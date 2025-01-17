package entities

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/ecs/types"
)

// Rocket represents a rocket entity
type Rocket struct {
	ID           int
	Position     types.Vector3
	Velocity     types.Vector3
	Acceleration types.Vector3
	Mass         float64
	Motor        *Motor    // Reference to the Motor entity
	Nosecone     *Nosecone // Reference to the Nosecone entity
}

// NewRocket creates a new rocket instance
func NewRocket(id int, mass float64, motor *Motor, nosecone *Nosecone) *Rocket {
	return &Rocket{
		ID:           id,
		Position:     types.Vector3{X: 0, Y: 0, Z: 0},
		Velocity:     types.Vector3{X: 0, Y: 0, Z: 0},
		Acceleration: types.Vector3{X: 0, Y: 0, Z: 0},
		Mass:         mass,
		Motor:        motor,
		Nosecone:     nosecone,
	}
}

// Update updates the rocket
func (r *Rocket) Update(dt float64) error {
	// Update the motor
	if r.Motor != nil {
		err := r.Motor.Update(dt)
		if err != nil {
			return err
		}
	}

	// Update the nosecone if needed (e.g., for drag calculations)
	if r.Nosecone != nil {
		// Perform any necessary updates for the nosecone
	}

	// Update other rocket properties (e.g., position, velocity) here

	return nil
}

// String returns a string representation of the Rocket
func (r *Rocket) String() string {
	return fmt.Sprintf("Rocket{ID: %d, Position: %v, Velocity: %v, Acceleration: %v, Mass: %.2f, Motor: %s, Nosecone: %s}", r.ID, r.Position, r.Velocity, r.Acceleration, r.Mass, r.Motor, r.Nosecone)
}
