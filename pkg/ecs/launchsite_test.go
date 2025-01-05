package ecs_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs"
)

// TEST: GIVEN a new Launchsite instance WHEN Describe is called THEN a string representation of the Launchsite is returned
func TestLaunchsiteDescribe(t *testing.T) {
	ls := ecs.NewLaunchsite(0.0, 0.0, 0.0)

	expected := "Lat: 0.00°, Lon: 0.00°, Alt: 0.00m"
	actual := ls.Describe()

	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}

}

// TEST: GIVEN a new Launchsite instance WHEN NewLaunchsite is called THEN a new Launchsite instance is returned with the correct values
func TestNewLaunchsite(t *testing.T) {
	ls := ecs.NewLaunchsite(0.0, 0.0, 0.0)

	if ls.Latitude != 0.0 {
		t.Errorf("Expected Latitude to be 0.0, got %f", ls.Latitude)
	}

	if ls.Longitude != 0.0 {
		t.Errorf("Expected Longitude to be 0.0, got %f", ls.Longitude)
	}

	if ls.Altitude != 0.0 {
		t.Errorf("Expected Altitude to be 0.0, got %f", ls.Altitude)
	}

}
