package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	logf "github.com/zerodha/logf"
)

// HiprEuroc24Benchmark implements the Benchmark interface for the specific dataset.
type HiprEuroc24Benchmark struct {
	name string
	// Ground truth data fields removed - loaded within Run
	// Config field removed - passed into Run
}

// NewHiprEuroc24Benchmark creates a new instance.
func NewHiprEuroc24Benchmark() *HiprEuroc24Benchmark {
	return &HiprEuroc24Benchmark{
		name: "hipr-euroc24",
		// config removed
	}
}

// Name returns the benchmark name.
func (b *HiprEuroc24Benchmark) Name() string {
	return b.name
}

// Setup method removed.

// LoadData method removed.

// Run executes the benchmark: runs simulation, loads ground truth, compares results.
func (b *HiprEuroc24Benchmark) Run(entry config.BenchmarkEntry, logger *logf.Logger, runDir string) ([]BenchmarkResult, error) {
	logger.Info("--- Starting Benchmark Run --- ", "name", b.Name())

	// --- 1. Prepare Simulation Input ---
	logger.Info("Preparing simulation input", "motor", entry.MotorDesignation, "design_file", entry.DesignFile)

	// Get absolute path for design file relative to CWD or assume absolute
	designFilePath, err := filepath.Abs(entry.DesignFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for design file '%s': %w", entry.DesignFile, err)
	}
	if _, err := os.Stat(designFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("design file '%s' not found at resolved path '%s'", entry.DesignFile, designFilePath)
	}
	logger.Debug("Using design file", "path", designFilePath)

	// Fetch thrust curve data
	httpClient := http_client.NewHTTPClient()
	loggerForThrustcurves := logf.New(logf.Opts{Level: logf.FatalLevel})
	motorData, err := thrustcurves.Load(entry.MotorDesignation, httpClient, loggerForThrustcurves)
	if err != nil {
		return nil, fmt.Errorf("failed to load thrust curve data for designation '%s': %w", entry.MotorDesignation, err)
	}
	if motorData == nil {
		return nil, fmt.Errorf("no thrust curve data found for designation '%s'", entry.MotorDesignation)
	}
	logger.Debug("Thrust curve data loaded successfully", "motor", entry.MotorDesignation, "points", len(motorData.Thrust))

	// --- 2. Run Simulation ---
	logger.Info("Running simulation...", "output_dir", runDir)
	// Get config needed for Simulation Manager
	cfg, err := config.GetConfig()
	if err != nil {
		// Log the error but attempt to proceed if possible? Or fail fast?
		// For now, fail fast as config is likely essential.
		return nil, fmt.Errorf("failed to get config for simulation manager: %w", err)
	}
	simManager := simulation.NewManager(cfg, *logger)

	// --- Initialize Simulation Manager with Storage ---
	logger.Debug("Initializing simulation manager storage", "run_dir", runDir)
	motionStore, err := storage.NewStorage(runDir, storage.MOTION, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create motion storage in %s: %w", runDir, err)
	}
	eventsStore, err := storage.NewStorage(runDir, storage.EVENTS, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create events storage in %s: %w", runDir, err)
	}
	dynamicsStore, err := storage.NewStorage(runDir, storage.DYNAMICS, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamics storage in %s: %w", runDir, err)
	}
	stores := &storage.Stores{
		Motion:   motionStore,
		Events:   eventsStore,
		Dynamics: dynamicsStore,
	}
	if err := simManager.Initialize(stores); err != nil {
		return nil, fmt.Errorf("failed to initialize simulation manager: %w", err)
	}
	// Defer closing stores
	defer motionStore.Close()
	defer eventsStore.Close()
	defer dynamicsStore.Close()
	// --- End Initialize ---

	err = simManager.Run()
	if err != nil {
		return nil, fmt.Errorf("simulation run failed: %w", err)
	}
	logger.Info("Simulation completed successfully")

	// --- 3. Load Ground Truth Data ---
	logger.Info("Loading ground truth data", "source_dir", entry.DataDir)
	// Ensure DataDir is absolute or resolve relative to CWD
	gtDataDir, err := filepath.Abs(entry.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for ground truth data dir '%s': %w", entry.DataDir, err)
	}
	if _, err := os.Stat(gtDataDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("ground truth data dir '%s' not found at resolved path '%s'", entry.DataDir, gtDataDir)
	}
	logger.Debug("Using ground truth data directory", "path", gtDataDir)

	flightInfoPath := filepath.Join(gtDataDir, "fl001 - flight_info_processed.csv")
	flightInfoGroundTruth, err := LoadFlightInfo(flightInfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ground truth flight info from '%s': %w", flightInfoPath, err)
	}
	eventInfoPath := filepath.Join(gtDataDir, "fl001 - event_info_processed.csv")
	eventInfoGroundTruth, err := LoadEventInfo(eventInfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ground truth event info from '%s': %w", eventInfoPath, err)
	}
	// Note: FlightState ground truth might not be needed if comparing only high-level metrics
	logger.Info("Ground truth data loaded", "flight_info_count", len(flightInfoGroundTruth), "event_info_count", len(eventInfoGroundTruth))

	// --- 4. Load Simulation Results ---
	logger.Info("Loading simulation results")
	simDynamicsPath := filepath.Join(runDir, "dynamics.csv")
	simDynamicsData, err := LoadSimDynamicsData(simDynamicsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load simulation dynamics data from '%s': %w", simDynamicsPath, err)
	}
	simEventInfoPath := filepath.Join(runDir, "events.csv")
	simEventInfo, err := LoadSimEventData(simEventInfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load simulation event info from '%s': %w", simEventInfoPath, err)
	}
	logger.Info("Simulation results loaded", "dynamics_count", len(simDynamicsData), "event_info_count", len(simEventInfo))

	// --- 5. Compare Results ---
	logger.Info("Comparing ground truth and simulation results...")
	var results []BenchmarkResult
	const tolerance = 0.05

	// --- Compare Apogee ---
	gtApogee, gtApogeeTime := findGroundTruthApogee(flightInfoGroundTruth)
	simApogee, simApogeeTime := findSimApogee(simDynamicsData)
	logger.Debug("Apogee Comparison", "ground_truth", gtApogee, "sim", simApogee)
	benchmarkResultApogee := compareFloat("Apogee", "Maximum altitude reached (m)", gtApogee, simApogee, tolerance)
	results = append(results, benchmarkResultApogee)
	if !benchmarkResultApogee.Passed {
		logger.Warn("Comparison Failed: Apogee", "expected", gtApogee, "actual", simApogee, "tolerance", tolerance)
	}

	logger.Debug("Apogee Time Comparison", "ground_truth", gtApogeeTime, "sim", simApogeeTime)
	benchmarkResultApogeeTime := compareFloat("Apogee Time", "Time of maximum altitude (s)", gtApogeeTime, simApogeeTime, tolerance)
	results = append(results, benchmarkResultApogeeTime)
	if !benchmarkResultApogeeTime.Passed {
		logger.Warn("Comparison Failed: Apogee Time", "expected", gtApogeeTime, "actual", simApogeeTime, "tolerance", tolerance)
	}

	// --- Compare Max Velocity ---
	gtMaxVel, gtMaxVelTime := findGroundTruthMaxVelocity(flightInfoGroundTruth)
	simMaxVel, simMaxVelTime := findSimMaxVelocity(simDynamicsData)
	logger.Debug("Max Velocity Comparison", "ground_truth", gtMaxVel, "sim", simMaxVel)
	benchmarkResultMaxVel := compareFloat("Max Velocity", "Maximum velocity reached (m/s)", gtMaxVel, simMaxVel, tolerance)
	results = append(results, benchmarkResultMaxVel)
	if !benchmarkResultMaxVel.Passed {
		logger.Warn("Comparison Failed: Max Velocity", "expected", gtMaxVel, "actual", simMaxVel, "tolerance", tolerance)
	}

	logger.Debug("Max Velocity Time Comparison", "ground_truth", gtMaxVelTime, "sim", simMaxVelTime)
	benchmarkResultMaxVelTime := compareFloat("Max Velocity Time", "Time of maximum velocity (s)", gtMaxVelTime, simMaxVelTime, tolerance)
	results = append(results, benchmarkResultMaxVelTime)
	if !benchmarkResultMaxVelTime.Passed {
		logger.Warn("Comparison Failed: Max Velocity Time", "expected", gtMaxVelTime, "actual", simMaxVelTime, "tolerance", tolerance)
	}

	// --- Compare Impact Velocity ---
	gtImpactVel, _ := findGroundTruthImpactVelocity(flightInfoGroundTruth) // Assuming time is not compared here
	simImpactVel, _ := findSimImpactVelocity(simDynamicsData)              // Assuming time is not compared here
	logger.Debug("Impact Velocity Comparison", "ground_truth", gtImpactVel, "sim", simImpactVel)
	benchmarkResultImpactVel := compareFloat("Impact Velocity", "Velocity at impact (m/s)", gtImpactVel, simImpactVel, tolerance)
	results = append(results, benchmarkResultImpactVel)
	if !benchmarkResultImpactVel.Passed {
		logger.Warn("Comparison Failed: Impact Velocity", "expected", gtImpactVel, "actual", simImpactVel, "tolerance", tolerance)
	}

	// --- Compare Flight Duration ---
	gtDuration, _ := findGroundTruthFlightDuration(flightInfoGroundTruth) // Assuming event time is the duration?
	simDuration, _ := findSimFlightDuration(simDynamicsData)              // Assuming last time step is duration?
	logger.Debug("Flight Duration Comparison", "ground_truth", gtDuration, "sim", simDuration)
	benchmarkResultDuration := compareFloat("Flight Duration", "Total flight time (s)", gtDuration, simDuration, tolerance)
	results = append(results, benchmarkResultDuration)
	if !benchmarkResultDuration.Passed {
		logger.Warn("Comparison Failed: Flight Duration", "expected", gtDuration, "actual", simDuration, "tolerance", tolerance)
	}

	// --- Compare Event Times ---
	eventMappings := map[string]string{
		// GroundTruth Event Name : Sim Event Name
		"EV_MAIN_DEPLOYMENT": "MAIN_DEPLOYMENT",
		// Standard events that usually match or are derived differently
		// "EV_LIFTOFF": "LIFTOFF", // Liftoff usually time 0 or handled separately
		// "EV_APOGEE": "APOGEE", // Apogee usually derived from dynamics data
		// "EV_MAX_V": "MAX_VELOCITY", // Max Velocity usually derived from dynamics data
	}

	for gtEvent, simEvent := range eventMappings {
		gtEventTime := findGroundTruthEventTime(eventInfoGroundTruth, gtEvent, logger)
		if gtEventTime < 0 {
			logger.Warn("Target ground truth event not found in data", "event_name", gtEvent)
			metricName := fmt.Sprintf("%s Time", gtEvent)
			results = append(results, BenchmarkResult{
				Metric:        metricName,
				Description:   fmt.Sprintf("Compare %s time", gtEvent),
				Expected:      math.NaN(),
				Actual:        findSimEventTime(simEventInfo, simEvent, logger),
				Tolerance:     0.05,
				ToleranceType: "relative",
				Passed:        false,
			})
			continue // Skip comparison if GT event is missing
		}

		// Find corresponding sim time
		simEventTime := findSimEventTime(simEventInfo, simEvent, logger)
		if simEventTime < 0 {
			logger.Warn("Target sim event not found", "event_name", simEvent)
			metricName := fmt.Sprintf("%s Time", gtEvent)
			results = append(results, BenchmarkResult{
				Metric:        metricName,
				Description:   fmt.Sprintf("Compare %s time", gtEvent),
				Expected:      gtEventTime,
				Actual:        math.NaN(),
				Tolerance:     0.05,
				ToleranceType: "relative",
				Passed:        false,
			})
			continue // Skip comparison if sim event is missing
		}

		// Compare times
		logger.Debug("Event Time Comparison", "event", gtEvent, "ground_truth", gtEventTime, "sim", simEventTime)
		benchmarkResultEventTime := compareFloat(fmt.Sprintf("%s Time", gtEvent), fmt.Sprintf("Time of %s event (s)", strings.ToLower(gtEvent)), gtEventTime, simEventTime, tolerance)
		results = append(results, benchmarkResultEventTime)
		if !benchmarkResultEventTime.Passed {
			logger.Warn("Comparison Failed: Event Time", "event", gtEvent, "expected", gtEventTime, "actual", simEventTime, "tolerance", tolerance)
		}
	}

	logger.Info("Comparison finished")
	return results, nil
}

