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

// Use object pools for vectors and matrices
var (
	vectorPool = sync.Pool{
		New: func() interface{} {
			return &types.Vector3{}
		},
	}
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
		if entity.Entity.ID() == basic.ID() {
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

// Update applies forces to entities
func (s *PhysicsSystem) Update(dt float64) error {
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
					s.logger.Warn("Skipping gravity calculation due to invalid mass", "entity_id", entity.Entity.ID())
				}

				// Calculate Thrust Force (acts along rocket body axis, rotated to global frame)
				var thrustForce types.Vector3
				if entity.Motor != nil && !entity.Motor.IsCoasting() && entity.Orientation != nil && entity.Orientation.Quat != (types.Quaternion{}) {
					thrustMagnitude := entity.Motor.GetThrust()
					// Assume thrust acts along the rocket's local +Y axis
					localThrust := types.Vector3{Y: thrustMagnitude}
					thrustForce = *entity.Orientation.Quat.RotateVector(&localThrust)
					entity.AccumulatedForce = entity.AccumulatedForce.Add(thrustForce)
				} else if entity.Motor != nil && !entity.Motor.IsCoasting() {
					// Fallback if orientation is missing/invalid? Assume vertical thrust? Log warning?
					s.logger.Warn("Calculating thrust without valid orientation, assuming vertical", "entity_id", entity.Entity.ID())
					thrustMagnitude := entity.Motor.GetThrust()
					thrustForce = types.Vector3{Y: thrustMagnitude} // Vertical thrust
					entity.AccumulatedForce = entity.AccumulatedForce.Add(thrustForce)
				}

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
	entityID := uint64(0) // Default ID if entity or entity.Entity is nil
	if entity != nil && entity.Entity != nil {
		entityID = entity.Entity.ID()
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
