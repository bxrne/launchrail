package config_test

import (
	"github.com/bxrne/launchrail/internal/config"
	"testing"
)

var (
	configPath = "../../testdata/config.yaml"
)

func TestGetConfig(t *testing.T) {
	cfg := config.GetConfig(configPath)
	if cfg == nil {
		t.Error("Expected config to be non-nil")
	}
}

func TestGetConfigSingleton(t *testing.T) {
	cfg1 := config.GetConfig(configPath)
	cfg2 := config.GetConfig(configPath)

	if cfg1 != cfg2 {
		t.Error("Expected config to be a singleton")
	}
}