// --- Helper methods for SIMULATION data --- //

// findSimApogee finds the maximum altitude from simulation dynamics data.
func findSimApogee(simData []SimDynamicsData) (float64, float64) {
	if len(simData) == 0 {
		return 0, 0
	}
	maxAltitude := simData[0].PositionZ
	timestamp := simData[0].Timestamp
	for _, p := range simData {
		if p.PositionZ > maxAltitude {
			maxAltitude = p.PositionZ
			timestamp = p.Timestamp
		}
	}
	return maxAltitude, timestamp
}

// findSimMaxVelocity finds the maximum velocity magnitude from simulation dynamics data.
func findSimMaxVelocity(simData []SimDynamicsData) (float64, float64) {
	if len(simData) == 0 {
		return 0, 0
	}
	maxVelocityMag := 0.0
	timestamp := 0.0
	if len(simData) > 0 {
		timestamp = simData[0].Timestamp
	}

	for _, p := range simData {
		// Calculate velocity magnitude
		velMag := math.Sqrt(p.VelocityX*p.VelocityX + p.VelocityY*p.VelocityY + p.VelocityZ*p.VelocityZ)
		if velMag > maxVelocityMag {
			maxVelocityMag = velMag
			timestamp = p.Timestamp
		}
	}
	return maxVelocityMag, timestamp
}

