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

	assert.Contains(t, markdown, "# Benchmark Results")
	assert.Contains(t, markdown, "Cool Rocket Test")
	assert.Contains(t, markdown, "(hipr-euroc24)")
	assert.Contains(t, markdown, "Another Test")
	assert.Contains(t, markdown, "(simple-test)")
	assert.Contains(t, markdown, "| Apogee | :white_check_mark: PASS |")
	assert.Contains(t, markdown, "| Max Velocity | :x: FAIL |")
	assert.Contains(t, markdown, "| **Overall** | **:x: FAIL** |")
	assert.Contains(t, markdown, "**Plugins:** `./test-plugins`")
	assert.Contains(t, markdown, "## Table of Contents")
	assert.Contains(t, markdown, "- [Cool Rocket Test](#cool-rocket-test)")
	assert.Contains(t, markdown, "- [Another Test](#another-test)")
}
