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
	Thrustcurve [][]float64
	Mass        float64
	thrust      float64
	Props       *thrustcurves.MotorData
	mu          sync.RWMutex
}

// defined elsewhere but within scope
// type MotorData struct {
// 	Designation  designation.Designation
// 	ID           string
// 	Thrust       [][]float64 // [[time, thrust], ...]
// 	TotalImpulse float64     // Newton-seconds
// 	BurnTime     float64     // Seconds
// 	AvgThrust    float64     // Newtons
// 	TotalMass    float64     // Kg
// 	WetMass      float64     // Kg
// 	MaxThrust    float64     // Newtons
// }

// String returns a string representation of the MotorData
func (m *Motor) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf("Motor{Position: %v, Mass: %.2f, Thrust: %.2f}", m.Position, m.Mass, m.thrust)
}

// GetThrustAfter returns the thrust of the Motor at given time
func (m *Motor) GetThrustAfter(total_dt float64) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find the thrust value based on the current time
	for _, sample := range m.Thrustcurve {
		if sample[0] <= total_dt {
			return sample[1]
		}

	}
	return 0
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

	m.thrust = m.GetThrustAfter(dt)
	m.Mass -= m.Props.WetMass / m.Props.BurnTime * dt

}

// NewMotor creates a new motor instance
func NewMotor(md *thrustcurves.MotorData) *Motor {
	m := &Motor{
		Position:    types.Vector3{X: 0, Y: 0, Z: 0},
		Thrustcurve: md.Thrust,
		Mass:        md.TotalMass,
		Props:       md,
		thrust:      0,
	}

	// Initialize thrust to the first thrust value in the curve
	if len(m.Thrustcurve) > 0 {
		m.thrust = m.Thrustcurve[0][1]
	}
	return m
}
