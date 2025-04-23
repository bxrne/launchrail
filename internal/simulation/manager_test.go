package simulation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zerodha/logf"
)

func TestNewManager(t *testing.T) {
	cfg := &config.Config{}
	log := logf.New(logf.Opts{})

	manager := simulation.NewManager(cfg, &log)
	assert.NotNil(t, manager)
}

func TestManager_Initialize(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setupConfig    func() *config.Config
		expectedError  bool
		expectedStatus simulation.SimulationStatus
	}{
		{
			name: "successful initialization",
			setupConfig: func() *config.Config {
				return &config.Config{
					Setup: config.Setup{
						App: config.App{
							BaseDir: tempDir,
						},
					},
					Engine: config.Engine{
						Options: config.Options{
							MotorDesignation: "269H110-14A",
							OpenRocketFile:   "../../testdata/openrocket/l1.ork",
						},
						External: config.External{
							OpenRocketVersion: "23.09",
						},
						Simulation: config.Simulation{
							Step:    0.01,
							MaxTime: 10,
						},
					},
				}
			},
			expectedError:  false,
			expectedStatus: simulation.StatusIdle,
		},
		{
			name: "invalid motor designation",
			setupConfig: func() *config.Config {
				return &config.Config{
					Setup: config.Setup{
						App: config.App{
							BaseDir: tempDir,
						},
					},
					Engine: config.Engine{
						Options: config.Options{
							MotorDesignation: "invalid-motor",
							OpenRocketFile:   "../../testdata/openrocket/l1.ork",
						},
					},
				}
			},
			expectedError:  true,
			expectedStatus: simulation.StatusIdle,
		},
		{
			name: "invalid OpenRocket file",
			setupConfig: func() *config.Config {
				return &config.Config{
					Setup: config.Setup{
						App: config.App{
							BaseDir: tempDir,
						},
					},
					Engine: config.Engine{
						Options: config.Options{
							MotorDesignation: "269H110-14A",
							OpenRocketFile:   "nonexistent.ork",
						},
					},
				}
			},
			expectedError:  true,
			expectedStatus: simulation.StatusIdle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			log := logf.New(logf.Opts{})

			manager := simulation.NewManager(cfg, &log)
			err := manager.Initialize()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestManager_Run(t *testing.T) {
	tempDir := t.TempDir()
	log := logf.New(logf.Opts{})
	tests := []struct {
		name           string
		setupManager   func() *simulation.Manager
		expectedError  bool
		expectedStatus simulation.SimulationStatus
	}{
		{
			name: "successful run",
			setupManager: func() *simulation.Manager {
				cfg := &config.Config{
					Setup: config.Setup{
						App: config.App{
							BaseDir: tempDir,
						},
					},
					Engine: config.Engine{
						Simulation: config.Simulation{
							Step:    0.01,
							MaxTime: 10,
						},
						Options: config.Options{
							MotorDesignation: "269H110-14A",
							OpenRocketFile:   "../../testdata/openrocket/l1.ork",
						},
						External: config.External{
							OpenRocketVersion: "23.09",
						},
					},
				}
				manager := simulation.NewManager(cfg, &log)
				err := manager.Initialize()
				require.NoError(t, err)
				return manager
			},
			expectedError:  false,
			expectedStatus: simulation.StatusComplete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := tt.setupManager()
			err := manager.Run()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestManager_Close(t *testing.T) {
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "motion.csv")
	log := logf.New(logf.Opts{})

	// Create a test file to verify it gets closed
	f, err := os.Create(dataFile)
	require.NoError(t, err)
	f.Close()

	cfg := &config.Config{
		Setup: config.Setup{
			App: config.App{
				BaseDir: tempDir,
			},
		},
		Engine: config.Engine{
			Options: config.Options{
				MotorDesignation: "269H110-14A",
				OpenRocketFile:   "../../testdata/openrocket/l1.ork",
			},
			External: config.External{
				OpenRocketVersion: "23.09",
			},
			Simulation: config.Simulation{
				Step:    0.01,
				MaxTime: 10,
			},
		},
	}

	manager := simulation.NewManager(cfg, &log)
	err = manager.Initialize()
	require.NoError(t, err)

	err = manager.Close()
	assert.NoError(t, err)
}
