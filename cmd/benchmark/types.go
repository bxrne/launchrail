package main

import (
	"fmt"
	"time"

	"github.com/bxrne/launchrail/internal/config"
)

// MetricResult holds the outcome of a single metric comparison.
type MetricResult struct {
	Name     string      `json:"name"`
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
	Passed   bool        `json:"passed"`
	Diff     string      `json:"diff,omitempty"`  // Optional: difference details
	Error    error       `json:"error,omitempty"` // Optional: error during metric evaluation
}

// BenchmarkResult holds the outcomes of all metrics for a benchmark run.
type BenchmarkResult struct {
	Name        string         `json:"name"`
	Metrics     []MetricResult `json:"metrics"`
	Passed      bool           `json:"passed"` // Overall pass/fail for the benchmark
	Duration    time.Duration  `json:"duration"`
	SetupError  error          `json:"setup_error,omitempty"`
	RunError    error          `json:"run_error,omitempty"`
	ReportNotes string         `json:"report_notes,omitempty"` // General notes or summary for the report
}

// Benchmark defines the interface for a benchmark suite.
type Benchmark interface {
	Name() string
	Run(cfg *config.Config, benchdataPath string) (*BenchmarkResult, error)
}

// String representation for MetricResult for easy printing
func (mr *MetricResult) String() string {
	status := "PASS"
	if !mr.Passed {
		status = "FAIL"
	}
	base := fmt.Sprintf("  Metric: %-30s | Status: %s", mr.Name, status)
	if !mr.Passed {
		base += fmt.Sprintf(" | Expected: %v, Actual: %v", mr.Expected, mr.Actual)
		if mr.Diff != "" {
			base += fmt.Sprintf(" | Diff: %s", mr.Diff)
		}
	}
	if mr.Error != nil {
		base += fmt.Sprintf(" | Error: %v", mr.Error)
	}
	return base
}

// String representation for BenchmarkResult for easy printing
func (br *BenchmarkResult) String() string {
	status := "PASSED"
	if !br.Passed {
		status = "FAILED"
	}
	resultStr := fmt.Sprintf("Benchmark: %s | Overall Status: %s | Duration: %s\n", br.Name, status, br.Duration)
	if br.SetupError != nil {
		resultStr += fmt.Sprintf("Setup Error: %v\n", br.SetupError)
	}
	if br.RunError != nil {
		resultStr += fmt.Sprintf("Run Error: %v\n", br.RunError)
	}
	for _, metric := range br.Metrics {
		resultStr += metric.String() + "\n"
	}
	if br.ReportNotes != "" {
		resultStr += fmt.Sprintf("Notes: %s\n", br.ReportNotes)
	}
	return resultStr
}
