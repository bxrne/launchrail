package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/diff"
)

func main() {
	// Load config (which now handles flags and resolves output dir)
	cfg, err := config.GetConfig()
	if err != nil {
		// Use a basic logger or fmt.Println if logger init depends on config
		fmt.Printf("Critical error: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Determine simulation base output directory from config (with env vars)
	homedir := os.Getenv("HOME")
	outputBase := filepath.Join(homedir, ".launchrail")
	logsDir := filepath.Join(outputBase, "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		fmt.Printf("Failed to create logs directory: %v\n", err)
		os.Exit(1)
	}
	logFilePath := filepath.Join(logsDir, "simulation-20250510-152030.log")

	// Initialize logger (write to both stdout and file)
	log := logger.GetLogger(cfg.Setup.Logging.Level, logFilePath)
	log.Info("Logger initialized", "level", cfg.Setup.Logging.Level)
	log.Info("Using simulation base output directory", "path", outputBase)
	// Ensure base output directory exists
	if err := os.MkdirAll(outputBase, 0o755); err != nil {
		log.Fatal("Failed to create simulation base output directory", "path", outputBase, "error", err)
	}

	// Generate unique run ID based on timestamp
	ork_file_bytes, err := os.ReadFile(cfg.Engine.Options.OpenRocketFile)
	if err != nil {
		log.Fatal("Failed to read openrocket file", "path", cfg.Engine.Options.OpenRocketFile, "error", err)
	}
	hash := diff.CombinedHash(cfg.Bytes(), ork_file_bytes)
	runDir := filepath.Join(outputBase, hash)
	log.Info("Creating simulation run directory", "runID", hash, "path", runDir)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		log.Fatal("Failed to create simulation run directory", "path", runDir, "error", err)
	}

	// Create and initialize simulation manager
	simManager := simulation.NewManager(cfg, *log) // Dereference pointer to pass interface value

	// Create storage for the run using the run-specific directory
	motionStore, err := storage.NewStorage(runDir, storage.MOTION)
	if err != nil {
		log.Fatal("Failed to create motion storage", "error", err)
	}
	if err := motionStore.Init(); err != nil {
		motionStore.Close()
		log.Fatal("Failed to initialize motion storage headers", "error", err)
	}
	eventsStore, err := storage.NewStorage(runDir, storage.EVENTS)
	if err != nil {
		motionStore.Close() // Clean up previously opened store
		log.Fatal("Failed to create events storage", "error", err)
	}
	if err := eventsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		log.Fatal("Failed to initialize events storage headers", "error", err)
	}
	dynamicsStore, err := storage.NewStorage(runDir, storage.DYNAMICS)
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

	log.Info("Simulation completed successfully.", "output_dir", runDir)
}
