package reporting

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/storage"
	logf "github.com/zerodha/logf"
)

// ReportData holds all the necessary information for generating a report.
// TODO: Consolidate redundant fields (e.g., LandingVelocity vs motionMetrics.LandingVelocityMPS)
type ReportData struct {
	Version            string // Application version
	RecordID           string
	AtmospherePlotPath string
	ThrustPlotPath     string
	TrajectoryPlotPath string
	DynamicsPlotPath   string
	GPSMapImagePath    string

	// Environmental Conditions
	LaunchSiteName  string
	LaunchLatitude  float64
	LaunchLongitude float64
	LaunchElevation float64 // meters AMSL
	WindSpeed       float64 // m/s at a reference altitude (e.g., 10m)
	WindDirection   float64 // degrees from North
	Temperature     float64 // Celsius at launch site
	Pressure        float64 // Pascals at launch site
	Humidity        float64 // Percentage

	// Flight Summary
	ApogeeMeters        float64
	MaxVelocityMPS      float64
	MaxAccelerationMPS2 float64
	TotalFlightTimeSec  float64
	LandingVelocityMPS  float64
	AllEvents           []FlightEvent

	// Event Summaries / Highlights
	MotorSummary     MotorHighlights
	ParachuteSummary ParachuteHighlights
	PhaseSummary     RocketPhaseHighlights

	// Landing Information
	LandingLatitude          float64
	LandingLongitude         float64
	LandingDistanceMeters    float64 // Distance from launch site
	LandingRadius95PctMeters float64 // Placeholder for 95th percentile landing radius
}

// MotorHighlights summarizes key motor-related events.
type MotorHighlights struct {
	IgnitionTimeSec   float64
	BurnoutTimeSec    float64
	BurnDurationSec   float64
	HasMotorEvents    bool // True if any motor-specific events were identified
}

// ParachuteEventDetail provides specifics for a single parachute event.
type ParachuteEventDetail struct {
	Name                     string  // e.g., "Drogue Deployed", "Main Deployed"
	DeploymentTimeSec        float64
	DeploymentAltitudeMeters float64
	TimeToDeploySec          float64 // Relative to liftoff
}

// ParachuteHighlights summarizes parachute deployment events.
type ParachuteHighlights struct {
	Events               []ParachuteEventDetail
	HasParachuteEvents   bool // True if any parachute events were identified
}

// RocketPhaseHighlights summarizes major flight phases and events.
type RocketPhaseHighlights struct {
	LiftoffTimeSec      float64 // From motion data or "Liftoff" event
	ApogeeTimeSec       float64 // From motion data or "Apogee" event
	LandingTimeSec      float64 // From "Landing" event
	CoastStartTimeSec   float64 // Typically after motor burnout
	CoastEndTimeSec     float64 // Typically at apogee or first parachute deployment
	CoastDurationSec    float64
	HasLiftoffEvent     bool
	HasApogeeEvent      bool
	HasLandingEvent     bool
}

// FlightEvent represents a significant event during the simulation.
// It can be used to populate tables in the report.
type FlightEvent struct {
	Name           string
	TimeSec        float64
	AltitudeMeters float64
	// Add other relevant data like velocity, status, etc.
}

// Generator handles report generation using text/template.
type Generator struct {
	template *template.Template
}

// NewGenerator creates a new report generator by reading and parsing the template file.
func NewGenerator() (*Generator, error) {
	// Assuming report.md.tmpl is now in the same directory (internal/reporting)
	templatePath := "internal/reporting/report.md.tmpl"
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		// Attempt to read from relative path if absolute fails (e.g., during tests)
		wd, _ := os.Getwd()
		altPath := filepath.Join(wd, templatePath)
		tmplBytes, err = os.ReadFile(altPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read markdown template file from %s or %s: %w", templatePath, altPath, err)
		}
	}

	tmpl, err := template.New("report").Parse(string(tmplBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse markdown template file %s: %w", templatePath, err)
	}
	return &Generator{
		template: tmpl,
	}, nil
}

