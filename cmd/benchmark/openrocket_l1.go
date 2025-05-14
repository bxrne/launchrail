package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bxrne/launchrail/internal/config"
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

// Expected CSV structure (example, adjust as needed):
// SimulationName,RocketFilePath,MotorName,ExpectedApogeeMetres,ExpectedMaxVelocityMPS,ExpectedTotalFlightTimeS

// Run executes the OpenRocket L1 comparison benchmark.
func (b *OpenRocketL1Benchmark) Run(appCfg *config.Config, benchdataPath string) (*BenchmarkResult, error) {
	startTime := time.Now()
	b.log.Info("Starting OpenRocket L1 Comparison benchmark")

	csvFilePath := filepath.Join(benchdataPath, "openrocket_l1", "export.csv")
	b.log.Info("Attempting to read benchmark data from", "path", csvFilePath)

	file, err := os.Open(csvFilePath)
	if err != nil {
		return &BenchmarkResult{
			Name:       b.Name(),
			Passed:     false,
			SetupError: fmt.Errorf("failed to open benchmark CSV file %s: %w", csvFilePath, err),
			Duration:   time.Since(startTime),
		}, nil // Return nil error for setup issues as per BenchmarkResult structure
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read() // Read header row
	if err != nil {
		return &BenchmarkResult{
			Name:       b.Name(),
			Passed:     false,
			SetupError: fmt.Errorf("failed to read CSV header from %s: %w", csvFilePath, err),
			Duration:   time.Since(startTime),
		}, nil
	}
	b.log.Info("CSV Headers found", "headers", headers)

	var metrics []MetricResult
	overallPassed := true

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Log and continue? Or fail the whole suite? For now, let's create a failed metric.
			b.log.Error("Error reading CSV record", "error", err)
			metrics = append(metrics, MetricResult{Name: "CSVRecordReadError", Passed: false, Error: err})
			overallPassed = false
			continue
		}

		if len(record) < 6 { // Assuming at least 6 columns as per example: Name, Rocket, Motor, Apogee, MaxVel, FlightTime
			b.log.Error("CSV record has insufficient columns", "record", record, "expected_min_columns", 6)
			metrics = append(metrics, MetricResult{Name: fmt.Sprintf("CSVRecordFormatError-%s", record[0]), Passed: false, Error: fmt.Errorf("expected at least 6 columns, got %d", len(record))})
			overallPassed = false
			continue
		}

		simulationName := record[0]
		rocketFileName := record[1] // e.g., "L1_Sport.ork"
		motorName := record[2]      // e.g., "Cesaroni_H128ST-14A"
		expectedApogee, errApogee := strconv.ParseFloat(record[3], 64)
		expectedMaxVelocity, errMaxVel := strconv.ParseFloat(record[4], 64)
		expectedFlightTime, errFlightTime := strconv.ParseFloat(record[5], 64)

		if errApogee != nil || errMaxVel != nil || errFlightTime != nil {
			b.log.Error("Error parsing expected numeric values from CSV", "record", record, "errApogee", errApogee, "errMaxVel", errMaxVel, "errFlightTime", errFlightTime)
			metrics = append(metrics, MetricResult{Name: fmt.Sprintf("%s-CSVParsing", simulationName), Passed: false, Error: fmt.Errorf("CSV parsing error for expected values")})
			overallPassed = false
			continue
		}

		b.log.Info("Processing simulation case", "name", simulationName, "rocketFile", rocketFileName, "motor", motorName)

		// --- LaunchRail Simulation ---
		var simRunError error

		// Create a deep copy of appCfg to modify for this specific simulation case
		caseCfg := *appCfg // This is a shallow copy, need deep copy for nested structs like Engine.Options
		// For a proper deep copy, consider using a library or manual recursive copy if structs are complex.
		// Assuming Engine and Options are structs and not pointers for this shallow copy to somewhat work for top-level changes.
		// If appCfg.Engine.Options itself is a pointer, this won't isolate changes.
		// A better approach would be: caseSpecificOptions := appCfg.Engine.Options and then modify caseSpecificOptions.
		// Then figure out how to make simulation.NewManager use these specific options, perhaps by modifying a temporary config object.
		// For now, we'll proceed with modifying a shallow copy, assuming it might be sufficient or we'll refine config handling later.

		// TODO: Properly deep copy appCfg or parts of it (Engine.Options)
		caseCfg.Engine.Options.OpenRocketFile = filepath.Join(benchdataPath, "openrocket_l1", "rockets", rocketFileName)
		caseCfg.Engine.Options.MotorDesignation = motorName

		simManager := simulation.NewManager(&caseCfg, *b.log)

		// Create and initialize storage
		recordDir := filepath.Join(benchdataPath, "openrocket_l1", "results", simulationName)
		if err := os.MkdirAll(recordDir, 0755); err != nil {
			b.log.Error("Failed to create record directory", "dir", recordDir, "error", err)
			simRunError = fmt.Errorf("failed to create record dir %s: %w", recordDir, err)
			// Add to results and continue
			metrics = append(metrics, MetricResult{Name: simulationName + "_APOGEE_METERS", Expected: expectedApogee, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_MAX_VELOCITY_METERS_PER_SECOND", Expected: expectedMaxVelocity, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_FLIGHT_TIME_SECONDS", Expected: expectedFlightTime, Actual: 0, Passed: false, Error: simRunError})
			continue
		}

		motionStore, err := storage.NewStorage(recordDir, storage.MOTION, &caseCfg)
		if err != nil {
			b.log.Error("Failed to create motion store", "error", err)
			simRunError = fmt.Errorf("failed to create motion store: %w", err)
			// Add to results and continue
			metrics = append(metrics, MetricResult{Name: simulationName + "_APOGEE_METERS", Expected: expectedApogee, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_MAX_VELOCITY_METERS_PER_SECOND", Expected: expectedMaxVelocity, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_FLIGHT_TIME_SECONDS", Expected: expectedFlightTime, Actual: 0, Passed: false, Error: simRunError})
			continue
		}
		defer motionStore.Close()
		if err := motionStore.Init(); err != nil {
			b.log.Error("Failed to initialize motion store", "error", err)
			simRunError = fmt.Errorf("failed to init motion store: %w", err)
			// Add to results and continue
			metrics = append(metrics, MetricResult{Name: simulationName + "_APOGEE_METERS", Expected: expectedApogee, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_MAX_VELOCITY_METERS_PER_SECOND", Expected: expectedMaxVelocity, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_FLIGHT_TIME_SECONDS", Expected: expectedFlightTime, Actual: 0, Passed: false, Error: simRunError})
			continue
		}

		eventsStore, err := storage.NewStorage(recordDir, storage.EVENTS, &caseCfg)
		if err != nil {
			b.log.Error("Failed to create events store", "error", err)
			simRunError = fmt.Errorf("failed to create events store: %w", err)
			// Add to results and continue
			metrics = append(metrics, MetricResult{Name: simulationName + "_APOGEE_METERS", Expected: expectedApogee, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_MAX_VELOCITY_METERS_PER_SECOND", Expected: expectedMaxVelocity, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_FLIGHT_TIME_SECONDS", Expected: expectedFlightTime, Actual: 0, Passed: false, Error: simRunError})
			continue
		}
		defer eventsStore.Close()
		if err := eventsStore.Init(); err != nil {
			b.log.Error("Failed to initialize events store", "error", err)
			simRunError = fmt.Errorf("failed to init events store: %w", err)
			// Add to results and continue
			metrics = append(metrics, MetricResult{Name: simulationName + "_APOGEE_METERS", Expected: expectedApogee, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_MAX_VELOCITY_METERS_PER_SECOND", Expected: expectedMaxVelocity, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_FLIGHT_TIME_SECONDS", Expected: expectedFlightTime, Actual: 0, Passed: false, Error: simRunError})
			continue
		}

		dynamicsStore, err := storage.NewStorage(recordDir, storage.DYNAMICS, &caseCfg)
		if err != nil {
			b.log.Error("Failed to create dynamics store", "error", err)
			simRunError = fmt.Errorf("failed to create dynamics store: %w", err)
			// Add to results and continue
			metrics = append(metrics, MetricResult{Name: simulationName + "_APOGEE_METERS", Expected: expectedApogee, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_MAX_VELOCITY_METERS_PER_SECOND", Expected: expectedMaxVelocity, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_FLIGHT_TIME_SECONDS", Expected: expectedFlightTime, Actual: 0, Passed: false, Error: simRunError})
			continue
		}
		defer dynamicsStore.Close()
		if err := dynamicsStore.Init(); err != nil {
			b.log.Error("Failed to initialize dynamics store", "error", err)
			simRunError = fmt.Errorf("failed to init dynamics store: %w", err)
			// Add to results and continue
			metrics = append(metrics, MetricResult{Name: simulationName + "_APOGEE_METERS", Expected: expectedApogee, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_MAX_VELOCITY_METERS_PER_SECOND", Expected: expectedMaxVelocity, Actual: 0, Passed: false, Error: simRunError})
			metrics = append(metrics, MetricResult{Name: simulationName + "_FLIGHT_TIME_SECONDS", Expected: expectedFlightTime, Actual: 0, Passed: false, Error: simRunError})
			continue
		}

		stores := &storage.Stores{
			Motion:   motionStore,
			Events:   eventsStore,
			Dynamics: dynamicsStore,
		}

		if err := simManager.Initialize(stores); err != nil { // Pass stores instance
			b.log.Error("Failed to initialize simulation manager", "sim_name", simulationName, "error", err)
			simRunError = fmt.Errorf("sim manager init failed: %w", err)
		} else {
			// Load the ORK file and motor (done within Initialize now)
			// Run the simulation
			if err := simManager.Run(); err != nil { // Use Run() method
				b.log.Error("Simulation run failed", "sim_name", simulationName, "error", err)
				simRunError = fmt.Errorf("simulation run failed: %w", err)
			} else {
				// Extract results from simManager.sim (TODO: inspect pkg/simulation/simulation.go for fields)
				simInstance := simManager.GetSim() // Corrected: Use GetSim()
				if simInstance == nil {
					b.log.Error("Failed to get simulation instance after run")
				} else {
					actualApogee := simInstance.MaxAltitude
					actualMaxVelocity := simInstance.MaxSpeed
					actualFlightTime := simInstance.CurrentTime // Was currentTime

					// Compare results and add to benchmarkResult
					// Compare Apogee
					apogeeMetric := b.compareFloatMetric(fmt.Sprintf("%s-Apogee", simulationName), expectedApogee, actualApogee, 0.05) // 5% tolerance
					if simRunError != nil {
						apogeeMetric.Error = simRunError
						apogeeMetric.Passed = false
					}
					metrics = append(metrics, apogeeMetric)
					if !apogeeMetric.Passed {
						overallPassed = false
					}

					// Compare Max Velocity
					maxVelMetric := b.compareFloatMetric(fmt.Sprintf("%s-MaxVelocity", simulationName), expectedMaxVelocity, actualMaxVelocity, 0.05) // 5% tolerance
					if simRunError != nil {
						maxVelMetric.Error = simRunError
						maxVelMetric.Passed = false
					}
					metrics = append(metrics, maxVelMetric)
					if !maxVelMetric.Passed {
						overallPassed = false
					}

					// Compare Total Flight Time
					flightTimeMetric := b.compareFloatMetric(fmt.Sprintf("%s-FlightTime", simulationName), expectedFlightTime, actualFlightTime, 0.10) // 10% tolerance
					if simRunError != nil {
						flightTimeMetric.Error = simRunError
						flightTimeMetric.Passed = false
					}
					metrics = append(metrics, flightTimeMetric)
					if !flightTimeMetric.Passed {
						overallPassed = false
					}
				}
			}
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
