package types

type SimulationMode int

const (
	ThreeDOF SimulationMode = iota
	SixDOF
)

type State struct {
	Position Vector3
	Velocity Vector3
	Mass     float64
	Time     float64

	// INFO: 6DOF specific fields
	Orientation *Quaternion `json:",omitempty"`
	AngularVel  *Vector3    `json:",omitempty"`
}
