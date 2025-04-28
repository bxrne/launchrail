package main

import (
	"fmt"
	"math"

	"github.com/bxrne/launchrail/internal/logger" // Import custom logger
	"github.com/bxrne/launchrail/internal/storage"
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
	LoadData(dataPath string) error
	Setup() error
	Run() ([]BenchmarkResult, error) // Removed simData parameter
}

// BenchmarkSuite manages and runs a collection of benchmarks.
type BenchmarkSuite struct {
	Config     BenchmarkConfig
	Benchmarks []Benchmark
}

// BenchmarkConfig holds configuration for the suite.
type BenchmarkConfig struct {
	BenchdataPath string
	SimRecordHash string             // Added: Hash of the simulation record to use
	RecordManager *storage.RecordManager // Added: Record manager to load the sim record
}

// NewBenchmarkSuite creates a new benchmark suite.
func NewBenchmarkSuite(config BenchmarkConfig) *BenchmarkSuite {
	// Initialize the Benchmarks slice
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
// It returns a map of benchmark names to their results, a boolean indicating overall success,
// and an error if any benchmark setup or run fails catastrophically.
func (s *BenchmarkSuite) RunAll() (map[string][]BenchmarkResult, bool, error) {
	// Get logger instance
	benchLogger := logger.GetLogger("info") // Assuming info level for suite logs

	allResults := make(map[string][]BenchmarkResult)
	overallPass := true

	// Use the actual benchmarks added to the suite
	for _, bench := range s.Benchmarks {
		benchLogger.Info("--- Running Benchmark ---", "name", bench.Name())

		// 1. Load Data
		if err := bench.LoadData(s.Config.BenchdataPath); err != nil {
			// Error already logged by the specific benchmark's LoadData (potentially)
			// Return the error to be handled by the caller (main.go)
			return nil, false, fmt.Errorf("benchmark '%s' LoadData failed: %w", bench.Name(), err)
		}

		// 2. Setup (Optional - if needed, specific benchmarks implement it)
		if err := bench.Setup(); err != nil {
			// Log and return the error
			benchLogger.Error("Error setting up benchmark", "name", bench.Name(), "error", err)
			return nil, false, fmt.Errorf("benchmark '%s' Setup failed: %w", bench.Name(), err)
		}

		// 3. Run Comparison
		// The benchmark now uses its internal config (SimRecordHash, RecordManager)
		results, err := bench.Run() // Call Run without simData
		if err != nil {
			// Log and return the error
			benchLogger.Error("Error running benchmark comparison", "name", bench.Name(), "error", err)
			return nil, false, fmt.Errorf("benchmark '%s' Run failed: %w", bench.Name(), err)
		}

		allResults[bench.Name()] = results

		// Check if any result within this benchmark failed
		benchmarkPass := true
		for _, result := range results {
			if !result.Passed {
				benchmarkPass = false
				overallPass = false // If any metric in any benchmark fails, overall fails
				// No need to break, collect all results
			}
		}

		if benchmarkPass {
			benchLogger.Info("Benchmark finished", "name", bench.Name(), "status", "PASS")
		} else {
			benchLogger.Info("Benchmark finished", "name", bench.Name(), "status", "FAIL")
		}
		benchLogger.Info("---------------------------------") // Simple separator
	}

	return allResults, overallPass, nil
}

// --- Helper Function (Placeholder) ---

// TODO: Implement the actual simulation run based on benchmark needs
// func runSimulationForBenchmark(b Benchmark) (interface{}, error) {
// 	 fmt.Printf("Simulating for %s...\n", b.Name())
// 	 // ... simulation logic ...
// 	 time.Sleep(1 * time.Second) // Simulate work
// 	 return map[string]float64{"Apogee": 7400.0, "MaxVelocity": 1050.0}, nil // Example simulated data
// }

// compareFloat compares two floats within a tolerance, handling zero expected values.
func compareFloat(name, description string, expected, actual, tolerancePercent float64) BenchmarkResult {
	passed := false
	diff := actual - expected
	absoluteDiff := math.Abs(diff)
	toleranceType := "relative"
	var calculatedTolerance float64

	if expected == 0 {
		// For zero expected values, tolerance is absolute.
		toleranceType = "absolute"
		calculatedTolerance = tolerancePercent // Treat the input as absolute tolerance here
		if absoluteDiff <= calculatedTolerance {
			passed = true
		}
	} else {
		// Calculate relative tolerance
		calculatedTolerance = math.Abs(expected * tolerancePercent)
		if absoluteDiff <= calculatedTolerance {
			passed = true
		}
	}

	return BenchmarkResult{
		Name:          name, // Use the provided metric name
		Description:   description,
		Metric:        name, // Use the provided metric name here as well
		Expected:      expected,
		Actual:        actual,
		Difference:    diff, // Report the signed difference
		Tolerance:     calculatedTolerance, // Report the calculated absolute tolerance value
		ToleranceType: toleranceType,
		Passed:        passed,
	}
}
