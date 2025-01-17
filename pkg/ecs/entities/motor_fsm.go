package entities

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
)

// MotorState represents the states of the motor
const (
	StateIdle    = "idle"
	StateBurning = "burning"
)

// MotorFSM manages the state of the motor
type MotorFSM struct {
	fsm *fsm.FSM
}

// NewMotorFSM creates a new MotorFSM instance
func NewMotorFSM() *MotorFSM {
	return &MotorFSM{
		fsm: fsm.NewFSM(
			StateIdle,
			fsm.Events{
				{Name: "ignite", Src: []string{StateIdle}, Dst: StateBurning},
				{Name: "extinguish", Src: []string{StateBurning}, Dst: StateIdle},
			},
			fsm.Callbacks{},
		),
	}
}

// UpdateState updates the state based on mass and elapsed time
func (fsm *MotorFSM) UpdateState(mass float64, elapsedTime float64, burnTime float64) error {
	ctx := context.Background() // Create a background context
	currentState := fsm.fsm.Current()

	if mass > 0 && elapsedTime <= burnTime {
		return fsm.handlePotentiallyActiveState(ctx, currentState)
	}
	return fsm.handlePotentiallyInactiveState(ctx, currentState)
}

// handlePotentiallyActiveState handles state transitions when the motor is active
func (fsm *MotorFSM) handlePotentiallyActiveState(ctx context.Context, currentState string) error {
	switch currentState {
	case StateIdle:
		return fsm.fsm.Event(ctx, "ignite")
	case StateBurning:
		// Already in burning state, no action needed
		return nil
	default:
		return fmt.Errorf("invalid state: %s", currentState)
	}
}

// handlePotentiallyInactiveState handles state transitions when the motor is inactive
func (fsm *MotorFSM) handlePotentiallyInactiveState(ctx context.Context, currentState string) error {
	switch currentState {
	case StateBurning:
		return fsm.fsm.Event(ctx, "extinguish")
	case StateIdle:
		// Already in idle state, no action needed
		return nil
	default:
		return fmt.Errorf("invalid state: %s", currentState)
	}
}

// GetState returns the current state of the FSM
func (fsm *MotorFSM) GetState() string {
	return fsm.fsm.Current()
}
