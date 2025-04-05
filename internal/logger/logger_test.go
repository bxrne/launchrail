package logger_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
)

// Barebones config for testing
var cfg = &config.Config{
	Setup: config.Setup{
		Logging: config.Logging{
			Level: "info",
		},
	},
}

// TEST: GIVEN GetLogger is called THEN a non-nil logger is returned
func TestGetLogger(t *testing.T) {
	log := logger.GetLogger(cfg)
	if log == nil {
		t.Error("Expected logger to be non-nil")
	}
}

// TEST: GIVEN GetLogger is called multiple times THEN the logger is a singleton
func TestGetLoggerSingleton(t *testing.T) {
	log1 := logger.GetLogger(cfg)
	log2 := logger.GetLogger(cfg)

	if log1 != log2 {
		t.Error("Expected logger to be a singleton")
	}
}
