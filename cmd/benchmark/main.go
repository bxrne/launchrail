package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/olekukonko/tablewriter"
	logf "github.com/zerodha/logf"
)

// BenchmarkServer manages and runs benchmark suites.
// It uses the Benchmark interface type.
// Ensure types.go in package main defines this interface.
type BenchmarkServer struct {
	log             *logf.Logger
	benchmarkSuites []Benchmark
	appConfig       *config.Config
	benchdataPath   string
}

// NewBenchmarkServer creates a new server for managing benchmarks.
func NewBenchmarkServer(lg *logf.Logger, appCfg *config.Config, benchdataPath string) *BenchmarkServer {
	return &BenchmarkServer{
		log:             lg,
		benchmarkSuites: make([]Benchmark, 0),
		appConfig:       appCfg,
		benchdataPath:   benchdataPath,
	}
}

// RegisterSuite adds a new benchmark suite to the server.
// It accepts any type that implements the Benchmark interface.
func (s *BenchmarkServer) RegisterSuite(suite Benchmark) {
	s.benchmarkSuites = append(s.benchmarkSuites, suite)
	s.log.Info("Registered benchmark suite", "name", suite.Name())
}

// RunAll executes all registered benchmark suites and returns their results.
func (s *BenchmarkServer) RunAll() []BenchmarkResult {
	var allResults []BenchmarkResult
	s.log.Info("Starting execution of all registered benchmark suites.")

	for _, suite := range s.benchmarkSuites {
		s.log.Info("Running benchmark suite", "name", suite.Name())
		startTime := time.Now()

		// The Benchmark's Run method is responsible for its own specific configuration loading
		// using the provided appConfig and base benchdataPath.
		result, err := suite.Run(s.appConfig, s.benchdataPath)

		duration := time.Since(startTime)

		if err != nil {
			s.log.Error("Benchmark suite run failed with error", "name", suite.Name(), "error", err)
			// Ensure a result object exists to record the error
			if result == nil {
				result = &BenchmarkResult{Name: suite.Name(), Passed: false, RunError: err, Duration: duration}
			} else {
				result.Passed = false
				result.RunError = err // Ensure error is captured if result was partially populated
				result.Duration = duration
			}
		} else if result == nil {
			s.log.Error("Benchmark suite returned nil result without error", "name", suite.Name())
			result = &BenchmarkResult{Name: suite.Name(), Passed: false, RunError: fmt.Errorf("benchmark returned nil result"), Duration: duration}
		} else {
			result.Duration = duration // Ensure duration is set on successful runs too
			s.log.Info("Benchmark suite completed", "name", suite.Name(), "passed", result.Passed, "duration", duration)
		}
		allResults = append(allResults, *result)
	}

	s.log.Info("Finished execution of all benchmark suites.")
	return allResults
}

// printDetailedBenchmarkResult prints a detailed view of a single benchmark result, including individual metrics.
func (s *BenchmarkServer) printDetailedBenchmarkResult(result *BenchmarkResult) {
	fmt.Printf("\n--- Benchmark Details: %s ---\n", result.Name)
	status := "PASSED"
	if !result.Passed {
		status = "FAILED"
	}
	fmt.Printf("Overall Status: %s | Duration: %s\n", status, result.Duration.Truncate(time.Millisecond))

	if result.SetupError != nil {
		fmt.Printf("Setup Error: %v\n", result.SetupError)
	}
	if result.RunError != nil {
		fmt.Printf("Run Error: %v\n", result.RunError)
	}

	if len(result.Metrics) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"Metric", "Expected", "Actual", "Status", "Difference", "Error"})

		for _, metric := range result.Metrics {
			metricStatus := "PASS"
			if !metric.Passed {
				metricStatus = "FAIL"
			}

			expectedStr := fmt.Sprintf("%v", metric.Expected)
			actualStr := fmt.Sprintf("%v", metric.Actual)
			errorStr := ""
			if metric.Error != nil {
				errorStr = metric.Error.Error()
			}

			_ = table.Append([]string{
				metric.Name,
				expectedStr,
				actualStr,
				metricStatus,
				metric.Diff,
				errorStr,
			})
		}
		_ = table.Render()
	}

	if result.ReportNotes != "" {
		fmt.Printf("Notes: %s\n", result.ReportNotes)
	}
	fmt.Println(strings.Repeat("-", 50)) // Use strings.Repeat for clarity
}

// printTable renders a summary of benchmark results in a formatted table.
func (s *BenchmarkServer) printTable(results []BenchmarkResult) {
	fmt.Printf("\n--- Benchmark Summary ---\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Name", "Passed", "Duration", "Error"})

	for _, result := range results {
		var errorStr string
		if result.RunError != nil {
			errorStr = result.RunError.Error()
		}

		_ = table.Append([]string{
			result.Name,
			fmt.Sprintf("%t", result.Passed),
			result.Duration.String(),
			errorStr,
		})
	}
	_ = table.Render()
}

func main() {
	appCfg, err := config.GetConfig()
	if err != nil {
		// Use a default logger for this critical startup error
		// as the main logger might depend on the config itself.
		// The logger package's GetLogger should handle initialization safely.
		critLog := logger.GetLogger("fatal", "")
		critLog.Fatal("Failed to load application configuration", "error", err)
	}

	// Initialize logger based on app config or defaults
	var lg *logf.Logger
	if appCfg.Setup.Logging.Level != "" {
		// config.Logging does not have FilePaths, so we pass an empty string for file-based logging for now.
		// If file logging for benchmarks is needed, config.yaml and config.go would need a dedicated field.
		lg = logger.GetLogger(appCfg.Setup.Logging.Level, "")
		lg.Info("Logger initialized from app config for benchmark tool", "level", appCfg.Setup.Logging.Level, "filePath", "<stdout>")
	} else {
		// Fallback to a default logger if not specified in config
		lg = logger.GetLogger("info", "")
		lg.Info("Logger initialized with default settings for benchmark tool")
	}

	lg.Info("Starting Launchrail Benchmark Tool", "version", appCfg.Setup.App.Version)

	benchdataPath := "./benchdata"
	lg.Info("Using benchmark data from", "path", benchdataPath)

	// Create Benchmark Server
	server := NewBenchmarkServer(lg, appCfg, benchdataPath)

	// Instantiate and add benchmark suites
	orL1Bench := NewOpenRocketL1Benchmark(lg)
	server.RegisterSuite(orL1Bench)

	if len(server.benchmarkSuites) == 0 {
		lg.Warn("No benchmark suites registered. Exiting.")
		os.Exit(0)
		return
	}

	results := server.RunAll()

	// Print results
	var overallSuccess = true
	for i := range results { // Iterate by index to pass pointer to printDetailedBenchmarkResult
		server.printDetailedBenchmarkResult(&results[i]) // Print detailed results first
		if !results[i].Passed {
			overallSuccess = false
		}
	}
	server.printTable(results) // Then print the summary table

	if overallSuccess {
		lg.Info("All benchmarks PASSED")
		os.Exit(0)
	} else {
		lg.Error("One or more benchmarks FAILED")
		os.Exit(1)
	}
}
