package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bxrne/launchrail/internal/config" // Import config package
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
)

// benchmarkRegistry maps benchmark tags/names to their constructor functions.
// This allows dynamic instantiation based on config.
var benchmarkRegistry = map[string]func(BenchmarkConfig) Benchmark{
	"hipr-euroc24": func(config BenchmarkConfig) Benchmark { // Wrap the constructor
		return NewHiprEuroc24Benchmark(config) // Return the concrete type which implements the interface
	},
	// Add other benchmarks here following the same pattern:
	// "tag-name": func(config BenchmarkConfig) Benchmark { return NewOtherBenchmarkConstructor(config) },
}

func main() {
	// --- Command Line Flags --- -> Replaced with tag flag
	benchmarkTag := flag.String("tag", "", "Tag of the specific benchmark to run (optional, runs all enabled if empty).")
	flag.Parse()

	// --- Load Configuration --- (Moved earlier)
	cfg, err := config.GetConfig()
	if err != nil {
		// Use a temporary basic logger for config loading errors
		basicLogger := logger.GetLogger("error") // Default to error if config fails
		basicLogger.Fatal("Failed to load configuration", "error", err)
	}

	// --- Initialize Logger with Config Level --- (Moved earlier)
	benchLogger := logger.GetLogger(cfg.Setup.Logging.Level)

	// --- Initialize Overall Results --- 
	overallResults := make(map[string][]BenchmarkResult)
	overallFailedCount := 0
	overallPassedCount := 0
	hasRunBenchmark := false // Track if any benchmark was actually selected and run

	// --- Loop Through Configured Benchmarks --- 
	benchLogger.Info("Processing configured benchmarks...", "requested_tag", *benchmarkTag)
	for tag, benchmarkEntry := range cfg.Benchmarks {
		// Skip if a specific tag is requested and it doesn't match
		if *benchmarkTag != "" && *benchmarkTag != tag {
			benchLogger.Debug("Skipping benchmark: tag mismatch", "tag", tag, "requested_tag", *benchmarkTag)
			continue
		}

		// Skip if benchmark is not enabled in config
		if !benchmarkEntry.Enabled {
			benchLogger.Info("Skipping benchmark: disabled in config", "tag", tag, "name", benchmarkEntry.Name)
			continue
		}

		benchLogger.Info("--- Starting Benchmark Run --- ", "tag", tag, "name", benchmarkEntry.Name)
		hasRunBenchmark = true

		// --- Validate and Resolve Paths from Config --- 
		absDesignFilePath, err := filepath.Abs(benchmarkEntry.DesignFile)
		if err != nil {
			benchLogger.Error("Error getting absolute path for benchmark design file", "tag", tag, "path", benchmarkEntry.DesignFile, "error", err)
			continue // Skip this benchmark
		}
		if _, err := os.Stat(absDesignFilePath); os.IsNotExist(err) {
			benchLogger.Error("Benchmark design file not found", "tag", tag, "path", absDesignFilePath)
			continue // Skip this benchmark
		}
		benchLogger.Info("Using design file for simulation", "tag", tag, "path", absDesignFilePath)

		absBenchdataPath, err := filepath.Abs(benchmarkEntry.DataDir)
		if err != nil {
			benchLogger.Error("Error getting absolute path for benchmark data directory", "tag", tag, "path", benchmarkEntry.DataDir, "error", err)
			continue // Skip this benchmark
		}
		if _, err := os.Stat(absBenchdataPath); os.IsNotExist(err) {
			benchLogger.Error("Benchmark data directory not found", "tag", tag, "path", absBenchdataPath)
			continue // Skip this benchmark
		}
		benchLogger.Info("Using benchmark data directory", "tag", tag, "path", absBenchdataPath)
		
		// --- Update Config with Current Benchmark's ORK File Path ---
		// Create a copy of the engine config to avoid modifying the original for subsequent benchmarks
		currentEngineConfig := cfg.Engine
		currentEngineConfig.Options.OpenRocketFile = absDesignFilePath
		benchLogger.Info("Updated engine config with benchmark ORK file path", "tag", tag)

		// --- Initialize and Run Simulation for this Benchmark --- 
		benchLogger.Info("Initializing simulation manager for benchmark...", "tag", tag)
		// Create a temporary config with the updated engine settings for this run
		tempCfg := *cfg // Shallow copy is okay here as we only change Engine
		tempCfg.Engine = currentEngineConfig

		simManager := simulation.NewManager(&tempCfg, *benchLogger)
		if err := simManager.Initialize(); err != nil {
			benchLogger.Error("Failed to initialize simulation manager for benchmark", "tag", tag, "error", err)
			simManager.Close() // Attempt cleanup
			continue // Skip this benchmark
		}

		simHash := simManager.GetSimHash()
		if simHash == "" {
			benchLogger.Error("Failed to retrieve simulation hash after initialization", "tag", tag)
			simManager.Close() // Attempt cleanup
			continue // Skip this benchmark
		}
		benchLogger.Info("Simulation initialized for benchmark", "tag", tag, "recordHash", simHash)

		benchLogger.Info("Running simulation for benchmark...", "tag", tag)
		if err := simManager.Run(); err != nil {
			benchLogger.Error("Simulation run failed for benchmark", "tag", tag, "error", err)
			simManager.Close() // Attempt cleanup
			continue // Skip this benchmark
		}
		benchLogger.Info("Simulation run completed successfully for benchmark", "tag", tag)

		// Close simulation manager before accessing results
		if err := simManager.Close(); err != nil {
			benchLogger.Error("Error closing simulation manager for benchmark", "tag", tag, "error", err)
			// Continue, as simulation might have finished, but benchmark might fail
		}

		// --- Setup Benchmark Suite for this Single Benchmark ---
		benchLogger.Info("Initializing record manager for benchmark results...", "tag", tag)
		rm, err := storage.NewRecordManager(cfg.Setup.App.BaseDir)
		if err != nil {
			benchLogger.Error("Failed to initialize record manager for benchmark", "tag", tag, "error", err)
			continue // Skip this benchmark
		}

		benchmarkConfig := BenchmarkConfig{
			BenchdataPath: absBenchdataPath,
			SimRecordHash: simHash,
			RecordManager: rm,
		}
		suite := NewBenchmarkSuite(benchmarkConfig) // Suite for just this one benchmark run

		// Get the correct benchmark constructor from the registry
		constructor, exists := benchmarkRegistry[tag]
		if !exists {
			benchLogger.Error("No benchmark implementation registered for tag", "tag", tag)
			continue // Skip this benchmark
		}
		suite.AddBenchmark(constructor(benchmarkConfig))

		// --- Run This Single Benchmark --- 
		benchLogger.Info("Starting benchmark comparison...", "tag", tag)
		results, _, err := suite.RunAll() // Run the suite containing only the current benchmark
		if err != nil {
			benchLogger.Error("Error running benchmark comparison", "tag", tag, "error", err)
			// Don't exit immediately, record failure and continue if possible
			overallFailedCount++ // Assume failure if the run errors out
			continue
		}

		// --- Merge Results and Count Pass/Fail --- 
		benchLogger.Info("--- Benchmark Results Summary --- ", "tag", tag, "name", benchmarkEntry.Name)
		benchmarkFailed := false
		for benchmarkName, benchmarkResults := range results { // Should only be one entry from the single-benchmark suite
			overallResults[benchmarkName] = benchmarkResults // Add/overwrite in overall map
			for _, res := range benchmarkResults {
				if !res.Passed {
					overallFailedCount++
					benchmarkFailed = true
				} else {
					overallPassedCount++
				}
				// Log individual metric results here if desired (using benchLogger)
				statusStr := "PASSED"
				if !res.Passed { statusStr = "FAILED" }
				benchLogger.Debug("Metric result", "tag", tag, "benchmark", benchmarkName, "metric", res.Metric, "status", statusStr)
			}
		}
		if benchmarkFailed {
			benchLogger.Error("Benchmark run finished with failed metrics", "tag", tag, "name", benchmarkEntry.Name)
		} else {
			benchLogger.Info("Benchmark run finished successfully", "tag", tag, "name", benchmarkEntry.Name)
		}

	} // End of benchmark loop

	// --- Check if any benchmark was run --- 
	// Restore this check before proceeding
	if !hasRunBenchmark {
		if *benchmarkTag != "" {
			benchLogger.Error("No enabled benchmark found matching the requested tag", "tag", *benchmarkTag)
		} else {
			benchLogger.Warn("No enabled benchmarks found in the configuration.")
		}
		os.Exit(1) // Exit with error if no benchmarks were run
	}

	// --- Get Git Commit Hash --- 
	commitHash := "N/A"
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		commitHash = strings.TrimSpace(out.String())
	} else {
		benchLogger.Warn("Could not get git commit hash", "error", err)
	}

	// --- Format and Print Overall Results to Stdout --- 
	markdownOutput := formatResultsToMarkdown(overallResults, commitHash, cfg) // Pass hash and cfg
	fmt.Println("\n--- MARKDOWN OUTPUT START ---") // Marker for easy capture
	fmt.Print(markdownOutput)
	fmt.Println("--- MARKDOWN OUTPUT END ---")

	// --- Print Overall Summary to Stdout/Logs --- 
	benchLogger.Info("--- Overall Benchmark Suite Summary ---") // Log summary
	fmt.Println("\n--- Overall Summary ---")                 // Stdout summary
	fmt.Printf("Total Metrics Passed: %d\n", overallPassedCount)
	fmt.Printf("Total Metrics Failed: %d\n", overallFailedCount)

	if overallFailedCount > 0 {
		benchLogger.Error("Benchmark suite finished with failed metrics", "count", overallFailedCount) // Log final status
		fmt.Println("\nStatus: FAILED")
		os.Exit(1) // Exit with non-zero status code if any metric failed
	} else {
		benchLogger.Info("Benchmark suite finished successfully.") // Log final status
		fmt.Println("\nStatus: PASSED")
		os.Exit(0) // Exit with zero status code if all metrics passed
	}
}

