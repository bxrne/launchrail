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
		logFile, err := os.OpenFile("launchrail.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}

		instance = log.NewWithOptions(logFile, log.Options{
			ReportCaller:    true,
			ReportTimestamp: true,
			TimeFormat:      time.ANSIC,
			Level:           log.DebugLevel,
		})
	})
	return instance
}

