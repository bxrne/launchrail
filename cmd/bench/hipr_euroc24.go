package main

import (
	"fmt"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/logger" // Import custom logger
)

// HiprEuroc24Benchmark implements the Benchmark interface for the specific dataset.
type HiprEuroc24Benchmark struct {
	name        string
	DataDir     string
	flightInfo  []FlightInfo
	eventInfo   []EventInfo
	flightState []FlightState
	// Add fields for other loaded data (Baro, IMU, etc.)
}

// NewHiprEuroc24Benchmark creates a new instance.
func NewHiprEuroc24Benchmark(config BenchmarkConfig) *HiprEuroc24Benchmark {
	return &HiprEuroc24Benchmark{
		name: "hipr-euroc24",
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

// LoadData loads all necessary CSV files for this benchmark.
func (b *HiprEuroc24Benchmark) LoadData(benchDataDir string) error {
	var err error
	benchLogger := logger.GetLogger("info") // Get logger instance

	benchLogger.Info("Loading data for benchmark", "name", b.Name(), "sourceDir", benchDataDir)

	basePath := filepath.Join(benchDataDir, b.Name()) // Use benchmark name to find subdir

	flightInfoPath := filepath.Join(basePath, "fl001 - flight_info_processed.csv")
	b.flightInfo, err = LoadFlightInfo(flightInfoPath)
	if err != nil {
		return fmt.Errorf("failed to load flight info: %w", err)
	}
	benchLogger.Info("Loaded flight info", "count", len(b.flightInfo), "file", flightInfoPath)

	eventInfoPath := filepath.Join(basePath, "fl001 - event_info_processed.csv")
	b.eventInfo, err = LoadEventInfo(eventInfoPath)
	if err != nil {
		return fmt.Errorf("failed to load event info: %w", err)
	}
	benchLogger.Info("Loaded event info", "count", len(b.eventInfo), "file", eventInfoPath)

	flightStatePath := filepath.Join(basePath, "fl001 - flight_states_processed.csv")
	b.flightState, err = LoadFlightStates(flightStatePath)
	if err != nil {
		return fmt.Errorf("failed to load flight states: %w", err)
	}
	benchLogger.Info("Loaded flight states", "count", len(b.flightState), "file", flightStatePath)

	return nil
}

// Run performs the comparison between benchmark data and simulated data.
func (b *HiprEuroc24Benchmark) Run(simData interface{}) ([]BenchmarkResult, error) {
	results := []BenchmarkResult{}
	benchLogger := logger.GetLogger("info") // Get logger instance

	benchLogger.Info("Running comparison...")

	// --- Define Ground Truth Expected Values for hipr-euroc24 ---
	// These values represent the known correct results for this dataset.
	// TODO: Potentially load these from a config file instead of hardcoding.
	const expectedApogeeGroundTruth = 7448.0
	const expectedMaxVGroundTruth = 1055.31
	const expectedLiftoffTimeGroundTruth = 0.0
	const expectedApogeeTimeGroundTruth = 20.60 // Time of Apogee Event
	const expectedTouchdownTimeGroundTruth = 122.12 // Time of Touchdown State

	// --- Calculate Actual Values from Loaded Benchmark Data ---
	actualApogeeFromData, _ := b.findApogee()
	actualMaxVFromData, _ := b.findMaxVelocity()
	actualLiftoffTimeFromData := b.findEventTime("EV_LIFTOFF")
	actualApogeeEventTimeFromData := b.findEventTime("EV_APOGEE")
	actualTouchdownTimeFromData := b.findStateTime("TOUCHDOWN")

	// --- Perform Comparisons ---

	// 1. Apogee Comparison
	apogeeTolerance := 0.05 // 5% tolerance
	results = append(results, compareFloat("Apogee", "Compare apogee height", expectedApogeeGroundTruth, actualApogeeFromData, apogeeTolerance))

	// 2. Max Velocity Comparison
	maxVTolerance := 0.05 // 5% tolerance
	results = append(results, compareFloat("Max Velocity", "Compare maximum velocity", expectedMaxVGroundTruth, actualMaxVFromData, maxVTolerance))

	// 3. Event Timing Comparisons (Example)
	liftoffTolerance := 0.1 // 100ms tolerance
	results = append(results, compareFloat("Liftoff Time", "Compare liftoff event time", expectedLiftoffTimeGroundTruth, actualLiftoffTimeFromData, liftoffTolerance))

	apogeeTimeTolerance := 0.5 // 500ms tolerance
	// Note: findEventTime returns the *first* occurrence.
	results = append(results, compareFloat("Apogee Event Time", "Compare apogee event time", expectedApogeeTimeGroundTruth, actualApogeeEventTimeFromData, apogeeTimeTolerance))

	touchdownTolerance := 1.0 // 1s tolerance
	results = append(results, compareFloat("Touchdown Time", "Compare touchdown state time", expectedTouchdownTimeGroundTruth, actualTouchdownTimeFromData, touchdownTolerance))

	// TODO: Add more comparisons (e.g., trajectory matching, other event times)

	benchLogger.Info("Comparison complete.")
	return results, nil
}

// --- Helper methods for HiprEuroc24Benchmark ---

func (b *HiprEuroc24Benchmark) findApogee() (maxHeight float64, timestamp float64) {
	if len(b.flightInfo) == 0 {
		return 0, 0
	}
	maxHeight = b.flightInfo[0].Height
	timestamp = b.flightInfo[0].Timestamp
	for _, p := range b.flightInfo {
		if p.Height > maxHeight {
			maxHeight = p.Height
			timestamp = p.Timestamp
		}
	}
	return maxHeight, timestamp
}

func (b *HiprEuroc24Benchmark) findMaxVelocity() (maxVelocity float64, timestamp float64) {
	if len(b.flightInfo) == 0 {
		return 0, 0
	}
	maxVelocity = b.flightInfo[0].Velocity
	timestamp = b.flightInfo[0].Timestamp
	for _, p := range b.flightInfo {
		if p.Velocity > maxVelocity {
			maxVelocity = p.Velocity
			timestamp = p.Timestamp
		}
	}
	return maxVelocity, timestamp
}

func (b *HiprEuroc24Benchmark) findEventTime(eventName string) float64 {
	for _, e := range b.eventInfo {
		if e.Event == eventName {
			return e.Timestamp // Return first occurrence
		}
	}
	return -1 // Indicate not found
}

func (b *HiprEuroc24Benchmark) findStateTime(stateName string) float64 {
	for _, s := range b.flightState {
		if s.State == stateName {
			return s.Timestamp // Return first occurrence
		}
	}
	return -1 // Indicate not found
}
