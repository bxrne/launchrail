package main

import (
	"fmt"
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
		fmt.Printf("Error getting logger: %v\n", err)
		os.Exit(1)
	}

	log.Info("Starting Launchrail application")

	// 	sim := simulation.NewSimulation([]entity.Entity{}, []entity.System{})
	log.Info("Exiting Launchrail application")
}
