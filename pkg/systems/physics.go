package systems

import (
	"math"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/barrowman"
	"github.com/bxrne/launchrail/pkg/types"
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
	entities        []*PhysicsEntity // Changed to store pointers
	cpCalculator    *barrowman.CPCalculator
	workers         int
	workChan        chan PhysicsEntity
	resultChan      chan types.Vector3
	gravity         float64
	groundTolerance float64
}

// calculateStabilityForces calculates stability forces for an entity
func (s *PhysicsSystem) calculateStabilityForces(force *types.Vector3, stabilityMargin float64, entity PhysicsEntity) {
	// Basic stability force calculation
	const stabilityFactor = 0.1
	_ = entity

	// Apply corrective force based on stability margin
	correctionForce := stabilityMargin * stabilityFactor
	force.Y += correctionForce
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
func NewPhysicsSystem(world *ecs.World, cfg *config.Config) *PhysicsSystem {
	workers := 4
	return &PhysicsSystem{
		world:           world,
		entities:        make([]*PhysicsEntity, 0),
		workers:         workers,
		workChan:        make(chan PhysicsEntity, workers),
		resultChan:      make(chan types.Vector3, workers),
		cpCalculator:    barrowman.NewCPCalculator(), // Initialize calculator
		gravity:         cfg.Options.Launchsite.Atmosphere.ISAConfiguration.GravitationalAccel,
		groundTolerance: cfg.Simulation.GroundTolerance,
	}
}

// Update applies forces to entities
func (s *PhysicsSystem) Update(dt float64) error {
	var wg sync.WaitGroup
	workChan := make(chan *PhysicsEntity, len(s.entities)) // Changed to pointer channel
	resultChan := make(chan types.Vector3, len(s.entities))

	// Launch multiple workers for concurrent force calculations
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entity := range workChan {
				force := vectorPool.Get().(*types.Vector3)
				*force = types.Vector3{} // reset force
				// ...existing or refined stability/forces code...
				s.calculateStabilityForces(force, 0.0, *entity) // Dereference for backwards compatibility
				resultChan <- *force
				vectorPool.Put(force)
			}
		}()
	}

	for _, entity := range s.entities {
		workChan <- entity
	}
	close(workChan)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	i := 0
	for force := range resultChan {
		s.applyForce(s.entities[i], force, dt)
		i++
	}
	return nil
}

func (s *PhysicsSystem) handleGroundCollision(entity *PhysicsEntity) bool {
	groundTolerance := s.groundTolerance
	if entity.Position.Vec.Y <= groundTolerance {
		if entity.Position.Vec.Y <= 0 {
			entity.Position.Vec.Y = 0
			entity.Velocity.Vec.Y = 0
			entity.Acceleration.Vec.Y = 0
			return true
		}
	}
	return false
}

func (s *PhysicsSystem) calculateNetForce(entity *PhysicsEntity, force types.Vector3) float64 {
	var netForce float64

	// Add thrust if motor is active
	if entity.Motor != nil && !entity.Motor.IsCoasting() {
		thrust := entity.Motor.GetThrust()
		if !math.IsNaN(thrust) {
			netForce += thrust
		}
	}

	// Calculate velocity magnitude for drag
	velocity := math.Sqrt(entity.Velocity.Vec.X*entity.Velocity.Vec.X + entity.Velocity.Vec.Y*entity.Velocity.Vec.Y)

	if velocity > 0 {
		rho := getAtmosphericDensity(entity.Position.Vec.Y)
		if math.IsNaN(rho) {
			rho = 1.225 // Use sea level density as fallback
		}

		area := calculateReferenceArea(entity.Nosecone, entity.Bodytube)
		cd := 0.3 // Base drag coefficient
		if velocity > 100 {
			cd = 0.5 // Increased drag at higher velocities
		}

		dragForce := 0.5 * rho * cd * area * velocity * velocity

		// Apply drag in opposite direction of velocity
		if entity.Velocity.Vec.Y > 0 {
			netForce -= dragForce
		} else {
			netForce += dragForce
		}

		// Add external force
		netForce += force.Y
	}

	return netForce
}

func (s *PhysicsSystem) updateEntityState(entity *PhysicsEntity, netForce float64, dt float64) {
	entity.Acceleration.Vec.Y += netForce / entity.Mass.Value

	// Semi-implicit Euler integration
	newVelocity := entity.Velocity.Vec.Y + entity.Acceleration.Vec.Y*dt
	newPosition := entity.Position.Vec.Y + newVelocity*dt

	if newPosition <= 0 {
		s.handleGroundCollision(entity)
		return
	}

	entity.Velocity.Vec.Y = newVelocity
	entity.Position.Vec.Y = newPosition
}

func (s *PhysicsSystem) applyForce(entity *PhysicsEntity, force types.Vector3, dt float64) {
	// Add nil checks for required components
	if entity.Bodytube == nil || entity.Nosecone == nil || entity.Mass == nil {
		return
	}

	// Validate timestep and mass
	dt64 := float64(dt)
	if dt64 <= 0 || math.IsNaN(dt64) || dt64 > 0.1 || entity.Mass.Value <= 0 {
		return
	}

	// Check current state for landing condition
	if s.handleGroundCollision(entity) {
		return
	}

	// Reset acceleration and apply gravity
	entity.Acceleration.Vec.X = 0
	entity.Acceleration.Vec.Y = -s.gravity

	// Calculate and apply forces
	netForce := s.calculateNetForce(entity, force)
	s.updateEntityState(entity, netForce, dt64)
}

// Add adds an entity to the system
func (s *PhysicsSystem) Add(pe *PhysicsEntity) {
	s.entities = append(s.entities, pe) // Store pointer directly
}

// Priority returns the system priority
func (s *PhysicsSystem) Priority() int {
	return 1 // Runs after forces are applied
}

// String returns the system name
func (s *PhysicsSystem) String() string {
	return "PhysicsSystem"
}
