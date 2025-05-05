package main

import (
	"errors"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	logf "github.com/zerodha/logf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBenchmark is a mock implementation of the Benchmark interface for testing.
type MockBenchmark struct {
	MockName      string
	MockSetupErr  error
	MockRunErr    error
	MockResults   []BenchmarkResult
	MockLoadDataErr error
	LoadDataCalled bool
	SetupCalled   bool
	RunCalled     bool
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

func TestBenchmarkSuite_RunAll_Success(t *testing.T) {
	suite := NewBenchmarkSuite(BenchmarkConfig{})
	mockBench1 := &MockBenchmark{
		MockName: "Bench1",
		MockResults: []BenchmarkResult{
			{Name: "MetricA", Passed: true, Description: "Desc A"},
		},
	}
	mockBench2 := &MockBenchmark{
		MockName: "Bench2",
		MockResults: []BenchmarkResult{
			{Name: "MetricB", Passed: true, Description: "Desc B"},
			{Name: "MetricC", Passed: false, Description: "Desc C"},
		},
	}
	suite.AddBenchmark(mockBench1)
	suite.AddBenchmark(mockBench2)

	results, overallPass, err := suite.RunAll()

	require.NoError(t, err)
	assert.True(t, mockBench1.LoadDataCalled, "Bench1 LoadData should be called")
	assert.True(t, mockBench1.SetupCalled, "Bench1 Setup should be called")
	assert.True(t, mockBench1.RunCalled, "Bench1 Run should be called")
	assert.True(t, mockBench2.LoadDataCalled, "Bench2 LoadData should be called")
	assert.True(t, mockBench2.SetupCalled, "Bench2 Setup should be called")
	assert.True(t, mockBench2.RunCalled, "Bench2 Run should be called")

	require.Len(t, results, 2)
	assert.Equal(t, mockBench1.MockResults, results["Bench1"])
	assert.Equal(t, mockBench2.MockResults, results["Bench2"])

	assert.False(t, overallPass, "Overall result should be false due to MetricC failure")
}

func TestBenchmarkSuite_RunAll_SetupError(t *testing.T) {
	suite := NewBenchmarkSuite(BenchmarkConfig{})
	setupErr := errors.New("failed to setup")
	mockBench1 := &MockBenchmark{MockName: "Bench1", MockSetupErr: setupErr}
	mockBench2 := &MockBenchmark{MockName: "Bench2"}
	suite.AddBenchmark(mockBench1)
	suite.AddBenchmark(mockBench2)

	results, overallPass, err := suite.RunAll()

	require.Error(t, err)
	assert.ErrorIs(t, err, setupErr)
	assert.True(t, mockBench1.LoadDataCalled, "Bench1 LoadData should be called")
	assert.True(t, mockBench1.SetupCalled)
	assert.False(t, mockBench1.RunCalled, "Run should not be called if Setup fails")
	assert.False(t, mockBench2.LoadDataCalled, "Subsequent benchmarks should not run LoadData if previous fails")
	assert.False(t, mockBench2.SetupCalled, "Subsequent benchmarks should not run Setup if previous fails")
	assert.False(t, mockBench2.RunCalled)
	assert.Nil(t, results)
	assert.False(t, overallPass)
}

func TestBenchmarkSuite_RunAll_RunError(t *testing.T) {
	suite := NewBenchmarkSuite(BenchmarkConfig{})
	runErr := errors.New("failed to run")
	mockBench1 := &MockBenchmark{MockName: "Bench1", MockRunErr: runErr}
	mockBench2 := &MockBenchmark{MockName: "Bench2"}
	suite.AddBenchmark(mockBench1)
	suite.AddBenchmark(mockBench2)

	results, overallPass, err := suite.RunAll()

	require.Error(t, err)
	assert.ErrorIs(t, err, runErr)
	assert.True(t, mockBench1.LoadDataCalled, "Bench1 LoadData should be called")
	assert.True(t, mockBench1.SetupCalled)
	assert.True(t, mockBench1.RunCalled)
	assert.False(t, mockBench2.LoadDataCalled, "Subsequent benchmarks should not run LoadData if previous fails")
	assert.False(t, mockBench2.SetupCalled, "Subsequent benchmarks should not run Setup if previous fails")
	assert.False(t, mockBench2.RunCalled)
	assert.Nil(t, results)
	assert.False(t, overallPass)
}

func TestBenchmarkSuite_RunAll_LoadDataError(t *testing.T) {
	suite := NewBenchmarkSuite(BenchmarkConfig{})
	loadErr := errors.New("failed to load data")
	mockBench1 := &MockBenchmark{MockName: "Bench1", MockLoadDataErr: loadErr}
	mockBench2 := &MockBenchmark{MockName: "Bench2"}
	suite.AddBenchmark(mockBench1)
	suite.AddBenchmark(mockBench2)

	results, overallPass, err := suite.RunAll()

	require.Error(t, err)
	assert.ErrorIs(t, err, loadErr)
	assert.True(t, mockBench1.LoadDataCalled)
	assert.False(t, mockBench1.SetupCalled, "Setup should not be called if LoadData fails")
	assert.False(t, mockBench1.RunCalled)
	assert.False(t, mockBench2.LoadDataCalled, "Subsequent benchmarks should not run LoadData if previous fails")
	assert.False(t, mockBench2.SetupCalled)
	assert.False(t, mockBench2.RunCalled)
	assert.Nil(t, results)
	assert.False(t, overallPass)
}
