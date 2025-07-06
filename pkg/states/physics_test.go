package states_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPhysicsState_BasicEntity tests that BasicEntity is properly embedded
func TestPhysicsState_BasicEntity(t *testing.T) {
	state := &states.PhysicsState{}

	// Test that BasicEntity is embedded correctly
	basicEntity := ecs.NewBasic()
	state.BasicEntity = basicEntity

	assert.Equal(t, basicEntity.ID(), state.ID())
}

// TestPhysicsState_Initialization tests initialization of PhysicsState with all components
func TestPhysicsState_Initialization(t *testing.T) {
	state := &states.PhysicsState{
		BasicEntity:         ecs.NewBasic(),
		Position:            &types.Position{Vec: types.Vector3{X: 1.0, Y: 2.0, Z: 3.0}},
		Velocity:            &types.Velocity{Vec: types.Vector3{X: 0.1, Y: 0.2, Z: 0.3}},
		Acceleration:        &types.Acceleration{Vec: types.Vector3{X: 0.01, Y: 0.02, Z: 0.03}},
		AngularVelocity:     &types.Vector3{X: 0.001, Y: 0.002, Z: 0.003},
		AngularAcceleration: &types.Vector3{X: 0.0001, Y: 0.0002, Z: 0.0003},
		Orientation:         &types.Orientation{},
		Mass:                &types.Mass{Value: 10.5},
		Time:                5.5,
		AccumulatedForce:    types.Vector3{X: 100.0, Y: 200.0, Z: 300.0},
		AccumulatedMoment:   types.Vector3{X: 10.0, Y: 20.0, Z: 30.0},
		CurrentEvent:        types.Liftoff,
	}

	// Test all fields are properly set
	assert.NotNil(t, state.Position)
	assert.NotNil(t, state.Velocity)
	assert.NotNil(t, state.Acceleration)
	assert.NotNil(t, state.AngularVelocity)
	assert.NotNil(t, state.AngularAcceleration)
	assert.NotNil(t, state.Orientation)
	assert.NotNil(t, state.Mass)

	assert.Equal(t, 1.0, state.Position.Vec.X)
	assert.Equal(t, 2.0, state.Position.Vec.Y)
	assert.Equal(t, 3.0, state.Position.Vec.Z)

	assert.Equal(t, 0.1, state.Velocity.Vec.X)
	assert.Equal(t, 0.2, state.Velocity.Vec.Y)
	assert.Equal(t, 0.3, state.Velocity.Vec.Z)

	assert.Equal(t, 0.01, state.Acceleration.Vec.X)
	assert.Equal(t, 0.02, state.Acceleration.Vec.Y)
	assert.Equal(t, 0.03, state.Acceleration.Vec.Z)

	assert.Equal(t, 0.001, state.AngularVelocity.X)
	assert.Equal(t, 0.002, state.AngularVelocity.Y)
	assert.Equal(t, 0.003, state.AngularVelocity.Z)

	assert.Equal(t, 0.0001, state.AngularAcceleration.X)
	assert.Equal(t, 0.0002, state.AngularAcceleration.Y)
	assert.Equal(t, 0.0003, state.AngularAcceleration.Z)

	assert.Equal(t, 10.5, state.Mass.Value)
	assert.Equal(t, 5.5, state.Time)

	assert.Equal(t, 100.0, state.AccumulatedForce.X)
	assert.Equal(t, 200.0, state.AccumulatedForce.Y)
	assert.Equal(t, 300.0, state.AccumulatedForce.Z)

	assert.Equal(t, 10.0, state.AccumulatedMoment.X)
	assert.Equal(t, 20.0, state.AccumulatedMoment.Y)
	assert.Equal(t, 30.0, state.AccumulatedMoment.Z)

	assert.Equal(t, types.Liftoff, state.CurrentEvent)
}

// TestPhysicsState_NilFields tests behavior with nil fields
func TestPhysicsState_NilFields(t *testing.T) {
	state := &states.PhysicsState{}

	// Test that nil fields are handled gracefully
	assert.Nil(t, state.Position)
	assert.Nil(t, state.Velocity)
	assert.Nil(t, state.Acceleration)
	assert.Nil(t, state.AngularVelocity)
	assert.Nil(t, state.AngularAcceleration)
	assert.Nil(t, state.Orientation)
	assert.Nil(t, state.Mass)
	assert.Nil(t, state.Motor)
	assert.Nil(t, state.Bodytube)
	assert.Nil(t, state.Nosecone)
	assert.Nil(t, state.Finset)
	assert.Nil(t, state.Parachute)
}

