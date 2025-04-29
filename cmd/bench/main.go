package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	logf "github.com/zerodha/logf"
)

// Global logger instance
var benchLogger *logf.Logger

func main() {
	// --- Command Line Flags ---
	benchmarkTag := flag.String("tag", "", "Tag of the specific benchmark to run (optional, runs all enabled if empty).")
	resultsDir := flag.String("resultsdir", "", "Path to the directory containing simulation result files (e.g., MOTION.csv). Required.")
	outputMarkdownPath := flag.String("outputmd", "", "Path to write the benchmark results in Markdown format (e.g., BENCHMARK.md). Optional.")
	flag.Parse()

	// --- Load Configuration ---
	cfg, err := config.GetConfig()
	if err != nil {
		// Use a temporary basic logger for config loading errors
		basicLogger := logger.GetLogger("error") // Default to error if config fails
		basicLogger.Fatal("Failed to load configuration", "error", err)
	}

	// --- Initialize Logger with Config Level ---
	benchLogger = logger.GetLogger(cfg.Setup.Logging.Level)

	// --- Validate Required Flags ---
	if *resultsDir == "" {
		benchLogger.Fatal("Missing required flag: --resultsdir")
	}
	absResultsDir, err := filepath.Abs(*resultsDir)
	if err != nil {
		benchLogger.Fatal("Failed to get absolute path for --resultsdir", "path", *resultsDir, "error", err)
	}
	if _, err := os.Stat(absResultsDir); os.IsNotExist(err) {
		benchLogger.Fatal("Results directory specified by --resultsdir does not exist", "path", absResultsDir)
	}
	benchLogger.Info("Using simulation results directory", "path", absResultsDir)

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

		// --- Setup Benchmark Suite for this Benchmark ---
		benchConfig := BenchmarkConfig{
			BenchdataPath: absBenchdataPath, // Path to EXPECTED data
			ResultDirPath: absResultsDir,    // Path to ACTUAL data (from --resultsdir flag)
		}

		suite := NewBenchmarkSuite(benchConfig)

		// --- Register Specific Benchmark Implementation ---
		// Use tag to select benchmark
		var benchmarkToRun Benchmark
		switch tag {
		case "hipr-euroc24":
			benchmarkToRun = NewHiprEuroc24Benchmark(benchConfig) // Use constructor
		// Add cases for other benchmark tags here
		// case "another-tag":
		// 	benchmarkToRun = NewAnotherBenchmark(benchConfig)
		default:
			benchLogger.Warn("Skipping benchmark: Unknown benchmark tag in config", "tag", tag)
			continue
		}
		suite.AddBenchmark(benchmarkToRun) // Use AddBenchmark

		// --- Run the Benchmark Suite ---
		benchLogger.Info("Running benchmark comparisons...", "tag", tag)
		// RunAll returns map[string][]BenchmarkResult, bool, error
		resultsMap, suitePass, err := suite.RunAll()
		if err != nil {
			benchLogger.Error("Benchmark suite failed", "tag", tag, "error", err)
			// Store a generic failure result for this benchmark
			overallResults[benchmarkEntry.Name] = []BenchmarkResult{ // Use benchmarkEntry.Name as key
				{
					Metric:      "Suite Error",
					Description: err.Error(),
					Passed:      false, // Mark as failed
					// Other fields default to zero/empty
				},
			}
			overallFailedCount++
			continue // Move to the next benchmark
		}
		benchLogger.Debug("Benchmark suite run completed", "tag", tag, "overall_suite_pass", suitePass) // Log the returned status

		// --- Store and Summarize Results for this Benchmark ---
		// Since RunAll returns a map, and we added only one benchmark, get its results
		var currentResults []BenchmarkResult
		if res, ok := resultsMap[benchmarkEntry.Name]; ok { // Use Name from config entry for lookup
			currentResults = res
			overallResults[benchmarkEntry.Name] = currentResults // Store under Name
		} else {
			benchLogger.Warn("Did not find results for expected benchmark name in suite output", "tag", tag, "expectedName", benchmarkEntry.Name)
			// Store a generic failure result?
			overallResults[benchmarkEntry.Name] = []BenchmarkResult{
				{Metric: "Result Mismatch", Description: "Results not found under expected name", Passed: false},
			}
			overallFailedCount++
			continue
		}

		passedCount := 0
		failedCount := 0
		for _, res := range currentResults {
			if res.Passed { // Use 'Passed' field
				passedCount++
			} else {
				failedCount++
			}
		}
		overallPassedCount += passedCount
		overallFailedCount += failedCount

		benchLogger.Info("--- Benchmark Run Finished --- ", "tag", tag, "name", benchmarkEntry.Name, "passed", passedCount, "failed", failedCount)

	} // End loop through configured benchmarks

	// --- Final Summary ---
	if !hasRunBenchmark {
		benchLogger.Warn("No benchmarks were selected or enabled to run.")
		// Decide exit code? Exit 0 for now if nothing ran.
		os.Exit(0)
	}

	benchLogger.Info("--- Overall Benchmark Summary --- ")
	for name, results := range overallResults { // Iterate over Name
		passed := 0
		failed := 0
		for _, r := range results {
			if r.Passed { // Use 'Passed' field
				passed++
			} else {
				failed++
			}
		}
		status := "PASS"
		if failed > 0 {
			status = "FAIL"
		}
		benchLogger.Info("Benchmark", "name", name, "status", status, "passed", passed, "failed", failed) // Log Name
		// Print detailed results if needed (already printed during run?)
		// for _, r := range results {
		// 	benchLogger.Debug("Detail", "name", name, "metric", r.Metric, "status", r.Passed, "description", r.Description)
		// }
	}
	benchLogger.Info("--------------------------------", "Total Passed", overallPassedCount, "Total Failed", overallFailedCount)

	// --- Write Markdown Output (if requested) ---
	if *outputMarkdownPath != "" {
		markdownContent := formatResultsMarkdown(overallResults, cfg.Benchmarks)
		absMarkdownPath, err := filepath.Abs(*outputMarkdownPath)
		if err != nil {
			benchLogger.Error("Failed to get absolute path for markdown output", "path", *outputMarkdownPath, "error", err)
			// Continue to exit status, but log the error
		} else {
			err = os.WriteFile(absMarkdownPath, []byte(markdownContent), 0644)
			if err != nil {
				benchLogger.Error("Failed to write benchmark results to markdown file", "path", absMarkdownPath, "error", err)
			} else {
				benchLogger.Info("Benchmark results written to markdown file", "path", absMarkdownPath)
			}
		}
	}

	// --- Exit with Status Code ---
	if overallFailedCount > 0 {
		benchLogger.Error("One or more benchmarks failed.")
		os.Exit(1)
	}

	benchLogger.Info("All benchmarks passed.")
	os.Exit(0)
}