// GenerateMarkdownFile creates the markdown report content and saves it to a file.
func (g *Generator) GenerateMarkdownFile(data ReportData, outputDir string) error {
	var mdOutput bytes.Buffer
	if err := g.template.Execute(&mdOutput, data); err != nil {
		return fmt.Errorf("failed to execute markdown template: %w", err)
	}

	mdFilePath := filepath.Join(outputDir, "report.md")
	if err := os.WriteFile(mdFilePath, mdOutput.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write markdown report to %s: %w", mdFilePath, err)
	}
	return nil
}

// createDummyAsset creates a minimal 1x1 transparent PNG at the given path.
// In a real scenario, this would generate actual plot images.
func createDummyAsset(assetPath string) error {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.Transparent)

	f, err := os.Create(assetPath)
	if err != nil {
		return fmt.Errorf("failed to create dummy asset file %s: %w", assetPath, err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("failed to encode dummy PNG %s: %w", assetPath, err)
	}
	return nil
}

// LoadSimulationData loads the necessary data for a report from storage and creates dummy assets.
func LoadSimulationData(rm *storage.RecordManager, recordID string, reportSpecificDir string, log *logf.Logger) (ReportData, error) {
	log.Debug("Attempting to load simulation data for record", "recordID", recordID)

	record, err := rm.GetRecord(recordID)
	if err != nil {
		log.Error("Failed to get record for report data", "recordID", recordID, "error", err)
		return ReportData{}, fmt.Errorf("failed to load record %s: %w", recordID, err)
	}
	defer record.Close() // Ensure record files are closed

	log.Info("Loaded record for report", "recordID", recordID, "creationTime", record.CreationTime)

	data := ReportData{
		RecordID:                 record.Hash,
		Version:                  "v0.0.0-dev",
		LaunchSiteName:           "Default Launch Site",
		LaunchLatitude:           37.7749,
		LaunchLongitude:          -122.4194,
		LaunchElevation:          10.0,
		WindSpeed:                5.0,
		WindDirection:            270.0,
		Temperature:              20.0,
		Pressure:                 101325.0,
		Humidity:                 60.0,
		LandingLatitude:          37.7750,
		LandingLongitude:         -122.4200,
		LandingDistanceMeters:    100.0,
		LandingRadius95PctMeters: 50.0,
	}

	// --- Process MOTION.csv ---
	if record.Motion != nil {
		allMotionData, err := record.Motion.ReadAll()
		if err != nil {
			log.Error("Failed to read all data from MOTION.csv", "recordID", recordID, "error", err)
			return ReportData{}, fmt.Errorf("failed to read motion data for %s: %w", recordID, err)
		}
		log.Debug("Raw motion data for record %s (first 5 rows)", "recordID", recordID, "data", allMotionData[:min(5, len(allMotionData))])
		if len(allMotionData) < 2 { // Need at least headers and one data row
			log.Error("MOTION.csv is empty or contains only headers", "recordID", recordID)
			return ReportData{}, fmt.Errorf("motion data is insufficient for record %s (rows: %d)", recordID, len(allMotionData))
		}
		motionMetrics, err := parseMotionData(allMotionData, data.LaunchElevation, log)
		if err != nil {
			log.Error("Failed to parse MOTION.csv data", "recordID", recordID, "error", err)
			return ReportData{}, fmt.Errorf("failed to parse motion data for %s: %w", recordID, err)
		}
		data.ApogeeMeters = motionMetrics.ApogeeMeters
		data.MaxVelocityMPS = motionMetrics.MaxVelocityMPS
		data.MaxAccelerationMPS2 = motionMetrics.MaxAccelerationMPS2
		data.TotalFlightTimeSec = motionMetrics.TotalFlightTimeSec
		data.LandingVelocityMPS = motionMetrics.LandingVelocityMPS
		log.Debug("Parsed motion data for record %s", "recordID", recordID, "apogeeM", data.ApogeeMeters, "maxVeloMPS", data.MaxVelocityMPS, "flightTimeS", data.TotalFlightTimeSec)

	} else {
		log.Error("MOTION.csv storage not available for record", "recordID", recordID)
		return ReportData{}, fmt.Errorf("motion storage not available for record %s", recordID)
	}

	// --- Process EVENTS.csv ---
	if record.Events != nil {
		allEventsData, err := record.Events.ReadAll()
		if err != nil {
			log.Error("Failed to read all data from EVENTS.csv", "recordID", recordID, "error", err)
			// Decide whether to return error or continue without events
			// Continuing for now, highlights will be empty/defaulted
		} else {
			log.Debug("Raw events data for record %s (first 5 rows)", "recordID", recordID, "data", allEventsData[:min(5, len(allEventsData))])
			if len(allEventsData) == 0 { // Completely empty file
				log.Warn("EVENTS.csv is empty", "recordID", recordID)
			} else if len(allEventsData) == 1 { // Only headers, no data rows
				log.Warn("EVENTS.csv contains only headers", "recordID", recordID)
			} else { // Has headers and data
				headers := allEventsData[0]
				dataRows := allEventsData[1:]

				colIndices := make(map[string]int)
				for i, header := range headers {
					colIndices[strings.ToLower(strings.TrimSpace(header))] = i
				}

				requiredCols := []string{"time", "event_name"}
				var missingCols []string
				for _, colName := range requiredCols {
					if _, ok := colIndices[colName]; !ok {
						missingCols = append(missingCols, colName)
					}
				}
				if len(missingCols) > 0 {
					log.Error("EVENTS.csv missing required columns", "recordID", recordID, "missing", missingCols, "available", headers)
					// Decide whether to return error or continue without events
					// Continuing for now, highlights will be empty/defaulted
				} else {
					timeIdx := colIndices["time"]
					eventNameIdx := colIndices["event_name"]
					altIdx, altColExists := colIndices["altitude"]

					for i, row := range dataRows {
						if len(row) <= timeIdx || len(row) <= eventNameIdx || (altColExists && len(row) <= altIdx) {
							log.Warn("Skipping malformed row in EVENTS.csv", "recordID", recordID, "rowIndex", i+1, "rowLength", len(row), "requiredIndices", []int{timeIdx, eventNameIdx, altIdx})
							continue
						}
						timeStr := row[timeIdx]
						eventName := row[eventNameIdx]

						time, err := strconv.ParseFloat(timeStr, 64)
						if err != nil {
							log.Warn("Failed to parse time in EVENTS.csv row, skipping event", "recordID", recordID, "rowIndex", i+1, "value", timeStr, "error", err)
							continue // Skip this event
						}

						altitude := 0.0
						if altColExists {
							altStr := row[altIdx]
							altVal, parseErr := strconv.ParseFloat(altStr, 64)
							if parseErr == nil {
								altitude = altVal
							} else {
								// Log quietly if altitude parsing fails, use 0.0
								log.Debug("Failed to parse altitude for event, using 0.0", "recordID", recordID, "rowIndex", i+1, "event", eventName, "value", altStr, "error", parseErr)
							}
						}

						data.AllEvents = append(data.AllEvents, FlightEvent{Name: eventName, TimeSec: time, AltitudeMeters: altitude})
					}
				}
			}
		}
	} else {
		log.Error("EVENTS.csv storage not available for record", "recordID", recordID)
		// Continue without events, highlights will be empty/defaulted
	}

	// --- Process Highlights & Placeholders ---
	data.MotorSummary, data.ParachuteSummary, data.PhaseSummary = processEventHighlights(data.AllEvents, parsedMotionMetrics{ApogeeMeters: data.ApogeeMeters, MaxVelocityMPS: data.MaxVelocityMPS, MaxAccelerationMPS2: data.MaxAccelerationMPS2, TotalFlightTimeSec: data.TotalFlightTimeSec, LandingVelocityMPS: data.LandingVelocityMPS}, data.LaunchElevation)

	// If AllEvents is still empty after trying to parse (or if file was empty/headers only), add placeholders
	if len(data.AllEvents) == 0 {
		log.Info("No key flight events parsed or available, using placeholders", "recordID", recordID)
		// Ensure ApogeeMeters and TotalFlightTimeSec are not zero to avoid division by zero or nonsensical time
		apogeeTimeEstimate := 0.0
		if data.ApogeeMeters > 0 && data.MaxVelocityMPS > 0 { // Basic sanity check for apogee time estimation
			apogeeTimeEstimate = data.ApogeeMeters / (data.MaxVelocityMPS / 2) // Very rough estimate
			if apogeeTimeEstimate > data.TotalFlightTimeSec && data.TotalFlightTimeSec > 0 {
				apogeeTimeEstimate = data.TotalFlightTimeSec / 2
			}
		}
		data.AllEvents = []FlightEvent{
			{"Liftoff (placeholder)", 0.0, data.LaunchElevation},
			{"Apogee (placeholder)", math.Max(0, apogeeTimeEstimate), data.ApogeeMeters},
			{"Landing (placeholder)", math.Max(0, data.TotalFlightTimeSec), data.LaunchElevation},
		}
	}

	// --- Create dummy plot assets ---
	assetSubDir := "assets"
	plotPaths := map[string]*string{
		"atmosphere_plot.svg":     &data.AtmospherePlotPath,
		"thrust_plot.svg":         &data.ThrustPlotPath,
		"trajectory_plot.svg":     &data.TrajectoryPlotPath,
		"dynamics_plot.svg":       &data.DynamicsPlotPath,
		"gps_map.svg":             &data.GPSMapImagePath,
	}

	for name, pathVar := range plotPaths {
		relPath := filepath.Join(assetSubDir, name)
		*pathVar = relPath
		if err := createDummyAsset(filepath.Join(reportSpecificDir, relPath)); err != nil {
			// Log error but continue, report might be partially useful
			log.Error("Failed to create dummy asset for report", "assetName", name, "recordID", recordID, "error", err)
		}
	}

	return data, nil
}

