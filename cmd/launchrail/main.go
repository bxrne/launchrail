package main

import (
	"fmt"
	"os"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/diff"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("CRITICAL: Failed to load configuration: %v\n", err) // Use fmt for pre-logger errors
		os.Exit(1)
	}

	// Initialize logger using the new centralized function
	log, err := logger.InitFileLogger(cfg.Setup.Logging.Level, "launchrail-cli")
	if err != nil {
		fmt.Printf("CRITICAL: Failed to initialize file logger: %v\n", err) // Use fmt for pre-logger errors
		os.Exit(1)
	}

	// Determine simulation base output directory (for records, not logs)
	homedir := os.Getenv("HOME") // Consider using os.UserHomeDir() for robustness
	if homedir == "" {
		usr, uerr := os.UserHomeDir()
		if uerr != nil {
			log.Fatal("Failed to determine home directory", "error", uerr)
		}
		homedir = usr
	}

	// Generate unique run ID based on timestamp
	ork_file_bytes, err := os.ReadFile(cfg.Engine.Options.OpenRocketFile)
	if err != nil {
		log.Fatal("Failed to read openrocket file", "path", cfg.Engine.Options.OpenRocketFile, "error", err)
	}
	hash := diff.CombinedHash(cfg.Bytes(), ork_file_bytes)
	log.Info("Creating simulation run directory", "runID", hash)
	if err := os.MkdirAll(hash, 0o755); err != nil {
		log.Fatal("Failed to create simulation run directory", "path", hash, "error", err)
	}

	// Create and initialize simulation manager
	simManager := simulation.NewManager(cfg, *log) // Dereference pointer to pass interface value

	// Create storage for the run using the run-specific directory
	motionStore, err := storage.NewStorage(hash, storage.MOTION)
	if err != nil {
		log.Fatal("Failed to create motion storage", "error", err)
	}
	if err := motionStore.Init(); err != nil {
		motionStore.Close()
		log.Fatal("Failed to initialize motion storage headers", "error", err)
	}
	eventsStore, err := storage.NewStorage(hash, storage.EVENTS)
	if err != nil {
		motionStore.Close() // Clean up previously opened store
		log.Fatal("Failed to create events storage", "error", err)
	}
	if err := eventsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		log.Fatal("Failed to initialize events storage headers", "error", err)
	}
	dynamicsStore, err := storage.NewStorage(hash, storage.DYNAMICS)
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

	log.Info("Simulation completed successfully.", "SHA256", hash)
}
