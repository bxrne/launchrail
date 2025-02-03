package systems

import (
	"math"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/barrowman"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/types"
)

// Use object pools for vectors and matrices
var (
	vectorPool = sync.Pool{
		New: func() interface{} {
			return &types.Vector3{}
		},
	}

	forcePool = sync.Pool{
		New: func() interface{} {
			return make([]types.Vector3, 0, 10)
		},
	}
)

type PhysicsSystem struct {
	world        *ecs.World
	entities     []physicsEntity
	cpCalculator *barrowman.CPCalculator
	workers      int
	workChan     chan physicsEntity
	resultChan   chan types.Vector3
}

// Add TrapezoidFinset to physicsEntity
type physicsEntity struct {
	*ecs.BasicEntity
	*components.Position
	*components.Velocity
	*components.Acceleration
	*components.Mass
	*components.Motor
	*components.Bodytube
	*components.Nosecone
	*components.TrapezoidFinset // Add this field
}

func (s *PhysicsSystem) calculateCG(entity physicsEntity) float64 {
	// Simple CG calculation - assuming uniform mass distribution
	totalMass := entity.Mass.Value
	totalMoment := 0.0

	// Add moments from each component
	if entity.Nosecone != nil {
		totalMoment += entity.Nosecone.GetMass() * entity.Nosecone.Position.X
	}
	if entity.Bodytube != nil {
		totalMoment += entity.Bodytube.GetMass() * entity.Bodytube.Position.X
	}
	if entity.Motor != nil {
		totalMoment += entity.Motor.GetMass() * entity.Motor.Position.X
	}

	return totalMoment / totalMass
}

func (s *PhysicsSystem) calculateStabilityForces(force *types.Vector3, stabilityMargin float64, entity physicsEntity) {
	// Basic stability force calculation
	const stabilityFactor = 0.1

	// Apply corrective force based on stability margin
	correctionForce := stabilityMargin * stabilityFactor
	force.Y += correctionForce
}

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

func NewPhysicsSystem(world *ecs.World) *PhysicsSystem {
	workers := 4
	return &PhysicsSystem{
		world:        world,
		entities:     make([]physicsEntity, 0),
		workers:      workers,
		workChan:     make(chan physicsEntity, workers),
		resultChan:   make(chan types.Vector3, workers),
		cpCalculator: barrowman.NewCPCalculator(), // Initialize calculator
	}
}

func (s *PhysicsSystem) Update(dt float32) {
	var wg sync.WaitGroup
	workChan := make(chan physicsEntity, len(s.entities))
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
}

func (s *PhysicsSystem) processForces(wg *sync.WaitGroup) {
	defer wg.Done()
	for entity := range s.workChan {
		// Get vector from pool
		force := vectorPool.Get().(*types.Vector3)
		defer vectorPool.Put(force)

		// Only calculate CP if we have a finset
		var stabilityMargin float64
		if entity.TrapezoidFinset != nil {
			// Calculate CP and CG
			cp := s.cpCalculator.CalculateCP(entity.Nosecone, entity.Bodytube, entity.TrapezoidFinset)
			cg := s.calculateCG(entity)
			stabilityMargin = (cp - cg) / entity.Bodytube.Radius
		}

		// Calculate forces based on stability
		s.calculateStabilityForces(force, stabilityMargin, entity)

		s.resultChan <- *force
	}
}

func (s *PhysicsSystem) applyForce(entity physicsEntity, force types.Vector3, dt float32) {
	// Validate timestep
	dt64 := float64(dt)
	if dt64 <= 0 || math.IsNaN(dt64) || dt64 > 0.1 {
		return
	}

	// Reset acceleration each frame
	entity.Acceleration.X = 0
	entity.Acceleration.Y = -9.81 // Gravity is constant

	// Calculate forces
	var netForce float64

	// Add thrust in POSITIVE Y direction (upward)
	if !entity.Motor.IsCoasting() {
		thrust := entity.Motor.GetThrust()
		netForce += thrust // Add thrust as positive force
	}

	// Add weight (already accounted for in initial acceleration)
	// weight := entity.Mass.Value * 9.81
	// netForce -= weight   // REMOVE THIS - we already have gravity in acceleration

	// Calculate drag
	velocity := math.Sqrt(entity.Velocity.X*entity.Velocity.X +
		entity.Velocity.Y*entity.Velocity.Y)
	if velocity > 0 {
		// Air density decreases with altitude
		rho := 1.225 * math.Exp(-entity.Position.Y/7400.0)

		// Reference area
		area := math.Pi * entity.Bodytube.Radius * entity.Bodytube.Radius

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
	}

	// Calculate final acceleration (adding to gravity)
	entity.Acceleration.Y += netForce / entity.Mass.Value // ADD to existing acceleration

	// Semi-implicit Euler integration
	newVelocity := entity.Velocity.Y + entity.Acceleration.Y*dt64
	newPosition := entity.Position.Y + newVelocity*dt64

	// Ground collision check
	if newPosition <= 0 && newVelocity < 0 {
		// entity.Position.Y = 0
		// entity.Velocity.Y = 0
		// entity.Acceleration.Y = 0
		// panic("Simulation ended: Ground impact")
	}

	// Update state
	entity.Velocity.Y = newVelocity
	entity.Position.Y = newPosition
}

func (s *PhysicsSystem) Add(entity *ecs.BasicEntity, pos *components.Position,
	vel *components.Velocity, acc *components.Acceleration, mass *components.Mass, motor *components.Motor, bodytube *components.Bodytube, nosecone *components.Nosecone) {
	s.entities = append(s.entities, physicsEntity{entity, pos, vel, acc, mass, motor, bodytube, nosecone, nil})
}

func (s *PhysicsSystem) Priority() int {
	return 1 // Runs after forces are applied
}
func (s *PhysicsSystem) String() string {
	return "PhysicsSystem"
}
