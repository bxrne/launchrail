package simulation

import (
	"fmt"
	"math"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/plugin"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/entities"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/stats"
	"github.com/bxrne/launchrail/pkg/systems"

	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/zerodha/logf"
)

// Simulation represents a rocket simulation
type Simulation struct {
	world             *ecs.World
	physicsSystem     *systems.PhysicsSystem
	aerodynamicSystem *systems.AerodynamicSystem
	logParasiteSystem *systems.LogParasiteSystem
	motionParasite    *systems.StorageParasiteSystem
	eventsParasite    *systems.StorageParasiteSystem
	rulesSystem       *systems.RulesSystem
	rocket            *entities.RocketEntity
	config            *config.Config
	logger            *logf.Logger
	updateChan        chan struct{}
	doneChan          chan struct{}
	stateChan         chan systems.RocketState
	stats             *stats.FlightStats
	launchRailSystem  *systems.LaunchRailSystem
	currentTime       float64
	systems           []systems.System
	pluginManager     *plugin.Manager
}

// NewSimulation creates a new rocket simulation
func NewSimulation(cfg *config.Config, log *logf.Logger, stores *storage.Stores) (*Simulation, error) {
	world := &ecs.World{}

	sim := &Simulation{
		world:         world,
		config:        cfg,
		logger:        log,
		updateChan:    make(chan struct{}),
		doneChan:      make(chan struct{}),
		stateChan:     make(chan systems.RocketState, 100),
		pluginManager: plugin.NewManager(*log),
	}

	for _, pluginPath := range cfg.Plugins.Paths {
		if err := sim.pluginManager.LoadPlugin(pluginPath); err != nil {
			return nil, err
		}
	}

	// Initialize systems with optimized worker counts
	sim.physicsSystem = systems.NewPhysicsSystem(world, cfg)
	sim.aerodynamicSystem = systems.NewAerodynamicSystem(world, 4, cfg)
	rules := systems.NewRulesSystem(world, cfg)

	sim.rulesSystem = rules

	// Initialize launch rail system with config values
	sim.launchRailSystem = systems.NewLaunchRailSystem(
		world,
		cfg.Options.Launchrail.Length,
		cfg.Options.Launchrail.Angle,
		cfg.Options.Launchrail.Orientation,
	)

	// Initialize parasite systems with specific store types
	sim.logParasiteSystem = systems.NewLogParasiteSystem(world, log)
	sim.motionParasite = systems.NewStorageParasiteSystem(world, stores.Motion, storage.MOTION)
	sim.eventsParasite = systems.NewStorageParasiteSystem(world, stores.Events, storage.EVENTS)

	// Start parasites (only once)
	sim.logParasiteSystem.Start(sim.stateChan)
	err := sim.motionParasite.Start(sim.stateChan)
	if err != nil {
		return nil, err
	}

	err = sim.eventsParasite.Start(sim.stateChan)
	if err != nil {
		return nil, err
	}

	sim.stats = stats.NewFlightStats()

	// Add systems to the slice - Note: we should NOT add the event parasite here
	// as it's meant to be independent
	sim.systems = []systems.System{
		sim.physicsSystem,
		sim.aerodynamicSystem,
		sim.rulesSystem,
		sim.launchRailSystem,
		sim.logParasiteSystem,
		sim.motionParasite,
	}

	return sim, nil
}

// LoadRocket loads a rocket entity into the simulation
func (s *Simulation) LoadRocket(orkData *openrocket.RocketDocument, motorData *thrustcurves.MotorData) error {
	// Create motor component with logger
	motor, err := components.NewMotor(ecs.NewBasic(), motorData, *s.logger)
	if err != nil {
		return err
	}

	// Create rocket entity with all components
	s.rocket = entities.NewRocketEntity(s.world, orkData, motor)

	// Create a single PhysicsEntity to reuse for all systems
	sysEntity := &systems.PhysicsEntity{
		Entity:       s.rocket.BasicEntity,
		Position:     s.rocket.Position,
		Velocity:     s.rocket.Velocity,
		Acceleration: s.rocket.Acceleration,
		Mass:         s.rocket.Mass,
		Motor:        motor,
		Bodytube:     s.rocket.GetComponent("bodytube").(*components.Bodytube),
		Nosecone:     s.rocket.GetComponent("nosecone").(*components.Nosecone),
		Finset:       s.rocket.GetComponent("finset").(*components.TrapezoidFinset),
		Parachute:    s.rocket.GetComponent("parachute").(*components.Parachute),
	}

	// Add to all systems
	s.physicsSystem.Add(sysEntity)
	s.aerodynamicSystem.Add(sysEntity)
	s.rulesSystem.Add(sysEntity)
	s.launchRailSystem.Add(sysEntity)
	s.logParasiteSystem.Add(sysEntity)
	s.motionParasite.Add(sysEntity)

	return nil
}

