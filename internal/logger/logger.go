package logger

import (
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zerodha/logf"
)

var (
	logger      logf.Logger
	once        sync.Once
	logFile     *os.File
	logFilePath string
	defaultOpts = logf.Opts{
		EnableCaller:    true,
		TimestampFormat: "15:04:05",
		EnableColor:     false, // Disable color globally so log file output has no ANSI codes
		Level:           logf.InfoLevel, // Default level
	}
)

// GetLogger returns the singleton instance of the logger.
// The 'level' parameter is only effective on the first call that initializes the logger.
// The 'filePath' parameter is optional and only effective on the first call.
func GetLogger(level string, filePath ...string) *logf.Logger {
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
			logLevel = defaultOpts.Level
		}
		currentOpts.Level = logLevel

		var writers []io.Writer
		writers = append(writers, os.Stdout)
		if len(filePath) > 0 && filePath[0] != "" {
			var err error
			logFilePath = filePath[0]
			logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("[logger] Failed to open log file '%s': %v", logFilePath, err)
			} else {
				writers = append(writers, logFile)
			}
		}
		currentOpts.Writer = io.MultiWriter(writers...)
		logger = logf.New(currentOpts)
	})
	return &logger
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
	logger = logf.Logger{}
}
