package ecs

import "fmt"

// Launchrail represents the launch rail
type Launchrail struct {
	Length      float64
	Angle       float64
	Orientation float64
}

// NewLaunchrail creates a new Launchrail instance
func NewLaunchrail(length float64, angle float64, orientation float64) *Launchrail {
	return &Launchrail{
		Length:      length,
		Angle:       angle,
		Orientation: orientation,
	}
}

// Describe returns a string representation of the Launchrail with units
func (l *Launchrail) Describe() string {
	return fmt.Sprintf("Len: %.2fm, Angle: %.2f°, Orient: %.2f°", l.Length, l.Angle, l.Orientation)
}
