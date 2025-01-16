package components

import (
	"context"
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
	f := &MotorFSM{
		fsm: fsm.NewFSM(
			StateIdle,
			fsm.Events{
				{Name: "ignite", Src: []string{StateIdle}, Dst: StateBurning},
				{Name: "extinguish", Src: []string{StateBurning}, Dst: StateIdle},
			},
			fsm.Callbacks{},
		),
	}
	return f
}

// UpdateState updates the state based on mass and elapsed time
func (fsm *MotorFSM) UpdateState(mass float64, elapsedTime float64, burnTime float64) {
	ctx := context.Background() // Create a background context

	if mass > 0 && elapsedTime <= burnTime {
		fsm.fsm.Event(ctx, "ignite")
	} else {
		fsm.fsm.Event(ctx, "extinguish")
	}
}

// GetState returns the current state of the FSM
func (fsm *MotorFSM) GetState() string {
	return fsm.fsm.Current()
}
