package logger_test

import (
	"github.com/bxrne/launchrail/internal/logger"
	"testing"
)

func TestGetLogger(t *testing.T) {
	log := logger.GetLogger()
	if log == nil {
		t.Error("Expected logger to be non-nil")
	}
}

func TestGetLoggerSingleton(t *testing.T) {
	log1 := logger.GetLogger()
	log2 := logger.GetLogger()

	if log1 != log2 {
		t.Error("Expected logger to be a singleton")
	}
}
