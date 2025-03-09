package systems

import (
	"fmt"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/states"
)

// StorageParasiteSystem logs rocket state data to storage
type StorageParasiteSystem struct {
	world     *ecs.World
	storage   *storage.Storage
	entities  []states.PhysicsState
	dataChan  chan *states.PhysicsState
	done      chan struct{}
	storeType storage.StorageType
}

// NewStorageParasiteSystem creates a new StorageParasiteSystem
func NewStorageParasiteSystem(world *ecs.World, storage *storage.Storage, storeType storage.StorageType) *StorageParasiteSystem {
	return &StorageParasiteSystem{
		world:     world,
		storage:   storage,
		entities:  make([]states.PhysicsState, 0),
		done:      make(chan struct{}),
		storeType: storeType,
	}
}

// Start the StorageParasiteSystem
func (s *StorageParasiteSystem) Start(dataChan chan *states.PhysicsState) error {
	s.dataChan = dataChan

	s.storage.Init()

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
			switch s.storeType {
			case storage.MOTION:
				record := []string{
					fmt.Sprintf("%.6f", state.Time),
					fmt.Sprintf("%.6f", state.Position.Vec.Y),
					fmt.Sprintf("%.6f", state.Velocity.Vec.Y),
					fmt.Sprintf("%.6f", state.Acceleration.Vec.Y),
					fmt.Sprintf("%.6f", state.Motor.GetThrust()),
				}
				if err := s.storage.Write(record); err != nil {
					fmt.Printf("Error writing motion record: %v\n", err)
				}
			case storage.EVENTS:
				parachuteStatus := "NOT_DEPLOYED"
				if state.Parachute.IsDeployed() {
					parachuteStatus = "DEPLOYED"
				}

				record := []string{
					fmt.Sprintf("%.6f", state.Time),
					state.Motor.GetState(),
					parachuteStatus,
				}
				if err := s.storage.Write(record); err != nil {
					fmt.Printf("Error writing event record: %v\n", err)
				}

			case storage.DYNAMICS:
				record := []string{
					fmt.Sprintf("%.6f", state.Time),
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
					"0", // TODO: Add dynamics data
				}
				if err := s.storage.Write(record); err != nil {
					fmt.Printf("Error writing dynamics record: %v\n", err)
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
func (s *StorageParasiteSystem) Add(pe *states.PhysicsState) {
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
