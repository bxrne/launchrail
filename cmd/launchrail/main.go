package main

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

func main() {
	log := logger.GetLogger()
	cfg := config.GetConfig()

	log.Info().Str("app_name", cfg.App.Name).Str("app_version", cfg.App.Version).Msg("Application configuration")

	motor, err := thrustcurves.Load(cfg.Dev.MotorDesignation)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load motor data")
	}

	log.Info().Str("motor", motor.String()).Msg("Motor data loaded")
}
