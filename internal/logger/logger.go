package logger

import (
	"sync"

	"github.com/zerodha/logf"
)

var (
	logger      logf.Logger
	once        sync.Once
	defaultOpts = logf.Opts{
		EnableCaller:    true,
		TimestampFormat: "15:04:05",
		EnableColor:     true,
		Level:           logf.InfoLevel, // Default level
	}
)

// GetLogger returns the singleton instance of the logger.
// The 'level' parameter is only effective on the first call that initializes the logger.
// Subsequent calls will return the already initialized logger, ignoring the 'level' parameter.
func GetLogger(level string) *logf.Logger {
	once.Do(func() {
		currentOpts := defaultOpts // Start with defaults
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
			// If an unrecognized level is passed on the first call,
			// or if level is empty, use the level from defaultOpts.
			logLevel = defaultOpts.Level
			// TODO: Consider logging a one-time warning if a non-empty, unrecognized level is provided,
			// e.g., using a temporary logger instance here if this initialization itself needs logging.
		}
		currentOpts.Level = logLevel
		logger = logf.New(currentOpts)
	})
	return &logger
}

// Reset is for testing so that we can reset the logger singleton
func Reset() {
	once = sync.Once{}
	// Setting logger to an empty struct effectively makes it a no-op logger until next GetLogger call.
	// Or, one could set it to nil and add nil checks, but an empty struct is safer for direct use.
	logger = logf.Logger{}
}
