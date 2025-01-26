package components

import (
	"fmt"
	"sync"

	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

type Motor struct {
	ID          ecs.EntityID
	Position    types.Vector3
	Thrustcurve [][]float64
	Mass        float64
	thrust      float64
	Props       *thrustcurves.MotorData
	FSM         *MotorFSM
	elapsedTime float64
	mu          sync.RWMutex
}

func NewMotor(id ecs.EntityID, md *thrustcurves.MotorData) *Motor {
	m := &Motor{
		ID:          id,
		Position:    types.Vector3{},
		Thrustcurve: md.Thrust,
		Mass:        md.TotalMass,
		Props:       md,
		thrust:      0,
		FSM:         NewMotorFSM(),
	}

	if len(m.Thrustcurve) > 0 {
		m.thrust = m.Thrustcurve[0][1]
	}
	return m
}

func (m *Motor) Update(dt float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if dt <= 0 {
		return fmt.Errorf("invalid timestep: dt must be > 0")
	}

	m.elapsedTime += dt

	// Update FSM state
	err := m.FSM.UpdateState(m.Mass, m.elapsedTime, m.Props.BurnTime)
	if err != nil {
		return err
	}

	// Update thrust and mass based on state
	if m.FSM.GetState() == StateBurning {
		m.thrust = m.interpolateThrust(m.elapsedTime)
		if m.Mass > 0 && m.thrust > 0 {
			massLoss := (m.thrust * dt) / m.Props.AvgThrust
			m.Mass = max(0, m.Mass-massLoss)
		}
	} else {
		m.thrust = 0
	}

	return nil
}

func (m *Motor) GetThrust() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.thrust
}

func (m *Motor) GetMass() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Mass
}

func (m *Motor) Type() string {
	return ecs.ComponentMotor
}

func (m *Motor) interpolateThrust(totalDt float64) float64 {
	if len(m.Thrustcurve) == 0 {
		return 0
	}

	for i := 0; i < len(m.Thrustcurve)-1; i++ {
		if m.Thrustcurve[i][0] <= totalDt && totalDt < m.Thrustcurve[i+1][0] {
			t1, thrust1 := m.Thrustcurve[i][0], m.Thrustcurve[i][1]
			t2, thrust2 := m.Thrustcurve[i+1][0], m.Thrustcurve[i+1][1]
			return thrust1 + (thrust2-thrust1)*(totalDt-t1)/(t2-t1)
		}
	}

	if totalDt >= m.Thrustcurve[len(m.Thrustcurve)-1][0] {
		return m.Thrustcurve[len(m.Thrustcurve)-1][1]
	}

	return 0
}

func (m *Motor) String() string {
	return fmt.Sprintf("Motor{ID: %d, Position: %s, Mass: %f, Thrust: %f}", m.ID, m.Position.String(), m.Mass, m.thrust)
}
