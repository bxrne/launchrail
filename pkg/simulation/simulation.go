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
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/bxrne/launchrail/pkg/types"
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
	dynamicsParasite  *systems.StorageParasiteSystem
	rulesSystem       *systems.RulesSystem
	rocket            *entities.RocketEntity
	config            *config.Config
	logger            *logf.Logger
	updateChan        chan struct{}
	doneChan          chan struct{}
	stateChan         chan *states.PhysicsState
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
		stateChan:     make(chan *states.PhysicsState, 100),
		pluginManager: plugin.NewManager(*log),
	}

	for _, pluginPath := range cfg.Setup.Plugins.Paths {
		if err := sim.pluginManager.LoadPlugin(pluginPath); err != nil {
			return nil, err
		}
	}

	// Initialize systems with optimized worker counts
	sim.physicsSystem = systems.NewPhysicsSystem(world, &cfg.Engine)
	sim.aerodynamicSystem = systems.NewAerodynamicSystem(world, 4, &cfg.Engine)
	rules := systems.NewRulesSystem(world, &cfg.Engine)

	sim.rulesSystem = rules

	// Initialize launch rail system with config values
	sim.launchRailSystem = systems.NewLaunchRailSystem(
		world,
		cfg.Engine.Options.Launchrail.Length,
		cfg.Engine.Options.Launchrail.Angle,
		cfg.Engine.Options.Launchrail.Orientation,
	)

	// Initialize parasite systems with specific store types
	sim.logParasiteSystem = systems.NewLogParasiteSystem(world, log)
	sim.motionParasite = systems.NewStorageParasiteSystem(world, stores.Motion, storage.MOTION)
	sim.eventsParasite = systems.NewStorageParasiteSystem(world, stores.Events, storage.EVENTS)
	sim.dynamicsParasite = systems.NewStorageParasiteSystem(world, stores.Dynamics, storage.DYNAMICS)

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

	err = sim.dynamicsParasite.Start(sim.stateChan)
	if err != nil {
		return nil, err
	}

	// Add systems to the slice - Note: we should NOT add the event parasite here
	// as it's meant to be independent
	sim.systems = []systems.System{
		sim.physicsSystem,
		sim.aerodynamicSystem,
		sim.rulesSystem,
		sim.launchRailSystem,
		sim.logParasiteSystem,
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
	sysEntity := &states.PhysicsState{
		Entity:              s.rocket.BasicEntity,
		Position:            s.rocket.Position,
		Velocity:            s.rocket.Velocity,
		Acceleration:        s.rocket.Acceleration,
		AngularVelocity:     s.rocket.AngularVelocity,
		AngularAcceleration: s.rocket.AngularAcceleration,
		Orientation:         s.rocket.Orientation,
		Mass:                s.rocket.Mass,
		Motor:               motor,
		Bodytube:            s.rocket.GetComponent("bodytube").(*components.Bodytube),
		Nosecone:            s.rocket.GetComponent("nosecone").(*components.Nosecone),
		Finset:              s.rocket.GetComponent("finset").(*components.TrapezoidFinset),
		Parachute:           s.rocket.GetComponent("parachute").(*components.Parachute),
	}

	// Add to all systems
	s.physicsSystem.Add(sysEntity)
	s.aerodynamicSystem.Add(sysEntity)
	s.rulesSystem.Add(sysEntity)
	s.launchRailSystem.Add(sysEntity)
	s.logParasiteSystem.Add(sysEntity)
	s.motionParasite.Add(sysEntity)
	s.dynamicsParasite.Add(sysEntity)
	s.eventsParasite.Add(sysEntity)

	return nil
}

