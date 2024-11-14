package logger_test

import (
	"os"
	"testing"

	"github.com/bxrne/launchrail/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN the logger is requested WHEN it is the first time THEN it should create a new instance
func TestGetLogger_FirstTime(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "launchrail*.log")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	log := logger.GetLogger(tmpFile.Name())
	assert.NotNil(t, log, "Expected logger instance to be created")
}

// TEST: GIVEN the logger is requested WHEN it is not the first time THEN it should return the existing instance
func TestGetLogger_SubsequentCalls(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "launchrail*.log")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	firstInstance := logger.GetLogger(tmpFile.Name())
	secondInstance := logger.GetLogger(tmpFile.Name())

	assert.Equal(t, firstInstance, secondInstance, "Expected the same logger instance to be returned")
}

// TEST: GIVEN the logger is requested WHEN it is created THEN it should create a log file
func TestGetLogger_LogFileCreation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "launchrail*.log")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	logger.GetLogger(tmpFile.Name())

	_, err = os.Stat(tmpFile.Name())
	assert.False(t, os.IsNotExist(err), "Expected log file to be created")
}
