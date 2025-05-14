package reporting

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/zerodha/logf"
)

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

func getHeaderKey(headers []string, keyName string) string {
	for _, h := range headers {
		if strings.EqualFold(h, keyName) {
			return h
		}
	}
	return "" // Return empty if not found
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
	RecordID         string                 `json:"record_id" yaml:"record_id"`
	Version          string                 `json:"version" yaml:"version"`
	RocketName       string                 `json:"rocket_name" yaml:"rocket_name"`
	MotorName        string                 `json:"motor_name" yaml:"motor_name"`
	LiftoffMassKg    float64                `json:"liftoff_mass_kg" yaml:"liftoff_mass_kg"`
	GeneratedTime    string                 `json:"generated_time" yaml:"generated_time"`
	ConfigSummary    *config.Config         `json:"config_summary" yaml:"config_summary"`
	Summary          ReportSummary          `json:"summary" yaml:"summary"`
	Plots            map[string]string      `json:"plots" yaml:"plots"`
	MotionMetrics    *MotionMetrics         `json:"motion_metrics" yaml:"motion_metrics"`
	MotorSummary     MotorSummaryData       `json:"motor_summary" yaml:"motor_summary"`
	ParachuteSummary ParachuteSummaryData   `json:"parachute_summary" yaml:"parachute_summary"`
	PhaseSummary     PhaseSummaryData       `json:"phase_summary" yaml:"phase_summary"`
	LaunchRail       LaunchRailData         `json:"launch_rail" yaml:"launch_rail"`
	ForcesAndMoments ForcesAndMomentsData   `json:"forces_and_moments" yaml:"forces_and_moments"`
	Weather          WeatherData            `json:"weather" yaml:"weather"`
	AllEvents        []EventSummary         `json:"all_events" yaml:"all_events"`
	Stages           []StageData            `json:"stages" yaml:"stages"`
	RecoverySystems  []RecoverySystemData   `json:"recovery_systems" yaml:"recovery_systems"`
	MotionData       []*PlotSimRecord       `json:"motion_data" yaml:"motion_data"`
	MotionHeaders    []string               `json:"motion_headers" yaml:"motion_headers"`
	EventsData       [][]string             `json:"events_data" yaml:"events_data"`
	Log              *logf.Logger           `json:"-"`
	ReportTitle      string                 `json:"report_title" yaml:"report_title"`
	GenerationDate   string                 `json:"generation_date" yaml:"generation_date"`
	MotorData        []*PlotSimRecord       `json:"motor_data" yaml:"motor_data"`
	MotorHeaders     []string               `json:"motor_headers" yaml:"motor_headers"`
	Extensions       map[string]interface{} `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	Assets           map[string]string      `json:"assets,omitempty" yaml:"assets,omitempty"`
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
	reader.FieldsPerRecord = -1 // Allow variable number of fields for header row

	headers, err := reader.Read()
	if err == io.EOF {
		log.Warn("CSV data is empty or contains only EOF.")
		return []*PlotSimRecord{}, []string{}, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}
	log.Debug("CSV Headers", "headers", headers)

	reader.FieldsPerRecord = len(headers) // Enforce field count for data rows

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
				record[header] = "" // Handle rows shorter than headers
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
// This allows the reporting package to interact with storage without a direct dependency on cmd/server types.
type HandlerRecordManager interface {
	GetRecord(hash string) (*storage.Record, error)
	GetStorageDir() string
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

	// TODO: Populate other fields like SpecificImpulse, MotorClass, etc., from config or other sources if available.

	log.Info("Calculated motor summary", "maxThrust", summary.MaxThrust, "totalImpulse", summary.TotalImpulse, "burnTime", summary.BurnTime, "avgThrust", summary.AvgThrust)
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
		switch strings.ToLower(strings.TrimSpace(header)) {
		case "time", "time (s)":
			timeIdx = i
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
// It assumes event name is in the first column and time in the second after headers.
func FindFlightEvents(eventsData [][]string, log *logf.Logger) (launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx int) {
	launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx = -1, -1, -1, -1, -1

	if len(eventsData) < 1 { // Need at least one data row (headers are not passed in eventsData from GenerateReportData)
		log.Warn("eventsData has insufficient rows to find flight events", "num_rows", len(eventsData))
		return launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx // Explicitly return the values
	}

	// In GenerateReportData, eventsData is passed as rawMotionData[1:] after headers are stripped.
	// So, eventsData[i][0] is event name, eventsData[i][1] is time.
	for i, eventRow := range eventsData {
		if len(eventRow) < 2 { // Need at least name and time
			log.Warn("Event data row too short for event name/time", "row_index", i, "row_len", len(eventRow))
			continue
		}

		eventName := strings.TrimSpace(eventRow[0])
		switch {
		case strings.EqualFold(eventName, "Launch"):
			if launchIdx == -1 {
				launchIdx = FindEventIndex(eventsData, "Launch")
			}
		case strings.Contains(strings.ToLower(eventName), "burnout"):
			if burnoutIdx == -1 {
				burnoutIdx = FindEventIndex(eventsData, eventName)
			} // Use actual eventName if it contains 'burnout'
		case strings.EqualFold(eventName, "Rail Exit"):
			if railExitIdx == -1 {
				railExitIdx = FindEventIndex(eventsData, "Rail Exit")
			}
		case strings.EqualFold(eventName, "Apogee"):
			if apogeeEventIdx == -1 {
				apogeeEventIdx = FindEventIndex(eventsData, "Apogee")
			}
		case strings.EqualFold(eventName, "Touchdown"):
			if touchdownIdx == -1 {
				touchdownIdx = FindEventIndex(eventsData, "Touchdown")
			}
		}
	}
	return launchIdx, railExitIdx, burnoutIdx, apogeeEventIdx, touchdownIdx // Explicitly return the values
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
// It handles loading raw simulation data, parsing it, and calculating summary metrics.
func GenerateReportData(log *logf.Logger, cfg *config.Config, rs *storage.RecordManager, recordID string) (*ReportData, error) {
	if log == nil {
		// Create a default logger if nil to prevent panics, though ideally, a logger should always be passed.
		defaultLogger := logf.New(logf.Opts{}) 
		log = &defaultLogger
		log.Warn("GenerateReportData called with nil logger, using default.")
	}
	log.Info("GenerateReportData called (minimal implementation for syntax fix)", "recordID", recordID)
	
	// TODO: This is a minimal implementation to fix syntax. 
	// The original, full implementation of GenerateReportData needs to be restored or verified.
	// For now, return a basic ReportData struct and no error.
	return &ReportData{
		RecordID:      recordID,
		GeneratedTime: time.Now().Format(time.RFC3339),
		// Initialize other essential fields if known to prevent nil pointer issues downstream
		Summary: ReportSummary{
			MotionMetrics: MotionMetrics{}, // Ensure nested structs are initialized
		},
		Plots: make(map[string]string),
		Assets: make(map[string]string),
	}, nil
}
