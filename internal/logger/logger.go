package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zerodha/logf"
)

var (
	globalLogger    logf.Logger
	once            sync.Once
	logFile         *os.File
	defaultOpts     = logf.Opts{
		EnableCaller:    true,
		TimestampFormat: "15:04:05",
		EnableColor:     false,
		Level:           logf.InfoLevel,
	}
	// UserCurrentFunc is a variable that holds the function to get the current user.
	// It's exported for testing purposes to allow mocking.
	UserCurrentFunc = user.Current
)

// GetDefaultOpts returns a copy of the default logger options.
// This is useful for tests that need to modify options for a specific logger instance.
func GetDefaultOpts() logf.Opts {
	return defaultOpts
}

// InitFileLogger sets up the global logger with file output.
// It ensures the log directory exists and creates a timestamped log file.
func InitFileLogger(configuredLevel string, appName string) (*logf.Logger, error) {
	usr, err := UserCurrentFunc() // Use the exported function variable
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	homedir := usr.HomeDir
	outputBase := filepath.Join(homedir, ".launchrail")
	logsDir := filepath.Join(outputBase, "logs")

	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory '%s': %w", logsDir, err)
	}

	currentTime := time.Now().Format("2006-01-02_15-04-05")
	logFileName := fmt.Sprintf("%s-%s.log", appName, currentTime)
	fullLogFilePath := filepath.Join(logsDir, logFileName)

	// GetLogger will be called with the determined path and level.
	// The sync.Once inside GetLogger will handle the singleton initialization.
	lg := GetLogger(configuredLevel, fullLogFilePath)
	lg.Info("File logger initialized", "app", appName, "path", fullLogFilePath, "level", configuredLevel)
	return lg, nil
}

// GetLogger returns the singleton instance of the logger.
// If filePath is provided (typically by InitFileLogger), it attempts to set up file logging.
// Otherwise, or if file opening fails, it defaults to stdout.
// The 'level' and 'filePath' parameters are only effective on the first call that initializes the logger.
func GetLogger(level string, filePath ...string) *logf.Logger {
	once.Do(func() {
		currentOpts := GetDefaultOpts()
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
			logLevel = currentOpts.Level // Use default if level string is unrecognized
		}
		currentOpts.Level = logLevel

		var writers []io.Writer
		writers = append(writers, os.Stdout) // Always log to stdout

		if len(filePath) > 0 && filePath[0] != "" {
			var err error
			// Use the first path provided. It's expected to be set by InitFileLogger.
			actualLogFilePath := filePath[0]
			logFile, err = os.OpenFile(actualLogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				// Use standard log package for this specific error, as logf isn't fully set up yet.
				// This message will go to stdout.
				log.Printf("[logger] Failed to open log file '%s': %v. Continuing with stdout only.", actualLogFilePath, err)
			} else {
				writers = append(writers, logFile)
			}
		}
		currentOpts.Writer = io.MultiWriter(writers...)
		globalLogger = logf.New(currentOpts)
	})
	return &globalLogger
}

// LoggingMiddleware returns a Gin middleware that logs all HTTP requests with details.
func LoggingMiddleware(log *logf.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		log.Info("HTTP Request",
			"status", status,
			"method", method,
			"path", path,
			"query", query,
			"ip", clientIP,
			"latency", latency.String(),
			"user_agent", userAgent,
		)
	}
}

// Reset is for testing so that we can reset the logger singleton
func Reset() {
	once = sync.Once{}
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
	globalLogger = logf.Logger{}
}
