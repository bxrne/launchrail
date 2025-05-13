package reporting

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/zerodha/logf"
	"golang.org/x/net/html"
)

// MotionMetrics holds summary statistics about the rocket's motion during flight.
type MotionMetrics struct {
	TimeAtApogee             float64 `json:"time_at_apogee" yaml:"time_at_apogee"`                                   // Time of apogee from liftoff [s] (sensor data based)
	FlightTime               float64 `json:"flight_time" yaml:"flight_time"`                                         // Total flight time from liftoff to landing [s] (event based)
	BurnoutTime              float64 `json:"burnout_time" yaml:"burnout_time"`                                       // Time of motor burnout from liftoff [s]
	MaxAltitudeAGL           float64 `json:"max_altitude_agl" yaml:"max_altitude_agl"`                               // Max altitude above ground level [m]
	MaxSpeed                 float64 `json:"max_speed" yaml:"max_speed"`                                             // Max speed achieved during flight [m/s]
	MaxAcceleration          float64 `json:"max_acceleration" yaml:"max_acceleration"`                               // Max acceleration achieved (positive magnitude) [m/s^2]
	RailExitVelocity         float64 `json:"rail_exit_velocity" yaml:"rail_exit_velocity"`                           // Speed at launch rail clearance [m/s]
	LaunchRailClearanceTime  float64 `json:"launch_rail_clearance_time" yaml:"launch_rail_clearance_time"`           // Time at launch rail clearance [s]
	BurnoutAltitude          float64 `json:"burnout_altitude" yaml:"burnout_altitude"`                               // Altitude at motor burnout AGL [m]
	TimeToApogee             float64 `json:"time_to_apogee" yaml:"time_to_apogee"`                                   // Time from launch to apogee [s]
	LandingSpeed             float64 `json:"landing_speed" yaml:"landing_speed"`                                     // Speed at landing [m/s]
	DescentTime              float64 `json:"descent_time" yaml:"descent_time"`                                       // Time from apogee to landing [s]
	CoastToApogeeTime        float64 `json:"coast_to_apogee_time" yaml:"coast_to_apogee_time"`                       // Time from burnout to apogee [s]
	LaunchStabilityMetric    float64 `json:"launch_stability_metric" yaml:"launch_stability_metric"`                 // Barrowman stability margin (calibers)
	LaunchRailExitMachNumber float64 `json:"launch_rail_exit_mach_number" yaml:"launch_rail_exit_mach_number"`       // Mach number at launch rail exit
	AverageDescentSpeed      float64 `json:"average_descent_speed,omitempty" yaml:"average_descent_speed,omitempty"` // Average speed during descent [m/s]

	// Additional optional metrics
	MaxAltitudeASL   float64 `json:"max_altitude_asl,omitempty" yaml:"max_altitude_asl,omitempty"`   // Max altitude above sea level [m]
	TerminalVelocity float64 `json:"terminal_velocity,omitempty" yaml:"terminal_velocity,omitempty"` // Estimated terminal velocity during descent [m/s]

	// Error field for reporting issues during calculation
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// MotorSummaryData holds key performance indicators for the rocket motor.
type MotorSummaryData struct {
	BurnTime          float64
	MaxThrust         float64 // same as PeakThrust but with standardized name
	AvgThrust         float64 // same as AverageThrust but with standardized name
	TotalImpulse      float64
	SpecificImpulse   float64
	ThrustEfficiency  float64
	MotorClass        string
	MotorManufacturer string
	PropellantMass    float64
}

// ParachuteSummaryData holds summary statistics for parachute performance.
type ParachuteSummaryData struct {
	DeploymentTime     float64
	DeploymentAltitude float64
	DeploymentVelocity float64
	DragCoefficient    float64
	DescentRate        float64
	OpeningForce       float64
	Diameter           float64
	ParachuteType      string
}

// PhaseSummaryData holds summary statistics for flight phases.
type PhaseSummaryData struct {
	ApogeeTimeSec float64
	MaxAltitudeM  float64
	// Add other relevant metrics
}

// EventSummary provides a concise summary of a flight event.
type EventSummary struct {
	Time     float64
	Name     string // Changed from Event to Name for consistency
	Altitude float64
	Velocity float64
	Details  string // Optional additional details
}

// ReportSummary aggregates all summary statistics for the report.
type ReportSummary struct {
	MotionMetrics    MotionMetrics
	MotorSummary     MotorSummaryData
	ParachuteSummary ParachuteSummaryData
	PhaseSummary     PhaseSummaryData
	EventsTimeline   []EventSummary // e.g., liftoff, burnout, apogee, parachute deploy
	// Add other summary sections as needed
}

// PlotInfo stores information about a generated plot.
type PlotInfo struct {
	Title    string
	Filename string // Relative path to the plot image in the assets directory
	Type     string // e.g., "altitude_vs_time", "velocity_vs_time"
}

// PlotSimRecord represents a single row of parsed simulation data, typically motion data.
// Using a map allows flexibility with varying CSV headers.
// Values can be float64 (for numeric data) or string (for non-numeric or unconverted data).
type PlotSimRecord map[string]interface{}

// StageData holds data for a single stage of the rocket.
type StageData struct {
	Name        string
	MassKg      float64
	BurnTimeSec float64
	// Add other relevant metrics
}

// RecoverySystemData holds data for a recovery system.
type RecoverySystemData struct {
	Type        string
	Deployment  float64
	DescentRate float64
	// Add other relevant metrics
}

// LaunchRailData holds data about launch rail performance
type LaunchRailData struct {
	Length            float64
	Angle             float64
	DepartureVelocity float64
	MaxForce          float64
	DepartureTime     float64
	StabilityMargin   float64
}

// ForcesAndMomentsData holds aerodynamic force and moment data
type ForcesAndMomentsData struct {
	MaxAngleOfAttack   float64
	MaxNormalForce     float64
	MaxAxialForce      float64
	MaxRollRate        float64
	MaxPitchMoment     float64
	MaxDynamicPressure float64
	CenterOfPressure   float64
	CenterOfGravity    float64
	StabilityMargin    float64
}

// WeatherData holds atmospheric conditions data
type WeatherData struct {
	Latitude              float64 // Added
	Longitude             float64 // Added
	ElevationAMSL         float64 // Added
	WindSpeed             float64
	WindDirection         float64
	WindDirectionCardinal string
	Temperature           float64
	Pressure              float64
	Density               float64
	Humidity              float64
}

// Note: EventSummary is defined above with the required fields for the FRR report

// motionPoint is a helper struct for processing motion data rows.
type motionPoint struct {
	Time   float64
	AltAGL float64
	AltASL float64
	VelTot float64
	AccTot float64
	VelV   float64
}

// func populateDefaultValues(rData *ReportData, appCfg *config.Config) {
// 	// Set default values for LaunchRail
// 	rData.LaunchRail = LaunchRailData{
// 		Length:            2.0,  // Default 2m rail
// 		Angle:             5.0,  // Default 5 degree launch angle
// 		DepartureVelocity: 15.0, // Typical minimum safe velocity
// 		MaxForce:          50.0, // Placeholder
// 		DepartureTime:     0.5,  // Placeholder
// 		StabilityMargin:   1.5,  // Typical calibers for stability
// 	}

// 	// Set default values for MotorSummary (additional fields)
// 	rData.MotorSummary.SpecificImpulse = 200.0 // Typical amateur motor Isp
// 	rData.MotorSummary.ThrustEfficiency = 95.0 // Percentage
// 	if rData.MotorName != "" {
// 		// Basic parsing of motor name to extract motor class
// 		if len(rData.MotorName) >= 1 {
// 			rData.MotorSummary.MotorClass = string(rData.MotorName[0])
// 		}
// 	}

// 	// Set default values for ParachuteSummary (additional fields)
// 	rData.ParachuteSummary.DeploymentAltitude = rData.MotionMetrics.MaxAltitudeAGL * 0.9 // 90% of apogee
// 	rData.ParachuteSummary.DeploymentVelocity = 20.0
// 	rData.ParachuteSummary.DragCoefficient = 1.5      // Typical parachute drag coefficient
// 	rData.ParachuteSummary.OpeningForce = 100.0       // Placeholder
// 	rData.ParachuteSummary.Diameter = 0.8             // Default diameter in meters
// 	rData.ParachuteSummary.ParachuteType = "Toroidal" // Default type

// 	// Set default values for Weather
// 	rData.Weather = WeatherData{
// 		Latitude:              appCfg.Engine.Options.Launchsite.Latitude, // Added
// 		Longitude:             appCfg.Engine.Options.Launchsite.Longitude, // Added
// 		ElevationAMSL:         appCfg.Engine.Options.Launchsite.Altitude, // Added
// 		WindSpeed:             3.0,     // Light breeze
// 		WindDirection:         45.0,    // NE
// 		WindDirectionCardinal: "NE",    // Cardinal direction
// 		Temperature:           20.0,    // 20°C
// 		Pressure:              1013.25, // Standard pressure
// 		Density:               1.225,   // Sea level air density
// 		Humidity:              50.0,    // 50%
// 	}

// 	// Set default values for ForcesAndMoments
// 	rData.ForcesAndMoments = ForcesAndMomentsData{
// 		MaxAngleOfAttack:   5.0,    // Degrees
// 		MaxNormalForce:     200.0,  // Newtons
// 		MaxAxialForce:      150.0,  // Newtons
// 		MaxRollRate:        10.0,   // Degrees/second
// 		MaxPitchMoment:     2.0,    // Nm
// 		MaxDynamicPressure: 5000.0, // Pascals
// 		CenterOfPressure:   0.8,    // m from nose
// 		CenterOfGravity:    0.6,    // m from nose
// 		StabilityMargin:    1.5,    // Calibers
// 	}

// 	// Ensure MotionMetrics fields are populated for report template
// 	if rData.MotionMetrics != nil {
// 		// Default/example values if not set elsewhere
// 		if rData.MotionMetrics.MaxAltitudeAGL == 0 {
// 			rData.MotionMetrics.MaxAltitudeAGL = 500.0
// 		}
// 		if rData.MotionMetrics.MaxSpeed == 0 {
// 			rData.MotionMetrics.MaxSpeed = 200.0
// 		}
// 		if rData.MotionMetrics.MaxAcceleration == 0 {
// 			rData.MotionMetrics.MaxAcceleration = 100.0
// 		}
// 		if rData.MotionMetrics.FlightTime == 0 {
// 			rData.MotionMetrics.FlightTime = 120.0
// 		}
// 		if rData.MotionMetrics.LandingSpeed == 0 {
// 			rData.MotionMetrics.LandingSpeed = 5.0
// 		}

// 		// Populate other motion metrics with reasonable values
// 		rData.MotionMetrics.MaxAltitudeASL = rData.MotionMetrics.MaxAltitudeAGL * 1.1
// 	}

// 	// Add some sample events if none exist
// 	if len(rData.AllEvents) == 0 {
// 		rData.AllEvents = []EventSummary{
// 			{Time: 0.0, Name: "Launch", Altitude: 0.0, Velocity: 0.0, Details: "Rocket leaves the launch pad"},
// 			{Time: 2.0, Name: "Motor Burnout", Altitude: 200.0, Velocity: 150.0, Details: "Motor has consumed all propellant"},
// 			{Time: 15.0, Name: "Apogee", Altitude: rData.MotionMetrics.MaxAltitudeAGL, Velocity: 0.0, Details: "Maximum altitude reached"},
// 			{Time: 15.1, Name: "Parachute Deployment", Altitude: rData.MotionMetrics.MaxAltitudeAGL - 5.0, Velocity: 5.0, Details: "Main parachute deployed"},
// 			{Time: rData.MotionMetrics.FlightTime, Name: "Landing", Altitude: 0.0, Velocity: rData.MotionMetrics.LandingSpeed, Details: "Touchdown"},
// 		}
// 	}
// }

// parseSimData parses raw CSV data into a slice of PlotSimRecord and headers.
// It attempts to convert numeric fields to float64, otherwise stores as string.
func parseSimData(log *logf.Logger, csvData []byte, delimiter string) ([]*PlotSimRecord, []string, error) {
	if log == nil {
		return nil, nil, fmt.Errorf("parseSimData called with nil logger")
	}

	log.Debug("Parsing sim data", "length_bytes", len(csvData))
	if len(csvData) == 0 {
		log.Warn("CSV data is empty.")
		return []*PlotSimRecord{}, []string{}, nil
	}

	reader := csv.NewReader(bytes.NewReader(csvData))
	if len(delimiter) == 0 {
		delimiter = ","
	}
	reader.Comma = rune(delimiter[0])
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err == io.EOF {
		log.Warn("CSV data is empty or contains only EOF.")
		return []*PlotSimRecord{}, []string{}, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}
	log.Debug("CSV Headers", "headers", headers)

	reader.FieldsPerRecord = len(headers)

	var records []*PlotSimRecord
	rowIndex := 0
	for {
		row, err := reader.Read()
		rowIndex++
		if err == io.EOF {
			break
		}
		if err != nil {
			if parseErr, ok := err.(*csv.ParseError); ok && parseErr.Err == csv.ErrFieldCount {
				log.Warn("Skipping row due to mismatched column count", "row_index", rowIndex, "expected_cols", len(headers), "actual_cols", len(row), "row_data", row, "error", err)
				continue
			}
			log.Warn("Error reading CSV row", "row_index", rowIndex, "row_data", row, "error", err)
			continue
		}

		record := make(PlotSimRecord)
		for i, header := range headers {
			if i < len(row) {
				rawValue := row[i]
				if valFloat, errFloat := strconv.ParseFloat(rawValue, 64); errFloat == nil {
					record[header] = valFloat
				} else {
					record[header] = rawValue
				}
			} else {
				record[header] = ""
				log.Warn("Row shorter than headers", "row_index", rowIndex, "header_missing", header)
			}
		}
		records = append(records, &record)
	}

	if len(records) == 0 && len(headers) > 0 {
		log.Warn("CSV data contained headers but no data rows.")
	}

	log.Debug("Parsed data records", "num_records", len(records), "num_headers", len(headers))
	return records, headers, nil
}

// HandlerRecordManager defines the interface required for record management in reports.
type HandlerRecordManager interface {
	GetRecord(hash string) (*storage.Record, error)
	GetStorageDir() string
}

// calculateMotorSummary computes summary statistics from motor telemetry data.
// It assumes 'Time (s)' and 'Thrust (N)' are present in the headers.
// motorPoint represents a single data point of motor thrust over time
type motorPoint struct {
	Time   float64
	Thrust float64
}

// findMotorDataIndices finds the indices for time and thrust columns in the motor headers
func findMotorDataIndices(motorHeaders []string, log *logf.Logger) (timeIdx, thrustIdx int) {
	log.Debug("calculateMotorSummary: processing with headers", "headers", motorHeaders)

	timeIdx, thrustIdx = -1, -1
	for i, header := range motorHeaders {
		switch strings.ToLower(strings.TrimSpace(header)) {
		case "time (s)":
			timeIdx = i
		case "time":
			if timeIdx == -1 {
				timeIdx = i
			}
		case "thrust (n)":
			thrustIdx = i
		case "thrust":
			if thrustIdx == -1 {
				thrustIdx = i
			}
		}
	}
	if timeIdx == -1 || thrustIdx == -1 {
		log.Warn("calculateMotorSummary: Required columns not found.", "searched_time", "'Time (s)' or 'Time'", "searched_thrust", "'Thrust (N)' or 'Thrust'", "received_headers", motorHeaders, "time_col_found", timeIdx != -1, "thrust_col_found", thrustIdx != -1)
	}
	return timeIdx, thrustIdx
}

// extractMotorPoints converts raw motor data records to a structured format
func extractMotorPoints(motorData []*PlotSimRecord, motorHeaders []string, timeIdx, thrustIdx int, log *logf.Logger) []motorPoint {
	points := make([]motorPoint, 0, len(motorData))

	for i, record := range motorData {
		timeVal, okTime := (*record)[motorHeaders[timeIdx]].(float64)
		thrustVal, okThrust := (*record)[motorHeaders[thrustIdx]].(float64)

		if !okTime || !okThrust {
			log.Warn("Failed to parse time or thrust value from motor data record", "recordIndex", i, "record", record)
			continue // Skip this record
		}
		points = append(points, motorPoint{Time: timeVal, Thrust: thrustVal})
	}

	// Sort points by time to ensure correct calculation order
	if len(points) > 1 {
		sort.Slice(points, func(i, j int) bool {
			return points[i].Time < points[j].Time
		})
	}

	return points
}

// calculateThrustMetrics computes max thrust, burn time, and other thrust-related metrics
func calculateThrustMetrics(points []motorPoint) (maxThrust, burnStartTime, burnEndTime float64) {
	burnStartTime = -1

	for _, p := range points {
		// Track maximum thrust
		if p.Thrust > maxThrust {
			maxThrust = p.Thrust
		}

		// Burn time detection (first to last point with thrust > 0.01 N)
		if p.Thrust > 0.01 { // Threshold to consider motor active
			if burnStartTime == -1 {
				burnStartTime = p.Time
			}
			burnEndTime = p.Time
		}
	}

	return maxThrust, burnStartTime, burnEndTime
}

// calculateImpulse computes total impulse using the trapezoidal rule
func calculateImpulse(points []motorPoint) float64 {
	var totalImpulse float64
	var lastTime, lastThrust float64

	for i, p := range points {
		// Total Impulse (Trapezoidal rule)
		if i > 0 {
			dt := p.Time - lastTime
			if dt > 0 { // Avoid division by zero or negative dt if data is messy
				totalImpulse += (p.Thrust + lastThrust) * dt / 2.0
			}
		}
		lastTime = p.Time
		lastThrust = p.Thrust
	}

	return totalImpulse
}

// calculateMotorSummary computes summary statistics from motor thrust data
func calculateMotorSummary(motorData []*PlotSimRecord, motorHeaders []string, log *logf.Logger) MotorSummaryData {
	summary := MotorSummaryData{}
	if len(motorData) == 0 {
		log.Warn("Motor data is empty, cannot calculate motor summary.")
		return summary
	}

	// Find required data column indices
	timeIdx, thrustIdx := findMotorDataIndices(motorHeaders, log)
	if timeIdx == -1 || thrustIdx == -1 {
		log.Warn("Required columns 'Time (s)'/'Time' or 'Thrust (N)'/'Thrust' not found in motor headers. Cannot calculate motor summary.", "headers", motorHeaders)
		return summary
	}

	// Extract and process data points
	points := extractMotorPoints(motorData, motorHeaders, timeIdx, thrustIdx, log)
	if len(points) < 2 { // Need at least two points for impulse/burn time calculation
		log.Warn("Not enough valid motor data points to calculate summary.")
		// Still, we might have a max thrust from a single point if thrust > 0
		if len(points) == 1 && points[0].Thrust > 0 {
			summary.MaxThrust = points[0].Thrust
			summary.BurnTime = 0 // Or some small epsilon if thrust is instantaneous
		}
		return summary
	}

	// Calculate key metrics
	maxThrust, burnStartTime, burnEndTime := calculateThrustMetrics(points)
	totalImpulse := calculateImpulse(points)

	// Populate summary
	summary.MaxThrust = maxThrust
	if burnStartTime != -1 && burnEndTime >= burnStartTime {
		summary.BurnTime = burnEndTime - burnStartTime
	} else {
		summary.BurnTime = 0
	}

	summary.TotalImpulse = totalImpulse
	if summary.BurnTime > 0 {
		summary.AvgThrust = totalImpulse / summary.BurnTime
	} else {
		summary.AvgThrust = 0 // Avoid division by zero
	}

	// TODO: Populate other fields like SpecificImpulse, MotorClass, etc., from config or other sources if available.

	log.Info("Calculated motor summary", "maxThrust", summary.MaxThrust, "totalImpulse", summary.TotalImpulse, "burnTime", summary.BurnTime, "avgThrust", summary.AvgThrust)
	return summary
}

// func findLandingSpeed(points []motionPoint, eventsData [][]string, touchdownIdx int) float64 {
// 	if touchdownIdx < 0 || len(eventsData) <= touchdownIdx || len(eventsData[touchdownIdx]) == 0 {
// 		return 0
// 	}

// 	touchdownTime, err := strconv.ParseFloat(eventsData[touchdownIdx][0], 64)
// 	if err != nil {
// 		return 0
// 	}

// 	// Find velocity closest to the touchdown time
// 	for i := len(points) - 1; i >= 0; i-- {
// 		// Look for velocity data within 0.1s of touchdown or the last valid velocity before touchdown
// 		if points[i].Time <= touchdownTime && points[i].VelTot >= 0 {
// 			return math.Abs(points[i].VelTot) // Speed magnitude at landing
// 		}
// 	}

// 	return 0
// }

// func calculateStabilityMetrics(railExitVelocity, launchRailLength float64) (machNumber, stabilityMetric float64) {
// 	if launchRailLength <= 0 || railExitVelocity <= 0 {
// 		return 0, 0
// 	}

// 	machNumber = railExitVelocity / 340.29                      // Approximate speed of sound at sea level
// 	stabilityMetric = railExitVelocity * 0.3 / launchRailLength // Simplified stability calculation

// 	return
// }

// CalculateMotionMetrics computes summary motion statistics from telemetry and event data.
func CalculateMotionMetrics(motionData []*PlotSimRecord, motionHeaders []string, eventsData [][]string, launchRailLength float64, log *logf.Logger) *MotionMetrics {
	log.Debug("CalculateMotionMetrics: processing with motion headers", "motion_headers", motionHeaders)

	metrics := &MotionMetrics{}

	if eventsData == nil {
		log.Warn("CalculateMotionMetrics: eventsData is nil. Event-dependent metrics will not be calculated or will use fallbacks if available.")
		// metrics.Error could be set here, or allow individual calculations to note missing data.
	}

	// 1. Find data column indices from motionHeaders
	timeIdx, altitudeIdx, velocityIdx, accelIdx := FindMotionDataIndices(motionHeaders)
	if timeIdx == -1 { // Time is essential and index must be valid
		log.Error("Essential 'Time' or 'Time (s)' column not found in motion data headers. Motion-point-based metrics will not be calculated.", "received_headers", motionHeaders)
		// Do not return early; allow event-based metrics to be calculated.
		// motionPoints will be empty if timeIdx is -1, and subsequent logic handles this.
	}

	// 2. Extract motion points from motionData
	motionPoints := ExtractMotionPoints(motionData, motionHeaders, timeIdx, altitudeIdx, velocityIdx, accelIdx, log)
	if len(motionPoints) == 0 && timeIdx != -1 { // Log this only if timeIdx was found but still no points (e.g. empty motionData)
		log.Warn("No motion points extracted from motionData. Point-based metrics may be inaccurate or unavailable.")
	}

	// Event-based metrics calculation (should proceed even if motionPoints is empty)
	var launchIdx, apogeeIdx, burnoutIdx, touchdownIdx, railExitIdx = -1, -1, -1, -1, -1
	if eventsData != nil {
		launchIdx = findEventIndex(eventsData, "Launch")
		burnoutIdx = findEventIndex(eventsData, "Motor Burnout") // Changed from "Burnout"
		apogeeIdx = findEventIndex(eventsData, "Apogee")
		touchdownIdx = findEventIndex(eventsData, "Touchdown")
		railExitIdx = findEventIndex(eventsData, "Rail Exit")
		log.Debug("Event indices found", "launch", launchIdx, "burnout", burnoutIdx, "apogee", apogeeIdx, "touchdown", touchdownIdx, "railExit", railExitIdx)
	} else {
		log.Info("CalculateMotionMetrics: Skipping event index searches as eventsData is nil.")
	}

	// Calculate metrics

	// 2. Flight Time (event-based)
	if eventsData != nil && launchIdx != -1 && touchdownIdx != -1 {
		metrics.FlightTime = CalculateFlightTime(eventsData, launchIdx, touchdownIdx)
		if metrics.FlightTime < 0 { // Check for error return from CalculateFlightTime
			log.Warn("Failed to calculate valid flight time from events.")
			metrics.FlightTime = 0 // Reset or handle as per desired error representation
		}
	} else {
		log.Warn("Cannot calculate FlightTime: events data missing, or Launch/Touchdown events not found.")
	}

	// 3. Time to Apogee (event-based)
	if eventsData != nil && launchIdx != -1 && apogeeIdx != -1 {
		apogeeEventTimeStr := eventsData[apogeeIdx][1] // Assuming time is column 1
		launchEventTimeStr := eventsData[launchIdx][1]
		apogeeEventTimeFloat, errApogee := strconv.ParseFloat(apogeeEventTimeStr, 64)
		launchEventTimeFloat, errLaunch := strconv.ParseFloat(launchEventTimeStr, 64)

		if errApogee == nil && errLaunch == nil {
			metrics.TimeToApogee = apogeeEventTimeFloat - launchEventTimeFloat
			metrics.TimeAtApogee = apogeeEventTimeFloat // Corrected: Use TimeAtApogee
		} else {
			log.Warn("Could not parse apogee or launch event times to calculate TimeToApogee.", "errApogee", errApogee, "errLaunch", errLaunch)
		}
	} else if len(motionPoints) > 0 {
		// Fallback for TimeToApogee if event-based calculation failed or was not possible
		currentMaxAlt := -1.0
		var apogeeTimeFromMotion float64 = -1
		for _, p := range motionPoints {
			if p.AltAGL > currentMaxAlt {
				currentMaxAlt = p.AltAGL
				apogeeTimeFromMotion = p.Time // Capture time of max altitude
			}
		}
		if currentMaxAlt != -1.0 {
			launchTimeFromMotion := 0.0
			if len(motionPoints) > 0 {
				launchTimeFromMotion = motionPoints[0].Time
				log.Warn("Using first motion point time as approximate launch time for TimeToApogee calculation from motion data.")
			}
			metrics.TimeToApogee = apogeeTimeFromMotion - launchTimeFromMotion
			metrics.TimeAtApogee = apogeeTimeFromMotion
		} else {
			log.Warn("Could not determine apogee altitude from motion points.")
		}
	} else {
		log.Warn("Cannot calculate TimeToApogee: events data missing, or Launch/Apogee events not found.")
	}

	// 4. Burnout Time (Time from launch to motor burnout, event-based)
	if eventsData != nil && launchIdx != -1 && burnoutIdx != -1 {
		burnoutEventTimeStr := eventsData[burnoutIdx][1] // Assuming time is column 1
		launchEventTimeStr := eventsData[launchIdx][1]
		burnoutEventTimeFloat, errBurnout := strconv.ParseFloat(burnoutEventTimeStr, 64)
		launchEventTimeFloat, errLaunch := strconv.ParseFloat(launchEventTimeStr, 64)
		if errBurnout == nil && errLaunch == nil {
			metrics.BurnoutTime = burnoutEventTimeFloat - launchEventTimeFloat
		} else {
			log.Warn("Could not parse burnout or launch event times to calculate BurnoutTime.", "errBurnout", errBurnout, "errLaunch", errLaunch)
		}
	} else {
		log.Warn("Cannot calculate BurnoutTime: events data missing, or Launch/Burnout events not found.")
	}

	log.Debug("Metrics before DescentTime calc", "FlightTime", metrics.FlightTime, "TimeToApogee", metrics.TimeToApogee)
	// 7. Descent Time (Time from apogee to landing/touchdown)
	// Requires FlightTime and TimeToApogee
	if metrics.FlightTime > 0 && metrics.TimeToApogee > 0 && metrics.FlightTime > metrics.TimeToApogee {
		metrics.DescentTime = metrics.FlightTime - metrics.TimeToApogee
	} else {
		log.Warn("Cannot calculate DescentTime: FlightTime or TimeToApogee is missing or invalid.")
	}

	log.Debug("Metrics before CoastToApogeeTime calc", "BurnoutTime", metrics.BurnoutTime, "TimeToApogee", metrics.TimeToApogee)
	// 8. Coast to Apogee Time (Time from burnout to apogee)
	// Requires BurnoutTime (preferably event-based) and TimeToApogee (preferably event-based or motion-based)
	if metrics.BurnoutTime > 0 && metrics.TimeToApogee > 0 && metrics.TimeToApogee > metrics.BurnoutTime {
		metrics.CoastToApogeeTime = metrics.TimeToApogee - metrics.BurnoutTime
	} else if metrics.BurnoutTime > 0 && metrics.TimeToApogee > 0 {
		log.Warn("Calculated TimeToApogee is not greater than BurnoutTime, CoastToApogeeTime may be invalid.", "timeToApogee", metrics.TimeToApogee, "burnoutTime", metrics.BurnoutTime)
	} else {
		log.Warn("Cannot calculate CoastToApogeeTime: BurnoutTime or TimeToApogee is missing or invalid.")
	}

	// 9. Burnout Altitude (Altitude at motor burnout)
	// Requires BurnoutTime (preferably event-based) and motion data
	if metrics.BurnoutTime > 0 {
		burnoutTimeFromLaunch := metrics.BurnoutTime // This is duration from launch
		actualBurnoutTimestamp := 0.0
		if eventsData != nil && launchIdx != -1 { // Get actual launch timestamp if possible
			launchEventTimeStr := eventsData[launchIdx][1]
			lt, err := strconv.ParseFloat(launchEventTimeStr, 64)
			if err == nil {
				actualBurnoutTimestamp = lt + burnoutTimeFromLaunch
			}
		} else if len(motionPoints) > 0 {
			// Fallback: assume first motion point time is launch time
			actualBurnoutTimestamp = motionPoints[0].Time + burnoutTimeFromLaunch
			log.Warn("Using first motion point time as approximate launch time for BurnoutAltitude calculation.")
		}

		if actualBurnoutTimestamp > 0 {
			closestPointToBurnout := FindClosestMotionPoint(motionPoints, actualBurnoutTimestamp)
			if closestPointToBurnout != nil {
				metrics.BurnoutAltitude = closestPointToBurnout.AltAGL
			} else {
				log.Warn("Could not find a motion point close to burnout time to determine BurnoutAltitude.")
			}
		} else {
			log.Warn("Cannot determine actual burnout timestamp for BurnoutAltitude calculation.")
		}
	} else {
		log.Warn("Cannot calculate BurnoutAltitude: BurnoutTime is missing or invalid.")
	}

	// 10. Landing Velocity/Speed (from motion data, at touchdown)
	// Requires touchdown event for time, then find closest motion point.
	if eventsData != nil && touchdownIdx != -1 {
		touchdownTimeStr := eventsData[touchdownIdx][1]
		touchdownTime, err := strconv.ParseFloat(touchdownTimeStr, 64)
		if err == nil {
			closestPointToTouchdown := FindClosestMotionPoint(motionPoints, touchdownTime)
			if closestPointToTouchdown != nil {
				metrics.LandingSpeed = closestPointToTouchdown.VelTot
			} else {
				log.Warn("Could not find a motion point close to touchdown time to determine LandingSpeed.")
			}
		} else {
			log.Warn("Could not parse touchdown event time to determine LandingSpeed.", "error", err)
		}
	} else if len(motionPoints) > 0 { // Fallback if event data is nil or touchdown event not found
		lastPoint := motionPoints[len(motionPoints)-1]
		if lastPoint.AltAGL <= 1.0 { // Consider 'near ground' if altitude is <= 1.0m
			metrics.LandingSpeed = lastPoint.VelTot
			log.Info("Calculated LandingSpeed using last motion point near ground level.", "velocity", metrics.LandingSpeed, "altAGL", lastPoint.AltAGL)
		} else {
			log.Warn("Cannot calculate LandingSpeed from motion data: Touchdown event not found/eventsData nil, and last motion point not near ground.", "lastPointAlt", lastPoint.AltAGL)
		}
	} else {
		log.Warn("Cannot calculate LandingSpeed: Touchdown event not found or eventsData is nil, and no motion points available.")
	}

	// If motionPoints are available, use them for point-based metrics which can be more precise
	// or serve as fallbacks.
	if len(motionPoints) > 0 {
		// 5. Apogee Altitude (from motion data if available, more precise)
		currentMaxAlt := -1.0
		var apogeeTimeFromMotion float64 = -1
		for _, p := range motionPoints {
			if p.AltAGL > currentMaxAlt {
				currentMaxAlt = p.AltAGL
				apogeeTimeFromMotion = p.Time // Capture time of max altitude
			}
		}
		if currentMaxAlt != -1.0 {
			metrics.MaxAltitudeAGL = currentMaxAlt            // Corrected: Use MaxAltitudeAGL
			if metrics.TimeToApogee == 0 && launchIdx != -1 { // If event-based TimeToApogee wasn't set, use motion data
				launchTimeFromMotion := 0.0
				if eventsData != nil && launchIdx != -1 { // get launch time from events if possible
					launchEventTimeStr := eventsData[launchIdx][1]
					lt, err := strconv.ParseFloat(launchEventTimeStr, 64)
					if err == nil {
						launchTimeFromMotion = lt
					}
				} else if len(motionPoints) > 0 {
					// Fallback: assume first motion point is close to launch if no event data for launch time
					// This is a rough approximation
					launchTimeFromMotion = motionPoints[0].Time
					log.Warn("Using first motion point time as approximate launch time for TimeToApogee calculation from motion data.")
				}
				metrics.TimeToApogee = apogeeTimeFromMotion - launchTimeFromMotion
			}
			if metrics.TimeAtApogee == 0 { // If event-based TimeAtApogee wasn't set
				metrics.TimeAtApogee = apogeeTimeFromMotion
			}
		} else {
			log.Warn("Could not determine apogee altitude from motion points.")
		}

		// 6. Max Speed & Max Acceleration (from motion data)
		currentMaxSpeed := -1.0
		currentMaxAccel := -1.0
		for _, p := range motionPoints {
			if p.VelTot > currentMaxSpeed {
				currentMaxSpeed = p.VelTot
			}
			if p.AccTot > currentMaxAccel {
				currentMaxAccel = p.AccTot
			}
		}
		if currentMaxSpeed != -1.0 {
			metrics.MaxSpeed = currentMaxSpeed
		}
		if currentMaxAccel != -1.0 {
			metrics.MaxAcceleration = currentMaxAccel
		}

		// 7. Rail Exit Velocity
		if eventsData != nil && railExitIdx != -1 {
			railExitTimeStr := eventsData[railExitIdx][1]
			railExitTime, err := strconv.ParseFloat(railExitTimeStr, 64)
			if err == nil {
				closestPoint := FindClosestMotionPoint(motionPoints, railExitTime)
				if closestPoint != nil {
					metrics.RailExitVelocity = closestPoint.VelTot
				}
			} else {
				log.Warn("Could not parse rail exit event time.", "error", err)
			}
		} else if launchRailLength > 0 { // Fallback using launchRailLength if Rail Exit event is not available
			for _, p := range motionPoints {
				if p.AltAGL >= launchRailLength { // Assuming AltAGL is altitude above ground
					metrics.RailExitVelocity = p.VelTot
					log.Debug("Calculated rail exit velocity based on launch rail length", "velocity", metrics.RailExitVelocity, "altitude_agl", p.AltAGL, "rail_length", launchRailLength)
					break
				}
			}
			if metrics.RailExitVelocity == 0 {
				log.Warn("Could not determine rail exit velocity using launch rail length. Rocket might not have cleared the rail or data is insufficient.")
			}
		}

		// 8. Coast to Apogee Time (Time from burnout to apogee)
		// Requires BurnoutTime (preferably event-based) and TimeToApogee (preferably event-based or motion-based)
		if metrics.BurnoutTime > 0 && metrics.TimeToApogee > 0 && metrics.TimeToApogee > metrics.BurnoutTime {
			metrics.CoastToApogeeTime = metrics.TimeToApogee - metrics.BurnoutTime
		} else if metrics.BurnoutTime > 0 && metrics.TimeToApogee > 0 {
			log.Warn("Calculated TimeToApogee is not greater than BurnoutTime, CoastToApogeeTime may be invalid.", "timeToApogee", metrics.TimeToApogee, "burnoutTime", metrics.BurnoutTime)
		} else {
			log.Warn("Cannot calculate CoastToApogeeTime: BurnoutTime or TimeToApogee is missing or invalid.")
		}

		// 9. Burnout Altitude (Altitude at motor burnout)
		// Requires BurnoutTime (preferably event-based) and motion data
		if metrics.BurnoutTime > 0 {
			burnoutTimeFromLaunch := metrics.BurnoutTime // This is duration from launch
			actualBurnoutTimestamp := 0.0
			if eventsData != nil && launchIdx != -1 { // Get actual launch timestamp if possible
				launchEventTimeStr := eventsData[launchIdx][1]
				lt, err := strconv.ParseFloat(launchEventTimeStr, 64)
				if err == nil {
					actualBurnoutTimestamp = lt + burnoutTimeFromLaunch
				}
			} else if len(motionPoints) > 0 {
				// Fallback: assume first motion point time is launch time
				actualBurnoutTimestamp = motionPoints[0].Time + burnoutTimeFromLaunch
				log.Warn("Using first motion point time as approximate launch time for BurnoutAltitude calculation.")
			}

			if actualBurnoutTimestamp > 0 {
				closestPointToBurnout := FindClosestMotionPoint(motionPoints, actualBurnoutTimestamp)
				if closestPointToBurnout != nil {
					metrics.BurnoutAltitude = closestPointToBurnout.AltAGL
				} else {
					log.Warn("Could not find a motion point close to burnout time to determine BurnoutAltitude.")
				}
			} else {
				log.Warn("Cannot determine actual burnout timestamp for BurnoutAltitude calculation.")
			}
		} else {
			log.Warn("Cannot calculate BurnoutAltitude: BurnoutTime is missing or invalid.")
		}

		// 10. Landing Velocity/Speed (from motion data, at touchdown)
		// Requires touchdown event for time, then find closest motion point.
		if eventsData != nil && touchdownIdx != -1 {
			touchdownTimeStr := eventsData[touchdownIdx][1]
			touchdownTime, err := strconv.ParseFloat(touchdownTimeStr, 64)
			if err == nil {
				closestPointToTouchdown := FindClosestMotionPoint(motionPoints, touchdownTime)
				if closestPointToTouchdown != nil {
					metrics.LandingSpeed = closestPointToTouchdown.VelTot
				} else {
					log.Warn("Could not find a motion point close to touchdown time to determine LandingSpeed.")
				}
			} else {
				log.Warn("Could not parse touchdown event time to determine LandingSpeed.", "error", err)
			}
		} else if len(motionPoints) > 0 { // Fallback if event data is nil or touchdown event not found
			lastPoint := motionPoints[len(motionPoints)-1]
			if lastPoint.AltAGL <= 1.0 { // Consider 'near ground' if altitude is <= 1.0m
				metrics.LandingSpeed = lastPoint.VelTot
				log.Info("Calculated LandingSpeed using last motion point near ground level.", "velocity", metrics.LandingSpeed, "altAGL", lastPoint.AltAGL)
			} else {
				log.Warn("Cannot calculate LandingSpeed from motion data: Touchdown event not found/eventsData nil, and last motion point not near ground.", "lastPointAlt", lastPoint.AltAGL)
			}
		} else {
			log.Warn("Cannot calculate LandingSpeed: Touchdown event not found or eventsData is nil, and no motion points available.")
		}
	} else {
		log.Warn("Motion points are not available. Point-based metrics (ApogeeAltitude, MaxSpeed, MaxAcceleration, RailExitVelocity (motion-based), CoastToApogeeTime, BurnoutAltitude, LandingSpeed) cannot be calculated.")
	}

	log.Debug("Calculated motion metrics", "metrics", fmt.Sprintf("%+v", metrics))
	return metrics
}

// FindMotionDataIndices finds the indices of key motion data headers.
func FindMotionDataIndices(motionHeaders []string) (timeIdx, altitudeIdx, velocityIdx, accelIdx int) {
	timeIdx, altitudeIdx, velocityIdx, accelIdx = -1, -1, -1, -1 // Initialize to -1 (not found)
	for i, header := range motionHeaders {
		switch strings.ToLower(strings.TrimSpace(header)) {
		case "time":
			timeIdx = i
		case "time (s)":
			if timeIdx == -1 {
				timeIdx = i
			}
		case "altitude agl (m)":
			altitudeIdx = i
		case "total velocity (m/s)":
			velocityIdx = i
		case "total acceleration (m/s^2)":
			accelIdx = i
		}
	}
	return timeIdx, altitudeIdx, velocityIdx, accelIdx
}

// FindFlightEvents processes eventsData to find indices for key flight events.
func FindFlightEvents(eventsData [][]string, log *logf.Logger) (launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx int) {
	launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx = -1, -1, -1, -1, -1

	if len(eventsData) < 2 { // Need at least a header and one data row
		log.Warn("eventsData has insufficient rows to find flight events", "num_rows", len(eventsData))
		return
	}

	headerRow := eventsData[0]
	eventNameCol, eventTimeCol := -1, -1

	for i, header := range headerRow {
		normalizedHeader := strings.ToLower(strings.TrimSpace(header))
		if strings.Contains(normalizedHeader, "event") && strings.Contains(normalizedHeader, "name") {
			eventNameCol = i
		} else if strings.Contains(normalizedHeader, "time") {
			eventTimeCol = i // Assuming the first 'time' column is the event time
		}
	}

	if eventNameCol == -1 {
		log.Warn("Could not find 'Event Name' column in eventsData header", "header_row", headerRow)
		// Attempt to default if specific known structures are used, e.g., if only 2 columns, assume 0 is name, 1 is time.
		if len(headerRow) >= 1 {
			eventNameCol = 0
		} // Default to first column if not found
	}
	if eventTimeCol == -1 {
		log.Warn("Could not find 'Time' column in eventsData header", "header_row", headerRow)
		if len(headerRow) >= 2 {
			eventTimeCol = 1
		} // Default to second column if not found
	}

	// Ensure eventNameCol and eventTimeCol are valid before proceeding to avoid panic
	if eventNameCol == -1 || eventTimeCol == -1 || eventNameCol >= len(headerRow) || eventTimeCol >= len(headerRow) {
		log.Error("Critical: Event name or time column index is invalid after header processing.", "eventNameCol", eventNameCol, "eventTimeCol", eventTimeCol, "numHeaderCols", len(headerRow))
		return // Cannot proceed without valid column indices
	}

	log.Debug("Event header processing complete", "eventNameCol", eventNameCol, "eventTimeCol", eventTimeCol)

	for i, eventRow := range eventsData[1:] { // Iterate data rows, starting from index 1
		dataRowIndex := i + 1 // Actual index in original eventsData slice
		if len(eventRow) <= eventNameCol {
			log.Warn("Event data row too short for event name column", "row_index", dataRowIndex, "row_len", len(eventRow), "expected_col", eventNameCol)
			continue
		}

		eventName := strings.TrimSpace(eventRow[eventNameCol])
		switch {
		case strings.EqualFold(eventName, "Launch"):
			if launchIdx == -1 { // Take the first occurrence
				launchIdx = dataRowIndex
			}
		// Match "Burnout" or "Motor Burnout"
		case strings.Contains(strings.ToLower(eventName), "burnout"):
			if burnoutIdx == -1 {
				burnoutIdx = dataRowIndex
			}
		case strings.EqualFold(eventName, "Rail Exit"):
			if railExitIdx == -1 {
				railExitIdx = dataRowIndex
			}
		case strings.EqualFold(eventName, "Apogee"):
			if apogeeEventIdx == -1 {
				apogeeEventIdx = dataRowIndex
			}
		case strings.EqualFold(eventName, "Touchdown"):
			if touchdownIdx == -1 {
				touchdownIdx = dataRowIndex
			}
		}
	}

	if launchIdx == -1 {
		log.Warn("Launch event not found in eventsData")
	}
	if railExitIdx == -1 {
		log.Warn("Rail Exit event not found in eventsData")
	}
	if burnoutIdx == -1 {
		log.Warn("Burnout event not found in eventsData")
	}
	if apogeeEventIdx == -1 {
		log.Warn("Apogee event not found in eventsData")
	}
	if touchdownIdx == -1 {
		log.Warn("Touchdown event not found in eventsData")
	}

	return launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx
}

// ExtractMotionPoints extracts motion points from raw motion data.
func ExtractMotionPoints(motionData []*PlotSimRecord, motionHeaders []string, timeIdx, altitudeIdx, velocityIdx, accelIdx int, log *logf.Logger) []motionPoint {
	var points []motionPoint
	if timeIdx == -1 || timeIdx >= len(motionHeaders) { // Time is essential and index must be valid
		log.Warn("Time column header not found or index out of bounds, cannot extract motion points.")
		return points
	}
	timeHeader := motionHeaders[timeIdx]

	var altHeader, velHeader, accHeader string
	if altitudeIdx != -1 && altitudeIdx < len(motionHeaders) {
		altHeader = motionHeaders[altitudeIdx]
	}
	if velocityIdx != -1 && velocityIdx < len(motionHeaders) {
		velHeader = motionHeaders[velocityIdx]
	}
	if accelIdx != -1 && accelIdx < len(motionHeaders) {
		accHeader = motionHeaders[accelIdx]
	}

	for _, recordMapPtr := range motionData {
		if recordMapPtr == nil {
			continue
		}
		recordMap := *recordMapPtr
		mp := motionPoint{Time: -1, AltAGL: -1, VelTot: -1, AccTot: -1} // Initialize with invalid values

		// Parse Time
		if rawTime, ok := recordMap[timeHeader]; ok {
			t, err := GetFloat64Value(rawTime)
			if err == nil {
				mp.Time = t
			} else {
				log.Debug("Failed to parse time value", "header", timeHeader, "value", rawTime, "error", err)
				continue // Skip record if time is unparseable
			}
		} else {
			log.Debug("Time header not found in record map", "header", timeHeader)
			continue
		}

		// Parse Altitude AGL if header is valid
		if altHeader != "" {
			if rawAlt, ok := recordMap[altHeader]; ok {
				alt, err := GetFloat64Value(rawAlt)
				if err == nil {
					mp.AltAGL = alt
				} else {
					log.Debug("Failed to parse altitude value", "header", altHeader, "value", rawAlt, "error", err)
				}
			}
		}

		// Parse Total Velocity if header is valid
		if velHeader != "" {
			if rawVel, ok := recordMap[velHeader]; ok {
				vel, err := GetFloat64Value(rawVel)
				if err == nil {
					mp.VelTot = vel
				} else {
					log.Debug("Failed to parse velocity value", "header", velHeader, "value", rawVel, "error", err)
				}
			}
		}

		// Parse Total Acceleration if header is valid
		if accHeader != "" {
			if rawAcc, ok := recordMap[accHeader]; ok {
				acc, err := GetFloat64Value(rawAcc)
				if err == nil {
					mp.AccTot = acc
				} else {
					log.Debug("Failed to parse acceleration value", "header", accHeader, "value", rawAcc, "error", err)
				}
			}
		}
		points = append(points, mp)
	}
	return points
}

// GetFloat64Value tries to convert an interface{} to float64.
// It handles cases where the value might already be a float64 or a string representation of a float.
func GetFloat64Value(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("unsupported type for float64 conversion: %T", v)
	}
}