// formatResultsToMarkdown formats the benchmark results into a Markdown string.
// Includes commit hash and config info.
func formatResultsToMarkdown(results map[string][]BenchmarkResult, commitHash string, cfg *config.Config) string {
	var markdown bytes.Buffer

	// --- Header --- 
	now := time.Now().Format(time.RFC1123)
	markdown.WriteString("# Benchmark Results\n\n")
	markdown.WriteString(fmt.Sprintf("**Date:** %s\n", now))
	markdown.WriteString(fmt.Sprintf("**Commit:** %s\n", commitHash))
	if cfg != nil && len(cfg.Setup.Plugins.Paths) > 0 {
		markdown.WriteString(fmt.Sprintf("**Plugins:** `%s`\n", strings.Join(cfg.Setup.Plugins.Paths, ", ")))
	}
	markdown.WriteString("\n")

	// --- Table of Contents --- 
	if len(results) > 1 { // Only show TOC if more than one benchmark
		markdown.WriteString("## Table of Contents\n")
		// Sort benchmark names for consistent TOC order
		benchmarkNames := make([]string, 0, len(results))
		for name := range results {
			benchmarkNames = append(benchmarkNames, name)
		}
		sort.Strings(benchmarkNames) // Requires importing "sort"

		for _, benchmarkName := range benchmarkNames {
			// Generate GitHub-compatible anchor link (lowercase, spaces to dashes, remove non-alphanumeric)
			anchor := strings.ToLower(benchmarkName)
			anchor = strings.ReplaceAll(anchor, " ", "-")
			// Basic alphanumeric filtering for anchor
			reg := regexp.MustCompile(`[^a-z0-9-]`) // Requires importing "regexp"
			anchor = reg.ReplaceAllString(anchor, "")
			markdown.WriteString(fmt.Sprintf("- [%s](#%s)\n", benchmarkName, anchor))
		}
		markdown.WriteString("\n")
	}

	// --- Benchmark Details --- 
	// Sort names again for consistent output order
	benchmarkNames := make([]string, 0, len(results))
	for name := range results {
		benchmarkNames = append(benchmarkNames, name)
	}
	sort.Strings(benchmarkNames) // Requires importing "sort"

	for _, benchmarkName := range benchmarkNames {
		benchmarkResults := results[benchmarkName]
		markdown.WriteString(fmt.Sprintf("## %s\n\n", benchmarkName))
		markdown.WriteString("| Metric        | Description   | Expected | Actual   | Diff     | Tolerance | Type     | Status |\n")
		markdown.WriteString("|---------------|---------------|----------|----------|----------|-----------|----------|--------|\n")

		for _, res := range benchmarkResults {
			status := ":white_check_mark: PASSED"
			if !res.Passed {
				status = ":x: FAILED"
			}
			// Format floats for better readability
			// Using %.3f for precision, adjust as needed
			expectedStr := fmt.Sprintf("%.3f", res.Expected)
			actualStr := fmt.Sprintf("%.3f", res.Actual)
			diffStr := fmt.Sprintf("%.3f", res.Difference)
			toleranceStr := fmt.Sprintf("%.3f", res.Tolerance)

			// Escape pipe characters in description AND metric if any
			metric := strings.ReplaceAll(res.Metric, "|", "\\|") // <-- Added escaping for Metric
			description := strings.ReplaceAll(res.Description, "|", "\\|")

			markdown.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s | %s |\n",
				metric, // Use escaped metric
				description, expectedStr, actualStr, diffStr, toleranceStr, res.ToleranceType, status))
		}
		markdown.WriteString("\n") // Add space between benchmarks
	}

	return markdown.String()
}
