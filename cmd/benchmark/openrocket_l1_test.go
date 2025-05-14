package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/bxrne/launchrail/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	logf "github.com/zerodha/logf"
)

func newTestLogger() *logf.Logger {
	opts := logger.GetDefaultOpts() // This should provide reasonable defaults
	opts.Level = logf.ErrorLevel    // Set a specific quiet level, e.g., ErrorLevel
	opts.Writer = io.Discard        // Discard output for quiet tests
	lg := logf.New(opts)
	return &lg // Return address of the logger instance
}

func TestOpenRocketL1Benchmark_Name(t *testing.T) {
	lg := newTestLogger()
	bench := NewOpenRocketL1Benchmark(lg)
	assert.Equal(t, "OpenRocketL1Comparison", bench.Name())
}

func TestOpenRocketL1Benchmark_compareFloatMetric(t *testing.T) {
	lg := newTestLogger()
	bench := NewOpenRocketL1Benchmark(lg)

	tests := []struct {
		name             string
		expectedVal      float64
		actualVal        float64
		tolerancePercent float64
		wantPassed       bool
		wantDiffSubstr   string // Substring to check in the MetricResult.Diff field
	}{
		{
			name:             "Actual within positive tolerance",
			expectedVal:      100.0,
			actualVal:        103.0,
			tolerancePercent: 5.0,
			wantPassed:       true,
			wantDiffSubstr:   "3.00 (3.00%)",
		},
		{
			name:             "Actual within negative tolerance",
			expectedVal:      100.0,
			actualVal:        97.0,
			tolerancePercent: 5.0,
			wantPassed:       true,
			wantDiffSubstr:   "-3.00 (-3.00%)",
		},
		{
			name:             "Actual outside positive tolerance",
			expectedVal:      100.0,
			actualVal:        106.0,
			tolerancePercent: 5.0,
			wantPassed:       false,
			wantDiffSubstr:   "6.00 (6.00%)",
		},
		{
			name:             "Actual outside negative tolerance",
			expectedVal:      100.0,
			actualVal:        94.0,
			tolerancePercent: 5.0,
			wantPassed:       false,
			wantDiffSubstr:   "-6.00 (-6.00%)",
		},
		{
			name:             "Actual exactly at positive tolerance boundary",
			expectedVal:      100.0,
			actualVal:        105.0,
			tolerancePercent: 5.0,
			wantPassed:       true,
			wantDiffSubstr:   "5.00 (5.00%)",
		},
		{
			name:             "Actual exactly at negative tolerance boundary",
			expectedVal:      100.0,
			actualVal:        95.0,
			tolerancePercent: 5.0,
			wantPassed:       true,
			wantDiffSubstr:   "-5.00 (-5.00%)",
		},
		{
			name:             "Expected is zero, actual within absolute tolerance",
			expectedVal:      0.0,
			actualVal:        0.04,
			tolerancePercent: 0.05, // Interpreted as absolute 0.05
			wantPassed:       true,
			wantDiffSubstr:   "0.04 (+Inf%)",
		},
		{
			name:             "Expected is zero, actual outside absolute tolerance",
			expectedVal:      0.0,
			actualVal:        0.06,
			tolerancePercent: 0.05, // Interpreted as absolute 0.05
			wantPassed:       false,
			wantDiffSubstr:   "0.06 (+Inf%)",
		},
		{
			name:             "Both zero",
			expectedVal:      0.0,
			actualVal:        0.0,
			tolerancePercent: 1.0,
			wantPassed:       true,
			wantDiffSubstr:   "0.00 (NaN%)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricName := "TestMetric"
			result := bench.compareFloatMetric(metricName, tt.expectedVal, tt.actualVal, tt.tolerancePercent)

			assert.Equal(t, metricName, result.Name)
			assert.Equal(t, tt.expectedVal, result.Expected)
			assert.Equal(t, tt.actualVal, result.Actual)
			assert.Equal(t, tt.wantPassed, result.Passed)
			assert.Contains(t, result.Diff, tt.wantDiffSubstr)
		})
	}
}

// Helper function to create a temporary CSV file for testing
func createTempCSV(t *testing.T, content string) string {
	t.Helper()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_export.csv")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to write temp CSV file")
	return filePath
}