// FindPeakValues finds the peak values from motion points.
func FindPeakValues(points []motionPoint) (maxAltitude, maxSpeed, maxAcceleration, apogeeDataTime float64) {
	if len(points) == 0 {
		return 0, 0, 0, 0
	}

	// Initialize with the first point's values, assuming they are valid
	// or with a very small number if we want to ensure any positive value is greater.
	// For simplicity, let's start with the first point if available, or zero if not.
	maxAltitude = -1e9 // Using a very small number to ensure any actual altitude is greater
	maxSpeed = -1e9
	maxAcceleration = -1e9
	apogeeDataTime = 0

	firstValidAltitude := false

	for _, p := range points {
		if p.AltAGL > maxAltitude {
			maxAltitude = p.AltAGL
			apogeeDataTime = p.Time
			firstValidAltitude = true
		}
		if p.VelTot > maxSpeed {
			maxSpeed = p.VelTot
		}
		// Use math.Abs for acceleration if we're interested in magnitude,
		// or direct value if direction matters (e.g. max positive Gs)
		// Assuming we want the peak magnitude for 'max acceleration'.
		currentAccelMag := math.Abs(p.AccTot)
		if currentAccelMag > maxAcceleration {
			maxAcceleration = currentAccelMag
		}
	}

	// If no valid altitude was found (e.g., all were -1 or some initial sentinel),
	// reset maxAltitude and apogeeDataTime to 0 to indicate no data.
	if !firstValidAltitude {
		maxAltitude = 0
		apogeeDataTime = 0
	}
	// Same for speed and acceleration if they remained at their initial sentinel values
	if maxSpeed == -1e9 {
		maxSpeed = 0
	}
	if maxAcceleration == -1e9 {
		maxAcceleration = 0
	}

	return maxAltitude, maxSpeed, maxAcceleration, apogeeDataTime
}

