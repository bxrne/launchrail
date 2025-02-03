package main

import (
	"fmt"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/simulation"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	log := logger.GetLogger(cfg)
	log.Info("Config loaded", "Name", cfg.App.Name, "Version", cfg.App.Version)

	// Load motor data
	motorData, err := thrustcurves.Load(cfg.Options.MotorDesignation, http_client.NewHTTPClient())
	if err != nil {
		log.Fatal("Failed to load motor data", "Error", err)
	}
	log.Debug("Motor data loaded", "Designation", motorData.Designation, "TotalMass", motorData.TotalMass)

	// Load OpenRocket data
	orkData, err := openrocket.Load(cfg.Options.OpenRocketFile, cfg.External.OpenRocketVersion)
	if err != nil {
		log.Fatal("Failed to load OpenRocket data", "Error", err)
	}
	log.Debug("OpenRocket data loaded", "Version", orkData.Version, "Creator", orkData.Creator)

	// Initialize storage
	motionStore, err := storage.NewStorage(cfg.App.BaseDir, "motion")
	if err != nil {
		log.Fatal("Failed to create motion storage", "Error", err)
	}
	log.Debug("Storage for motion data initialized", "BaseDir", cfg.App.BaseDir)

	// Create and run simulation with logger
	sim, err := simulation.NewSimulation(cfg, log, motionStore)
	if err != nil {
		log.Fatal("Failed to create simulation", "Error", err)
	}
	log.Debug("Simulation created")

	err = sim.LoadRocket(&orkData.Rocket, motorData)
	if err != nil {
		log.Fatal("Failed to load rocket data", "Error", err)
	}
	log.Debug("Rocket data loaded")

	err = sim.Run()
	if err != nil {
		log.Fatal("Simulation failed", "Error", err)
	}

	log.Info("Simulation completed successfully")
}
