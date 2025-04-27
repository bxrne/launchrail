package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// --- Data Structures for CSVs ---

// FlightInfo holds data from flight_info_processed.csv
type FlightInfo struct {
	Timestamp    float64
	Height       float64
	Velocity     float64
	Acceleration float64
}

// EventInfo holds data from event_info_processed.csv
type EventInfo struct {
	Timestamp float64
	Event     string
	OutIdx    int // Assuming this is an integer index
}

// FlightState holds data from flight_states_processed.csv
type FlightState struct {
	Timestamp float64
	State     string
}

// TODO: Add structs for other CSV files as needed (Baro, IMU, Filtered, GNSS, etc.)

// --- CSV Loading Functions ---

// loadCSV loads a generic CSV file, skipping the header.
func loadCSV(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	// Skip header row
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header from %s: %w", filePath, err)
	}

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read records from %s: %w", filePath, err)
	}
	return records, nil
}

// parseFloat parses a string to float64, returning an error if invalid.
func parseFloat(s string, rowIdx int, colName string, fileName string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float value '%s' in %s, row %d, column %s: %w", s, filepath.Base(fileName), rowIdx+2, colName, err) // +2 for 1-based index and header
	}
	return v, nil
}

// parseInt parses a string to int, returning an error if invalid.
func parseInt(s string, rowIdx int, colName string, fileName string) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid int value '%s' in %s, row %d, column %s: %w", s, filepath.Base(fileName), rowIdx+2, colName, err) // +2 for 1-based index and header
	}
	return v, nil
}

// LoadFlightInfo loads data from flight_info_processed.csv
func LoadFlightInfo(filePath string) ([]FlightInfo, error) {
	records, err := loadCSV(filePath)
	if err != nil {
		return nil, err
	}

	data := make([]FlightInfo, 0, len(records))
	for i, record := range records {
		if len(record) < 4 {
			return nil, fmt.Errorf("unexpected number of columns in %s, row %d: got %d, want >= 4", filepath.Base(filePath), i+2, len(record))
		}

		ts, err := parseFloat(record[0], i, "ts", filePath)
		if err != nil {
			return nil, err
		}
		h, err := parseFloat(record[1], i, "height", filePath)
		if err != nil {
			return nil, err
		}
		v, err := parseFloat(record[2], i, "velocity", filePath)
		if err != nil {
			return nil, err
		}
		a, err := parseFloat(record[3], i, "acceleration", filePath)
		if err != nil {
			return nil, err
		}

		data = append(data, FlightInfo{Timestamp: ts, Height: h, Velocity: v, Acceleration: a})
	}
	return data, nil
}

// LoadEventInfo loads event data from a CSV file.
func LoadEventInfo(filePath string) ([]EventInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	// Skip header row
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header from %s: %w", filePath, err)
	}

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read records from %s: %w", filePath, err)
	}

	data := make([]EventInfo, 0, len(records))
	for i, record := range records {
		// Expect 4 columns: #, ts, event, out_idx
		if len(record) != 4 {
			return nil, fmt.Errorf("unexpected number of columns in %s, row %d: got %d, want 4", filepath.Base(filePath), i+1, len(record))
		}

		// Parse column 1 (index 1) as timestamp
		ts, err := parseFloat(record[1], i, "ts", filePath)
		if err != nil {
			return nil, err
		}

		// Column 2 (index 2) is the event name (string)
		eventName := record[2]

		// Columns 0 (#) and 3 (out_idx) are ignored for this struct

		data = append(data, EventInfo{Timestamp: ts, Event: eventName})
	}
	return data, nil
}

// LoadFlightStates loads data from flight_states_processed.csv
func LoadFlightStates(filePath string) ([]FlightState, error) {
	records, err := loadCSV(filePath)
	if err != nil {
		return nil, err
	}

	data := make([]FlightState, 0, len(records))
	for i, record := range records {
		if len(record) != 3 { // Expect exactly 3 columns
			return nil, fmt.Errorf("unexpected number of columns in %s, row %d: got %d, want 3", filepath.Base(filePath), i+1, len(record))
		}

		// Parse column 1 (index 1) as timestamp
		ts, err := parseFloat(record[1], i, "ts", filePath)
		if err != nil {
			return nil, err
		}

		// Take column 2 (index 2) as state string
		data = append(data, FlightState{Timestamp: ts, State: record[2]})
	}
	return data, nil
}

// TODO: Add loading functions for other CSV files as needed.
