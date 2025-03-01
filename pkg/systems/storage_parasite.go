package systems

import (
	"fmt"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/storage"
)

// StorageParasiteSystem logs rocket state data to storage
type StorageParasiteSystem struct {
	world     *ecs.World
	storage   *storage.Storage
	entities  []PhysicsEntity
	dataChan  chan RocketState
	done      chan struct{}
	storeType storage.StorageType
}

// NewStorageParasiteSystem creates a new StorageParasiteSystem
func NewStorageParasiteSystem(world *ecs.World, storage *storage.Storage, storeType storage.StorageType) *StorageParasiteSystem {
	return &StorageParasiteSystem{
		world:     world,
		storage:   storage,
		entities:  make([]PhysicsEntity, 0),
		done:      make(chan struct{}),
		storeType: storeType,
	}
}

// Start the StorageParasiteSystem
func (s *StorageParasiteSystem) Start(dataChan chan RocketState) error {
	s.dataChan = dataChan

	// Initialize headers based on store type
	switch s.storeType {
	case storage.MOTION:
		err := s.storage.Init([]string{"Time", "Altitude", "Velocity", "Acceleration", "Thrust"})
		if err != nil {
			return fmt.Errorf("error initializing storage: %v", err)
		}
	case storage.EVENTS:
		err := s.storage.Init([]string{"time", "motor_status", "parachute_status"})
		if err != nil {
			return fmt.Errorf("error initializing storage: %v", err)
		}
	}

	go s.processData()

	return nil
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
			if s.storeType == storage.MOTION {
				record := []string{
					fmt.Sprintf("%.6f", state.Time),
					fmt.Sprintf("%.6f", state.Altitude),
					fmt.Sprintf("%.6f", state.Velocity),
					fmt.Sprintf("%.6f", state.Acceleration),
					fmt.Sprintf("%.6f", state.Thrust),
				}
				if err := s.storage.Write(record); err != nil {
					fmt.Printf("Error writing motion record: %v\n", err)
				}
			} else if s.storeType == storage.EVENTS {
				parachuteStatus := "NOT_DEPLOYED"
				if state.ParachuteDeployed {
					parachuteStatus = "DEPLOYED"
				}

				record := []string{
					fmt.Sprintf("%.6f", state.Time),
					state.MotorState,
					parachuteStatus,
				}
				if err := s.storage.Write(record); err != nil {
					fmt.Printf("Error writing event record: %v\n", err)
				}
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
func (s *StorageParasiteSystem) Update(dt float64) error {
	// No need to track time here - data comes from simulation state
	return nil
}

// Add adds entities to the system
func (s *StorageParasiteSystem) Add(pe *PhysicsEntity) {
	s.entities = append(s.entities, PhysicsEntity{pe.Entity, pe.Position, pe.Velocity, pe.Acceleration, pe.Mass, pe.Motor, pe.Bodytube, pe.Nosecone, pe.Finset, pe.Parachute})
}
