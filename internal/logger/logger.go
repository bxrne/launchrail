package logger

import (
	"sync"
	"time"

	"github.com/bxrne/launchrail/internal/config"
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
func GetLogger(cfg *config.Config) *logf.Logger {
	switch cfg.Logging.Level {
	case "debug":
		opts.Level = logf.DebugLevel
	case "info":
		opts.Level = logf.InfoLevel
	case "warn":
		opts.Level = logf.WarnLevel
	case "error":
		opts.Level = logf.ErrorLevel
	case "fatal":
		opts.Level = logf.FatalLevel
	}
	once.Do(func() {
		logger = logf.New(opts)
	})

	return &logger
}
