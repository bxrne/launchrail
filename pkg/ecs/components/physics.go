package components

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/types"
)

// Physics represents the physics component of an entity.
type Physics struct {
	Mass         float64
	Position     types.Vector3
	Velocity     types.Vector3
	Acceleration types.Vector3
	Forces       []types.Vector3
	Gravity      types.Vector3
}

// NewPhysics creates a new physics component.
func NewPhysics(gravityAcceleration, mass float64) *Physics {
	if mass <= 0 {
		panic("Mass must be greater than zero to prevent division by zero errors.")
	}

	return &Physics{
		Mass:     mass,
		Position: types.Vector3{},
		Velocity: types.Vector3{},
		Gravity:  types.Vector3{X: 0, Y: -gravityAcceleration, Z: 0},
	}
}

// AddForce applies a force to the physics component.
func (p *Physics) AddForce(force types.Vector3) {
	p.Forces = append(p.Forces, force)
}

// Update updates the physics state based on forces and time.
func (p *Physics) Update(dt float64) error {
	if dt <= 0 {
		return fmt.Errorf("invalid timestep: dt must be > 0")
	}

	// Calculate the net force acting on the entity
	netForce := p.Gravity
	for _, force := range p.Forces {
		netForce = netForce.Add(force)
	}

	// Calculate acceleration (mass assumed to be non-zero)
	p.Acceleration = netForce.DivideScalar(p.Mass)

	// Update velocity and position
	p.Velocity = p.Velocity.Add(p.Acceleration.MultiplyScalar(dt))
	p.Position = p.Position.Add(p.Velocity.MultiplyScalar(dt))

	// Clear the forces after applying them
	p.Forces = []types.Vector3{}
	return nil
}

// String returns a string representation of the Physics component.
func (p *Physics) String() string {
	return fmt.Sprintf(
		"Physics{Mass: %.2f, Position: %s, Velocity: %s, Acceleration: %s, Gravity: %s}",
		p.Mass, p.Position.String(), p.Velocity.String(), p.Acceleration.String(), p.Gravity.String(),
	)
}

// Type returns the component type
func (p *Physics) Type() string {
	return ecs.ComponentPhysics
}

// GetMass returns the mass of the entity
func (p *Physics) GetMass() float64 {
	return p.Mass
}

// GetVelocity returns the velocity of the entity
func (p *Physics) GetVelocity() types.Vector3 {
	return p.Velocity
}
