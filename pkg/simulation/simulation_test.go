package simulation_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/simulation"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/zerodha/logf"
)

// TEST: GIVEN nothing WHEN NewSimulation is called THEN a new Simulation is returned
func TestNewSimulation(t *testing.T) {
	cfg := &config.Config{}
	log := &logf.Logger{}
	motionStore := &storage.Storage{}
	sim, err := simulation.NewSimulation(cfg, log, motionStore)
	if err != nil {
		t.Errorf("Error creating new simulation: %v", err)
	}
	if sim == nil {
		t.Errorf("New simulation is nil")
	}
}

// TEST: GIVEN a Simulation when LoadRocket is called THEN a new Rocket is returned
func TestLoadRocket(t *testing.T) {
	cfg := &config.Config{}
	log := &logf.Logger{}
	motionStore := &storage.Storage{}
	sim, _ := simulation.NewSimulation(cfg, log, motionStore)

	orkData, err := openrocket.Load("../../testdata/openrocket/l1.ork", "23.09")
	if err != nil {
		t.Errorf("Error loading OpenRocket file: %v", err)
	}

	motorData := &thrustcurves.MotorData{
		Designation:  designation.Designation("269H110-14A"),
		ID:           "1",
		Thrust:       [][]float64{{0, 0}, {1, 1}},
		TotalImpulse: 0,
		BurnTime:     0,
		AvgThrust:    0,
		TotalMass:    0,
		WetMass:      0,
		MaxThrust:    0,
	}

	err = sim.LoadRocket(&orkData.Rocket, motorData)
	if err != nil {
		t.Errorf("Error loading rocket: %v", err)
	}
}

// TEST: GIVEN a Simulation WHEN Run is called THEN the simulation runs
func TestRun(t *testing.T) {
	cfg := &config.Config{}
	log := &logf.Logger{}
	motionStore := &storage.Storage{}
	sim, _ := simulation.NewSimulation(cfg, log, motionStore)

	orkData, err := openrocket.Load("../../testdata/openrocket/l1.ork", "23.09")
	if err != nil {
		t.Errorf("Error loading OpenRocket file: %v", err)
	}

	motorData := &thrustcurves.MotorData{
		Designation:  designation.Designation("269H110-14A"),
		ID:           "1",
		Thrust:       [][]float64{{0, 0}, {1, 1}},
		TotalImpulse: 0,
		BurnTime:     0,
		AvgThrust:    0,
		TotalMass:    0,
		WetMass:      0,
		MaxThrust:    0,
	}

	err = sim.LoadRocket(&orkData.Rocket, motorData)
	if err != nil {
		t.Errorf("Error loading rocket: %v", err)
	}

	// nolint: errcheck
	go sim.Run()
}
