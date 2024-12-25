package main

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

func main() {
	cfg := config.GetConfig("config")

	// NOTE: Get thrust curve from API
	_ = thrustcurves.Load(cfg.Dev.MotorDesignation)

	// TODO: Get rocket configuration from OpenRocket

	// TODO: Get launch configuration from OpenRocket

	// TODO: Get conditions

	// TODO: Setup simulation

	// TODO: Run simulation

	// TODO: Output results

	// TODO: Cleanup
}
