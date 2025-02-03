package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/zerodha/logf"
)

type LogParasiteSystem struct {
	world    *ecs.World
	logger   *logf.Logger
	entities []physicsEntity
	dataChan chan RocketState
	done     chan struct{}
}

func NewLogParasiteSystem(world *ecs.World, logger *logf.Logger) *LogParasiteSystem {
	return &LogParasiteSystem{
		world:    world,
		logger:   logger,
		entities: make([]physicsEntity, 0),
		done:     make(chan struct{}),
	}
}

func (s *LogParasiteSystem) Start(dataChan chan RocketState) {
	s.dataChan = dataChan
	go s.processData()
}

func (s *LogParasiteSystem) Stop() {
	close(s.done)
}

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

func (s *LogParasiteSystem) Priority() int {
	return 1
}

func (s *LogParasiteSystem) Update(dt float32) {
	// No need to track time here - data comes from simulation state
	return
}

func (s *LogParasiteSystem) Add(entity *ecs.BasicEntity, pos *components.Position,
	vel *components.Velocity, acc *components.Acceleration, mass *components.Mass,
	motor *components.Motor, bodytube *components.Bodytube, nosecone *components.Nosecone,
	finset *components.TrapezoidFinset) {
	s.entities = append(s.entities, physicsEntity{entity, pos, vel, acc, mass, motor, bodytube, nosecone, finset})
}
