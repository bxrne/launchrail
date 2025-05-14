package reporting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"

	"github.com/bxrne/launchrail/internal/storage"
)

func TestParseMotionDataForPlotting(t *testing.T) {
	logger := logf.New(logf.Opts{})
	loggerPtr := &logger

	// Test case 1: Normal case with valid motion data
	simData := &storage.SimulationData{
		MotionHeaders: []string{"Time (s)", "Altitude (m)", "Velocity (m/s)"},
		MotionData: [][]string{
			{"0", "0", "0"},
			{"1", "10.5", "20.5"},
			{"2", "35.7", "30.2"},
			{"3", "70.3", "25.6"},
		},
	}

	records, headers := ParseMotionDataForPlotting(simData, loggerPtr)

	assert.NotNil(t, records, "Should return non-nil records")
	assert.Equal(t, 4, len(records), "Should parse all 4 data rows")
	assert.Equal(t, 3, len(headers), "Should return all 3 headers")

	// Verify first record values are parsed as float64
	assert.Equal(t, 0.0, (*records[0])["Time (s)"], "Should parse time as float64")
	assert.Equal(t, 0.0, (*records[0])["Altitude (m)"], "Should parse altitude as float64")
	assert.Equal(t, 0.0, (*records[0])["Velocity (m/s)"], "Should parse velocity as float64")

	// Verify last record values
	assert.Equal(t, 3.0, (*records[3])["Time (s)"], "Should parse time as float64")
	assert.Equal(t, 70.3, (*records[3])["Altitude (m)"], "Should parse altitude as float64")
	assert.Equal(t, 25.6, (*records[3])["Velocity (m/s)"], "Should parse velocity as float64")

	// Test case 2: Empty data
	emptyData := &storage.SimulationData{
		MotionHeaders: []string{"Time (s)", "Altitude (m)", "Velocity (m/s)"},
		MotionData:    [][]string{},
	}

	emptyRecords, emptyHeaders := ParseMotionDataForPlotting(emptyData, loggerPtr)
	assert.Equal(t, 0, len(emptyRecords), "Should return empty records for empty data")
	assert.Equal(t, 3, len(emptyHeaders), "Should still return headers even if no data")

	// Test case 3: Nil data
	nilData := &storage.SimulationData{
		MotionHeaders: nil,
		MotionData:    nil,
	}

	nilRecords, nilHeaders := ParseMotionDataForPlotting(nilData, loggerPtr)
	assert.Nil(t, nilRecords, "Should return nil records for nil data")
	assert.Nil(t, nilHeaders, "Should return nil headers for nil data")

	// Test case 4: Non-numeric data
	mixedData := &storage.SimulationData{
		MotionHeaders: []string{"Time (s)", "Event", "Value"},
		MotionData: [][]string{
			{"0", "Launch", "10"},
			{"1.5", "Burnout", "abc"},
			{"2.7", "Apogee", ""},
		},
	}

	mixedRecords, _ := ParseMotionDataForPlotting(mixedData, loggerPtr)
	assert.Equal(t, 3, len(mixedRecords), "Should parse all rows with mixed data")

	// Check numeric conversion worked for numbers and strings were preserved
	assert.Equal(t, 0.0, (*mixedRecords[0])["Time (s)"], "Should parse time as float64")
	assert.Equal(t, "Launch", (*mixedRecords[0])["Event"], "Should keep strings as strings")
	assert.Equal(t, 10.0, (*mixedRecords[0])["Value"], "Should parse numeric strings as float64")
	assert.Equal(t, "abc", (*mixedRecords[1])["Value"], "Should keep non-numeric strings as strings")

	// Test case 5: Row with missing columns
	shortRowData := &storage.SimulationData{
		MotionHeaders: []string{"Time (s)", "Altitude (m)", "Velocity (m/s)"},
		MotionData: [][]string{
			{"0", "0", "0"},
			{"1", "10"}, // Missing velocity column
			{"2", "35.7", "30.2"},
		},
	}

	shortRowRecords, _ := ParseMotionDataForPlotting(shortRowData, loggerPtr)
	assert.Equal(t, 2, len(shortRowRecords), "Should only parse rows with complete data")
}