// CalculateFlightTime calculates the flight time from events.
func CalculateFlightTime(eventsData [][]string, launchIdx, touchdownIdx int) float64 {
	if launchIdx == -1 || touchdownIdx == -1 || launchIdx >= len(eventsData) || touchdownIdx >= len(eventsData) {
		return 0 // Not enough data or invalid indices
	}

	launchEventRow := eventsData[launchIdx]
	touchdownEventRow := eventsData[touchdownIdx]

	// Event time is typically in the second column (index 1) of the eventData row
	const eventTimeColumnIndex = 1

	if len(launchEventRow) <= eventTimeColumnIndex || len(touchdownEventRow) <= eventTimeColumnIndex {
		return 0 // Event rows don't have enough columns for time
	}

	launchTimeStr := launchEventRow[eventTimeColumnIndex]
	touchdownTimeStr := touchdownEventRow[eventTimeColumnIndex]

	launchTime, errLaunch := strconv.ParseFloat(launchTimeStr, 64)
	if errLaunch != nil {
		// Log this error if a logger was available, for now, return 0
		return 0
	}

	touchdownTime, errTouchdown := strconv.ParseFloat(touchdownTimeStr, 64)
	if errTouchdown != nil {
		// Log this error if a logger was available, for now, return 0
		return 0
	}

	if touchdownTime < launchTime {
		return 0 // Touchdown before launch is not logical
	}

	return touchdownTime - launchTime
}

