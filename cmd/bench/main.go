package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/template"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	logf "github.com/zerodha/logf"
)

// Global logger instance
var benchLogger *logf.Logger

const reportTemplate = `
# Benchmark Results: {{.Tag}}

**Status:** {{.OverallStatus}} (Passed: {{.PassedCount}}, Failed: {{.FailedCount}})

| Metric | Expected | Actual | Difference | Tolerance | Tol Type | Status |
|---|---|---|---|---|---|---|{{range .Results}}
| {{.Metric}} | {{printf "%.4f" .Expected}} | {{printf "%.4f" .Actual}} | {{printf "%.4f" .Difference}} | {{printf "%.4f" .Tolerance}} | {{.ToleranceType}} | {{if .Passed}}PASS{{else}}FAIL{{end}} |{{end}}
`

type reportData struct {
	Tag           string
	OverallStatus string
	PassedCount   int
	FailedCount   int
	Results       []BenchmarkResult
}

// writeMarkdownReport generates a markdown report file in the specified run directory.
func writeMarkdownReport(runDir, tag string, results []BenchmarkResult) error {
	passedCount := 0
	failedCount := 0
	for _, r := range results {
		if r.Passed {
			passedCount++
		} else {
			failedCount++
		}
	}

	status := "PASS"
	if failedCount > 0 {
		status = "FAIL"
	}

	data := reportData{
		Tag:           tag,
		OverallStatus: status,
		PassedCount:   passedCount,
		FailedCount:   failedCount,
		Results:       results,
	}

	tmpl, err := template.New("report").Parse(reportTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse markdown template: %w", err)
	}

	var reportBuf bytes.Buffer
	if err := tmpl.Execute(&reportBuf, data); err != nil {
		return fmt.Errorf("failed to execute markdown template: %w", err)
	}

	// Trim leading/trailing whitespace potentially added by template
	reportContent := bytes.TrimSpace(reportBuf.Bytes())

	reportPath := filepath.Join(runDir, "result.md")
	if err := os.WriteFile(reportPath, reportContent, 0644); err != nil {
		return fmt.Errorf("failed to write markdown report to %s: %w", reportPath, err)
	}
	return nil
}

func main() {
	// --- Setup Logger ---
	// Initialize logger early for setup messages
	// Use GetLogger with a default level initially
	benchLogger = logger.GetLogger("debug") // Use debug for benchmark setup

	// --- Load Configuration ---
	cfg, err := config.GetConfig()
	if err != nil {
		benchLogger.Fatal("Failed to load configuration", "error", err)
	}
	// Re-initialize logger with config level
	benchLogger = logger.GetLogger(cfg.Setup.Logging.Level)

	benchLogger.Info("--- Starting Benchmark Run --- ")

	// --- Create Base Output Directory ---
	homeDir, err := os.UserHomeDir()
	if err != nil {
		benchLogger.Fatal("Failed to get user home directory", "error", err)
	}
	timestamp := time.Now().Format("20060102-150405")
	baseOutputDir := filepath.Join(homeDir, ".launchrail", "benchmarks", timestamp)
	if err := os.MkdirAll(baseOutputDir, 0755); err != nil {
		benchLogger.Fatal("Failed to create base output directory", "path", baseOutputDir, "error", err)
	}
	benchLogger.Info("Benchmark results will be stored in", "directory", baseOutputDir)

	// --- Run Benchmarks from Config ---
	overallResults := make(map[string][]BenchmarkResult)
	overallPassedCount := 0
	overallFailedCount := 0

	benchmarkFound := false
	for tag, benchmarkEntry := range cfg.Benchmarks {
		if !benchmarkEntry.Enabled {
			benchLogger.Info("Skipping disabled benchmark", "tag", tag)
			continue
		}
		benchmarkFound = true
		benchLogger.Info("--- Running Benchmark --- ", "tag", tag, "name", benchmarkEntry.Name)

		// Create specific directory for this run
		runDir := filepath.Join(baseOutputDir, tag)
		if err := os.MkdirAll(runDir, 0755); err != nil {
			benchLogger.Error("Failed to create benchmark run directory", "path", runDir, "error", err)
			overallResults[tag] = []BenchmarkResult{
				{Metric: "Setup Error", Description: "Failed to create run directory", Passed: false},
			}
			overallFailedCount++
			continue
		}

		// --- Instantiate the correct benchmark based on tag ---
		var benchmark Benchmark // Use the interface type
		switch tag {
		case "hipr-euroc24":
			benchmark = &HiprEuroc24Benchmark{}
		// Add cases for other benchmarks here
		// case "another-benchmark-tag":
		// 	 benchmark = &AnotherBenchmarkImplementation{}
		default:
			benchLogger.Error("No benchmark implementation found for tag", "tag", tag)
			overallResults[tag] = []BenchmarkResult{
				{Metric: "Configuration Error", Description: "No implementation for benchmark tag", Passed: false},
			}
			overallFailedCount++
			continue
		}

		// --- Execute the Benchmark ---
		// NOTE: The Run signature needs to be updated in benchmark.go and hipr_euroc24.go
		// Expected new signature: Run(cfg config.BenchmarkEntry, logger *logf.Logger, runDir string) ([]BenchmarkResult, error)
		currentResults, err := benchmark.Run(benchmarkEntry, benchLogger, runDir) // Pass config and runDir. Lint error 4463794b-e453-41f8-91f5-eb464281b522 is expected here and will be fixed by updating the interface.
		if err != nil {
			benchLogger.Error("Benchmark execution failed", "tag", tag, "error", err)
			// Ensure some result is recorded even on error
			if len(currentResults) == 0 {
				currentResults = append(currentResults, BenchmarkResult{
					Metric: "Execution Error", Description: err.Error(), Passed: false,
				})
			} else {
				// Mark existing results as failed if an overarching error occurred
				for i := range currentResults {
					currentResults[i].Passed = false
				}
			}
		}
		overallResults[tag] = currentResults

		// --- Write Markdown Report ---
		if err := writeMarkdownReport(runDir, tag, currentResults); err != nil {
			benchLogger.Error("Failed to write markdown report", "tag", tag, "runDir", runDir, "error", err)
			// Optionally treat report writing failure as a benchmark failure
			// overallFailedCount++ // Uncomment if needed
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

	}

	if !benchmarkFound {
		benchLogger.Warn("No enabled benchmarks found in configuration.")
	}

	// --- Final Summary ---
	fmt.Println("\n--- Overall Benchmark Summary ---") // Print header to stdout
	// Sort keys for consistent output order
	benchmarkNames := make([]string, 0, len(overallResults))
	for name := range overallResults {
		benchmarkNames = append(benchmarkNames, name)
	}
	sort.Strings(benchmarkNames)

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
		fmt.Printf("Benchmark [%s]: %s (Passed: %d, Failed: %d)\n", name, status, passed, failed)
	}

	fmt.Println("--------------------------------") // Separator
	fmt.Printf("Total Passed: %d, Total Failed: %d\n", overallPassedCount, overallFailedCount)
	fmt.Println("--------------------------------") // Footer separator

	// --- Exit Status ---
	if overallFailedCount > 0 {
		benchLogger.Error("Exiting with failure status due to benchmark errors.")
		os.Exit(1) // Exit with error code if any benchmarks failed
	}

	benchLogger.Info("All benchmarks passed. Exiting successfully.")
	os.Exit(0) // Exit successfully
}
