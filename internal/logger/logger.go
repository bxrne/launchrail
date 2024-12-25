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
		EnableCaller:    true,
		TimestampFormat: time.RFC3339Nano,
		Level:           logf.DebugLevel,
	}
)

// GetLogger returns the singleton instance of the logger.
func GetLogger() logf.Logger {
	once.Do(func() {
		logger = logf.New(opts)
	})

	return logger
}
