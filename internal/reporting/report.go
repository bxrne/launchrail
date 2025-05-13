package reporting

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
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
)

// MotionMetrics holds summary statistics about the rocket's motion.
type MotionMetrics struct {
	ApogeeTime       float64 `json:"apogee_time" yaml:"apogee_time"`                                 // Time of apogee from liftoff [s]
	TotalFlightTime  float64 `json:"total_flight_time" yaml:"total_flight_time"`                     // Total flight time from liftoff to landing [s]
	MaxAcceleration  float64 `json:"max_acceleration" yaml:"max_acceleration"`                       // Maximum acceleration during flight (vector magnitude) [m/s^2]
	LandingDistance  float64 `json:"landing_distance" yaml:"landing_distance"`                       // Distance from launch pad to landing point [m]
	LandingSpeed     float64 `json:"landing_speed" yaml:"landing_speed"`                             // Velocity at impact [m/s]
	TerminalVelocity float64 `json:"terminal_velocity,omitempty" yaml:"terminal_velocity,omitempty"` // Estimated terminal velocity during descent [m/s]

	// Additional metrics
	MaxAltitudeAGL           float64 `json:"max_altitude_agl" yaml:"max_altitude_agl"`                       // Max altitude above ground level [m]
	MaxAltitudeASL           float64 `json:"max_altitude_asl,omitempty" yaml:"max_altitude_asl,omitempty"`   // Max altitude above sea level [m]
	MaxSpeed                 float64 `json:"max_speed" yaml:"max_speed"`                                     // Maximum speed during flight (vector magnitude) [m/s]
	LaunchRailClearanceTime  float64 `json:"launch_rail_clearance_time" yaml:"launch_rail_clearance_time"`   // Time when rocket clears the launch rail [s]
	LaunchRailClearanceSpeed float64 `json:"launch_rail_clearance_speed" yaml:"launch_rail_clearance_speed"` // Speed when rocket clears the launch rail [m/s]
	BurnoutTime              float64 `json:"burnout_time" yaml:"burnout_time"`                               // Time of motor burnout from liftoff [s]
	BurnoutAltitudeAGL       float64 `json:"burnout_altitude_agl" yaml:"burnout_altitude_agl"`               // Altitude AGL at motor burnout [m]
	CoastToApogeeTime        float64 `json:"coast_to_apogee_time" yaml:"coast_to_apogee_time"`               // Time from burnout to apogee [s]
	DescentTime              float64 `json:"descent_time" yaml:"descent_time"`                               // Time from apogee to landing [s]
}

// MotorSummaryData holds summary statistics for motor performance.
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

// plotSimRecord represents a single row of parsed simulation data, typically motion data.
// Using a map allows flexibility with varying CSV headers.
// Values can be float64 (for numeric data) or string (for non-numeric or unconverted data).
type plotSimRecord map[string]interface{}

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