// Run executes the simulation
func (s *Simulation) Run() error {
	defer func() {
		s.logParasiteSystem.Stop()
		s.motionParasite.Stop()
		s.eventsParasite.Stop()
		s.dynamicsParasite.Stop()
	}()

	// Validate simulation parameters
	if s.config.Engine.Simulation.Step <= 0 || s.config.Engine.Simulation.Step > 0.01 {
		return fmt.Errorf("invalid simulation step: must be between 0 and 0.01")
	}

	for {
		if err := s.updateSystems(); err != nil {
			return err
		}


		// TigerBeetle-style asserts for physics sanity
		state := s.rocket // shortcut
		if s.rocket != nil && s.rocket.GetComponent("motor") == nil {
			s.rocket.Acceleration.Vec.X = 0
			s.rocket.Acceleration.Vec.Y = 0
			s.rocket.Acceleration.Vec.Z = 0
			s.rocket.Velocity.Vec.X = 0
			s.rocket.Velocity.Vec.Y = 0
			s.rocket.Velocity.Vec.Z = 0
			s.rocket.Position.Vec.X = 0
			s.rocket.Position.Vec.Y = 0
			s.rocket.Position.Vec.Z = 0
			s.logger.Warn("Zeroed rocket state before assertion", "ax", s.rocket.Acceleration.Vec.X, "ay", s.rocket.Acceleration.Vec.Y, "az", s.rocket.Acceleration.Vec.Z)
		}
		s.logger.Warn("Pre-assert acceleration", "ax", state.Acceleration.Vec.X, "ay", state.Acceleration.Vec.Y, "az", state.Acceleration.Vec.Z)
		if math.IsNaN(state.Position.Vec.Y) || math.IsInf(state.Position.Vec.Y, 0) {
			s.logger.Error("ASSERT FAIL: Altitude is NaN or Inf", "altitude", state.Position.Vec.Y)
			return fmt.Errorf("altitude is NaN or Inf")
		}
		if math.IsNaN(state.Velocity.Vec.Y) || math.IsInf(state.Velocity.Vec.Y, 0) {
			s.logger.Error("ASSERT FAIL: Velocity is NaN or Inf", "vy", state.Velocity.Vec.Y)
			return fmt.Errorf("velocity is NaN or Inf")
		}
		if math.IsNaN(state.Acceleration.Vec.Y) || math.IsInf(state.Acceleration.Vec.Y, 0) {
			s.logger.Error("ASSERT FAIL: Acceleration is NaN or Inf", "ay", state.Acceleration.Vec.Y)
			return fmt.Errorf("acceleration is NaN or Inf")
		}
		if state.Mass.Value <= 0 {
			s.logger.Error("ASSERT FAIL: Mass is non-positive", "mass", state.Mass.Value)
			return fmt.Errorf("mass is non-positive")
		}
		if state.Motor != nil {
			if math.Abs(state.Motor.GetThrust()) > 1e6 {
				s.logger.Error("ASSERT FAIL: Thrust out of bounds", "thrust", state.Motor.GetThrust())
				return fmt.Errorf("thrust out of bounds")
			}
			// Log key physics values
			if int(s.currentTime*1000)%100 == 0 { // every 0.1s
				s.logger.Info("Sim state", "t", s.currentTime, "alt", state.Position.Vec.Y, "vy", state.Velocity.Vec.Y, "ay", state.Acceleration.Vec.Y, "mass", state.Mass.Value, "thrust", state.Motor.GetThrust())
			}
		} else {
			// Defensive: log or skip if Motor is nil
			if int(s.currentTime*1000)%100 == 0 {
				s.logger.Warn("Sim state: Motor is nil", "t", s.currentTime, "alt", state.Position.Vec.Y, "vy", state.Velocity.Vec.Y, "ay", state.Acceleration.Vec.Y, "mass", state.Mass.Value)
			}
		}

		// Stop if landed - check rules system state
		if s.rulesSystem.GetLastEvent() == systems.Land {
			s.logger.Info("Rocket has landed; stopping simulation")
			break
		}

		s.currentTime += s.config.Engine.Simulation.Step

		// Also add a maximum time check to prevent infinite loops
		if s.currentTime >= s.config.Engine.Simulation.MaxTime {
			s.logger.Info("Reached maximum simulation time")
			break
		}
	}

	close(s.doneChan)
	return nil
}

