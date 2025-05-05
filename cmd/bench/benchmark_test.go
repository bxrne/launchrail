package main

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	logf "github.com/zerodha/logf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBenchmark is a mock implementation of the Benchmark interface for testing.
type MockBenchmark struct {
	MockName        string
	MockSetupErr    error
	MockRunErr      error
	MockResults     []BenchmarkResult
	MockLoadDataErr error
	LoadDataCalled  bool
	SetupCalled     bool
	RunCalled       bool
}

func (m *MockBenchmark) Name() string {
	return m.MockName
}

// LoadData implements the Benchmark interface for MockBenchmark.
func (m *MockBenchmark) LoadData(benchDataDir string) error {
	m.LoadDataCalled = true
	// We don't use benchDataDir in the mock, but it's needed for the interface
	return m.MockLoadDataErr
}

func (m *MockBenchmark) Setup() error {
	m.SetupCalled = true
	return m.MockSetupErr
}

func (m *MockBenchmark) Run(entry config.BenchmarkEntry, logger *logf.Logger, runDir string) ([]BenchmarkResult, error) {
	m.RunCalled = true
	return m.MockResults, m.MockRunErr
}

func TestNewBenchmarkSuite(t *testing.T) {
	config := BenchmarkConfig{BenchdataPath: "/tmp/benchdata"}
	suite := NewBenchmarkSuite(config)
	require.NotNil(t, suite)
	assert.Equal(t, config, suite.Config)
	assert.NotNil(t, suite.Benchmarks)
	assert.Empty(t, suite.Benchmarks)
}

func TestBenchmarkSuite_AddBenchmark(t *testing.T) {
	suite := NewBenchmarkSuite(BenchmarkConfig{})
	mockBench := &MockBenchmark{MockName: "TestBench"}
	suite.AddBenchmark(mockBench)

	require.Len(t, suite.Benchmarks, 1)
	assert.Equal(t, mockBench, suite.Benchmarks[0])
}

// TODO: Add TestBenchmarkSuite_RunAll_Failure cases
// func TestBenchmarkSuite_Run_Single(t *testing.T) { ... } // Example for future single run test
