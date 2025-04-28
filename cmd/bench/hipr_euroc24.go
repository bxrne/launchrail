package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bxrne/launchrail/internal/logger"
)

// HiprEuroc24Benchmark implements the Benchmark interface for the specific dataset.
type HiprEuroc24Benchmark struct {
	name string
	// Ground truth data loaded from CSVs
	flightInfoGroundTruth []FlightInfo // Renamed for clarity
	// eventInfoGroundTruth  []EventInfo   // Renamed for clarity
	// flightStateGroundTruth []FlightState // Renamed for clarity

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

	basePath := filepath.Join(benchDataDir, b.Name()) // Use benchmark name to find subdir

	flightInfoPath := filepath.Join(basePath, "fl001 - flight_info_processed.csv")
	b.flightInfoGroundTruth, err = LoadFlightInfo(flightInfoPath) // Load into renamed field
	if err != nil {
		return fmt.Errorf("failed to load ground truth flight info: %w", err)
	}
	benchLogger.Info("Loaded ground truth flight info", "count", len(b.flightInfoGroundTruth), "file", flightInfoPath)

	// eventInfoPath := filepath.Join(basePath, "fl001 - event_info_processed.csv")
	// b.eventInfoGroundTruth, err = LoadEventInfo(eventInfoPath)
	// if err != nil {
	// 	return fmt.Errorf("failed to load ground truth event info: %w", err)
	// }
	// benchLogger.Info("Loaded ground truth event info", "count", len(b.eventInfoGroundTruth), "file", eventInfoPath)

	// flightStatePath := filepath.Join(basePath, "fl001 - flight_states_processed.csv")
	// b.flightStateGroundTruth, err = LoadFlightStates(flightStatePath)
	// if err != nil {
	// 	return fmt.Errorf("failed to load ground truth flight states: %w", err)
	// }
	// benchLogger.Info("Loaded ground truth flight states", "count", len(b.flightStateGroundTruth), "file", flightStatePath)

	return nil
}

// Run performs the comparison between simulation results and ground truth data.
func (b *HiprEuroc24Benchmark) Run() ([]BenchmarkResult, error) {
	results := []BenchmarkResult{}
	benchLogger := logger.GetLogger("info") // Get logger instance

	benchLogger.Info("Running benchmark comparison", "name", b.name, "simulationRecord", b.config.SimRecordHash)

	// --- Load Simulation Record Data --- // NEW
	if b.config.RecordManager == nil {
		return nil, fmt.Errorf("RecordManager is nil in benchmark config for '%s'", b.name)
	}
	if b.config.SimRecordHash == "" {
		return nil, fmt.Errorf("SimRecordHash is empty in benchmark config for '%s'", b.name)
	}

	simRecord, err := b.config.RecordManager.GetRecord(b.config.SimRecordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get simulation record '%s': %w", b.config.SimRecordHash, err)
	}
	defer simRecord.Close()

	// Load and parse motion data from the simulation record storage
	headers, rawMotionData, err := simRecord.Motion.ReadHeadersAndData()
	if err != nil {
		return nil, fmt.Errorf("failed to read simulation motion data: %w", err)
	}

	simMotionData, err := parseMotionData(headers, rawMotionData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse simulation motion data: %w", err)
	}

	// --- Define Ground Truth Expected Values (from loaded CSVs) ---
	expectedApogeeGroundTruth, _ := b.findGroundTruthApogee()
	expectedMaxVGroundTruth, _ := b.findGroundTruthMaxVelocity()
	// expectedLiftoffTimeGroundTruth := b.findGroundTruthEventTime("EV_LIFTOFF") // Example if event data were loaded
	// expectedApogeeEventTimeGroundTruth := b.findGroundTruthEventTime("EV_APOGEE") // Example if event data were loaded
	// expectedTouchdownTimeGroundTruth := b.findGroundTruthStateTime("TOUCHDOWN") // Example if state data were loaded

	// --- Calculate Actual Values from SIMULATION Data --- // NEW
	actualApogeeFromSim, _ := findSimApogee(simMotionData)
	actualMaxVFromSim, _ := findSimMaxVelocity(simMotionData)
	// actualLiftoffTimeFromSim := findSimulationEventTime(telemetry, model.EventLiftoff) // Placeholder
	// actualApogeeEventTimeFromSim := findSimulationEventTime(telemetry, model.EventApogee) // Placeholder
	// actualTouchdownTimeFromSim := findSimulationStateTime(telemetry, model.StateTouchdown) // Placeholder

	// --- Perform Comparisons --- // UPDATED: Compare sim actual vs ground truth expected

	// 1. Apogee Comparison
	apogeeTolerance := 0.05 // 5% tolerance
	results = append(results, compareFloat("Apogee", "Compare simulation apogee height vs ground truth", expectedApogeeGroundTruth, actualApogeeFromSim, apogeeTolerance))

	// 2. Max Velocity Comparison
	maxVTolerance := 0.05 // 5% tolerance
	results = append(results, compareFloat("Max Velocity", "Compare simulation max velocity vs ground truth", expectedMaxVGroundTruth, actualMaxVFromSim, maxVTolerance))

	// 3. Event Timing Comparisons (Example - if implemented)
	// liftoffTolerance := 0.1 // 100ms tolerance
	// results = append(results, compareFloat("Liftoff Time", "Compare sim liftoff event time vs ground truth", expectedLiftoffTimeGroundTruth, actualLiftoffTimeFromSim, liftoffTolerance))

	// apogeeTimeTolerance := 0.5 // 500ms tolerance
	// results = append(results, compareFloat("Apogee Event Time", "Compare sim apogee event time vs ground truth", expectedApogeeEventTimeGroundTruth, actualApogeeEventTimeFromSim, apogeeTimeTolerance))

	// touchdownTolerance := 1.0 // 1s tolerance
	// results = append(results, compareFloat("Touchdown Time", "Compare sim touchdown state time vs ground truth", expectedTouchdownTimeGroundTruth, actualTouchdownTimeFromSim, touchdownTolerance))

	benchLogger.Info("Comparison complete.")
	return results, nil
}

