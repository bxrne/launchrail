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
	StateIgnited = "IGNITED"
)

// MotorFSM represents the finite state machine for the motor
type MotorFSM struct {
	*fsm.FSM
	elapsedTime float64
}

// NewMotorFSM creates a new FSM for the motor
func NewMotorFSM() *MotorFSM {
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
	}
}

// UpdateState updates the state based on elapsed time only
func (fsm *MotorFSM) UpdateState(mass float64, elapsedTime float64, burnTime float64) error {
	// Clamp negative values
	if elapsedTime < 0 {
		elapsedTime = 0
	}
	if burnTime < 0 {
		burnTime = 0
	}

	ctx := context.Background()
	currentState := fsm.FSM.Current()

	// Active burning period
	if elapsedTime < burnTime && mass > 0 {
		switch currentState {
		case StateIdle:
			// Idle -> Burning
			if err := fsm.FSM.Event(ctx, "ignite"); err != nil {
				return err
			}
			if err := fsm.FSM.Event(ctx, "start_burning"); err != nil {
				return err
			}
		case StateIgnited:
			// Ignited -> Burning
			if err := fsm.FSM.Event(ctx, "start_burning"); err != nil {
				return err
			}
		}
		return nil
	}

	// Transition to idle after burn or if mass depleted
	if currentState == StateBurning {
		return fsm.FSM.Event(ctx, "burnout")
	}
	return nil
}

// handlePotentiallyActiveState handles state transitions when the motor is active
func (fsm *MotorFSM) handlePotentiallyActiveState(ctx context.Context, currentState string) error {
	switch currentState {
	case StateIdle:
		return fsm.Event(ctx, "ignite")
	case StateIgnited:
		return fsm.Event(ctx, "start_burning")
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
	case StateIgnited:
		return fsm.Event(ctx, "burnout")
	default:
		return fmt.Errorf("invalid state: %s", currentState)
	}
}

// GetState returns the current state of the FSM
func (fsm *MotorFSM) GetState() string {
	return fsm.Current()
}
