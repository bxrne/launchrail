package components

import (
	"fmt"
	"math"
	"sync"

	"github.com/EngoEngine/ecs"

	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

// MotorState represents the current state of the motor
type MotorState string

// MotorState constants
const (
	MotorIgnited  MotorState = "IGNITED"
	MotorBurning  MotorState = "BURNING"
	MotorBurnout  MotorState = "BURNOUT"
	MotorCoasting MotorState = "COASTING"
)

// Motor represents a rocket motor component
type Motor struct {
	ID          ecs.BasicEntity
	Position    types.Vector3
	Thrustcurve [][]float64
	Mass        float64
	thrust      float64
	Props       *thrustcurves.MotorData
	FSM         *MotorFSM
	elapsedTime float64
	mu          sync.RWMutex
	burnTime    float64
	isCoasting  bool
	logger      logf.Logger // Add logger
	state       MotorState  // Add motor state
}

// NewMotor creates a new motor component from thrust curve data
func NewMotor(id ecs.BasicEntity, md *thrustcurves.MotorData, logger logf.Logger) *Motor {
	if md == nil || len(md.Thrust) == 0 {
		panic("invalid motor data")
	}

	m := &Motor{
		ID:          id,
		Position:    types.Vector3{},
		Thrustcurve: validateThrustCurve(md.Thrust),
		Mass:        md.TotalMass,
		Props:       md,
		thrust:      0,
		FSM:         NewMotorFSM(),
		burnTime:    md.BurnTime,
		isCoasting:  false,
		logger:      logger,       // Initialize logger
		state:       MotorIgnited, // Initial state
	}

	// Initialize with first thrust point
	m.thrust = m.Thrustcurve[0][1]
	return m
}

// validateThrustCurve ensures thrust curve data is valid and properly formatted
func validateThrustCurve(curve [][]float64) [][]float64 {
	if len(curve) < 2 {
		panic("thrust curve must have at least 2 points")
	}

	// Ensure time points are monotonically increasing
	for i := 1; i < len(curve); i++ {
		if curve[i][0] <= curve[i-1][0] {
			panic("thrust curve time points must be strictly increasing")
		}
	}

	// Ensure no negative thrust values
	for _, point := range curve {
		if point[1] < 0 {
			panic("negative thrust values are invalid")
		}
	}

	return curve
}

// Update updates the motor state based on the current time step
func (m *Motor) Update(dt float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if dt <= 0 || math.IsNaN(dt) {
		return fmt.Errorf("invalid timestep")
	}

	// Calculate thrust based on time step
	if !m.isCoasting && m.elapsedTime < m.burnTime {
		m.thrust = m.interpolateThrust(m.elapsedTime)

		// Calculate mass loss proportional to timestep
		if m.Mass > 0 && m.thrust > 0 {
			massLossRate := (m.Props.TotalMass - m.Props.WetMass) / m.burnTime
			m.Mass = math.Max(0, m.Mass-(massLossRate*dt))
		}

		m.elapsedTime += dt
	}

	// Call FSM to update internal state:
	err := m.FSM.UpdateState(m.Mass, m.elapsedTime, m.burnTime)
	if err != nil {
		return err
	}
	// Reflect FSM state in Motor.state
	switch m.FSM.GetState() {
	case StateIdle:
		m.state = MotorCoasting
	case StateBurning:
		m.state = MotorBurning
	}

	// Strict burn time and thrust validation
	if m.elapsedTime >= m.burnTime {
		if m.thrust > 0 {
			return fmt.Errorf("motor still producing thrust after burn time: t=%v, thrust=%v",
				m.elapsedTime, m.thrust)
		}
		m.isCoasting = true
		m.thrust = 0
		m.state = MotorBurnout // Update state
		return nil
	}

	m.elapsedTime += dt

	// Only update thrust during burn
	if !m.isCoasting {
		m.thrust = m.interpolateThrust(m.elapsedTime)

		// Update mass only during thrust
		if m.Mass > 0 && m.thrust > 0 {
			massLoss := (m.thrust * dt) / m.Props.AvgThrust
			m.Mass = math.Max(0, m.Mass-massLoss)
		}

		// Update state based on current thrust calculation
		if m.elapsedTime > 0 {
			m.state = MotorBurning
		}
	}

	return nil
}

// GetThrust returns the current thrust of the motor
func (m *Motor) GetThrust() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.isCoasting || m.elapsedTime >= m.burnTime {
		return 0
	}

	return m.thrust
}

// IsCoasting returns true if the motor has completed its burn
func (m *Motor) IsCoasting() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state == MotorCoasting || m.state == MotorBurnout
}

// GetState returns the current state of the motor
func (m *Motor) GetState() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return string(m.state)
}

// Reset resets the motor state for potential reuse
func (m *Motor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.elapsedTime = 0
	m.isCoasting = false
	m.thrust = m.Thrustcurve[0][1]
	m.Mass = m.Props.TotalMass
	m.FSM = NewMotorFSM()
	m.state = MotorCoasting // Reset state
}

// GetMass returns the current mass of the motor
func (m *Motor) GetMass() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Mass
}

// Type returns the type of the motor component
func (m *Motor) Type() string {
	return "Motor"
}

// interpolateThrust interpolates the thrust value at a given time point
func (m *Motor) interpolateThrust(totalDt float64) float64 {
	if m.isCoasting || totalDt >= m.burnTime {
		return 0
	}

	// Clamp to valid range
	if totalDt < 0 {
		return m.Thrustcurve[0][1]
	}

	// Find interpolation points
	for i := 0; i < len(m.Thrustcurve)-1; i++ {
		t1, thrust1 := m.Thrustcurve[i][0], m.Thrustcurve[i][1]
		t2, thrust2 := m.Thrustcurve[i+1][0], m.Thrustcurve[i+1][1]

		if t1 <= totalDt && totalDt < t2 {
			ratio := (totalDt - t1) / (t2 - t1)
			return thrust1 + (thrust2-thrust1)*ratio
		}
	}

	return 0
}

// String returns a string representation of the motor component
func (m *Motor) String() string {
	return fmt.Sprintf("Motor{ID: %d, Position: %s, Mass: %f, Thrust: %f}", m.ID.ID(), m.Position.String(), m.Mass, m.thrust)
}

// GetPlanformArea returns the planform area of the motor
func (m *Motor) GetPlanformArea() float64 {
	return 0
}

// Add method to get elapsed time
func (m *Motor) GetElapsedTime() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.elapsedTime
}
