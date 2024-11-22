package logger

import (
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

// Global singleton instances for logger management
// INFO: These variables are package-level to maintain singleton pattern
var (
	once     sync.Once
	instance *log.Logger
)

// GetLogger returns a singleton instance of the logger configured with the specified output file.
// It ensures thread-safe initialization using sync.Once.
//
// INFO: This is the primary method for obtaining a logger instance throughout the application
//
// Parameters:
//   - fileOutPath: The file path where logs will be written
//
// Returns:
//   - *log.Logger: Configured logger instance
//   - error: Any error encountered during initialization
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

// Reset clears the logger instance and resets the singleton state.
// This should only be used in testing scenarios.
//
// WARN: This method is dangerous in production and should only be used in tests
// WARN: Calling this while other goroutines are using the logger may cause panic
func Reset() {
	instance = nil
	once = sync.Once{}
}
