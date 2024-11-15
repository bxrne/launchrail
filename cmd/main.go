package main

import (
	"fmt"
	"os"

	"github.com/bxrne/launchrail/pkg/config"
	"github.com/bxrne/launchrail/pkg/logger"
	"github.com/bxrne/launchrail/pkg/tui"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error loading configuration: %v", err)
		os.Exit(1) // WARNING: Process exit
	}

	log, err := logger.GetLogger(cfg.Logs.File)
	if err != nil {
		fmt.Printf("Error getting logger: %v", err)
		os.Exit(1) // WARNING: Process exit
	}

	log.Info("Starting Launchrail application")

	svc := tui.New(cfg, log)
	if _, err := svc.Run(); err != nil {
		log.Errorf("Error running program: %v", err)
		os.Exit(1) // WARNING: Process exit
	}

	log.Info("Exiting Launchrail application")
}
