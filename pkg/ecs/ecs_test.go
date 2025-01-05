package ecs_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/ecs"
)

// TEST: GIVEN a new ECS instance WHEN Describe is called THEN a string representation of the ECS is returned
func TestECSDescribe(t *testing.T) {
	e := &ecs.ECS{
		World:      ecs.NewWorld(),
		Launchrail: ecs.NewLaunchrail(0.0, 0.0, 0.0),
		Launchsite: ecs.NewLaunchsite(0.0, 0.0, 0.0),
	}

	expected := "Rail: Len: 0.00m, Angle: 0.00°, Orient: 0.00°, Site: Lat: 0.00°, Lon: 0.00°, Alt: 0.00m"
	actual := e.Describe()

	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}

}

// TEST: GIVEN a new ECS instance WHEN NewECS is called THEN a new ECS instance is returned
func TestNewECS(t *testing.T) {
	cfg, err := config.GetConfig()
	if err != nil {
		t.Errorf("Failed to get configuration: %v", err)
	}

	_, err = ecs.NewECS(cfg)
	if err != nil {
		t.Errorf("Failed to create ECS: %v", err)
	}
}
