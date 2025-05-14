package storage

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
	Motor *MotorData `json:"motor" yaml:"motor"`
	// Add other simulation-wide data here, e.g., calculated apogee, max velocity from a full sim run, etc.
}
