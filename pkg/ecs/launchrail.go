package ecs

// Launchrail represents the launch rail
type Launchrail struct {
	Length      float64
	Angle       float64
	Orientation float64
}

// NewLaunchrail creates a new Launchrail instance
func NewLaunchrail() *Launchrail {
	return &Launchrail{
		Length:      0,
		Angle:       0,
		Orientation: 0,
	}
}
