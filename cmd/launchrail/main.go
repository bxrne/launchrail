package main

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/entities"
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

	motor_data, err := thrustcurves.Load(cfg.Options.MotorDesignation, http_client.NewHTTPClient())
	if err != nil {
		log.Fatal("Failed to load motor data", "Error", err)
	}

	motor_descripton, err := motor_data.Designation.Describe()
	if err != nil {
		log.Fatal("Failed to describe motor", "Error", err)
	}
	log.Info("Motor data loaded", "Description", motor_descripton)

	// NOTE: Get rocket configuration from OpenRocket
	ork_data, err := openrocket.Load(cfg.Options.OpenRocketFile, cfg.External.OpenRocketVersion)
	if err != nil {
		log.Fatal("Failed to load OpenRocket data", "Error", err)
	}

	// NOTE: Validate OpenRocket data
	err = ork_data.Validate(cfg)
	if err != nil {
		log.Fatal("Failed to validate OpenRocket data", "Error", err)
	}
	log.Info("OpenRocket file loaded", "Description", ork_data.Describe())

	// NOTE: Create the ecs
	ecs := ecs.NewECS()
	motor := entities.NewMotor(1, motor_data)
	nosecone := entities.NewNoseconeFromORK(1, &ork_data.Rocket)
	rocket := entities.NewRocket(1, 10.0, motor, nosecone)
	ecs.AddEntity(1, rocket)

	log.Info("ECS created", "Description", ecs.String())

	log.Info("Running simulation", "Step", cfg.Simulation.Step, "MaxTime", cfg.Simulation.MaxTime)
	for i := 0.0; i < cfg.Simulation.MaxTime; i++ {
		log.Debug("Running simulation step", "Step", i)
		ecs.Update(i)
		log.Debug("States updated", "Motor", motor.String())
	}
	log.Info("Simulation complete")
}
