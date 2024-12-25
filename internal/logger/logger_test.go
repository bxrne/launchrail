package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bxrne/launchrail/internal/logger"
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

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	logger.SetLoggerOutput(&buf)
	logger.GetLogger().Info("test message")
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("Expected logger output to contain 'test message', got: %s", buf.String())
	}
}
