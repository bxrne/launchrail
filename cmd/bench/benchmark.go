package main

import (
	"math"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	logf "github.com/zerodha/logf"
)

// BenchmarkResult holds the outcome of a single benchmark comparison.
type BenchmarkResult struct {
	Name          string
	Description   string
	Metric        string // e.g., "Apogee", "Max Velocity"
	Expected      float64
	Actual        float64
	Difference    float64
	Tolerance     float64
	ToleranceType string // Type of tolerance applied ("relative" or "absolute")
	Passed        bool
}

// Benchmark defines the interface for a runnable benchmark case.
type Benchmark interface {
	Name() string
	Run(entry config.BenchmarkEntry, logger *logf.Logger, runDir string) ([]BenchmarkResult, error)
}

// BenchmarkSuite manages and runs a collection of benchmarks.
type BenchmarkSuite struct {
	Config     BenchmarkConfig
	Benchmarks []Benchmark
}

// BenchmarkConfig holds configuration for the benchmark suite.
type BenchmarkConfig struct {
	BenchdataPath string // Path to the directory containing expected benchmark data CSVs
	ResultDirPath string // Path to the directory containing the actual simulation result CSVs
}

// NewBenchmarkSuite creates a new benchmark suite.
func NewBenchmarkSuite(config BenchmarkConfig) *BenchmarkSuite {
	return &BenchmarkSuite{
		Config:     config,
		Benchmarks: make([]Benchmark, 0),
	}
}

// AddBenchmark adds a benchmark to the suite.
func (s *BenchmarkSuite) AddBenchmark(b Benchmark) {
	s.Benchmarks = append(s.Benchmarks, b)
}

// RunAll runs all benchmarks in the suite.
func (s *BenchmarkSuite) RunAll(logLevel string) (map[string][]BenchmarkResult, bool, error) {
	benchLogger := logger.GetLogger(logLevel)

	allResults := make(map[string][]BenchmarkResult)
	overallPass := true

	for _, bench := range s.Benchmarks {
		benchLogger.Info("--- Starting Benchmark --- ", "name", bench.Name())

		benchLogger.Warn("BenchmarkSuite.RunAll is deprecated/broken due to interface changes", "benchmark", bench.Name())
		overallPass = false
		allResults[bench.Name()] = []BenchmarkResult{
			{Metric: "Suite Error", Description: "BenchmarkSuite.RunAll is deprecated", Passed: false},
		}
	}

	return allResults, overallPass, nil
}

// compareFloat compares two floats within a tolerance, handling zero expected values.
func compareFloat(name, description string, expected, actual, tolerancePercent float64) BenchmarkResult {
	passed := false
	diff := actual - expected
	absoluteDiff := math.Abs(diff)
	toleranceType := "relative"
	var calculatedTolerance float64

	if expected == 0 {
		toleranceType = "absolute"
		calculatedTolerance = tolerancePercent
		if absoluteDiff <= calculatedTolerance {
			passed = true
		}
	} else {
		calculatedTolerance = math.Abs(expected * tolerancePercent)
		if absoluteDiff <= calculatedTolerance {
			passed = true
		}
	}

	return BenchmarkResult{
		Name:          name,
		Description:   description,
		Metric:        name,
		Expected:      expected,
		Actual:        actual,
		Difference:    diff,
		Tolerance:     calculatedTolerance,
		ToleranceType: toleranceType,
		Passed:        passed,
	}
}