// --- Helper methods for GROUND TRUTH data (from CSV) ---

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

/* // Example if event/state ground truth data loading were enabled
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
*/

// --- Helper methods for SIMULATION data (from storage.TelemetryData) --- // NEW

type parsedMotionData struct {
	time         []float64
	altitude     []float64
	velocity     []float64
	acceleration []float64
	thrust       []float64
}

// findSimApogee finds the maximum altitude from simulation data.
func findSimApogee(simData *parsedMotionData) (maxHeight float64, maxHeightTime float64) {
	if len(simData.altitude) == 0 {
		return 0, 0
	}
	maxHeight = simData.altitude[0]
	maxHeightTime = simData.time[0]
	for i := 1; i < len(simData.altitude); i++ {
		if simData.altitude[i] > maxHeight {
			maxHeight = simData.altitude[i]
			maxHeightTime = simData.time[i]
		}
	}
	return maxHeight, maxHeightTime
}

// findSimMaxVelocity finds the maximum velocity from simulation data.
func findSimMaxVelocity(simData *parsedMotionData) (maxVelocity float64, maxVelocityTime float64) {
	if len(simData.velocity) == 0 {
		return 0, 0
	}
	maxVelocity = simData.velocity[0]
	maxVelocityTime = simData.time[0]
	for i := 1; i < len(simData.velocity); i++ {
		if simData.velocity[i] > maxVelocity {
			maxVelocity = simData.velocity[i]
			maxVelocityTime = simData.time[i]
		}
	}
	return maxVelocity, maxVelocityTime
}

// Helper function to parse motion data from CSV strings
func parseMotionData(headers []string, data [][]string) (*parsedMotionData, error) {
	colIndices := make(map[string]int)
	for i, h := range headers {
		colIndices[strings.ToLower(strings.TrimSpace(h))] = i
	}

	requiredCols := []string{"time", "altitude", "velocity", "acceleration", "thrust"}
	for _, col := range requiredCols {
		if _, ok := colIndices[col]; !ok {
			return nil, fmt.Errorf("required motion column '%s' not found in headers", col)
		}
	}

	parsed := &parsedMotionData{
		time:         make([]float64, len(data)),
		altitude:     make([]float64, len(data)),
		velocity:     make([]float64, len(data)),
		acceleration: make([]float64, len(data)),
		thrust:       make([]float64, len(data)),
	}

	var err error
	for i, row := range data {
		parsed.time[i], err = strconv.ParseFloat(row[colIndices["time"]], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time at row %d: %v", i, err)
		}

		parsed.altitude[i], err = strconv.ParseFloat(row[colIndices["altitude"]], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse altitude at row %d: %v", i, err)
		}

		parsed.velocity[i], err = strconv.ParseFloat(row[colIndices["velocity"]], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse velocity at row %d: %v", i, err)
		}

		parsed.acceleration[i], err = strconv.ParseFloat(row[colIndices["acceleration"]], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse acceleration at row %d: %v", i, err)
		}

		parsed.thrust[i], err = strconv.ParseFloat(row[colIndices["thrust"]], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse thrust at row %d: %v", i, err)
		}
	}

	return parsed, nil
}