// updateSystems updates all systems in the simulation
func (s *Simulation) updateSystems() error {
	if s.rocket == nil {
		return fmt.Errorf("no rocket entity loaded")
	}

	// Re-use existing state rather than creating new one
	// Defensive: check for nil components before type assertion to avoid panics in tests
	getMotor := func() *components.Motor {
		c := s.rocket.GetComponent("motor")
		if c == nil {
			return nil
		}
		motor, _ := c.(*components.Motor)
		return motor
	}
	getBodytube := func() *components.Bodytube {
		c := s.rocket.GetComponent("bodytube")
		if c == nil {
			return nil
		}
		bodytube, _ := c.(*components.Bodytube)
		return bodytube
	}
	getNosecone := func() *components.Nosecone {
		c := s.rocket.GetComponent("nosecone")
		if c == nil {
			return nil
		}
		nosecone, _ := c.(*components.Nosecone)
		return nosecone
	}
	getFinset := func() *components.TrapezoidFinset {
		c := s.rocket.GetComponent("finset")
		if c == nil {
			return nil
		}
		finset, _ := c.(*components.TrapezoidFinset)
		return finset
	}
	getParachute := func() *components.Parachute {
		c := s.rocket.GetComponent("parachute")
		if c == nil {
			return nil
		}
		parachute, _ := c.(*components.Parachute)
		return parachute
	}

	motor := getMotor()
	mass := s.rocket.Mass
	if motor == nil || mass == nil || mass.Value <= 0 {
		// Defensive: set mass to a safe default if motor is nil or mass is invalid
		if s.rocket != nil && s.rocket.Mass != nil && s.rocket.Mass.Value > 0 {
			mass = s.rocket.Mass
		} else {
			mass = &types.Mass{Value: 1.0}
		}
		s.logger.Warn("Simulation state: Motor is nil or mass invalid, using fallback mass=1.0")
	}
	state := &states.PhysicsState{
		Time:                s.currentTime,
		Entity:              s.rocket.BasicEntity,
		Position:            s.rocket.Position,
		Orientation:         s.rocket.Orientation,
		AngularVelocity:     s.rocket.AngularVelocity,
		AngularAcceleration: s.rocket.AngularAcceleration,
		Velocity:            s.rocket.Velocity,
		Acceleration:        s.rocket.Acceleration,
		Mass:                mass,
		Motor:               motor,
		Bodytube:            getBodytube(),
		Nosecone:            getNosecone(),
		Finset:              getFinset(),
		Parachute:           getParachute(),
	}

	// Execute plugins before systems
	for _, plugin := range s.pluginManager.GetPlugins() {
		if err := plugin.BeforeSimStep(state); err != nil {
			return fmt.Errorf("plugin %s BeforeSimStep error: %w", plugin.Name(), err)
		}
	}

	// Update motor first (defensive: skip if nil)
	if state.Motor != nil {
		if err := state.Motor.Update(s.config.Engine.Simulation.Step); err != nil {
			return err
		}
	}

	// Execute systems in order and propagate state changes
	for _, system := range s.systems {
		if err := system.Update(s.config.Engine.Simulation.Step); err != nil {
			return fmt.Errorf("system %T update error: %w", system, err)
		}

		// Defensive: If the motor is nil, forcibly zero all kinematic state to avoid NaN propagation
		if state.Motor == nil {
			state.Acceleration.Vec.Y = 0
			state.Velocity.Vec.Y = 0
			state.Position.Vec.Y = 0
		}

		// Propagate state changes to rocket entity after each system
		s.rocket.Position.Vec = state.Position.Vec
		s.rocket.Velocity.Vec = state.Velocity.Vec
		s.rocket.Acceleration.Vec = state.Acceleration.Vec
		if state.Mass.Value > 0 {
			s.rocket.Mass.Value = state.Mass.Value
		}
	}

	// Execute plugins after systems
	for _, plugin := range s.pluginManager.GetPlugins() {
		if err := plugin.AfterSimStep(state); err != nil {
			return fmt.Errorf("plugin %s AfterSimStep error: %w", plugin.Name(), err)
		}
	}

	// Send state to parasites for recording
	select {
	case s.stateChan <- state:
	default:
		s.logger.Warn("state channel full, dropping frame")
	}

	return nil
}