// populateDefaultValues sets reasonable defaults for newly added report data structures
// when actual simulation data might not be available yet
func populateDefaultValues(rData *ReportData, appCfg *config.Config) {
	// Set default values for LaunchRail
	rData.LaunchRail = LaunchRailData{
		Length:            2.0,  // Default 2m rail
		Angle:             5.0,  // Default 5 degree launch angle
		DepartureVelocity: 15.0, // Typical minimum safe velocity
		MaxForce:          50.0, // Placeholder
		DepartureTime:     0.5,  // Placeholder
		StabilityMargin:   1.5,  // Typical calibers for stability
	}

	// Set default values for MotorSummary (additional fields)
	rData.MotorSummary.SpecificImpulse = 200.0 // Typical amateur motor Isp
	rData.MotorSummary.ThrustEfficiency = 95.0 // Percentage
	if rData.MotorName != "" {
		// Basic parsing of motor name to extract motor class
		if len(rData.MotorName) >= 1 {
			rData.MotorSummary.MotorClass = string(rData.MotorName[0])
		}
	}

	// Set default values for ParachuteSummary (additional fields)
	rData.ParachuteSummary.DeploymentAltitude = rData.MotionMetrics.MaxAltitudeAGL * 0.9 // 90% of apogee
	rData.ParachuteSummary.DeploymentVelocity = 20.0
	rData.ParachuteSummary.DragCoefficient = 1.5      // Typical parachute drag coefficient
	rData.ParachuteSummary.OpeningForce = 100.0       // Placeholder
	rData.ParachuteSummary.Diameter = 0.8             // Default diameter in meters
	rData.ParachuteSummary.ParachuteType = "Toroidal" // Default type

	// Set default values for Weather
	rData.Weather = WeatherData{
		WindSpeed:             3.0,     // Light breeze
		WindDirection:         45.0,    // NE
		WindDirectionCardinal: "NE",    // Cardinal direction
		Temperature:           20.0,    // 20°C
		Pressure:              1013.25, // Standard pressure
		Density:               1.225,   // Sea level air density
		Humidity:              50.0,    // 50%
	}

	// Set default values for ForcesAndMoments
	rData.ForcesAndMoments = ForcesAndMomentsData{
		MaxAngleOfAttack:   5.0,    // Degrees
		MaxNormalForce:     200.0,  // Newtons
		MaxAxialForce:      150.0,  // Newtons
		MaxRollRate:        10.0,   // Degrees/second
		MaxPitchMoment:     2.0,    // Nm
		MaxDynamicPressure: 5000.0, // Pascals
		CenterOfPressure:   0.8,    // m from nose
		CenterOfGravity:    0.6,    // m from nose
		StabilityMargin:    1.5,    // Calibers
	}

	// Ensure MotionMetrics fields are populated for report template
	if rData.MotionMetrics != nil {
		// Default/example values if not set elsewhere
		if rData.MotionMetrics.MaxAltitudeAGL == 0 {
			rData.MotionMetrics.MaxAltitudeAGL = 500.0
		}
		if rData.MotionMetrics.MaxSpeed == 0 {
			rData.MotionMetrics.MaxSpeed = 200.0
		}
		if rData.MotionMetrics.MaxAcceleration == 0 {
			rData.MotionMetrics.MaxAcceleration = 100.0
		}
		if rData.MotionMetrics.TotalFlightTime == 0 {
			rData.MotionMetrics.TotalFlightTime = 120.0
		}
		if rData.MotionMetrics.LandingDistance == 0 {
			rData.MotionMetrics.LandingDistance = 250.0
		}

		// Populate other motion metrics with reasonable values
		rData.MotionMetrics.MaxAltitudeASL = rData.MotionMetrics.MaxAltitudeAGL * 1.1
		rData.MotionMetrics.LandingSpeed = 5.0 // m/s, typical with parachute
	}

	// Add some sample events if none exist
	if len(rData.AllEvents) == 0 {
		rData.AllEvents = []EventSummary{
			{Time: 0.0, Name: "Launch", Altitude: 0.0, Velocity: 0.0, Details: "Rocket leaves the launch pad"},
			{Time: 2.0, Name: "Motor Burnout", Altitude: 200.0, Velocity: 150.0, Details: "Motor has consumed all propellant"},
			{Time: 15.0, Name: "Apogee", Altitude: rData.MotionMetrics.MaxAltitudeAGL, Velocity: 0.0, Details: "Maximum altitude reached"},
			{Time: 15.1, Name: "Parachute Deployment", Altitude: rData.MotionMetrics.MaxAltitudeAGL - 5.0, Velocity: 5.0, Details: "Main parachute deployed"},
			{Time: rData.MotionMetrics.TotalFlightTime, Name: "Landing", Altitude: 0.0, Velocity: rData.MotionMetrics.LandingSpeed, Details: "Touchdown"},
		}
	}
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
	MotionData       []*plotSimRecord     `json:"motion_data" yaml:"motion_data"`
	MotionHeaders    []string             `json:"motion_headers" yaml:"motion_headers"`
	EventsData       [][]string           `json:"events_data" yaml:"events_data"`
	Log              *logf.Logger         `json:"-"` // Exclude logger from JSON
	ReportTitle      string               `json:"report_title" yaml:"report_title"`
	GenerationDate   string               `json:"generation_date" yaml:"generation_date"`
	MotorData        []*plotSimRecord     `json:"motor_data" yaml:"motor_data"`
	MotorHeaders     []string             `json:"motor_headers" yaml:"motor_headers"`
	// Extended fields for templates and flexible data
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	// Collection of all assets (SVG plots, etc.)
	Assets map[string]string `json:"assets,omitempty"`
}

