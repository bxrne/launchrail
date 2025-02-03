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
	sim.logParasiteSystem = systems.NewLogParasiteSystem(world, log)                 // For logging
	sim.storageParasiteSystem = systems.NewStorageParasiteSystem(world, motionStore) // For motion storage

	// Start parasites
	sim.logParasiteSystem.Start(sim.stateChan)
	sim.storageParasiteSystem.Start(sim.stateChan)

	sim.stats = stats.NewFlightStats()

	return sim, nil
}

func (s *Simulation) LoadRocket(orkData *openrocket.RocketDocument, motorData *thrustcurves.MotorData) error {
	// Create motor component with logger
	motor := components.NewMotor(ecs.NewBasic(), motorData, *s.logger)

	// Create rocket entity with all components
	s.rocket = entities.NewRocketEntity(s.world, orkData, motor)

	// Add components to physics system
	s.physicsSystem.Add(
		s.rocket.BasicEntity,
		s.rocket.Position,
		s.rocket.Velocity,
		s.rocket.Acceleration,
		s.rocket.Mass,
		motor,
		s.rocket.GetComponent("bodytube").(*components.Bodytube),
		s.rocket.GetComponent("nosecone").(*components.Nosecone),
	)

	// Get optional finset component
	finset := s.rocket.GetComponent("finset").(*components.TrapezoidFinset)

	// Add to aerodynamic system with finset
	s.aerodynamicSystem.Add(
		s.rocket.BasicEntity,
		s.rocket.Position,
		s.rocket.Velocity,
		s.rocket.Acceleration,
		s.rocket.Mass,
		motor,
		s.rocket.GetComponent("bodytube").(*components.Bodytube),
		s.rocket.GetComponent("nosecone").(*components.Nosecone),
		finset, // This may be nil
	)

	// Add to rules system
	s.rulesSystem.Add(
		s.rocket.BasicEntity,
		s.rocket.Position,
		s.rocket.Velocity,
		s.rocket.Acceleration,
		s.rocket.Mass,
		motor,
		s.rocket.GetComponent("bodytube").(*components.Bodytube),
		s.rocket.GetComponent("nosecone").(*components.Nosecone),
		finset,
	)

	// Add to launch rail system
	s.launchRailSystem.Add(
		s.rocket.BasicEntity,
		s.rocket.Position,
		s.rocket.Velocity,
		s.rocket.Acceleration,
		s.rocket.Mass,
		motor,
		s.rocket.GetComponent("bodytube").(*components.Bodytube),
		s.rocket.GetComponent("nosecone").(*components.Nosecone),
		finset,
	)

	// Add to log parasite system
	s.logParasiteSystem.Add(
		s.rocket.BasicEntity,
		s.rocket.Position,
		s.rocket.Velocity,
		s.rocket.Acceleration,
		s.rocket.Mass,
		motor,
		s.rocket.GetComponent("bodytube").(*components.Bodytube),
		s.rocket.GetComponent("nosecone").(*components.Nosecone),
		finset, // This may be nil
	)

	// Add to storage parasite system
	s.storageParasiteSystem.Add(
		s.rocket.BasicEntity,
		s.rocket.Position,
		s.rocket.Velocity,
		s.rocket.Acceleration,
		s.rocket.Mass,
		motor,
		s.rocket.GetComponent("bodytube").(*components.Bodytube),
		s.rocket.GetComponent("nosecone").(*components.Nosecone),
		finset, // This may be nil
	)

	return nil
}

func (s *Simulation) Run() error {
	defer func() {
		s.logParasiteSystem.Stop()
		s.storageParasiteSystem.Stop()
	}()

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

		// Send current state before handling events
		state := systems.RocketState{
			Time:         float64(currentTime),
			Altitude:     s.rocket.Position.Y,
			Velocity:     s.rocket.Velocity.Y,
			Acceleration: s.rocket.Acceleration.Y,
			Thrust:       thrust,
			MotorState:   motorState,
		}
		s.stateChan <- state

		// Update stats
		mach := math.Abs(s.rocket.Velocity.Y) / 340.0 // approximate sound speed
		s.stats.Update(
			float64(currentTime),
			s.rocket.Position.Y,
			s.rocket.Velocity.Y,
			s.rocket.Acceleration.Y,
			mach,
		)

		// Handle events
		switch event {
		case systems.Apogee:
			s.logger.Info("Apogee detected",
				"time", currentTime,
				"altitude", s.rocket.Position.Y,
				"velocity", s.rocket.Velocity.Y,
			)

		case systems.Land:
			s.logger.Info("Landing detected - simulation complete",
				"time", currentTime,
				"altitude", s.rocket.Position.Y,
				"velocity", s.rocket.Velocity.Y,
				"acceleration", s.rocket.Acceleration.Y,
			)

			// Ensure final state shows landed position
			s.rocket.Position.Y = 0
			s.rocket.Velocity.Y = 0
			s.rocket.Acceleration.Y = 0

			// Send final state update
			landedState := systems.RocketState{
				Time:         float64(currentTime),
				Altitude:     0,
				Velocity:     0,
				Acceleration: 0,
				Thrust:       0,
				MotorState:   "LANDED",
			}
			s.stateChan <- landedState

			// Print final stats
			s.logger.Info("Flight Statistics",
				"stats", s.stats.String(),
			)

			// Close channels and return
			close(s.doneChan)
			return nil
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
