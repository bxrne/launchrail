package main

import (
	"fmt"
	"os"

	"github.com/bxrne/launchrail/internal/openrocket"
	"github.com/bxrne/launchrail/pkg/report"
	"github.com/bxrne/launchrail/pkg/systems"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
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

	// Run sim
	simulation := systems.NewSimulation(ecs)
	log.Info("Simulation starting")
	simulation.Run()
	log.Info("Simulation finished")

	// Crunch nums
	report.Generate(ecs, cfg.Dev.OutFile)
	log.Infof("Report saved to %v, exiting.", cfg.Dev.OutFile)
}
