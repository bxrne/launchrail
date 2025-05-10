package logger_test

import (
	"os"
	"strings"
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
			continue
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

// TEST: GIVEN logs are written to a file THEN the output contains no ANSI color codes
func TestLogFileHasNoColorCodes(t *testing.T) {
	logger.Reset()
	logFile := "test_no_color.log"
	defer func() { _ = os.Remove(logFile) }()

	log := logger.GetLogger(cfg.Setup.Logging.Level, logFile)
	log.Info("No color test log entry")
	// No log.Sync() on logf.Logger; log file should be flushed immediately or on close

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	logContent := string(data)
	if containsANSICodes(logContent) {
		t.Errorf("Log file contains ANSI color codes: %q", logContent)
	}
}

// containsANSICodes returns true if the string contains ANSI escape sequences
func containsANSICodes(s string) bool {
	// ANSI escape sequences start with \x1b[ or \033[
	return strings.Contains(s, "\x1b[") || strings.Contains(s, "\033[")
}