// Helper struct for parsed motion data
type parsedMotionMetrics struct {
	ApogeeMeters        float64
	MaxVelocityMPS      float64
	MaxAccelerationMPS2 float64
	TotalFlightTimeSec  float64
	LandingVelocityMPS  float64
}

func parseMotionData(allRows [][]string, launchElevation float64, log *logf.Logger) (parsedMotionMetrics, error) {
	metrics := parsedMotionMetrics{
		ApogeeMeters:        launchElevation,
		MaxVelocityMPS:      -1e9,
		MaxAccelerationMPS2: -1e9,
	}

	if len(allRows) < 2 { // Headers + at least one data row
		// Log this error using the passed-in logger if it's critical enough
		// log.Error("MOTION.csv data error", "message", "Requires at least a header and one data row", "rows_received", len(allRows))
		return metrics, fmt.Errorf("MOTION.csv data requires at least a header and one data row, got %d total rows", len(allRows))
	}

	headers := allRows[0]
	dataRows := allRows[1:]

	colIndices := make(map[string]int)
	for i, header := range headers {
		colIndices[strings.ToLower(strings.TrimSpace(header))] = i
	}

	requiredCols := []string{"time", "altitude", "velocity", "acceleration"}
	for _, colName := range requiredCols {
		if _, ok := colIndices[colName]; !ok {
			// log.Error("MOTION.csv missing column", "column", colName, "available_headers", headers)
			return metrics, fmt.Errorf("MOTION.csv missing required column: %s. Available: %v", colName, headers)
		}
	}

	timeIdx := colIndices["time"]
	altIdx := colIndices["altitude"]
	velIdx := colIndices["velocity"]
	accelIdx := colIndices["acceleration"]

	var lastTime, lastAltitude, lastVelocity float64
	apogeeReached := false

	log.Debug("Processing MOTION.csv data rows")
	for _, row := range dataRows {
		timeStr := strings.TrimSpace(row[timeIdx])
		altStr := strings.TrimSpace(row[altIdx])
		velStr := strings.TrimSpace(row[velIdx])
		accelStr := strings.TrimSpace(row[accelIdx])

		log.Debug("Parsing motion data row", "timeStr", timeStr, "altStr", altStr, "velStr", velStr, "accelStr", accelStr)

		time, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			log.Error("Failed to parse time in MOTION.csv row", "value", timeStr, "error", err)
			return metrics, fmt.Errorf("failed to parse time '%s' in MOTION.csv: %w", timeStr, err)
		}
		log.Debug("Parsed time", "value", time)

		altitude, err := strconv.ParseFloat(altStr, 64)
		if err != nil {
			log.Error("Failed to parse altitude in MOTION.csv row", "value", altStr, "error", err)
			return metrics, fmt.Errorf("failed to parse altitude '%s' in MOTION.csv: %w", altStr, err)
		}
		log.Debug("Parsed altitude", "value", altitude)

		velocity, err := strconv.ParseFloat(velStr, 64)
		if err != nil {
			log.Error("Failed to parse velocity in MOTION.csv row", "value", velStr, "error", err)
			return metrics, fmt.Errorf("failed to parse velocity '%s' in MOTION.csv: %w", velStr, err)
		}
		log.Debug("Parsed velocity", "value", velocity)

		acceleration, err := strconv.ParseFloat(accelStr, 64)
		if err != nil {
			log.Error("Failed to parse acceleration in MOTION.csv row", "value", accelStr, "error", err)
			return metrics, fmt.Errorf("failed to parse acceleration '%s' in MOTION.csv: %w", accelStr, err)
		}
		log.Debug("Parsed acceleration", "value", acceleration)

		if altitude > metrics.ApogeeMeters {
			metrics.ApogeeMeters = altitude
			apogeeReached = true
		}
		if velocity > metrics.MaxVelocityMPS {
			metrics.MaxVelocityMPS = velocity
		}
		if math.Abs(acceleration) > metrics.MaxAccelerationMPS2 {
			metrics.MaxAccelerationMPS2 = math.Abs(acceleration)
		}

		lastTime = time
		lastAltitude = altitude
		lastVelocity = velocity

		if apogeeReached && math.Abs(altitude-launchElevation) < 1.0 {
			metrics.LandingVelocityMPS = velocity
		}
	}

	metrics.TotalFlightTimeSec = lastTime
	if metrics.LandingVelocityMPS == 0 && lastTime > 0 {
		if math.Abs(lastAltitude-launchElevation) < 5.0 {
			metrics.LandingVelocityMPS = lastVelocity
		} else {
			metrics.LandingVelocityMPS = 0
		}
	}
	// Ensure MaxVelocity and MaxAcceleration are not the initial small numbers if no data processed meaningfully
	if metrics.MaxVelocityMPS == -1e9 {
		metrics.MaxVelocityMPS = 0
	}
	if metrics.MaxAccelerationMPS2 == -1e9 {
		metrics.MaxAccelerationMPS2 = 0
	}

	return metrics, nil
}