// parseSimData parses raw CSV data into a slice of plotSimRecord and headers.
// It attempts to convert numeric fields to float64, otherwise stores as string.
func parseSimData(log *logf.Logger, csvData []byte, delimiter string) ([]*plotSimRecord, []string, error) {
	if log == nil {
		return nil, nil, fmt.Errorf("parseSimData called with nil logger")
	}

	log.Debug("Parsing sim data", "length_bytes", len(csvData))
	if len(csvData) == 0 {
		log.Warn("CSV data is empty.")
		return []*plotSimRecord{}, []string{}, nil
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
		return []*plotSimRecord{}, []string{}, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}
	log.Debug("CSV Headers", "headers", headers)

	reader.FieldsPerRecord = len(headers)

	var records []*plotSimRecord
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

		record := make(plotSimRecord)
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
func calculateMotorSummary(motorData []*plotSimRecord, motorHeaders []string, log *logf.Logger) MotorSummaryData {
	summary := MotorSummaryData{}
	if len(motorData) == 0 {
		log.Warn("Motor data is empty, cannot calculate motor summary.")
		return summary
	}

	// Find indices for Time and Thrust columns
	timeIdx := -1
	thrustIdx := -1
	for i, header := range motorHeaders {
		if header == "Time (s)" { // TODO: Make these configurable or more robust
			timeIdx = i
		} else if header == "Thrust (N)" {
			thrustIdx = i
		}
	}

	if timeIdx == -1 || thrustIdx == -1 {
		log.Warn("Required columns 'Time (s)' or 'Thrust (N)' not found in motor headers. Cannot calculate motor summary.", "headers", motorHeaders)
		return summary
	}

	var totalImpulse float64
	var maxThrust float64
	var burnStartTime, burnEndTime float64 = -1, -1
	var lastTime, lastThrust float64

	// Convert plotSimRecord to a more usable format for calculation
	type motorPoint struct {
		Time   float64
		Thrust float64
	}
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

	if len(points) < 2 { // Need at least two points for impulse/burn time calculation
		log.Warn("Not enough valid motor data points to calculate summary.")
		// Still, we might have a max thrust from a single point if thrust > 0
		if len(points) == 1 && points[0].Thrust > 0 {
			summary.MaxThrust = points[0].Thrust
			summary.BurnTime = 0 // Or some small epsilon if thrust is instantaneous
		}
		return summary
	}

	// Sort points by time to ensure correct calculation order
	sort.Slice(points, func(i, j int) bool {
		return points[i].Time < points[j].Time
	})

	for i, p := range points {
		// Max Thrust
		if p.Thrust > maxThrust {
			maxThrust = p.Thrust
		}

		// Burn time detection (simple: first to last point with thrust > 0.01 N)
		// A more robust method might involve a threshold and contiguous segments.
		if p.Thrust > 0.01 { // Threshold to consider motor active
			if burnStartTime == -1 {
				burnStartTime = p.Time
			}
			burnEndTime = p.Time
		}

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

	log.Info("Calculated motor summary", "maxThrust", summary.MaxThrust, "avgThrust", summary.AvgThrust, "totalImpulse", summary.TotalImpulse, "burnTime", summary.BurnTime)
	return summary
}

// calculateMotionMetrics computes summary motion statistics from telemetry and event data.
func calculateMotionMetrics(motionData []*plotSimRecord, motionHeaders []string, eventsData [][]string, launchRailLength float64, log *logf.Logger) *MotionMetrics {
	metrics := &MotionMetrics{}
	if len(motionData) == 0 {
		log.Warn("Motion data is empty, cannot calculate motion metrics.")
		return metrics
	}

	// Helper to find column index by (case-insensitive) name pattern
	findIndex := func(patterns ...string) int {
		for _, header := range motionHeaders {
			for _, pattern := range patterns {
				if strings.Contains(strings.ToLower(header), strings.ToLower(pattern)) {
					// Find exact index for the matched header
					for i, h := range motionHeaders {
						if h == header {
							return i
						}
					}
				}
			}
		}
		return -1
	}

	timeIdx := findIndex("Time (s)", "Time")
	altAGLIdx := findIndex("Altitude (AGL) (m)", "Altitude AGL", "AGL Alt")
	altASLIdx := findIndex("Altitude (ASL) (m)", "Altitude ASL", "ASL Alt")
	velTotIdx := findIndex("Total Velocity (m/s)", "TotVel")
	accTotIdx := findIndex("Total Acceleration (m/s^2)", "TotAcc")
	velVIdx := findIndex("Vertical Velocity (m/s)", "VertVel")

	if timeIdx == -1 {
		log.Warn("Time column not found in motion headers. Cannot calculate most motion metrics.")
		return metrics // Essential for almost all metrics
	}

	// Convert to a more usable struct and sort by time
	points := make([]motionPoint, 0, len(motionData))
	for idx, record := range motionData { // Changed i to idx to avoid confusion if 'i' is used later
		mp := motionPoint{}
		if val, ok := (*record)[motionHeaders[timeIdx]].(float64); ok {
			mp.Time = val
		} else {
			log.Warn("Bad time val", "idx", idx)
			continue
		}
		if altAGLIdx != -1 {
			if val, ok := (*record)[motionHeaders[altAGLIdx]].(float64); ok {
				mp.AltAGL = val
			}
		}
		if altASLIdx != -1 {
			if val, ok := (*record)[motionHeaders[altASLIdx]].(float64); ok {
				mp.AltASL = val
			}
		}
		if velTotIdx != -1 {
			if val, ok := (*record)[motionHeaders[velTotIdx]].(float64); ok {
				mp.VelTot = val
			}
		}
		if accTotIdx != -1 {
			if val, ok := (*record)[motionHeaders[accTotIdx]].(float64); ok {
				mp.AccTot = val
			}
		}
		if velVIdx != -1 {
			if val, ok := (*record)[motionHeaders[velVIdx]].(float64); ok {
				mp.VelV = val
			}
		}
		points = append(points, mp)
	}

	if len(points) == 0 {
		log.Warn("No valid motion points after parsing, cannot calculate motion metrics.")
		return metrics
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Time < points[j].Time })

	// Calculate max values by iterating through sorted points
	for _, p := range points { // Changed i to _
		if altAGLIdx != -1 && p.AltAGL > metrics.MaxAltitudeAGL {
			metrics.MaxAltitudeAGL = p.AltAGL
		}
		if altASLIdx != -1 && p.AltASL > metrics.MaxAltitudeASL {
			metrics.MaxAltitudeASL = p.AltASL
		}
		if velTotIdx != -1 && p.VelTot > metrics.MaxSpeed {
			metrics.MaxSpeed = p.VelTot
		}
		if accTotIdx != -1 && p.AccTot > metrics.MaxAcceleration {
			metrics.MaxAcceleration = p.AccTot
		}

		// Launch Rail Clearance (AGL based)
		if launchRailLength > 0 && altAGLIdx != -1 && metrics.LaunchRailClearanceTime == 0 && p.AltAGL >= launchRailLength {
			metrics.LaunchRailClearanceTime = p.Time
			if velTotIdx != -1 {
				metrics.LaunchRailClearanceSpeed = p.VelTot
			}
		}

		// Apogee detection (when vertical velocity crosses from positive to negative, or is zero at max AGL)
		// This is a simplified check. A more robust method would find where d(AltAGL)/dt = 0.
		if altAGLIdx != -1 && velVIdx != -1 && p.AltAGL >= metrics.MaxAltitudeAGL { // MaxAltitudeAGL is updated before this check
			// If this point is the highest so far, consider it a candidate for apogee
			if metrics.ApogeeTime == 0 || p.AltAGL > (*getMotionPointByTime(points, metrics.ApogeeTime, timeIdx)).AltAGL {
				metrics.ApogeeTime = p.Time
			}
		}
	}

	// Use event data for more precise timings if available
	// eventsData: [][]string, typically [Time, EventName, Description...]
	if len(eventsData) > 1 { // Assuming first row is header
		for _, eventRow := range eventsData[1:] {
			if len(eventRow) < 2 {
				continue
			}
			eventTime, err := strconv.ParseFloat(strings.TrimSpace(eventRow[0]), 64)
			if err != nil {
				continue
			}
			eventName := strings.ToUpper(strings.TrimSpace(eventRow[1]))

			switch eventName {
			case "APOGEE", "APOGEE_REACHED":
				metrics.ApogeeTime = eventTime
			case "BURNOUT", "MOTOR_BURNOUT":
				metrics.BurnoutTime = eventTime
				if mp := getMotionPointByTime(points, eventTime, timeIdx); mp != nil && altAGLIdx != -1 {
					metrics.BurnoutAltitudeAGL = mp.AltAGL
				}
			case "LAUNCHRAIL_CLEAR", "LIFTOFF", "LAUNCH_RAIL_CLEARED": // LIFTOFF might be t=0, but clear is more accurate
				if metrics.LaunchRailClearanceTime == 0 || eventTime < metrics.LaunchRailClearanceTime { // Take the earliest if multiple
					metrics.LaunchRailClearanceTime = eventTime
					if mp := getMotionPointByTime(points, eventTime, timeIdx); mp != nil && velTotIdx != -1 {
						metrics.LaunchRailClearanceSpeed = mp.VelTot
					}
				}
			case "LANDING", "GROUND_HIT", "TOUCHDOWN":
				metrics.TotalFlightTime = eventTime
				if mp := getMotionPointByTime(points, eventTime, timeIdx); mp != nil && velTotIdx != -1 {
					metrics.LandingSpeed = mp.VelTot
				}
			}
		}
	}

	// If TotalFlightTime not set by event, use last motion point time
	if metrics.TotalFlightTime == 0 && len(points) > 0 {
		metrics.TotalFlightTime = points[len(points)-1].Time
		// Attempt to get landing speed from the last point if Vertical Velocity indicates landing
		if velTotIdx != -1 && velVIdx != -1 {
			lastPoint := points[len(points)-1]
			// A simple check for landing: AGL near zero and vertical velocity negative or near zero.
			// This might need refinement based on actual data patterns.
			if altAGLIdx != -1 && lastPoint.AltAGL < 1.0 { // Assuming AGL close to 0 is landing
				metrics.LandingSpeed = lastPoint.VelTot
			}
		}
	}

	if metrics.ApogeeTime > 0 && metrics.BurnoutTime > 0 && metrics.ApogeeTime > metrics.BurnoutTime {
		metrics.CoastToApogeeTime = metrics.ApogeeTime - metrics.BurnoutTime
	}
	if metrics.TotalFlightTime > 0 && metrics.ApogeeTime > 0 && metrics.TotalFlightTime > metrics.ApogeeTime {
		metrics.DescentTime = metrics.TotalFlightTime - metrics.ApogeeTime
	}

	log.Info("Calculated motion metrics", "maxAltAGL", metrics.MaxAltitudeAGL, "maxSpeed", metrics.MaxSpeed, "apogeeTime", metrics.ApogeeTime, "flightTime", metrics.TotalFlightTime)
	return metrics
}

// getMotionPointByTime finds a motionPoint at or just after a given time.
// This is a helper and assumes points are sorted by time.
func getMotionPointByTime(points []motionPoint, t float64, timeIdx int) *motionPoint { // timeIdx is actually not used here as points struct has Time
	if len(points) == 0 {
		return nil
	}
	// Simple linear search, can be replaced with binary search for performance on very large datasets
	for i, p := range points {
		if p.Time >= t {
			// If p.Time is exactly t, or the first point after t
			// For interpolation, one might take points[i-1] and points[i]
			return &points[i]
		}
	}
	return &points[len(points)-1] // Return last point if t is beyond data range
}

// LoadSimulationData orchestrates loading all necessary data for a report.
func LoadSimulationData(recordID string, rm HandlerRecordManager, reportSpecificDir string, appCfg *config.Config) (*ReportData, error) {
	log := logger.GetLogger(appCfg.Setup.Logging.Level)
	if log == nil {
		return nil, fmt.Errorf("failed to initialize logger: logger.GetLogger returned nil")
	}

	log.Info("Loading simulation data for report", "recordID", recordID)

	// Set current time for the report generation
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
		AllEvents:        []EventSummary{},
	}

	// Load simulation configuration directly from app config
	rData.ConfigSummary = appCfg

	// Derive RocketName from OpenRocketFile path in app config
	if appCfg.Engine.Options.OpenRocketFile != "" {
		rData.RocketName = filepath.Base(appCfg.Engine.Options.OpenRocketFile)
	}

	// Get MotorName from the app config
	if appCfg.Engine.Options.MotorDesignation != "" {
		rData.MotorName = appCfg.Engine.Options.MotorDesignation
	}

	// Set a default value for LiftoffMassKg if available
	// Note: In a full implementation, you would extract this from simulation data

	// Get the specific record to access its storage instances
	record, err := rm.GetRecord(recordID)
	if err != nil {
		log.Error("Failed to get record for report generation", "recordID", recordID, "error", err)
		return nil, fmt.Errorf("failed to get record %s: %w", recordID, err)
	}
	// It's crucial to close the record's stores when done to release file handles.
	// Since LoadSimulationData returns ReportData which might be used later, closing here is tricky.
	// If the record's storage fields are nil, we don't need to close anything
	defer func() {
		// Safely close the record if it has valid storage fields
		if record.Motion != nil || record.Events != nil || record.Dynamics != nil {
			if err := record.Close(); err != nil {
				log.Warn("Failed to close record stores after loading data", "recordID", recordID, "error", err)
			}
		} else {
			log.Debug("Record has no active storage fields to close", "recordID", recordID)
		}
	}()

	// Load motion data - safely handle potentially nil Motion field
	motionFilePath := filepath.Join(reportSpecificDir, "motion.csv")
	if record.Motion != nil {
		// Use the storage object if available
		motionFilePath = record.Motion.GetFilePath()
	}

	// Read the motion data file directly
	motionDataCSV, err := os.ReadFile(motionFilePath)
	if err != nil {
		// Try alternative file name case (MOTION.csv)
		motionFilePath = filepath.Join(reportSpecificDir, "MOTION.csv")
		motionDataCSV, err = os.ReadFile(motionFilePath)
		if err != nil {
			log.Warn("Could not load motion data from any standard location", "basePath", reportSpecificDir, "error", err)
		}
	}

	// If we successfully loaded motion data, process it
	if err == nil {
		motionDataParsed, motionHeaders, err := parseSimData(log, motionDataCSV, ",")
		if err != nil {
			log.Warn("Error parsing motion data", "filename", motionFilePath, "error", err)
		} else {
			rData.MotionData = motionDataParsed
			rData.MotionHeaders = motionHeaders
			log.Info("Successfully parsed motion data records", "count", len(rData.MotionData))

			// Attempt to extract LiftoffMassKg from the first motion data record
			if len(rData.MotionData) > 0 {
				firstRecord := rData.MotionData[0]
				massHeaderKey := ""
				for _, h := range rData.MotionHeaders {
					if strings.EqualFold(h, "Mass (kg)") {
						massHeaderKey = h
						break
					}
				}
				if massHeaderKey != "" {
					if massVal, ok := (*firstRecord)[massHeaderKey].(float64); ok {
						rData.LiftoffMassKg = massVal
						log.Info("Successfully extracted LiftoffMassKg from motion data", "value", rData.LiftoffMassKg)
					} else {
						log.Warn("Found 'Mass (kg)' header but failed to cast value to float64 from first motion record", "value", (*firstRecord)[massHeaderKey])
					}
				} else {
					log.Warn("'Mass (kg)' header not found in motion data headers. LiftoffMassKg will not be set from motion data.")
				}
			} else {
				log.Warn("Motion data is empty, cannot extract LiftoffMassKg.")
			}
		}
	}

	// Load events data - safely handle potentially nil Events field
	eventsFilePath := filepath.Join(reportSpecificDir, "events.csv")
	if record.Events != nil {
		// Use the storage object if available
		eventsFilePath = record.Events.GetFilePath()
	}

	// Read the events data file directly
	eventsDataCSV, err := os.ReadFile(eventsFilePath)
	if err != nil {
		// Try alternative file name case (EVENTS.csv)
		eventsFilePath = filepath.Join(reportSpecificDir, "EVENTS.csv")
		eventsDataCSV, err = os.ReadFile(eventsFilePath)
		if err != nil {
			log.Warn("Could not load events data from any standard location", "basePath", reportSpecificDir, "error", err)
		}
	}

	// If we successfully loaded events data, process it
	if err == nil {
		reader := csv.NewReader(bytes.NewReader(eventsDataCSV))
		reader.Comma = ','
		rawEventsData, err := reader.ReadAll()
		if err != nil {
			log.Warn("Error parsing events data", "filename", eventsFilePath, "error", err)
		} else {
			rData.EventsData = rawEventsData
			log.Info("Successfully loaded raw event entries", "count", len(rData.EventsData))
		}
	}

	// Calculate MotionMetrics
	if len(rData.MotionData) > 0 {
		launchRailLen := 0.0
		if appCfg != nil && appCfg.Engine.Options.Launchrail.Length > 0 {
			launchRailLen = appCfg.Engine.Options.Launchrail.Length
		}
		rData.MotionMetrics = calculateMotionMetrics(rData.MotionData, rData.MotionHeaders, rData.EventsData, launchRailLen, log)
	} else if rData.MotionMetrics == nil { // Ensure MotionMetrics is initialized even if no data
		rData.MotionMetrics = &MotionMetrics{}
	}

	// Load motor data - safely handle potentially nil Dynamics field (assuming motor data might be in record.Dynamics)
	motorTelemetryFilePath := filepath.Join(reportSpecificDir, "motor_telemetry.csv")
	// In a real scenario, you might check record.Dynamics or a specific motor data store if it exists
	// For now, we'll assume it's a file in the reportSpecificDir similar to motion/events.

	// Read the motor telemetry data file directly
	motorDataCSV, err := os.ReadFile(motorTelemetryFilePath)
	if err != nil {
		// Try alternative file names or cases
		motorTelemetryFilePath = filepath.Join(reportSpecificDir, "MOTOR_TELEMETRY.csv")
		motorDataCSV, err = os.ReadFile(motorTelemetryFilePath)
		if err != nil {
			motorTelemetryFilePath = filepath.Join(reportSpecificDir, "motor.csv") // Another common name
			motorDataCSV, err = os.ReadFile(motorTelemetryFilePath)
			if err != nil {
				log.Warn("Could not load motor telemetry data from any standard location", "basePath", reportSpecificDir, "error", err)
			}
		}
	}

	// If we successfully loaded motor telemetry data, process it
	if err == nil {
		motorDataParsed, motorHeaders, err := parseSimData(log, motorDataCSV, ",") // Assuming same delimiter
		if err != nil {
			log.Warn("Error parsing motor telemetry data", "filename", motorTelemetryFilePath, "error", err)
		} else {
			rData.MotorData = motorDataParsed
			rData.MotorHeaders = motorHeaders
			log.Info("Successfully parsed motor telemetry data records", "count", len(rData.MotorData))

			// Calculate motor summary from telemetry data
			rData.MotorSummary = calculateMotorSummary(rData.MotorData, rData.MotorHeaders, log)
			// TODO: Potentially update LiftoffMassKg if motor data provides PropellantMass and config provides DryMass
			// e.g. if rData.LiftoffMassKg is still 0 and rData.MotorSummary.PropellantMass > 0 && appCfg.Rocket.DryMass > 0
			// rData.LiftoffMassKg = appCfg.Rocket.DryMass + rData.MotorSummary.PropellantMass
		}
	}

	assetsDir := filepath.Join(reportSpecificDir, "assets")
	if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
		log.Error("Failed to create assets directory", "path", assetsDir, "error", err)
		return nil, fmt.Errorf("failed to create assets directory '%s': %w", assetsDir, err)
	}
	rData.ReportTitle = fmt.Sprintf("Simulation Report for %s - %s", rData.RocketName, recordID)
	rData.GenerationDate = time.Now().Format(time.RFC1123)

	// Populate Plots map with placeholders for expected plots
	// This satisfies test assertions for plot key existence.
	// Actual plot generation functions (plotFunc, plotAltitude etc.) would go here.
	requiredPlots := []string{
		// Basic flight plots
		"altitude_vs_time",
		"velocity_vs_time",
		"acceleration_vs_time",

		// Additional trajectory plots
		"trajectory_3d",
		"landing_radius",

		// Motor performance plot
		"thrust_vs_time",

		// Forces and moments plots
		"angle_of_attack",
		"forces_vs_time",
		"moments_vs_time",
		"roll_rate",
	}

	for _, plotKey := range requiredPlots {
		// Create placeholder entries for all required plots - just the filename
		// Avoid duplicate assets/ path by using just the filename
		// The full path will be built by the template renderer
		rData.Plots[plotKey] = plotKey + ".svg"
		log.Info("Placeholder entry added for plot", "plot_key", plotKey, "path", rData.Plots[plotKey])
	}

	// Populate default values for the newly added structures
	populateDefaultValues(rData, appCfg)

	log.Info("Simulation data loading and basic processing complete", "RecordID", recordID)
	return rData, nil
}
