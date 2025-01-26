package main

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/systems"
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

	// Create ECS world
	world := ecs.NewWorld()

	// Add systems
	world.AddSystem(systems.NewRocketSystem())
	world.AddSystem(systems.NewPhysicsSystem(4)) // 4 worker threads

	// Create rocket entity
	rocketID := world.CreateEntity()

	// Add components
	motorComp := components.NewMotor(rocketID, motor_data)
	physicsComp := components.NewPhysics(9.81, 1.0)
	aeroComp := components.NewAerodynamics(0.5, 3.14159*(0.1*0.1)) // Example area

	world.AddComponent(rocketID, motorComp)
	world.AddComponent(rocketID, physicsComp)
	world.AddComponent(rocketID, aeroComp)

	// Run simulation
	timeStep := 0.016 // 60Hz
	maxTime := 10.0   // Example max time
	for t := 0.0; t < maxTime; t += timeStep {
		if err := world.Update(timeStep); err != nil {
			log.Fatal(err.Error())
		}
		log.Debug(
			"Rocket",
			"Time", t,
			"Position", physicsComp.Position,
			"Velocity", physicsComp.Velocity,
			"Thrust", motorComp.GetThrust(),
		)
	}

	log.Info("Simulation complete")
}
