package config_test

import (
	"reflect"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
)

// TEST: GIVEN a populated configuration WHEN String is called THEN a map of configuration values is returned
func TestConfigString(t *testing.T) {
	cfg := config.Config{
		App: config.App{
			Name:    "launchrail-test",
			Version: "0.0.0",
		},
		Logging: config.Logging{
			Level: "info",
		},
		External: config.External{
			OpenRocketVersion: "15.03",
		},
		Options: config.Options{
			MotorDesignation: "G80-7T",
			OpenRocketFile:   "test/fixtures/rocket.ork",
			Launchrail: config.Launchrail{
				Length:      0.00,
				Angle:       0.00,
				Orientation: 0.00,
			},
			Launchsite: config.Launchsite{
				Latitude:  0.00,
				Longitude: 0.00,
				Altitude:  0.00,
			},
		},
		Simulation: config.Simulation{
			Step:    0.00,
			MaxTime: 0.00,
		},
	}

	expected := map[string]string{
		"app.name":                       "launchrail-test",
		"app.version":                    "0.0.0",
		"logging.level":                  "info",
		"external.openrocket_version":    "15.03",
		"options.motor_designation":      "G80-7T",
		"options.openrocket_file":        "test/fixtures/rocket.ork",
		"options.launchrail.length":      "0.00",
		"options.launchrail.angle":       "0.00",
		"options.launchrail.orientation": "0.00",
		"options.launchsite.latitude":    "0.00",
		"options.launchsite.longitude":   "0.00",
		"options.launchsite.altitude":    "0.00",
		"simulation.step":                "0.00",
		"simulation.max_time":            "0.00",
	}

	actual := cfg.String()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
