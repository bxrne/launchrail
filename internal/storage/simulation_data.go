package storage

// MotorData holds information about the rocket motor.
// This is a basic structure and can be expanded as needed.
type MotorData struct {
	Name string `json:"name" yaml:"name"`
	// Add other motor-specific fields here, e.g., Manufacturer, ThrustCurve, etc.
}

// SimulationData holds overall data from a simulation run that isn't part of the static config.
// It's intended to be passed to functions like report generation.
type SimulationData struct {
	Motor *MotorData `json:"motor" yaml:"motor"`
	// Add other simulation-wide data here, e.g., calculated apogee, max velocity from a full sim run, etc.
}
