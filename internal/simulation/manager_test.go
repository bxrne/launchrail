package simulation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/plugin"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zerodha/logf"
)

func TestNewManager(t *testing.T) {
	cfg := &config.Config{}
	log := logf.New(logf.Opts{})

	manager := simulation.NewManager(cfg, log)
	assert.NotNil(t, manager)
}

func TestMain(m *testing.M) {
	// Mock plugin compilation for all tests in this package
	originalCompilePlugins := plugin.CompilePlugins
	plugin.CompilePlugins = func(sourceDir, outputDir string, logger logf.Logger) error {
		// No-op for tests, assume plugins are pre-compiled or not needed
		logger.Info("[Test Mock] Skipping plugin compilation")
		return nil
	}

	// Run tests
	exitCode := m.Run()

	// Restore original function
	plugin.CompilePlugins = originalCompilePlugins

	os.Exit(exitCode)
}

func TestManager_Initialize(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setupConfig    func() *config.Config
		expectedError  bool
		expectedStatus simulation.ManagerStatus
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
		{
			name: "invalid simulation step (too low)",
			setupConfig: func() *config.Config {
				return &config.Config{
					Setup: config.Setup{App: config.App{BaseDir: tempDir}},
					Engine: config.Engine{
						Options: config.Options{MotorDesignation: "269H110-14A", OpenRocketFile: "../../testdata/openrocket/l1.ork"},
						Simulation: config.Simulation{
							Step:    0,
							MaxTime: 10,
						},
					},
				}
			},
			expectedError:  true,
			expectedStatus: simulation.StatusIdle,
		},
		{
			name: "invalid simulation step (too high)",
			setupConfig: func() *config.Config {
				return &config.Config{
					Setup: config.Setup{App: config.App{BaseDir: tempDir}},
					Engine: config.Engine{
						Options: config.Options{MotorDesignation: "269H110-14A", OpenRocketFile: "../../testdata/openrocket/l1.ork"},
						Simulation: config.Simulation{
							Step:    0.2,
							MaxTime: 10,
						},
					},
				}
			},
			expectedError:  true,
			expectedStatus: simulation.StatusIdle,
		},
		{
			name: "invalid simulation max time",
			setupConfig: func() *config.Config {
				return &config.Config{
					Setup: config.Setup{App: config.App{BaseDir: tempDir}},
					Engine: config.Engine{
						Options: config.Options{MotorDesignation: "269H110-14A", OpenRocketFile: "../../testdata/openrocket/l1.ork"},
						Simulation: config.Simulation{
							Step:    0.01,
							MaxTime: -5,
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

			tempDir := t.TempDir()
			recordDir := filepath.Join(tempDir, "init_test_record")
			motionStore, err := storage.NewStorage(recordDir, storage.MOTION)
			require.NoError(t, err)
			defer motionStore.Close()
			eventsStore, err := storage.NewStorage(recordDir, storage.EVENTS)
			require.NoError(t, err)
			defer eventsStore.Close()
			dynamicsStore, err := storage.NewStorage(recordDir, storage.DYNAMICS)
			require.NoError(t, err)
			defer dynamicsStore.Close()
			stores := &storage.Stores{
				Motion:   motionStore,
				Events:   eventsStore,
				Dynamics: dynamicsStore,
			}

			manager := simulation.NewManager(cfg, log)
			err = manager.Initialize(stores)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			require.NotNil(t, manager)
			require.Equal(t, simulation.StatusIdle, manager.GetStatus())
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
		expectedStatus simulation.ManagerStatus
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
				tempDir := t.TempDir()
				recordDir := filepath.Join(tempDir, "run_test_record")
				motionStore, err := storage.NewStorage(recordDir, storage.MOTION)
				require.NoError(t, err)
				defer motionStore.Close()
				eventsStore, err := storage.NewStorage(recordDir, storage.EVENTS)
				require.NoError(t, err)
				defer eventsStore.Close()
				dynamicsStore, err := storage.NewStorage(recordDir, storage.DYNAMICS)
				require.NoError(t, err)
				defer dynamicsStore.Close()
				stores := &storage.Stores{
					Motion:   motionStore,
					Events:   eventsStore,
					Dynamics: dynamicsStore,
				}

				manager := simulation.NewManager(cfg, log)
				err = manager.Initialize(stores)
				require.NoError(t, err, "Initialize should not fail for run test")
				return manager
			},
			expectedError:  false,
			expectedStatus: simulation.StatusCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := tt.setupManager()
			err := manager.Run()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, "Run should not return an error")
			}

			require.NotNil(t, manager)
			require.Equal(t, simulation.StatusCompleted, manager.GetStatus())
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

	tempDir = t.TempDir()
	recordDir := filepath.Join(tempDir, "close_test_record")
	motionStore, err := storage.NewStorage(recordDir, storage.MOTION)
	require.NoError(t, err)
	// Don't defer close here, we want manager.Close to handle it
	eventsStore, err := storage.NewStorage(recordDir, storage.EVENTS)
	require.NoError(t, err)
	dynamicsStore, err := storage.NewStorage(recordDir, storage.DYNAMICS)
	require.NoError(t, err)
	stores := &storage.Stores{
		Motion:   motionStore,
		Events:   eventsStore,
		Dynamics: dynamicsStore,
	}

	manager := simulation.NewManager(cfg, log)
	err = manager.Initialize(stores)
	require.NoError(t, err)

	// Simulate some activity (optional, but good practice)

	err = manager.Close()
	assert.NoError(t, err)
}