// parseEventsDataFromCSV processes raw string data from EVENTS.csv into a slice of FlightEvent.
// This function now primarily focuses on parsing the CSV content.
func parseEventsDataFromCSV(allRows [][]string) ([]FlightEvent, error) {
	var events []FlightEvent
	if len(allRows) == 0 { // No data at all
		return events, nil
	}
	if len(allRows) == 1 { // Only headers, no data rows
		return events, nil
	}

	headers := allRows[0]
	dataRows := allRows[1:]

	colIndices := make(map[string]int)
	for i, header := range headers {
		colIndices[strings.ToLower(strings.TrimSpace(header))] = i
	}

	requiredCols := []string{"time", "event_name"}
	for _, colName := range requiredCols {
		if _, ok := colIndices[colName]; !ok {
			return nil, fmt.Errorf("EVENTS.csv missing required column: %s. Available: %v", colName, headers)
		}
	}

	timeIdx := colIndices["time"]
	eventNameIdx := colIndices["event_name"]
	altIdx, altColExists := colIndices["altitude"]

	for _, row := range dataRows {
		timeStr := row[timeIdx]
		eventName := row[eventNameIdx]

		time, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			// logger.GetLogger("").Warn("Failed to parse time in EVENTS.csv row, skipping event", "value", timeStr, "error", err); continue
			return nil, fmt.Errorf("failed to parse time '%s' in EVENTS.csv: %w", timeStr, err)
		}

		altitude := 0.0
		if altColExists {
			altStr := row[altIdx]
			altVal, parseErr := strconv.ParseFloat(altStr, 64)
			if parseErr == nil {
				altitude = altVal
			}
		}

		events = append(events, FlightEvent{Name: eventName, TimeSec: time, AltitudeMeters: altitude})
	}
	return events, nil
}

