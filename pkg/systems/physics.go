package systems

import (
	"fmt"
	"math"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/barrowman"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

// PhysicsSystem calculates forces on entities
type PhysicsSystem struct {
	world           *ecs.World
	entities        []*states.PhysicsState // Changed to store pointers
	cpCalculator    *barrowman.CPCalculator
	workers         int
	gravity         float64
	groundTolerance float64
	logger          logf.Logger // Use non-pointer interface type
}

// Remove removes an entity from the system
func (s *PhysicsSystem) Remove(basic ecs.BasicEntity) {
	var deleteIndex int
	for i, entity := range s.entities {
		if entity.BasicEntity.ID() == basic.ID() {
			deleteIndex = i
			break
		}
	}
	s.entities = append(s.entities[:deleteIndex], s.entities[deleteIndex+1:]...)
}

// NewPhysicsSystem creates a new PhysicsSystem
func NewPhysicsSystem(world *ecs.World, cfg *config.Engine, logger logf.Logger, workers int) *PhysicsSystem {
	if workers <= 0 {
		workers = 4 // Default number of workers
	}
	return &PhysicsSystem{
		world:           world,
		entities:        make([]*states.PhysicsState, 0),
		cpCalculator:    barrowman.NewCPCalculator(), // Initialize calculator
		workers:         workers,
		gravity:         cfg.Options.Launchsite.Atmosphere.ISAConfiguration.GravitationalAccel,
		groundTolerance: cfg.Simulation.GroundTolerance,
		logger:          logger, // Use logger directly
	}
}

// Update implements ecs.System interface
func (s *PhysicsSystem) Update(dt float32) {
	_ = s.update(float64(dt))
}

// UpdateWithError implements System interface
func (s *PhysicsSystem) UpdateWithError(dt float64) error {
	return s.update(dt)
}

// update is the internal update method
func (s *PhysicsSystem) update(dt float64) error {
	if dt <= 0 || math.IsNaN(dt) {
		return fmt.Errorf("invalid timestep: %v", dt)
	}

	type result struct {
		err error
	}

	workChan := make(chan *states.PhysicsState)
	results := make(chan result, len(s.entities))
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entity := range workChan {
				if err := s.validateEntity(entity); err != nil {
					results <- result{err: err}
					continue
				}

				// --- Force Accumulation and State Update ---
				// Reset forces for this physics update step
				// entity.AccumulatedForce = types.Vector3{} // THIS LINE IS REMOVED - AccumulatedForce is managed by simulation.go

				// Calculate Gravity Force (acts downwards in global frame)
				if entity.Mass != nil && entity.Mass.Value > 0 { // Only apply gravity if mass is valid
					gravityForce := types.Vector3{Y: -s.gravity * entity.Mass.Value}
					entity.AccumulatedForce = entity.AccumulatedForce.Add(gravityForce)
				} else {
					s.logger.Warn("Skipping gravity calculation due to invalid mass", "entity_id", entity.BasicEntity.ID())
				}

				// NOTE: Thrust force is now handled in simulation.go's RK4 integrator
				// to avoid double-counting thrust forces

				// *** Apply accumulated forces and update state *** // THIS COMMENT IS MISLEADING NOW
				// s.updateEntityState(entity, entity.AccumulatedForce, dt) // THIS CALL IS REMOVED

				results <- result{err: nil} // Report success for this entity
			}
		}()
	}

	// Send all entities to workChan
	for _, entity := range s.entities {
		workChan <- entity
	}
	close(workChan)

	// Wait for all workers to finish in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect the first error encountered
	for res := range results {
		if res.err != nil {
			return res.err
		}
	}

	return nil
}

func (s *PhysicsSystem) validateEntity(entity *states.PhysicsState) error {
	entityID := uint64(0) // Default ID if entity is nil
	if entity != nil {
		entityID = entity.BasicEntity.ID()
	}
	if entity == nil {
		return fmt.Errorf("nil entity")
	}
	if entity.Position == nil || entity.Velocity == nil || entity.Acceleration == nil {
		return fmt.Errorf("entity missing required vectors")
	}
	if entity.Mass == nil {
		return fmt.Errorf("entity missing mass")
	}
	if entity.Mass.Value <= 0 {
		return fmt.Errorf("invalid entity mass value: %v", entity.Mass.Value)
	}
	// Check for essential geometry needed for drag/aerodynamics
	if entity.Nosecone == nil {
		return fmt.Errorf("entity %d missing Nosecone component", entityID)
	}
	if entity.Bodytube == nil {
		return fmt.Errorf("entity %d missing Bodytube component", entityID)
	}
	return nil
}

// Add adds an entity to the system
func (s *PhysicsSystem) Add(pe *states.PhysicsState) {
	s.entities = append(s.entities, pe) // Store pointer directly
}

// String returns the system name
func (s *PhysicsSystem) String() string {
	return "PhysicsSystem"
}
