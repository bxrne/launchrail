package logger_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/logger"
)

// TEST: GIVEN GetLogger is called THEN a non-nil logger is returned
func TestGetLogger(t *testing.T) {
	log := logger.GetLogger()
	if log == nil {
		t.Error("Expected logger to be non-nil")
	}
}

// TEST: GIVEN GetLogger is called multiple times THEN the logger is a singleton
func TestGetLoggerSingleton(t *testing.T) {
	log1 := logger.GetLogger()
	log2 := logger.GetLogger()

	if log1 != log2 {
		t.Error("Expected logger to be a singleton")
	}
}
