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
	results := []BenchmarkResult{}

	benchLogger.Info("Running benchmark comparison", "name", b.Name(), "simRecordHash", b.config.SimRecordHash)

	// --- Load SIMULATION Data from Record --- // NEW LOGIC
	simRecord, err := b.config.RecordManager.GetRecord(b.config.SimRecordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get simulation record '%s': %w", b.config.SimRecordHash, err)
	}
	benchLogger.Info("Loaded simulation record", "hash", simRecord.Hash, "path", simRecord.Path)

	// Construct paths to simulation CSVs within the record's directory
	simFlightInfoPath := filepath.Join(simRecord.Path, "flight_info.csv")
	simEventInfoPath := filepath.Join(simRecord.Path, "events.csv")           // Assuming standard name
	simFlightStatesPath := filepath.Join(simRecord.Path, "flight_states.csv") // Assuming standard name

	// Load simulation data using existing loaders
	simFlightInfo, err := LoadFlightInfo(simFlightInfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load simulation flight info from %s: %w", simFlightInfoPath, err)
	}
	benchLogger.Info("Loaded simulation flight info", "count", len(simFlightInfo))

	simEventInfo, err := LoadEventInfo(simEventInfoPath)
	if err != nil {
		// Note: Allow missing events file for now? Or return error?
		benchLogger.Warn("Failed to load simulation event info, continuing without event comparison", "file", simEventInfoPath, "error", err)
		// return nil, fmt.Errorf("failed to load simulation event info from %s: %w", simEventInfoPath, err)
	}
	benchLogger.Info("Loaded simulation event info", "count", len(simEventInfo))

	simFlightStates, err := LoadFlightStates(simFlightStatesPath)
	if err != nil {
		// Note: Allow missing states file for now? Or return error?
		benchLogger.Warn("Failed to load simulation flight states, continuing without state comparison", "file", simFlightStatesPath, "error", err)
		// return nil, fmt.Errorf("failed to load simulation flight states from %s: %w", simFlightStatesPath, err)
	}
	benchLogger.Info("Loaded simulation flight states", "count", len(simFlightStates))

	// --- Define Ground Truth Expected Values (from loaded CSVs) ---
	expectedApogeeGroundTruth, _ := b.findGroundTruthApogee()
	expectedMaxVGroundTruth, _ := b.findGroundTruthMaxVelocity()
	expectedLiftoffTimeGroundTruth := b.findGroundTruthEventTime("EV_LIFTOFF")
	expectedApogeeEventTimeGroundTruth := b.findGroundTruthEventTime("EV_APOGEE")
	expectedTouchdownTimeGroundTruth := b.findGroundTruthStateTime("TOUCHDOWN")

	// --- Calculate Actual Values from SIMULATION Data --- // UPDATED with sim data
	actualApogeeFromSim, _ := findSimApogee(simFlightInfo)
	actualMaxVFromSim, _ := findSimMaxVelocity(simFlightInfo)
	actualLiftoffTimeFromSim := findSimEventTime(simEventInfo, "Liftoff")            // Use names from dummy data
	actualApogeeEventTimeFromSim := findSimEventTime(simEventInfo, "ApogeeDetected") // Use names from dummy data
	// actualTouchdownTimeFromSim := findSimStateTime(simFlightStates, "???") // Need the correct state name from sim output
	actualTouchdownTimeFromSim := 0.0 // Placeholder until state name is known
	benchLogger.Warn("Touchdown time comparison skipped: Unknown state name in simulation output CSV")

	// --- Perform Comparisons --- // UPDATED: Compare sim actual vs ground truth expected

	// 1. Apogee Comparison
	apogeeTolerance := 0.05 // 5% tolerance
	results = append(results, compareFloat("Apogee Altitude", "Compare simulation apogee vs ground truth", expectedApogeeGroundTruth, actualApogeeFromSim, apogeeTolerance))

	// 2. Max Velocity Comparison
	maxVTolerance := 0.05 // 5% tolerance
	results = append(results, compareFloat("Max Velocity", "Compare simulation max velocity vs ground truth", expectedMaxVGroundTruth, actualMaxVFromSim, maxVTolerance))

	// 3. Event Timing Comparisons
	liftoffTolerance := 0.1 // 100ms tolerance
	results = append(results, compareFloat("Liftoff Time", "Compare sim liftoff event time vs ground truth", expectedLiftoffTimeGroundTruth, actualLiftoffTimeFromSim, liftoffTolerance))

	apogeeTimeTolerance := 0.5 // 500ms tolerance
	results = append(results, compareFloat("Apogee Event Time", "Compare sim apogee event time vs ground truth", expectedApogeeEventTimeGroundTruth, actualApogeeEventTimeFromSim, apogeeTimeTolerance))

	// 4. State Timing Comparison (Placeholder)
	touchdownTolerance := 1.0 // 1s tolerance
	results = append(results, compareFloat("Touchdown Time", "Compare sim touchdown state time vs ground truth", expectedTouchdownTimeGroundTruth, actualTouchdownTimeFromSim, touchdownTolerance))

	benchLogger.Info("Comparison complete.")
	return results, nil
}