// findSimEventTime finds the timestamp for a specific event from simulation event info.
// Note: Matches event name string case-insensitively.
func findSimEventTime(simEvents []SimEventInfo, targetEventName string, logger *logf.Logger) float64 {
	for _, e := range simEvents {
		// Case-insensitive comparison
		if strings.EqualFold(e.EventName, targetEventName) {
			logger.Debug("Found target sim event", "target", targetEventName, "foundEvent", e.EventName, "timestamp", e.Time)
			return e.Time
		}
	}

	logger.Warn("Target sim event not found", "targetEvent", targetEventName)
	return -1
}

// --- Helper methods for GROUND TRUTH data --- //

// findGroundTruthApogee finds the maximum altitude from ground truth flight info.
func findGroundTruthApogee(gtData []FlightInfo) (float64, float64) {
	if len(gtData) == 0 {
		return 0, 0
	}
	maxAltitude := gtData[0].Height
	timestamp := gtData[0].Timestamp
	for _, p := range gtData {
		if p.Height > maxAltitude {
			maxAltitude = p.Height
			timestamp = p.Timestamp
		}
	}
	return maxAltitude, timestamp
}

// findGroundTruthMaxVelocity finds the maximum velocity from ground truth flight info.
func findGroundTruthMaxVelocity(gtData []FlightInfo) (float64, float64) {
	if len(gtData) == 0 {
		return 0, 0
	}
	maxVelocity := gtData[0].Velocity
	timestamp := gtData[0].Timestamp
	for _, p := range gtData {
		if p.Velocity > maxVelocity {
			maxVelocity = p.Velocity
			timestamp = p.Timestamp
		}
	}
	return maxVelocity, timestamp
}