// Run executes the simulation
func (s *Simulation) Run() error {
	defer func() {
		s.logParasiteSystem.Stop()
		s.motionParasite.Stop()
	}()

	// Validate simulation parameters
	if s.config.Simulation.Step <= 0 || s.config.Simulation.Step > 0.01 {
		return fmt.Errorf("invalid simulation step: must be between 0 and 0.01")
	}

	for {
		if err := s.updateSystems(); err != nil {
			return err
		}

		// Stop if landed - check rules system state
		if s.rulesSystem.GetLastEvent() == systems.Land {
			s.logger.Info("Rocket has landed; stopping simulation")
			break
		}

		s.currentTime += s.config.Simulation.Step

		// Also add a maximum time check to prevent infinite loops
		if s.currentTime >= s.config.Simulation.MaxTime {
			s.logger.Info("Reached maximum simulation time")
			break
		}
	}

	// Print stats after landing
	s.logger.Info("Flight Statistics",
		"stats", s.stats.String(),
	)

	close(s.doneChan)
	return nil
}

// updateCoreSystems updates the core systems in the simulation (no plugins)
func (s *Simulation) updateCoreSystems(state *systems.RocketState) error {
	// Update core systems
	for _, system := range s.systems {
		if err := system.Update(float64(s.config.Simulation.Step)); err != nil {
			return err
		}
	}

	// Apply parachute effects if deployed
	if parachute := s.rocket.GetComponent("parachute").(*components.Parachute); parachute.Deployed {
		// Apply additional drag force from parachute
		rho := s.aerodynamicSystem.GetAirDensity(float64(s.rocket.Position.Vec.Y))
		vel := s.rocket.Velocity.Vec.Y
		dragForce := -0.5 * float64(rho) * parachute.DragCoefficient * parachute.Area * vel * math.Abs(vel)
		s.rocket.Acceleration.Vec.Y += dragForce / s.rocket.Mass.Value
	}

	// Calculate velocity and acceleration
	vel := s.rocket.Velocity.Vec.Y
	acc := s.rocket.Acceleration.Vec.Y
	if math.IsNaN(acc) {
		acc = 0
	}

	// Calculate Mach number
	speedOfSound := s.aerodynamicSystem.GetSpeedOfSound(float64(s.rocket.Position.Vec.Y))
	mach := 0.0
	if speedOfSound > 1e-8 {
		mach = vel / float64(speedOfSound)
	}
	if math.IsNaN(mach) || math.IsInf(mach, 0) {
		mach = 0
	}

	// Update statistics
	s.stats.Update(
		s.currentTime,
		s.rocket.Position.Vec.Y,
		vel,
		acc,
		mach,
	)

	// Update motor component
	motorComp, ok := s.rocket.GetComponent("motor").(*components.Motor)
	if ok {
		if err := motorComp.Update(s.config.Simulation.Step); err != nil {
			return fmt.Errorf("motor update error: %w", err)
		}

		// Update state with final values
		state.Velocity = vel
		state.Acceleration = acc
		state.Thrust = motorComp.GetThrust()
		state.MotorState = motorComp.GetState()
	}

	return nil
}

// updateSystems updates all systems in the simulation
func (s *Simulation) updateSystems() error {
	// Create initial state with parachute status
	parachute := s.rocket.GetComponent("parachute").(*components.Parachute)
	state := &systems.RocketState{
		Time:              s.currentTime,
		Altitude:          s.rocket.Position.Vec.Y,
		Velocity:          s.rocket.Velocity.Vec.Y,
		Acceleration:      s.rocket.Acceleration.Vec.Y,
		ParachuteDeployed: parachute.IsDeployed(),
	}

	// Execute plugin BeforeSimStep hooks
	for _, plugin := range s.pluginManager.GetPlugins() {
		if err := plugin.BeforeSimStep(state); err != nil {
			return fmt.Errorf("plugin %s BeforeSimStep error: %w", plugin.Name(), err)
		}
	}

	// Update core systems
	if err := s.updateCoreSystems(state); err != nil {
		return err
	}

	// Execute plugin AfterSimStep hooks
	for _, plugin := range s.pluginManager.GetPlugins() {
		if err := plugin.AfterSimStep(state); err != nil {
			return fmt.Errorf("plugin %s AfterSimStep error: %w", plugin.Name(), err)
		}
	}

	// Send final state to channel
	s.stateChan <- *state

	return nil
}
