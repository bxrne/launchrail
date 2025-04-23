package systems_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN Nothing WHEN NewRulesSystem is called THEN a new rules system is returned
func TestNewRulesSystem(t *testing.T) {
	rs := systems.NewRulesSystem(&ecs.World{}, &config.Engine{})
	assert.NotNil(t, rs)
}

// TEST: GIVEN a rules system WHEN Add is called with an entities state THEN its stored
func TestAdd(t *testing.T) {
	rs := systems.NewRulesSystem(&ecs.World{}, &config.Engine{})
	en := &states.PhysicsState{}

	rs.Add(en)
}

// TEST: GIVEN an entity WHEN it reaches apogee THEN the event is detected
func TestApogeeDetection(t *testing.T) {
	cfg := &config.Engine{
		Simulation: config.Simulation{
			GroundTolerance: 0.1,
		},
	}
	rs := systems.NewRulesSystem(&ecs.World{}, cfg)

	motor_fsm := components.NewMotorFSM()
	motor_fsm.SetState("BURNOUT")

	entity := &states.PhysicsState{
		Position:  &types.Position{Vec: types.Vector3{Y: 100}},
		Velocity:  &types.Velocity{Vec: types.Vector3{Y: -10}}, 
		Motor:     &components.Motor{FSM: motor_fsm},
		Parachute: &components.Parachute{Trigger: "apogee", Deployed: false},
	}

	rs.Add(entity)
	_ = rs.Update(0)
	assert.Equal(t, systems.Apogee, rs.GetLastEvent())
	assert.True(t, entity.Parachute.Deployed)
}

// TEST: GIVEN an entity WHEN it lands THEN the event is detected
func TestLandingDetection(t *testing.T) {
	cfg := &config.Engine{
		Simulation: config.Simulation{
			GroundTolerance: 0.1,
		},
	}
	rs := systems.NewRulesSystem(&ecs.World{}, cfg)
	motor_fsm := components.NewMotorFSM()
	entity := &states.PhysicsState{
		Position:     &types.Position{Vec: types.Vector3{Y: 0.05}}, // Slightly above ground tolerance
		Velocity:     &types.Velocity{Vec: types.Vector3{Y: -0.1}}, // Simulate downward velocity
		Acceleration: &types.Acceleration{Vec: types.Vector3{}},    // Initialize acceleration
		Motor:        &components.Motor{FSM: motor_fsm},                          // Initialize motor
		Parachute:    &components.Parachute{Trigger: "apogee", Deployed: true},
	}

	rs.Add(entity)
	_ = rs.Update(0)
	assert.Equal(t, systems.Land, rs.GetLastEvent())
}

// TEST: GIVEN an invalid entity WHEN processed THEN no event is triggered
func TestInvalidEntityHandling(t *testing.T) {
	rs := systems.NewRulesSystem(&ecs.World{}, &config.Engine{})
	entity := &states.PhysicsState{}

	rs.Add(entity)
	_ = rs.Update(0)
	assert.Equal(t, systems.None, rs.GetLastEvent())
}
