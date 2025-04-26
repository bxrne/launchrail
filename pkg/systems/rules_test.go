package systems_test

import (
	"testing"

	"io"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"
)

func TestProcessRules_NilEntity(t *testing.T) {
	rs := systems.NewRulesSystem(&ecs.World{}, &config.Engine{})
	result := rs.ProcessRules(nil)
	if result != systems.None {
		t.Errorf("Expected None for nil entity, got %v", result)
	}
}

func TestDetectApogee_HighVelocity(t *testing.T) {
	lg := logger.GetLogger("debug")
	rs := systems.NewRulesSystem(&ecs.World{}, &config.Engine{Simulation: config.Simulation{GroundTolerance: 0.1}})
	entity := &states.PhysicsState{
		Velocity: &types.Velocity{Vec: types.Vector3{Y: 10}},
		Motor:    &components.Motor{FSM: components.NewMotorFSM(nil, *lg)},
		Position: &types.Position{Vec: types.Vector3{Y: 100}},
	}
	result := rs.DetectApogee(entity)
	if result {
		t.Errorf("Expected false for high velocity, got %v", result)
	}
}

func TestNewRulesSystem(t *testing.T) {
	rs := systems.NewRulesSystem(&ecs.World{}, &config.Engine{})
	assert.NotNil(t, rs)
}

func TestAdd(t *testing.T) {
	rs := systems.NewRulesSystem(&ecs.World{}, &config.Engine{})
	en := &states.PhysicsState{}

	rs.Add(en)
}

func TestApogeeDetection(t *testing.T) {
	cfg := &config.Engine{
		Simulation: config.Simulation{
			GroundTolerance: 0.1,
		},
	}
	rs := systems.NewRulesSystem(&ecs.World{}, cfg)

	logger := logf.New(logf.Opts{Writer: io.Discard})
	motorProps := &thrustcurves.MotorData{}
	motor := &components.Motor{
		Props: motorProps,
	}
	motor.FSM = components.NewMotorFSM(motor, logger)
	motor.FSM.SetState(components.StateIdle)

	entity := &states.PhysicsState{
		Position:  &types.Position{Vec: types.Vector3{Y: 100}},
		Velocity:  &types.Velocity{Vec: types.Vector3{Y: -0.01}},
		Motor:     motor,
		Parachute: &components.Parachute{Trigger: "apogee", Deployed: false},
	}

	rs.Add(entity)
	_ = rs.Update(0)
	assert.Equal(t, systems.Apogee, rs.GetLastEvent())
	assert.True(t, entity.Parachute.Deployed)
}

func TestLandingDetection(t *testing.T) {
	cfg := &config.Engine{
		Simulation: config.Simulation{
			GroundTolerance: 0.1,
		},
	}
	rs := systems.NewRulesSystem(&ecs.World{}, cfg)
	logger := logf.New(logf.Opts{Writer: io.Discard})
	motorProps := &thrustcurves.MotorData{}
	motor := &components.Motor{
		Props: motorProps,
	}
	motor.FSM = components.NewMotorFSM(motor, logger)
	entity := &states.PhysicsState{
		Position:     &types.Position{Vec: types.Vector3{Y: 100}},
		Velocity:     &types.Velocity{Vec: types.Vector3{Y: 0.0}},
		Acceleration: &types.Acceleration{Vec: types.Vector3{}},
		Motor:        motor,
		Parachute:    &components.Parachute{Trigger: "apogee", Deployed: false},
	}

	rs.Add(entity)
	_ = rs.Update(0)
	assert.Equal(t, systems.Apogee, rs.GetLastEvent())
	assert.True(t, entity.Parachute.Deployed)

	entity.Position.Vec.Y = 0.05
	entity.Velocity.Vec.Y = -0.1
	_ = rs.Update(0)
	assert.Equal(t, systems.Land, rs.GetLastEvent())
}

func TestInvalidEntityHandling(t *testing.T) {
	rs := systems.NewRulesSystem(&ecs.World{}, &config.Engine{})
	entity := &states.PhysicsState{}

	rs.Add(entity)
	_ = rs.Update(0)
	assert.Equal(t, systems.None, rs.GetLastEvent())
}
