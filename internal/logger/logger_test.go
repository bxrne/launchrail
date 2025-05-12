package logger_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"
)

// Barebones config for testing
var cfg = &config.Config{
	Setup: config.Setup{
		Logging: config.Logging{
			Level: "info",
		},
	},
}

// TEST: GIVEN GetLogger is called THEN a non-nil logger is returned
func TestGetLogger(t *testing.T) {
	log := logger.GetLogger(cfg.Setup.Logging.Level)
	if log == nil {
		t.Error("Expected logger to be non-nil")
	}
}

// TEST: GIVEN GetLogger is called multiple times THEN the logger is a singleton
func TestGetLoggerSingleton(t *testing.T) {
	log1 := logger.GetLogger(cfg.Setup.Logging.Level)
	log2 := logger.GetLogger(cfg.Setup.Logging.Level)

	if log1 != log2 {
		t.Error("Expected logger to be a singleton")
	}
}

// TEST: GIVEN GetLogger is called with different levels THEN the logger level is set correctly
func TestGetLoggerDifferentLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "fatal"}

	for _, level := range levels {
		logger.Reset() // Reset the logger to ensure a fresh instance
		cfg.Setup.Logging.Level = level
		log := logger.GetLogger(level)
		if log == nil {
			t.Errorf("Expected logger to be non-nil for level %s", level)
			continue
		}
		if log.Level.String() != level {
			t.Errorf("Expected logger level to be %s, got %s", level, log.Level.String())
		}
	}
}

// TEST: GIVEN Reset is called THEN the logger is reset
func TestReset(t *testing.T) {
	logger.Reset() // Reset the logger to ensure a fresh instance
	log1 := logger.GetLogger(cfg.Setup.Logging.Level)
	if log1 == nil {
		t.Error("Expected logger to be non-nil after reset")
	}
}

// TEST: GIVEN logs are written to a file THEN the output contains no ANSI color codes
func TestLogFileHasNoColorCodes(t *testing.T) {
	logger.Reset()
	logFile := "test_no_color.log"
	defer func() { _ = os.Remove(logFile) }()

	log := logger.GetLogger(cfg.Setup.Logging.Level, logFile)
	log.Info("No color test log entry")
	// No log.Sync() on logf.Logger; log file should be flushed immediately or on close

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	logContent := string(data)
	if containsANSICodes(logContent) {
		t.Errorf("Log file contains ANSI color codes: %q", logContent)
	}
}

// containsANSICodes returns true if the string contains ANSI escape sequences
func containsANSICodes(s string) bool {
	// ANSI escape sequences start with \x1b[ or \033[
	return strings.Contains(s, "\x1b[") || strings.Contains(s, "\033[")
}

// TEST: GIVEN GetLogger is called with an unrecognized level THEN the logger level defaults to info
func TestGetLogger_UnrecognizedLevelDefaultsToInfo(t *testing.T) {
	logger.Reset()
	cfg.Setup.Logging.Level = "verywronglevel" // An unrecognized level
	logInstance := logger.GetLogger(cfg.Setup.Logging.Level) // Pass the level string
	if logInstance == nil {
		t.Fatal("Expected logger to be non-nil for unrecognized level")
	}
	if logInstance.Level != logf.InfoLevel { // Default level is InfoLevel in defaultOpts
		t.Errorf("Expected logger level to be '%s' for unrecognized level, got '%s'", logf.InfoLevel.String(), logInstance.Level.String())
	}
}

// TEST: GIVEN GetLogger is called with a filePath that cannot be opened THEN it logs an error and uses stdout
func TestGetLogger_FileOpenError(t *testing.T) {
	logger.Reset()

	var buf bytes.Buffer
	originalStdLogOutput := log.Writer() // log.Writer() returns current output
	log.SetOutput(&buf)                 // Intercepts output from standard `log` package
	defer func() {
		log.SetOutput(originalStdLogOutput)
	}()

	// Use a path that should fail: the temp directory itself (which is a directory, not a file)
	invalidLogFilePath := os.TempDir()

	logInstance := logger.GetLogger("info", invalidLogFilePath) // Use the invalid path
	if logInstance == nil {
		t.Fatal("Expected logger to be non-nil even with file open error")
	}

	loggedOutput := buf.String()
	expectedErrorMsg := "Failed to open log file"
	if !strings.Contains(loggedOutput, expectedErrorMsg) {
		t.Errorf("Expected log output to contain '%s', but got '%s'", expectedErrorMsg, loggedOutput)
	}
	// Check if the path is mentioned in the error. The logger formats it like "'%s'"
	if !strings.Contains(loggedOutput, "'"+invalidLogFilePath+"'") {
		t.Errorf("Expected log output to contain the invalid path '%s', but got '%s'", invalidLogFilePath, loggedOutput)
	}
}