// formatResultsMarkdown generates a markdown report from the benchmark results.
func formatResultsMarkdown(results map[string][]BenchmarkResult, benchmarks map[string]config.BenchmarkEntry) string {
	var builder strings.Builder

	// Timestamp
	builder.WriteString(fmt.Sprintf("## Benchmark Results (%s)\n\n", time.Now().Format(time.RFC1123)))

	// Summary Table
	builder.WriteString("### Summary\n\n")
	builder.WriteString("| Name | Status | Passed | Failed |\n") // Use Name
	builder.WriteString("|------|--------|--------|--------|\n")
	overallPassed := 0
	overallFailed := 0
	// Need to iterate consistently, perhaps get keys and sort?
	names := make([]string, 0, len(results))
	for name := range results {
		names = append(names, name)
	}
	sort.Strings(names) // Need to import "sort"

	for _, name := range names {
		resList := results[name]
		passed := 0
		failed := 0
		for _, r := range resList {
			if r.Passed { // Use 'Passed'
				passed++
			} else {
				failed++
			}
		}
		status := "PASS"
		if failed > 0 {
			status = "FAIL"
		}
		builder.WriteString(fmt.Sprintf("| %s | %s | %d | %d |\n", name, status, passed, failed))
		overallPassed += passed
		overallFailed += failed
	}
	status := "PASS"
	if overallFailed > 0 {
		status = "FAIL"
	}
	builder.WriteString(fmt.Sprintf("| **Overall** | **%s** | **%d** | **%d** |\n\n", status, overallPassed, overallFailed))

	// Detailed Results per Benchmark
	builder.WriteString("### Details\n\n")
	for _, name := range names { // Iterate using sorted names
		resList := results[name]
		// Find the corresponding tag (less efficient, maybe restructure results map?)
		var tag string
		for t, entry := range benchmarks {
			if entry.Name == name {
				tag = t
				break
			}
		}
		builder.WriteString(fmt.Sprintf("#### %s (%s)\n\n", name, tag))                                      // Show Name and Tag
		builder.WriteString("| Metric | Status | Expected | Actual | Diff | Tolerance | Type | Details |\n") // Restore columns
		builder.WriteString("|--------|--------|----------|--------|------|-----------|------|---------|\n")
		for _, r := range resList {
			statusIcon := "PASS"
			if !r.Passed { // Use 'Passed'
				statusIcon = "FAIL"
			}
			// Escape potential pipe characters in the message for table rendering
			safeDescription := strings.ReplaceAll(r.Description, "|", "\\")
			builder.WriteString(fmt.Sprintf("| %s | %s | %.3f | %.3f | %.3f | %.3f | %s | %s |\n",
				r.Metric, statusIcon, r.Expected, r.Actual, r.Difference, r.Tolerance, r.ToleranceType, safeDescription))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
