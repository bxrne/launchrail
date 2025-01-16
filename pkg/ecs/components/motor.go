package components

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

// Motor represents the motor component of a rocket
type Motor struct {
	Position    types.Vector3
	Thrustcurve *thrustcurves.MotorData
	Mass        float64
	thrust      float64
}

// String returns a string representation of the MotorData
func (m *Motor) String() string {
	return fmt.Sprintf("Motor{Position: %v, Mass: %.2f, Thrust: %.2f}", m.Position, m.Thrustcurve.Designation, m.Mass, m.thrust)
}

// Update updates the motor (uses thrust curves and reduces mass)
func (m *Motor) Update(dt float64) {
	for _, sample := range m.Thrustcurve.Thrust {
		if sample[0] <= dt {
			m.thrust = sample[1]
		}
	}

	m.Mass -= m.thrust * dt
}

// NewMotor creates a new motor instance
func NewMotor(thrustcurve *thrustcurves.MotorData, mass float64) *Motor {
	return &Motor{
		Position:    types.Vector3{X: 0, Y: 0, Z: 0},
		Thrustcurve: thrustcurve,
		Mass:        mass,
		thrust:      0,
	}
}
