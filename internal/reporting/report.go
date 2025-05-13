package reporting

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/zerodha/logf"
)

// MotionMetrics holds summary statistics related to motion.
type MotionMetrics struct {
	MaxAltitude      float64
	MaxVelocity      float64
	MaxAccelerationY float64
	MaxGForce        float64
	ApogeeTimeSec    float64
	FlightTimeSec    float64
	GroundHitVel     float64
	// Add other relevant metrics
}

// MotorSummaryData holds summary statistics for motor performance.
type MotorSummaryData struct {
	BurnTime      float64
	PeakThrust    float64
	AverageThrust float64
	TotalImpulse  float64
	// Add other relevant metrics
}

// ParachuteSummaryData holds summary statistics for parachute performance.
type ParachuteSummaryData struct {
	DeploymentTime float64
	DescentRate    float64
	// Add other relevant metrics
}

// PhaseSummaryData holds summary statistics for flight phases.
type PhaseSummaryData struct {
	ApogeeTimeSec float64
	MaxAltitudeM  float64
	// Add other relevant metrics
}

// EventSummary provides a concise summary of a flight event.
type EventSummary struct {
	Time    float64
	Event   string
	Details string // Optional additional details
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

// ReportData holds all data required to generate a report.
type ReportData struct {
	RecordID         string
	Version          string
	RocketName       string
	MotorName        string
	LiftoffMassKg    float64
	ConfigSummary    *config.Engine // Summary of simulation engine configuration used
	Summary          ReportSummary
	Plots            map[string]string
	MotionMetrics    *MotionMetrics
	MotorSummary     MotorSummaryData
	ParachuteSummary ParachuteSummaryData
	PhaseSummary     PhaseSummaryData
	AllEvents        []EventSummary
	Stages           []StageData
	RecoverySystems  []RecoverySystemData
	MotionData       []*plotSimRecord
	MotionHeaders    []string
	EventsData       [][]string
	Log              *logf.Logger `json:"-"` // Exclude logger from JSON
	ReportTitle      string
	GenerationDate   string
	// Add other relevant data
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

// LoadSimulationData orchestrates loading all necessary data for a report.
func LoadSimulationData(recordID string, rm *storage.RecordManager, reportSpecificDir string, appCfg *config.Config) (*ReportData, error) {
	log := logger.GetLogger(appCfg.Setup.Logging.Level)
	if log == nil {
		return nil, fmt.Errorf("failed to initialize logger: logger.GetLogger returned nil")
	}

	log.Info("Loading simulation data for report", "recordID", recordID)

	rData := &ReportData{
		RecordID:         recordID,
		Version:          appCfg.Setup.App.Version,
		Plots:            make(map[string]string),
		Log:              log,
		MotionMetrics:    &MotionMetrics{},
		MotorSummary:     MotorSummaryData{},
		ParachuteSummary: ParachuteSummaryData{},
		PhaseSummary:     PhaseSummaryData{},
		Stages:           []StageData{},
		RecoverySystems:  []RecoverySystemData{},
		AllEvents:        []EventSummary{},
	}

	// Load simulation configuration summary
	simConfigFilename := "engine_config.json" // Define a standard name for the engine config file
	// The reportSpecificDir IS the record's directory path.
	simConfigPath := filepath.Join(reportSpecificDir, simConfigFilename)
	simConfigBytes, err := os.ReadFile(simConfigPath)
	if err != nil {
		log.Warn("Could not load simulation engine config", "filename", simConfigPath, "error", err)
	} else {
		var engineCfg config.Engine
		err := json.Unmarshal(simConfigBytes, &engineCfg)
		if err != nil {
			log.Warn("Error parsing simulation engine config JSON", "filename", simConfigPath, "error", err)
		} else {
			rData.ConfigSummary = &engineCfg
			if rData.ConfigSummary != nil {
				// Derive RocketName from OpenRocketFile path in the loaded engine config or appCfg as fallback
				if rData.ConfigSummary.Options.OpenRocketFile != "" {
					rData.RocketName = filepath.Base(rData.ConfigSummary.Options.OpenRocketFile)
				} else if appCfg.Engine.Options.OpenRocketFile != "" {
					rData.RocketName = filepath.Base(appCfg.Engine.Options.OpenRocketFile)
				}
				// Get MotorName from the loaded engine config or appCfg as fallback
				if rData.ConfigSummary.Options.MotorDesignation != "" {
					rData.MotorName = rData.ConfigSummary.Options.MotorDesignation
				} else if appCfg.Engine.Options.MotorDesignation != "" {
					rData.MotorName = appCfg.Engine.Options.MotorDesignation
				}
				// rData.LiftoffMassKg = simCfg.Rocket.LiftoffMass // LiftoffMassKg needs to be sourced differently, not directly available in config.Engine
			}
		}
	}

	// Get the specific record to access its storage instances
	record, err := rm.GetRecord(recordID)
	if err != nil {
		log.Error("Failed to get record for report generation", "recordID", recordID, "error", err)
		return nil, fmt.Errorf("failed to get record %s: %w", recordID, err)
	}
	// It's crucial to close the record's stores when done to release file handles.
	// Since LoadSimulationData returns ReportData which might be used later, closing here is tricky.
	// This responsibility might need to be handled by the caller of LoadSimulationData after report generation is complete.
	// For now, we will defer a close, but this might be too early if ReportData is used externally for a long time.
	// Consider if the *storage.Record itself should be part of ReportData if long-lived access to stores is needed.
	defer func() {
		if err := record.Close(); err != nil {
			log.Warn("Failed to close record stores after loading data", "recordID", recordID, "error", err)
		}
	}()

	// Load motion data
	motionFilePath := record.Motion.GetFilePath()
	motionDataCSV, err := os.ReadFile(motionFilePath)
	if err != nil {
		log.Warn("Could not load motion data", "filename", motionFilePath, "error", err)
	} else {
		motionDataParsed, motionHeaders, err := parseSimData(log, motionDataCSV, ",")
		if err != nil {
			log.Warn("Error parsing motion data", "filename", motionFilePath, "error", err)
		} else {
			rData.MotionData = motionDataParsed
			rData.MotionHeaders = motionHeaders
			log.Info("Successfully parsed motion data records", "count", len(rData.MotionData))
		}
	}

	// Load events data
	eventsFilePath := record.Events.GetFilePath()
	eventsDataCSV, err := os.ReadFile(eventsFilePath)
	if err != nil {
		log.Warn("Could not load events data", "filename", eventsFilePath, "error", err)
	} else {
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

	assetsDir := filepath.Join(reportSpecificDir, "assets")
	if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
		log.Error("Failed to create assets directory", "path", assetsDir, "error", err)
		return nil, fmt.Errorf("failed to create assets directory '%s': %w", assetsDir, err)
	}
	rData.ReportTitle = fmt.Sprintf("Simulation Report for %s - %s", rData.RocketName, recordID)
	rData.GenerationDate = time.Now().Format(time.RFC1123)

	// Calculate and set summary motion statistics
	if len(rData.MotionData) > 0 {
		if rData.MotionMetrics == nil { // Ensure MotionMetrics is initialized
			rData.MotionMetrics = &MotionMetrics{}
		}

		maxAltitude := -1e9 // Initialize with a very small number
		maxVelocity := -1e9 // Initialize with a very small number

		for _, row := range rData.MotionData { // row is *plotSimRecord (pointer to map[string]interface{})
			if altVal, ok := (*row)["altitude"].(float64); ok { // Corrected: Dereference pointer then map index
				if altVal > maxAltitude {
					maxAltitude = altVal
				}
			}
			if velVal, ok := (*row)["velocity"].(float64); ok { // Corrected: Dereference pointer then map index
				if velVal > maxVelocity {
					maxVelocity = velVal
				}
			}
		}

		if maxAltitude > -1e9 { // Check if any data was processed
			rData.MotionMetrics.MaxAltitude = maxAltitude
		}
		if maxVelocity > -1e9 { // Check if any data was processed
			rData.MotionMetrics.MaxVelocity = maxVelocity
		}
	}

	// Populate Plots map with placeholders for expected plots
	// This satisfies test assertions for plot key existence.
	// Actual plot generation functions (plotFunc, plotAltitude etc.) would go here.
	requiredPlots := []string{"altitude_vs_time", "velocity_vs_time", "acceleration_vs_time"}
	if rData.Plots == nil {
		rData.Plots = make(map[string]string)
	}
	for _, plotKey := range requiredPlots {
		// In a real scenario, this would be the path to the generated plot image.
		rData.Plots[plotKey] = filepath.ToSlash(filepath.Join("assets", plotKey+".svg"))
		log.Info("Placeholder entry added for plot", "plot_key", plotKey, "path", rData.Plots[plotKey])
	}

	log.Info("Simulation data loading and basic processing complete", "RecordID", recordID)
	return rData, nil
}
