package reporting

import (
	"bytes"
	_ "embed"
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

	logger "github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/zerodha/logf"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

//go:embed report.md.tmpl
var reportTemplateBytes []byte

// ReportData holds all the necessary information for generating a report.
type ReportData struct {
	Version              string // Application version
	RecordID             string
	AtmospherePlotPath   string
	ThrustPlotPath       string
	TrajectoryPlotPath   string
	DynamicsPlotPath     string
	GPSMapImagePath      string
	VelocityPlotPath     string
	AccelerationPlotPath string

	SimulationName string // Added to hold simulation name

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

	MotionMetrics MotionMetrics // Consolidated motion metrics

	// Flight Summary
	ApogeeMeters        float64
	MaxVelocityMPS      float64
	MaxAccelerationMPS2 float64
	TotalFlightTimeSec  float64
	LandingVelocityMPS  float64
	LandingAltitude     float64
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
	LandingTime              float64 // Time of landing event in seconds

	Assets map[string]string
}

// MotionMetrics holds calculated summary values from motion data.
type MotionMetrics struct {
	ApogeeMeters        float64
	MaxVelocityMPS      float64
	MaxAccelerationMPS2 float64
	TotalFlightTimeSec  float64
	LandingVelocityMPS  float64
}

// MotorHighlights summarizes key motor-related events.
type MotorHighlights struct {
	IgnitionTimeSec  float64
	BurnoutTimeSec   float64
	BurnDurationSec  float64
	HasMotorEvents   bool    // True if any motor-specific events were identified
	IgnitionAltitude float64 // Altitude at ignition
	BurnoutAltitude  float64 // Altitude at burnout
}

// ParachuteEventDetail captures information about a single parachute deployment event.
type ParachuteEventDetail struct {
	Type           string  // e.g., "Drogue", "Main"
	TimeSec        float64 // Time of deployment
	AltitudeMeters float64 // Altitude at deployment
}

// ParachuteHighlights summarizes key parachute-related events.
type ParachuteHighlights struct {
	Events               []ParachuteEventDetail
	HasParachuteEvents   bool    // True if any parachute events were identified
	DrogueDeployTimeSec  float64 // Convenience field for template
	DrogueDeployAltitude float64 // Convenience field for template
	MainDeployTimeSec    float64 // Convenience field for template
	MainDeployAltitude   float64 // Convenience field for template
}

// RocketPhaseHighlights summarizes key phases of the rocket's flight.
type RocketPhaseHighlights struct {
	LiftoffTimeSec              float64 // From motion data or "Liftoff" event
	ApogeeTimeSec               float64 // From motion data or "Apogee" event
	LandingTimeSec              float64 // From "Landing" event
	CoastStartTimeSec           float64 // Typically after motor burnout
	CoastEndTimeSec             float64 // Typically at apogee or first parachute deployment
	CoastDurationSec            float64
	HasLiftoffEvent             bool
	HasApogeeEvent              bool
	HasLandingEvent             bool
	PoweredAscentDurationSec    float64
	ApogeeAltitude              float64
	CoastToApogeeDurationSec    float64
	MainChuteDescentDurationSec float64
	DrogueDescentDurationSec    float64
	FreeFallDurationSec         float64
}

// FlightEvent represents a discrete event during the simulation.
type FlightEvent struct {
	Name           string
	TimeSec        float64
	AltitudeMeters float64
}

// Generator handles report generation using text/template.
type Generator struct {
	template *template.Template
	log      *logf.Logger
}

// NewGenerator creates a new report generator by parsing the embedded template.
func NewGenerator() (*Generator, error) {
	log := logger.GetLogger("reporting")

	// Parse the embedded template
	tmpl, err := template.New("report.md.tmpl").Parse(string(reportTemplateBytes))
	if err != nil {
		log.Error("Failed to parse embedded markdown template", "error", err)
		return nil, fmt.Errorf("failed to parse embedded markdown template: %w", err)
	}

	return &Generator{
		template: tmpl,
		log:      log,
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

// LoadSimulationData loads the necessary data for a report from storage.
func LoadSimulationData(recordID string, rm *storage.RecordManager, reportSpecificDir string) (ReportData, error) {
	log := logger.GetLogger("reporting")
	log.Debug("Loading simulation data", "record_id", recordID, "output_dir", reportSpecificDir)

	var err error
	var record *storage.Record
	record, err = rm.GetRecord(recordID)
	if err != nil {
		log.Error("Failed to get record for report data", "recordID", recordID, "error", err)
		return ReportData{}, fmt.Errorf("failed to load record %s: %w", recordID, err)
	}
	defer record.Close()

	var timeData, altitudeData, velocityData, accelerationData []float64
	// var motionMetrics MotionMetrics // No longer needed here, part of ReportData

	data := ReportData{
		RecordID: record.Hash, Version: "v0.0.0-dev", LaunchSiteName: "Default Launch Site",
		LaunchLatitude: 37.7749, LaunchLongitude: -122.4194, LaunchElevation: 10.0,
		WindSpeed: 5.0, WindDirection: 270.0, Temperature: 20.0, Pressure: 101325.0, Humidity: 60.0,
		LandingLatitude: 37.7750, LandingLongitude: -122.4200, LandingDistanceMeters: 100.0,
		LandingRadius95PctMeters: 50.0, Assets: map[string]string{},
		SimulationName: record.Name, // Populate SimulationName from record.Name
	}

	if record.Motion != nil {
		allMotionData, readErr := record.Motion.ReadAll()
		if readErr != nil {
			return ReportData{}, fmt.Errorf("failed to read motion data: %w", readErr)
		}
		if len(allMotionData) < 2 {
			return ReportData{}, fmt.Errorf("motion data insufficient, rows: %d", len(allMotionData))
		}

		var parseErr error
		// motionMetrics, parseErr = parseMotionData(allMotionData, data.LaunchElevation)
		data.MotionMetrics, parseErr = parseMotionData(allMotionData, data.LaunchElevation) // Populate data.MotionMetrics
		if parseErr != nil {
			return ReportData{}, fmt.Errorf("failed to parse motion data: %w", parseErr)
		}
		data.ApogeeMeters, data.MaxVelocityMPS, data.MaxAccelerationMPS2, data.TotalFlightTimeSec, data.LandingVelocityMPS =
			data.MotionMetrics.ApogeeMeters, data.MotionMetrics.MaxVelocityMPS, data.MotionMetrics.MaxAccelerationMPS2, data.MotionMetrics.TotalFlightTimeSec, data.MotionMetrics.LandingVelocityMPS

		if len(allMotionData) > 1 {
			headers, dataRows := allMotionData[0], allMotionData[1:]
			timeIdx, altIdx, velIdx, accIdx := -1, -1, -1, -1
			for i, h := range headers {
				switch strings.ToLower(strings.TrimSpace(h)) {
				case "time", "time (s)":
					timeIdx = i
				case "altitude", "altitude (m)":
					altIdx = i
				case "velocity", "velocity (m/s)":
					velIdx = i
				case "acceleration", "acceleration (m/s^2)":
					accIdx = i
				}
			}
			if timeIdx != -1 && altIdx != -1 && velIdx != -1 && accIdx != -1 {
				for _, row := range dataRows {
					t, tErr := parseFloatFromRow(row, timeIdx)
					a, aErr := parseFloatFromRow(row, altIdx)
					v, vErr := parseFloatFromRow(row, velIdx)
					ac, acErr := parseFloatFromRow(row, accIdx)
					if tErr == nil && aErr == nil && vErr == nil && acErr == nil {
						timeData = append(timeData, t)
						altitudeData = append(altitudeData, a)
						velocityData = append(velocityData, v)
						accelerationData = append(accelerationData, ac)
					}
				}
			}
		}
		// Plot generation
		if len(timeData) > 1 { // Need at least two points to plot a line
			altitudePlotData := newXYs(timeData, altitudeData)
			if altSVG, err := generatePlotSVG("Altitude vs. Time", "Time (s)", "Altitude (m)", altitudePlotData, reportSpecificDir, "altitude_vs_time.svg"); err == nil {
				data.Assets["altitude_vs_time.svg"] = altSVG
			} else {
				log.Error("Failed to generate altitude plot", "error", err)
			}

			velocityPlotData := newXYs(timeData, velocityData)
			if velSVG, err := generatePlotSVG("Velocity vs. Time", "Time (s)", "Velocity (m/s)", velocityPlotData, reportSpecificDir, "velocity_vs_time.svg"); err == nil {
				data.Assets["velocity_vs_time.svg"] = velSVG
			} else {
				log.Error("Failed to generate velocity plot", "error", err)
			}

			accelerationPlotData := newXYs(timeData, accelerationData)
			if accSVG, err := generatePlotSVG("Acceleration vs. Time", "Time (s)", "Acceleration (m/s^2)", accelerationPlotData, reportSpecificDir, "acceleration_vs_time.svg"); err == nil {
				data.Assets["acceleration_vs_time.svg"] = accSVG
			} else {
				log.Error("Failed to generate acceleration plot", "error", err)
			}
		}
	} else {
		return ReportData{}, fmt.Errorf("motion storage not available for record %s", recordID)
	}

	data.AllEvents = []FlightEvent{}
	var landingTime float64 = -1
	if record.Events != nil {
		allEventsData, readErr := record.Events.ReadAll()
		if readErr == nil && len(allEventsData) >= 2 {
			headers, dataRows := allEventsData[0], allEventsData[1:]
			timeIdx, eventIdx := -1, -1
			for i, h := range headers {
				switch strings.ToLower(strings.TrimSpace(h)) {
				case "time (s)", "time":
					timeIdx = i
				case "event name", "event_name":
					eventIdx = i
				}
			}
			if timeIdx != -1 && eventIdx != -1 {
				for _, row := range dataRows {
					if len(row) <= max(timeIdx, eventIdx) {
						continue
					}
					timeStr, eventName := row[timeIdx], strings.TrimSpace(row[eventIdx])
					t, tErr := strconv.ParseFloat(timeStr, 64)
					if tErr != nil {
						continue
					}
					altAtEvent := findAltitudeAtTime(timeData, altitudeData, t)
					data.AllEvents = append(data.AllEvents, FlightEvent{Name: eventName, TimeSec: t, AltitudeMeters: altAtEvent})
					if strings.EqualFold(eventName, "Landing") {
						landingTime, data.LandingAltitude, data.LandingTime = t, altAtEvent, t
					}
				}
			}
		}
	}

	if landingTime == -1 && len(altitudeData) > 0 && data.TotalFlightTimeSec > 0 {
		data.LandingAltitude, data.LandingTime = altitudeData[len(altitudeData)-1], data.TotalFlightTimeSec
		log.Warn("Landing event not found, using fallback from MOTION.csv", "alt", data.LandingAltitude, "time", data.LandingTime)
	}

	processEventHighlights(data.AllEvents, &data) // Pass pointer to data, it will be modified in place
	// data.MotorSummary, data.ParachuteSummary, data.PhaseSummary are now populated by processEventHighlights

	log.Info("Finished loading data for report", "recordID", recordID)
	return data, nil
}

// parseMotionData processes raw string data from MOTION.csv to calculate MotionMetrics.
func parseMotionData(allRows [][]string, launchElevation float64) (MotionMetrics, error) {
	log := logger.GetLogger("reporting.parseMotionData")

	metrics := MotionMetrics{
		ApogeeMeters:        launchElevation, // Initialize apogee with launch elevation
		MaxVelocityMPS:      -1e9,            // Initialize with a very small number
		MaxAccelerationMPS2: -1e9,            // Initialize with a very small number
	}

	if len(allRows) < 2 { // Headers + at least one data row
		log.Error("MOTION.csv data error: Requires at least a header and one data row", "rows_received", len(allRows))
		return metrics, fmt.Errorf("MOTION.csv data requires at least a header and one data row, got %d total rows", len(allRows))
	}

	headers := allRows[0]
	dataRows := allRows[1:]

	colIndices := make(map[string]int)
	for i, header := range headers {
		colIndices[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// Define expected column names (case-insensitive)
	// Allowing for variations like "Time" or "Time (s)"
	getTimeIdx := func() int {
		if idx, ok := colIndices["time"]; ok {
			return idx
		}
		if idx, ok := colIndices["time (s)"]; ok {
			return idx
		}
		return -1
	}
	getAltIdx := func() int {
		if idx, ok := colIndices["altitude"]; ok {
			return idx
		}
		if idx, ok := colIndices["altitude (m)"]; ok {
			return idx
		}
		return -1
	}
	getVelIdx := func() int {
		if idx, ok := colIndices["velocity"]; ok {
			return idx
		}
		if idx, ok := colIndices["velocity (m/s)"]; ok {
			return idx
		}
		return -1
	}
	getAccelIdx := func() int {
		if idx, ok := colIndices["acceleration"]; ok {
			return idx
		}
		if idx, ok := colIndices["acceleration (m/s^2)"]; ok {
			return idx
		}
		return -1
	}

	timeIdx := getTimeIdx()
	altIdx := getAltIdx()
	velIdx := getVelIdx()
	accelIdx := getAccelIdx()

	if timeIdx == -1 || altIdx == -1 || velIdx == -1 || accelIdx == -1 {
		missingCols := []string{}
		if timeIdx == -1 {
			missingCols = append(missingCols, "Time")
		}
		if altIdx == -1 {
			missingCols = append(missingCols, "Altitude")
		}
		if velIdx == -1 {
			missingCols = append(missingCols, "Velocity")
		}
		if accelIdx == -1 {
			missingCols = append(missingCols, "Acceleration")
		}
		log.Error("MOTION.csv missing required columns", "missing_columns", strings.Join(missingCols, ", "), "available_headers", headers)
		return metrics, fmt.Errorf("MOTION.csv missing required columns: %s. Available headers: %v", strings.Join(missingCols, ", "), headers)
	}

	var lastTime, lastAltitude, lastVelocity float64
	// apogeeReached := false // This variable was used in a more complex landing velocity detection; simplified for now

	log.Debug("Processing MOTION.csv data rows", "count", len(dataRows))
	for i, row := range dataRows {
		// Find the maximum index required from the row to ensure it's not out of bounds
		maxRequiredIndex := max(max(timeIdx, altIdx), max(velIdx, accelIdx))
		if len(row) <= maxRequiredIndex { // Ensure row has enough columns
			log.Warn("Skipping malformed row in MOTION.csv", "row_index", i+1, "row_length", len(row), "expected_max_idx", maxRequiredIndex)
			continue
		}

		time, err := parseFloatFromRow(row, timeIdx)
		if err != nil {
			log.Warn("Failed to parse time in MOTION.csv row", "row_index", i+1, "value", row[timeIdx], "error", err)
			continue // Skip row if time is unparseable
		}

		altitude, err := parseFloatFromRow(row, altIdx)
		if err != nil {
			log.Warn("Failed to parse altitude in MOTION.csv row", "row_index", i+1, "value", row[altIdx], "error", err)
			// Potentially continue or use a default/previous value
			altitude = lastAltitude // For simplicity, use last known if current is bad
		}

		velocity, err := parseFloatFromRow(row, velIdx)
		if err != nil {
			log.Warn("Failed to parse velocity in MOTION.csv row", "row_index", i+1, "value", row[velIdx], "error", err)
			velocity = lastVelocity
		}

		acceleration, err := parseFloatFromRow(row, accelIdx)
		if err != nil {
			log.Warn("Failed to parse acceleration in MOTION.csv row", "row_index", i+1, "value", row[accelIdx], "error", err)
			// For acceleration, a default of 0 might be safer if unparseable
			acceleration = 0
		}

		if altitude > metrics.ApogeeMeters {
			metrics.ApogeeMeters = altitude
			// apogeeReached = true
		}
		if velocity > metrics.MaxVelocityMPS { // For max speed, consider magnitude if bi-directional expected
			metrics.MaxVelocityMPS = velocity
		}
		if math.Abs(acceleration) > metrics.MaxAccelerationMPS2 {
			metrics.MaxAccelerationMPS2 = math.Abs(acceleration)
		}

		lastTime = time
		lastAltitude = altitude
		lastVelocity = velocity
	}

	metrics.TotalFlightTimeSec = lastTime

	// Simplified landing velocity: use the last recorded velocity if altitude is close to launch elevation.
	// More robust logic might involve checking for negative velocity after apogee.
	if math.Abs(lastAltitude-launchElevation) < 5.0 { // Within 5m of launch elevation
		metrics.LandingVelocityMPS = lastVelocity
	} else {
		// If not near launch elevation at end, landing velocity might be zero or needs other estimation
		// For now, if it's high, it implies it didn't 'land' in the data timeframe relative to launch elevation
		metrics.LandingVelocityMPS = lastVelocity // Or set to 0 if lastAltitude is still high
		log.Debug("Final altitude far from launch elevation", "last_alt", lastAltitude, "launch_elev", launchElevation, "assigned_landing_velo", metrics.LandingVelocityMPS)
	}

	// Ensure MaxVelocity and MaxAcceleration are not the initial small numbers if no data processed meaningfully
	if metrics.MaxVelocityMPS == -1e9 {
		metrics.MaxVelocityMPS = 0
	}
	if metrics.MaxAccelerationMPS2 == -1e9 {
		metrics.MaxAccelerationMPS2 = 0
	}

	log.Debug("Finished parsing MOTION.csv", "apogee", metrics.ApogeeMeters, "max_velo", metrics.MaxVelocityMPS, "flight_time", metrics.TotalFlightTimeSec, "landing_velo", metrics.LandingVelocityMPS)
	return metrics, nil
}

// Helper function to parse a float from a specific column in a CSV row
func parseFloatFromRow(row []string, index int) (float64, error) {
	if index < 0 || index >= len(row) {
		return 0, fmt.Errorf("index %d out of bounds for row length %d", index, len(row))
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(row[index]), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float '%s': %w", row[index], err)
	}
	return val, nil
}

// newXYs converts two slices of float64 (x and y values) into a plotter.XYs structure.
func newXYs(xData []float64, yData []float64) plotter.XYs {
	pts := make(plotter.XYs, len(xData))
	for i := range xData {
		if i < len(yData) { // Ensure we don't go out of bounds for yData if lengths differ
			pts[i].X = xData[i]
			pts[i].Y = yData[i]
		} else {
			// Handle mismatch if necessary, e.g., by logging or truncating
			// For now, assume xData is the primary length controller
			pts = pts[:i] // Truncate if yData is shorter
			break
		}
	}
	return pts
}

// generatePlotSVG creates a plot with the given data and saves it as an SVG file.
// It returns the path to the saved SVG file or an error.
func generatePlotSVG(title, xLabel, yLabel string, data plotter.XYs, reportSpecificDir, filename string) (string, error) {
	log := logger.GetLogger("reporting.generatePlotSVG")

	p := plot.New()

	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	// Add a line plotter for the data
	l, err := plotter.NewLine(data)
	if err != nil {
		log.Error("Failed to create new line plotter", "title", title, "error", err)
		return "", fmt.Errorf("failed to create line plotter for %s: %w", title, err)
	}
	l.LineStyle.Width = vg.Points(1)
	l.LineStyle.Color = plotutil.Color(0) // Use the first color in the default palette

	p.Add(l)

	// Add a grid
	p.Add(plotter.NewGrid())

	// Ensure the reportSpecificDir exists
	if err := os.MkdirAll(reportSpecificDir, 0755); err != nil {
		log.Error("Failed to create directory for plot SVG", "directory", reportSpecificDir, "error", err)
		return "", fmt.Errorf("failed to create directory %s: %w", reportSpecificDir, err)
	}

	filePath := filepath.Join(reportSpecificDir, filename)

	// Save the plot to an SVG file.
	// Dimensions are in vg.Inch units.
	if err := p.Save(8*vg.Inch, 4*vg.Inch, filePath); err != nil {
		log.Error("Failed to save plot SVG", "file_path", filePath, "error", err)
		return "", fmt.Errorf("failed to save plot %s: %w", filePath, err)
	}

	log.Info("Successfully generated plot SVG", "file_path", filePath)
	return filePath, nil
}

// Helper function to find altitude at a specific time from MOTION.csv data
// This assumes allRows includes headers and is sorted by time.
func findAltitudeAtTime(timeData, altitudeData []float64, targetTime float64) float64 {
	if len(timeData) == 0 || len(timeData) != len(altitudeData) {
		return 0
	}
	closestIdx, minDiff := 0, math.Abs(timeData[0]-targetTime)
	for i := 1; i < len(timeData); i++ {
		diff := math.Abs(timeData[i] - targetTime)
		if diff < minDiff {
			minDiff, closestIdx = diff, i
		}
	}
	return altitudeData[closestIdx]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// processEventHighlights processes a list of flight events to populate summary structures.
// It modifies the data struct directly.
func processEventHighlights(allEvents []FlightEvent, data *ReportData) {
	log := logger.GetLogger("reporting.processEventHighlights")

	// Sort events by time to make phase processing easier
	sort.SliceStable(allEvents, func(i, j int) bool {
		return allEvents[i].TimeSec < allEvents[j].TimeSec
	})
	data.AllEvents = allEvents // Store sorted events

	// Initialize summaries (some convenience fields in ParachuteHighlights will be set later)
	data.MotorSummary = MotorHighlights{BurnoutTimeSec: -1, IgnitionTimeSec: -1}
	// data.ParachuteSummary will be built by appending events
	data.ParachuteSummary.Events = []ParachuteEventDetail{}
	data.PhaseSummary = RocketPhaseHighlights{}

	var apogeeEventTime float64 = -1

	for _, event := range allEvents {
		log.Debug("Processing event", "name", event.Name, "time", event.TimeSec, "alt", event.AltitudeMeters)
		switch strings.ToLower(event.Name) {
		// Motor Events
		case "motor_ignition":
			data.MotorSummary.IgnitionTimeSec = event.TimeSec
			data.MotorSummary.IgnitionAltitude = event.AltitudeMeters
			data.MotorSummary.HasMotorEvents = true
		case "motor_burnout", "burnout":
			data.MotorSummary.BurnoutTimeSec = event.TimeSec
			data.MotorSummary.BurnoutAltitude = event.AltitudeMeters
			data.MotorSummary.HasMotorEvents = true
			if data.MotorSummary.IgnitionTimeSec > 0 {
				data.MotorSummary.BurnDurationSec = event.TimeSec - data.MotorSummary.IgnitionTimeSec
				data.PhaseSummary.PoweredAscentDurationSec = data.MotorSummary.BurnDurationSec
			}

		// Parachute Events
		case "drogue_parachute_deploy", "drogue_deploy":
			data.ParachuteSummary.Events = append(data.ParachuteSummary.Events, ParachuteEventDetail{
				Type: "Drogue", TimeSec: event.TimeSec, AltitudeMeters: event.AltitudeMeters,
			})
			data.ParachuteSummary.HasParachuteEvents = true
			// Set convenience fields
			data.ParachuteSummary.DrogueDeployTimeSec = event.TimeSec
			data.ParachuteSummary.DrogueDeployAltitude = event.AltitudeMeters
		case "main_parachute_deploy", "main_deploy":
			data.ParachuteSummary.Events = append(data.ParachuteSummary.Events, ParachuteEventDetail{
				Type: "Main", TimeSec: event.TimeSec, AltitudeMeters: event.AltitudeMeters,
			})
			data.ParachuteSummary.HasParachuteEvents = true
			// Set convenience fields
			data.ParachuteSummary.MainDeployTimeSec = event.TimeSec
			data.ParachuteSummary.MainDeployAltitude = event.AltitudeMeters

		// Phase Events (Apogee is critical for phase calculations)
		case "apogee":
			data.PhaseSummary.ApogeeTimeSec = event.TimeSec
			data.PhaseSummary.ApogeeAltitude = event.AltitudeMeters // This should match data.ApogeeMeters from motion data
			data.PhaseSummary.HasApogeeEvent = true
			apogeeEventTime = event.TimeSec
			if data.MotorSummary.BurnoutTimeSec > 0 { // Ensure burnout happened
				data.PhaseSummary.CoastToApogeeDurationSec = event.TimeSec - data.MotorSummary.BurnoutTimeSec
				data.PhaseSummary.CoastStartTimeSec = data.MotorSummary.BurnoutTimeSec
				data.PhaseSummary.CoastEndTimeSec = event.TimeSec
			}

		// Landing Event (can be used to determine descent duration)
		case "landing", "landed", "ground_hit":
			data.PhaseSummary.HasLandingEvent = true
			data.PhaseSummary.LandingTimeSec = event.TimeSec // Prefer event time for landing
			// LandingTime in ReportData is also set from this event in LoadSimulationData, which is good for consistency.
			// Descent durations are calculated later, using the most accurate landing time.

		case "liftoff":
			data.PhaseSummary.LiftoffTimeSec = event.TimeSec
			data.PhaseSummary.HasLiftoffEvent = true
		}
	}

	// Determine landing time to use for duration calculations.
	// Prefer specific landing event time, then motion data landing time, then total flight time.
	finalLandingTime := data.TotalFlightTimeSec // Fallback to total flight time from motion
	if data.PhaseSummary.HasLandingEvent && data.PhaseSummary.LandingTimeSec > 0 {
		finalLandingTime = data.PhaseSummary.LandingTimeSec
	} else if data.LandingTime > 0 { // data.LandingTime is from EVENTS.csv (or motion fallback) in LoadSimulationData
		finalLandingTime = data.LandingTime
	}

	// Refine descent durations based on the determined finalLandingTime
	if finalLandingTime > 0 {
		mainDeployed := false
		var mainDeployTime float64
		drogueDeployed := false
		var drogueDeployTime float64

		for _, pEvent := range data.ParachuteSummary.Events {
			if strings.ToLower(pEvent.Type) == "main" && pEvent.TimeSec < finalLandingTime {
				mainDeployed = true
				mainDeployTime = pEvent.TimeSec
				// Use the latest main deployment if multiple (though unlikely)
			} else if strings.ToLower(pEvent.Type) == "drogue" && pEvent.TimeSec < finalLandingTime {
				drogueDeployed = true
				drogueDeployTime = pEvent.TimeSec
			}
		}

		if mainDeployed {
			data.PhaseSummary.MainChuteDescentDurationSec = finalLandingTime - mainDeployTime
		} else if drogueDeployed {
			data.PhaseSummary.DrogueDescentDurationSec = finalLandingTime - drogueDeployTime
		} else if apogeeEventTime > 0 && finalLandingTime > apogeeEventTime {
			data.PhaseSummary.FreeFallDurationSec = finalLandingTime - apogeeEventTime
		}
	}

	log.Debug("Finished processing event highlights", "motor_summary", data.MotorSummary, "parachute_summary", data.ParachuteSummary, "phase_summary", data.PhaseSummary)
}

// GenerateReportPackage creates a report package directory containing the report and its assets.
// It returns the path to the report directory or an error.
func GenerateReportPackage(recordHash string, rm *storage.RecordManager, reportsBaseDir string) (string, error) {
	reportDir := filepath.Join(reportsBaseDir, recordHash)
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		return "", err
	}

	// Load simulation data
	data, err := LoadSimulationData(recordHash, rm, reportDir)
	if err != nil {
		return "", err
	}

	// Set asset paths in ReportData and create dummy assets for all expected plots
	assetsDir := filepath.Join(reportDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		return "", err
	}
	data.GPSMapImagePath = "assets/gps_map.png"
	data.AtmospherePlotPath = "assets/atmosphere_plot.png"
	data.ThrustPlotPath = "assets/thrust_plot.png"
	data.TrajectoryPlotPath = "assets/trajectory_plot.png"
	data.DynamicsPlotPath = "assets/dynamics_plot.png"

	for _, asset := range []string{
		"gps_map.png",
		"atmosphere_plot.png",
		"thrust_plot.png",
		"trajectory_plot.png",
		"dynamics_plot.png",
	} {
		if err := createDummyAsset(filepath.Join(assetsDir, asset)); err != nil {
			return "", err
		}
	}

	// Generate the markdown report
	gen, err := NewGenerator()
	if err != nil {
		return "", err
	}
	if err := gen.GenerateMarkdownFile(data, reportDir); err != nil {
		return "", err
	}

	// Ensure assets directory and gps_map.png exist for report references
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		return "", err
	}

	// TODO: Copy other assets, plots, etc. if needed for full packaging
	// TODO: Optionally zip the directory if required by API contract

	return reportDir, nil
}
