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
	world        *ecs.World
	entities     []PhysicsEntity
	cpCalculator *barrowman.CPCalculator
	workers      int
	workChan     chan PhysicsEntity
	resultChan   chan types.Vector3
	gravity      float64
}

// calculateStabilityForces calculates stability forces for an entity
func (s *PhysicsSystem) calculateStabilityForces(force *types.Vector3, stabilityMargin float64, entity PhysicsEntity) {
	// Basic stability force calculation
	const stabilityFactor = 0.1

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
		world:        world,
		entities:     make([]PhysicsEntity, 0),
		workers:      workers,
		workChan:     make(chan PhysicsEntity, workers),
		resultChan:   make(chan types.Vector3, workers),
		cpCalculator: barrowman.NewCPCalculator(), // Initialize calculator
		gravity:      cfg.Options.Launchsite.Atmosphere.ISAConfiguration.GravitationalAccel,
	}
}

// Update applies forces to entities
func (s *PhysicsSystem) Update(dt float32) error {
	var wg sync.WaitGroup
	workChan := make(chan PhysicsEntity, len(s.entities))
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
				s.calculateStabilityForces(force, 0.0, entity)
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

// applyForce applies forces to an entity
func (s *PhysicsSystem) applyForce(entity PhysicsEntity, force types.Vector3, dt float32) {
	// Validate timestep
	dt64 := float64(dt)
	if dt64 <= 0 || math.IsNaN(dt64) || dt64 > 0.1 {
		return
	}

	// Reset acceleration each frame
	entity.Acceleration.X = 0
	entity.Acceleration.Y = -s.gravity // Use configured gravity

	// Calculate forces
	var netForce float64
	// Add thrust in POSITIVE Y direction (upward)
	if !entity.Motor.IsCoasting() {
		thrust := entity.Motor.GetThrust()
		netForce += thrust // Add thrust as positive force
	}

	// Calculate drag
	velocity := math.Sqrt(entity.Velocity.X*entity.Velocity.X +
		entity.Velocity.Y*entity.Velocity.Y)
	if velocity > 0 {
		// Air density decreases with altitude
		rho := getAtmosphericDensity(entity.Position.Y)

		// Reference area
		area := calculateReferenceArea(entity.Nosecone, entity.Bodytube)

		// Drag coefficient increases with velocity
		cd := 0.3
		if velocity > 100 {
			cd = 0.5
		}

		// Calculate drag force
		dragForce := 0.5 * rho * cd * area * velocity * velocity

		// Apply drag opposite to velocity
		if entity.Velocity.Y > 0 {
			netForce -= dragForce
		} else {
			netForce += dragForce
		}

		// apply passed in force
		netForce += force.Y

	}

	// Calculate lift
	liftCoefficient := 0.5 // Simplified lift coefficient
	area := math.Pi * entity.Bodytube.Radius * entity.Bodytube.Radius
	rho := 1.225 * math.Exp(-entity.Position.Y/7400.0)
	liftForce := 0.5 * rho * liftCoefficient * area * velocity * velocity
	netForce += liftForce

	// Calculate final acceleration (adding to gravity)
	entity.Acceleration.Y += netForce / entity.Mass.Value // ADD to existing acceleration

	// Semi-implicit Euler integration
	newVelocity := entity.Velocity.Y + entity.Acceleration.Y*dt64
	newPosition := entity.Position.Y + newVelocity*dt64

	// Ground collision check with proper landing detection
	if newPosition <= 0 {
		if newVelocity < 0 {
			// Landing - stop movement
			entity.Position.Y = 0
			entity.Velocity.Y = 0
			entity.Acceleration.Y = 0
		}
		return // Stop updating after landing
	}

	// Update state if not landed
	entity.Velocity.Y = newVelocity
	entity.Position.Y = newPosition
}

// Add adds an entity to the system
func (s *PhysicsSystem) Add(pe *PhysicsEntity) {
	s.entities = append(s.entities, PhysicsEntity{pe.Entity, pe.Position, pe.Velocity, pe.Acceleration, pe.Mass, pe.Motor, pe.Bodytube, pe.Nosecone, pe.Finset})
}

// Priority returns the system priority
func (s *PhysicsSystem) Priority() int {
	return 1 // Runs after forces are applied
}

// String returns the system name
func (s *PhysicsSystem) String() string {
	return "PhysicsSystem"
}
