package components

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

// Motor represents the motor component of a rocket
type Motor struct {
	Position    types.Vector3
	Thrustcurve [][]float64 // [[time, thrust], ...]
	Mass        float64
	thrust      float64
	Props       *thrustcurves.MotorData
	fsm         *MotorFSM // FSM instance
	elapsedTime float64   // Time elapsed since ignition
}

// String returns a string representation of the MotorData
func (m *Motor) String() string {
	return fmt.Sprintf("Motor{Position: %v, Mass: %.3f, Thrust: %.3f, State: %s}", m.Position, m.Mass, m.thrust, m.fsm.GetState())
}

// Update updates the motor (uses thrust curves and reduces mass)
func (m *Motor) Update(dt float64) error {
	// Update elapsed time
	m.elapsedTime += dt

	// Update the FSM state
	err := m.fsm.UpdateState(m.Mass, m.elapsedTime, m.Props.BurnTime)
	if err != nil {
		return err
	}

	// If in burning state, calculate thrust and update mass
	if m.fsm.GetState() == StateBurning {
		m.thrust = m.GetThrustAfter(m.elapsedTime)

		// Calculate mass loss based on thrust and time
		if m.Mass > 0 {
			massLoss := (m.thrust * dt) / m.Props.AvgThrust // Average thrust is used for mass loss calculation
			newMass := m.Mass - massLoss

			// Ensure mass does not go negative
			if newMass < 0 {
				newMass = 0
			}
			m.Mass = newMass
		}
	} else {
		m.thrust = 0 // No thrust if idle
	}

	return nil
}

// NewMotor creates a new motor instance
func NewMotor(md *thrustcurves.MotorData) *Motor {
	m := &Motor{
		Position:    types.Vector3{X: 0, Y: 0, Z: 0},
		Thrustcurve: md.Thrust,
		Mass:        md.TotalMass,
		Props:       md,
		thrust:      0,
		fsm:         NewMotorFSM(), // Initialize the FSM
	}

	// Initialize thrust to the first thrust value in the curve
	if len(m.Thrustcurve) > 0 {
		m.thrust = m.Thrustcurve[0][1]
	}
	return m
}

// GetThrustAfter returns the thrust of the Motor at a given time using linear interpolation
func (m *Motor) GetThrustAfter(totalDt float64) float64 {
	// If the thrust curve is empty, return 0
	if len(m.Thrustcurve) == 0 {
		return 0
	}

	// Find the appropriate segment for interpolation
	for i := 0; i < len(m.Thrustcurve)-1; i++ {
		if m.Thrustcurve[i][0] <= totalDt && totalDt < m.Thrustcurve[i+1][0] {
			// Perform linear interpolation
			t1, thrust1 := m.Thrustcurve[i][0], m.Thrustcurve[i][1]
			t2, thrust2 := m.Thrustcurve[i+1][0], m.Thrustcurve[i+1][1]

			// Linear interpolation formula
			return thrust1 + (thrust2-thrust1)*(totalDt-t1)/(t2-t1)
		}
	}

	// If totalDt is beyond the last sample, return the last thrust value
	if totalDt >= m.Thrustcurve[len(m.Thrustcurve)-1][0] {
		return m.Thrustcurve[len(m.Thrustcurve)-1][1]
	}

	// If totalDt is before the first sample, return 0
	return 0
}
