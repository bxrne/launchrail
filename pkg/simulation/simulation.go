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
	world             *ecs.World
	physicsSystem     *systems.PhysicsSystem
	aerodynamicSystem *systems.AerodynamicSystem
	parasiteSystem    *systems.ParasiteSystem
	loggingSystem     *systems.ParasiteSystem
	rocket            *entities.RocketEntity
	config            *config.Config
	logger            *logf.Logger
}

func NewSimulation(cfg *config.Config, log *logf.Logger, motionStore *storage.Storage) (*Simulation, error) {
	world := &ecs.World{}

	sim := &Simulation{
		world:  world,
		config: cfg,
		logger: log,
	}

	// Initialize systems with optimized worker counts
	sim.physicsSystem = systems.NewPhysicsSystem(world)
	sim.aerodynamicSystem = systems.NewAerodynamicSystem(world, 4) // Add worker count
	sim.parasiteSystem = systems.NewParasiteSystem(world, motionStore, false, log)
	sim.loggingSystem = systems.NewParasiteSystem(world, nil, true, log)

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

	// Add to parasite system with finset
	s.parasiteSystem.Add(
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

	// Similar for other systems...

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

	s.loggingSystem.Add(
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

	for currentTime < maxTime {
		// Update motor first to check for errors
		if entity := s.rocket.GetComponent("motor").(*components.Motor); entity != nil {
			if err := entity.Update(float64(dt)); err != nil {
				return fmt.Errorf("motor error at t=%v: %w", currentTime, err)
			}
		}

		// Rest of simulation updates
		s.physicsSystem.Update(dt)
		if err := s.aerodynamicSystem.Update(dt); err != nil {
			return err
		}
		s.parasiteSystem.Update(dt)
		s.loggingSystem.Update(dt)
		currentTime += dt
	}

	return nil
}