// FindClosestMotionPoint finds the motionPoint in a sorted list of points
// that is closest in time to the targetTime.
// Assumes points are sorted by Time.
func FindClosestMotionPoint(points []motionPoint, targetTime float64) *motionPoint {
	if len(points) == 0 {
		return nil
	}

	// Binary search to find the closest point or insertion point
	i := sort.Search(len(points), func(j int) bool { return points[j].Time >= targetTime })

	// Check boundaries and select the closest one
	if i == 0 {
		// targetTime is before or at the first point
		return &points[0]
	}
	if i == len(points) {
		// targetTime is after the last point
		return &points[len(points)-1]
	}

	// targetTime is between points[i-1] and points[i]
	// Compare which one is closer
	if (targetTime - points[i-1].Time) < (points[i].Time - targetTime) {
		return &points[i-1]
	}
	return &points[i]

}

// LoadSimulationData orchestrates loading all necessary data for a report.
func LoadSimulationData(recordID string, rm HandlerRecordManager, reportSpecificDir string, appCfg *config.Config) (*ReportData, error) {
	log := logger.GetLogger(appCfg.Setup.Logging.Level)
	if log == nil {
		return nil, fmt.Errorf("failed to initialize logger: logger.GetLogger returned nil")
	}

	log.Info("Preparing to load simulation data for report", "recordID", recordID, "reportSpecificDir", reportSpecificDir)

	simDataForReport := &storage.SimulationData{
		Motor: &storage.MotorData{},
	}

	engineConfigPath := filepath.Join(reportSpecificDir, "engine_config.json")
	engineConfigBytes, err := os.ReadFile(engineConfigPath)
	if err == nil {
		var recordEngineCfg config.Engine
		if unmarshalErr := json.Unmarshal(engineConfigBytes, &recordEngineCfg); unmarshalErr == nil {
			log.Info("Successfully loaded engine_config.json from record directory", "path", engineConfigPath)
			if recordEngineCfg.Options.MotorDesignation != "" {
				simDataForReport.Motor.Name = recordEngineCfg.Options.MotorDesignation
				log.Info("Motor name set from record's engine_config.json", "motorName", simDataForReport.Motor.Name)
			}
			// Basic parsing of motor name to extract motor class
			// Removed line: simDataForReport.Motor.Class = string(recordEngineCfg.Options.MotorDesignation[0])
		} else {
			log.Warn("Failed to unmarshal engine_config.json from record directory", "path", engineConfigPath, "error", unmarshalErr)
		}
	} else {
		log.Warn("engine_config.json not found in record directory or could not be read", "path", engineConfigPath, "error", err)
		// Motor name will rely on fallback logic within GenerateReportData or be empty if no appCfg fallback
	}

	return GenerateReportData(simDataForReport, recordID, rm, reportSpecificDir, appCfg)
}

