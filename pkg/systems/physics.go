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

	return s.processEntities(dt)
}

// Result represents the outcome of a physics calculation
type Result struct {
	err error
}

// processEntities handles the concurrent processing of all physics entities
func (s *PhysicsSystem) processEntities(dt float64) error {
	workChan := make(chan *states.PhysicsState)
	results := make(chan Result, len(s.entities))

	// Start the worker pool
	s.startWorkerPool(workChan, results)

	// Send all entities to workChan
	for _, entity := range s.entities {
		workChan <- entity
	}
	close(workChan)

	// Collect results and return the first error encountered
	return s.collectResults(results)
}

// startWorkerPool initializes worker goroutines to process physics updates
func (s *PhysicsSystem) startWorkerPool(workChan <-chan *states.PhysicsState, results chan<- Result) {
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.processEntityWorker(workChan, results)
		}()
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()
}

// processEntityWorker handles physics updates for entities received on the workChan
func (s *PhysicsSystem) processEntityWorker(workChan <-chan *states.PhysicsState, results chan<- Result) {
	for entity := range workChan {
		if err := s.validateEntity(entity); err != nil {
			results <- Result{err: err}
			continue
		}

		// Apply physics forces to the entity
		s.applyPhysicsForces(entity)

		results <- Result{err: nil} // Report success
	}
}

// applyPhysicsForces applies physics forces to an entity
func (s *PhysicsSystem) applyPhysicsForces(entity *states.PhysicsState) {
	// Calculate Gravity Force (acts downwards in global frame)
	if entity.Mass != nil && entity.Mass.Value > 0 {
		gravityForce := types.Vector3{Y: -s.gravity * entity.Mass.Value}
		entity.AccumulatedForce = entity.AccumulatedForce.Add(gravityForce)
	} else {
		s.logger.Warn("Skipping gravity calculation due to invalid mass", "entity_id", entity.BasicEntity.ID())
	}

	// NOTE: Thrust force is now handled in simulation.go's RK4 integrator
	// to avoid double-counting thrust forces
}

// collectResults collects processing results and returns the first error encountered
func (s *PhysicsSystem) collectResults(results <-chan Result) error {
	for res := range results {
		if res.err != nil {
			return res.err
		}
	}
	return nil
}

// validateEntity checks if an entity has all required components and properties
func (s *PhysicsSystem) validateEntity(entity *states.PhysicsState) error {
	// Check if entity exists
	if entity == nil {
		return fmt.Errorf("nil entity")
	}

	// Get entity ID for error messages
	entityID := entity.BasicEntity.ID()

	// Validate required vectors
	if err := s.validateRequiredVectors(entity); err != nil {
		return err
	}

	// Validate mass
	if err := s.validateMass(entity); err != nil {
		return err
	}

	// Validate geometry components
	if entity.Nosecone == nil {
		return fmt.Errorf("entity %d missing Nosecone component", entityID)
	}
	if entity.Bodytube == nil {
		return fmt.Errorf("entity %d missing Bodytube component", entityID)
	}

	return nil
}

// validateRequiredVectors checks if an entity has all required vector components
func (s *PhysicsSystem) validateRequiredVectors(entity *states.PhysicsState) error {
	if entity.Position == nil || entity.Velocity == nil || entity.Acceleration == nil {
		return fmt.Errorf("entity missing required vectors")
	}
	return nil
}

// validateMass checks if an entity has a valid mass component
func (s *PhysicsSystem) validateMass(entity *states.PhysicsState) error {
	if entity.Mass == nil {
		return fmt.Errorf("entity missing mass")
	}
	if entity.Mass.Value <= 0 {
		return fmt.Errorf("invalid entity mass value: %v", entity.Mass.Value)
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
