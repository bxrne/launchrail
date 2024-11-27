package components

// RocketComponent is the base interface for all rocket components.
type RocketComponent struct {
	Name     string
	Mass     float64
	Position float64
}

// Stage represents a rocket stage.
type Stage struct {
	Name string
}

// FinSet represents a set of fins.
type FinSet struct {
	Name     string
	FinCount int
	FinShape string
}

// BodyTube represents a body tube.
type BodyTube struct {
	Name   string
	Radius float64
	Length float64
}

// NoseCone represents a nose cone.
type NoseCone struct {
	Name string
}

// Transition represents a transition component.
type Transition struct {
	Name string
}

// LaunchLug represents a launch lug.
type LaunchLug struct {
	Name    string
	LugType string
}

// TrapezoidalFinSet represents a trapezoidal fin set.
type TrapezoidalFinSet struct {
	Name             string
	TrapezoidalShape string
}

// EllipticalFinSet represents an elliptical fin set.
type EllipticalFinSet struct {
	Name            string
	EllipticalShape string
}

// FreeformFinSet represents a freeform fin set.
type FreeformFinSet struct {
	Name          string
	FreeformShape string
}
