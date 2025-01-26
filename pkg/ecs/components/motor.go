package components

import (
	"fmt"
	"sync"

	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/looplab/fsm"
)

type Motor struct {
	ID          ecs.EntityID
	Position    types.Vector3
	Thrustcurve [][]float64
	Mass        float64
	thrust      float64
	Props       *thrustcurves.MotorData
	fsm         *fsm.FSM
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
		fsm: fsm.NewFSM(
			"idle",
			fsm.Events{
				{Name: "ignite", Src: []string{"idle"}, Dst: "burning"},
				{Name: "extinguish", Src: []string{"burning"}, Dst: "idle"},
			},
			fsm.Callbacks{},
		),
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

	// Update state
	if m.Mass > 0 && m.elapsedTime <= m.Props.BurnTime {
		if m.fsm.Current() == "idle" {
			m.fsm.Event(nil, "ignite")
		}
	} else if m.fsm.Current() == "burning" {
		m.fsm.Event(nil, "extinguish")
	}

	// Update thrust and mass
	if m.fsm.Current() == "burning" {
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

func (m *Motor) Type() string {
	return ecs.ComponentMotor
}

// interpolateThrust returns the thrust of the Motor at a given time using linear interpolation
func (m *Motor) interpolateThrust(totalDt float64) float64 {
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

func (m *Motor) String() string {
	return fmt.Sprintf("Motor{ID: %d, Position: %s, Mass: %f, Thrust: %f}", m.ID, m.Position.String(), m.Mass, m.thrust)
}

func (m *Motor) GetMass() float64 {
	return m.Mass
}
