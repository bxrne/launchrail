package main

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

func main() {
	log := logger.GetLogger()
	log.Debug("Starting...")

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("Failed to get configuration", "error", err)
	}
	log.Info("Config loaded", "Name", cfg.App.Name, "Version", cfg.App.Version)

	// NOTE: Get thrust curve from API
	motor_data, err := thrustcurves.Load(cfg.Options.MotorDesignation, &http_client.DefaultHTTPClient{}, &designation.DefaultDesignationValidator{})
	if err != nil {
		log.Fatal("Failed to load motor data", "Error", err)
	}

	motor_descripton, err := motor_data.Designation.Describe()
	if err != nil {
		log.Fatal("Failed to describe motor", "Error", err)
	}
	log.Info("Motor loaded", "Description", motor_descripton)

	// NOTE: Get rocket configuration from OpenRocket
	ork_data, err := openrocket.Load(cfg.Options.OpenRocketFile)
	if err != nil {
		log.Fatal("Failed to load OpenRocket data", "Error", err)
	}
	log.Info("OpenRocket file loaded", "Description", ork_data.Describe())
	log.Debug("Openrocket data", "Data", ork_data.String())

	log.Debug("Finished")
}
