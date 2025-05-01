package main

import (
	"fmt"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/logger"
)

// HiprEuroc24Benchmark implements the Benchmark interface for the specific dataset.
type HiprEuroc24Benchmark struct {
	name string
	// Ground truth data loaded from CSVs
	flightInfoGroundTruth  []FlightInfo  // Renamed for clarity
	eventInfoGroundTruth   []EventInfo   // Renamed for clarity
	flightStateGroundTruth []FlightState // Renamed for clarity

	config BenchmarkConfig // Added configuration
}

// NewHiprEuroc24Benchmark creates a new instance.
func NewHiprEuroc24Benchmark(config BenchmarkConfig) *HiprEuroc24Benchmark {
	return &HiprEuroc24Benchmark{
		name:   "hipr-euroc24",
		config: config, // Store config
	}
}

// Name returns the benchmark name.
func (b *HiprEuroc24Benchmark) Name() string {
	return b.name
}

// Setup performs any necessary setup for the benchmark (currently none needed).
func (b *HiprEuroc24Benchmark) Setup() error {
	// No specific setup needed for this benchmark yet.
	return nil
}

// LoadData loads all necessary CSV files for the ground truth data.
func (b *HiprEuroc24Benchmark) LoadData(benchDataDir string) error {
	var err error
	benchLogger := logger.GetLogger("info") // Get logger instance

	benchLogger.Info("Loading GROUND TRUTH data for benchmark", "name", b.Name(), "sourceDir", benchDataDir)

	// The benchDataDir argument already points to the specific benchmark's data directory.
	// No need to join with b.Name() again.
	basePath := benchDataDir // Use benchDataDir directly

	flightInfoPath := filepath.Join(basePath, "fl001 - flight_info_processed.csv")
	b.flightInfoGroundTruth, err = LoadFlightInfo(flightInfoPath) // Load into renamed field
	if err != nil {
		return fmt.Errorf("failed to load ground truth flight info: %w", err)
	}
	benchLogger.Info("Loaded ground truth flight info", "count", len(b.flightInfoGroundTruth), "file", flightInfoPath)

	eventInfoPath := filepath.Join(basePath, "fl001 - event_info_processed.csv")
	b.eventInfoGroundTruth, err = LoadEventInfo(eventInfoPath)
	if err != nil {
		return fmt.Errorf("failed to load ground truth event info: %w", err)
	}
	benchLogger.Info("Loaded ground truth event info", "count", len(b.eventInfoGroundTruth), "file", eventInfoPath)

	flightStatePath := filepath.Join(basePath, "fl001 - flight_states_processed.csv")
	b.flightStateGroundTruth, err = LoadFlightStates(flightStatePath)
	if err != nil {
		return fmt.Errorf("failed to load ground truth flight states: %w", err)
	}
	benchLogger.Info("Loaded ground truth flight states", "count", len(b.flightStateGroundTruth), "file", flightStatePath)

	return nil
}

