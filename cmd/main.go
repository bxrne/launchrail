package main

import (
	"fmt"
	"github.com/bxrne/launchrail/internal/openrocket"
	"github.com/bxrne/launchrail/pkg/report"
	"github.com/bxrne/launchrail/pkg/systems"
	"os"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.GetLogger(cfg.Logs.File)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Started Launchrail application")
	rocketData, err := openrocket.Decompress(cfg.Dev.ORKFile)
	if err != nil {
		log.Fatal(err)
	}

	ecs := rocketData.ConvertToECS()
	log.Debug("Parsed OpenRocket data to ecs")

	log.Debug("Simulation start")

	simulation := systems.NewSimulation(ecs)
	simulation.Run()

	log.Debug("Simulation over")
	report.Generate(ecs, "simulation_report.csv")

	log.Info("Exiting Launchrail application")
}
