package logger

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	once     sync.Once
	instance zerolog.Logger
)

// GetLogger returns the singleton instance of the logger.
func GetLogger() zerolog.Logger {
	once.Do(func() {
		instance = zerolog.New(
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		).Level(zerolog.TraceLevel).With().Timestamp().Caller().Logger()
	})
	return instance
}