func TestOpenRocketL1Benchmark_loadOpenRocketExportData(t *testing.T) {
	lg := newTestLogger()

	tests := []struct {
		name          string
		csvContent    string
		wantApogee    float64
		wantMaxVel    float64
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name: "Valid CSV full data",
			csvContent: `# Exported from OpenRocket
# Version: OpenRocket 15.03
# Exported on 2024-07-15 12:00:00
#
# Simulation #1: Apogee, Max Velocity, etc.
Time (s),Altitude (m),Vertical velocity (m/s),Vertical acceleration (m/s²),Total velocity (m/s),Total acceleration (m/s²),Position East (m),Position North (m),Lateral distance (m),Lateral direction (°),Latitude (deg),Longitude (deg),Gravitational acceleration (m/s²),Angle of attack (°),Roll rate (°/s),Pitch rate (°/s),Yaw rate (°/s),Mass (g),Propellant mass (g),Longitudinal moment of inertia (kg·m²),Rotational moment of inertia (kg·m²),CP location (cm),CG location (cm),Stability margin calibers (난류),Mach number (난류),Reynolds number (난류),Thrust (N),Drag coefficient (난류),Axial drag coefficient (난류),Pressure (Pa),Temperature (°C),Speed of sound (m/s),Air density (kg/m³),Dynamic viscosity (Pa·s)
0.0,0.0,0.0,0.0,0.0,0.0,0.0,0.0,0.0,0.0,0.0,0.0,9.81,0.0,0.0,0.0,0.0,1000.0,100.0,0.1,0.01,50.0,40.0,1.5,0.0,0.0,0.0,0.0,0.0,101325.0,15.0,340.0,1.225,0.0000181
1.0,50.0,100.0,10.0,100.0,10.0,0.0,0.0,0.0,0.0,0.0,0.0,9.81,0.0,0.0,0.0,0.0,950.0,50.0,0.1,0.01,50.0,40.0,1.5,0.3,100000.0,200.0,0.3,0.3,100000.0,10.0,330.0,1.2,0.0000180
2.0,150.0,80.0,-20.0,80.0,-20.0,0.0,0.0,0.0,0.0,0.0,0.0,9.81,0.0,0.0,0.0,0.0,900.0,0.0,0.1,0.01,50.0,40.0,1.5,0.2,80000.0,0.0,0.4,0.4,98000.0,5.0,320.0,1.1,0.0000179
3.0,180.0,0.0,-10.0,10.0,-10.0,0.0,0.0,0.0,0.0,0.0,0.0,9.81,0.0,0.0,0.0,0.0,900.0,0.0,0.1,0.01,50.0,40.0,1.5,0.03,10000.0,0.0,0.5,0.5,97000.0,3.0,310.0,1.0,0.0000178
4.0,150.0,-50.0,-10.0,50.0,-10.0,0.0,0.0,0.0,0.0,0.0,0.0,9.81,0.0,0.0,0.0,0.0,900.0,0.0,0.1,0.01,50.0,40.0,1.5,0.15,50000.0,0.0,0.6,0.6,96000.0,1.0,300.0,0.9,0.0000177
`,
			wantApogee: 180.0,
			wantMaxVel: 100.0,
			wantErr:    false,
		},
		{
			name: "Valid CSV minimal data for apogee and max_vel",
			csvContent: `# Simulation #1: Apogee, Max Velocity, etc.
Time (s),Altitude (m),Vertical velocity (m/s)
0.0,0.0,0.0
1.0,50.0,100.0
2.0,150.0,80.0
3.0,180.0,0.0
4.0,150.0,-50.0
`,
			wantApogee: 180.0,
			wantMaxVel: 100.0,
			wantErr:    false,
		},
		{
			name: "CSV no data rows",
			csvContent: `# Simulation #1: Apogee, Max Velocity, etc.
Time (s),Altitude (m),Vertical velocity (m/s)
`,
			wantErr:       true,
			wantErrSubstr: "no data rows found",
		},
		{
			name: "CSV malformed header line ID (no # Simulation #n)",
			csvContent: `# Some other comment
Time (s),Altitude (m),Vertical velocity (m/s)
0.0,0.0,0.0
`,
			wantErr:       true,
			wantErrSubstr: "could not find OpenRocket data header line",
		},
		{
			name: "CSV missing Altitude column",
			csvContent: `# Simulation #1: Apogee, Max Velocity, etc.
Time (s),Vertical velocity (m/s)
0.0,0.0
`,
			wantErr:       true,
			wantErrSubstr: "could not find required column 'Altitude (m)'",
		},
		{
			name: "CSV missing Vertical velocity column",
			csvContent: `# Simulation #1: Apogee, Max Velocity, etc.
Time (s),Altitude (m)
0.0,0.0
`,
			wantErr:       true,
			wantErrSubstr: "could not find required column 'Vertical velocity (m/s)'",
		},
		{
			name: "CSV non-numeric altitude",
			csvContent: `# Simulation #1: Apogee, Max Velocity, etc.
Time (s),Altitude (m),Vertical velocity (m/s)
0.0,abc,0.0
`,
			wantErr:       true,
			wantErrSubstr: "error parsing altitude",
		},
		{
			name: "CSV non-numeric vertical velocity",
			csvContent: `# Simulation #1: Apogee, Max Velocity, etc.
Time (s),Altitude (m),Vertical velocity (m/s)
0.0,0.0,xyz
`,
			wantErr:       true,
			wantErrSubstr: "error parsing vertical_velocity",
		},
		{
			name:          "Empty CSV file",
			csvContent:    ``,
			wantErr:       true,
			wantErrSubstr: "could not find OpenRocket data header line", // or "no records found" or similar
		},
		{
			name:          "Non-existent file",
			csvContent:    "", // Special case, filePath will be set to a non-existent path
			wantErr:       true,
			wantErrSubstr: "no such file or directory", // OS-dependent message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.name == "Non-existent file" {
				filePath = filepath.Join(t.TempDir(), "non_existent_file.csv")
			} else {
				filePath = createTempCSV(t, tt.csvContent)
			}

			apogee, maxVel, _, err := loadOpenRocketExportData(filePath, lg)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrSubstr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantApogee, apogee, "Apogee mismatch")
				assert.Equal(t, tt.wantMaxVel, maxVel, "MaxVelocity mismatch")
			}
		})
	}
}

// Test for Run() method will be next.
