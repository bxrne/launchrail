package main

import (
	"fmt"

	"github.com/bxrne/launchrail/internal/logger" // Import custom logger
)

// BenchmarkResult holds the outcome of a single benchmark comparison.
type BenchmarkResult struct {
	Name        string
	Passed      bool
	Metric      string // e.g., "Apogee", "Max Velocity"
	Expected    float64
	Actual      float64
	Difference  float64
	Tolerance   float64
	Description string
}

// Benchmark defines the interface for a runnable benchmark case.
type Benchmark interface {
	Name() string
	LoadData(dataPath string) error
	Setup() error
	Run(simData interface{}) ([]BenchmarkResult, error) // simData needs definition
}

// BenchmarkSuite manages and runs a collection of benchmarks.
type BenchmarkSuite struct {
	Config     BenchmarkConfig
	Benchmarks []Benchmark
}

// BenchmarkConfig holds configuration for the suite.
type BenchmarkConfig struct {
	BenchdataPath string
	// Add other config like simulation parameters if needed
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
		// TODO: This needs the actual simulation data passed in.
		// For now, passing nil, which means benchmarks might use internal/loaded data.
		results, err := bench.Run(nil) // Pass simulation data here eventually
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
