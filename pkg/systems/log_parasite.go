package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/zerodha/logf"
)

// LogParasiteSystem logs rocket state data
type LogParasiteSystem struct {
	world    *ecs.World
	logger   *logf.Logger
	entities []physicsEntity
	dataChan chan RocketState
	done     chan struct{}
}

// NewLogParasiteSystem creates a new LogParasiteSystem
func NewLogParasiteSystem(world *ecs.World, logger *logf.Logger) *LogParasiteSystem {
	return &LogParasiteSystem{
		world:    world,
		logger:   logger,
		entities: make([]physicsEntity, 0),
		done:     make(chan struct{}),
	}
}

// Start the LogParasiteSystem
func (s *LogParasiteSystem) Start(dataChan chan RocketState) {
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
				"altitude", state.Altitude,
				"velocity", state.Velocity,
				"acceleration", state.Acceleration,
				"thrust", state.Thrust,
				"motor_state", state.MotorState,
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
func (s *LogParasiteSystem) Update(dt float32) {
	// No need to track time here - data comes from simulation state
}

// Add adds entities to the system
func (s *LogParasiteSystem) Add(se *SystemEntity) {
	s.entities = append(s.entities, physicsEntity{se.Entity, se.Pos, se.Vel, se.Acc, se.Mass, se.Motor, se.Bodytube, se.Nosecone, se.Finset})
}
