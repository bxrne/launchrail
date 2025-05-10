package components_test

import (
	"testing"
	"time"

	"github.com/bxrne/launchrail/pkg/components"
	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"
	"io"
)

// TEST: GIVEN a new MotorFSM WHEN InitialState is called THEN the state is "idle"
func TestMotorFSM_InitialState(t *testing.T) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	motor := &components.Motor{}
	fsm := components.NewMotorFSM(motor, logger)
	assert.Equal(t, components.StateIdle, fsm.GetState(), "FSM should start in 'idle' state")
}

// TEST: GIVEN a new MotorFSM WHEN StateTransitions is called THEN the state transitions are correct
func TestMotorFSM_StateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		mass        float64
		elapsedTime float64
		burnTime    float64
		initState   string
		wantState   string
	}{
		{"Idle to Burning", 10.0, 1.0, 5.0, components.StateIdle, components.StateBurning},
		{"Burning to Coasting (Time Exceeded)", 10.0, 6.0, 5.0, components.StateBurning, components.StateCoasting},
		{"Burning to Coasting (No Mass)", 0.0, 3.0, 5.0, components.StateBurning, components.StateCoasting},
		{"Stay in Idle", 0.0, 0.0, 5.0, components.StateIdle, components.StateIdle},
		{"Stay in Burning", 10.0, 3.0, 5.0, components.StateBurning, components.StateBurning},
		{"Edge Case - Exact Burn Time", 10.0, 5.0, 5.0, components.StateBurning, components.StateCoasting},
		{"Edge Case - Zero Mass at Start", 0.0, 0.0, 5.0, components.StateIdle, components.StateIdle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logf.New(logf.Opts{Writer: io.Discard})
			motor := &components.Motor{}
			fsm := components.NewMotorFSM(motor, logger)

			// Set initial state if needed
			if tt.initState == components.StateBurning {
				_ = fsm.UpdateState(10.0, 1.0, 5.0) // Force to burning state
			}

			err := fsm.UpdateState(tt.mass, tt.elapsedTime, tt.burnTime)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantState, fsm.GetState())
		})
	}
}

// TEST: GIVEN a new MotorFSM WHEN ConcurrentUpdates is called THEN the state is valid
func TestMotorFSM_ConcurrentUpdates(t *testing.T) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	motor := &components.Motor{}
	fsm := components.NewMotorFSM(motor, logger)

	// Run multiple updates concurrently
	for i := 0; i < 10; i++ {
		go func() {
			_ = fsm.UpdateState(10.0, 1.0, 5.0)
		}()
	}

	// Allow time for goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// State should be valid
	state := fsm.GetState()
	assert.Contains(t, []string{components.StateIdle, components.StateBurning, components.StateCoasting}, state)
}

// TEST: GIVEN a new MotorFSM WHEN NegativeValues is called THEN the state is valid
func TestMotorFSM_NegativeValues(t *testing.T) {
	tests := []struct {
		name        string
		mass        float64
		elapsedTime float64
		burnTime    float64
	}{
		{"Negative Mass", -1.0, 1.0, 5.0},
		{"Negative Elapsed Time", 10.0, -1.0, 5.0},
		{"Negative Burn Time", 10.0, 1.0, -5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logf.New(logf.Opts{Writer: io.Discard})
			motor := &components.Motor{}
			fsm := components.NewMotorFSM(motor, logger)
			err := fsm.UpdateState(tt.mass, tt.elapsedTime, tt.burnTime)
			assert.NoError(t, err) // Should handle negative values gracefully
		})
	}
}

// TEST: GIVEN a new MotorFSM WHEN InvalidState is called THEN the state is valid
func TestMotorFSM_StateString(t *testing.T) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	motor := &components.Motor{}
	fsm := components.NewMotorFSM(motor, logger)
	assert.Equal(t, "idle", fsm.GetState())

	_ = fsm.UpdateState(10.0, 1.0, 5.0)
	assert.Equal(t, "burning", fsm.GetState())
}

// TEST: GIVEN a new MotorFSM WHEN RapidStateChanges is called THEN the state is valid
func TestMotorFSM_RapidStateChanges(t *testing.T) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	motor := &components.Motor{}
	fsm := components.NewMotorFSM(motor, logger)

	// Rapidly toggle between states
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			_ = fsm.UpdateState(10.0, 1.0, 5.0) // Should transition to burning
		} else {
			_ = fsm.UpdateState(0.0, 6.0, 5.0) // Should transition to coast
		}
	}

	// Final state should be valid
	state := fsm.GetState()
	assert.Contains(t, []string{components.StateIdle, components.StateBurning, components.StateCoasting}, state)
}
