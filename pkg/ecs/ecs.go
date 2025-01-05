package ecs

import "github.com/bxrne/launchrail/internal/config"

// ECS represents all non fixed objects
type ECS struct {
	World      *World
	Launchrail *Launchrail
	Launchsite *Launchsite
}

// Describe returns a string representation of the ecs
func (e *ECS) Describe() string {
	return "Rail: " + e.Launchrail.Describe() + ", Site: " + e.Launchsite.Describe()
}

// New creates a new ECS instance
func New(cfg *config.Config) (*ECS, error) {
	return &ECS{
		World:      NewWorld(),
		Launchrail: NewLaunchrail(cfg.Options.Launchrail.Length, cfg.Options.Launchrail.Angle, cfg.Options.Launchrail.Orientation),
		Launchsite: NewLaunchsite(cfg.Options.Launchsite.Latitude, cfg.Options.Launchsite.Longitude, cfg.Options.Launchsite.Altitude),
	}, nil
}