// processEventHighlights analyzes parsed events and motion metrics to populate summary structures.
func processEventHighlights(allEvents []FlightEvent, motionMetrics parsedMotionMetrics, launchElevation float64) (MotorHighlights, ParachuteHighlights, RocketPhaseHighlights) {
	motorSummary := MotorHighlights{HasMotorEvents: false, IgnitionTimeSec: -1, BurnoutTimeSec: -1}
	parachuteSummary := ParachuteHighlights{HasParachuteEvents: false}
	phaseSummary := RocketPhaseHighlights{
		LiftoffTimeSec:    0, // Usually time 0
		ApogeeTimeSec:     motionMetrics.ApogeeMeters, // From motion data
		LandingTimeSec:    motionMetrics.TotalFlightTimeSec,  // From motion data (overall flight time)
		HasLiftoffEvent:   true, // Assume liftoff at t=0 if no specific event
		HasApogeeEvent:    motionMetrics.ApogeeMeters > 0,
		HasLandingEvent:   motionMetrics.TotalFlightTimeSec > 0 && motionMetrics.LandingVelocityMPS != 0, // Crude check
		CoastStartTimeSec: -1,
		CoastEndTimeSec:   -1,
	}

	var liftoffTime float64 = 0 // Default to 0, can be updated by Liftoff event

	for _, event := range allEvents {
		nameLower := strings.ToLower(event.Name)

		// Motor Events
		if strings.Contains(nameLower, "motor ignition") || strings.Contains(nameLower, "ignition") && !strings.Contains(nameLower, "parachute") {
			if motorSummary.IgnitionTimeSec < 0 || event.TimeSec < motorSummary.IgnitionTimeSec {
				motorSummary.IgnitionTimeSec = event.TimeSec
				motorSummary.HasMotorEvents = true
			}
		}
		if strings.Contains(nameLower, "motor burnout") || strings.Contains(nameLower, "burnout") && !strings.Contains(nameLower, "parachute") {
			if motorSummary.BurnoutTimeSec < 0 || event.TimeSec > motorSummary.BurnoutTimeSec { // Take the latest burnout if multiple
				motorSummary.BurnoutTimeSec = event.TimeSec
				motorSummary.HasMotorEvents = true
			}
		}

		// Parachute Events
		if strings.Contains(nameLower, "parachute deployed") || strings.Contains(nameLower, "deploy") && strings.Contains(nameLower, "parachute") {
			parachuteDetail := ParachuteEventDetail{
				Name:                     event.Name,
				DeploymentTimeSec:        event.TimeSec,
				DeploymentAltitudeMeters: event.AltitudeMeters,
				TimeToDeploySec:          event.TimeSec - liftoffTime, // Calculated relative to liftoff
			}
			parachuteSummary.Events = append(parachuteSummary.Events, parachuteDetail)
			parachuteSummary.HasParachuteEvents = true
		}

		// Phase Events
		if strings.Contains(nameLower, "liftoff") {
			phaseSummary.LiftoffTimeSec = event.TimeSec
			liftoffTime = event.TimeSec // Update liftoff time for parachute calculations
			phaseSummary.HasLiftoffEvent = true
		}
		if strings.Contains(nameLower, "apogee") {
			// Prefer motion data if available and event confirms, otherwise use event
			if !(motionMetrics.ApogeeMeters > 0) || (motionMetrics.ApogeeMeters > 0 && math.Abs(motionMetrics.ApogeeMeters-event.TimeSec) < 5.0) { // 5s tolerance
				phaseSummary.ApogeeTimeSec = event.TimeSec
			}
			phaseSummary.HasApogeeEvent = true
		}
		if strings.Contains(nameLower, "landing") {
			phaseSummary.LandingTimeSec = event.TimeSec
			phaseSummary.HasLandingEvent = true
		}
	}

	// Calculate motor burn duration
	if motorSummary.IgnitionTimeSec >= 0 && motorSummary.BurnoutTimeSec > motorSummary.IgnitionTimeSec {
		motorSummary.BurnDurationSec = motorSummary.BurnoutTimeSec - motorSummary.IgnitionTimeSec
	}

	// Calculate coast phase (very basic)
	if motorSummary.HasMotorEvents && motorSummary.BurnoutTimeSec >= 0 {
		phaseSummary.CoastStartTimeSec = motorSummary.BurnoutTimeSec
	} else if phaseSummary.HasLiftoffEvent { // If no motor burnout, assume coast starts after liftoff (e.g. glider)
		phaseSummary.CoastStartTimeSec = phaseSummary.LiftoffTimeSec
	}

	if phaseSummary.HasApogeeEvent && phaseSummary.ApogeeTimeSec > 0 {
		phaseSummary.CoastEndTimeSec = phaseSummary.ApogeeTimeSec
		// If parachute deploys before apogee, coast might end sooner
		if parachuteSummary.HasParachuteEvents && len(parachuteSummary.Events) > 0 {
			firstParachuteDeployTime := -1.0
			for _, pEvent := range parachuteSummary.Events {
				if firstParachuteDeployTime < 0 || pEvent.DeploymentTimeSec < firstParachuteDeployTime {
					firstParachuteDeployTime = pEvent.DeploymentTimeSec
				}
			}
			if firstParachuteDeployTime >= 0 && firstParachuteDeployTime < phaseSummary.ApogeeTimeSec {
				phaseSummary.CoastEndTimeSec = firstParachuteDeployTime
			}
		}
	}

	if phaseSummary.CoastStartTimeSec >= 0 && phaseSummary.CoastEndTimeSec > phaseSummary.CoastStartTimeSec {
		phaseSummary.CoastDurationSec = phaseSummary.CoastEndTimeSec - phaseSummary.CoastStartTimeSec
	}

	// Sort parachute events by time for consistent reporting
	sort.Slice(parachuteSummary.Events, func(i, j int) bool {
		return parachuteSummary.Events[i].DeploymentTimeSec < parachuteSummary.Events[j].DeploymentTimeSec
	})

	return motorSummary, parachuteSummary, phaseSummary
}

