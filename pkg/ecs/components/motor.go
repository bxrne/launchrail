package components

import (
	"fmt"
	"sync"

	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

// Motor represents the motor component of a rocket
type Motor struct {
	Position    types.Vector3
	Thrustcurve *thrustcurves.MotorData
	Mass        float64
	thrust      float64
	mu          sync.RWMutex
}

// String returns a string representation of the MotorData
func (m *Motor) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf("Motor{Position: %v, Mass: %.2f, Thrust: %.2f}", m.Position, m.Mass, m.thrust)
}

// GetThrust returns the thrust of the Motor
func (m *Motor) GetThrust() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.thrust
}

// Update updates the motor (uses thrust curves and reduces mass)
func (m *Motor) Update(dt float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Initialize thrust to 0 for the current update
	m.thrust = 0

	// Find the thrust value based on the current time
	for _, sample := range m.Thrustcurve.Thrust {
		if sample[0] <= dt {
			m.thrust = sample[1]
		}
	}

}

// NewMotor creates a new motor instance
func NewMotor(thrustcurve *thrustcurves.MotorData, mass float64) *Motor {
	m := &Motor{
		Position:    types.Vector3{X: 0, Y: 0, Z: 0},
		Thrustcurve: thrustcurve,
		Mass:        mass,
		thrust:      0,
	}

	m.thrust = m.Thrustcurve.Thrust[0][1]
	return m
}
