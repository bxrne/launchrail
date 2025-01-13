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

// TEST: GIVEN a mock component and a delta time WHEN Update is called THEN the component is updated.
func TestMockComponent_Update(t *testing.T) {
	mock := components.NewMockComponent("Mock")
	mock.Update(1.0)
}

// TEST: GIVEN nothing WHEN NewMockComponent is called THEN a new MockComponent instance is returned.
func TestNewMockComponent(t *testing.T) {
	mock := components.NewMockComponent("Mock")
	if mock == nil {
		t.Errorf("Expected MockComponent instance, got nil")
	}
}
