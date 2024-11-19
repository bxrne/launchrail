package main

import (
	"fmt"
	"os"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/tui"
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

	t := tui.New(cfg, log)
	_, err = t.Run()
	if err != nil {
		log.Errorf("Error running TUI: %v", err)
		os.Exit(1)
	}

	log.Info("Exiting Launchrail application")
}
