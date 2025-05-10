package simulation_test

import (
	"os"
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
							Name:    "TestApp",
							Version: "0.0.1",
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
							Name:    "TestApp",
							Version: "0.0.1",
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
							Name:    "TestApp",
							Version: "0.0.1",
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
					Setup: config.Setup{
						App: config.App{
							Name:    "TestApp",
							Version: "0.0.1",
						},
					},
					Engine: config.Engine{
						Options: config.Options{
							MotorDesignation: "269H110-14A",
							OpenRocketFile:   "../../testdata/openrocket/l1.ork",
						},
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
					Setup: config.Setup{
						App: config.App{
							Name:    "TestApp",
							Version: "0.0.1",
						},
					},
					Engine: config.Engine{
						Options: config.Options{
							MotorDesignation: "269H110-14A",
							OpenRocketFile:   "../../testdata/openrocket/l1.ork",
						},
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
					Setup: config.Setup{
						App: config.App{
							Name:    "TestApp",
							Version: "0.0.1",
						},
					},
					Engine: config.Engine{
						Options: config.Options{
							MotorDesignation: "269H110-14A",
							OpenRocketFile:   "../../testdata/openrocket/l1.ork",
						},
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

			recordDir := t.TempDir()
			motionStore, err := storage.NewStorage(recordDir, storage.MOTION)
			require.NoError(t, err)
			// defer motionStore.Close() // Manager is responsible for closing
			eventsStore, err := storage.NewStorage(recordDir, storage.EVENTS)
			require.NoError(t, err)
			// defer eventsStore.Close() // Manager is responsible for closing
			dynamicsStore, err := storage.NewStorage(recordDir, storage.DYNAMICS)
			require.NoError(t, err)
			// defer dynamicsStore.Close() // Manager is responsible for closing
			stores := &storage.Stores{
				Motion:   motionStore,
				Events:   eventsStore,
				Dynamics: dynamicsStore,
			}

			manager := simulation.NewManager(cfg, log)
			err = manager.Initialize(stores)

			// Check error and status based on whether an error was expected
			if tt.expectedError {
				assert.Error(t, err) // Error should be present
				require.NotNil(t, manager)
				require.Equal(t, simulation.StatusFailed, manager.GetStatus(), "Expected status to be Failed when initialization returns an error")
			} else {
				assert.NoError(t, err) // No error should be present
				require.NotNil(t, manager)
				require.Equal(t, simulation.StatusIdle, manager.GetStatus(), "Expected status to be Idle after successful initialization")
			}
		})
	}
}

func TestManager_Run(t *testing.T) {
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
							Name:    "TestApp",
							Version: "0.0.1",
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
				recordDir := t.TempDir()
				motionStore, err := storage.NewStorage(recordDir, storage.MOTION)
				require.NoError(t, err)
				// defer motionStore.Close() // Manager is responsible for closing
				eventsStore, err := storage.NewStorage(recordDir, storage.EVENTS)
				require.NoError(t, err)
				// defer eventsStore.Close() // Manager is responsible for closing
				dynamicsStore, err := storage.NewStorage(recordDir, storage.DYNAMICS)
				require.NoError(t, err)
				// defer dynamicsStore.Close() // Manager is responsible for closing
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
	log := logf.New(logf.Opts{})
	cfg := &config.Config{
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

	recordDir := t.TempDir()
	motionStore, err := storage.NewStorage(recordDir, storage.MOTION)
	require.NoError(t, err)
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

	// First close
	err = manager.Close()
	assert.NoError(t, err)
	assert.Equal(t, simulation.StatusClosed, manager.GetStatus())

	// Second close - should be no-op and no error
	err = manager.Close()
	assert.NoError(t, err)
	assert.Equal(t, simulation.StatusClosed, manager.GetStatus())
}

func TestManager_GetStatus(t *testing.T) {
	log := logf.New(logf.Opts{})
	cfg := &config.Config{
		Engine: config.Engine{
			Simulation: config.Simulation{
				Step:    0.01,
				MaxTime: 0.01, // Short run time
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

	recordDir := t.TempDir()
	motionStore, err := storage.NewStorage(recordDir, storage.MOTION)
	require.NoError(t, err)
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

	// 1. Before Initialize
	assert.Equal(t, simulation.StatusIdle, manager.GetStatus(), "Initial status should be Idle")

	// 2. Initialize
	err = manager.Initialize(stores)
	require.NoError(t, err)
	assert.Equal(t, simulation.StatusIdle, manager.GetStatus(), "Status after successful Initialize should be Idle")

	// 3. Run
	err = manager.Run()
	require.NoError(t, err)
	assert.Equal(t, simulation.StatusCompleted, manager.GetStatus(), "Status after successful Run should be Completed")

	// 4. Close
	err = manager.Close()
	require.NoError(t, err)
	assert.Equal(t, simulation.StatusClosed, manager.GetStatus(), "Status after Close should be Closed")
}

// TODO: Add test case for Initialize failing due to plugin compilation
// TODO: Add test case for Run failing due to simulation error
