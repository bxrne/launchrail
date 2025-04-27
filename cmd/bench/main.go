package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/config" // Import config package
	"github.com/bxrne/launchrail/internal/logger"
)

func main() {
	// --- Load Configuration ---
	cfg, err := config.GetConfig()
	if err != nil {
		// Use a temporary basic logger for config loading errors
		basicLogger := logger.GetLogger("error") // Default to error if config fails
		basicLogger.Fatal("Failed to load configuration", "error", err)
	}

	// --- Initialize Logger with Config Level ---
	benchLogger := logger.GetLogger(cfg.Setup.Logging.Level)

	// --- Define Benchmark Data Path (Hardcoded) ---
	const benchdataPath = "./benchdata"

	absBenchdataPath, err := filepath.Abs(benchdataPath)
	if err != nil {
		// Log the error and exit
		benchLogger.Error("Error getting absolute path for benchdata", "path", benchdataPath, "error", err)
		os.Exit(1)
	}

	benchLogger.Info("Benchmark Data Directory", "path", absBenchdataPath)

	// --- Benchmark Suite Setup ---
	config := BenchmarkConfig{
		BenchdataPath: absBenchdataPath,
	}
	suite := NewBenchmarkSuite(config)

	// Add specific benchmarks to the suite
	suite.AddBenchmark(NewHiprEuroc24Benchmark(config)) // Pass config here
	// Add more benchmarks here as needed

	// --- Run Benchmarks ---
	benchLogger.Info("Starting benchmark suite...")
	// Assign overallPass to _ as we now check failedCount directly
	results, _, err := suite.RunAll()
	if err != nil {
		// Use Error as RunAll now returns the error instead of logging and exiting directly
		// Log the error and exit
		benchLogger.Error("Error running benchmark suite", "error", err)
		os.Exit(1) // Exit explicitly if RunAll returns an error
	}

	// --- Print Results Summary ---
	failedCount := 0
	passedCount := 0
	benchLogger.Info("--- Benchmark Results Summary ---")
	for benchmarkName, benchmarkResults := range results {
		fmt.Printf("\nResults for %s:\n", benchmarkName)
		benchmarkFailed := false
		for _, res := range benchmarkResults {
			status := "PASSED"
			if !res.Passed {
				status = "FAILED"
				failedCount++
				benchmarkFailed = true // Mark the benchmark as failed if any metric failed
			} else {
				passedCount++
			}
			fmt.Printf("  [%s] %s: %s\n", status, res.Name, res.Description)
		}
		if benchmarkFailed {
			benchLogger.Error("Benchmark failed", "name", benchmarkName)
		} else {
			benchLogger.Info("Benchmark passed", "name", benchmarkName)
		}
	}

	fmt.Println("\n--- Overall ---")
	fmt.Printf("Total Metrics Passed: %d\n", passedCount)
	fmt.Printf("Total Metrics Failed: %d\n", failedCount)

	if failedCount > 0 {
		// Log the error and exit
		benchLogger.Error("Benchmark suite finished with failed metrics", "count", failedCount)
		os.Exit(1) // Exit with non-zero status code if any metric failed
	} else {
		benchLogger.Info("Benchmark suite finished successfully.")
		os.Exit(0) // Exit with zero status code if all metrics passed
	}
}
