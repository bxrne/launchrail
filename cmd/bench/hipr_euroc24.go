package main

import (
	"fmt"
	"math" // Added for velocity magnitude calculation
	"os"
	"path/filepath"
	"strings"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client" // Added import
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage" // Added import
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
func NewHiprEuroc24Benchmark() *HiprEuroc24Benchmark { // Removed config param
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
	httpClient := http_client.NewHTTPClient() // Create default HTTP client
	motorData, err := thrustcurves.Load(entry.MotorDesignation, httpClient)
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
	simManager := simulation.NewManager(cfg, *logger) // Dereference logger

	// --- Initialize Simulation Manager with Storage ---
	logger.Debug("Initializing simulation manager storage", "run_dir", runDir)
	motionStore, err := storage.NewStorage(runDir, storage.MOTION)
	if err != nil {
		return nil, fmt.Errorf("failed to create motion storage in %s: %w", runDir, err)
	}
	eventsStore, err := storage.NewStorage(runDir, storage.EVENTS)
	if err != nil {
		return nil, fmt.Errorf("failed to create events storage in %s: %w", runDir, err)
	}
	dynamicsStore, err := storage.NewStorage(runDir, storage.DYNAMICS)
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
	flightInfoGroundTruth, err := LoadFlightInfo(flightInfoPath) // Using function from datastore.go
	if err != nil {
		return nil, fmt.Errorf("failed to load ground truth flight info from '%s': %w", flightInfoPath, err)
	}
	eventInfoPath := filepath.Join(gtDataDir, "fl001 - event_info_processed.csv")
	eventInfoGroundTruth, err := LoadEventInfo(eventInfoPath) // Using function from datastore.go
	if err != nil {
		return nil, fmt.Errorf("failed to load ground truth event info from '%s': %w", eventInfoPath, err)
	}
	// Note: FlightState ground truth might not be needed if comparing only high-level metrics
	logger.Info("Ground truth data loaded", "flight_info_count", len(flightInfoGroundTruth), "event_info_count", len(eventInfoGroundTruth))

	// --- 4. Load Simulation Results ---
	logger.Info("Loading simulation results")
	simDynamicsPath := filepath.Join(runDir, "dynamics.csv")     // Load dynamics.csv
	simDynamicsData, err := LoadSimDynamicsData(simDynamicsPath) // Use LoadSimDynamicsData
	if err != nil {
		return nil, fmt.Errorf("failed to load simulation dynamics data from '%s': %w", simDynamicsPath, err)
	}
	simEventInfoPath := filepath.Join(runDir, "events.csv") // Load events.csv
	simEventInfo, err := LoadSimEventData(simEventInfoPath) // Use LoadSimEventData
	if err != nil {
		return nil, fmt.Errorf("failed to load simulation event info from '%s': %w", simEventInfoPath, err)
	}
	logger.Info("Simulation results loaded", "dynamics_count", len(simDynamicsData), "event_info_count", len(simEventInfo))

	// --- 5. Compare Results ---
	logger.Info("Comparing ground truth and simulation results...")
	var results []BenchmarkResult
	const tolerance = 0.05 // 5% tolerance

	// --- Compare Apogee ---
	gtApogee, gtApogeeTime := findGroundTruthApogee(flightInfoGroundTruth)
	simApogee, simApogeeTime := findSimApogee(simDynamicsData) // Pass correct data type
	logger.Debug("Apogee Comparison", "ground_truth", gtApogee, "sim", simApogee)
	results = append(results, compareFloat("Apogee", "Maximum altitude reached (m)", gtApogee, simApogee, tolerance))
	logger.Debug("Apogee Time Comparison", "ground_truth", gtApogeeTime, "sim", simApogeeTime)
	results = append(results, compareFloat("Apogee Time", "Time of maximum altitude (s)", gtApogeeTime, simApogeeTime, tolerance))

	// --- Compare Max Velocity ---
	gtMaxVel, gtMaxVelTime := findGroundTruthMaxVelocity(flightInfoGroundTruth)
	simMaxVel, simMaxVelTime := findSimMaxVelocity(simDynamicsData) // Pass correct data type
	logger.Debug("Max Velocity Comparison", "ground_truth", gtMaxVel, "sim", simMaxVel)
	results = append(results, compareFloat("Max Velocity", "Maximum velocity reached (m/s)", gtMaxVel, simMaxVel, tolerance))
	logger.Debug("Max Velocity Time Comparison", "ground_truth", gtMaxVelTime, "sim", simMaxVelTime)
	results = append(results, compareFloat("Max Velocity Time", "Time of maximum velocity (s)", gtMaxVelTime, simMaxVelTime, tolerance))

	// --- Compare Event Times ---
	eventMappings := map[string]string{
		// GroundTruth Event Name : Sim Event Name
		"Burnout":         "MOTOR_BURNOUT", // Corrected sim event name
		"Drogue Deployed": "DROGUE_DEPLOY", // Corrected sim event name
		"Main Deployed":   "MAIN_DEPLOY",   // Corrected sim event name
		// Add more mappings if needed
	}

	for gtEvent, simEvent := range eventMappings {
		gtEventTime := findGroundTruthEventTime(eventInfoGroundTruth, gtEvent)
		if gtEventTime < 0 {
			logger.Warn("Target ground truth event not found in data", "event_name", gtEvent)
			metricName := fmt.Sprintf("%s Time", gtEvent)
			results = append(results, BenchmarkResult{
				Metric:        metricName,
				Description:   fmt.Sprintf("Compare %s time", gtEvent),
				Expected:      math.NaN(),
				Actual:        findSimEventTime(simEventInfo, simEvent, logger),
				Tolerance:     0.05, // Revert to hardcoded 5% tolerance
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
				Tolerance:     0.05, // Revert to hardcoded 5% tolerance
				ToleranceType: "relative",
				Passed:        false,
			})
			continue // Skip comparison if sim event is missing
		}

		// Compare times
		logger.Debug("Event Time Comparison", "event", gtEvent, "ground_truth", gtEventTime, "sim", simEventTime)
		results = append(results, compareFloat(fmt.Sprintf("%s Time", gtEvent), fmt.Sprintf("Time of %s event (s)", strings.ToLower(gtEvent)), gtEventTime, simEventTime, tolerance))
	}

	logger.Info("Comparison finished")
	return results, nil
}

// --- Helper methods for SIMULATION data --- //

// findSimApogee finds the maximum altitude from simulation dynamics data.
func findSimApogee(simData []SimDynamicsData) (float64, float64) { // Changed type
	if len(simData) == 0 {
		return 0, 0
	}
	maxAltitude := simData[0].PositionZ // Use PositionZ for altitude
	timestamp := simData[0].Timestamp
	for _, p := range simData {
		if p.PositionZ > maxAltitude { // Use PositionZ
			maxAltitude = p.PositionZ // Use PositionZ
			timestamp = p.Timestamp
		}
	}
	return maxAltitude, timestamp
}

// findSimMaxVelocity finds the maximum velocity magnitude from simulation dynamics data.
func findSimMaxVelocity(simData []SimDynamicsData) (float64, float64) { // Changed type
	if len(simData) == 0 {
		return 0, 0
	}
	maxVelocityMag := 0.0
	timestamp := 0.0
	if len(simData) > 0 {
		timestamp = simData[0].Timestamp // Initial timestamp
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
func findSimEventTime(simEvents []SimEventInfo, targetEventName string, logger *logf.Logger) float64 { // Pass logger
	for _, e := range simEvents {
		// Case-insensitive comparison
		if strings.EqualFold(e.EventName, targetEventName) {
			logger.Debug("Found target sim event", "target", targetEventName, "foundEvent", e.EventName, "timestamp", e.Time)
			return e.Time
		}
	}

	logger.Warn("Target sim event not found", "targetEvent", targetEventName)
	return -1 // Not found
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
	return maxVelocity, timestamp // Corrected return value
}

// findGroundTruthEventTime finds the timestamp for a specific event from ground truth event info.
func findGroundTruthEventTime(gtEvents []EventInfo, eventName string) float64 {
	for _, e := range gtEvents {
		// Need case-insensitive comparison or ensure ground truth event names match sim names exactly
		if strings.EqualFold(e.Event, eventName) { // Make comparison case-insensitive
			return e.Timestamp
		}
	}
	return -1 // Indicate not found
}
