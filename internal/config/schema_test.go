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
		Logging: struct {
			Level string `mapstructure:"level"`
		}{
			Level: "info",
		},
		Options: struct {
			MotorDesignation string `mapstructure:"motor_designation"`
			OpenRocketFile   string `mapstructure:"openrocket_file"`
		}{
			MotorDesignation: "G80-7T",
			OpenRocketFile:   "test/fixtures/rocket.ork",
		},
	}

	expected := map[string]string{
		"app.name":                  "launchrail-test",
		"app.version":               "0.0.0",
		"logging.level":             "info",
		"options.motor_designation": "G80-7T",
		"options.openrocket_file":   "test/fixtures/rocket.ork",
	}

	actual := cfg.String()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
