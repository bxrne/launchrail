package simulation

import (
	"fmt"
	"math"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
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
	world                 *ecs.World
	physicsSystem         *systems.PhysicsSystem
	aerodynamicSystem     *systems.AerodynamicSystem
	logParasiteSystem     *systems.LogParasiteSystem
	storageParasiteSystem *systems.StorageParasiteSystem
	rulesSystem           *systems.RulesSystem
	rocket                *entities.RocketEntity
	config                *config.Config
	logger                *logf.Logger
	updateChan            chan struct{}
	doneChan              chan struct{}
	stateChan             chan systems.RocketState
	stats                 *stats.FlightStats
	launchRailSystem      *systems.LaunchRailSystem
}

// NewSimulation creates a new rocket simulation
func NewSimulation(cfg *config.Config, log *logf.Logger, motionStore *storage.Storage) (*Simulation, error) {
	world := &ecs.World{}

	sim := &Simulation{
		world:      world,
		config:     cfg,
		logger:     log,
		updateChan: make(chan struct{}),
		doneChan:   make(chan struct{}),
		stateChan:  make(chan systems.RocketState, 100), // Buffered channel
	}

	// Initialize systems with optimized worker counts
	sim.physicsSystem = systems.NewPhysicsSystem(world)
	sim.aerodynamicSystem = systems.NewAerodynamicSystem(world, 4) // Add worker count
	sim.rulesSystem = systems.NewRulesSystem(world)                // Add this line

	// Initialize launch rail system with config values
	sim.launchRailSystem = systems.NewLaunchRailSystem(
		world,
		cfg.Options.Launchrail.Length,
		cfg.Options.Launchrail.Angle,
		cfg.Options.Launchrail.Orientation,
	)

	// Initialize parasite systems
	sim.logParasiteSystem = systems.NewLogParasiteSystem(world, log)
	sim.storageParasiteSystem = systems.NewStorageParasiteSystem(world, motionStore)

	// Start parasites
	sim.logParasiteSystem.Start(sim.stateChan)
	sim.storageParasiteSystem.Start(sim.stateChan)

	sim.stats = stats.NewFlightStats()

	return sim, nil
}

// LoadRocket loads a rocket entity into the simulation
func (s *Simulation) LoadRocket(orkData *openrocket.RocketDocument, motorData *thrustcurves.MotorData) error {
	// Create motor component with logger
	motor := components.NewMotor(ecs.NewBasic(), motorData, *s.logger)

	// Create rocket entity with all components
	s.rocket = entities.NewRocketEntity(s.world, orkData, motor)

	// Create a single SystemEntity to reuse for all systems
	sysEntity := &systems.SystemEntity{
		Entity:   s.rocket.BasicEntity,
		Pos:      s.rocket.Position,
		Vel:      s.rocket.Velocity,
		Acc:      s.rocket.Acceleration,
		Mass:     s.rocket.Mass,
		Motor:    motor,
		Bodytube: s.rocket.GetComponent("bodytube").(*components.Bodytube),
		Nosecone: s.rocket.GetComponent("nosecone").(*components.Nosecone),
		Finset:   s.rocket.GetComponent("finset").(*components.TrapezoidFinset),
	}

	// Add to all systems
	s.physicsSystem.Add(sysEntity)
	s.aerodynamicSystem.Add(sysEntity)
	s.rulesSystem.Add(sysEntity)
	s.launchRailSystem.Add(sysEntity)
	s.logParasiteSystem.Add(sysEntity)
	s.storageParasiteSystem.Add(sysEntity)

	return nil
}

// Run executes the simulation
func (s *Simulation) Run() error {
	defer func() {
		s.logParasiteSystem.Stop()
		s.storageParasiteSystem.Stop()
	}()

	// Validate simulation parameters
	if s.config.Simulation.Step <= 0 || s.config.Simulation.Step > 0.01 {
		return fmt.Errorf("invalid simulation step: must be between 0 and 0.01")
	}
	if s.config.Simulation.MaxTime <= 0 || s.config.Simulation.MaxTime > 120 {
		return fmt.Errorf("invalid max time: must be between 0 and 120")
	}

	dt := float32(s.config.Simulation.Step)
	currentTime := float32(0)
	maxTime := float32(s.config.Simulation.MaxTime)

	for currentTime < maxTime {
		// Update launch rail constraints first
		if err := s.launchRailSystem.Update(dt); err != nil {
			return fmt.Errorf("launch rail error at t=%v: %w", currentTime, err)
		}

		// Update motor first to check for errors
		motorState := "COASTING"
		thrust := 0.0
		if motor := s.rocket.GetComponent("motor").(*components.Motor); motor != nil {
			if err := motor.Update(float64(dt)); err != nil {
				return fmt.Errorf("motor error at t=%v: %w", currentTime, err)
			}
			thrust = motor.GetThrust()
			motorState = motor.GetState()
		}

		// Update physics and aerodynamics
		s.physicsSystem.Update(dt)
		if err := s.aerodynamicSystem.Update(dt); err != nil {
			s.logger.Error("Aerodynamic system update failed", "error", err)
		}

		// Check rules
		event := s.rulesSystem.Update(dt)

		// Create current state
		state := systems.RocketState{
			Time:         float64(currentTime),
			Altitude:     s.rocket.Position.Y,
			Velocity:     s.rocket.Velocity.Y,
			Acceleration: s.rocket.Acceleration.Y,
			Thrust:       thrust,
			MotorState:   motorState,
		}

		// Update stats before handling events
		mach := math.Abs(s.rocket.Velocity.Y) / 340.0
		s.stats.Update(
			float64(currentTime),
			s.rocket.Position.Y,
			s.rocket.Velocity.Y,
			s.rocket.Acceleration.Y,
			mach,
		)

		// Handle events and potentially modify state
		switch event {
		case systems.Apogee:
			s.logger.Info("Apogee detected",
				"time", currentTime,
				"altitude", state.Altitude,
				"velocity", state.Velocity,
			)
			s.stateChan <- state

		case systems.Land:
			// Log the final state before zeroing
			s.stateChan <- state

			s.logger.Info("Landing detected - simulation complete",
				"time", currentTime,
				"altitude", state.Altitude,
				"velocity", state.Velocity,
				"acceleration", state.Acceleration,
			)

			// Send final landed state
			landedState := systems.RocketState{
				Time:         float64(currentTime),
				Altitude:     0,
				Velocity:     0,
				Acceleration: 0,
				Thrust:       0,
				MotorState:   "LANDED",
			}

			// Print stats before final state
			s.logger.Info("Flight Statistics",
				"stats", s.stats.String(),
			)

			s.stateChan <- landedState
			close(s.doneChan)
			return nil

		default:
			s.stateChan <- state
		}

		currentTime += dt
	}

	s.logger.Warn("Simulation reached max time without landing",
		"maxTime", maxTime,
		"finalAltitude", s.rocket.Position.Y)

	// Print stats even if max time reached
	s.logger.Info("Flight Statistics",
		"stats", s.stats.String(),
	)

	close(s.doneChan)
	return nil
}