// GenerateReportData orchestrates loading all necessary data for a report.
func GenerateReportData(simData *storage.SimulationData, recordID string, rm HandlerRecordManager, reportSpecificDir string, appCfg *config.Config) (*ReportData, error) {
	log := logger.GetLogger(appCfg.Setup.Logging.Level)
	if log == nil {
		return nil, fmt.Errorf("failed to initialize logger: logger.GetLogger returned nil")
	}

	log.Info("Generating simulation report data", "recordID", recordID)

	currentTime := time.Now().Format(time.RFC1123)

	rData := &ReportData{
		RecordID:         recordID,
		Version:          appCfg.Setup.App.Version,
		GeneratedTime:    currentTime,
		Plots:            make(map[string]string),
		Log:              log,
		MotionMetrics:    &MotionMetrics{},
		MotorSummary:     MotorSummaryData{},
		ParachuteSummary: ParachuteSummaryData{},
		PhaseSummary:     PhaseSummaryData{},
		LaunchRail:       LaunchRailData{},
		ForcesAndMoments: ForcesAndMomentsData{},
		Weather:          WeatherData{},
		Stages:           []StageData{},
		RecoverySystems:  []RecoverySystemData{},
		MotionData:       []*PlotSimRecord{},
		MotionHeaders:    []string{},
		EventsData:       [][]string{},
		GenerationDate:   currentTime,
		MotorData:        []*PlotSimRecord{},
		MotorHeaders:     []string{},
		Extensions:       make(map[string]interface{}),
		Assets:           make(map[string]string),
	}

	if simData != nil && simData.Motor != nil && simData.Motor.Name != "" {
		rData.MotorName = simData.Motor.Name
		log.Info("Motor name set from provided simData", "motorName", rData.MotorName)
	} else if appCfg.Engine.Options.MotorDesignation != "" {
		rData.MotorName = appCfg.Engine.Options.MotorDesignation
		log.Warn("Motor name from simData was empty, falling back to appCfg.Engine.Options.MotorDesignation", "motorName", rData.MotorName)
	} else {
		log.Warn("Motor name not available from simData or appCfg.Engine.Options.MotorDesignation. MotorName will be empty.")
	}

	rData.ConfigSummary = appCfg

	if appCfg.Engine.Options.OpenRocketFile != "" {
		rData.RocketName = filepath.Base(appCfg.Engine.Options.OpenRocketFile)
		rData.ReportTitle = fmt.Sprintf("Simulation Report for %s", rData.RocketName)
	} else {
		rData.ReportTitle = "Simulation Report"
	}

	// Populate WeatherData with LaunchSite information from appCfg
	rData.Weather.Latitude = appCfg.Engine.Options.Launchsite.Latitude
	rData.Weather.Longitude = appCfg.Engine.Options.Launchsite.Longitude
	rData.Weather.ElevationAMSL = appCfg.Engine.Options.Launchsite.Altitude

	// Keep existing weather data population if any, or ensure they are initialized
	// For now, assuming other weather fields (WindSpeed, Temp, etc.) are populated elsewhere or are fine as zero values if not set.
	// If they were meant to be sourced from appCfg.Engine.Options.Launchsite.Atmosphere or similar, that logic would go here.
	// Based on current WeatherData struct, fields like Temperature, Pressure, Humidity are part of it but not directly in Launchsite struct from config.
	// These might come from simulation results or a different part of config not yet examined for weather specifics.

	record, err := rm.GetRecord(recordID)
	if err != nil {
		log.Error("Failed to get record for report generation", "recordID", recordID, "error", err)
		return nil, fmt.Errorf("failed to get record %s: %w", recordID, err)
	}
	defer func() {
		if record.Motion != nil || record.Events != nil || record.Dynamics != nil {
			if closeErr := record.Close(); closeErr != nil {
				log.Warn("Failed to close record stores after generating report data", "recordID", recordID, "error", closeErr)
			}
		} else {
			log.Debug("Record has no active storage fields to close", "recordID", recordID)
		}
	}()

	motionFilePath := filepath.Join(reportSpecificDir, "motion.csv")
	if record.Motion != nil && record.Motion.GetFilePath() != "" {
		motionFilePath = record.Motion.GetFilePath()
	}
	motionDataCSV, err := os.ReadFile(motionFilePath)
	if err != nil {
		altMotionFilePath := filepath.Join(reportSpecificDir, "MOTION.csv")
		motionDataCSV, err = os.ReadFile(altMotionFilePath)
		if err != nil {
			log.Warn("Could not load motion data", "error", err)
		} else {
			motionFilePath = altMotionFilePath
		}
	}

	if err == nil {
		motionDataParsed, motionHeaders, parseErr := parseSimData(log, motionDataCSV, ",")
		if parseErr != nil {
			log.Warn("Error parsing motion data", "filename", motionFilePath, "error", parseErr)
		} else {
			rData.MotionData = motionDataParsed
			rData.MotionHeaders = motionHeaders
			if len(rData.MotionData) > 0 {
				firstRecord := rData.MotionData[0]
				massHeaderKey := getHeaderKey(rData.MotionHeaders, "Mass (kg)")
				if massHeaderKey != "" {
					if massVal, ok := (*firstRecord)[massHeaderKey].(float64); ok {
						rData.LiftoffMassKg = massVal
					}
				}
			}
		}
	}

	eventsFilePath := filepath.Join(reportSpecificDir, "events.csv")
	if record.Events != nil && record.Events.GetFilePath() != "" {
		eventsFilePath = record.Events.GetFilePath()
	}
	eventsDataCSV, err := os.ReadFile(eventsFilePath)
	if err != nil {
		altEventsFilePath := filepath.Join(reportSpecificDir, "EVENTS.csv")
		eventsDataCSV, err = os.ReadFile(altEventsFilePath)
		if err != nil {
			log.Warn("Could not load events data", "error", err)
		} else {
			eventsFilePath = altEventsFilePath
		}
	}

	if err == nil {
		reader := csv.NewReader(bytes.NewReader(eventsDataCSV))
		reader.Comma = ','
		rawEventsData, parseErr := reader.ReadAll()
		if parseErr != nil {
			log.Warn("Error parsing events data", "filename", eventsFilePath, "error", parseErr)
		} else {
			rData.EventsData = rawEventsData
		}
	}

	calculatedMetrics := CalculateMotionMetrics(rData.MotionData, rData.MotionHeaders, rData.EventsData, appCfg.Engine.Options.Launchrail.Length, log)
	rData.MotionMetrics = calculatedMetrics

	rData.MotorSummary = calculateMotorSummary(rData.MotionData, rData.MotionHeaders, log)

	plotKeys := []string{
		"altitude_vs_time", "velocity_vs_time", "acceleration_vs_time", "trajectory_3d",
		"landing_radius", "thrust_vs_time", "angle_of_attack", "forces_vs_time",
		"moments_vs_time", "roll_rate",
	}
	for _, key := range plotKeys {
		plotPath := fmt.Sprintf("%s.svg", key)
		rData.Plots[key] = plotPath
		rData.Assets[key] = plotPath
	}

	log.Info("Simulation data loading and basic processing complete", "RecordID", recordID)
	return rData, nil
}

