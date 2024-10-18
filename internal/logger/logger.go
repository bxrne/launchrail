package logger

import (
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

var (
	once     sync.Once
	instance *log.Logger
)

// NOTE: Returns the singleton logger instance
func GetLogger() *log.Logger {
	once.Do(func() {
		instance = log.NewWithOptions(os.Stderr, log.Options{
			ReportCaller:    true,
			ReportTimestamp: true,
			TimeFormat:      time.ANSIC,
			Level:           log.DebugLevel,
		})
	})

	return instance
}
