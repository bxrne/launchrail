package systems

import (
	"fmt"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/storage"
)

// StorageParasiteSystem logs rocket state data to storage
type StorageParasiteSystem struct {
	world    *ecs.World
	storage  *storage.Storage
	entities []physicsEntity
	dataChan chan RocketState
	done     chan struct{}
}

// NewStorageParasiteSystem creates a new StorageParasiteSystem
func NewStorageParasiteSystem(world *ecs.World, storage *storage.Storage) *StorageParasiteSystem {
	return &StorageParasiteSystem{
		world:    world,
		storage:  storage,
		entities: make([]physicsEntity, 0),
		done:     make(chan struct{}),
	}
}

// Start the StorageParasiteSystem
func (s *StorageParasiteSystem) Start(dataChan chan RocketState) {
	s.dataChan = dataChan
	go s.processData()
}

// Stop the StorageParasiteSystem
func (s *StorageParasiteSystem) Stop() {
	close(s.done)
}

// processData logs rocket state data
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

// Priority returns the system priority
func (s *StorageParasiteSystem) Priority() int {
	return 1
}

// Update the StorageParasiteSystem
func (s *StorageParasiteSystem) Update(dt float32) error {
	// No need to track time here - data comes from simulation state
	return nil
}

// Add adds entities to the system
func (s *StorageParasiteSystem) Add(se *SystemEntity) {
	s.entities = append(s.entities, physicsEntity{se.Entity, se.Pos, se.Vel, se.Acc, se.Mass, se.Motor, se.Bodytube, se.Nosecone, se.Finset})
}
