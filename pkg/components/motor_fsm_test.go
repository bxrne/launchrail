package components_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/components"
	"github.com/stretchr/testify/assert"
)

func TestMotorFSM_InitialState(t *testing.T) {
	fsm := components.NewMotorFSM()
	assert.Equal(t, components.StateIdle, fsm.GetState(), "FSM should start in 'idle' state")
}

func TestMotorFSM_TransitionToBurning(t *testing.T) {
	fsm := components.NewMotorFSM()
	err := fsm.UpdateState(10.0, 1.0, 5.0)
	assert.NoError(t, err, "Transition to 'burning' should not produce an error")
	assert.Equal(t, components.StateBurning, fsm.GetState(), "FSM should transition to 'burning' state")
}

func TestMotorFSM_TransitionToIdle(t *testing.T) {
	fsm := components.NewMotorFSM()
	_ = fsm.UpdateState(10.0, 1.0, 5.0)
	assert.Equal(t, components.StateBurning, fsm.GetState(), "FSM should be in 'burning' state")

	err := fsm.UpdateState(0.0, 6.0, 5.0)
	assert.NoError(t, err, "Transition to 'idle' should not produce an error")
	assert.Equal(t, components.StateIdle, fsm.GetState(), "FSM should transition back to 'idle' state")
}

func TestMotorFSM_NoTransition(t *testing.T) {
	fsm := components.NewMotorFSM()
	err := fsm.UpdateState(0.0, 0.0, 5.0)
	assert.NoError(t, err, "No transition should not produce an error")
	assert.Equal(t, components.StateIdle, fsm.GetState(), "FSM should remain in 'idle' state")

	_ = fsm.UpdateState(10.0, 1.0, 5.0)
	assert.Equal(t, components.StateBurning, fsm.GetState(), "FSM should transition to 'burning' state")

	err = fsm.UpdateState(10.0, 2.0, 5.0)
	assert.NoError(t, err, "No transition should not produce an error")
	assert.Equal(t, components.StateBurning, fsm.GetState(), "FSM should remain in 'burning' state")
}

func TestMotorFSM_InvalidState(t *testing.T) {
	fsm := components.NewMotorFSM()
	err := fsm.UpdateState(-1.0, 1.0, 5.0)
	assert.NoError(t, err, "Invalid state should not produce an error")
}
