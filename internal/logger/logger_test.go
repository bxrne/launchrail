package logger_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/logger"
	"github.com/rs/zerolog"
)

func TestGetLogger(t *testing.T) {
	log := logger.GetLogger()

	if log.GetLevel() != zerolog.TraceLevel {
		t.Errorf("Expected log level to be TraceLevel, but got %v", log.GetLevel())
	}

	log.Info().Msg("This is an info message")
	log.Debug().Msg("This is a debug message")
	log.Error().Msg("This is an error message")
}
