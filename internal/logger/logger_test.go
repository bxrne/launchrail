package logger_test

import (
	"os"
	"testing"

	"github.com/bxrne/launchrail/internal/logger"

	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a valid file path WHEN GetLogger is called THEN it should return a logger instance
func TestGetLogger_ValidFilePath(t *testing.T) {
	logger.Reset()

	tmpFile, err := os.CreateTemp("", "logger_test_*.log")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	logInstance, err := logger.GetLogger(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, logInstance)
}

// TEST: GIVEN an invalid file path WHEN GetLogger is called THEN it should return an error
func TestGetLogger_InvalidFilePath(t *testing.T) {
	logger.Reset()

	logInstance, err := logger.GetLogger("/invalid/path/logger_test.log")
	assert.Error(t, err)
	assert.Nil(t, logInstance)
}

// TEST: GIVEN a valid file path WHEN GetLogger is called multiple times THEN it should return the same instance
func TestGetLogger_Singleton(t *testing.T) {
	logger.Reset()

	tmpFile, err := os.CreateTemp("", "logger_test_*.log")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	logInstance1, err := logger.GetLogger(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, logInstance1)

	logInstance2, err := logger.GetLogger(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, logInstance2)

	assert.Equal(t, logInstance1, logInstance2)
}
