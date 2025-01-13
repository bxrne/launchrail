package ecs

import "fmt"

// Launchsite represents the launch site
type Launchsite struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
}

// NewLaunchsite creates a new Launchsite from config
func NewLaunchsite(lat float64, lon float64, alt float64) *Launchsite {
	return &Launchsite{
		Latitude:  lat,
		Longitude: lon,
		Altitude:  alt,
	}
}

// Describe returns a string representation of the Launchsite with units
func (l *Launchsite) Describe() string {
	return fmt.Sprintf("Lat: %.2f°, Lon: %.2f°, Alt: %.2fm", l.Latitude, l.Longitude, l.Altitude)
}
