package main

import (
	"fmt"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
)

func main() {
	// Load config
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// Initialize logger
	log := logger.GetLogger(cfg.Setup.Logging.Level)
	log.Info("Logger initialized", "level", cfg.Setup.Logging.Level)

	// Create and initialize simulation manager
	simManager := simulation.NewManager(cfg, *log) // Dereference pointer to pass interface value

	// Create storage for the CLI run
	cliRecordDir := filepath.Join(cfg.Setup.App.BaseDir, "cli_run")
	motionStore, err := storage.NewStorage(cliRecordDir, storage.MOTION)
	if err != nil {
		log.Fatal("Failed to create motion storage", "error", err)
	}
	eventsStore, err := storage.NewStorage(cliRecordDir, storage.EVENTS)
	if err != nil {
		motionStore.Close() // Clean up previously opened store
		log.Fatal("Failed to create events storage", "error", err)
	}
	dynamicsStore, err := storage.NewStorage(cliRecordDir, storage.DYNAMICS)
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		log.Fatal("Failed to create dynamics storage", "error", err)
	}
	stores := &storage.Stores{
		Motion:   motionStore,
		Events:   eventsStore,
		Dynamics: dynamicsStore,
	}

	if err := simManager.Initialize(stores); err != nil {
		log.Fatal("Failed to initialize simulation manager", "error", err)
	}

	// Run simulation
	if err := simManager.Run(); err != nil {
		log.Fatal("Simulation failed", "Error", err)
	}

	// Close the manager (which now closes the stores)
	if err := simManager.Close(); err != nil {
		log.Error("Failed to close simulation manager", "Error", err)
	}

	log.Info("Simulation completed successfully.")
}