func TestFindApogeeFromMotionData(t *testing.T) {
	logger := logf.New(logf.Opts{})
	loggerPtr := &logger

	// Test case 1: Normal flight with clear apogee
	headers := []string{"Time (s)", "Altitude (m)", "Velocity (m/s)"}
	motionData := []*PlotSimRecord{
		{
			"Time (s)":       0.0,
			"Altitude (m)":   0.0,
			"Velocity (m/s)": 0.0,
		},
		{
			"Time (s)":       1.0,
			"Altitude (m)":   50.0,
			"Velocity (m/s)": 40.0,
		},
		{
			"Time (s)":       2.0,
			"Altitude (m)":   120.0,
			"Velocity (m/s)": 30.0,
		},
		{
			"Time (s)":       3.0,
			"Altitude (m)":   150.0, // Apogee
			"Velocity (m/s)": 0.0,
		},
		{
			"Time (s)":       4.0,
			"Altitude (m)":   135.0,
			"Velocity (m/s)": -20.0,
		},
		{
			"Time (s)":       5.0,
			"Altitude (m)":   100.0,
			"Velocity (m/s)": -25.0,
		},
	}

	apogeeTime, apogeeAlt := findApogeeFromMotionData(motionData, headers, loggerPtr)
	assert.Equal(t, 3.0, apogeeTime, "Should find the correct apogee time")
	assert.Equal(t, 150.0, apogeeAlt, "Should find the correct apogee altitude")

	// Test case 2: Empty data
	emptyTime, emptyAlt := findApogeeFromMotionData([]*PlotSimRecord{}, headers, loggerPtr)
	assert.Equal(t, 0.0, emptyTime, "Should return zero time for empty data")
	assert.Equal(t, 0.0, emptyAlt, "Should return zero altitude for empty data")

	// Test case 3: Missing altitude column
	badHeaders := []string{"Time (s)", "BadCol", "Velocity (m/s)"}
	badHeaderTime, badHeaderAlt := findApogeeFromMotionData(motionData, badHeaders, loggerPtr)
	assert.Equal(t, 0.0, badHeaderTime, "Should return zero time when altitude column is missing")
	assert.Equal(t, 0.0, badHeaderAlt, "Should return zero altitude when altitude column is missing")

	// Test case 4: Different column naming
	altHeaders := []string{"Time", "Height", "Speed"}
	altTime, altAlt := findApogeeFromMotionData(motionData, altHeaders, loggerPtr)
	assert.Equal(t, 3.0, altTime, "Should find apogee with alternative column names")
	assert.Equal(t, 150.0, altAlt, "Should find apogee altitude with alternative column names")

	// Test case 5: Non-numeric data
	nonNumericData := []*PlotSimRecord{
		{
			"Time (s)":       0.0,
			"Altitude (m)":   "zero", // Non-numeric
			"Velocity (m/s)": 0.0,
		},
		{
			"Time (s)":       1.0,
			"Altitude (m)":   50.0,
			"Velocity (m/s)": 40.0,
		},
	}

	nonNumericTime, nonNumericAlt := findApogeeFromMotionData(nonNumericData, headers, loggerPtr)
	assert.Equal(t, 1.0, nonNumericTime, "Should find apogee with non-numeric data by ignoring those points")
	assert.Equal(t, 50.0, nonNumericAlt, "Should find apogee with non-numeric data by ignoring those points")
}

