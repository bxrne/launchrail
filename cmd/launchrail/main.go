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
	// Load config
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// Initialize logger
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

	// Initialize storage with headers
	storage, err := storage.NewStorage(cfg.App.BaseDir, "motion")
	if err != nil {
		log.Fatal("Failed to create storage", "error", err)
	}
	defer storage.Close()

	// Set headers for storage of motion data
	err = storage.Init([]string{
		"time",
		"altitude",     // Changed from position_y for clarity
		"velocity",     // Changed from velocity_y for clarity
		"acceleration", // Changed from acceleration_y for clarity
		"thrust",
	})
	if err != nil {
		log.Fatal("Failed to init storage", "error", err)
	}

	// Configure logger with additional debug level
	log.Debug("Storage initialized",
		"path", storage.GetFilePath(),
		"headers", fmt.Sprintf("%v", []string{"time", "altitude", "velocity", "acceleration", "thrust"}),
	)

	// Create simulation
	sim, err := simulation.NewSimulation(cfg, log, storage)
	if err != nil {
		log.Fatal("Failed to create simulation", "Error", err)
	}
	log.Debug("Simulation created")

	// Load rocket data
	err = sim.LoadRocket(&orkData.Rocket, motorData)
	if err != nil {
		log.Fatal("Failed to load rocket data", "Error", err)
	}
	log.Debug("Rocket data loaded")

	// Run simulation
	err = sim.Run()
	if err != nil {
		log.Fatal("Simulation failed", "Error", err)
	}

	log.Info("Simulation completed successfully")
	log.Debug("Simulation data saved", "Path", storage.GetFilePath())
}
