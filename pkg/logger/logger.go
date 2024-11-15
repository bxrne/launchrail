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
func GetLogger(fileOutPath string) (*log.Logger, error) {
	var err error

	once.Do(func() {
		var logFile *os.File

		logFile, err = os.OpenFile(fileOutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return
		}

		instance = log.NewWithOptions(logFile, log.Options{
			ReportCaller:    true,
			ReportTimestamp: true,
			TimeFormat:      time.ANSIC,
			Level:           log.DebugLevel,
		})
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

// WARN: Do not use. It is for testing
func Reset() {
	instance = nil
	once = sync.Once{}
}
