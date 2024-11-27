package main

import (
	"fmt"
	"os"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/openrocket"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/entities"
	"github.com/bxrne/launchrail/pkg/report"
	"github.com/bxrne/launchrail/pkg/systems"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error loading configuration: %v", err)
		os.Exit(1)
	}

	log, err := logger.GetLogger(cfg.Logs.File)
	if err != nil {
		fmt.Printf("Error intialising logger: %v", err)
		os.Exit(1)
	}
	log.Debug("Initialised")

	// Parse .ork
	rocketData, err := openrocket.Decompress(cfg.Dev.ORKFile)
	if err != nil {
		log.Fatal(err)
	}
	ecs := rocketData.ConvertToECS()

	// Initialize velocity, position, and acceleration components
	for entity := entities.Entity(1); entity < ecs.GetNextEntity(); entity++ {
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
		ecs.AddComponent(entity, &components.Acceleration{}, "Acceleration")
	}

	// Run the simulation
	simulation := systems.NewSimulation(ecs)
	simulation.Run()

	// Get the stored states from the physics system
	states := simulation.GetStates()

	// Generate report
	err = report.Generate(states, "simulation_report.csv")
	if err != nil {
		log.Fatalf("Error generating report: %v\n", err)
	}

	log.Info("Simulation completed and report generated.")
}
