package components

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

// MotorFSM represents the finite state machine for the motor
type MotorFSM struct {
	*fsm.FSM
}

// NewMotorFSM creates a new FSM for the motor
func NewMotorFSM() *MotorFSM {
	return &MotorFSM{
		FSM: fsm.NewFSM(
			string(StateIdle), // Set initial state to "idle"
			fsm.Events{
				{Name: "ignite", Src: []string{string(StateIdle)}, Dst: string(StateBurning)},
				{Name: "burnout", Src: []string{string(StateBurning)}, Dst: string(StateIdle)},
			},
			fsm.Callbacks{},
		),
	}
}

// UpdateState updates the state based on mass and elapsed time
func (fsm *MotorFSM) UpdateState(mass float64, elapsedTime float64, burnTime float64) error {
	ctx := context.Background() // Create a background context
	currentState := fsm.Current()

	// Force the motor to go idle when empty or time exceeded
	if mass <= 0 || elapsedTime >= burnTime {
		return fsm.handlePotentiallyInactiveState(ctx, currentState)
	}

	// Use strictly less than for active state
	if mass > 0 && elapsedTime < burnTime {
		return fsm.handlePotentiallyActiveState(ctx, currentState)
	}
	return fsm.handlePotentiallyInactiveState(ctx, currentState)
}

// handlePotentiallyActiveState handles state transitions when the motor is active
func (fsm *MotorFSM) handlePotentiallyActiveState(ctx context.Context, currentState string) error {
	switch currentState {
	case StateIdle:
		return fsm.Event(ctx, "ignite")
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
		return fsm.Event(ctx, "burnout")
	case StateIdle:
		// Already in idle state, no action needed
		return nil
	default:
		return fmt.Errorf("invalid state: %s", currentState)
	}
}

// GetState returns the current state of the FSM
func (fsm *MotorFSM) GetState() string {
	return fsm.Current()
}
