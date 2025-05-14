package reporting

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/zerodha/logf"
)

// RecordManager defines the interface for accessing record data, used by GenerateReportData.
// This allows for easier testing and decoupling from concrete storage implementations.
// It is satisfied by *storage.RecordManager and test mocks.
// Ensure HandlerRecordManager in cmd/server/handlers.go is compatible.
type RecordManager interface {
	GetRecord(hash string) (*storage.Record, error)
	GetStorageDir() string
	// Add other methods if GenerateReportData needs them, e.g., ListRecords, DeleteRecord.
	// For now, only GetRecord and GetStorageDir seem to be used based on ReportAPIV2's needs.
}

// findEventIndex searches for an event by name in the eventsData and returns its index.
// Assumes event name is in the first column (index 0) of each event row.
func FindEventIndex(eventsData [][]string, eventName string) int {
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

// ConvertMarkdownToSimpleHTML converts a markdown string to a very basic HTML representation.
func ConvertMarkdownToSimpleHTML(mdContent string, recordID string) string {
	htmlOutput := "<!DOCTYPE html>\n"
	htmlOutput += "<html>\n"
	htmlOutput += "<head>\n"
	htmlOutput += "<style>body { font-family: sans-serif; margin: 20px; } h1, h2, h3 { color: #333; } pre { background-color: #f5f5f5; padding: 10px; border: 1px solid #ddd; overflow-x: auto; }</style>\n"
	htmlOutput += "</head>\n"
	htmlOutput += "<body>\n"
	htmlOutput += fmt.Sprintf("<h1>Simulation Report: %s</h1>\n", recordID)

	lines := strings.Split(strings.ReplaceAll(mdContent, "\r\n", "\n"), "\n\n")
	for _, paragraph := range lines {
		if strings.HasPrefix(paragraph, "```") {
			codeBlock := strings.TrimPrefix(paragraph, "```")
			codeBlock = strings.TrimSuffix(codeBlock, "```")
			codeBlock = strings.TrimSpace(codeBlock)
			escapedCode := strings.ReplaceAll(codeBlock, "<", "&lt;")
			escapedCode = strings.ReplaceAll(escapedCode, ">", "&gt;")
			htmlOutput += "<pre><code>" + escapedCode + "</code></pre>\n"
		} else if strings.HasPrefix(paragraph, "### ") {
			htmlOutput += "<h3>" + strings.ReplaceAll(strings.TrimPrefix(paragraph, "### "), "<", "&lt;") + "</h3>\n"
		} else if strings.HasPrefix(paragraph, "## ") {
			htmlOutput += "<h2>" + strings.ReplaceAll(strings.TrimPrefix(paragraph, "## "), "<", "&lt;") + "</h2>\n"
		} else if strings.HasPrefix(paragraph, "# ") {
			htmlOutput += "<h1>" + strings.ReplaceAll(strings.TrimPrefix(paragraph, "# "), "<", "&lt;") + "</h1>\n"
		} else if strings.TrimSpace(paragraph) != "" {
			htmlOutput += "<p>" + strings.ReplaceAll(paragraph, "<", "&lt;") + "</p>\n"
		}
	}

	htmlOutput += "</body>\n"
	htmlOutput += "</html>"
	return htmlOutput
}

// ReportData holds all data required to generate a report.
type ReportData struct {
	RecordID         string                  `json:"record_id" yaml:"record_id"`
	Version          string                  `json:"version" yaml:"version"`
	RocketName       string                  `json:"rocket_name" yaml:"rocket_name"`
	MotorName        string                  `json:"motor_name" yaml:"motor_name"`
	LiftoffMassKg    float64                 `json:"liftoff_mass_kg" yaml:"liftoff_mass_kg"`
	GeneratedTime    string                  `json:"generated_time" yaml:"generated_time"`
	Config           config.Config           `json:"config" yaml:"config"`
	Summary          ReportSummary           `json:"summary" yaml:"summary"`
	Plots            map[string]string       `json:"plots" yaml:"plots"`
	MotionMetrics    *MotionMetrics          `json:"motion_metrics" yaml:"motion_metrics"`
	MotorSummary     MotorSummaryData        `json:"motor_summary" yaml:"motor_summary"`
	ParachuteSummary ParachuteSummaryData    `json:"parachute_summary" yaml:"parachute_summary"`
	PhaseSummary     PhaseSummaryData        `json:"phase_summary" yaml:"phase_summary"`
	LaunchRail       LaunchRailData          `json:"launch_rail" yaml:"launch_rail"`
	ForcesAndMoments ForcesAndMomentsData    `json:"forces_and_moments" yaml:"forces_and_moments"`
	Weather          WeatherData             `json:"weather" yaml:"weather"`
	AllEvents        []EventSummary          `json:"all_events" yaml:"all_events"`
	Stages           []StageData             `json:"stages" yaml:"stages"`
	RecoverySystems  []RecoverySystemData    `json:"recovery_systems" yaml:"recovery_systems"`
	MotionData       []*PlotSimRecord        `json:"motion_data" yaml:"motion_data"`
	MotionHeaders    []string                `json:"motion_headers" yaml:"motion_headers"`
	EventsData       [][]string              `json:"events_data" yaml:"events_data"`
	Log              *logf.Logger            `json:"-"`
	ReportTitle      string                  `json:"report_title" yaml:"report_title"`
	GenerationDate   string                  `json:"generation_date" yaml:"generation_date"`
	MotorData        []*PlotSimRecord        `json:"motor_data" yaml:"motor_data"`
	MotorHeaders     []string                `json:"motor_headers" yaml:"motor_headers"`
	Extensions       map[string]interface{}  `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	Assets           map[string]string       `json:"assets,omitempty" yaml:"assets,omitempty"`
	RawData          *storage.SimulationData `json:"raw_data" yaml:"raw_data"`
}

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
	RocketName        string        `json:"rocket_name" yaml:"rocket_name"`
	MotorDesignation  string        `json:"motor_designation" yaml:"motor_designation"`
	LaunchSite        string        `json:"launch_site" yaml:"launch_site"`
	TargetApogeeFt    float64       `json:"target_apogee_ft" yaml:"target_apogee_ft"` // Ensure this field exists in config or is handled
	LiftoffMassKg     float64       `json:"liftoff_mass_kg" yaml:"liftoff_mass_kg"`
	CoastToApogeeTime float64       `json:"coast_to_apogee_time" yaml:"coast_to_apogee_time"`
	MotionMetrics     MotionMetrics `json:"motion_metrics" yaml:"motion_metrics"`
	// Use existing RecoverySystemData type for the recovery system information
	RecoverySystem []RecoverySystemData `json:"recovery_system" yaml:"recovery_system"`
	Notes          string               `json:"notes,omitempty" yaml:"notes,omitempty"`
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
	Latitude              float64
	Longitude             float64
	ElevationAMSL         float64
	WindSpeed             float64
	WindDirection         float64
	WindDirectionCardinal string
	Pressure              float64 // Atmospheric pressure at launch site altitude (Pa)
	SeaLevelPressure      float64 // Standard sea level pressure (Pa)
	Density               float64 // Air density at launch site (kg/m³)
	LocalGravity          float64 // Local gravity at launch site (m/s²)
	SpeedOfSound          float64 // Speed of sound in current conditions (m/s)
	TemperatureK          float64 // Temperature in Kelvin
}

// Note: EventSummary is defined above with the required fields for the FRR report

// parseSimData parses raw CSV data into a slice of PlotSimRecord and headers.
// It attempts to convert numeric fields to float64, otherwise stores as string.
// REFACTORED: Now accepts pre-read headers and data rows, instead of raw bytes.
func parseSimData(log *logf.Logger, headers []string, data [][]string) ([]*PlotSimRecord, error) {
	if log == nil {
		return nil, fmt.Errorf("parseSimData called with nil logger")
	}

	log.Debug("Parsing sim data from pre-read content", "num_headers", len(headers), "num_data_rows", len(data))

	if len(headers) == 0 {
		log.Warn("Headers are empty.")
		// Depending on strictness, this could be an error or return empty.
		// For now, assume it's valid to have no headers, meaning no processable data for plots.
		return []*PlotSimRecord{}, nil
	}

	if len(data) == 0 {
		log.Warn("Data rows are empty.")
		return []*PlotSimRecord{}, nil
	}

	var records []*PlotSimRecord
	for rowIndex, row := range data {
		if len(row) != len(headers) {
			log.Warn("Skipping row due to mismatched column count",
				"row_index", rowIndex,
				"expected_cols", len(headers),
				"actual_cols", len(row),
				"row_data", row)
			continue
		}

		record := make(PlotSimRecord)
		for i, header := range headers {
			rawValue := row[i]
			if valFloat, errFloat := strconv.ParseFloat(rawValue, 64); errFloat == nil {
				record[header] = valFloat
			} else {
				record[header] = rawValue
			}
		}
		records = append(records, &record)
	}

	if len(records) == 0 && len(data) > 0 {
		log.Warn("Data rows were provided but no records could be parsed (e.g., all rows had mismatched columns).")
	}

	log.Debug("Parsed data records", "num_records", len(records))
	return records, nil
}

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
			if timeIdx == -1 { // Prefer "Time (s)" but accept "Time"
				timeIdx = i
			}
		case "thrust (n)":
			thrustIdx = i
		case "thrust":
			if thrustIdx == -1 { // Prefer "Thrust (N)" but accept "Thrust"
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
	burnStartTime = -1 // Initialize to indicate not yet found

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

	// Calculate Specific Impulse (Isp = total impulse / (propellant mass * g0))
	// If propellant mass is unknown, use a default estimate based on total impulse
	// Typically, commercial motors have ~200-250 N·s/kg specific impulse
	estimatedPropellantMass := totalImpulse / 220.0 // Reasonable estimate
	summary.PropellantMass = estimatedPropellantMass

	// Standard gravity constant (m/s²)
	const g0 = 9.80665

	// Calculate specific impulse (Isp in seconds)
	if estimatedPropellantMass > 0 {
		summary.SpecificImpulse = totalImpulse / (estimatedPropellantMass * g0)
	}

	// Calculate thrust efficiency (avg thrust / max thrust ratio)
	if summary.MaxThrust > 0 {
		summary.ThrustEfficiency = summary.AvgThrust / summary.MaxThrust
	}

	// Determine motor class based on total impulse (using NAR/TRA classification)
	// https://www.nar.org/standards-and-testing-committee/nar-standards/
	summary.MotorClass = designation.DetermineMotorClass(totalImpulse)

	// Set manufacturer if available (or use a default or extract from designation if available)
	summary.MotorManufacturer = "Unknown" // Default

	log.Info("Calculated motor summary",
		"maxThrust", summary.MaxThrust,
		"totalImpulse", summary.TotalImpulse,
		"burnTime", summary.BurnTime,
		"avgThrust", summary.AvgThrust,
		"specificImpulse", summary.SpecificImpulse,
		"motorClass", summary.MotorClass)
	return summary
}

// motionPoint struct for internal use in CalculateMotionMetrics and helpers
type motionPoint struct {
	Time   float64
	AltAGL float64
	VelTot float64
	AccTot float64
}

// CalculateMotionMetrics computes summary motion statistics from telemetry and event data.
func CalculateMotionMetrics(motionData []*PlotSimRecord, motionHeaders []string, eventsData [][]string, launchRailLength float64, log *logf.Logger) *MotionMetrics {
	log.Debug("CalculateMotionMetrics: processing with motion headers", "motion_headers", motionHeaders)

	metrics := &MotionMetrics{}

	if eventsData == nil {
		log.Warn("CalculateMotionMetrics: eventsData is nil. Event-dependent metrics will not be calculated or will use fallbacks if available.")
	}

	timeIdx, altitudeIdx, velocityIdx, accelIdx := FindMotionDataIndices(motionHeaders)
	if timeIdx == -1 {
		log.Error("Essential 'Time' or 'Time (s)' column not found in motion data headers. Motion-point-based metrics will not be calculated.", "received_headers", motionHeaders)
	}

	motionPoints := ExtractMotionPoints(motionData, motionHeaders, timeIdx, altitudeIdx, velocityIdx, accelIdx, log)
	if len(motionPoints) == 0 && timeIdx != -1 {
		log.Warn("No motion points extracted from motionData. Point-based metrics may be inaccurate or unavailable.")
	}

	launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx := FindFlightEvents(eventsData, log)
	log.Debug("Event indices found", "launch", launchIdx, "railExit", railExitIdx, "burnout", burnoutIdx, "apogee", apogeeEventIdx, "touchdown", touchdownIdx)

	// FlightTime
	if launchIdx != -1 && touchdownIdx != -1 {
		metrics.FlightTime = CalculateFlightTime(eventsData, launchIdx, touchdownIdx)
		if metrics.FlightTime < 0 {
			log.Warn("Failed to calculate valid flight time from events.")
			metrics.FlightTime = 0
		}
	} else {
		log.Warn("Cannot calculate FlightTime: Launch or Touchdown events not found or eventsData missing.")
	}

	// TimeToApogee & TimeAtApogee (event-based first, then motion-based fallback)
	launchTime := 0.0
	if launchIdx != -1 {
		ltStr := eventsData[launchIdx][1] // Assuming time is column 1
		lt, err := strconv.ParseFloat(ltStr, 64)
		if err == nil {
			launchTime = lt
		} else {
			log.Warn("Could not parse launch event time for metrics.", "error", err)
		}
	}

	if apogeeEventIdx != -1 && launchIdx != -1 {
		apogeeTimeStr := eventsData[apogeeEventIdx][1]
		apogeeTime, err := strconv.ParseFloat(apogeeTimeStr, 64)
		if err == nil && launchTime > -1 { // ensure launchTime was parsed
			metrics.TimeToApogee = apogeeTime - launchTime
			metrics.TimeAtApogee = apogeeTime
		} else {
			log.Warn("Could not parse apogee event time or launch time invalid.", "error", err)
		}
	}

	// MaxAltitudeAGL and motion-based apogee time updates
	maxAltFromMotion, _, _, apogeeTimeFromMotion := FindPeakValues(motionPoints)
	if maxAltFromMotion > 0 {
		metrics.MaxAltitudeAGL = maxAltFromMotion
		if metrics.TimeAtApogee == 0 { // If event-based was not set
			metrics.TimeAtApogee = apogeeTimeFromMotion
			if launchTime > -1 && apogeeTimeFromMotion >= launchTime { // Ensure launchTime was valid
				metrics.TimeToApogee = apogeeTimeFromMotion - launchTime
			} else if launchTime == -1 { // Fallback if launch event time wasn't available for motion based TimeToApogee
				if len(motionPoints) > 0 {
					metrics.TimeToApogee = apogeeTimeFromMotion - motionPoints[0].Time
					log.Info("Used first motion point time as launch for motion-based TimeToApogee.")
				}
			}
		}
	}

	// BurnoutTime
	if burnoutIdx != -1 && launchIdx != -1 {
		burnoutTimeStr := eventsData[burnoutIdx][1]
		burnoutTime, err := strconv.ParseFloat(burnoutTimeStr, 64)
		if err == nil && launchTime > -1 {
			metrics.BurnoutTime = burnoutTime - launchTime
		} else {
			log.Warn("Could not parse burnout event time or launch time invalid.", "error", err)
		}
	}

	// DescentTime
	if metrics.FlightTime > 0 && metrics.TimeToApogee > 0 && metrics.FlightTime > metrics.TimeToApogee {
		metrics.DescentTime = metrics.FlightTime - metrics.TimeToApogee
	} else {
		log.Warn("Cannot calculate DescentTime: FlightTime or TimeToApogee is missing or invalid.")
	}

	// CoastToApogeeTime
	if metrics.BurnoutTime > 0 && metrics.TimeToApogee > 0 && metrics.TimeToApogee > metrics.BurnoutTime {
		metrics.CoastToApogeeTime = metrics.TimeToApogee - metrics.BurnoutTime
	} else if metrics.BurnoutTime > 0 && metrics.TimeToApogee > 0 {
		log.Warn("Calculated TimeToApogee is not greater than BurnoutTime, CoastToApogeeTime may be invalid.", "timeToApogee", metrics.TimeToApogee, "burnoutTime", metrics.BurnoutTime)
	} else {
		log.Warn("Cannot calculate CoastToApogeeTime: BurnoutTime or TimeToApogee is missing or invalid.")
	}

	// BurnoutAltitude
	if metrics.BurnoutTime > 0 && launchTime > -1 { // Requires BurnoutTime and valid launchTime
		actualBurnoutTimestamp := launchTime + metrics.BurnoutTime
		closestPointToBurnout := FindClosestMotionPoint(motionPoints, actualBurnoutTimestamp)
		if closestPointToBurnout != nil {
			metrics.BurnoutAltitude = closestPointToBurnout.AltAGL
		} else {
			log.Warn("Could not find a motion point close to burnout time to determine BurnoutAltitude.")
		}
	}

	// MaxSpeed & MaxAcceleration from motion data
	_, maxSpeedFromMotion, maxAccelFromMotion, _ := FindPeakValues(motionPoints)
	if maxSpeedFromMotion > 0 {
		metrics.MaxSpeed = maxSpeedFromMotion
	}
	if maxAccelFromMotion > 0 {
		metrics.MaxAcceleration = maxAccelFromMotion
	}

	// RailExitVelocity & LaunchRailClearanceTime
	if railExitIdx != -1 {
		railExitTimeStr := eventsData[railExitIdx][1]
		railExitTime, err := strconv.ParseFloat(railExitTimeStr, 64)
		if err == nil {
			closestPoint := FindClosestMotionPoint(motionPoints, railExitTime)
			if closestPoint != nil {
				metrics.RailExitVelocity = closestPoint.VelTot
			}
			if launchTime > -1 {
				metrics.LaunchRailClearanceTime = railExitTime - launchTime
			}
		} else {
			log.Warn("Could not parse rail exit event time.", "error", err)
		}
	} else if launchRailLength > 0 && len(motionPoints) > 0 { // Fallback using launchRailLength
		for _, p := range motionPoints {
			if p.AltAGL >= launchRailLength {
				metrics.RailExitVelocity = p.VelTot
				if launchTime > -1 && p.Time >= launchTime {
					metrics.LaunchRailClearanceTime = p.Time - launchTime
				} else if launchTime == -1 && len(motionPoints) > 0 { // Absolute fallback if no launch time
					metrics.LaunchRailClearanceTime = p.Time - motionPoints[0].Time
				}
				log.Debug("Calculated rail exit velocity and time based on launch rail length", "velocity", metrics.RailExitVelocity, "time", metrics.LaunchRailClearanceTime)
				break
			}
		}
		if metrics.RailExitVelocity == 0 {
			log.Warn("Could not determine rail exit velocity using launch rail length.")
		}
	}

	// LandingSpeed
	if touchdownIdx != -1 {
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
	} else if len(motionPoints) > 0 { // Fallback for landing speed
		lastPoint := motionPoints[len(motionPoints)-1]
		if lastPoint.AltAGL <= 1.0 { // Assuming near ground
			metrics.LandingSpeed = lastPoint.VelTot
			log.Info("Calculated LandingSpeed using last motion point near ground level.", "velocity", metrics.LandingSpeed, "altAGL", lastPoint.AltAGL)
		}
	}

	log.Debug("Calculated motion metrics", "metrics", fmt.Sprintf("%+v", metrics))
	return metrics
}

// FindMotionDataIndices finds the indices of key motion data headers.
func FindMotionDataIndices(motionHeaders []string) (timeIdx, altitudeIdx, velocityIdx, accelIdx int) {
	timeIdx, altitudeIdx, velocityIdx, accelIdx = -1, -1, -1, -1
	for i, header := range motionHeaders {
		cleanHeader := strings.ToLower(strings.TrimSpace(header))

		// Check for time column
		if timeIdx == -1 && (cleanHeader == "time" || cleanHeader == "time (s)" || strings.Contains(cleanHeader, "time")) {
			timeIdx = i
		}

		// Check for altitude column
		if altitudeIdx == -1 && (cleanHeader == "altitude" || cleanHeader == "altitude agl (m)" || strings.Contains(cleanHeader, "altitude")) {
			altitudeIdx = i
		}

		// Check for velocity column
		if velocityIdx == -1 && (cleanHeader == "velocity" || cleanHeader == "total velocity (m/s)" || strings.Contains(cleanHeader, "velocity")) {
			velocityIdx = i
		}

		// Check for acceleration column
		if accelIdx == -1 && (cleanHeader == "acceleration" || cleanHeader == "total acceleration (m/s^2)" || strings.Contains(cleanHeader, "acceleration")) {
			accelIdx = i
		}
	}
	return timeIdx, altitudeIdx, velocityIdx, accelIdx
}

// FindFlightEvents processes eventsData to find indices for key flight events.
// It assumes event name is in the first column and time in the second after headers.
func FindFlightEvents(eventsData [][]string, log *logf.Logger) (launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx int) {
	launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx = -1, -1, -1, -1, -1

	if len(eventsData) == 0 {
		log.Warn("No event data available for finding flight events.")
		return // Return with all -1
	}

	// First, try to directly find events by name
	launchIdx = FindEventIndex(eventsData, "Launch")
	railExitIdx = FindEventIndex(eventsData, "Rail Exit")
	burnoutIdx = FindEventIndex(eventsData, "Motor Burnout")
	apogeeEventIdx = FindEventIndex(eventsData, "Apogee")
	touchdownIdx = FindEventIndex(eventsData, "Touchdown")

	// If we didn't find all events, try looser matching (case-insensitive)
	for i, row := range eventsData {
		if len(row) < 1 {
			continue
		}

		eventName := row[0]
		switch {
		case strings.EqualFold(eventName, "Launch"):
			if launchIdx == -1 {
				launchIdx = i
			}
		case strings.EqualFold(eventName, "Motor Burnout"):
			if burnoutIdx == -1 {
				burnoutIdx = i
			}
		case strings.EqualFold(eventName, EventRailExit):
			if railExitIdx == -1 {
				railExitIdx = FindEventIndex(eventsData, EventRailExit)
			}
		case strings.EqualFold(eventName, "Apogee"):
			if apogeeEventIdx == -1 {
				apogeeEventIdx = i
			}
		case strings.EqualFold(eventName, "Touchdown"):
			if touchdownIdx == -1 {
				touchdownIdx = i
			}
		}
	}
	return launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx // Explicitly return the values
}

// FindParachuteEvents searches the event data for parachute deployment events with DEPLOYED status
func FindParachuteEvents(eventsData [][]string, log *logf.Logger) []RecoverySystemData {
	if len(eventsData) == 0 {
		log.Warn("No event data available for finding parachute events.")
		return nil
	}

	// Find column indices once at the beginning
	timeIdx, eventNameIdx, statusIdx, parachuteStatusIdx, parachuteTypeIdx := findColumnIndices(eventsData[0], log)

	// Map to store unique parachute types and their deployment times
	parachuteMap := make(map[string]RecoverySystemData)

	// Process each row of event data (skip header row)
	for i := 1; i < len(eventsData); i++ {
		row := eventsData[i]
		if len(row) <= timeIdx { // Need at least time value
			continue
		}

		// Parse deployment time
		deploymentTime, ok := parseDeploymentTime(row[timeIdx], log)
		if !ok {
			continue
		}

		// Process each detection method in order
		if processParachuteStatusColumn(row, parachuteStatusIdx, eventNameIdx, parachuteTypeIdx, deploymentTime, log, parachuteMap) {
			continue // Skip other methods if already found
		}

		if processParachuteEventName(row, eventNameIdx, statusIdx, deploymentTime, log, parachuteMap) {
			continue // Skip next method if already found
		}

		processParachuteInColumnValues(row, eventsData[0], timeIdx, eventNameIdx, statusIdx, deploymentTime, log, parachuteMap)
	}

	return sortRecoverySystems(parachuteMap, log)
}

// processParachuteStatusColumn checks for parachute info in a dedicated status column
func processParachuteStatusColumn(row []string, parachuteStatusIdx, eventNameIdx, parachuteTypeIdx int,
	deploymentTime float64, log *logf.Logger, parachuteMap map[string]RecoverySystemData) bool {

	// Skip if we don't have a dedicated parachute status column
	if parachuteStatusIdx < 0 || parachuteStatusIdx >= len(row) {
		return false
	}

	statusValue := strings.ToUpper(strings.TrimSpace(row[parachuteStatusIdx]))
	if statusValue != StatusDeployed {
		return false
	}

	// Determine parachute type from available data
	parachuteType := determineParachuteType(row, eventNameIdx, parachuteTypeIdx, "Parachute")
	processParachuteDeployment(parachuteType, deploymentTime, log, parachuteMap)
	return true
}

// processParachuteEventName looks for parachute info in event name
func processParachuteEventName(row []string, eventNameIdx, statusIdx int, deploymentTime float64,
	log *logf.Logger, parachuteMap map[string]RecoverySystemData) bool {

	// Skip if we don't have event name column
	if eventNameIdx < 0 || eventNameIdx >= len(row) {
		return false
	}

	eventName := strings.ToLower(row[eventNameIdx])
	if !strings.Contains(eventName, "parachute") && !strings.Contains(eventName, "recovery") {
		return false
	}

	// Check if deployed status
	deployed := true // Assume deployed if not specified
	if statusIdx >= 0 && statusIdx < len(row) {
		deployed = isDeployedStatus(row[statusIdx])
	}

	if !deployed {
		return false
	}

	// Determine parachute type from event name
	parachuteType := "Parachute"
	if strings.Contains(eventName, "drogue") {
		parachuteType = RecoverySystemDrogue
	} else if strings.Contains(eventName, "main") {
		parachuteType = RecoverySystemMain
	}

	log.Info("Found parachute event by name", "type", parachuteType, "time", deploymentTime)
	processParachuteDeployment(parachuteType, deploymentTime, log, parachuteMap)
	return true
}

// processParachuteInColumnValues searches all columns for parachute status information
func processParachuteInColumnValues(row, headers []string, timeIdx, eventNameIdx, statusIdx int,
	deploymentTime float64, log *logf.Logger, parachuteMap map[string]RecoverySystemData) bool {

	found := false

	for colIdx, colValue := range row {
		// Skip already processed columns
		if colIdx == timeIdx || colIdx == eventNameIdx || colIdx == statusIdx {
			continue
		}

		// Check if this column has parachute status info
		colValueLower := strings.ToLower(colValue)
		if !strings.Contains(colValueLower, "parachute") || !strings.Contains(colValueLower, "status") {
			continue
		}

		// Check for DEPLOYED status
		if !isDeployedStatus(colValue) {
			continue
		}

		// Try to determine parachute type
		parachuteType := "Parachute"
		for j, header := range headers {
			headerLower := strings.ToLower(header)
			if strings.Contains(headerLower, "type") && j < len(row) {
				typeVal := strings.TrimSpace(row[j])
				if typeVal != "" {
					parachuteType = typeVal
					break
				}
			}
		}

		log.Info("Found parachute with status in columns", "type", parachuteType, "time", deploymentTime)
		processParachuteDeployment(parachuteType, deploymentTime, log, parachuteMap)
		found = true
	}

	return found
}

// sortRecoverySystems converts the map to a sorted slice
func sortRecoverySystems(parachuteMap map[string]RecoverySystemData, log *logf.Logger) []RecoverySystemData {
	var recoverySystems []RecoverySystemData
	for _, system := range parachuteMap {
		recoverySystems = append(recoverySystems, system)
	}

	// Sort by deployment time
	sort.Slice(recoverySystems, func(i, j int) bool {
		return recoverySystems[i].Deployment < recoverySystems[j].Deployment
	})

	log.Info("Found recovery systems", "count", len(recoverySystems))
	return recoverySystems
}

// ExtractMotionPoints converts raw motion data records to structured motionPoint slices.
func ExtractMotionPoints(motionData []*PlotSimRecord, motionHeaders []string, timeIdx, altitudeIdx, velocityIdx, accelIdx int, log *logf.Logger) []motionPoint {
	points := make([]motionPoint, 0, len(motionData))

	if timeIdx == -1 {
		log.Error("Cannot extract motion points: 'Time' column index not found.")
		return points
	}

	for i, record := range motionData {
		mp := motionPoint{}
		timeVal, ok := GetFloat64Value(*record, motionHeaders[timeIdx])
		if !ok {
			log.Warn("Failed to parse time value from motion data record", "recordIndex", i, "record", record)
			continue
		}
		mp.Time = timeVal

		if altitudeIdx != -1 {
			mp.AltAGL, _ = GetFloat64Value(*record, motionHeaders[altitudeIdx])
		}
		if velocityIdx != -1 {
			mp.VelTot, _ = GetFloat64Value(*record, motionHeaders[velocityIdx])
		}
		if accelIdx != -1 {
			mp.AccTot, _ = GetFloat64Value(*record, motionHeaders[accelIdx])
		}
		points = append(points, mp)
	}
	return points
}

// GetFloat64Value safely extracts a float64 value from a PlotSimRecord.
func GetFloat64Value(record PlotSimRecord, key string) (float64, bool) {
	val, exists := record[key]
	if !exists {
		return 0, false
	}
	fVal, ok := val.(float64)
	return fVal, ok
}

// FindPeakValues iterates through motion points to find max altitude, speed, and acceleration.
func FindPeakValues(motionPoints []motionPoint) (maxAlt, maxSpeed, maxAccel, timeAtApogee float64) {
	for _, p := range motionPoints {
		if p.AltAGL > maxAlt {
			maxAlt = p.AltAGL
			timeAtApogee = p.Time // Update timeAtApogee when new maxAlt is found
		}
		if p.VelTot > maxSpeed {
			maxSpeed = p.VelTot
		}
		if math.Abs(p.AccTot) > maxAccel { // Consider magnitude for max acceleration
			maxAccel = math.Abs(p.AccTot)
		}
	}
	return
}

// CalculateFlightTime calculates total flight time from event data.
func CalculateFlightTime(eventsData [][]string, launchIdx, touchdownIdx int) float64 {
	if launchIdx == -1 || touchdownIdx == -1 || launchIdx >= len(eventsData) || touchdownIdx >= len(eventsData) {
		return -1 // Invalid indices
	}
	launchTimeStr := eventsData[launchIdx][1] // Assuming time is at index 1
	touchdownTimeStr := eventsData[touchdownIdx][1]

	launchTime, errLaunch := strconv.ParseFloat(launchTimeStr, 64)
	touchdownTime, errTouchdown := strconv.ParseFloat(touchdownTimeStr, 64)

	if errLaunch != nil || errTouchdown != nil {
		return -1 // Error parsing times
	}

	return touchdownTime - launchTime
}

// FindClosestMotionPoint finds the motion point closest to a given timestamp.
func FindClosestMotionPoint(motionPoints []motionPoint, timestamp float64) *motionPoint {
	if len(motionPoints) == 0 {
		return nil
	}

	closest := &motionPoints[0]
	minDiff := math.Abs(motionPoints[0].Time - timestamp)

	for i := 1; i < len(motionPoints); i++ {
		diff := math.Abs(motionPoints[i].Time - timestamp)
		if diff < minDiff {
			minDiff = diff
			closest = &motionPoints[i]
		}
	}
	return closest
}

// GenerateReportData creates a new ReportData struct, populating it with information
// from the specified recordID, configuration, and storage backend.
// It now accepts a RecordManager interface instead of a concrete *storage.RecordManager.
func GenerateReportData(log *logf.Logger, cfg *config.Config, rm RecordManager, recordID string) (*ReportData, error) {
	log.Info("Generating report data", "recordID", recordID)

	// Get the record using the RecordManager interface
	record, err := rm.GetRecord(recordID)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("Record not found for report generation", "recordID", recordID, "error", err)
			return nil, storage.ErrRecordNotFound // Propagate the specific error
		}
		log.Error("Failed to get record for report", "recordID", recordID, "error", err)
		return nil, fmt.Errorf("failed to retrieve record %s: %w", recordID, err)
	}
	// Ensure record and its necessary components are not nil
	if record == nil {
		log.Error("Retrieved record is nil", "recordID", recordID)
		return nil, fmt.Errorf("retrieved record %s is nil", recordID)
	}

	// Load the simulation config that was used when this simulation was run
	// First, check if engine_config.json exists in the record's directory
	simConfig := cfg // Default to current config if no stored config found
	engineConfigPath := filepath.Join(record.Path, "engine_config.json")

	if _, err := os.Stat(engineConfigPath); err == nil {
		// Engine config file exists, try to load it
		configData, err := os.ReadFile(engineConfigPath)
		if err == nil {
			// Successfully read the file
			log.Info("Found stored engine configuration file", "path", engineConfigPath)

			// Parse the JSON into a config struct
			var storedConfig config.Config
			if err := json.Unmarshal(configData, &storedConfig); err == nil {
				simConfig = &storedConfig
				log.Info("Using stored engine configuration for report generation")
			} else {
				log.Warn("Failed to parse stored engine configuration, using current config", "error", err)
			}
		} else {
			log.Warn("Failed to read stored engine configuration file, using current config", "error", err)
		}
	} else {
		log.Warn("No stored engine configuration found, using current config", "path", engineConfigPath)
	}

	// Initialize SimulationData struct
	simData := &storage.SimulationData{}

	// Attempt to read OpenRocket document if available
	orkFilePath := filepath.Join(record.Path, "simulation.ork")
	if _, err := os.Stat(orkFilePath); err == nil {
		// Use openrocket.Load with file path and version from config
		orkDoc, loadErr := openrocket.Load(orkFilePath, cfg.Engine.External.OpenRocketVersion)
		if loadErr != nil {
			log.Warn("Failed to load .ork file, proceeding without it", "path", orkFilePath, "error", loadErr)
		} else {
			simData.ORKDoc = orkDoc
			log.Info("Successfully parsed .ork file", "path", orkFilePath)
		}
	} else {
		log.Info(".ork file not found, proceeding without it", "path", orkFilePath)
	}

	// Load data from CSV files (Motion, Events, Dynamics)
	if record.Motion != nil {
		simData.MotionHeaders, simData.MotionData, err = record.Motion.ReadHeadersAndData()
		if err != nil {
			log.Error("Failed to read motion data", "recordID", recordID, "error", err)
			// Depending on requirements, this might be a critical error. For now, log and continue.
		} else {
			log.Info("Successfully read motion data", "recordID", recordID, "headers_count", len(simData.MotionHeaders), "data_rows", len(simData.MotionData))
		}
	} else {
		log.Warn("Motion data store is nil for record", "recordID", recordID)
	}

	// Prepare data for plotting (parsed from motion.csv)
	var parsedMotionPlotRecords []*PlotSimRecord
	var parsedMotionPlotHeaders []string

	// If we have raw motion data (already loaded into simData.MotionData and simData.MotionHeaders),
	// parse it for plotting using the refactored parseSimData.
	if record.Motion != nil && len(simData.MotionData) > 0 && len(simData.MotionHeaders) > 0 {
		log.Debug("Attempting to parse pre-loaded motion data for plotting", "num_headers", len(simData.MotionHeaders), "num_rows", len(simData.MotionData))
		var parseErr error
		parsedMotionPlotRecords, parseErr = parseSimData(log, simData.MotionHeaders, simData.MotionData)
		if parseErr != nil {
			log.Error("Failed to parse motion data from pre-loaded content", "error", parseErr)
			// Depending on requirements, this might be a critical error or just mean no plots.
		} else {
			parsedMotionPlotHeaders = simData.MotionHeaders // Headers are passed directly
			log.Info("Successfully parsed pre-loaded motion data into PlotSimRecord format", "num_records", len(parsedMotionPlotRecords), "num_headers", len(parsedMotionPlotHeaders))
		}
	} else {
		log.Warn("Skipping parsing of motion data for plotting due to missing pre-loaded data or headers",
			"recordID", recordID,
			"motion_store_nil", record.Motion == nil,
			"motion_data_empty", len(simData.MotionData) == 0,
			"motion_headers_empty", len(simData.MotionHeaders) == 0)
	}

	if record.Events != nil {
		simData.EventsHeaders, simData.EventsData, err = record.Events.ReadHeadersAndData()
		if err != nil {
			log.Error("Failed to read events data", "recordID", recordID, "error", err)
			// Continue if possible, or return error
		} else {
			log.Info("Successfully read events data", "recordID", recordID, "headers_count", len(simData.EventsHeaders), "data_rows", len(simData.EventsData))
		}
	} else {
		log.Warn("Events data store is nil for record", "recordID", recordID)
	}

	// Populate ReportSummary from the loaded simulation config and placeholders
	summary := ReportSummary{
		RocketName:       simConfig.Setup.App.Name, // Default to app name from the sim config
		MotorDesignation: simConfig.Engine.Options.MotorDesignation,
		LaunchSite:       fmt.Sprintf("Lat: %.4f, Lon: %.4f, Alt: %.1fm", simConfig.Engine.Options.Launchsite.Latitude, simConfig.Engine.Options.Launchsite.Longitude, simConfig.Engine.Options.Launchsite.Altitude),
	}

	// Handle TargetApogeeFt - check if it exists in config
	if targetApogee, exists := getTargetApogeeFromConfig(simConfig); exists {
		summary.TargetApogeeFt = targetApogee
		log.Info("Found target apogee", "target_apogee_ft", targetApogee)
	} else {
		// Default to 0 if not available
		summary.TargetApogeeFt = 0
		log.Info("No target apogee found in config, using default value")
	}

	// Calculate liftoff mass from simulation data
	liftoffMass := calculateLiftoffMass(simData, parsedMotionPlotRecords, parsedMotionPlotHeaders)
	summary.LiftoffMassKg = liftoffMass
	log.Info("Calculated liftoff mass", "mass_kg", liftoffMass)

	if simData.ORKDoc != nil && simData.ORKDoc.Rocket.Name != "" {
		summary.RocketName = simData.ORKDoc.Rocket.Name
	}

	// Calculate motion metrics if we have motion data
	var motionMetrics *MotionMetrics
	if len(parsedMotionPlotRecords) > 0 && len(parsedMotionPlotHeaders) > 0 {
		// Use the existing CalculateMotionMetrics function
		// Default to 1.0m if launch rail length not specified
		launchRailLength := 1.0
		// Try to get launch rail length from config if available
		if cfg.Engine.Options.Launchrail.Length > 0 {
			launchRailLength = cfg.Engine.Options.Launchrail.Length
		}
		motionMetrics = CalculateMotionMetrics(parsedMotionPlotRecords, parsedMotionPlotHeaders, simData.EventsData, launchRailLength, log)
		log.Info("Calculated motion metrics", "maxAltitude", motionMetrics.MaxAltitudeAGL, "maxSpeed", motionMetrics.MaxSpeed)
	} else {
		log.Warn("No motion data available for metrics calculation")
		motionMetrics = &MotionMetrics{} // Initialize with empty values
	}

	// Try to load real motor data from the ThrustCurve API if a motor designation is available
	var motorSummary MotorSummaryData
	var motorData []*PlotSimRecord = nil
	var motorHeaders []string = []string{"Time (s)", "Thrust (N)"}

	if cfg.Engine.Options.MotorDesignation != "" {
		log.Info("Attempting to load motor data from ThrustCurve API", "designation", cfg.Engine.Options.MotorDesignation)

		// Create an HTTP client for the thrustcurves package
		httpClient := http_client.NewHTTPClient()

		// Load motor data from ThrustCurve API
		motorProps, err := thrustcurves.Load(cfg.Engine.Options.MotorDesignation, httpClient, *log)
		if err == nil && motorProps != nil {
			log.Info("Successfully loaded motor data from ThrustCurve API",
				"designation", cfg.Engine.Options.MotorDesignation,
				"motorID", motorProps.ID,
				"thrust_points", len(motorProps.Thrust))

			// Convert the thrust curve to PlotSimRecord format
			motorData = make([]*PlotSimRecord, len(motorProps.Thrust))
			for i, point := range motorProps.Thrust {
				record := make(PlotSimRecord)
				record["Time (s)"] = point[0]
				record["Thrust (N)"] = point[1]
				motorData[i] = &record
			}

			// Create a motor summary from the loaded data
			motorSummary = MotorSummaryData{
				MaxThrust:         motorProps.MaxThrust,
				BurnTime:          motorProps.BurnTime,
				TotalImpulse:      motorProps.TotalImpulse,
				AvgThrust:         motorProps.AvgThrust,
				PropellantMass:    motorProps.WetMass,
				MotorClass:        string(motorProps.Designation)[0:1], // Get class from first letter of designation
				ThrustEfficiency:  motorProps.AvgThrust / motorProps.MaxThrust,
				MotorManufacturer: "ThrustCurve", // Default, ideally this would come from the API
			}

			// Calculate specific impulse if propellant mass is available
			if motorProps.WetMass > 0 {
				const g0 = 9.80665 // Standard gravity (m/s²)
				motorSummary.SpecificImpulse = motorProps.TotalImpulse / (motorProps.WetMass * g0)
			}

			log.Info("Populated motor summary from ThrustCurve API data",
				"maxThrust", motorSummary.MaxThrust,
				"burnTime", motorSummary.BurnTime,
				"totalImpulse", motorSummary.TotalImpulse,
				"propellantMass", motorSummary.PropellantMass,
				"motorClass", motorSummary.MotorClass)
		} else {
			log.Warn("Failed to load motor data from ThrustCurve API, falling back to calculated summary",
				"designation", cfg.Engine.Options.MotorDesignation,
				"error", err)
		}
	}

	// Fall back to calculating motor summary from simulation data if we couldn't get it from ThrustCurve
	if motorSummary.MaxThrust == 0 && len(parsedMotionPlotRecords) > 0 && len(parsedMotionPlotHeaders) > 0 && record.Motion != nil {
		log.Info("Using simulation data to calculate motor summary")
		motorSummary = calculateMotorSummary(parsedMotionPlotRecords, parsedMotionPlotHeaders, log)
		log.Info("Calculated motor metrics from simulation data", "maxThrust", motorSummary.MaxThrust, "burnTime", motorSummary.BurnTime, "totalImpulse", motorSummary.TotalImpulse)
	} else if motorSummary.MaxThrust == 0 {
		log.Warn("No motor data available for motor summary calculation")
	}

	// Find flight events
	var launchIdx, apogeeIdx, touchdownIdx int
	// Unused but declared to match function signature: railExitIdx, burnoutIdx
	var _, _ int
	if len(simData.EventsData) > 0 {
		launchIdx, _, _, apogeeIdx, touchdownIdx = FindFlightEvents(simData.EventsData, log)
		log.Info("Found flight events", "apogeeIdx", apogeeIdx, "launchIdx", launchIdx, "touchdownIdx", touchdownIdx)
	} else {
		log.Warn("No events data available for event extraction")
		launchIdx, apogeeIdx, touchdownIdx = -1, -1, -1
	}

	// Create phase summary data
	phaseSummary := PhaseSummaryData{}

	// If we have apogee data, add it to the phase summary
	if apogeeIdx >= 0 && apogeeIdx < len(simData.EventsData) {
		// Extract time from events data
		apogeeTime, err := strconv.ParseFloat(simData.EventsData[apogeeIdx][1], 64)
		if err == nil {
			phaseSummary.ApogeeTimeSec = apogeeTime
			phaseSummary.MaxAltitudeM = motionMetrics.MaxAltitudeAGL
			log.Info("Added apogee to phase summary", "time", apogeeTime, "altitude", motionMetrics.MaxAltitudeAGL)
		} else {
			log.Warn("Could not parse apogee time", "error", err)
		}
	}

	// We'll initialize an empty recovery systems slice for now
	// This can be enhanced in the future when recovery system data structure is better defined

	// Calculate flight time
	var flightTime float64
	if launchIdx >= 0 && touchdownIdx >= 0 && launchIdx < len(simData.EventsData) && touchdownIdx < len(simData.EventsData) {
		launchTime, err := strconv.ParseFloat(simData.EventsData[launchIdx][1], 64)
		if err == nil {
			touchdownTime, err := strconv.ParseFloat(simData.EventsData[touchdownIdx][1], 64)
			if err == nil {
				flightTime = touchdownTime - launchTime
				log.Info("Calculated flight time", "flightTime", flightTime)
			} else {
				log.Warn("Could not parse touchdown time", "error", err)
			}
		} else {
			log.Warn("Could not parse launch time", "error", err)
		}
	}

	// If we have calculated motion metrics but no flight time was found in events,
	// update the flight time in the motion metrics
	if motionMetrics != nil && flightTime > 0 && motionMetrics.FlightTime <= 0 {
		motionMetrics.FlightTime = flightTime
		log.Info("Updated motion metrics flight time from events", "flightTime", flightTime)
	}

	// Create comprehensive data for the report
	// 1. Create accurate launch conditions using location and physics values from the simulation config
	// Calculate local gravity at the launch site based on latitude and altitude
	latRad := simConfig.Engine.Options.Launchsite.Latitude * math.Pi / 180.0
	// Standard gravity formula based on latitude (ignoring altitude effects)
	localGravity := 9.780327 * (1 + 0.0053024*math.Pow(math.Sin(latRad), 2) - 0.0000058*math.Pow(math.Sin(2*latRad), 2))

	// Adjust for altitude (approximation)
	altitudeM := simConfig.Engine.Options.Launchsite.Altitude
	// Earth radius in meters
	earthRadiusM := 6371000.0
	localGravity = localGravity * math.Pow(earthRadiusM/(earthRadiusM+altitudeM), 2)

	// Get atmosphere config directly from the provided config
	isa := simConfig.Engine.Options.Launchsite.Atmosphere.ISAConfiguration

	// Calculate speed of sound at the pad using temperature
	// Use standard atmosphere temperature lapse rate if not specified
	temperatureK := 288.15 // Default temperature at sea level (15°C in Kelvin)
	if isa.SeaLevelTemperature > 0 {
		// Convert from Kelvin to Celsius if needed (assuming config might have K)
		if isa.SeaLevelTemperature > 100 {
			temperatureK = isa.SeaLevelTemperature
		} else {
			temperatureK = isa.SeaLevelTemperature + 273.15
		}
	}

	// Apply temperature lapse rate to altitude
	tempLapseRate := 0.0065 // Standard atmosphere lapse rate in K/m
	if isa.TemperatureLapseRate > 0 {
		tempLapseRate = isa.TemperatureLapseRate
	}

	// Adjust temperature for altitude
	temperatureK = temperatureK - tempLapseRate*altitudeM
	// Calculate speed of sound: c = sqrt(gamma * R * T)
	// Where gamma is ratio of specific heats, R is specific gas constant, T is temperature in Kelvin
	gamma := 1.4 // Ratio of specific heats for air
	if isa.RatioSpecificHeats > 0 {
		gamma = isa.RatioSpecificHeats
	}

	specificGasConstant := 287.05 // J/(kg·K) for dry air
	if isa.SpecificGasConstant > 0 {
		specificGasConstant = isa.SpecificGasConstant
	}

	speedOfSound := math.Sqrt(gamma * specificGasConstant * temperatureK)

	// Calculate atmospheric pressure at altitude using barometric formula
	seaLevelPressure := 101325.0 // Standard sea level pressure in Pa
	if isa.SeaLevelPressure > 0 {
		seaLevelPressure = isa.SeaLevelPressure
	}

	// Pressure formula: P = P0 * (1 - L*h/T0)^(g*M/R0*L)
	// where L is lapse rate, h is height, T0 is sea level temp, g is gravity
	// M is molar mass of air, R0 is universal gas constant
	pressureAtAltitude := seaLevelPressure * math.Pow(
		1-(tempLapseRate*altitudeM)/288.15,
		(localGravity*0.0289644)/(8.31447*tempLapseRate))

	// Calculate air density: ρ = P/(R*T)
	airDensity := pressureAtAltitude / (specificGasConstant * temperatureK)

	// Simplified weather model with comprehensive atmospheric properties
	// No wind direction modeling as requested

	weatherData := WeatherData{
		Latitude:      cfg.Engine.Options.Launchsite.Latitude,
		Longitude:     cfg.Engine.Options.Launchsite.Longitude,
		ElevationAMSL: altitudeM,
		// Set minimal wind data since the wind direction model was removed
		WindSpeed:        0.0,                // Set to zero as requested to remove defaulted fields
		Pressure:         pressureAtAltitude, // Atmospheric pressure at the launch site (Pa)
		SeaLevelPressure: seaLevelPressure,   // Sea level pressure reference (Pa)
		Density:          airDensity,         // Air density (kg/m³)
		LocalGravity:     localGravity,       // Local gravity (m/s²)
		SpeedOfSound:     speedOfSound,       // Speed of sound in air (m/s)
		TemperatureK:     temperatureK,       // Temperature in Kelvin
	}

	// 2. Create recovery systems data from actual parachute deployment events
	recoverySystems := FindParachuteEvents(simData.EventsData, log)

	// If no parachute events were found in the events data, infer them from motion data
	if len(recoverySystems) == 0 {
		log.Info("No parachute events found in simulation data, attempting to infer from motion data")

		// Get apogee time from motion metrics
		recoveryTimeDeployment := 0.0
		if motionMetrics != nil && motionMetrics.TimeToApogee > 0 {
			recoveryTimeDeployment = motionMetrics.TimeToApogee
		} else {
			// Try to find apogee from motion data
			apogeeTime, maxAlt := findApogeeFromMotionData(parsedMotionPlotRecords, parsedMotionPlotHeaders, log)
			if apogeeTime > 0 {
				recoveryTimeDeployment = apogeeTime
				log.Info("Found apogee from motion data", "time", apogeeTime, "altitude", maxAlt)
			} else {
				// Fallback to reasonable value based on typical rocket flight
				recoveryTimeDeployment = 10.0
				log.Warn("Could not determine apogee time, using default value", "default_time", recoveryTimeDeployment)
			}
		}

		// Calculate realistic descent rates based on motion data if available
		drogueDescentRate, mainDescentRate := calculateDescentRates(parsedMotionPlotRecords, parsedMotionPlotHeaders, recoveryTimeDeployment, log)

		recoverySystems = []RecoverySystemData{
			{
				Type:        RecoverySystemDrogue,
				Deployment:  recoveryTimeDeployment,
				DescentRate: drogueDescentRate,
			},
			{
				Type:        RecoverySystemMain,
				Deployment:  recoveryTimeDeployment + 5.0, // Typically deployed 5 seconds after drogue
				DescentRate: mainDescentRate,
			},
		}
	} else {
		log.Info("Using parachute events from simulation data", "count", len(recoverySystems))
	}

	// 3. Make sure phase summary has valid data
	if phaseSummary.ApogeeTimeSec <= 0 {
		if motionMetrics != nil && motionMetrics.TimeToApogee > 0 {
			phaseSummary.ApogeeTimeSec = motionMetrics.TimeToApogee
		} else {
			phaseSummary.ApogeeTimeSec = 11.5 // Sample value if no actual data
		}
	}

	if phaseSummary.MaxAltitudeM <= 0 {
		if motionMetrics != nil && motionMetrics.MaxAltitudeAGL > 0 {
			phaseSummary.MaxAltitudeM = motionMetrics.MaxAltitudeAGL
		} else {
			phaseSummary.MaxAltitudeM = 500.0 // Sample altitude if no actual data
		}
	}

	return &ReportData{
		RecordID:        recordID,
		Version:         cfg.Setup.App.Version,
		GeneratedTime:   time.Now().Format(time.RFC3339),
		Config:          *cfg,
		Summary:         summary,
		RawData:         simData,
		MotionData:      parsedMotionPlotRecords,
		MotionHeaders:   parsedMotionPlotHeaders,
		MotorData:       motorData,                           // Real motor thrust data from ThrustCurve API
		MotorHeaders:    motorHeaders,                        // Motor headers for the thrust data
		Plots:           make(map[string]string),             // To be populated by GeneratePlots
		Assets:          make(map[string]string),             // To be populated by PrepareReportAssets
		MotionMetrics:   motionMetrics,                       // Add calculated motion metrics
		MotorSummary:    motorSummary,                        // Add calculated motor summary from ThrustCurve API
		PhaseSummary:    phaseSummary,                        // Add flight phases with actual/sample data
		MotorName:       cfg.Engine.Options.MotorDesignation, // Motor name from config
		RocketName:      summary.RocketName,                  // Rocket name from summary
		Weather:         weatherData,                         // Add weather data
		RecoverySystems: recoverySystems,                     // Add recovery systems data
	}, nil
}
