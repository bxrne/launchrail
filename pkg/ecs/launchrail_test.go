package ecs_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs"
)

// TEST: GIVEN params WHEN NewLaunchrail is called THEN a new Launchrail instance is returned with the correct values
func TestNewLaunchrail(t *testing.T) {
	lr := ecs.NewLaunchrail(0.0, 0.0, 0.0)

	if lr.Length != 0.0 {
		t.Errorf("Expected Length to be 0.0, got %f", lr.Length)
	}

	if lr.Angle != 0.0 {
		t.Errorf("Expected Angle to be 0.0, got %f", lr.Angle)
	}

	if lr.Orientation != 0.0 {
		t.Errorf("Expected Orientation to be 0.0, got %f", lr.Orientation)
	}

}

// TEST: GIVEN a new Launchrail instance WHEN Describe is called THEN a string representation of the Launchrail is returned
func TestLaunchrailDescribe(t *testing.T) {
	lr := ecs.NewLaunchrail(0.0, 0.0, 0.0)

	expected := "Len: 0.00m, Angle: 0.00°, Orient: 0.00°"
	actual := lr.Describe()

	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}

}
