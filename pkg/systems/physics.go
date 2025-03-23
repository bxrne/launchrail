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
		// Update motor state first
		if entity.Motor != nil {
			if err := entity.Motor.Update(dt); err != nil {
				return err
			}
		}

		// Calculate forces
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
	if entity.Position.Vec.Y <= s.groundTolerance && entity.Velocity.Vec.Y <= 0 {
		entity.Position.Vec.Y = 0
		entity.Velocity.Vec.Y = 0
		entity.Acceleration.Vec.Y = 0
		return true
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

	// Update translational state
	entity.Acceleration.Vec.Y = newAcceleration
	newVelocity := entity.Velocity.Vec.Y + entity.Acceleration.Vec.Y*dt
	newPosition := entity.Position.Vec.Y + newVelocity*dt
	if math.IsNaN(newVelocity) || math.IsInf(newVelocity, 0) {
		return
	}

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

	// Update rotation (6DOF)
	if entity.Orientation != nil && entity.AngularVelocity != nil && entity.AngularAcceleration != nil {
		// Simple angular integration
		entity.AngularVelocity.X += entity.AngularAcceleration.X * dt
		entity.AngularVelocity.Y += entity.AngularAcceleration.Y * dt
		entity.AngularVelocity.Z += entity.AngularAcceleration.Z * dt

		// Integrate orientation quaternion
		entity.Orientation.Integrate(*entity.AngularVelocity, dt)
	}
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
