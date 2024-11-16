package simulation

import "github.com/bxrne/launchrail/pkg/types"

type Event interface {
	Check(state types.State) bool
	Handle(state *types.State)
}

type ApogeeEvent struct {
	// TODO: Implementation
}

type BurnoutEvent struct {
	// TODO: Implementation
}
