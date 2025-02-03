package systems

import (
	"fmt"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/components"
)

type StorageParasiteSystem struct {
	world    *ecs.World
	storage  *storage.Storage
	entities []physicsEntity
	dataChan chan RocketState
	done     chan struct{}
}

func NewStorageParasiteSystem(world *ecs.World, storage *storage.Storage) *StorageParasiteSystem {
	return &StorageParasiteSystem{
		world:    world,
		storage:  storage,
		entities: make([]physicsEntity, 0),
		done:     make(chan struct{}),
	}
}

func (s *StorageParasiteSystem) Start(dataChan chan RocketState) {
	s.dataChan = dataChan
	go s.processData()
}

func (s *StorageParasiteSystem) Stop() {
	close(s.done)
}

func (s *StorageParasiteSystem) processData() {
	for {
		select {
		case state := <-s.dataChan:
			record := []string{
				fmt.Sprintf("%.6f", state.Time),
				fmt.Sprintf("%.6f", state.Altitude),
				fmt.Sprintf("%.6f", state.Velocity),
				fmt.Sprintf("%.6f", state.Acceleration),
				fmt.Sprintf("%.6f", state.Thrust),
			}
			if err := s.storage.Write(record); err != nil {
				fmt.Printf("Error writing record: %v\n", err)
			}
		case <-s.done:
			return
		}
	}
}

func (s *StorageParasiteSystem) Priority() int {
	return 1
}

func (s *StorageParasiteSystem) Update(dt float32) {
	// No need to track time here - data comes from simulation state
	return
}

func (s *StorageParasiteSystem) Add(entity *ecs.BasicEntity, pos *components.Position,
	vel *components.Velocity, acc *components.Acceleration, mass *components.Mass,
	motor *components.Motor, bodytube *components.Bodytube, nosecone *components.Nosecone,
	finset *components.TrapezoidFinset) {
	s.entities = append(s.entities, physicsEntity{entity, pos, vel, acc, mass, motor, bodytube, nosecone, finset})
}
