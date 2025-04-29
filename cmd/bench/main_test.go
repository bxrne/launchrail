package main

import (
	"strings"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/stretchr/testify/assert"
)

func createTestConfig() *config.Config {
	return &config.Config{
		Setup: config.Setup{
			App: config.App{
				Name:    "TestApp",
				Version: "1.0",
				BaseDir: ".", // Assuming current dir is fine for test
			},
			Logging: config.Logging{Level: "debug"},
			Plugins: config.Plugins{Paths: []string{"./test-plugins"}},
		},
	}
}

func TestFormatResultsToMarkdown(t *testing.T) {
	results := map[string][]BenchmarkResult{
		"BenchmarkA": {
			{
				Name:          "Some Internal Name 1",
				Description:   "Test metric one",
				Metric:        "Metric1",
				Expected:      100.0,
				Actual:        101.5,
				Difference:    1.5,
				Tolerance:     2.0,
				ToleranceType: "Absolute",
				Passed:        true,
			},
			{
				Name:          "Some Internal Name 2",
				Description:   "Test metric two failed",
				Metric:        "Metric2|Pipe",
				Expected:      50.0,
				Actual:        55.1,
				Difference:    5.1,
				Tolerance:     5.0,
				ToleranceType: "Absolute",
				Passed:        false,
			},
		},
		"BenchmarkB": {
			{
				Name:          "Some Internal Name 3",
				Description:   "Test metric three",
				Metric:        "Metric3",
				Expected:      0.0,
				Actual:        0.01,
				Difference:    0.01,
				Tolerance:     0.1,
				ToleranceType: "Absolute",
				Passed:        true,
			},
		},
	}

	commitHash := "testhash"
	testCfg := createTestConfig()

	markdown := formatResultsToMarkdown(results, commitHash, testCfg)

	expectedHeader := "# Benchmark Results"
	expectedBenchmarkAHeader := "## BenchmarkA"
	expectedBenchmarkBHeader := "## BenchmarkB"
	expectedTableHeader := "| Metric        | Description   | Expected | Actual   | Diff     | Tolerance | Type     | Status |"
	expectedSeparator := "|---------------|---------------|----------|----------|----------|-----------|----------|--------|"
	expectedRowA1 := "| Metric1 | Test metric one | 100.000 | 101.500 | 1.500 | 2.000 | Absolute | :white_check_mark: PASSED |"
	expectedRowA2 := "| Metric2\\|Pipe | Test metric two failed | 50.000 | 55.100 | 5.100 | 5.000 | Absolute | :x: FAILED |"
	expectedRowB1 := "| Metric3 | Test metric three | 0.000 | 0.010 | 0.010 | 0.100 | Absolute | :white_check_mark: PASSED |"

	assert.True(t, strings.Contains(markdown, expectedHeader), "Markdown should contain the main header")
	assert.True(t, strings.Contains(markdown, expectedBenchmarkAHeader), "Markdown should contain header for BenchmarkA")
	assert.True(t, strings.Contains(markdown, expectedBenchmarkBHeader), "Markdown should contain header for BenchmarkB")
	assert.True(t, strings.Contains(markdown, expectedTableHeader), "Markdown should contain the table header")
	assert.True(t, strings.Contains(markdown, expectedSeparator), "Markdown should contain the table separator")
	assert.True(t, strings.Contains(markdown, expectedRowA1), "Markdown should contain the correct row for BenchmarkA Metric1")
	assert.True(t, strings.Contains(markdown, expectedRowA2), "Markdown should contain the correct row for BenchmarkA Metric2 (with escaped pipe)")
	assert.True(t, strings.Contains(markdown, expectedRowB1), "Markdown should contain the correct row for BenchmarkB Metric3")

	// Check order roughly
	idxA := strings.Index(markdown, expectedBenchmarkAHeader)
	idxB := strings.Index(markdown, expectedBenchmarkBHeader)
	idxA1 := strings.Index(markdown, expectedRowA1)
	idxA2 := strings.Index(markdown, expectedRowA2)
	idxB1 := strings.Index(markdown, expectedRowB1)

	assert.True(t, idxA < idxB || idxB < idxA, "Benchmark headers should exist") // Order doesn't strictly matter between benchmarks
	assert.True(t, idxA < idxA1 && idxA1 < idxA2, "Rows within BenchmarkA should be in order")
	assert.True(t, idxB < idxB1, "Row within BenchmarkB should follow its header")

	// Additional checks
	assert.Contains(t, markdown, "**Commit:** testhash")
	assert.Contains(t, markdown, "**Plugins:** `./test-plugins`")
	assert.Contains(t, markdown, "## Table of Contents") // Check TOC exists
	assert.Contains(t, markdown, "- [BenchmarkA](#benchmarka)") // Corrected TOC link assertion
	assert.Contains(t, markdown, "- [BenchmarkB](#benchmarkb)")   // Corrected TOC link assertion
}
