package main

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestFormatResultsMarkdown(t *testing.T) {
	mockResults := map[string][]BenchmarkResult{
		"Cool Rocket Test": {
			{
				Name:          "Cool Rocket Test",
				Description:   "Test 1 Description",
				Metric:        "Apogee",
				Expected:      1000.0,
				Actual:        995.0,
				Difference:    -5.0,
				Tolerance:     10.0,
				ToleranceType: "absolute",
				Passed:        true,
			},
			{
				Name:          "Cool Rocket Test",
				Description:   "Test 2 Description",
				Metric:        "Max Velocity",
				Expected:      300.0,
				Actual:        310.0,
				Difference:    10.0,
				Tolerance:     0.05,
				ToleranceType: "relative",
				Passed:        false,
			},
		},
		"Another Test": {
			{
				Name:          "Another Test",
				Description:   "Single test",
				Metric:        "Burn Time",
				Expected:      5.0,
				Actual:        5.0,
				Difference:    0.0,
				Tolerance:     0.1,
				ToleranceType: "absolute",
				Passed:        true,
			},
		},
	}

	mockBenchmarkEntries := map[string]config.BenchmarkEntry{
		"hipr-euroc24": {
			Name:       "Cool Rocket Test",
			DesignFile: "/path/to/design1.ork",
			DataDir:    "/path/to/data1",
			Enabled:    true,
		},
		"simple-test": {
			Name:       "Another Test",
			DesignFile: "/path/to/design2.ork",
			DataDir:    "/path/to/data2",
			Enabled:    true,
		},
	}

	markdown := formatResultsMarkdown(mockResults, mockBenchmarkEntries)

	// General Structure Assertions
	assert.Contains(t, markdown, "# Benchmark Results")
	assert.Contains(t, markdown, "### Summary")
	assert.Contains(t, markdown, "| Name | Status | Passed | Failed |") // Summary header
	assert.Contains(t, markdown, "### Details")

	// Summary Content Assertions
	assert.Contains(t, markdown, "| Cool Rocket Test | FAIL | 1 | 1 |")
	assert.Contains(t, markdown, "| Another Test | PASS | 1 | 0 |")
	assert.Contains(t, markdown, "| **Overall** | **FAIL** | **2** | **1** |") // Corrected assertion with bold counts

	// Details Assertions
	assert.Contains(t, markdown, "#### Cool Rocket Test (hipr-euroc24)")
	assert.Contains(t, markdown, "| Apogee | PASS |")          // Plain PASS status
	assert.Contains(t, markdown, "| Max Velocity | FAIL |")    // Plain FAIL status

	assert.Contains(t, markdown, "#### Another Test (simple-test)")
	assert.Contains(t, markdown, "| Burn Time | PASS |")       // Plain PASS status

	// Ensure removed/never added elements are NOT present
	assert.NotContains(t, markdown, ":white_check_mark:")
	assert.NotContains(t, markdown, ":x:")
	assert.NotContains(t, markdown, "## Table of Contents")
	assert.NotContains(t, markdown, "- [Cool Rocket Test](#cool-rocket-test)")
	assert.NotContains(t, markdown, "- [Another Test](#another-test)")
	assert.NotContains(t, markdown, "**Plugins:**") // Plugin path is not part of this function's output
}
