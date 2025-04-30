package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
)

func main() {
	// Load config (which now handles flags and resolves output dir)
	cfg, err := config.GetConfig()
	if err != nil {
		// Use a basic logger or fmt.Println if logger init depends on config
		fmt.Printf("Critical error: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.GetLogger(cfg.Setup.Logging.Level)
	log.Info("Logger initialized", "level", cfg.Setup.Logging.Level)

	// Construct simulation output directory path
	// Get user's home directory
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Failed to get user's home directory", "error", err)
	}
	baseDir := filepath.Join(usr.HomeDir, ".launchrail") // Use ~/.launchrail

	// Construct simulation output directory path using the new baseDir
	outputDir := baseDir
	log.Info("Using simulation output directory", "path", outputDir)

	// Ensure output directory exists (Keep this here - app's responsibility)
	//if err := os.MkdirAll(outputDir, 0755); err != nil {
	//	log.Fatal("Failed to create output directory", "path", outputDir, "error", err)
	//}

	// Create and initialize simulation manager
	simManager := simulation.NewManager(cfg, *log) // Dereference pointer to pass interface value

	// Create storage for the run using the determined output directory
	motionStore, err := storage.NewStorage(outputDir, storage.MOTION)
	if err != nil {
		log.Fatal("Failed to create motion storage", "error", err)
	}
	if err := motionStore.Init(); err != nil {
		motionStore.Close()
		log.Fatal("Failed to initialize motion storage headers", "error", err)
	}
	eventsStore, err := storage.NewStorage(outputDir, storage.EVENTS)
	if err != nil {
		motionStore.Close() // Clean up previously opened store
		log.Fatal("Failed to create events storage", "error", err)
	}
	if err := eventsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		log.Fatal("Failed to initialize events storage headers", "error", err)
	}
	dynamicsStore, err := storage.NewStorage(outputDir, storage.DYNAMICS)
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		log.Fatal("Failed to create dynamics storage", "error", err)
	}
	if err := dynamicsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		dynamicsStore.Close()
		log.Fatal("Failed to initialize dynamics storage headers", "error", err)
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

	log.Info("Simulation completed successfully.", "output_dir", outputDir)
}
