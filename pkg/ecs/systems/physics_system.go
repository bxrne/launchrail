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

	// Process physics in parallel
	chunkSize := (len(entities) + s.workers - 1) / s.workers
	s.wg.Add(s.workers)

	for w := 0; w < s.workers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > len(entities) {
			end = len(entities)
		}

		go func(start, end int) {
			defer s.wg.Done()
			for i := start; i < end; i++ {
				entity := entities[i]
				if comp, ok := world.GetComponent(entity, ecs.ComponentPhysics); ok {
					if physics, ok := comp.(ecs.PhysicsComponent); ok {
						physics.Update(dt)
				}
				}
			}
		}(start, end)
	}

	s.wg.Wait()
	return nil
}
