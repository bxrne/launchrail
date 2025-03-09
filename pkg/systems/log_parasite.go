package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/zerodha/logf"
)

// LogParasiteSystem logs rocket state data
type LogParasiteSystem struct {
	world    *ecs.World
	logger   *logf.Logger
	entities []states.PhysicsState
	dataChan chan *states.PhysicsState
	done     chan struct{}
}

// NewLogParasiteSystem creates a new LogParasiteSystem
func NewLogParasiteSystem(world *ecs.World, logger *logf.Logger) *LogParasiteSystem {
	return &LogParasiteSystem{
		world:    world,
		logger:   logger,
		entities: make([]states.PhysicsState, 0),
		done:     make(chan struct{}),
	}
}

// Start the LogParasiteSystem
func (s *LogParasiteSystem) Start(dataChan chan *states.PhysicsState) {
	s.dataChan = dataChan
	go s.processData()
}

// Stop the LogParasiteSystem
func (s *LogParasiteSystem) Stop() {
	close(s.done)
}

// processData logs rocket state data
func (s *LogParasiteSystem) processData() {
	for {
		select {
		case state := <-s.dataChan:
			s.logger.Debug("rocket_state",
				"time", state.Time,
				"altitude", state.Position.Vec.Y,
				"velocity", state.Velocity.Vec.Y,
				"acceleration", state.Acceleration.Vec.Y,
				"thrust", state.Motor.GetThrust(),
				"motor_state", state.Motor.GetState(),
			)
		case <-s.done:
			return
		}
	}
}

// Priority returns the system priority
func (s *LogParasiteSystem) Priority() int {
	return 1
}

// Update the LogParasiteSystem
func (s *LogParasiteSystem) Update(dt float64) error {
	// No need to track time here - data comes from simulation state
	return nil
}

// Add adds entities to the system
func (s *LogParasiteSystem) Add(pe *states.PhysicsState) {
	s.entities = append(s.entities,
		states.PhysicsState{
			Entity:          pe.Entity,
			Position:        pe.Position,
			Velocity:        pe.Velocity,
			Acceleration:    pe.Acceleration,
			Orientation:     pe.Orientation,
			Mass:            pe.Mass,
			Motor:           pe.Motor,
			Bodytube:        pe.Bodytube,
			Nosecone:        pe.Nosecone,
			Finset:          pe.Finset,
			Parachute:       pe.Parachute,
			AngularVelocity: pe.AngularVelocity,
		})
}
