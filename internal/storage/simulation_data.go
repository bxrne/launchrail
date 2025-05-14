package storage

import (
	"github.com/bxrne/launchrail/pkg/openrocket"
)

// ThrustPoint represents a single point on a motor thrust curve
type ThrustPoint struct {
	Time   float64
	Thrust float64
}

// MotorData represents motor-specific simulation data
type MotorData struct {
	Name          string
	MaxThrust     float64
	TotalImpulse  float64
	BurnTime      float64
	AverageThrust float64
	ThrustData    []ThrustPoint
	Headers       []string   // Headers for the data columns
	Data          [][]string // Raw CSV data for motor
}

// SimulationData holds overall data from a simulation run that isn't part of the static config.
// It's intended to be passed to functions like report generation.
type SimulationData struct {
	Motor         *MotorData                     `json:"motor" yaml:"motor"`
	MotionHeaders []string                       `json:"motion_headers,omitempty" yaml:"motion_headers,omitempty"`
	MotionData    [][]string                     `json:"motion_data,omitempty" yaml:"motion_data,omitempty"`
	EventsHeaders []string                       `json:"events_headers,omitempty" yaml:"events_headers,omitempty"`
	EventsData    [][]string                     `json:"events_data,omitempty" yaml:"events_data,omitempty"`
	ORKDoc        *openrocket.OpenrocketDocument `json:"ork_doc,omitempty" yaml:"ork_doc,omitempty"`
	// Add other simulation-wide data here, e.g., calculated apogee, max velocity from a full sim run, etc.
}
