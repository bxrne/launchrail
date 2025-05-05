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
	world := &ecs.World{}
	cfg := &config.Engine{}
	logger := logf.New(logf.Opts{Writer: io.Discard})
	rs := systems.NewRulesSystem(world, cfg, logger)
	result := rs.ProcessRules(nil)
	if result != types.None {
		t.Errorf("Expected None for nil entity, got %v", result)
	}
}

func TestDetectApogee_HighVelocity(t *testing.T) {
	lg := logger.GetLogger("debug")
	world := &ecs.World{}
	cfg := &config.Engine{Simulation: config.Simulation{GroundTolerance: 0.1}}
	logger := logf.New(logf.Opts{Writer: io.Discard})
	rs := systems.NewRulesSystem(world, cfg, logger)
	motor := &components.Motor{}
	motor.FSM = components.NewMotorFSM(motor, *lg)
	entity := &states.PhysicsState{
		Velocity: &types.Velocity{Vec: types.Vector3{Y: 10}},
		Motor:    motor,
		Position: &types.Position{Vec: types.Vector3{Y: 100}},
	}
	result := rs.DetectApogee(entity)
	if result {
		t.Errorf("Expected false for high velocity, got %v", result)
	}
}

func TestNewRulesSystem(t *testing.T) {
	world := &ecs.World{}
	cfg := &config.Engine{}
	logger := logf.New(logf.Opts{Writer: io.Discard})
	rs := systems.NewRulesSystem(world, cfg, logger)
	assert.NotNil(t, rs)
}

func TestAdd(t *testing.T) {
	world := &ecs.World{}
	cfg := &config.Engine{}
	logger := logf.New(logf.Opts{Writer: io.Discard})
	rs := systems.NewRulesSystem(world, cfg, logger)
	en := &states.PhysicsState{}

	rs.Add(en)
}

func TestApogeeDetection(t *testing.T) {
	cfg := &config.Engine{
		Simulation: config.Simulation{
			GroundTolerance: 0.1,
		},
	}
	world := &ecs.World{}
	logger := logf.New(logf.Opts{Writer: io.Discard})
	rs := systems.NewRulesSystem(world, cfg, logger)

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

	// 1. Simulate Liftoff
	entity.Position.Vec.Y = 1.0 // Ensure above ground
	entity.Motor.FSM.SetState(components.StateBurning) // Use SetState to force motor state
	_ = rs.Update(0)
	assert.Equal(t, types.Liftoff, rs.GetLastEvent(), "Event should be Liftoff after first update")

	// 2. Simulate Apogee condition (motor idle, negative velocity)
	entity.Position.Vec.Y = 100.0 // Set altitude high for apogee
	entity.Velocity.Vec.Y = -0.01 // Negative velocity indicating descent
	entity.Motor.FSM.SetState(components.StateIdle)   // Use SetState to force motor state
	_ = rs.Update(0)

	// 3. Assert Apogee detection and parachute deployment
	assert.Equal(t, types.Apogee, rs.GetLastEvent(), "Event should be Apogee after second update")
	assert.True(t, entity.Parachute.Deployed, "Parachute should be deployed at apogee")
}

func TestLandingDetection(t *testing.T) {
	cfg := &config.Engine{
		Simulation: config.Simulation{
			GroundTolerance: 0.1,
		},
	}
	world := &ecs.World{}
	logger := logf.New(logf.Opts{Writer: io.Discard})
	rs := systems.NewRulesSystem(world, cfg, logger)
	motorProps := &thrustcurves.MotorData{}
	motor := &components.Motor{
		Props: motorProps,
	}
	motor.FSM = components.NewMotorFSM(motor, logger)
	entity := &states.PhysicsState{
		Position:     &types.Position{Vec: types.Vector3{Y: 0.0}}, // Start at ground level
		Velocity:     &types.Velocity{Vec: types.Vector3{Y: 0.0}},
		Acceleration: &types.Acceleration{Vec: types.Vector3{}},
		Motor:        motor,
		Parachute:    &components.Parachute{Trigger: "apogee", Deployed: false},
	}

	rs.Add(entity)

	// 1. Simulate Liftoff
	entity.Position.Vec.Y = 1.0 // Ensure above ground
	entity.Motor.FSM.SetState(components.StateBurning) // Force motor to burning state
	_ = rs.Update(0)
	assert.Equal(t, types.Liftoff, rs.GetLastEvent(), "Event should be Liftoff")

	// 2. Simulate Apogee
	entity.Position.Vec.Y = 100.0 // Set altitude high for apogee
	entity.Velocity.Vec.Y = -0.01 // Negative velocity indicating descent
	entity.Motor.FSM.SetState(components.StateIdle)   // Motor should be idle/coasting at apogee
	_ = rs.Update(0)
	assert.Equal(t, types.Apogee, rs.GetLastEvent(), "Event should be Apogee")
	assert.True(t, entity.Parachute.Deployed, "Parachute should be deployed at apogee")

	// 3. Simulate Landing
	entity.Position.Vec.Y = 0.05 // Below ground tolerance
	entity.Velocity.Vec.Y = -0.1 // Moving down
	_ = rs.Update(0)
	assert.Equal(t, types.Land, rs.GetLastEvent(), "Event should be Land")
}

func TestInvalidEntityHandling(t *testing.T) {
	world := &ecs.World{}
	cfg := &config.Engine{}
	logger := logf.New(logf.Opts{Writer: io.Discard})
	rs := systems.NewRulesSystem(world, cfg, logger)
	entity := &states.PhysicsState{}

	rs.Add(entity)
	_ = rs.Update(0)
	assert.Equal(t, types.None, rs.GetLastEvent())
}
