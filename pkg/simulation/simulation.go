package simulation

import (
	"fmt"

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
	state := &states.PhysicsState{
		Time:                s.currentTime,
		Entity:              s.rocket.BasicEntity,
		Position:            s.rocket.Position,
		Orientation:         s.rocket.Orientation,
		AngularVelocity:     s.rocket.AngularVelocity,
		AngularAcceleration: s.rocket.AngularAcceleration,
		Velocity:            s.rocket.Velocity,
		Acceleration:        s.rocket.Acceleration,
		Mass:                s.rocket.Mass,
		Motor:               s.rocket.GetComponent("motor").(*components.Motor),
		Bodytube:            s.rocket.GetComponent("bodytube").(*components.Bodytube),
		Nosecone:            s.rocket.GetComponent("nosecone").(*components.Nosecone),
		Finset:              s.rocket.GetComponent("finset").(*components.TrapezoidFinset),
		Parachute:           s.rocket.GetComponent("parachute").(*components.Parachute),
	}

	// Execute plugins before systems
	for _, plugin := range s.pluginManager.GetPlugins() {
		if err := plugin.BeforeSimStep(state); err != nil {
			return fmt.Errorf("plugin %s BeforeSimStep error: %w", plugin.Name(), err)
		}
	}

	// Update motor first
	if err := state.Motor.Update(s.config.Engine.Simulation.Step); err != nil {
		return err
	}

	// Execute systems in order and propagate state changes
	for _, system := range s.systems {
		if err := system.Update(s.config.Engine.Simulation.Step); err != nil {
			return fmt.Errorf("system %T update error: %w", system, err)
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
