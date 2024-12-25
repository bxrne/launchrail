package main

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

func main() {
	log := logger.GetLogger()
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("Failed to get configuration: %s", err)
	}
	log.Info("Config loaded", "name", cfg.App.Name, "version", cfg.App.Version)

	// NOTE: Get thrust curve from API
	motor_data, err := thrustcurves.Load(cfg.Options.MotorDesignation, &http_client.DefaultHTTPClient{}, &designation.DefaultDesignationValidator{})
	if err != nil {
		log.Fatal("Failed to load motor data: %s", err)
	}

	motor_descripton, err := motor_data.Designation.Describe()
	if err != nil {
		log.Fatal("Failed to describe motor: %s", err)
	}
	log.Info("Motor loaded", "description", motor_descripton)

	// TODO: Get rocket configuration from OpenRocket

	// TODO: Get launch configuration from OpenRocket

	// TODO: Get conditions

	// TODO: Setup simulation

	// TODO: Run simulation

	// TODO: Output results

	// TODO: Cleanup
}
