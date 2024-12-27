package systems_test

import (
	"github.com/bxrne/launchrail/pkg/ecs/systems"
	"testing"
)

// TEST: GIVEN no mock system WHEN NewMockSystem is called THEN a mock system is returned.
func TestNewMockSystem(t *testing.T) {
	mock := systems.NewMockSystem(0)
	if mock == nil {
		t.Errorf("NewMockSystem() = %v, want %v", mock, "MockSystem")
	}
}

// TEST: GIVEN a mock system WHEN Update is called THEN the system is updated.
func TestMockSystem_Update(t *testing.T) {
	mock := systems.NewMockSystem(0)
	mock.Update(1.0)
	if got := mock.GetUpdateCalled(); !got {
		t.Errorf("MockSystem.Update() = %v, want %v", got, true)
	}
}

// TEST: GIVEN a mock system WHEN Priority is called THEN the priority is returned.
func TestMockSystem_Priority(t *testing.T) {
	mock := systems.NewMockSystem(0)
	if got := mock.Priority(); got != 0 {
		t.Errorf("MockSystem.Priority() = %v, want %v", got, 0)
	}
}

// TEST: GIVEN a mock system WHEN Update is called THEN the delta time is returned.
func TestMockSystem_GetUpdateDt(t *testing.T) {
	mock := systems.NewMockSystem(0)
	mock.Update(1.0)
	if got := mock.GetUpdateDt(); got != 1.0 {
		t.Errorf("MockSystem.GetUpdateDt() = %v, want %v", got, 1.0)
	}
}

// TEST: GIVEN a mock system WHEN Update is called THEN the system is updated with the correct delta time.
func TestMockSystem_UpdateDt(t *testing.T) {
	mock := systems.NewMockSystem(0)
	mock.Update(1.0)
	if got := mock.GetUpdateDt(); got != 1.0 {
		t.Errorf("MockSystem.Update() = %v, want %v", got, 1.0)
	}
}
