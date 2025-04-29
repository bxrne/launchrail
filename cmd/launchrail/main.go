package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
)

func main() {
	// --- Command Line Flags ---
	outputDirFlag := flag.String("output-dir", "", "Directory to save simulation output files (MOTION.csv, EVENTS.csv, etc.). Optional, defaults relative to config base_dir.")
	flag.Parse()

	// Load config
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1) // Exit if config fails
	}

	// Initialize logger
	log := logger.GetLogger(cfg.Setup.Logging.Level)
	log.Info("Logger initialized", "level", cfg.Setup.Logging.Level)

	// Determine output directory
	var outputDir string
	if *outputDirFlag != "" {
		// Use flag if provided
		outputDir, err = filepath.Abs(*outputDirFlag)
		if err != nil {
			log.Fatal("Failed to get absolute path for --output-dir", "path", *outputDirFlag, "error", err)
		}
		log.Info("Using specified output directory", "path", outputDir)
	} else {
		// Default behavior: relative to config base dir
		outputDir = filepath.Join(cfg.Setup.App.BaseDir, "cli_run")
		log.Info("Using default output directory relative to config base_dir", "path", outputDir)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal("Failed to create output directory", "path", outputDir, "error", err)
	}

	// Create and initialize simulation manager
	simManager := simulation.NewManager(cfg, *log) // Dereference pointer to pass interface value

	// Create storage for the run using the determined output directory
	motionStore, err := storage.NewStorage(outputDir, storage.MOTION)
	if err != nil {
		log.Fatal("Failed to create motion storage", "error", err)
	}
	eventsStore, err := storage.NewStorage(outputDir, storage.EVENTS)
	if err != nil {
		motionStore.Close() // Clean up previously opened store
		log.Fatal("Failed to create events storage", "error", err)
	}
	dynamicsStore, err := storage.NewStorage(outputDir, storage.DYNAMICS)
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

	log.Info("Simulation completed successfully.", "output_dir", outputDir) // Log the actual output dir used
}
