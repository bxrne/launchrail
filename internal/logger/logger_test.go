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
	log := logger.GetLogger(cfg.Setup.Logging.Level)
	if log == nil {
		t.Error("Expected logger to be non-nil")
	}
}

// TEST: GIVEN GetLogger is called multiple times THEN the logger is a singleton
func TestGetLoggerSingleton(t *testing.T) {
	log1 := logger.GetLogger(cfg.Setup.Logging.Level)
	log2 := logger.GetLogger(cfg.Setup.Logging.Level)

	if log1 != log2 {
		t.Error("Expected logger to be a singleton")
	}
}

// TEST: GIVEN GetLogger is called with different levels THEN the logger level is set correctly
func TestGetLoggerDifferentLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "fatal"}

	for _, level := range levels {
		logger.Reset() // Reset the logger to ensure a fresh instance
		cfg.Setup.Logging.Level = level
		log := logger.GetLogger(level)
		if log == nil {
			t.Errorf("Expected logger to be non-nil for level %s", level)
		}

		if log.Level.String() != level {
			t.Errorf("Expected logger level to be %s, got %s", level, log.Level.String())
		}
	}
}

// TEST: GIVEN Reset is called THEN the logger is reset
func TestReset(t *testing.T) {
	logger.Reset() // Reset the logger to ensure a fresh instance
	log1 := logger.GetLogger(cfg.Setup.Logging.Level)
	if log1 == nil {
		t.Error("Expected logger to be non-nil after reset")
	}
}
