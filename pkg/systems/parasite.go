package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/zerodha/logf"
)

type ParasiteSystem struct {
	world    *ecs.World
	entities []physicsEntity
	storage  *storage.Storage
	log      bool
	logger   *logf.Logger
	time     float64
}

func NewParasiteSystem(world *ecs.World, storage *storage.Storage, log bool, logger *logf.Logger) *ParasiteSystem {
	return &ParasiteSystem{
		world:    world,
		entities: make([]physicsEntity, 0),
		storage:  storage,
		log:      log,
		logger:   logger,
		time:     0,
	}
}

func (s *ParasiteSystem) Priority() int {
	return 1
}

func (s *ParasiteSystem) Update(dt float32) {
	currentTime := s.time

	for _, entity := range s.entities {
		if s.storage != nil {
			// Log only time and thrust
			thrustValue := 0.0
			if entity.Motor != nil {
				thrustValue = entity.Motor.GetThrust()
			}

			s.logger.Debug("Motor thrust",
				"time", currentTime,
				"thrust", thrustValue)
		}
	}

	s.time += float64(dt)
}

func (s *ParasiteSystem) Add(entity *ecs.BasicEntity, pos *components.Position,
	vel *components.Velocity, acc *components.Acceleration, mass *components.Mass,
	motor *components.Motor, bodytube *components.Bodytube, nosecone *components.Nosecone,
	finset *components.TrapezoidFinset) {
	s.entities = append(s.entities, physicsEntity{entity, pos, vel, acc, mass, motor, bodytube, nosecone, finset})
}