// GenerateReportPackage orchestrates the generation of a self-contained report package.
func GenerateReportPackage(rm *storage.RecordManager, recordID string, baseReportsDir string) (string, error) {
	log := logger.GetLogger("info")
	reportSpecificDir := filepath.Join(baseReportsDir, recordID)

	if err := os.MkdirAll(reportSpecificDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create report directory %s: %w", reportSpecificDir, err)
	}

	// TODO: Get actual config and app version if needed by GenerateReportData
	// For now, passing nil config and placeholder version
	// We might need to restructure to pass Config down or load it here.
	data, err := LoadSimulationData(rm, recordID, reportSpecificDir, log)
	if err != nil {
		// LoadSimulationData now returns the data even on some errors, handle specific cases?
		log.Error("Error generating report data, attempting to create report with partial data", "recordID", recordID, "error", err)
		// Optionally return error here if data generation failure is critical
		// return "", fmt.Errorf("failed to generate report data for %s: %w", recordID, err)
	}

	// --- Generate Markdown using text/template Generator --- 
	gen, err := NewGenerator() // Use the reinstated generator
	if err != nil {
		return "", fmt.Errorf("failed to create report generator: %w", err)
	}
	if err := gen.GenerateMarkdownFile(data, reportSpecificDir); err != nil { // Call the generator method
		return "", fmt.Errorf("failed to generate markdown report file: %w", err)
	}

	// --- Copy Assets --- 
	// (Keep existing asset copying logic if any, or add it here)
	assetSourceDir := "internal/reporting/assets" // Relative path from project root
	assetDestDir := filepath.Join(reportSpecificDir, "assets")
	if err := copyAssets(assetSourceDir, assetDestDir); err != nil {
		log.Warn("Failed to copy assets for report, plots might be missing", "recordID", recordID, "source", assetSourceDir, "dest", assetDestDir, "error", err)
		// Decide if this is a fatal error. Continuing for now.
	}

	log.Info("Successfully generated report package", "recordID", recordID, "outputDir", reportSpecificDir)
	return reportSpecificDir, nil
}

// min is a helper function to prevent out-of-bounds access when logging slices.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// copyAssets copies files from sourceDir to destDir.
func copyAssets(sourceDir, destDir string) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source asset directory %s: %w", sourceDir, err)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination asset directory %s: %w", destDir, err)
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		if entry.IsDir() {
			// Skip directories for now, or implement recursive copy if needed
			continue
		}

		input, err := os.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to read source asset file %s: %w", sourcePath, err)
		}

		if err := os.WriteFile(destPath, input, 0644); err != nil {
			return fmt.Errorf("failed to write destination asset file %s: %w", destPath, err)
		}
	}
	return nil
}
