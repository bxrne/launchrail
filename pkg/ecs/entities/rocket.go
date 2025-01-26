package entities

import (
	"fmt"
	"math"

	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/types"
)

// Rocket represents a rocket entity
type Rocket struct {
	ID           int
	Motor        *components.Motor
	Nosecone     *Nosecone
	Physics      *components.Physics
	Aerodynamics *components.Aerodynamics
}

// NewRocket creates a new rocket instance
func NewRocket(id int, mass float64, motor *components.Motor, nosecone *Nosecone, dragCoefficient float64) *Rocket {
	area := math.Pi * (nosecone.Radius * nosecone.Radius)

	return &Rocket{
		ID:           id,
		Motor:        motor,
		Nosecone:     nosecone,
		Physics:      components.NewPhysics(9.81, mass),
		Aerodynamics: components.NewAerodynamics(dragCoefficient, area),
	}
}

// Update updates the rocket state based on the timestep
func (r *Rocket) Update(dt float64) error {
	if dt <= 0 {
		return fmt.Errorf("invalid timestep: dt must be > 0")
	}

	// Update motor first to get thrust
	if err := r.Motor.Update(dt); err != nil {
		return err
	}

	// Apply motor thrust as a force
	thrustVec := types.Vector3{X: 0, Y: r.Motor.GetThrust(), Z: 0}
	r.Physics.AddForce(thrustVec)

	// Apply aerodynamic drag
	dragForce := r.Aerodynamics.CalculateDrag(r.Physics.Velocity)
	r.Physics.AddForce(dragForce)

	// Update physics for movement
	err := r.Physics.Update(dt)
	if err != nil {
		return err
	}

	return nil
}

// String returns a string representation of the Rocket
func (r *Rocket) String() string {
	return fmt.Sprintf("Rocket{ID: %d, Motor: %s, Nosecone: %s, Physics: %s, Aerodynamics: %s}", r.ID, r.Motor.String(), r.Nosecone.String(), r.Physics.String(), r.Aerodynamics.String())
}
