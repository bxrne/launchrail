package components

import (
	"context"

	"github.com/looplab/fsm"
	"github.com/zerodha/logf"
)

// MotorState represents the states of the motor
const (
	StateIdle    = "idle"
	StateBurning = "burning"
	StateIgnited = "IGNITED"
)

// MotorFSM represents the finite state machine for the motor
type MotorFSM struct {
	*fsm.FSM
	motor *Motor // Reference to the motor component it controls
	log   logf.Logger
}

// NewMotorFSM creates a new FSM for the motor
func NewMotorFSM(motor *Motor, log logf.Logger) *MotorFSM {
	return &MotorFSM{
		FSM: fsm.NewFSM(
			string(StateIdle), // Set initial state to "idle"
			fsm.Events{
				{Name: "ignite", Src: []string{string(StateIdle)}, Dst: string(StateIgnited)},
				{Name: "burnout", Src: []string{string(StateBurning)}, Dst: string(StateIdle)},
				{Name: "start_burning", Src: []string{string(StateIgnited)}, Dst: string(StateBurning)},
			},
			fsm.Callbacks{},
		),
		motor: motor,
		log:   log,
	}
}

// handleBurningTransition manages FSM transitions during the burning period.
func (fsm *MotorFSM) handleBurningTransition(ctx context.Context, state string) error {
	switch state {
	case StateIdle:
		if err := fsm.triggerEvent(ctx, "ignite"); err != nil {
			return err
		}
		return fsm.triggerEvent(ctx, "start_burning")
	case StateIgnited:
		return fsm.triggerEvent(ctx, "start_burning")
	}
	return nil
}

// triggerEvent wraps FSM.Event with error propagation.
func (fsm *MotorFSM) triggerEvent(ctx context.Context, event string) error {
	if err := fsm.FSM.Event(ctx, event); err != nil {
		return err
	}
	return nil
}

// UpdateState updates the state based on elapsed time only
func (fsm *MotorFSM) UpdateState(mass float64, elapsedTime float64, burnTime float64) error {
	if elapsedTime < 0 {
		elapsedTime = 0
	}
	if burnTime < 0 {
		burnTime = 0
	}

	ctx := context.Background()
	currentState := fsm.FSM.Current()

	if elapsedTime < burnTime && mass > 0 {
		return fsm.handleBurningTransition(ctx, currentState)
	}

	if currentState == StateBurning {
		return fsm.triggerEvent(ctx, "burnout")
	}
	return nil
}

// GetState returns the current state of the FSM
func (fsm *MotorFSM) GetState() string {
	return fsm.Current()
}
