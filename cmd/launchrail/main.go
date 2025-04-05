package main

import (
	"fmt"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
)

func main() {
	// Load config
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// Initialize logger
	log := logger.GetLogger(cfg)
	log.Info("Config loaded", "Name", cfg.Setup.App.Name, "Version", cfg.Setup.App.Version)

	// Create and initialize simulation manager
	simManager := simulation.NewManager(cfg, log)
	defer simManager.Close()

	if err := simManager.Initialize(); err != nil {
		log.Fatal("Failed to initialize simulation", "Error", err)
	}

	// Run simulation
	if err := simManager.Run(); err != nil {
		log.Fatal("Simulation failed", "Error", err)
	}
}
