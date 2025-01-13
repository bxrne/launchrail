package ecs_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/entities"
	"github.com/bxrne/launchrail/pkg/ecs/systems"
)

// TEST: GIVEN nothing WHEN a NewWorld is called THEN a new World instance is returned
func TestNewWorld(t *testing.T) {
	r := entities.NewRocket(1.0)
	w := ecs.NewWorld(r)

	if w == nil {
		t.Errorf("Expected World instance, got nil")
	}
}

// TEST: GIVEN a World instance and a Rocket Entity WHEN String is called THEN a string representation of the World is returned
func TestWorld_String(t *testing.T) {
	r := entities.NewRocket(1.0)
	w := ecs.NewWorld(r)

	expected := "1 entities, 0 components, and 0 systems"
	if w.String() != expected {
		t.Errorf("Expected %s, got %s", expected, w.String())
	}
}

// TEST: GIVEN a World instance and a Rocket Entity WHEN Describe is called THEN a string representation of the World is returned
func TestWorld_Describe(t *testing.T) {
	r := entities.NewRocket(1.0)
	w := ecs.NewWorld(r)

	expected := "1 entities, 0 components, and 0 systems"
	if w.Describe() != expected {
		t.Errorf("Expected %s, got %s", expected, w.Describe())
	}
}

// TEST: GIVEN a World instance and a Rocket Entity WHEN AddEntity is called THEN the entity is added to the NewWorld
func TestWorld_AddEntity(t *testing.T) {
	r := entities.NewRocket(1.0)
	w := ecs.NewWorld(r)

	w.AddEntity(r)
	expected := "2 entities, 0 components, and 0 systems"
	if w.Describe() != expected {
		t.Errorf("Expected %s, got %s", expected, w.Describe())
	}
}

// TEST: GIVEN a World instance and a Component WHEN AddComponent is called THEN the component is added to the TestNewWorld
func TestWorld_AddComponent(t *testing.T) {
	r := entities.NewRocket(1.0)
	w := ecs.NewWorld(r)

	c := components.NewMockComponent("mock")
	w.AddComponent(c)
	expected := "1 entities, 1 components, and 0 systems"
	if w.Describe() != expected {
		t.Errorf("Expected %s, got %s", expected, w.Describe())
	}
}

// TEST: GIVEN a World instance and a System WHEN AddSystem is called THEN the system is added to the TestNewWorld
func TestWorld_AddSystem(t *testing.T) {
	r := entities.NewRocket(1.0)
	w := ecs.NewWorld(r)

	s := systems.NewMockSystem(1)
	w.AddSystem(s)
	expected := "1 entities, 0 components, and 1 systems"
	if w.Describe() != expected {
		t.Errorf("Expected %s, got %s", expected, w.Describe())
	}
}

// TEST: GIVEN a World instance and a System WHEN Update is called THEN the system is updated
func TestWorld_Update(t *testing.T) {
	r := entities.NewRocket(1.0)
	w := ecs.NewWorld(r)

	s := systems.NewMockSystem(1)
	w.AddSystem(s)
	w.Update(1.0)
}
