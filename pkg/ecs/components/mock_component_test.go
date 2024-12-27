package components_test

import (
	"github.com/bxrne/launchrail/pkg/ecs/components"
	"testing"
)

// TEST: GIVEN a mock component WHEN Type is called THEN the type is returned.
func TestMockComponent_Type(t *testing.T) {
	mock := components.NewMockComponent("Mock")
	if got := mock.Type(); got != "Mock" {
		t.Errorf("MockComponent.Type() = %v, want %v", got, "Mock")
	}
}
