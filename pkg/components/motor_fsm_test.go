package components_test

import (
	"testing"
	"time"

	"github.com/bxrne/launchrail/pkg/components"
	"github.com/stretchr/testify/assert"
)

func TestMotorFSM_InitialState(t *testing.T) {
	fsm := components.NewMotorFSM()
	assert.Equal(t, components.StateIdle, fsm.GetState(), "FSM should start in 'idle' state")
}

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
		{"Burning to Idle (Time Exceeded)", 10.0, 6.0, 5.0, components.StateBurning, components.StateIdle},
		{"Burning to Idle (No Mass)", 0.0, 3.0, 5.0, components.StateBurning, components.StateIdle},
		{"Stay in Idle", 0.0, 0.0, 5.0, components.StateIdle, components.StateIdle},
		{"Stay in Burning", 10.0, 3.0, 5.0, components.StateBurning, components.StateBurning},
		{"Edge Case - Exact Burn Time", 10.0, 5.0, 5.0, components.StateBurning, components.StateIdle},
		{"Edge Case - Zero Mass at Start", 0.0, 0.0, 5.0, components.StateIdle, components.StateIdle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsm := components.NewMotorFSM()

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

func TestMotorFSM_ConcurrentUpdates(t *testing.T) {
	fsm := components.NewMotorFSM()

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
	assert.Contains(t, []string{components.StateIdle, components.StateBurning}, state)
}

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
			fsm := components.NewMotorFSM()
			err := fsm.UpdateState(tt.mass, tt.elapsedTime, tt.burnTime)
			assert.NoError(t, err) // Should handle negative values gracefully
		})
	}
}

func TestMotorFSM_StateString(t *testing.T) {
	fsm := components.NewMotorFSM()
	assert.Equal(t, "idle", fsm.GetState())

	_ = fsm.UpdateState(10.0, 1.0, 5.0)
	assert.Equal(t, "burning", fsm.GetState())
}

func TestMotorFSM_RapidStateChanges(t *testing.T) {
	fsm := components.NewMotorFSM()

	// Rapidly toggle between states
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			_ = fsm.UpdateState(10.0, 1.0, 5.0) // Should transition to burning
		} else {
			_ = fsm.UpdateState(0.0, 6.0, 5.0) // Should transition to idle
		}
	}

	// Final state should be valid
	state := fsm.GetState()
	assert.Contains(t, []string{components.StateIdle, components.StateBurning}, state)
}
