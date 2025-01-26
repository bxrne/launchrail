package main

import (
	"fmt"
	"math"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/entities"
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

	// Storage via Parasite system
	motion_store, err := storage.NewStorage(cfg.App.BaseDir, "motion")
	if err != nil {
		log.Fatal("Failed to create motion storage", "Error", err)
	}
	motion_store.Init([]string{"time", "Sx", "Sy", "Sz", "Vx", "Vy", "Vz", "Ax", "Ay", "Az"})

	// Add systems
	world.AddSystem(systems.NewParasiteSystem())
	world.AddSystem(systems.NewRocketSystem())
	world.AddSystem(systems.NewPhysicsSystem(4)) // 4 worker threads

	// Create rocket entity
	rocketID := world.CreateEntity()
	nosecone := entities.NewNoseconeFromORK(rocketID, &ork_data.Rocket)
	err = world.AddComponent(rocketID, nosecone)
	if err != nil {
		log.Fatal("Failed to add Nosecone component", "Error", err)
	}

	// Add components
	motorComp := components.NewMotor(rocketID, motor_data)
	physicsComp := components.NewPhysics(9.81, 1.0)
	aeroComp := components.NewAerodynamics(0.5, math.Pi*nosecone.Radius*nosecone.Radius)

	err = world.AddComponent(rocketID, motorComp)
	if err != nil {
		log.Fatal("Failed to add Motor component", "Error", err)
	}

	err = world.AddComponent(rocketID, physicsComp)
	if err != nil {
		log.Fatal("Failed to add Physics component", "Error", err)
	}

	err = world.AddComponent(rocketID, aeroComp)
	if err != nil {
		log.Fatal("Failed to add Aerodynamics component", "Error", err)
	}

	for t := 0.0; t < cfg.Simulation.MaxTime; t += cfg.Simulation.Step {
		if err := world.Update(cfg.Simulation.Step); err != nil {
			log.Fatal(err.Error())
		}
		log.Debug(
			"Rocket",
			"Time", t,
			"Position", physicsComp.Position,
			"Velocity", physicsComp.Velocity,
			"Thrust", motorComp.GetThrust(),
		)
		motion_store.Write([]string{fmt.Sprintf("%f", t), fmt.Sprintf("%f", physicsComp.Position.X), fmt.Sprintf("%f", physicsComp.Position.Y), fmt.Sprintf("%f", physicsComp.Position.Z), fmt.Sprintf("%f", physicsComp.Velocity.X), fmt.Sprintf("%f", physicsComp.Velocity.Y), fmt.Sprintf("%f", physicsComp.Velocity.Z), fmt.Sprintf("%f", physicsComp.Acceleration.X), fmt.Sprintf("%f", physicsComp.Acceleration.Y), fmt.Sprintf("%f", physicsComp.Acceleration.Z)})

	}

	log.Info("Simulation complete")

	log.Debug("Exiting...")
}
