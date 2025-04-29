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

	// --- Load Actual Simulation Data ---
	// Construct paths based on the SimulationResultsDir from config
	// Assuming a single record ID for now, needs generalization later.
	// We need to determine the record ID to load. For now, hardcoding or simple logic.
	// Let's assume the benchmark name maps to a record ID or pattern.
	// This part needs refinement based on how record IDs are associated with benchmarks.
	// For hipr-euroc24, let's assume the record ID is 'rec001' or similar. 
	// TODO: Implement a robust way to find the relevant record ID.
	actualDataPath := b.config.ResultDirPath // Load directly from the results dir
	benchLogger.Info("Loading actual simulation data", "recordPath", actualDataPath)

	// Use generic filenames matching the actual simulation output
	actualMotionPath := filepath.Join(actualDataPath, "MOTION.csv") 
	actualMotionData, err := LoadFlightInfo(actualMotionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load actual motion data from record '%s': %w", actualMotionPath, err)
	}
	benchLogger.Debug("Loaded actual motion data", "count", len(actualMotionData))

	// Use generic filenames matching the actual simulation output
	actualEventsPath := filepath.Join(actualDataPath, "EVENTS.csv") 
	actualEventsData, err := LoadEventInfo(actualEventsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load actual event data from record '%s': %w", actualEventsPath, err)
	}
	benchLogger.Debug("Loaded actual event data", "count", len(actualEventsData))

	// Use generic filenames matching the actual simulation output
	actualDynamicsPath := filepath.Join(actualDataPath, "DYNAMICS.csv") 
	actualDynamicsData, err := LoadFlightStates(actualDynamicsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load actual dynamics data from record '%s': %w", actualDynamicsPath, err)
	}
	benchLogger.Debug("Loaded actual dynamics data", "count", len(actualDynamicsData))

	// --- Perform Comparisons ---
	benchLogger.Info("Performing comparisons...")

	// --- Define Ground Truth Expected Values (from loaded CSVs) --- 
	// Use the data loaded by LoadData method
	expectedApogeeGroundTruth, _ := findSimApogee(b.flightInfoGroundTruth) 
	expectedMaxVGroundTruth, _ := findSimMaxVelocity(b.flightInfoGroundTruth) 
	expectedLiftoffTimeGroundTruth := findSimEventTime(b.eventInfoGroundTruth, "LIFTOFF") 
	expectedApogeeEventTimeGroundTruth := findSimEventTime(b.eventInfoGroundTruth, "APOGEE") 
	expectedTouchdownTimeGroundTruth := 0.0 // findSimStateTime(b.flightStateGroundTruth, "TOUCHDOWN")
	benchLogger.Warn("Touchdown time ground truth determination skipped: findSimStateTime needs verification")

	// --- Calculate Actual Values from SIMULATION Data --- 
	actualApogeeFromSim, _ := findSimApogee(actualMotionData)
	actualMaxVFromSim, _ := findSimMaxVelocity(actualMotionData)
	actualLiftoffTimeFromSim := findSimEventTime(actualEventsData, "Liftoff")            
	actualApogeeEventTimeFromSim := findSimEventTime(actualEventsData, "ApogeeDetected") 
	actualTouchdownTimeFromSim := 0.0 
	benchLogger.Warn("Touchdown time comparison skipped: Unknown state name in simulation output CSV")

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
