package types_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
)

func TestEventString(t *testing.T) {
	tests := []struct {
		event types.Event
		want  string
	}{
		{types.None, "NONE"},
		{types.Liftoff, "LIFTOFF"},
		{types.Apogee, "APOGEE"},
		{types.Land, "LAND"},
		{types.ParachuteDeploy, "PARACHUTE_DEPLOY"},
		{types.Burnout, "BURNOUT"},
		{types.Event(99), "UNKNOWN_EVENT(99)"}, // Test unknown event
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.event.String(); got != tt.want {
				t.Errorf("Event(%d).String() = %q, want %q", tt.event, got, tt.want)
			}
		})
	}
}
