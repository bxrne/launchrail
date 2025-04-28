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
	storage   storage.StorageInterface
	entities  []*states.PhysicsState // Change to pointer slice
	dataChan  chan *states.PhysicsState
	done      chan struct{}
	storeType storage.SimStorageType
}

// NewStorageParasiteSystem creates a new StorageParasiteSystem
func NewStorageParasiteSystem(world *ecs.World, store storage.StorageInterface, storeType storage.SimStorageType) *StorageParasiteSystem {
	return &StorageParasiteSystem{
		world:     world,
		storage:   store,
		entities:  make([]*states.PhysicsState, 0),
		done:      make(chan struct{}),
		storeType: storeType,
	}
}

// Start the StorageParasiteSystem
func (s *StorageParasiteSystem) Start(dataChan chan *states.PhysicsState) error {
	s.dataChan = dataChan

	err := s.storage.Init()
	if err != nil {
		return err
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
			if state == nil {
				continue
			}
			switch s.storeType {
			case storage.MOTION:
				s.handleMotion(state)
			case storage.EVENTS:
				s.handleEvents(state)
			case storage.DYNAMICS:
				s.handleDynamics(state)
			}
		case <-s.done:
			return
		}
	}
}

func (s *StorageParasiteSystem) handleMotion(state *states.PhysicsState) {
	if state.Motor == nil || state.Position == nil || state.Velocity == nil || state.Acceleration == nil {
		return
	}
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
}

func (s *StorageParasiteSystem) handleEvents(state *states.PhysicsState) {
	if state.Motor == nil || state.Parachute == nil {
		return
	}
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
}

func (s *StorageParasiteSystem) handleDynamics(state *states.PhysicsState) {
	if state.Position == nil || state.Velocity == nil || state.Acceleration == nil || state.Orientation == nil {
		return
	}
	record := []string{
		fmt.Sprintf("%.6f", state.Time),
		fmt.Sprintf("%.6f", state.Position.Vec.X),
		fmt.Sprintf("%.6f", state.Position.Vec.Y),
		fmt.Sprintf("%.6f", state.Position.Vec.Z),
		fmt.Sprintf("%.6f", state.Velocity.Vec.X),
		fmt.Sprintf("%.6f", state.Velocity.Vec.Y),
		fmt.Sprintf("%.6f", state.Velocity.Vec.Z),
		fmt.Sprintf("%.6f", state.Acceleration.Vec.X),
		fmt.Sprintf("%.6f", state.Acceleration.Vec.Y),
		fmt.Sprintf("%.6f", state.Acceleration.Vec.Z),
		fmt.Sprintf("%.6f", state.Orientation.Quat.X),
		fmt.Sprintf("%.6f", state.Orientation.Quat.Y),
		fmt.Sprintf("%.6f", state.Orientation.Quat.Z),
		fmt.Sprintf("%.6f", state.Orientation.Quat.W),
	}
	if err := s.storage.Write(record); err != nil {
		fmt.Printf("Error writing dynamics record: %v\n", err)
	}
}

// Update the StorageParasiteSystem
func (s *StorageParasiteSystem) Update(dt float64) error {
	// No need to track time here - data comes from simulation state
	return nil
}

// Add adds entities to the system
func (s *StorageParasiteSystem) Add(pe *states.PhysicsState) {
	s.entities = append(s.entities, pe) // Store pointer directly
}
