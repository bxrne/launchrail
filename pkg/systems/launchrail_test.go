package systems_test

import (
	"math"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a new launch rail WHEN initialized THEN parameters are set correctly
func TestNewLaunchRailSystem(t *testing.T) {
	world := &ecs.World{}
	length := 2.0
	angle := 5.0
	orientation := 0.0
	logfLogger := logf.New(logf.Opts{})
	logger := &logfLogger

	system := systems.NewLaunchRailSystem(world, length, angle, orientation, logger)

	assert.NotNil(t, system)
	assert.Equal(t, length, system.GetRail().Length)
	assert.InDelta(t, angle*math.Pi/180.0, system.GetRail().Angle, 0.0001)
	assert.Equal(t, orientation, system.GetRail().Orientation)
}

// TEST: GIVEN a launch rail system WHEN adding an entity THEN entity is tracked
func TestAddEntity(t *testing.T) {
	world := &ecs.World{}
	logfLogger := logf.New(logf.Opts{})
	logger := &logfLogger
	system := systems.NewLaunchRailSystem(world, 2.0, 5.0, 0.0, logger)

	entity := &states.PhysicsState{
		Position:     &types.Position{Vec: types.Vector3{X: 0, Y: 0}},
		Velocity:     &types.Velocity{Vec: types.Vector3{X: 0, Y: 0}},
		Mass:         &types.Mass{Value: 1.0},
		Acceleration: &types.Acceleration{Vec: types.Vector3{}},
	}

	system.Add(entity)
	assert.Contains(t, system.GetEntities(), entity)
}

// TEST: GIVEN a stationary rocket WHEN no thrust THEN stays at start of rail
func TestUpdateNoThrust(t *testing.T) {
	world := &ecs.World{}
	logfLogger := logf.New(logf.Opts{})
	logger := &logfLogger
	system := systems.NewLaunchRailSystem(world, 2.0, 5.0, 0.0, logger)

	entity := &states.PhysicsState{
		Position:     &types.Position{Vec: types.Vector3{X: 0, Y: 0}},
		Velocity:     &types.Velocity{Vec: types.Vector3{X: 0, Y: 0}},
		Mass:         &types.Mass{Value: 1.0},
		Acceleration: &types.Acceleration{Vec: types.Vector3{}},
	}

	system.Add(entity)
	err := system.Update(0.01)

	assert.NoError(t, err)
	assert.Equal(t, 0.0, entity.Position.Vec.X)
	assert.Equal(t, 0.0, entity.Position.Vec.Y)
}