// TestPhysicsState_ComponentsAssignment tests assignment of rocket components
func TestPhysicsState_ComponentsAssignment(t *testing.T) {
	state := &states.PhysicsState{
		BasicEntity: ecs.NewBasic(),
	}

	// Create and assign components
	motor := &components.Motor{}
	bodytube := &components.Bodytube{}
	nosecone := &components.Nosecone{}
	finset := &components.TrapezoidFinset{}
	parachute := &components.Parachute{}

	state.Motor = motor
	state.Bodytube = bodytube
	state.Nosecone = nosecone
	state.Finset = finset
	state.Parachute = parachute

	// Test components are properly assigned
	assert.Equal(t, motor, state.Motor)
	assert.Equal(t, bodytube, state.Bodytube)
	assert.Equal(t, nosecone, state.Nosecone)
	assert.Equal(t, finset, state.Finset)
	assert.Equal(t, parachute, state.Parachute)
}

// TestPhysicsState_InertiaTensors tests inertia tensor operations
func TestPhysicsState_InertiaTensors(t *testing.T) {
	state := &states.PhysicsState{}

	// Test identity matrix setup
	identity := types.Matrix3x3{
		M11: 1, M12: 0, M13: 0,
		M21: 0, M22: 1, M23: 0,
		M31: 0, M32: 0, M33: 1,
	}

	state.InertiaTensorBody = identity
	state.InverseInertiaTensorBody = identity

	assert.Equal(t, identity, state.InertiaTensorBody)
	assert.Equal(t, identity, state.InverseInertiaTensorBody)

	// Test non-trivial inertia tensor
	nonTrivial := types.Matrix3x3{
		M11: 2.5, M12: 0.1, M13: 0.2,
		M21: 0.1, M22: 3.0, M23: 0.3,
		M31: 0.2, M32: 0.3, M33: 1.8,
	}

	state.InertiaTensorBody = nonTrivial
	assert.Equal(t, nonTrivial, state.InertiaTensorBody)
}

// TestPhysicsState_ForceAccumulation tests force and moment accumulation
func TestPhysicsState_ForceAccumulation(t *testing.T) {
	state := &states.PhysicsState{}

	// Test initial zero forces
	assert.Equal(t, 0.0, state.AccumulatedForce.X)
	assert.Equal(t, 0.0, state.AccumulatedForce.Y)
	assert.Equal(t, 0.0, state.AccumulatedForce.Z)

	assert.Equal(t, 0.0, state.AccumulatedMoment.X)
	assert.Equal(t, 0.0, state.AccumulatedMoment.Y)
	assert.Equal(t, 0.0, state.AccumulatedMoment.Z)

	// Test force accumulation
	force1 := types.Vector3{X: 10.0, Y: 20.0, Z: 30.0}
	force2 := types.Vector3{X: 5.0, Y: 15.0, Z: 25.0}

	state.AccumulatedForce = state.AccumulatedForce.Add(force1)
	state.AccumulatedForce = state.AccumulatedForce.Add(force2)

	expectedForce := types.Vector3{X: 15.0, Y: 35.0, Z: 55.0}
	assert.Equal(t, expectedForce, state.AccumulatedForce)

	// Test moment accumulation
	moment1 := types.Vector3{X: 1.0, Y: 2.0, Z: 3.0}
	moment2 := types.Vector3{X: 0.5, Y: 1.5, Z: 2.5}

	state.AccumulatedMoment = state.AccumulatedMoment.Add(moment1)
	state.AccumulatedMoment = state.AccumulatedMoment.Add(moment2)

	expectedMoment := types.Vector3{X: 1.5, Y: 3.5, Z: 5.5}
	assert.Equal(t, expectedMoment, state.AccumulatedMoment)
}

// TestPhysicsState_EventHandling tests event state management
func TestPhysicsState_EventHandling(t *testing.T) {
	state := &states.PhysicsState{}

	// Test initial event state
	assert.Equal(t, types.None, state.CurrentEvent)

	// Test event transitions
	events := []types.Event{
		types.Liftoff,
		types.Burnout,
		types.Apogee,
		types.ParachuteDeploy,
		types.Land,
	}

	for _, event := range events {
		state.CurrentEvent = event
		assert.Equal(t, event, state.CurrentEvent)
	}
}