// findGroundTruthEventTime finds the timestamp for a specific event from ground truth event info.
func findGroundTruthEventTime(gtEvents []EventInfo, eventName string, logger *logf.Logger) float64 {
	for _, e := range gtEvents {
		// Need case-insensitive comparison or ensure ground truth event names match sim names exactly
		logger.Debug("Comparing GT event", "expected", eventName, "actual_in_csv", e.Event, "equal_fold", strings.EqualFold(e.Event, eventName))
		if strings.EqualFold(e.Event, eventName) {
			return e.Timestamp
		}
	}
	logger.Warn("Target ground truth event not found", "targetEvent", eventName)
	return -1
}

// findGroundTruthImpactVelocity finds the impact velocity from ground truth flight info.
func findGroundTruthImpactVelocity(gtData []FlightInfo) (float64, float64) {
	if len(gtData) == 0 {
		return 0, 0
	}
	impactVelocity := gtData[len(gtData)-1].Velocity
	timestamp := gtData[len(gtData)-1].Timestamp
	return impactVelocity, timestamp
}

// findGroundTruthFlightDuration finds the flight duration from ground truth flight info.
func findGroundTruthFlightDuration(gtData []FlightInfo) (float64, float64) {
	if len(gtData) == 0 {
		return 0, 0
	}
	flightDuration := gtData[len(gtData)-1].Timestamp - gtData[0].Timestamp
	return flightDuration, 0
}

// findSimImpactVelocity finds the impact velocity from simulation dynamics data.
func findSimImpactVelocity(simData []SimDynamicsData) (float64, float64) {
	if len(simData) == 0 {
		return 0, 0
	}
	impactVelocity := math.Sqrt(simData[len(simData)-1].VelocityX*simData[len(simData)-1].VelocityX + simData[len(simData)-1].VelocityY*simData[len(simData)-1].VelocityY + simData[len(simData)-1].VelocityZ*simData[len(simData)-1].VelocityZ)
	timestamp := simData[len(simData)-1].Timestamp
	return impactVelocity, timestamp
}

// findSimFlightDuration finds the flight duration from simulation dynamics data.
func findSimFlightDuration(simData []SimDynamicsData) (float64, float64) {
	if len(simData) == 0 {
		return 0, 0
	}
	flightDuration := simData[len(simData)-1].Timestamp - simData[0].Timestamp
	return flightDuration, 0
}
