package ecs

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/ecs/entities"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

// ECS represents all non fixed objects
type ECS struct {
	World      *World
	Launchrail *Launchrail
	Launchsite *Launchsite
}

// Describe returns a string representation of the ecs
func (e *ECS) Describe() string {
	return "Rail: " + e.Launchrail.Describe() + ", Site: " + e.Launchsite.Describe() + ", World: " + e.World.Describe()
}

// New creates a new ECS instance
func NewECS(cfg *config.Config, orkData *openrocket.RocketDocument, motorData *thrustcurves.MotorData) (*ECS, error) {
	rocket, err := entities.NewRocketFromORK(orkData)
	if err != nil {
		return nil, err
	}

	return &ECS{
		World:      NewWorld(rocket),
		Launchrail: NewLaunchrail(cfg.Options.Launchrail.Length, cfg.Options.Launchrail.Angle, cfg.Options.Launchrail.Orientation),
		Launchsite: NewLaunchsite(cfg.Options.Launchsite.Latitude, cfg.Options.Launchsite.Longitude, cfg.Options.Launchsite.Altitude),
	}, nil
}
