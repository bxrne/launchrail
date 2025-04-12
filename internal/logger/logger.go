package logger

import (
	"sync"
	"time"

	"github.com/zerodha/logf"
)

var (
	once   sync.Once
	logger logf.Logger
	opts   logf.Opts = logf.Opts{
		EnableColor:     true,
		EnableCaller:    false,
		TimestampFormat: time.RFC3339Nano,
		Level:           logf.DebugLevel,
	}
)

// GetLogger returns the singleton instance of the logger.
func GetLogger(level string) *logf.Logger {
	var logLevel logf.Level
	switch level {
	case "debug":
		logLevel = logf.DebugLevel
	case "info":
		logLevel = logf.InfoLevel
	case "warn":
		logLevel = logf.WarnLevel
	case "error":
		logLevel = logf.ErrorLevel
	case "fatal":
		logLevel = logf.FatalLevel
	default:
		logLevel = logf.DebugLevel
	}

	opts.Level = logLevel

	once.Do(func() {
		logger = logf.New(opts)
	})

	return &logger
}

// Reset is for testing so that we can reset the logger singleton
// and create a new instance.
func Reset() {
	once = sync.Once{}
	logger = logf.Logger{}
}