func TestCalculateDescentRates(t *testing.T) {
	logger := logf.New(logf.Opts{})
	loggerPtr := &logger

	// Test case 1: Normal descent profile
	headers := []string{"Time (s)", "Altitude (m)", "Velocity (m/s)"}
	motionData := []*PlotSimRecord{
		// Pre-apogee data
		{
			"Time (s)":       0.0,
			"Altitude (m)":   0.0,
			"Velocity (m/s)": 0.0,
		},
		{
			"Time (s)":       2.0,
			"Altitude (m)":   120.0,
			"Velocity (m/s)": 30.0,
		},
		// Apogee at t=3.0
		{
			"Time (s)":       3.0,
			"Altitude (m)":   150.0,
			"Velocity (m/s)": 0.0,
		},
		// Drogue descent (fast)
		{
			"Time (s)":       4.0,
			"Altitude (m)":   130.0,
			"Velocity (m/s)": -20.0,
		},
		{
			"Time (s)":       6.0,
			"Altitude (m)":   90.0,
			"Velocity (m/s)": -20.0,
		},
		{
			"Time (s)":       8.0,
			"Altitude (m)":   50.0,
			"Velocity (m/s)": -20.0,
		},
		// Main descent (slower)
		{
			"Time (s)":       10.0,
			"Altitude (m)":   45.0,
			"Velocity (m/s)": -5.0,
		},
		{
			"Time (s)":       12.0,
			"Altitude (m)":   35.0,
			"Velocity (m/s)": -5.0,
		},
		{
			"Time (s)":       14.0,
			"Altitude (m)":   25.0,
			"Velocity (m/s)": -5.0,
		},
		{
			"Time (s)":       18.0,
			"Altitude (m)":   5.0,
			"Velocity (m/s)": -5.0,
		},
		{
			"Time (s)":       19.0,
			"Altitude (m)":   0.0,
			"Velocity (m/s)": 0.0,
		},
	}

	drogueRate, mainRate := calculateDescentRates(motionData, headers, 3.0, loggerPtr)
	assert.InDelta(t, 20.0, drogueRate, 0.1, "Should calculate correct drogue rate")
	assert.InDelta(t, 5.0, mainRate, 0.1, "Should calculate correct main rate")

	// Test case 2: Empty data
	emptyDrogue, emptyMain := calculateDescentRates([]*PlotSimRecord{}, headers, 0.0, loggerPtr)
	assert.Equal(t, 20.0, emptyDrogue, "Should return default drogue rate for empty data")
	assert.Equal(t, 5.0, emptyMain, "Should return default main rate for empty data")

	// Test case 3: Missing columns
	badHeaders := []string{"Time (s)", "Height", "BadCol"}
	badHeaderDrogue, badHeaderMain := calculateDescentRates(motionData, badHeaders, 3.0, loggerPtr)
	assert.Equal(t, 20.0, badHeaderDrogue, "Should return default drogue rate when velocity column is missing")
	assert.Equal(t, 5.0, badHeaderMain, "Should return default main rate when velocity column is missing")

	// Test case 4: Out of range values (sanity check)
	badValueData := []*PlotSimRecord{
		// Pre-apogee
		{
			"Time (s)":       0.0,
			"Altitude (m)":   0.0,
			"Velocity (m/s)": 0.0,
		},
		// Apogee
		{
			"Time (s)":       3.0,
			"Altitude (m)":   150.0,
			"Velocity (m/s)": 0.0,
		},
		// Unrealistic drogue rate
		{
			"Time (s)":       4.0,
			"Altitude (m)":   149.0,
			"Velocity (m/s)": -0.5, // Too slow for drogue
		},
		{
			"Time (s)":       5.0,
			"Altitude (m)":   148.0,
			"Velocity (m/s)": -1.0, // Too slow for drogue
		},
		// Unrealistic main rate
		{
			"Time (s)":       10.0,
			"Altitude (m)":   200.0,
			"Velocity (m/s)": -200.0, // Too fast for main
		},
	}

	badDrogue, badMain := calculateDescentRates(badValueData, headers, 3.0, loggerPtr)
	assert.Equal(t, 20.0, badDrogue, "Should use default drogue rate when calculated rate is unrealistic")
	assert.Equal(t, 5.0, badMain, "Should use default main rate when calculated rate is unrealistic")

	// Test case 5: Alternative column names
	altHeaders := []string{"Time", "Height", "Speed"}
	altDrogue, altMain := calculateDescentRates(motionData, altHeaders, 3.0, loggerPtr)
	assert.InDelta(t, 20.0, altDrogue, 0.1, "Should calculate rates with alternative column names")
	assert.InDelta(t, 5.0, altMain, 0.1, "Should calculate rates with alternative column names")
}
