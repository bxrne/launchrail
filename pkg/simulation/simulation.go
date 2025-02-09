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
	currentTime           float64
	systems               []systems.System // Now using the System interface
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

	// Initialize parasite systems
	sim.logParasiteSystem = systems.NewLogParasiteSystem(world, log)
	sim.storageParasiteSystem = systems.NewStorageParasiteSystem(world, motionStore)

	// Start parasites
	sim.logParasiteSystem.Start(sim.stateChan)
	sim.storageParasiteSystem.Start(sim.stateChan)

	sim.stats = stats.NewFlightStats()

	// Add systems to the slice
	sim.systems = []systems.System{
		sim.physicsSystem,
		sim.aerodynamicSystem,
		sim.rulesSystem,
		sim.launchRailSystem,
		sim.logParasiteSystem,
		sim.storageParasiteSystem,
	}

	return sim, nil
}

// LoadRocket loads a rocket entity into the simulation
func (s *Simulation) LoadRocket(orkData *openrocket.RocketDocument, motorData *thrustcurves.MotorData) error {
	// Create motor component with logger
	motor := components.NewMotor(ecs.NewBasic(), motorData, *s.logger)

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

	for {
		if err := s.updateSystems(); err != nil {
			return err
		}
		// Stop if landed
		if s.rulesSystem.GetLastEvent() == systems.Land {
			s.logger.Info("Rocket has landed; stopping simulation")
			break
		}
		s.currentTime += s.config.Simulation.Step
	}

	// Print stats after landing
	s.logger.Info("Flight Statistics",
		"stats", s.stats.String(),
	)

	close(s.doneChan)
	return nil
}

func (s *Simulation) updateSystems() error {
	for _, system := range s.systems {
		if err := system.Update(float32(s.config.Simulation.Step)); err != nil {
			return err
		}
	}

	vel := s.rocket.Velocity.Vec.Y
	acc := s.rocket.Acceleration.Vec.Y
	if math.IsNaN(acc) {
		acc = 0
	}
	speedOfSound := s.aerodynamicSystem.GetSpeedOfSound(float32(s.rocket.Position.Vec.Y))
	mach := 0.0
	if speedOfSound > 1e-8 {
		mach = vel / float64(speedOfSound)
	}
	if math.IsNaN(mach) || math.IsInf(mach, 0) {
		mach = 0
	}

	s.stats.Update(
		s.currentTime,
		s.rocket.Position.Vec.Y,
		vel,
		acc,
		mach,
	)

	motorComp, ok := s.rocket.GetComponent("motor").(*components.Motor)
	if ok {
		motorComp.Update(s.config.Simulation.Step)
		s.stateChan <- systems.RocketState{
			Time:         s.currentTime,
			Altitude:     s.rocket.Position.Vec.Y,
			Velocity:     vel,
			Acceleration: acc,
			Thrust:       motorComp.GetThrust(),
			MotorState:   motorComp.GetState(),
		}
	}

	return nil
}
