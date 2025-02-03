package simulation

import (
	"fmt"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/entities"
	"github.com/bxrne/launchrail/pkg/openrocket"
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
	rocket                *entities.RocketEntity
	config                *config.Config
	logger                *logf.Logger
	updateChan            chan struct{}
	doneChan              chan struct{}
	stateChan             chan systems.RocketState
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

	// Initialize parasite systems
	sim.logParasiteSystem = systems.NewLogParasiteSystem(world, log)                 // For logging
	sim.storageParasiteSystem = systems.NewStorageParasiteSystem(world, motionStore) // For motion storage

	// Start parasites
	sim.logParasiteSystem.Start(sim.stateChan)
	sim.storageParasiteSystem.Start(sim.stateChan)

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

	defer func() {
		if r := recover(); r != nil {
			switch r {
			case "Simulation ended: Ground impact":
				s.logger.Info("Simulation completed: Ground impact detected")
			default:
				// Log any other panics and re-panic
				s.logger.Error("Simulation failed", "error", r)
				panic(r)
			}
		}
	}()

	go s.runPhysicsSystem(dt)
	go s.runAerodynamicSystem(dt)

	for currentTime < maxTime {
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

		// Signal systems to update
		s.updateChan <- struct{}{}

		// Send state to parasites with current simulation time
		s.stateChan <- systems.RocketState{
			Time:         float64(currentTime),
			Altitude:     s.rocket.Position.Y,
			Velocity:     s.rocket.Velocity.Y,
			Acceleration: s.rocket.Acceleration.Y,
			Thrust:       thrust,
			MotorState:   motorState,
		}

		currentTime += dt
	}

	// Signal systems to stop
	close(s.doneChan)

	return nil
}

func (s *Simulation) runPhysicsSystem(dt float32) {
	for {
		select {
		case <-s.updateChan:
			s.physicsSystem.Update(dt)
		case <-s.doneChan:
			return
		}
	}
}

func (s *Simulation) runAerodynamicSystem(dt float32) {
	for {
		select {
		case <-s.updateChan:
			if err := s.aerodynamicSystem.Update(dt); err != nil {
				s.logger.Error("Aerodynamic system update failed", "error", err)
			}
		case <-s.doneChan:
			return
		}
	}
}
