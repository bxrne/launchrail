package simulation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/simulation"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/stretchr/testify/require"
	"github.com/zerodha/logf"
)

func setupTestStorage(t *testing.T) (*storage.Stores, func()) {
	testDir := filepath.Join(os.TempDir(), "launchrail-test")
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	motionStore, err := storage.NewStorage(testDir, "motion", storage.MOTION)
	require.NoError(t, err)

	eventsStore, err := storage.NewStorage(testDir, "events", storage.EVENTS)
	require.NoError(t, err)

	stores := &storage.Stores{
		Motion: motionStore,
		Events: eventsStore,
	}

	cleanup := func() {
		motionStore.Close()
		eventsStore.Close()
		os.RemoveAll(testDir)
	}

	return stores, cleanup
}

// TEST: GIVEN nothing WHEN NewSimulation is called THEN a new Simulation is returned
func TestNewSimulation(t *testing.T) {
	cfg := &config.Config{
		Simulation: config.Simulation{
			Step:    0.001,
			MaxTime: 300.0,
		},
	}
	log := logf.New(logf.Opts{})

	stores, cleanup := setupTestStorage(t)
	defer cleanup()

	sim, err := simulation.NewSimulation(cfg, &log, stores)
	require.NoError(t, err)
	require.NotNil(t, sim)
}

// TEST: GIVEN a Simulation when LoadRocket is called THEN a new Rocket is returned
func TestLoadRocket(t *testing.T) {
	cfg := &config.Config{
		Simulation: config.Simulation{
			Step:    0.001,
			MaxTime: 300.0,
		},
	}
	log := logf.New(logf.Opts{})

	stores, cleanup := setupTestStorage(t)
	defer cleanup()

	sim, err := simulation.NewSimulation(cfg, &log, stores)
	require.NoError(t, err)

	orkData, err := openrocket.Load("../../testdata/openrocket/l1.ork", "23.09")
	require.NoError(t, err)

	motorData := &thrustcurves.MotorData{
		Designation:  designation.Designation("269H110-14A"),
		ID:           "1",
		Thrust:       [][]float64{{0, 0}, {1, 1}},
		TotalImpulse: 100.0,
		BurnTime:     1.0,
		AvgThrust:    100.0,
		TotalMass:    0.1,
		WetMass:      0.2,
		MaxThrust:    200.0,
	}

	err = sim.LoadRocket(&orkData.Rocket, motorData)
	require.NoError(t, err)
}

// TEST: GIVEN a Simulation WHEN Run is called THEN the simulation runs
func TestRun(t *testing.T) {
	cfg := &config.Config{
		Simulation: config.Simulation{
			Step:    0.001,
			MaxTime: 1.0, // Short duration for test
		},
	}
	log := logf.New(logf.Opts{})

	stores, cleanup := setupTestStorage(t)
	defer cleanup()

	sim, err := simulation.NewSimulation(cfg, &log, stores)
	require.NoError(t, err)

	orkData, err := openrocket.Load("../../testdata/openrocket/l1.ork", "23.09")
	require.NoError(t, err)

	motorData := &thrustcurves.MotorData{
		Designation:  designation.Designation("269H110-14A"),
		ID:           "1",
		Thrust:       [][]float64{{0, 0}, {1, 1}},
		TotalImpulse: 100.0,
		BurnTime:     1.0,
		AvgThrust:    100.0,
		TotalMass:    0.1,
		WetMass:      0.2,
		MaxThrust:    200.0,
	}

	err = sim.LoadRocket(&orkData.Rocket, motorData)
	require.NoError(t, err)

	err = sim.Run()
	require.NoError(t, err)
}
