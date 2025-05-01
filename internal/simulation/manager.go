package simulation

import (
	"fmt"
	"sync"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/internal/plugin"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/simulation"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	logf "github.com/zerodha/logf"
)

// ManagerStatus represents the status of the simulation manager.
type ManagerStatus string

const (
	StatusIdle       ManagerStatus = "idle"
	StatusInitializing ManagerStatus = "initializing"
	StatusRunning      ManagerStatus = "running"
	StatusCompleted    ManagerStatus = "completed"
	StatusFailed       ManagerStatus = "failed"
	StatusClosed       ManagerStatus = "closed"
)

// Manager handles the overall simulation lifecycle.
type Manager struct {
	cfg    *config.Config
	log    logf.Logger
	mu     sync.Mutex
	status ManagerStatus
	sim    *simulation.Simulation
	stores *storage.Stores // Store the passed-in stores
}

// NewManager creates a new simulation manager.
func NewManager(cfg *config.Config, log logf.Logger) *Manager {
	return &Manager{
		cfg:    cfg,
		log:    log,
		status: StatusIdle,
	}
}

// Initialize sets up the simulation manager.
// It now accepts the storage.Stores instance created externally.
func (m *Manager) Initialize(stores *storage.Stores) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = StatusInitializing

	// Store the provided stores
	m.stores = stores

	// Compile plugins first
	m.log.Info("Compiling external plugins...")
	if err := plugin.CompilePlugins("./plugins", "./plugins", m.log); err != nil {
		m.log.Error("Failed to compile one or more plugins", "error", err)
		// Depending on requirements, we might want to allow proceeding even if some plugins fail.
		// For now, treat compilation failure as a fatal initialization error.
		m.status = StatusFailed
		return fmt.Errorf("plugin compilation failed: %w", err)
	}
	m.log.Info("Plugin compilation finished.")

	// Validate config consistency for step & max_time
	simStep := m.cfg.Engine.Simulation.Step
	simMax := m.cfg.Engine.Simulation.MaxTime
	if simStep <= 0 || simStep > 0.1 {
		m.status = StatusFailed
		return fmt.Errorf("invalid simulation step: must be >0 and <=0.1")
	}
	if simMax <= 0 {
		m.status = StatusFailed
		return fmt.Errorf("invalid simulation max_time: must be >0")
	}

	// Load motor data
	motorData, err := thrustcurves.Load(m.cfg.Engine.Options.MotorDesignation, http_client.NewHTTPClient())
	if err != nil {
		m.status = StatusFailed
		return err
	}
	m.log.Debug("Motor data loaded", "Designation", motorData.Designation)

	// Load OpenRocket data
	orkData, err := openrocket.Load(m.cfg.Engine.Options.OpenRocketFile, m.cfg.Engine.External.OpenRocketVersion)
	if err != nil {
		m.status = StatusFailed
		return err
	}
	m.log.Debug("OpenRocket data loaded", "Version", orkData.Version)

	// NOTE: Hash generation removed from here.
	// It should be handled by the storage.RecordManager.

	// NOTE: Storage initialization removed from here.
	// Stores are now passed in as an argument.

	// Create simulation, passing the provided stores
	m.sim, err = simulation.NewSimulation(m.cfg, m.log, m.stores)
	if err != nil {
		m.status = StatusFailed
		return err
	}

	// Load rocket data
	if err := m.sim.LoadRocket(&orkData.Rocket, motorData); err != nil {
		m.status = StatusFailed
		return err
	}

	m.status = StatusIdle // Ready to run
	return nil
}

func (m *Manager) Run() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = StatusRunning

	if err := m.sim.Run(); err != nil {
		m.status = StatusFailed
		return err
	}

	m.status = StatusCompleted
	m.log.Info("Simulation completed successfully")
	return nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status == StatusClosed {
		return nil // Already closed
	}

	m.status = StatusClosed
	m.log.Info("Simulation manager closed", "hash", "N/A - Hash now managed by storage.Record") // Update log message
	return nil
}

func (m *Manager) GetStatus() ManagerStatus {
	return m.status
}
