package main

import (
	"fmt"
	"os"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
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

	_, err = tea.NewProgram(initialModel(cfg, log)).Run()
	if err != nil {
		log.Errorf("Error starting Launchrail application: %v", err)
		os.Exit(1) // WARNING: Process exit
	}

	log.Info("Exiting Launchrail application")
}