// --- Helper methods for GROUND TRUTH data (from Benchmark struct fields) ---
// findGroundTruthApogee finds max height from loaded ground truth flight info.
func (b *HiprEuroc24Benchmark) findGroundTruthApogee() (maxHeight float64, timestamp float64) {
	if len(b.flightInfoGroundTruth) == 0 {
		return 0, 0
	}
	maxHeight = b.flightInfoGroundTruth[0].Height
	timestamp = b.flightInfoGroundTruth[0].Timestamp
	for _, p := range b.flightInfoGroundTruth {
		if p.Height > maxHeight {
			maxHeight = p.Height
			timestamp = p.Timestamp
		}
	}
	return maxHeight, timestamp
}

// findGroundTruthMaxVelocity finds max velocity from loaded ground truth flight info.
func (b *HiprEuroc24Benchmark) findGroundTruthMaxVelocity() (maxVelocity float64, timestamp float64) {
	if len(b.flightInfoGroundTruth) == 0 {
		return 0, 0
	}
	maxVelocity = b.flightInfoGroundTruth[0].Velocity
	timestamp = b.flightInfoGroundTruth[0].Timestamp
	for _, p := range b.flightInfoGroundTruth {
		if p.Velocity > maxVelocity {
			maxVelocity = p.Velocity
			timestamp = p.Timestamp
		}
	}
	return maxVelocity, timestamp
}

func (b *HiprEuroc24Benchmark) findGroundTruthEventTime(eventName string) float64 {
	for _, e := range b.eventInfoGroundTruth {
		if e.Event == eventName {
			return e.Timestamp // Return first occurrence
		}
	}
	return -1 // Indicate not found
}

func (b *HiprEuroc24Benchmark) findGroundTruthStateTime(stateName string) float64 {
	for _, s := range b.flightStateGroundTruth {
		if s.State == stateName {
			return s.Timestamp // Return first occurrence
		}
	}
	return -1 // Indicate not found
}

// --- Helper methods for SIMULATION data (from loaded CSV structs) --- // NEW

// findSimApogee finds the maximum altitude from simulation flight info.
func findSimApogee(simData []FlightInfo) (float64, float64) {
	if len(simData) == 0 {
		return 0, 0
	}
	maxAltitude := simData[0].Height
	timestamp := simData[0].Timestamp
	for _, p := range simData {
		if p.Height > maxAltitude {
			maxAltitude = p.Height
			timestamp = p.Timestamp
		}
	}
	return maxAltitude, timestamp
}

// findSimMaxVelocity finds the maximum velocity from simulation flight info.
func findSimMaxVelocity(simData []FlightInfo) (float64, float64) {
	if len(simData) == 0 {
		return 0, 0
	}
	maxVelocity := simData[0].Velocity
	timestamp := simData[0].Timestamp
	for _, p := range simData {
		if p.Velocity > maxVelocity {
			maxVelocity = p.Velocity
			timestamp = p.Timestamp
		}
	}
	return maxVelocity, timestamp
}

// findSimEventTime finds the timestamp for a specific event from simulation event info.
// Note: Matches event name string exactly.
func findSimEventTime(simEvents []EventInfo, eventName string) float64 {
	for _, e := range simEvents {
		if e.Event == eventName {
			return e.Timestamp
		}
	}
	return -1 // Indicate not found
}

// findSimStateTime finds the timestamp for a specific state from simulation state info.
// Note: Matches state name string exactly.
func findSimStateTime(simStates []FlightState, stateName string) float64 {
	for _, s := range simStates {
		if s.State == stateName {
			return s.Timestamp
		}
	}
	return -1 // Indicate not found
}

// compareFloat compares expected and actual float values within a tolerance.
// Handles expected == 0 by using absolute difference.