// Run performs the comparison between simulation results and ground truth data.
func (b *HiprEuroc24Benchmark) Run() ([]BenchmarkResult, error) {
	benchLogger := logger.GetLogger("info") // Get logger instance
	var results []BenchmarkResult

	// --- Load Actual Simulation Data --- // UPDATED
	actualDataPath := b.config.ResultDirPath // Load directly from the results dir
	benchLogger.Info("Loading actual simulation data", "recordPath", actualDataPath)

	actualMotionPath := filepath.Join(actualDataPath, "MOTION.csv")
	actualMotionData, err := LoadSimMotionData(actualMotionPath) // USE NEW LOADER
	if err != nil {
		return nil, fmt.Errorf("failed to load actual motion data from '%s': %w", actualMotionPath, err)
	}
	benchLogger.Debug("Loaded actual motion data", "count", len(actualMotionData))

	actualEventsPath := filepath.Join(actualDataPath, "EVENTS.csv")
	actualEventsData, err := LoadSimEventData(actualEventsPath) // USE NEW LOADER
	if err != nil {
		return nil, fmt.Errorf("failed to load actual event data from '%s': %w", actualEventsPath, err)
	}
	benchLogger.Debug("Loaded actual event data", "count", len(actualEventsData))

	// Load dynamics data, although it's not used in comparisons yet
	actualDynamicsPath := filepath.Join(actualDataPath, "DYNAMICS.csv")
	actualDynamicsData, err := LoadSimDynamicsData(actualDynamicsPath) // USE NEW LOADER
	if err != nil {
		// Log warning instead of error? Or keep error if dynamics are essential later?
		benchLogger.Warn("Failed to load actual dynamics data, comparisons using it will fail/be skipped", "file", actualDynamicsPath, "error", err)
		// return nil, fmt.Errorf("failed to load actual dynamics data from '%s': %w", actualDynamicsPath, err)
	}
	benchLogger.Debug("Loaded actual dynamics data", "count", len(actualDynamicsData))

	// --- Perform Comparisons ---
	benchLogger.Info("Performing comparisons...")

	// --- Define Ground Truth Expected Values (from ground truth CSVs) --- // UPDATED HELPER CALLS
	// Note: We still use the ground truth helpers (findGroundTruth*) for ground truth data.
	expectedApogeeGroundTruth, _ := findGroundTruthApogee(b.flightInfoGroundTruth)
	expectedMaxVGroundTruth, _ := findGroundTruthMaxVelocity(b.flightInfoGroundTruth)
	expectedLiftoffTimeGroundTruth := findGroundTruthEventTime(b.eventInfoGroundTruth, "LIFTOFF")
	expectedApogeeEventTimeGroundTruth := findGroundTruthEventTime(b.eventInfoGroundTruth, "APOGEE")
	// expectedTouchdownTimeGroundTruth := findGroundTruthStateTime(b.flightStateGroundTruth, "TOUCHDOWN") // Assuming a helper for states exists
	expectedTouchdownTimeGroundTruth := 0.0
	benchLogger.Warn("Touchdown time ground truth determination skipped: findGroundTruthStateTime needs implementation/verification")

	// --- Calculate Actual Values from SIMULATION Data --- // UPDATED HELPER CALLS
	actualApogeeFromSim, _ := findSimApogee(actualMotionData)             // Use sim helper
	actualMaxVFromSim, _ := findSimMaxVelocity(actualMotionData)        // Use sim helper
	actualLiftoffTimeFromSim := findSimEventTime(actualEventsData, "Liftoff") // Use sim helper
	actualApogeeEventTimeFromSim := findSimEventTime(actualEventsData, "ApogeeDetected") // Use sim helper
	actualTouchdownTimeFromSim := 0.0
	benchLogger.Warn("Touchdown time comparison skipped: Need logic to find touchdown in sim data")

	// --- Compare Metrics ---
	apogeeTolerance := 0.05 // 5% tolerance
	results = append(results, compareFloat("Apogee Altitude", "Compare simulation apogee vs ground truth", expectedApogeeGroundTruth, actualApogeeFromSim, apogeeTolerance))

	maxVTolerance := 0.05 // 5% tolerance
	results = append(results, compareFloat("Max Velocity", "Compare simulation max velocity vs ground truth", expectedMaxVGroundTruth, actualMaxVFromSim, maxVTolerance))

	liftoffTolerance := 0.1 // 100ms tolerance
	results = append(results, compareFloat("Liftoff Time", "Compare sim liftoff event time vs ground truth", expectedLiftoffTimeGroundTruth, actualLiftoffTimeFromSim, liftoffTolerance))

	apogeeTimeTolerance := 0.5 // 500ms tolerance
	results = append(results, compareFloat("Apogee Event Time", "Compare sim apogee event time vs ground truth", expectedApogeeEventTimeGroundTruth, actualApogeeEventTimeFromSim, apogeeTimeTolerance))

	touchdownTolerance := 1.0 // 1s tolerance
	results = append(results, compareFloat("Touchdown Time", "Compare sim touchdown state time vs ground truth", expectedTouchdownTimeGroundTruth, actualTouchdownTimeFromSim, touchdownTolerance))

	benchLogger.Info("Comparison complete.")
	return results, nil
}

