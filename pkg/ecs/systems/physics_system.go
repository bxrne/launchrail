package systems

import (
	"sync"

	"github.com/bxrne/launchrail/pkg/ecs"
)

type PhysicsSystem struct {
	workers int
	wg      sync.WaitGroup
}

func NewPhysicsSystem(workers int) *PhysicsSystem {
	if workers <= 0 {
		workers = 1
	}
	return &PhysicsSystem{workers: workers}
}

func (s *PhysicsSystem) Priority() int {
	return 100
}

func (s *PhysicsSystem) Update(world *ecs.World, dt float64) error {
	entities := world.Query(ecs.ComponentPhysics)

	// Create a channel to collect errors from goroutines
	errChan := make(chan error, s.workers)

	// Distribute work across workers
	chunkSize := (len(entities) + s.workers - 1) / s.workers
	s.wg.Add(s.workers)

	for w := 0; w < s.workers; w++ {
		go s.processEntityChunk(world, entities, w, chunkSize, dt, errChan)
	}

	// Wait for all workers to complete
	s.wg.Wait()
	close(errChan)

	// Collect first error if any
	return s.collectFirstError(errChan)
}

// Process a chunk of entities for a specific worker
func (s *PhysicsSystem) processEntityChunk(
	world *ecs.World,
	entities []ecs.EntityID,
	workerIndex,
	chunkSize int,
	dt float64,
	errChan chan<- error,
) {
	defer s.wg.Done()

	start := workerIndex * chunkSize
	end := start + chunkSize
	if end > len(entities) {
		end = len(entities)
	}

	for i := start; i < end; i++ {
		if err := s.updateEntityPhysics(world, entities[i], dt); err != nil {
			errChan <- err
			return
		}
	}
}

// Update physics for a single entity
func (s *PhysicsSystem) updateEntityPhysics(
	world *ecs.World,
	entity ecs.EntityID,
	dt float64,
) error {
	comp, ok := world.GetComponent(entity, ecs.ComponentPhysics)
	if !ok {
		return nil
	}

	physics, ok := comp.(ecs.PhysicsComponent)
	if !ok {
		return nil
	}

	return physics.Update(dt)
}

// Collect the first error from error channel
func (s *PhysicsSystem) collectFirstError(errChan <-chan error) error {
	for err := range errChan {
		return err
	}
	return nil
}
