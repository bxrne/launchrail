package logger_test

import (
	"bytes"
	"os"
	"sync"
	"testing"

	"github.com/bxrne/launchrail/internal/logger"
	"github.com/charmbracelet/log"
)

// TEST: GIVEN a logger WHEN GetLogger is called multiple times THEN it should always return the same instance.
func TestLoggerSingleton(t *testing.T) {
	logger1 := logger.GetLogger()
	logger2 := logger.GetLogger()

	if logger1 != logger2 {
		t.Error("Expected logger1 and logger2 to be the same instance")
	}
}

// TEST: GIVEN a logger WHEN GetLogger is accessed concurrently THEN it should return the same instance for all accesses.
func TestLoggerThreadSafety(t *testing.T) {
	var wg sync.WaitGroup
	loggers := make([]*log.Logger, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			loggers[i] = logger.GetLogger()
		}(i)
	}

	wg.Wait()

	for i := 1; i < len(loggers); i++ {
		if loggers[i] != loggers[0] {
			t.Errorf("Logger instance %d is not the same as logger instance 0", i)
		}
	}
}

// TEST: GIVEN the logger WHEN multiple levels of logs are used THEN the appropriate log level is respected.
func TestLoggerLevel(t *testing.T) {
	logger := logger.GetLogger()

	logger.SetLevel(log.ErrorLevel)

	var buf bytes.Buffer
	originalStderr := os.Stderr
	defer func() { os.Stderr = originalStderr }()
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger.Debug("This should not appear")

	w.Close()
	buf.ReadFrom(r)

	if bytes.Contains(buf.Bytes(), []byte("This should not appear")) {
		t.Error("Debug log should not be output at Error level")
	}
}
