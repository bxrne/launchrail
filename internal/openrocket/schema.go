package openrocket

import "encoding/xml"

// Openrocket represents the root of the OpenRocket XML structure.
type Openrocket struct {
	XMLName xml.Name `xml:"openrocket"`
	Rocket  Rocket   `xml:"rocket"`
}

// Rocket represents the rocket structure.
type Rocket struct {
	Name        string        `xml:"name"`
	Components  []Component   `xml:"component"`
	Stages      []Stage       `xml:"stage"`
	FinSets     []FinSet      `xml:"finset"`
	BodyTubes   []BodyTube    `xml:"bodyTube"`
	NoseCones   []NoseCone    `xml:"noseCone"`
	Transitions []Transition  `xml:"transition"`
	LaunchLugs  []LaunchLug   `xml:"launchLug"`
	Trapezoidal []Trapezoidal `xml:"trapezoidalFinSet"`
	Elliptical  []Elliptical  `xml:"ellipticalFinSet"`
	Freeform    []Freeform    `xml:"freeformFinSet"`
}

// Component represents a generic rocket component.
type Component struct {
	Name     string  `xml:"name"`
	Mass     float64 `xml:"mass"`
	Position float64 `xml:"position"`
}

// Stage represents a rocket stage.
type Stage struct {
	Name string `xml:"name"`
}

// FinSet represents a set of fins.
type FinSet struct {
	Name     string `xml:"name"`
	FinCount int    `xml:"finCount"`
	FinShape string `xml:"finShape"`
}

// BodyTube represents a body tube.
type BodyTube struct {
	Name   string  `xml:"name"`
	Radius float64 `xml:"radius"`
	Length float64 `xml:"length"`
}

// NoseCone represents a nose cone.
type NoseCone struct {
	Name string `xml:"name"`
}

// Transition represents a transition component.
type Transition struct {
	Name string `xml:"name"`
}

// LaunchLug represents a launch lug.
type LaunchLug struct {
	Name    string `xml:"name"`
	LugType string `xml:"lugType"`
}

// Trapezoidal represents a trapezoidal fin set.
type Trapezoidal struct {
	Name             string `xml:"name"`
	TrapezoidalShape string `xml:"trapezoidalShape"`
}

// Elliptical represents an elliptical fin set.
type Elliptical struct {
	Name            string `xml:"name"`
	EllipticalShape string `xml:"ellipticalShape"`
}

// Freeform represents a freeform fin set.
type Freeform struct {
	Name          string `xml:"name"`
	FreeformShape string `xml:"freeformShape"`
}
