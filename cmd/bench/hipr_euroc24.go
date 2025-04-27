package main

import (
	"fmt"
	"math"
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

	fmt.Println("Running comparison...")

	// --- Placeholder for extracting data from simData ---
	// This needs to be defined based on how simulation results are structured.
	// Example: simResults := simData.(map[string]float64) // Or a specific struct
	simApogee := 7448.0    // Placeholder value from dtl3.csv
	simMaxVelocity := 1055.31 // Placeholder value from dtl3.csv
	simLiftoffTime := 0.0     // Placeholder
	simApogeeTime := 20.60    // Placeholder
	simTouchdownTime := 122.12 // Placeholder

	// --- Perform Comparisons ---

	// 1. Apogee Comparison
	expectedApogee, _ := b.findApogee()
	apogeeTolerance := 0.05 // 5% tolerance
	results = append(results, compareFloat("Apogee", expectedApogee, simApogee, apogeeTolerance))

	// 2. Max Velocity Comparison
	expectedMaxV, _ := b.findMaxVelocity()
	maxVTolerance := 0.05 // 5% tolerance
	results = append(results, compareFloat("Max Velocity", expectedMaxV, simMaxVelocity, maxVTolerance))

	// 3. Event Timing Comparisons (Example)
	liftoffTolerance := 0.1 // 100ms tolerance
	expectedLiftoffTime := b.findEventTime("EV_LIFTOFF")
	results = append(results, compareFloat("Liftoff Time", expectedLiftoffTime, simLiftoffTime, liftoffTolerance))

	apogeeTimeTolerance := 0.5 // 500ms tolerance
	expectedApogeeEventTime := b.findEventTime("EV_APOGEE") // Note: This might need refinement if multiple apogee events exist
	results = append(results, compareFloat("Apogee Event Time", expectedApogeeEventTime, simApogeeTime, apogeeTimeTolerance))

	touchdownTolerance := 1.0 // 1s tolerance
	expectedTouchdownTime := b.findStateTime("TOUCHDOWN")
	results = append(results, compareFloat("Touchdown Time", expectedTouchdownTime, simTouchdownTime, touchdownTolerance))

	// TODO: Add more comparisons (e.g., trajectory matching, other event times)

	fmt.Println("Comparison complete.")
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

// compareFloat is a helper to create a BenchmarkResult for float comparison.
func compareFloat(metricName string, expected, actual, tolerancePercent float64) BenchmarkResult {
	diff := math.Abs(expected - actual)
	var toleranceValue float64

	// Handle expected value of zero (or close to zero) separately
	if math.Abs(expected) < 1e-9 { // Use epsilon comparison for float zero
		// If expected is zero, interpret tolerancePercent as the absolute allowed difference.
		toleranceValue = tolerancePercent // Treat as absolute tolerance
	} else {
		// Otherwise, calculate the tolerance value as a percentage of the expected value.
		toleranceValue = math.Abs(expected * tolerancePercent)
	}

	passed := diff <= toleranceValue

	return BenchmarkResult{
		Name:        metricName,
		Passed:      passed,
		Metric:      metricName,
		Expected:    expected,
		Actual:      actual,
		Difference:  diff,
		Tolerance:   toleranceValue,
		Description: fmt.Sprintf("%s: Expected=%.2f, Actual=%.2f, Diff=%.2f, Tolerance=%.2f (%.1f%% interpreted %s)",
			metricName,
			expected, actual, diff, toleranceValue,
			tolerancePercent*100,
			func() string { if math.Abs(expected) < 1e-9 { return "absolute" } else { return "relative" } }()),
	}
}
