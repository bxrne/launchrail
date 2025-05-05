package logger

import (
	"sync"

	"github.com/zerodha/logf"
)

var (
	logger logf.Logger
	once   sync.Once
	opts   = logf.Opts{
		EnableCaller:    true,
		TimestampFormat: "15:04:05",
		EnableColor:     true,
		Level:           logf.InfoLevel,
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
func Reset() {
	once = sync.Once{}
	logger = logf.Logger{}
}
