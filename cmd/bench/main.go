package main

import (
	"os"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	logf "github.com/zerodha/logf"
)

// Global logger instance
var benchLogger *logf.Logger

func main() {
	// --- Load Configuration ---
	cfg, err := config.GetConfig()
	if err != nil {
		basicLogger := logger.GetLogger("error")
		basicLogger.Fatal("Failed to load configuration", "error", err)
	}

	// --- Initialize Logger with Config Level ---
	benchLogger = logger.GetLogger(cfg.Setup.Logging.Level)

	// --- Determine Paths ---
	// Simulation results are expected here (consistent with launchrail main)
	simulationResultsDir := filepath.Join(cfg.Setup.App.BaseDir, "results")
	absSimulationResultsDir, err := filepath.Abs(simulationResultsDir)
	if err != nil {
		benchLogger.Fatal("Failed to get absolute path for simulation results directory", "path", simulationResultsDir, "error", err)
	}
	if _, err := os.Stat(absSimulationResultsDir); os.IsNotExist(err) {
		benchLogger.Warn("Simulation results directory does not exist yet, benchmarks might fail if they require it", "path", absSimulationResultsDir)
	} else {
		benchLogger.Info("Using simulation results directory", "path", absSimulationResultsDir)
	}

	// Base directory for finding benchmark data subdirectories
	benchmarkDataDir := cfg.BenchmarkDataDir
	if benchmarkDataDir == "" {
		benchLogger.Fatal("Missing required configuration: benchmark_data_dir")
	}
	absBenchmarkDataDir, err := filepath.Abs(benchmarkDataDir)
	if err != nil {
		benchLogger.Fatal("Failed to get absolute path for benchmark_data_dir", "path", benchmarkDataDir, "error", err)
	}
	if _, err := os.Stat(absBenchmarkDataDir); os.IsNotExist(err) {
		benchLogger.Fatal("Benchmark data directory not found", "path", absBenchmarkDataDir)
	}
	benchLogger.Info("Using benchmark data directory", "path", absBenchmarkDataDir)

	// Markdown output path (hardcoded for now)
	outputMarkdownPath := "BENCHMARK.md"
	benchLogger.Info("Markdown output will be written to", "path", outputMarkdownPath)

	// --- Discover Benchmarks (Subdirectories in benchmarkDataDir) ---
	discoveredTags := []string{}
	files, err := os.ReadDir(absBenchmarkDataDir)
	if err != nil {
		benchLogger.Fatal("Failed to read benchmark data directory", "path", absBenchmarkDataDir, "error", err)
	}
	for _, file := range files {
		if file.IsDir() {
			discoveredTags = append(discoveredTags, file.Name())
		}
	}
	//sort.Strings(discoveredTags) // Process in alphabetical order

	if len(discoveredTags) == 0 {
		benchLogger.Warn("No benchmark subdirectories found in benchmark_data_dir", "path", absBenchmarkDataDir)
		os.Exit(0) // Exit cleanly if no benchmarks found
	}
	benchLogger.Info("Discovered benchmark tags", "tags", discoveredTags)

	// --- Initialize Overall Results ---
	overallResults := make(map[string][]BenchmarkResult)
	overallFailedCount := 0
	overallPassedCount := 0

	// --- Loop Through Discovered Benchmark Tags ---
	for _, tag := range discoveredTags {
		benchLogger.Info("--- Starting Benchmark Run --- ", "tag", tag)

		// --- Determine Paths for this Benchmark ---
		currentBenchmarkDataPath := filepath.Join(absBenchmarkDataDir, tag)
		benchLogger.Info("Using benchmark data path for tag", "tag", tag, "path", currentBenchmarkDataPath)

		// Validate specific benchmark data path existence
		if _, err := os.Stat(currentBenchmarkDataPath); os.IsNotExist(err) {
			benchLogger.Error("Benchmark tag data directory not found, skipping", "tag", tag, "path", currentBenchmarkDataPath)
			// Store a failure result
			overallResults[tag] = []BenchmarkResult{
				{Metric: "Setup Error", Description: "Data directory not found", Passed: false},
			}
			overallFailedCount++
			continue
		}

		// --- Setup Benchmark Suite for this Benchmark ---
		benchConfig := BenchmarkConfig{
			BenchdataPath: currentBenchmarkDataPath, // Path to EXPECTED data (tag specific)
			ResultDirPath: absSimulationResultsDir,  // Path to ACTUAL data (shared)
		}

		suite := NewBenchmarkSuite(benchConfig)

		// --- Register Specific Benchmark Implementation ---
		// TODO: Implement a better mechanism to determine benchmark type from tag
		// For now, assume all tags correspond to HiprEuroc24 type benchmarks
		var benchmarkToRun Benchmark
		switch tag {
		// case "some-other-type":
		//     benchmarkToRun = NewSomeOtherBenchmark(benchConfig)
		default: // Assume HiprEuroc24 for any tag found
			benchLogger.Debug("Assuming benchmark type 'HiprEuroc24Benchmark' for tag", "tag", tag)
			benchmarkToRun = NewHiprEuroc24Benchmark(benchConfig)
		}

		// Add the benchmark instance to the suite
		suite.AddBenchmark(benchmarkToRun) // Pass only the benchmark instance

		// --- Run the Benchmark Suite ---
		benchLogger.Info("Running benchmark comparisons...", "tag", tag)
		resultsMap, suitePass, err := suite.RunAll() // RunAll should use the names provided in AddBenchmark
		if err != nil {
			benchLogger.Error("Benchmark suite failed", "tag", tag, "error", err)
			overallResults[tag] = []BenchmarkResult{
				{Metric: "Suite Error", Description: err.Error(), Passed: false},
			}
			overallFailedCount++
			continue
		}
		benchLogger.Debug("Benchmark suite run completed", "tag", tag, "overall_suite_pass", suitePass)

		// --- Store and Summarize Results ---
		var currentResults []BenchmarkResult
		if res, ok := resultsMap[tag]; ok { // Lookup using the name we added it with
			currentResults = res
			overallResults[tag] = currentResults
		} else {
			benchLogger.Warn("Did not find results for expected benchmark name in suite output", "tag", tag, "expectedName", tag)
			overallResults[tag] = []BenchmarkResult{
				{Metric: "Result Mismatch", Description: "Results not found under expected name", Passed: false},
			}
			overallFailedCount++
			continue
		}

		passedCount := 0
		failedCount := 0
		for _, res := range currentResults {
			if res.Passed {
				passedCount++
			} else {
				failedCount++
			}
		}
		overallPassedCount += passedCount
		overallFailedCount += failedCount

		benchLogger.Info("--- Benchmark Run Finished --- ", "tag", tag, "passed", passedCount, "failed", failedCount)

	} // End loop through discovered tags

	// --- Final Summary ---
	if len(discoveredTags) == 0 {
		// This case is handled earlier, but kept for safety
		benchLogger.Warn("No benchmarks were found or run.")
		os.Exit(0)
	}

	benchLogger.Info("--- Overall Benchmark Summary --- ")
	// Sort keys for consistent output order
	benchmarkNames := make([]string, 0, len(overallResults))
	for name := range overallResults {
		benchmarkNames = append(benchmarkNames, name)
	}
	//sort.Strings(benchmarkNames)

	for _, name := range benchmarkNames {
		results := overallResults[name]
		passed := 0
		failed := 0
		for _, r := range results {
			if r.Passed {
				passed++
			} else {
				failed++
			}
		}
		status := "PASS"
		if failed > 0 {
			status = "FAIL"
		}
		benchLogger.Info("Benchmark", "name", name, "status", status, "passed", passed, "failed", failed)
	}
	benchLogger.Info("--------------------------------", "Total Passed", overallPassedCount, "Total Failed", overallFailedCount)

	// --- Write Markdown Output ---
	if outputMarkdownPath != "" {
		// TODO: Re-implement or adapt markdown generation if needed.
		// The old formatResultsMarkdown likely relied on cfg.Benchmarks.
		// We might need a new function taking overallResults and reconstructing names/descriptions.
		// For now, commenting out the writing part to avoid errors.
		/*
			markdownContent := formatResultsMarkdown(overallResults) // Needs adaptation
			absMarkdownPath, err := filepath.Abs(outputMarkdownPath)
			if err != nil {
				benchLogger.Error("Failed to get absolute path for markdown output", "path", outputMarkdownPath, "error", err)
			} else {
				err = os.WriteFile(absMarkdownPath, []byte(markdownContent), 0644)
				if err != nil {
					benchLogger.Error("Failed to write markdown output file", "path", absMarkdownPath, "error", err)
				} else {
					benchLogger.Info("Benchmark results written to", "path", absMarkdownPath)
				}
			}
		*/
		benchLogger.Info("Markdown generation skipped (requires refactoring).")
	}

	// --- Exit Status ---
	if overallFailedCount > 0 {
		benchLogger.Error("Exiting with failure status due to benchmark errors.")
		os.Exit(1) // Exit with error code if any benchmarks failed
	}

	benchLogger.Info("All benchmarks passed. Exiting successfully.")
	os.Exit(0) // Exit successfully
}