func TestInitFileLogger_Success(t *testing.T) {
	logger.Reset()
	appName := "testAppSuccess"
	logLevel := "debug"

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		tempUser, err := user.Current()
		if err != nil {
			t.Skipf("Skipping TestInitFileLogger_Success: could not determine home directory: %v", err)
		}
		homeDir = tempUser.HomeDir
	}
	assert.NotEmpty(t, homeDir, "HOME directory must be available for this test")

	expectedLogDir := filepath.Join(homeDir, ".launchrail", "logs")

	// Clean up potential old log files from previous runs for this specific appName to ensure a clean test.
	// This helps in reliably checking if a new log file is created.
	files, _ := filepath.Glob(filepath.Join(expectedLogDir, appName+"-*.log"))
	for _, f := range files {
		_ = os.Remove(f)
	}

	logInstance, err := logger.InitFileLogger(logLevel, appName)

	assert.NoError(t, err)
	assert.NotNil(t, logInstance)
	assert.Equal(t, logLevel, logInstance.Level.String())

	createdFiles, ioErr := os.ReadDir(expectedLogDir)
	assert.NoError(t, ioErr, "Failed to read log directory")

	foundLogFile := false
	var createdLogFilePath string
	for _, file := range createdFiles {
		if strings.HasPrefix(file.Name(), appName+"-") && strings.HasSuffix(file.Name(), ".log") {
			foundLogFile = true
			createdLogFilePath = filepath.Join(expectedLogDir, file.Name())
			defer os.Remove(createdLogFilePath) // Clean up the created log file
			break
		}
	}
	assert.True(t, foundLogFile, "Expected log file to be created in %s with prefix %s", expectedLogDir, appName)
}

func TestInitFileLogger_UserError(t *testing.T) {
	logger.Reset()
	// Temporarily modify user.Current to simulate an error
	originalUserCurrent := logger.UserCurrentFunc // Use exported UserCurrentFunc
	logger.UserCurrentFunc = func() (*user.User, error) {
		return nil, fmt.Errorf("simulated user error")
	}
	defer func() { logger.UserCurrentFunc = originalUserCurrent }()

	_, err := logger.InitFileLogger("info", "testAppUserError")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get current user")
	assert.Contains(t, err.Error(), "simulated user error")
}

func TestInitFileLogger_MkdirError(t *testing.T) {
	logger.Reset()

	// Setup: Create a file where a directory is expected, to cause os.MkdirAll to fail.
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		tempUser, err := user.Current()
		if err != nil {
			t.Skipf("Skipping test: could not determine home directory: %v", err)
		}
		homeDir = tempUser.HomeDir
	}
	assert.NotEmpty(t, homeDir, "HOME directory must be available for this test")

	outputBase := filepath.Join(homeDir, ".launchrail")
	logsDirBlocker := filepath.Join(outputBase, "logs") // This path should be a dir

	// Ensure parent .launchrail exists
	_ = os.MkdirAll(outputBase, 0o755)

	// Remove any existing file or directory at logsDirBlocker to ensure we can create our blocking file.
	_ = os.RemoveAll(logsDirBlocker)

	// Create a file at 'logs' to block directory creation
	f, err := os.Create(logsDirBlocker)
	assert.NoError(t, err, "Setup: failed to create blocking file")
	f.Close()
	defer os.Remove(logsDirBlocker) // Clean up the blocking file

	_, err = logger.InitFileLogger("info", "testAppMkdirError")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create logs directory")
}

func TestLoggingMiddleware(t *testing.T) {
	logger.Reset()
	gin.SetMode(gin.TestMode)

	// Capture logger output
	var logOutput bytes.Buffer
	currentOpts := logger.GetDefaultOpts() // Assuming a way to get/set default opts or mock
	currentOpts.Writer = &logOutput
	currentOpts.EnableColor = false // Ensure no color codes for easier string matching
	logInstance := logf.New(currentOpts)

	// Setup Gin router with the middleware
	r := gin.New()
	r.Use(logger.LoggingMiddleware(&logInstance)) // Pass as pointer
	r.GET("/test_path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test_path?query=param", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	output := logOutput.String()
	t.Logf("Middleware log output: %s", output) // Log for debugging if test fails

	assert.Contains(t, output, "HTTP Request")
	assert.Contains(t, output, "status=200")
	assert.Contains(t, output, "method=GET")
	assert.Contains(t, output, "path=/test_path")
	assert.Contains(t, output, "query=param")
	assert.Contains(t, output, "user_agent=test-agent")
	// IP and latency are dynamic, so we check for their presence by key
	assert.Contains(t, output, "ip=")
	assert.Contains(t, output, "latency=")
}
