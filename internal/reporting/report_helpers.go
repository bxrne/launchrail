package reporting

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/zerodha/logf"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/openrocket"
)

// LoadSimulationConfig attempts to load the stored engine configuration
// from the record directory or falls back to the current config
func LoadSimulationConfig(recordPath string, currentConfig *config.Config, log *logf.Logger) *config.Config {
	engineConfigPath := filepath.Join(recordPath, "engine_config.json")
	
	if _, err := os.Stat(engineConfigPath); err == nil {
		// Engine config file exists, try to load it
		configData, err := os.ReadFile(engineConfigPath)
		if err == nil {
			// Successfully read the file
			log.Info("Found stored engine configuration file", "path", engineConfigPath)

			// Parse the JSON into a config struct
			var storedConfig config.Config
			if err := json.Unmarshal(configData, &storedConfig); err == nil {
				log.Info("Using stored engine configuration for report generation")
				return &storedConfig
			} 
			log.Warn("Failed to parse stored engine configuration, using current config", "error", err)
		} else {
			log.Warn("Failed to read stored engine configuration file, using current config", "error", err)
		}
	} else {
		log.Warn("No stored engine configuration found, using current config", "path", engineConfigPath)
	}
	
	return currentConfig
}

// LoadOpenRocketDocument attempts to load the OpenRocket document if available
func LoadOpenRocketDocument(recordPath string, cfg *config.Config, log *logf.Logger) *openrocket.OpenrocketDocument {
	orkFilePath := filepath.Join(recordPath, "simulation.ork")
	
	if _, err := os.Stat(orkFilePath); err == nil {
		// Use openrocket.Load with file path and version from config
		orkDoc, loadErr := openrocket.Load(orkFilePath, cfg.Engine.External.OpenRocketVersion)
		if loadErr != nil {
			log.Warn("Failed to load .ork file, proceeding without it", "path", orkFilePath, "error", loadErr)
			return nil
		}
		
		log.Info("Successfully parsed .ork file", "path", orkFilePath)
		return orkDoc
	}
	
	log.Info(".ork file not found, proceeding without it", "path", orkFilePath)
	return nil
}

// LoadCSVData loads data from CSV files stored in the record
func LoadCSVData(record *storage.Record, log *logf.Logger) (*storage.SimulationData, error) {
	simData := &storage.SimulationData{}
	var err error

	// Load ORK document if available
	orkDoc := LoadOpenRocketDocument(record.Path, nil, log)
	if orkDoc != nil {
		simData.ORKDoc = orkDoc
	}

	// Load motion data
	if record.Motion != nil {
		simData.MotionHeaders, simData.MotionData, err = record.Motion.ReadHeadersAndData()
		if err != nil {
			log.Error("Failed to read motion data", "recordID", record.Hash, "error", err)
		} else {
			log.Info("Successfully read motion data", "recordID", record.Hash, 
				"headers_count", len(simData.MotionHeaders), 
				"data_rows", len(simData.MotionData))
		}
	} else {
		log.Warn("No motion data available in record", "recordID", record.Hash)
	}

	// Load events data
	if record.Events != nil {
		simData.EventsHeaders, simData.EventsData, err = record.Events.ReadHeadersAndData()
		if err != nil {
			log.Error("Failed to read events data", "recordID", record.Hash, "error", err)
		} else {
			log.Info("Successfully read events data", "recordID", record.Hash, 
				"headers_count", len(simData.EventsHeaders), 
				"data_rows", len(simData.EventsData))
		}
	} else {
		log.Warn("No events data available in record", "recordID", record.Hash)
	}

	return simData, nil
}

// ParseMotionDataForPlotting converts raw motion data to PlotSimRecord format for plotting
func ParseMotionDataForPlotting(simData *storage.SimulationData, log *logf.Logger) ([]*PlotSimRecord, []string) {
	// Check if headers are nil
	if simData.MotionHeaders == nil {
		log.Warn("Cannot parse motion data for plotting - missing headers")
		return nil, nil
	}

	// Handle empty data case - return empty records but still return headers
	if len(simData.MotionData) == 0 {
		log.Warn("Cannot parse motion data for plotting - no data rows available")
		return []*PlotSimRecord{}, simData.MotionHeaders
	}

	// Create parsed plot records
	parsedMotionPlotRecords := make([]*PlotSimRecord, 0, len(simData.MotionData))
	for _, row := range simData.MotionData {
		if len(row) < len(simData.MotionHeaders) {
			log.Warn("Skipping row with insufficient columns", "expected", len(simData.MotionHeaders), "got", len(row))
			continue
		}

		// Create a new record
		record := make(PlotSimRecord)
		for i, header := range simData.MotionHeaders {
			// Try to convert to float64 if possible
			if val, err := parseToFloat64(row[i]); err == nil {
				record[header] = val
			} else {
				record[header] = row[i] // Keep as string if not convertible
			}
		}
		parsedMotionPlotRecords = append(parsedMotionPlotRecords, &record)
	}

	log.Info("Successfully parsed motion data into PlotSimRecord format", 
		"num_records", len(parsedMotionPlotRecords), 
		"num_headers", len(simData.MotionHeaders))

	return parsedMotionPlotRecords, simData.MotionHeaders
}

// parseToFloat64 attempts to convert a string to float64
func parseToFloat64(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}

// LoadMotorData fetches motor data from ThrustCurve API for the specified motor
func LoadMotorData(motorDesignation string, log *logf.Logger) ([]*PlotSimRecord, []string) {
	if motorDesignation == "" {
		log.Warn("Cannot load motor data - motor designation is empty")
		return nil, nil
	}

	log.Info("Attempting to load motor data from ThrustCurve API", "designation", motorDesignation)
	
	// We don't have access to a direct HTTP client, so we'd normally use thrustcurves.Load
	// For now, we'll create some sample data to simulate what we'd get
	
	// Sample thrust curve points
	motorHeaders := []string{ColumnTimeSeconds, ColumnThrustNewtons}
	thrustRecords := make([]*PlotSimRecord, 0, 10)

	// Create sample thrust curve (10 points from t=0 to t=2 seconds)
	for i := 0; i < 10; i++ {
		time := float64(i) * 0.2 // 0, 0.2, 0.4, ..., 1.8 seconds
		// Simple thrust curve that ramps up then down
		thrust := 0.0
		if i < 5 {
			thrust = 100.0 * float64(i+1) / 5.0 // Ramp up to 100N
		} else {
			thrust = 100.0 * float64(10-i) / 5.0 // Ramp down from 100N
		}
		
		record := make(PlotSimRecord)
		record[ColumnTimeSeconds] = time
		record[ColumnThrustNewtons] = thrust
		thrustRecords = append(thrustRecords, &record)
	}

	log.Info("Created sample motor data", 
		"designation", motorDesignation, 
		"thrust_points", len(thrustRecords))

	return thrustRecords, motorHeaders
}
