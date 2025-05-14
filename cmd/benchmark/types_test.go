package main_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	main "github.com/bxrne/launchrail/cmd/benchmark"
	"github.com/stretchr/testify/assert"
)

func TestMetricResult_String(t *testing.T) {
	tests := []struct {
		name     string
		mr       main.MetricResult
		wantSubs []string // Substrings we expect to find in the output
	}{
		{
			name: "PASS scenario",
			mr: main.MetricResult{
				Name:   "TestMetricPass",
				Passed: true,
			},
			wantSubs: []string{"Metric: TestMetricPass", "Status: PASS"},
		},
		{
			name: "FAIL scenario with diff",
			mr: main.MetricResult{
				Name:     "TestMetricFailDiff",
				Expected: 10,
				Actual:   5,
				Passed:   false,
				Diff:     "-5 (50%)",
			},
			wantSubs: []string{"Metric: TestMetricFailDiff", "Status: FAIL", "Expected: 10", "Actual: 5", "Diff: -5 (50%)"},
		},
		{
			name: "FAIL scenario with error",
			mr: main.MetricResult{
				Name:     "TestMetricFailErr",
				Expected: "foo",
				Actual:   "bar",
				Passed:   false,
				Error:    errors.New("evaluation error"),
			},
			wantSubs: []string{"Metric: TestMetricFailErr", "Status: FAIL", "Expected: foo", "Actual: bar", "Error: evaluation error"},
		},
		{
			name: "FAIL scenario with diff and error",
			mr: main.MetricResult{
				Name:     "TestMetricFailDiffErr",
				Expected: true,
				Actual:   false,
				Passed:   false,
				Diff:     "value mismatch",
				Error:    errors.New("some issue"),
			},
			wantSubs: []string{"Metric: TestMetricFailDiffErr", "Status: FAIL", "Expected: true", "Actual: false", "Diff: value mismatch", "Error: some issue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mr.String()
			for _, sub := range tt.wantSubs {
				assert.True(t, strings.Contains(got, sub), "String() output did not contain expected substring '%s'. Got: %s", sub, got)
			}
		})
	}
}

func TestBenchmarkResult_String(t *testing.T) {
	dummyMetricPass := main.MetricResult{Name: "SubMetricPass", Passed: true}
	dummyMetricFail := main.MetricResult{Name: "SubMetricFail", Expected: 1, Actual: 0, Passed: false, Diff: "-1"}

	tests := []struct {
		name     string
		br       main.BenchmarkResult
		wantSubs []string
	}{
		{
			name: "PASSED benchmark",
			br: main.BenchmarkResult{
				Name:     "TestBenchOverallPass",
				Metrics:  []main.MetricResult{dummyMetricPass, {Name: "AnotherPass", Passed: true}},
				Passed:   true,
				Duration: 123 * time.Millisecond,
			},
			wantSubs: []string{"Benchmark: TestBenchOverallPass", "Overall Status: PASSED", "Duration: 123ms", dummyMetricPass.String()},
		},
		{
			name: "FAILED benchmark",
			br: main.BenchmarkResult{
				Name:     "TestBenchOverallFail",
				Metrics:  []main.MetricResult{dummyMetricPass, dummyMetricFail},
				Passed:   false,
				Duration: 456 * time.Second,
			},
			wantSubs: []string{"Benchmark: TestBenchOverallFail", "Overall Status: FAILED", "Duration: 7m36s", dummyMetricPass.String(), dummyMetricFail.String()},
		},
		{
			name: "Benchmark with SetupError",
			br: main.BenchmarkResult{
				Name:       "TestBenchSetupErr",
				Passed:     false, // Typically false if setup fails
				Duration:   10 * time.Millisecond,
				SetupError: errors.New("could not set up"),
			},
			wantSubs: []string{"Benchmark: TestBenchSetupErr", "Overall Status: FAILED", "Setup Error: could not set up"},
		},
		{
			name: "Benchmark with RunError",
			br: main.BenchmarkResult{
				Name:     "TestBenchRunErr",
				Passed:   false,
				Duration: 20 * time.Millisecond,
				RunError: errors.New("runtime panic"),
			},
			wantSubs: []string{"Benchmark: TestBenchRunErr", "Overall Status: FAILED", "Run Error: runtime panic"},
		},
		{
			name: "Benchmark with ReportNotes",
			br: main.BenchmarkResult{
				Name:        "TestBenchNotes",
				Metrics:     []main.MetricResult{dummyMetricPass},
				Passed:      true,
				Duration:    50 * time.Millisecond,
				ReportNotes: "All systems nominal.",
			},
			wantSubs: []string{"Benchmark: TestBenchNotes", "Overall Status: PASSED", "Notes: All systems nominal.", dummyMetricPass.String()},
		},
		{
			name: "Benchmark with all bells and whistles (failed)",
			br: main.BenchmarkResult{
				Name:        "TestBenchComplexFail",
				Metrics:     []main.MetricResult{dummyMetricPass, dummyMetricFail},
				Passed:      false,
				Duration:    1 * time.Second,
				SetupError:  errors.New("bad setup"),
				RunError:    errors.New("bad run"),
				ReportNotes: "Disaster occurred.",
			},
			wantSubs: []string{
				"Benchmark: TestBenchComplexFail",
				"Overall Status: FAILED",
				"Duration: 1s",
				"Setup Error: bad setup",
				"Run Error: bad run",
				dummyMetricPass.String(),
				dummyMetricFail.String(),
				"Notes: Disaster occurred.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.br.String()
			for _, sub := range tt.wantSubs {
				assert.True(t, strings.Contains(got, sub), "String() output did not contain expected substring '%s'. Got: %s", sub, got)
			}
		})
	}
}