// findEventIndex searches for an event by name in the eventsData and returns its index.
// Assumes event name is in the first column (index 0) of each event row.
func findEventIndex(eventsData [][]string, eventName string) int {
	if eventsData == nil {
		return -1
	}
	for i, event := range eventsData {
		if len(event) > 0 && strings.EqualFold(strings.TrimSpace(event[0]), eventName) {
			return i
		}
	}
	return -1 // Event not found
}

func getHeaderKey(headers []string, keyName string) string {
	for _, h := range headers {
		if strings.EqualFold(h, keyName) {
			return h
		}
	}
	return "" // Return empty if not found
}

// ConvertMarkdownToSimpleHTML converts a markdown string to a very basic HTML representation.
// This is a placeholder and should be replaced with a proper markdown parsing library for production use.
func ConvertMarkdownToSimpleHTML(mdContent string, recordID string) string {
	// Replace # Title with <h1>Title</h1>, ## Subtitle with <h2>Subtitle</h2> etc.
	// This is a very naive implementation.
	htmlOutput := "<!DOCTYPE html>\n"
	htmlOutput += "<html>\n"
	htmlOutput += "<head>\n"
	htmlOutput += fmt.Sprintf("<title>Report: %s</title>\n", recordID)
	// Basic styling could be added here if desired
	htmlOutput += "<style>body { font-family: sans-serif; margin: 20px; } h1, h2, h3 { color: #333; } pre { background-color: #f5f5f5; padding: 10px; border: 1px solid #ddd; overflow-x: auto; }</style>\n"
	htmlOutput += "</head>\n"
	htmlOutput += "<body>\n"
	htmlOutput += fmt.Sprintf("<h1>Simulation Report: %s</h1>\n", recordID)

	// Simple conversion for paragraphs (split by double newline)
	// and attempt to convert markdown headers, and code blocks.
	lines := strings.Split(strings.ReplaceAll(mdContent, "\r\n", "\n"), "\n\n")
	for _, paragraph := range lines {
		if strings.HasPrefix(paragraph, "```") {
			codeBlock := strings.TrimPrefix(paragraph, "```")
			codeBlock = strings.TrimSuffix(codeBlock, "```")
			codeBlock = strings.TrimSpace(codeBlock)
			// Escape HTML characters within the code block
			escapedCode := html.EscapeString(codeBlock)
			htmlOutput += "<pre><code>" + escapedCode + "</code></pre>\n"
		} else if strings.HasPrefix(paragraph, "### ") {
			htmlOutput += "<h3>" + html.EscapeString(strings.TrimPrefix(paragraph, "### ")) + "</h3>\n"
		} else if strings.HasPrefix(paragraph, "## ") {
			htmlOutput += "<h2>" + html.EscapeString(strings.TrimPrefix(paragraph, "## ")) + "</h2>\n"
		} else if strings.HasPrefix(paragraph, "# ") {
			htmlOutput += "<h1>" + html.EscapeString(strings.TrimPrefix(paragraph, "# ")) + "</h1>\n"
		} else if strings.TrimSpace(paragraph) != "" {
			htmlOutput += "<p>" + html.EscapeString(paragraph) + "</p>\n"
		}
	}

	htmlOutput += "</body>\n"
	htmlOutput += "</html>"
	return htmlOutput
}

