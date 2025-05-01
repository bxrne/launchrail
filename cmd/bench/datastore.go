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

// SimMotionData holds data written to MOTION.csv by the simulation
type SimMotionData struct {
	Timestamp    float64
	Altitude     float64
	Velocity     float64
	Acceleration float64
	Thrust       float64
}

// SimEventData holds data written to EVENTS.csv by the simulation
type SimEventData struct {
	Timestamp       float64
	MotorStatus     string
	ParachuteStatus string
}

// SimDynamicsData holds data written to DYNAMICS.csv by the simulation
type SimDynamicsData struct {
	Timestamp     float64
	PositionX     float64
	PositionY     float64
	PositionZ     float64
	VelocityX     float64
	VelocityY     float64
	VelocityZ     float64
	AccelerationX float64
	AccelerationY float64
	AccelerationZ float64
	OrientationX  float64
	OrientationY  float64
	OrientationZ  float64
	OrientationW  float64
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

// LoadSimMotionData loads data from the simulation's MOTION.csv
func LoadSimMotionData(filePath string) ([]SimMotionData, error) {
	records, err := loadCSV(filePath)
	if err != nil {
		return nil, err
	}
	const expectedCols = 5 // time, altitude, velocity, acceleration, thrust
	data := make([]SimMotionData, 0, len(records))
	for i, record := range records {
		if len(record) != expectedCols {
			return nil, fmt.Errorf("unexpected number of columns in %s, row %d: got %d, want %d", filepath.Base(filePath), i+2, len(record), expectedCols)
		}

		ts, err := parseFloat(record[0], i, "time", filePath)
		if err != nil {
			return nil, err
		}
		alt, err := parseFloat(record[1], i, "altitude", filePath)
		if err != nil {
			return nil, err
		}
		vel, err := parseFloat(record[2], i, "velocity", filePath)
		if err != nil {
			return nil, err
		}
		acc, err := parseFloat(record[3], i, "acceleration", filePath)
		if err != nil {
			return nil, err
		}
		thrust, err := parseFloat(record[4], i, "thrust", filePath)
		if err != nil {
			return nil, err
		}

		data = append(data, SimMotionData{
			Timestamp:    ts,
			Altitude:     alt,
			Velocity:     vel,
			Acceleration: acc,
			Thrust:       thrust,
		})
	}
	return data, nil
}

// LoadSimEventData loads data from the simulation's EVENTS.csv
func LoadSimEventData(filePath string) ([]SimEventData, error) {
	records, err := loadCSV(filePath)
	if err != nil {
		return nil, err
	}
	const expectedCols = 3 // time, motor_status, parachute_status
	data := make([]SimEventData, 0, len(records))
	for i, record := range records {
		if len(record) != expectedCols {
			return nil, fmt.Errorf("unexpected number of columns in %s, row %d: got %d, want %d", filepath.Base(filePath), i+2, len(record), expectedCols)
		}

		ts, err := parseFloat(record[0], i, "time", filePath)
		if err != nil {
			return nil, err
		}
		motorStatus := record[1]
		parachuteStatus := record[2]

		data = append(data, SimEventData{
			Timestamp:       ts,
			MotorStatus:     motorStatus,
			ParachuteStatus: parachuteStatus,
		})
	}
	return data, nil
}

// LoadSimDynamicsData loads data from the simulation's DYNAMICS.csv
func LoadSimDynamicsData(filePath string) ([]SimDynamicsData, error) {
	records, err := loadCSV(filePath)
	if err != nil {
		return nil, err
	}
	const expectedCols = 14 // time, pos_x, ..., ori_w
	data := make([]SimDynamicsData, 0, len(records))
	for i, record := range records {
		if len(record) != expectedCols {
			return nil, fmt.Errorf("unexpected number of columns in %s, row %d: got %d, want %d", filepath.Base(filePath), i+2, len(record), expectedCols)
		}

		ts, err := parseFloat(record[0], i, "time", filePath)
		if err != nil { return nil, err }
		posX, err := parseFloat(record[1], i, "position_x", filePath)
		if err != nil { return nil, err }
		posY, err := parseFloat(record[2], i, "position_y", filePath)
		if err != nil { return nil, err }
		posZ, err := parseFloat(record[3], i, "position_z", filePath)
		if err != nil { return nil, err }
		velX, err := parseFloat(record[4], i, "velocity_x", filePath)
		if err != nil { return nil, err }
		velY, err := parseFloat(record[5], i, "velocity_y", filePath)
		if err != nil { return nil, err }
		velZ, err := parseFloat(record[6], i, "velocity_z", filePath)
		if err != nil { return nil, err }
		accX, err := parseFloat(record[7], i, "acceleration_x", filePath)
		if err != nil { return nil, err }
		accY, err := parseFloat(record[8], i, "acceleration_y", filePath)
		if err != nil { return nil, err }
		accZ, err := parseFloat(record[9], i, "acceleration_z", filePath)
		if err != nil { return nil, err }
		oriX, err := parseFloat(record[10], i, "orientation_x", filePath)
		if err != nil { return nil, err }
		oriY, err := parseFloat(record[11], i, "orientation_y", filePath)
		if err != nil { return nil, err }
		oriZ, err := parseFloat(record[12], i, "orientation_z", filePath)
		if err != nil { return nil, err }
		oriW, err := parseFloat(record[13], i, "orientation_w", filePath)
		if err != nil { return nil, err }

		data = append(data, SimDynamicsData{
			Timestamp:     ts,
			PositionX:     posX, PositionY:     posY, PositionZ:     posZ,
			VelocityX:     velX, VelocityY:     velY, VelocityZ:     velZ,
			AccelerationX: accX, AccelerationY: accY, AccelerationZ: accZ,
			OrientationX:  oriX, OrientationY:  oriY, OrientationZ:  oriZ, OrientationW: oriW,
		})
	}
	return data, nil
}

// TODO: Add loading functions for other CSV files as needed.
