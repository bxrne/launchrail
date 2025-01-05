package config_test

import (
	"reflect"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
)

// TEST: GIVEN a populated configuration WHEN String is called THEN a map of configuration values is returned
func TestConfigString(t *testing.T) {
	cfg := &config.Config{
		App: struct {
			Name    string `mapstructure:"name"`
			Version string `mapstructure:"version"`
		}{
			Name:    "launchrail-test",
			Version: "0.0.0",
		},
		External: struct {
			OpenRocketVersion string `mapstructure:"openrocket_version"`
		}{
			OpenRocketVersion: "15.03",
		},
		Logging: struct {
			Level string `mapstructure:"level"`
		}{
			Level: "info",
		},
		Options: struct {
			MotorDesignation string `mapstructure:"motor_designation"`
			OpenRocketFile   string `mapstructure:"openrocket_file"`
			Launchrail       struct {
				Length      float64 `mapstructure:"length"`
				Angle       float64 `mapstructure:"angle"`
				Orientation float64 `mapstructure:"orientation"`
			} `mapstructure:"launchrail"`
			Launchsite struct {
				Latitude  float64 `mapstructure:"latitude"`
				Longitude float64 `mapstructure:"longitude"`
				Altitude  float64 `mapstructure:"altitude"`
			} `mapstructure:"launchsite"`
		}{
			MotorDesignation: "G80-7T",
			OpenRocketFile:   "test/fixtures/rocket.ork",
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
	}

	actual := cfg.String()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
