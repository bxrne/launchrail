package types_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
)

// Parameters for testing
const (
	testEvent types.Event = types.Liftoff
)

func TestEventString(t *testing.T) {
	if testEvent.String() != "LIFTOFF" {
		t.Errorf("Event.String() = %s, want LIFTOFF", testEvent.String())
	}
}
