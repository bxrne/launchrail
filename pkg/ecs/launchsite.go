package ecs

// Launchsite represents the launch site
type Launchsite struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
}

// NewLaunchsite creates a new Launchsite instance
func NewLaunchsite() *Launchsite {
	return &Launchsite{
		Latitude:  0,
		Longitude: 0,
		Altitude:  0,
	}
}