// ReportData holds all data required to generate a report.
type ReportData struct {
	RecordID         string               `json:"record_id" yaml:"record_id"`
	Version          string               `json:"version" yaml:"version"`
	RocketName       string               `json:"rocket_name" yaml:"rocket_name"`
	MotorName        string               `json:"motor_name" yaml:"motor_name"`
	LiftoffMassKg    float64              `json:"liftoff_mass_kg" yaml:"liftoff_mass_kg"`
	GeneratedTime    string               `json:"generated_time" yaml:"generated_time"`
	ConfigSummary    *config.Config       `json:"config_summary" yaml:"config_summary"`
	Summary          ReportSummary        `json:"summary" yaml:"summary"`
	Plots            map[string]string    `json:"plots" yaml:"plots"`
	MotionMetrics    *MotionMetrics       `json:"motion_metrics" yaml:"motion_metrics"`
	MotorSummary     MotorSummaryData     `json:"motor_summary" yaml:"motor_summary"`
	ParachuteSummary ParachuteSummaryData `json:"parachute_summary" yaml:"parachute_summary"`
	PhaseSummary     PhaseSummaryData     `json:"phase_summary" yaml:"phase_summary"`
	LaunchRail       LaunchRailData       `json:"launch_rail" yaml:"launch_rail"`
	ForcesAndMoments ForcesAndMomentsData `json:"forces_and_moments" yaml:"forces_and_moments"`
	Weather          WeatherData          `json:"weather" yaml:"weather"`
	AllEvents        []EventSummary       `json:"all_events" yaml:"all_events"`
	Stages           []StageData          `json:"stages" yaml:"stages"`
	RecoverySystems  []RecoverySystemData `json:"recovery_systems" yaml:"recovery_systems"`
	MotionData       []*PlotSimRecord     `json:"motion_data" yaml:"motion_data"`
	MotionHeaders    []string             `json:"motion_headers" yaml:"motion_headers"`
	EventsData       [][]string           `json:"events_data" yaml:"events_data"`
	Log              *logf.Logger         `json:"-"` // Exclude logger from JSON
	ReportTitle      string               `json:"report_title" yaml:"report_title"`
	GenerationDate   string               `json:"generation_date" yaml:"generation_date"`
	MotorData        []*PlotSimRecord     `json:"motor_data" yaml:"motor_data"`
	MotorHeaders     []string             `json:"motor_headers" yaml:"motor_headers"`
	// Extended fields for templates and flexible data
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	// Collection of all assets (SVG plots, etc.)
	Assets map[string]string `json:"assets,omitempty"`
}
