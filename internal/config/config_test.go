package config_test

import (
	"github.com/bxrne/launchrail/internal/config"
	"testing"
)

func TestGetConfig(t *testing.T) {
	cfg := config.GetConfig("../../testdata/config.yaml")
	if cfg == nil {
		t.Error("Expected config to be non-nil")
	}
}

func TestGetConfigSingleton(t *testing.T) {
	cfg1 := config.GetConfig("../../testdata/config.yaml")
	cfg2 := config.GetConfig("../../testdata/config.yaml")

	if cfg1 != cfg2 {
		t.Error("Expected config to be a singleton")
	}
}
