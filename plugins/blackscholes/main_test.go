package main

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"
)

// Helper to create a default logger for tests
func newTestLogger() logf.Logger {
	opts := logf.Opts{EnableCaller: true, Level: logf.DebugLevel}
	return logf.New(opts)
}

// Helper to create a basic physics state for testing
func newTestPhysicsState(vx, vy, vz float64) *states.PhysicsState {
	return &states.PhysicsState{
		Velocity: &types.Velocity{
			Vec: types.Vector3{X: vx, Y: vy, Z: vz},
		},
		// Add other necessary fields if the plugin uses them, initialized appropriately
		Time: 0.0,
	}
}

func TestInitialize(t *testing.T) {
	plugin := &BlackScholesPlugin{}
	logger := newTestLogger()
	err := plugin.Initialize(logger)

	assert.NoError(t, err, "Initialize should not return an error")
	assert.NotNil(t, plugin.log, "Logger should be initialized")
	assert.NotNil(t, plugin.rng, "Random number generator should be initialized")
	assert.Equal(t, 0.05, plugin.turbulenceIntensity, "Default turbulence intensity should be set") // Check default
}

func TestNameVersion(t *testing.T) {
	plugin := &BlackScholesPlugin{}
	assert.Equal(t, "blackscholes", plugin.Name(), "Name should return 'blackscholes'")
	assert.Equal(t, "0.1.1", plugin.Version(), "Version should return '0.1.1'")
}

func TestCleanup(t *testing.T) {
	plugin := &BlackScholesPlugin{}
	logger := newTestLogger()
	_ = plugin.Initialize(logger) // Initialize first
	err := plugin.Cleanup()
	assert.NoError(t, err, "Cleanup should not return an error")
}

func TestAfterSimStep_ChangesState(t *testing.T) {
	plugin := &BlackScholesPlugin{}
	logger := newTestLogger()
	_ = plugin.Initialize(logger)
	plugin.turbulenceIntensity = 0.1 // Ensure non-zero intensity

	initialState := newTestPhysicsState(10.0, 20.0, 30.0)
	initialVelocity := initialState.Velocity.Vec // Copy initial velocity

	err := plugin.AfterSimStep(initialState)
	assert.NoError(t, err, "AfterSimStep should not return an error")

	// Check that velocity actually changed
	assert.NotEqual(t, initialVelocity, initialState.Velocity.Vec, "Velocity vector should change after step")
}

func TestAfterSimStep_IsRandom(t *testing.T) {
	plugin := &BlackScholesPlugin{}
	logger := newTestLogger()
	_ = plugin.Initialize(logger)
	plugin.turbulenceIntensity = 0.1 // Ensure non-zero intensity

	// Run 1
	state1 := newTestPhysicsState(10.0, 20.0, 30.0)
	err1 := plugin.AfterSimStep(state1)
	assert.NoError(t, err1)
	velocity1 := state1.Velocity.Vec

	// Run 2 - Re-initialize state to be identical to the start of Run 1
	state2 := newTestPhysicsState(10.0, 20.0, 30.0)
	err2 := plugin.AfterSimStep(state2)
	assert.NoError(t, err2)
	velocity2 := state2.Velocity.Vec

	// Because the RNG was seeded once and is used sequentially,
	// successive calls *should* produce different results.
	assert.NotEqual(t, velocity1, velocity2, "Successive calls with same initial state should produce different results due to RNG")
}

func TestAfterSimStep_ZeroIntensity(t *testing.T) {
	plugin := &BlackScholesPlugin{}
	logger := newTestLogger()
	_ = plugin.Initialize(logger)
	plugin.turbulenceIntensity = 0.0 // Set intensity to zero

	initialState := newTestPhysicsState(10.0, 20.0, 30.0)
	initialVelocity := initialState.Velocity.Vec

	err := plugin.AfterSimStep(initialState)
	assert.NoError(t, err, "AfterSimStep should not return an error")

	// Check that velocity did NOT change (noise terms should be zero)
	// Need tolerance due to potential float precision issues, although NormFloat64 * 0 should be exactly 0
	assert.InDelta(t, initialVelocity.X, initialState.Velocity.Vec.X, 1e-9, "Velocity X should not change with zero intensity")
	assert.InDelta(t, initialVelocity.Y, initialState.Velocity.Vec.Y, 1e-9, "Velocity Y should not change with zero intensity")
	assert.InDelta(t, initialVelocity.Z, initialState.Velocity.Vec.Z, 1e-9, "Velocity Z should not change with zero intensity")
	// Or assert the whole vector equality
	assert.Equal(t, initialVelocity, initialState.Velocity.Vec, "Velocity vector should not change with zero intensity")
}

// Optional: Test the scaling effect (harder to assert precisely)
func TestAfterSimStep_IntensityScaling(t *testing.T) {
	logger := newTestLogger()

	// Low intensity run
	pluginLow := &BlackScholesPlugin{}
	_ = pluginLow.Initialize(logger)
	pluginLow.turbulenceIntensity = 0.01
	stateLow := newTestPhysicsState(100.0, 0.0, 0.0) // High initial speed for effect
	initialVelocityLow := stateLow.Velocity.Vec
	_ = pluginLow.AfterSimStep(stateLow)
	deltaLowX := math.Abs(stateLow.Velocity.Vec.X - initialVelocityLow.X)
	deltaLowY := math.Abs(stateLow.Velocity.Vec.Y - initialVelocityLow.Y)
	deltaLowZ := math.Abs(stateLow.Velocity.Vec.Z - initialVelocityLow.Z)
	magnitudeDeltaLow := math.Sqrt(deltaLowX*deltaLowX + deltaLowY*deltaLowY + deltaLowZ*deltaLowZ)


	// High intensity run
	// Re-initialize plugin to reset RNG sequence for a *comparable* (though not identical) draw
	pluginHigh := &BlackScholesPlugin{}
	_ = pluginHigh.Initialize(logger) // Needs a different seed ideally, but re-init is baseline
	pluginHigh.turbulenceIntensity = 0.5 // Much higher intensity
	stateHigh := newTestPhysicsState(100.0, 0.0, 0.0) // Same initial state
	initialVelocityHigh := stateHigh.Velocity.Vec
	_ = pluginHigh.AfterSimStep(stateHigh)
	deltaHighX := math.Abs(stateHigh.Velocity.Vec.X - initialVelocityHigh.X)
	deltaHighY := math.Abs(stateHigh.Velocity.Vec.Y - initialVelocityHigh.Y)
	deltaHighZ := math.Abs(stateHigh.Velocity.Vec.Z - initialVelocityHigh.Z)
	magnitudeDeltaHigh := math.Sqrt(deltaHighX*deltaHighX + deltaHighY*deltaHighY + deltaHighZ*deltaHighZ)

	// We expect the magnitude of the change to be generally larger with higher intensity
	// This isn't a perfect test due to randomness, but checks the trend.
	assert.Greater(t, magnitudeDeltaHigh, magnitudeDeltaLow, "Magnitude of velocity change should generally be greater with higher turbulence intensity")
	t.Logf("Magnitude change (Low Intensity): %f", magnitudeDeltaLow)
	t.Logf("Magnitude change (High Intensity): %f", magnitudeDeltaHigh)

}