// TestPhysicsState_TimeProgression tests time progression tracking
func TestPhysicsState_TimeProgression(t *testing.T) {
	state := &states.PhysicsState{}

	// Test initial time
	assert.Equal(t, 0.0, state.Time)

	// Test time progression
	times := []float64{0.01, 0.1, 1.0, 10.0, 100.0}

	for _, time := range times {
		state.Time = time
		assert.Equal(t, time, state.Time)
		assert.True(t, state.Time >= 0.0)
	}
}

// TestPhysicsState_CompleteRocketState tests a complete rocket state setup
func TestPhysicsState_CompleteRocketState(t *testing.T) {
	state := &states.PhysicsState{
		BasicEntity:         ecs.NewBasic(),
		Position:            &types.Position{Vec: types.Vector3{X: 0, Y: 0, Z: 0}},
		Velocity:            &types.Velocity{Vec: types.Vector3{X: 0, Y: 50, Z: 0}},
		Acceleration:        &types.Acceleration{Vec: types.Vector3{X: 0, Y: 10, Z: 0}},
		AngularVelocity:     &types.Vector3{X: 0, Y: 0, Z: 0.1},
		AngularAcceleration: &types.Vector3{X: 0, Y: 0, Z: 0.01},
		Orientation:         &types.Orientation{},
		Mass:                &types.Mass{Value: 5.0},
		Time:                2.5,
		CurrentEvent:        types.Burnout,
		Motor:               &components.Motor{},
		Bodytube:            &components.Bodytube{},
		Nosecone:            &components.Nosecone{},
		Finset:              &components.TrapezoidFinset{},
		Parachute:           &components.Parachute{},
	}

	// Verify all components are present for a complete simulation
	require.NotNil(t, state.Position)
	require.NotNil(t, state.Velocity)
	require.NotNil(t, state.Acceleration)
	require.NotNil(t, state.AngularVelocity)
	require.NotNil(t, state.AngularAcceleration)
	require.NotNil(t, state.Orientation)
	require.NotNil(t, state.Mass)
	require.NotNil(t, state.Motor)
	require.NotNil(t, state.Bodytube)
	require.NotNil(t, state.Nosecone)
	require.NotNil(t, state.Finset)
	require.NotNil(t, state.Parachute)

	// Verify the rocket is in a reasonable flight state
	assert.Equal(t, 50.0, state.Velocity.Vec.Y)        // Moving upward
	assert.Equal(t, 10.0, state.Acceleration.Vec.Y)    // Accelerating upward
	assert.Equal(t, 5.0, state.Mass.Value)             // Has mass
	assert.Equal(t, types.Burnout, state.CurrentEvent) // In burnout phase
	assert.True(t, state.Time > 0.0)                   // Time has progressed
}

// TestPhysicsState_ZeroValues tests behavior with zero values
func TestPhysicsState_ZeroValues(t *testing.T) {
	state := &states.PhysicsState{
		Position:            &types.Position{Vec: types.Vector3{X: 0, Y: 0, Z: 0}},
		Velocity:            &types.Velocity{Vec: types.Vector3{X: 0, Y: 0, Z: 0}},
		Acceleration:        &types.Acceleration{Vec: types.Vector3{X: 0, Y: 0, Z: 0}},
		AngularVelocity:     &types.Vector3{X: 0, Y: 0, Z: 0},
		AngularAcceleration: &types.Vector3{X: 0, Y: 0, Z: 0},
		Mass:                &types.Mass{Value: 0},
		Time:                0,
	}

	// Test that zero values are handled correctly
	assert.Equal(t, 0.0, state.Position.Vec.X)
	assert.Equal(t, 0.0, state.Position.Vec.Y)
	assert.Equal(t, 0.0, state.Position.Vec.Z)

	assert.Equal(t, 0.0, state.Velocity.Vec.X)
	assert.Equal(t, 0.0, state.Velocity.Vec.Y)
	assert.Equal(t, 0.0, state.Velocity.Vec.Z)

	assert.Equal(t, 0.0, state.Acceleration.Vec.X)
	assert.Equal(t, 0.0, state.Acceleration.Vec.Y)
	assert.Equal(t, 0.0, state.Acceleration.Vec.Z)

	assert.Equal(t, 0.0, state.AngularVelocity.X)
	assert.Equal(t, 0.0, state.AngularVelocity.Y)
	assert.Equal(t, 0.0, state.AngularVelocity.Z)

	assert.Equal(t, 0.0, state.AngularAcceleration.X)
	assert.Equal(t, 0.0, state.AngularAcceleration.Y)
	assert.Equal(t, 0.0, state.AngularAcceleration.Z)

	assert.Equal(t, 0.0, state.Mass.Value)
	assert.Equal(t, 0.0, state.Time)
}
