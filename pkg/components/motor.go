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
	coasting    bool
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
		// Initialize thrust to first data point
		thrust:      md.Thrust[0][1],
		FSM:         NewMotorFSM(),
		burnTime:    md.BurnTime,
		logger:      logger,
	}

	m.logger.Info("Motor created", "ID", m.ID.ID(), "Mass", m.Mass, "BurnTime", m.burnTime)
	m.FSM.Event(context.Background(), "ignite")
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
	if dt < 0 {
		return fmt.Errorf("invalid negative timestep")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update elapsed time
	m.elapsedTime += dt
	// Update thrust based on thrust curve
	m.thrust = m.interpolateThrust(m.elapsedTime)
	// Update mass if still generating thrust
	if m.thrust > 0 {
		propellantMassFlow := m.Props.TotalMass / m.burnTime
		massLoss := propellantMassFlow * dt
		m.Mass = math.Max(0, m.Mass-massLoss)
	} else {
		m.coasting = true
	}
	// Update the motor FSM state
	_ = m.FSM.UpdateState(m.Mass, m.elapsedTime, m.burnTime)
	return nil
}

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

func (m *Motor) GetThrust() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.thrust
}

func (m *Motor) IsCoasting() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.coasting
}

func (m *Motor) GetState() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.FSM.Current()
}

func (m *Motor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.elapsedTime = 0
	m.Mass = m.Props.TotalMass
	m.thrust = m.Thrustcurve[0][1]
	m.coasting = false
	m.FSM = NewMotorFSM()
	m.FSM.Event(context.Background(), "ignite")
}

func (m *Motor) SetState(state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FSM.SetState(state)
}

func (m *Motor) GetMass() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Mass
}

func (m *Motor) Type() string {
	return "Motor"
}

func (m *Motor) String() string {
	return fmt.Sprintf("Motor{ID: %d, Position: %s, Mass: %f, Thrust: %f}", m.ID.ID(), m.Position.String(), m.Mass, m.thrust)
}

func (m *Motor) GetPlanformArea() float64 {
	return 0
}

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
