package logger_test

import (
	"os"
	"testing"

	"github.com/bxrne/launchrail/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for os.File
type MockFile struct {
	mock.Mock
}

func (m *MockFile) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockFile) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockFile) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// TEST: GIVEN a mocked file WHEN GetLogger is called THEN it should return a non-nil logger instance
func TestGetLogger(t *testing.T) {
	originalOpenFile := os.OpenFile
	defer func() { os.OpenFile = originalOpenFile }()

	mockFile := new(MockFile)
	mockFile.On("Write", mock.Anything).Return(0, nil)
	mockFile.On("Close").Return(nil)
	mockFile.On("Sync").Return(nil)

	os.OpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return mockFile, nil
	}

	logInstance := logger.GetLogger()
	assert.NotNil(t, logInstance, "Expected logger instance to be non-nil")

	logInstance2 := logger.GetLogger()
	assert.Equal(t, logInstance, logInstance2, "Expected the same logger instance to be returned")

	mockFile.AssertExpectations(t)
}