// --- Helper methods for SIMULATION data (from loaded CSV structs) --- // NEW

// findSimApogee finds the maximum altitude from simulation flight info.
func findSimApogee(simData []SimMotionData) (float64, float64) { // UPDATED TYPE
	if len(simData) == 0 {
		return 0, 0
	}
	maxAltitude := simData[0].Altitude // UPDATED FIELD
	timestamp := simData[0].Timestamp
	for _, p := range simData {
		if p.Altitude > maxAltitude { // UPDATED FIELD
			maxAltitude = p.Altitude // UPDATED FIELD
			timestamp = p.Timestamp
		}
	}
	return maxAltitude, timestamp
}

// findSimMaxVelocity finds the maximum velocity from simulation flight info.
func findSimMaxVelocity(simData []SimMotionData) (float64, float64) { // UPDATED TYPE
	if len(simData) == 0 {
		return 0, 0
	}
	maxVelocity := simData[0].Velocity // UPDATED FIELD
	timestamp := simData[0].Timestamp
	for _, p := range simData {
		if p.Velocity > maxVelocity { // UPDATED FIELD
			maxVelocity = p.Velocity // UPDATED FIELD
			timestamp = p.Timestamp
		}
	}
	return maxVelocity, timestamp
}

// findSimEventTime finds the timestamp for a specific event from simulation event info.
// Note: Matches event name string exactly.
func findSimEventTime(simEvents []SimEventData, eventName string) float64 { // UPDATED TYPE
	// The actual EVENTS.csv has status strings, not event names like the ground truth.
	// We need to infer events from status changes.
	benchLogger := logger.GetLogger("debug") // Use debug for potentially verbose logging

	if eventName == "Liftoff" {
		// Assumption: Liftoff occurs when motor status transitions from IDLE to BURNING.
		var prevMotorStatus string = ""
		if len(simEvents) > 0 {
			prevMotorStatus = simEvents[0].MotorStatus // Initialize with first status
		}
		for i := 1; i < len(simEvents); i++ { // Start from second record
			current := simEvents[i]
			if prevMotorStatus == "IDLE" && current.MotorStatus == "BURNING" {
				benchLogger.Debug("Found Liftoff event", "timestamp", current.Timestamp, "prevStatus", prevMotorStatus, "currentStatus", current.MotorStatus)
				return current.Timestamp
			}
			prevMotorStatus = current.MotorStatus
		}
		benchLogger.Warn("Liftoff event (IDLE -> BURNING transition) not found in sim data")
		return -1 // Not found

	} else if eventName == "ApogeeDetected" {
		// Assumption: ApogeeDetected corresponds to the first time ParachuteStatus is DEPLOYED.
		for _, e := range simEvents {
			if e.ParachuteStatus == "DEPLOYED" {
				benchLogger.Debug("Found ApogeeDetected event", "timestamp", e.Timestamp, "parachuteStatus", e.ParachuteStatus)
				return e.Timestamp
			}
		}
		benchLogger.Warn("ApogeeDetected event (ParachuteStatus == DEPLOYED) not found in sim data")
		return -1 // Not found

	} else {
		benchLogger.Warn("Unsupported event name for findSimEventTime", "targetEvent", eventName)
		return -1 // Unsupported event
	}
}

// --- Helper methods for GROUND TRUTH data --- // NEW SECTION

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
func findGroundTruthEventTime(gtEvents []EventInfo, eventName string) float64 {
	for _, e := range gtEvents {
		if e.Event == eventName {
			return e.Timestamp
		}
	}
	return -1 // Indicate not found
}

// TODO: Add findGroundTruthStateTime if needed for FlightState data comparisons
