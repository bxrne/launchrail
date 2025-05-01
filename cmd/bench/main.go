package main

import (
	"os"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/diff"
	logf "github.com/zerodha/logf"
)

// Global logger instance
var benchLogger *logf.Logger

func main() {
	// Markdown output path
	outputMarkdownPath := "BENCHMARK.md"

	// --- Load Configuration ---
	cfg, err := config.GetConfig()
	if err != nil {
		basicLogger := logger.GetLogger("error")
		basicLogger.Fatal("Failed to load configuration", "error", err)
	}

	// --- Initialize Logger with Config Level ---
	benchLogger = logger.GetLogger(cfg.Setup.Logging.Level)

	// --- Determine Paths ---
	// Simulation results directory must be specified in config
	simOutDir := cfg.Setup.App.SimulationOutputDir
	if simOutDir == "" {
		benchLogger.Fatal("setup.app.simulation_output_dir must be set in config")
	}
	absSimulationResultsDir, err := filepath.Abs(simOutDir)
	if err != nil {
		benchLogger.Fatal("Failed to get absolute path for simulation results directory", "path", absSimulationResultsDir, "error", err)
	}
	if _, err := os.Stat(absSimulationResultsDir); os.IsNotExist(err) {
		benchLogger.Warn("Simulation results directory does not exist yet, benchmarks might fail if they require it", "path", absSimulationResultsDir)
	} else {
		benchLogger.Info("Using simulation results directory", "path", absSimulationResultsDir)
	}

	// Ensure simulation results directory exists
	if err := os.MkdirAll(absSimulationResultsDir, 0o755); err != nil {
		benchLogger.Fatal("Failed to create simulation results directory", "path", absSimulationResultsDir, "error", err)
	}

	// Discover enabled benchmarks from config
	var discoveredTags []string
	for tag := range cfg.Benchmarks {
		if cfg.Benchmarks[tag].Enabled {
			discoveredTags = append(discoveredTags, tag)
		}
	}
	if len(discoveredTags) == 0 {
		benchLogger.Warn("No enabled benchmarks in config")
		os.Exit(0)
	}
	benchLogger.Info("Discovered enabled benchmark tags", "tags", discoveredTags)

	// --- Initialize Overall Results ---
	overallResults := make(map[string][]BenchmarkResult)
	overallFailedCount := 0
	overallPassedCount := 0

	// --- Loop Through Discovered Benchmark Tags ---
	for _, tag := range discoveredTags {
		benchLogger.Info("--- Starting Benchmark Run ---", "tag", tag)

		// Determine benchmark data path from config
		currentBenchmarkDataPath := cfg.Benchmarks[tag].DataDir
		benchLogger.Info("Using benchmark data path for tag", "tag", tag, "path", currentBenchmarkDataPath)

		// Validate benchmark data path exists
		absBenchmarkDataPath, err := filepath.Abs(currentBenchmarkDataPath)
		if err != nil {
			benchLogger.Fatal("Failed to get absolute path for benchmark data directory", "path", currentBenchmarkDataPath, "error", err)
		}
		if _, err := os.Stat(absBenchmarkDataPath); os.IsNotExist(err) {
			benchLogger.Error("Benchmark tag data directory not found, skipping", "tag", tag, "path", currentBenchmarkDataPath)
			// Store a failure result
			overallResults[tag] = []BenchmarkResult{
				{Metric: "Setup Error", Description: "Data directory not found", Passed: false},
			}
			overallFailedCount++
			continue
		}

		// Run a simulation for this benchmark into its own run folder
		baseDir := absSimulationResultsDir
		// Generate unique run ID
		// Get associated design file
		designFilePath := cfg.Benchmarks[tag].DesignFile
		designFileBytes, err := os.ReadFile(designFilePath)
		if err != nil {
			benchLogger.Fatal("Failed to read design file", "path", designFilePath, "error", err)
		}
		hash := diff.CombinedHash(cfg.Bytes(), designFileBytes)
		runDir := filepath.Join(baseDir, hash)
		// Create run directory
		if err := os.MkdirAll(runDir, 0o755); err != nil {
			benchLogger.Fatal("Failed to create run directory", "path", runDir, "error", err)
		}
		benchLogger.Info("Simulation run directory created", "tag", tag, "runID", hash)

		// Setup simulation storage and manager
		motionStore, err := storage.NewStorage(runDir, storage.MOTION)
		if err != nil {
			benchLogger.Fatal("Failed to create motion storage", "error", err)
		}
		eventsStore, err := storage.NewStorage(runDir, storage.EVENTS)
		if err != nil {
			motionStore.Close()
			benchLogger.Fatal("Failed to create events storage", "error", err)
		}
		dynamicsStore, err := storage.NewStorage(runDir, storage.DYNAMICS)
		if err != nil {
			motionStore.Close()
			eventsStore.Close()
			benchLogger.Fatal("Failed to create dynamics storage", "error", err)
		}
		stores := &storage.Stores{Motion: motionStore, Events: eventsStore, Dynamics: dynamicsStore}
		simManager := simulation.NewManager(cfg, *benchLogger)
		if err := simManager.Initialize(stores); err != nil {
			benchLogger.Fatal("Simulation initialization failed", "error", err)
		}
		if err := simManager.Run(); err != nil {
			benchLogger.Fatal("Simulation run failed", "error", err)
		}
		if err := simManager.Close(); err != nil {
			benchLogger.Warn("Simulation close error", "error", err)
		}

		// --- Setup Benchmark Suite for this Benchmark ---
		benchConfig := BenchmarkConfig{BenchdataPath: currentBenchmarkDataPath, ResultDirPath: runDir}

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
