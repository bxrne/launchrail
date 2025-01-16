package ecs_test

import (
	"os"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/entities"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

// TEST: GIVEN a new ECS instance WHEN Describe is called THEN a string representation of the ECS is returned
func TestECSDescribe(t *testing.T) {
	e := &ecs.ECS{
		World:      ecs.NewWorld(entities.NewRocket(1.0)),
		Launchrail: ecs.NewLaunchrail(0.0, 0.0, 0.0),
		Launchsite: ecs.NewLaunchsite(0.0, 0.0, 0.0),
	}

	expected := "Rail: Len: 0.00m, Angle: 0.00°, Orient: 0.00°, Site: Lat: 0.00°, Lon: 0.00°, Alt: 0.00m, World: 1 entities and 0 systems"
	actual := e.Describe()

	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}

}

// TEST: GIVEN a new ECS instance WHEN NewECS is called THEN a new ECS instance is returned
func TestNewECS(t *testing.T) {
	err := os.Chdir("../../")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	motor_data, err := thrustcurves.Load(cfg.Options.MotorDesignation, http_client.NewHTTPClient())
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	ork_data, err := openrocket.Load(cfg.Options.OpenRocketFile, cfg.External.OpenRocketVersion)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	ecs, err := ecs.NewECS(cfg, &ork_data.Rocket, motor_data)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	if ecs == nil {
		t.Errorf("Expected ECS instance, got nil")
	}

}
