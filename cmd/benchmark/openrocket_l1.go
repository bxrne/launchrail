package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
	logf "github.com/zerodha/logf"
)

// OpenRocketL1Benchmark implements the Benchmark interface for OpenRocket L1 comparison.
type OpenRocketL1Benchmark struct {
	log *logf.Logger
}

// NewOpenRocketL1Benchmark creates a new instance of OpenRocketL1Benchmark.
func NewOpenRocketL1Benchmark(lg *logf.Logger) *OpenRocketL1Benchmark {
	return &OpenRocketL1Benchmark{
		log: lg,
	}
}

// Name returns the name of the benchmark suite.
func (b *OpenRocketL1Benchmark) Name() string {
	return "OpenRocketL1Comparison"
}

const (
	colTimeS        = "Time (s)"
	colAltitudeM    = "Altitude (m)"
	colTotVelocity  = "Total velocity (m/s)"
	colVertVelocity = "Vertical velocity (m/s)"
)

// loadOpenRocketExportData parses an OpenRocket simulation export CSV file to extract key flight metrics.
// It looks for maximum altitude, maximum velocity, and the flight time until apogee is reached.
func loadOpenRocketExportData(filePath string, log *logf.Logger) (apogee float64, maxVelocity float64, flightTimeToApogee float64, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Failed to open OpenRocket export file", "path", filePath, "error", err)
		return 0, 0, 0, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var headerLine string
	var lineNumBeforeData int
	var foundOpenRocketSimLine bool

	// First, scan for OpenRocket simulation line and header line
	for scanner.Scan() {
		lineNumBeforeData++
		line := scanner.Text()

		// Check for OpenRocket simulation identifier in comments
		if strings.HasPrefix(line, "#") && strings.Contains(line, "Simulation #") {
			foundOpenRocketSimLine = true
			continue
		}

		// Special case: if a line contains column headers (including quoted versions), it's the header line
		// Handle both quoted and unquoted variants
		cleanLine := line

		// Handle quoted format - check if line starts with a quote and trim it
		if strings.HasPrefix(cleanLine, "\"") {
			cleanLine = strings.TrimPrefix(cleanLine, "\"")
		}

		// Check if this might be a header line (contains column names)
		if (strings.Contains(cleanLine, "Time (s)") && strings.Contains(cleanLine, "Altitude (m)")) ||
			(strings.Contains(cleanLine, "\"Time (s)\"") && strings.Contains(cleanLine, "\"Altitude (m)\"")) {
			// This is a header line (possibly commented, possibly quoted)

			// If it starts with #, remove the # prefix for parsing
			if strings.HasPrefix(cleanLine, "#") {
				cleanLine = strings.TrimPrefix(cleanLine, "#")
				cleanLine = strings.TrimSpace(cleanLine)
			}

			headerLine = cleanLine
			break
		}

		// Skip other comments and blank lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// First non-comment, non-blank line is the header
		headerLine = line
		break
	}

	if err := scanner.Err(); err != nil {
		log.Error("Error scanning for header in OpenRocket export file", "path", filePath, "error", err)
		return 0, 0, 0, fmt.Errorf("error scanning for header in %s: %w", filePath, err)
	}

	if headerLine == "" {
		log.Error("Could not find OpenRocket data header line in export file", "path", filePath)
		return 0, 0, 0, fmt.Errorf("could not find OpenRocket data header line in %s", filePath)
	}

	// Ensure we found the OpenRocket simulation identifier
	if !foundOpenRocketSimLine && !strings.Contains(headerLine, "Simulation #") {
		log.Error("Missing OpenRocket simulation identifier line", "path", filePath)
		return 0, 0, 0, fmt.Errorf("could not find OpenRocket data header line in %s", filePath)
	}

	// Parse the header line as CSV
	// Configure CSV reader to handle quoted fields
	csvReader := csv.NewReader(strings.NewReader(headerLine))
	csvReader.LazyQuotes = true // Allow LazyQuotes to handle quoted or unquoted CSV
	parsedHeader, err := csvReader.Read()
	if err != nil {
		log.Error("Failed to parse the extracted header line", "header_line", headerLine, "error", err)
		return 0, 0, 0, fmt.Errorf("failed to parse header line '%s': %w", headerLine, err)
	}

	timeColIdx, altColIdx, velColIdx := -1, -1, -1
	for i, h := range parsedHeader {
		switch strings.TrimSpace(h) {
		case colTimeS:
			timeColIdx = i
		case colAltitudeM:
			altColIdx = i
		case colTotVelocity, colVertVelocity:
			velColIdx = i
		}
	}

	// Check each required column and return specific error message for missing columns
	if timeColIdx == -1 {
		log.Error("Time column not found in OpenRocket export header", "path", filePath, "parsed_header", parsedHeader)
		return 0, 0, 0, fmt.Errorf("could not find required column '%s'", colTimeS)
	}

	if altColIdx == -1 {
		log.Error("Altitude column not found in OpenRocket export header", "path", filePath, "parsed_header", parsedHeader)
		return 0, 0, 0, fmt.Errorf("could not find required column '%s'", colAltitudeM)
	}

	if velColIdx == -1 {
		log.Error("Velocity column not found in OpenRocket export header", "path", filePath, "parsed_header", parsedHeader,
			"expected_columns", []string{colTotVelocity, colVertVelocity})
		return 0, 0, 0, fmt.Errorf("could not find required column '%s'", colVertVelocity)
	}
	log.Debug("Successfully parsed OpenRocket export header", "path", filePath, "timeIdx", timeColIdx, "altIdx", altColIdx, "velIdx", velColIdx)

	maxAltitude := -1.0
	currentMaxVelocity := 0.0
	apogeeTime := 0.0
	foundApogee := false
	dataLineNum := 0

	// Continue scanning for data rows
	for scanner.Scan() {
		lineNumBeforeData++ // keep track of original line number for logging
		line := scanner.Text()

		// Skip pure comment lines or empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		dataLineNum++
		csvReader := csv.NewReader(strings.NewReader(line))
		csvReader.LazyQuotes = true // Handle quoted or unquoted fields
		record, err := csvReader.Read()
		if err != nil {
			log.Warn("Failed to parse data line with CSV reader, skipping", "line_num_original", lineNumBeforeData, "line_content", line, "error", err)
			continue
		}

		// Data row processing (similar to before)
		maxRequiredIdx := timeColIdx
		if altColIdx > maxRequiredIdx {
			maxRequiredIdx = altColIdx
		}
		if velColIdx > maxRequiredIdx {
			maxRequiredIdx = velColIdx
		}
		if len(record) <= maxRequiredIdx {
			log.Warn("Skipping malformed data row, not enough columns for required fields", "line_num_original", lineNumBeforeData, "num_cols", len(record), "max_required_idx", maxRequiredIdx)
			continue
		}

		timeStr := strings.TrimSpace(record[timeColIdx])
		altStr := strings.TrimSpace(record[altColIdx])
		velStr := strings.TrimSpace(record[velColIdx])

		currentTime, errTime := strconv.ParseFloat(timeStr, 64)
		if errTime != nil {
			log.Warn("Failed to parse time value", "line_num_original", lineNumBeforeData, "time", timeStr, "error", errTime)
			return 0, 0, 0, fmt.Errorf("error parsing time value: %v", errTime)
		}

		currentAltitude, errAlt := strconv.ParseFloat(altStr, 64)
		if errAlt != nil {
			log.Warn("Failed to parse altitude value", "line_num_original", lineNumBeforeData, "altitude", altStr, "error", errAlt)
			return 0, 0, 0, fmt.Errorf("error parsing altitude: %v", errAlt)
		}

		currentVelocity, errVel := strconv.ParseFloat(velStr, 64)
		if errVel != nil {
			log.Warn("Failed to parse velocity value", "line_num_original", lineNumBeforeData, "velocity", velStr, "error", errVel)
			return 0, 0, 0, fmt.Errorf("error parsing vertical_velocity: %v", errVel)
		}

		if currentAltitude > maxAltitude {
			maxAltitude = currentAltitude
			apogeeTime = currentTime
			foundApogee = true
		}

		if currentVelocity > currentMaxVelocity {
			currentMaxVelocity = currentVelocity
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error("Error scanning data rows in OpenRocket export file", "path", filePath, "error", err)
		return 0, 0, 0, fmt.Errorf("error scanning data rows in %s: %w", filePath, err)
	}

	// Check for data rows
	if dataLineNum == 0 {
		log.Error("No data rows found in OpenRocket export file after header.", "path", filePath)
		return 0, 0, 0, fmt.Errorf("no data rows found in %s", filePath)
	}

	// Check if we found an apogee
	if !foundApogee {
		log.Warn("No apogee found in OpenRocket data", "path", filePath)
		return 0, 0, 0, fmt.Errorf("no apogee event or data found in %s after processing %d data lines", filePath, dataLineNum)
	}

	log.Info("Successfully parsed OpenRocket export data", "path", filePath, "apogee_m", maxAltitude, "max_velocity_mps", currentMaxVelocity, "time_to_apogee_s", apogeeTime)
	return maxAltitude, currentMaxVelocity, apogeeTime, nil
}

// Run executes the OpenRocket L1 comparison benchmark.
func (b *OpenRocketL1Benchmark) Run(appCfg *config.Config, benchdataPath string) (*BenchmarkResult, error) {
	startTime := time.Now()
	b.log.Info("Starting OpenRocket L1 Comparison benchmark")

	csvFilePath := filepath.Join(benchdataPath, "openrocket_l1", "export.csv")
	b.log.Info("Attempting to read OpenRocket export data from", "path", csvFilePath)

	expectedApogee, expectedMaxVelocity, expectedFlightTime, err := loadOpenRocketExportData(csvFilePath, b.log)
	if err != nil {
		return &BenchmarkResult{
			Name:       b.Name(),
			Passed:     false,
			SetupError: fmt.Errorf("failed to load OpenRocket export data from %s: %w", csvFilePath, err),
			Duration:   time.Since(startTime),
		}, nil
	}

	var metrics []MetricResult
	overallPassed := true

	// Simulation Name for reporting purposes
	simulationName := "LaunchRail_vs_OpenRocketL1"
	b.log.Info("Processing simulation case", "name", simulationName, "rocketFile_from_config", appCfg.Engine.Options.OpenRocketFile, "motor_from_config", appCfg.Engine.Options.MotorDesignation)

	// --- LaunchRail Simulation ---
	var simRunError error

	// Create a deep copy of appCfg to modify for this specific simulation case
	// No, for this simplified version, we assume appCfg is already set for the L1 rocket.
	// If rocket/motor needs to be overridden, it should be done before calling Run or via appCfg.

	// Setup simulation manager
	simLoggerOpts := logger.GetDefaultOpts() // Get base options from our logger package
	simLoggerOpts.DefaultFields = []any{"service", "simulation-manager", "benchmark_case", simulationName}
	simManagerLogger := logf.New(simLoggerOpts) // logf.New returns logf.Logger (value type)

	simManager := simulation.NewManager(appCfg, simManagerLogger) // Pass value type logger
	// TODO: Review if NewManager can intrinsically fail in a way that needs pre-check before Initialize.
	// For now, errors are caught during Initialize or Run phases.

	// Setup storage (unique for each benchmark run to avoid conflicts if run in parallel, though not currently)
	// For now, use a timestamp or a fixed sub-directory for benchmark outputs.
	// TODO: Use appCfg.Storage.RecordDir once StorageConfig is properly implemented in internal/config/config.go
	benchmarkOutputDir := filepath.Join(benchdataPath, "launchrail_outputs", simulationName+"_"+startTime.Format("20060102_150405"))

	if err := os.MkdirAll(benchmarkOutputDir, 0755); err != nil {
		b.log.Error("Failed to create benchmark output directory", "path", benchmarkOutputDir, "error", err)
		simRunError = fmt.Errorf("failed to create benchmark output dir: %w", err)
	} else {
		motionStore, errMotion := storage.NewStorage(benchmarkOutputDir, storage.MOTION, appCfg)
		eventsStore, errEvents := storage.NewStorage(benchmarkOutputDir, storage.EVENTS, appCfg)
		dynamicsStore, errDynamics := storage.NewStorage(benchmarkOutputDir, storage.DYNAMICS, appCfg)

		if errMotion != nil || errEvents != nil || errDynamics != nil {
			b.log.Error("Failed to create storage stores", "errMotion", errMotion, "errEvents", errEvents, "errDynamics", errDynamics)
			simRunError = fmt.Errorf("failed to create storage: m:%v e:%v d:%v", errMotion, errEvents, errDynamics)
		} else {
			defer motionStore.Close()
			defer eventsStore.Close()
			defer dynamicsStore.Close()

			if err := motionStore.Init(); err != nil || eventsStore.Init() != nil || dynamicsStore.Init() != nil {
				b.log.Error("Failed to initialize stores", "errMotion", motionStore.Init(), "errEvents", eventsStore.Init(), "errDynamics", dynamicsStore.Init())
				simRunError = fmt.Errorf("failed to init stores") // Simplified error
			} else {
				stores := &storage.Stores{
					Motion:   motionStore,
					Events:   eventsStore,
					Dynamics: dynamicsStore,
				}

				if err := simManager.Initialize(stores); err != nil {
					b.log.Error("Failed to initialize simulation manager", "sim_name", simulationName, "error", err)
					simRunError = fmt.Errorf("sim manager init failed: %w", err)
				} else {
					if err := simManager.Run(); err != nil {
						b.log.Error("Simulation run failed", "sim_name", simulationName, "error", err)
						simRunError = fmt.Errorf("simulation run failed: %w", err)
					} else {
						simInstance := simManager.GetSim()
						if simInstance == nil {
							b.log.Error("Failed to get simulation instance after run")
							simRunError = fmt.Errorf("GetSim() returned nil after run")
						} else {
							actualApogee := simInstance.MaxAltitude
							actualMaxVelocity := simInstance.MaxSpeed
							actualFlightTime := simInstance.CurrentTime // This is total simulation time, not necessarily apogee time.
							// For a more accurate flight time comparison, LaunchRail sim should also report apogee time.
							// Using expectedFlightTime which is APOGEE time from OpenRocket for now.

							// Compare Apogee
							apogeeMetric := b.compareFloatMetric(fmt.Sprintf("%s-Apogee_METERS", simulationName), expectedApogee, actualApogee, 0.05) // 5% tolerance
							metrics = append(metrics, apogeeMetric)
							if !apogeeMetric.Passed {
								overallPassed = false
							}

							// Compare Max Velocity
							maxVelMetric := b.compareFloatMetric(fmt.Sprintf("%s-MaxVelocity_MPS", simulationName), expectedMaxVelocity, actualMaxVelocity, 0.05) // 5% tolerance
							metrics = append(metrics, maxVelMetric)
							if !maxVelMetric.Passed {
								overallPassed = false
							}

							// Compare Total Flight Time (LaunchRail CurrentTime vs OpenRocket Apogee Time)
							flightTimeMetric := b.compareFloatMetric(fmt.Sprintf("%s-FlightTime_SECONDS", simulationName), expectedFlightTime, actualFlightTime, 0.10) // 10% tolerance
							metrics = append(metrics, flightTimeMetric)
							if !flightTimeMetric.Passed {
								overallPassed = false
							}
						}
					}
				}
			}
		}
	}

	// If there was a simulation run error, ensure metrics reflect this for clarity, even if some comparisons were made.
	if simRunError != nil {
		b.log.Error("Simulation run encountered an error, marking all metrics as failed if not already explicitly error-handled.", "error", simRunError)
		overallPassed = false
		// Ensure a generic error metric if no specific metrics were added or if they don't capture the sim error.
		if len(metrics) == 0 {
			metrics = append(metrics, MetricResult{Name: simulationName + "_SIMULATION_RUN_ERROR", Passed: false, Error: simRunError})
		} else {
			// Optionally, append error to existing metrics or add a new one.
			// For simplicity, we assume earlier metric creation for sim errors handles this, or rely on overallPassed.
		}
	}

	return &BenchmarkResult{
		Name:     b.Name(),
		Metrics:  metrics,
		Passed:   overallPassed,
		Duration: time.Since(startTime),
	}, nil
}

// compareFloatMetric is a helper to compare float values with a tolerance.
func (b *OpenRocketL1Benchmark) compareFloatMetric(name string, expected, actual, tolerancePercent float64) MetricResult {
	diff := actual - expected
	percentDiff := 0.0
	if expected != 0 {
		percentDiff = (diff / expected) * 100
	} else if diff != 0 { // expected is 0, actual is not. diff == actual.
		if diff > 0 {
			percentDiff = math.Inf(1)
		} else {
			percentDiff = math.Inf(-1)
		}
	} else { // expected is 0, actual is 0. diff is 0.
		percentDiff = math.NaN()
	}

	passed := false
	if expected == 0 { // Handle case where expected is zero, tolerance is absolute
		passed = math.Abs(diff) <= tolerancePercent // Interpret tolerancePercent as absolute tolerance if expected is 0
	} else {
		passed = math.Abs(percentDiff) <= tolerancePercent
	}

	return MetricResult{
		Name:     name,
		Expected: expected,
		Actual:   actual,
		Passed:   passed,
		Diff:     fmt.Sprintf("%.2f (%.2f%%)", diff, percentDiff),
	}
}
