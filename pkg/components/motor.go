package components

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/EngoEngine/ecs"

	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
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
	logger      logf.Logger
}

// NewMotor creates a new motor component from thrust curve data
func NewMotor(id ecs.BasicEntity, md *thrustcurves.MotorData, logger logf.Logger) (*Motor, error) {
	if md == nil || len(md.Thrust) == 0 {
		return nil, fmt.Errorf("thrust curve data is required")
	}

	if md.MaxThrust <= 0 {
		md.MaxThrust = findMaxThrust(md.Thrust)
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
		logger:      logger,
	}

	m.logger.Info("Motor created", "ID", m.ID.ID(), "Mass", m.Mass, "BurnTime", m.burnTime)
	// Initialize with first thrust point
	m.thrust = m.Thrustcurve[0][1]
	return m, nil
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

func (m *Motor) Update(dt float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if dt <= 0 || math.IsNaN(dt) {
		return fmt.Errorf("invalid timestep")
	}

	// Update elapsed time
	m.elapsedTime += dt

	// First interpolate thrust
	m.thrust = m.interpolateThrust(m.elapsedTime)

	// Handle state and mass updates
	if m.elapsedTime <= m.burnTime {
		// Force burning state
		if m.FSM.Current() == StateIdle {
			ctx := context.Background()
			if err := m.FSM.Event(ctx, "ignite"); err != nil {
				return fmt.Errorf("failed to transition to burning state: %v", err)
			}
		}

		// Update mass proportional to thrust
		if m.FSM.Current() == StateBurning {
			// Calculate mass loss based on thrust ratio
			thrustRatio := m.thrust / m.Props.MaxThrust
			massFlow := m.interpolateMassFlow(m.elapsedTime)
			massLoss := massFlow * dt * thrustRatio

			// Protect against numerical errors
			newMass := m.Mass - massLoss
			if newMass < 0 {
				newMass = 0
			}
			m.Mass = math.Max(0, newMass)
		}
	} else if m.elapsedTime >= m.burnTime && m.FSM.Current() == StateBurning {
		return m.handleBurnout()
	}

	return nil
}

// interpolateMassFlow calculates mass flow rate at a given time
func (m *Motor) interpolateMassFlow(t float64) float64 {
	if t <= 0 || t >= m.burnTime {
		return 0
	}

	// Find thrust at current time
	currentThrust := m.interpolateThrust(t)

	// Calculate instantaneous mass flow based on rocket equation
	// dm/dt = F/ve where ve is exhaust velocity (assumed constant)
	// We can approximate ve using average values
	averageThrust := m.Props.AvgThrust
	averageMassFlow := m.Props.TotalMass / m.burnTime

	if averageThrust <= 0 {
		return 0
	}

	// Scale mass flow by thrust ratio
	return (currentThrust / averageThrust) * averageMassFlow
}

func (m *Motor) handleBurnout() error {
	// Only transition to burnout if we're burning
	if m.FSM.Current() == StateBurning {
		m.thrust = 0
		ctx := context.Background()
		if err := m.FSM.Event(ctx, "burnout"); err != nil {
			m.logger.Error("failed to transition to idle state", "error", err)
		}
	}
	return nil
}

func (m *Motor) interpolateThrust(totalDt float64) float64 {
	// If before burn start, use initial thrust
	if totalDt <= m.Thrustcurve[0][0] {
		return m.Thrustcurve[0][1]
	}

	// If past burn time, return 0
	if totalDt > m.burnTime {
		return 0
	}

	// Find the surrounding data points for interpolation
	for i := 0; i < len(m.Thrustcurve)-1; i++ {
		t1, thrust1 := m.Thrustcurve[i][0], m.Thrustcurve[i][1]
		t2, thrust2 := m.Thrustcurve[i+1][0], m.Thrustcurve[i+1][1]

		if totalDt >= t1 && totalDt <= t2 {
			// Linear interpolation
			ratio := (totalDt - t1) / (t2 - t1)
			return thrust1 + (ratio * (thrust2 - thrust1))
		}
	}

	// If we're between last data point and burn time
	// Use the last thrust value
	return m.Thrustcurve[len(m.Thrustcurve)-1][1]
}

func (m *Motor) updateThrustAndMass(dt float64) {
	// Always interpolate thrust regardless of state
	m.thrust = m.interpolateThrust(m.elapsedTime)

	// Always update mass if we have thrust
	if m.Mass > 0 && m.thrust > 0 {
		propellantMassFlow := m.Props.TotalMass / m.burnTime
		massLoss := propellantMassFlow * dt
		m.Mass = math.Max(0, m.Mass-massLoss)
	}
}

// GetThrust returns the current thrust of the motor
func (m *Motor) GetThrust() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return interpolated thrust if burning
	if m.FSM.Current() == StateBurning {
		return m.interpolateThrust(m.elapsedTime)
	}
	return 0
}

// IsCoasting returns true if the motor has completed its burn
func (m *Motor) IsCoasting() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.FSM.Current() == StateIdle
}

// GetState returns the current state of the motor
func (m *Motor) GetState() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.FSM.Current()
}

// Reset resets the motor state for potential reuse
func (m *Motor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.elapsedTime = 0
	m.thrust = m.Thrustcurve[0][1]
	m.Mass = m.Props.TotalMass
	m.FSM = NewMotorFSM()
}

// SetState (testing only) sets the motor state to a specific value
func (m *Motor) SetState(state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FSM.SetState(state)
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

func findMaxThrust(thrustData [][]float64) float64 {
	maxVal := 0.0
	for _, point := range thrustData {
		if point[1] > maxVal {
			maxVal = point[1]
		}
	}
	return maxVal
}
