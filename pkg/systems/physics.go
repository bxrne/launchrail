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
	workChan        chan states.PhysicsState
	resultChan      chan types.Vector3
	gravity         float64
	groundTolerance float64
}

// calculateStabilityForces calculates stability forces for an entity
func (s *PhysicsSystem) calculateStabilityForces(force *types.Vector3, stabilityMargin float64, entity states.PhysicsState) {
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
		entities:        make([]*states.PhysicsState, 0),
		workers:         workers,
		workChan:        make(chan states.PhysicsState, workers),
		resultChan:      make(chan types.Vector3, workers),
		cpCalculator:    barrowman.NewCPCalculator(), // Initialize calculator
		gravity:         cfg.Options.Launchsite.Atmosphere.ISAConfiguration.GravitationalAccel,
		groundTolerance: cfg.Simulation.GroundTolerance,
	}
}

// Update applies forces to entities
func (s *PhysicsSystem) Update(dt float64) error {
	if dt <= 0 || math.IsNaN(dt) {
		return fmt.Errorf("invalid timestep: %v", dt)
	}

	for _, entity := range s.entities {
		// Skip invalid entities
		if entity == nil || entity.Mass == nil || entity.Mass.Value <= 0 {
			continue
		}

		// Calculate forces first
		netForce := s.calculateNetForce(entity, types.Vector3{})

		// Update state
		s.updateEntityState(entity, netForce, dt)

		// Validate new state
		if math.IsNaN(entity.Position.Vec.Y) ||
			math.IsNaN(entity.Velocity.Vec.Y) ||
			math.IsNaN(entity.Acceleration.Vec.Y) {
			return fmt.Errorf("NaN values in state update")
		}
	}

	return nil
}

func (s *PhysicsSystem) handleGroundCollision(entity *states.PhysicsState) bool {
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

func (s *PhysicsSystem) calculateNetForce(entity *states.PhysicsState, force types.Vector3) float64 {
	if entity == nil || entity.Mass == nil || entity.Mass.Value <= 0 {
		return 0
	}

	var netForce float64 = -s.gravity * entity.Mass.Value // Start with gravity

	// Add thrust if motor is active
	if entity.Motor != nil && !entity.Motor.IsCoasting() {
		thrust := entity.Motor.GetThrust()
		if !math.IsNaN(thrust) && !math.IsInf(thrust, 0) {
			netForce += thrust
		}
	}

	// Add drag force
	velocity := math.Sqrt(
		entity.Velocity.Vec.X*entity.Velocity.Vec.X +
			entity.Velocity.Vec.Y*entity.Velocity.Vec.Y)

	if velocity > 0 && !math.IsNaN(velocity) {
		rho := getAtmosphericDensity(entity.Position.Vec.Y)
		if !math.IsNaN(rho) && rho > 0 {
			area := calculateReferenceArea(entity.Nosecone, entity.Bodytube)
			cd := 0.3 // Base drag coefficient
			if velocity > 100 {
				cd = 0.5
			}
			dragForce := 0.5 * rho * cd * area * velocity * velocity
			netForce += -math.Copysign(dragForce, entity.Velocity.Vec.Y)
		}
	}

	// Add external force
	if !math.IsNaN(force.Y) && !math.IsInf(force.Y, 0) {
		netForce += force.Y
	}

	return netForce
}

func (s *PhysicsSystem) updateEntityState(entity *states.PhysicsState, netForce float64, dt float64) {
	if math.IsNaN(netForce) || math.IsInf(netForce, 0) {
		return // Skip update if force is invalid
	}

	// Calculate acceleration
	newAcceleration := netForce / entity.Mass.Value
	if math.IsNaN(newAcceleration) || math.IsInf(newAcceleration, 0) {
		return
	}
	entity.Acceleration.Vec.Y = newAcceleration

	// Semi-implicit Euler integration with validation
	newVelocity := entity.Velocity.Vec.Y + entity.Acceleration.Vec.Y*dt
	if math.IsNaN(newVelocity) || math.IsInf(newVelocity, 0) {
		return
	}

	newPosition := entity.Position.Vec.Y + newVelocity*dt
	if math.IsNaN(newPosition) || math.IsInf(newPosition, 0) {
		return
	}

	// Apply ground constraint
	if newPosition <= 0 {
		s.handleGroundCollision(entity)
		return
	}

	entity.Velocity.Vec.Y = newVelocity
	entity.Position.Vec.Y = newPosition
}

func (s *PhysicsSystem) applyForce(entity *states.PhysicsState, force types.Vector3, dt float64) {
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
func (s *PhysicsSystem) Add(pe *states.PhysicsState) {
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
