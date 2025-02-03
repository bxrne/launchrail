package simulation_test

import (
	"os"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/simulation"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zerodha/logf"
)

func setupTest(t *testing.T) (*config.Config, *logf.Logger, *storage.Storage, func()) {
	// Create test config
	cfg := &config.Config{
		App: config.App{
			Name:    "test-sim",
			Version: "0.0.1",
			BaseDir: "test_data",
		},
		Logging: config.Logging{
			Level: "debug",
		},
		Simulation: config.Simulation{
			Step:    0.001,
			MaxTime: 1.0,
		},
		Options: config.Options{
			Launchrail: config.Launchrail{
				Length:      2.0,
				Angle:       5.0,
				Orientation: 0.0,
			},
			Launchsite: config.Launchsite{
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						GravitationalAccel: 9.81,
					},
				},
			},
		},
	}

	// Create logger
	logger := logf.New(logf.Opts{EnableColor: false})

	// Create storage
	store, err := storage.NewStorage("test_data", "motion")
	require.NoError(t, err)

	err = store.Init([]string{"Time", "Altitude", "Velocity", "Acceleration", "Thrust"})
	require.NoError(t, err)

	cleanup := func() {
		store.Close()
		os.RemoveAll("test_data")
	}

	return cfg, &logger, store, cleanup
}

func createTestRocketData() *openrocket.RocketDocument {
	return &openrocket.RocketDocument{
		Name: "Test Rocket",
		Subcomponents: openrocket.Subcomponents{
			Stages: []openrocket.RocketStage{
				{
					Name: "Sustainer",
					SustainerSubcomponents: openrocket.SustainerSubcomponents{
						Nosecone: openrocket.Nosecone{
							Length:   0.3,
							Material: openrocket.Material{Type: "bulk", Density: 1500},
							Shape:    "ogive",
						},
						BodyTube: openrocket.BodyTube{
							Length:   1.0,
							Material: openrocket.Material{Type: "bulk", Density: 1500},
							Radius:   "0.025",
							Subcomponents: openrocket.BodyTubeSubcomponents{
								TrapezoidFinset: openrocket.TrapezoidFinset{
									FinCount:  4,
									RootChord: 0.1,
									TipChord:  0.05,
									Height:    0.1,
									Thickness: 0.003,
									Material:  openrocket.Material{Type: "bulk", Density: 1500},
								},
							},
						},
					},
				},
			},
		},
	}
}

// TEST: GIVEN valid configuration WHEN NewSimulation is called THEN a new simulation is created
func TestNewSimulation(t *testing.T) {
	cfg, logger, store, cleanup := setupTest(t)
	defer cleanup()

	sim, err := simulation.NewSimulation(cfg, logger, store)
	assert.NoError(t, err)
	assert.NotNil(t, sim)
}

// TEST: GIVEN valid rocket data WHEN LoadRocket is called THEN the rocket is loaded into simulation
func TestLoadRocket(t *testing.T) {
	cfg, logger, store, cleanup := setupTest(t)
	defer cleanup()

	sim, err := simulation.NewSimulation(cfg, logger, store)
	require.NoError(t, err)

	orkData := createTestRocketData()
	motorData := &thrustcurves.MotorData{
		ID:          "test-motor",
		Designation: "H123",
		Thrust:      [][]float64{{0, 100}, {1, 0}},
		TotalMass:   0.1,
	}

	err = sim.LoadRocket(orkData, motorData)
	assert.NoError(t, err)
}

// TEST: GIVEN loaded simulation WHEN Run is called THEN simulation executes successfully
func TestRun(t *testing.T) {
	cfg, logger, store, cleanup := setupTest(t)
	defer cleanup()

	// Update simulation parameters for more realistic test
	cfg.Simulation.Step = 0.01
	cfg.Simulation.MaxTime = 2.0

	sim, err := simulation.NewSimulation(cfg, logger, store)
	require.NoError(t, err)

	orkData := createTestRocketData()
	motorData := &thrustcurves.MotorData{
		ID:          "test-motor",
		Designation: "H123",
		TotalMass:   0.325,
		// More realistic thrust curve with proper burn time
		Thrust: [][]float64{
			{0.0, 0.0},
			{0.1, 100.0},
			{0.5, 80.0},
			{1.0, 50.0},
			{1.5, 20.0},
			{2.0, 0.0},
		},
	}

	err = sim.LoadRocket(orkData, motorData)
	require.NoError(t, err)

	err = sim.Run()
	assert.NoError(t, err)
}

// TEST: GIVEN invalid simulation parameters WHEN Run is called THEN returns error
func TestRun_InvalidParameters(t *testing.T) {
	cfg, logger, store, cleanup := setupTest(t)
	defer cleanup()

	// Set invalid simulation parameters
	cfg.Simulation.Step = -1
	cfg.Simulation.MaxTime = -1

	sim, err := simulation.NewSimulation(cfg, logger, store)
	require.NoError(t, err)

	orkData := createTestRocketData()

	motorData := &thrustcurves.MotorData{
		ID:          "test-motor",
		Designation: "H123",
		Thrust:      [][]float64{{0, 100}, {1, 0}},
	}

	err = sim.LoadRocket(orkData, motorData)
	require.NoError(t, err)

	err = sim.Run()
	assert.Error(t, err)
}
