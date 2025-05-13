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
	"github.com/zerodha/logf"
)

// ManagerStatus represents the status of the simulation manager.
type ManagerStatus string

const (
	StatusIdle         ManagerStatus = "idle"
	StatusInitializing ManagerStatus = "initializing"
	StatusRunning      ManagerStatus = "running"
	StatusCompleted    ManagerStatus = "completed"
	StatusFailed       ManagerStatus = "failed"
	StatusClosed       ManagerStatus = "closed"
)

// Manager handles the overall simulation lifecycle.
type Manager struct {
	cfg           *config.Config
	log           logf.Logger
	mu            sync.Mutex
	status        ManagerStatus
	sim           *simulation.Simulation
	stores        *storage.Stores // Store the passed-in stores
	pluginManager *plugin.Manager // Add plugin manager
}

// NewManager creates a new simulation manager.
func NewManager(cfg *config.Config, log logf.Logger) *Manager {
	pm := plugin.NewManager(log, cfg) // Create plugin manager
	return &Manager{
		cfg:           cfg,
		log:           log,
		status:        StatusIdle,
		pluginManager: pm, // Store plugin manager
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

	// 1. Validate config
	if err := m.validateSimConfigInternal(); err != nil {
		m.status = StatusFailed
		return err
	}

	// 2. Load data
	motorData, orkData, err := m.loadSimDataInternal()
	if err != nil {
		m.status = StatusFailed
		return err // Error already wrapped by helper
	}
	m.log.Debug("Motor data loaded", "Designation", motorData.Designation)
	m.log.Debug("OpenRocket data loaded", "Version", orkData.Version)

	// 3. Create and initialize simulation
	sim, err := m.createAndInitializeSimInternal(stores, orkData, motorData)
	if err != nil {
		m.status = StatusFailed
		return err // Error already wrapped by helper
	}
	m.sim = sim

	m.status = StatusIdle // Ready to run
	return nil
}

// validateSimConfigInternal checks critical simulation parameters from the manager's config.
func (m *Manager) validateSimConfigInternal() error {
	simStep := m.cfg.Engine.Simulation.Step
	simMax := m.cfg.Engine.Simulation.MaxTime
	if simStep <= 0 || simStep > 0.1 {
		return fmt.Errorf("invalid simulation step: must be >0 and <=0.1, got %f", simStep)
	}
	if simMax <= 0 {
		return fmt.Errorf("invalid simulation max_time: must be >0, got %f", simMax)
	}
	return nil
}

// loadSimDataInternal loads necessary external data like motor thrust curves and OpenRocket files.
func (m *Manager) loadSimDataInternal() (*thrustcurves.MotorData, *openrocket.OpenrocketDocument, error) {
	motorData, err := thrustcurves.Load(m.cfg.Engine.Options.MotorDesignation, http_client.NewHTTPClient(), m.log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load motor data for '%s': %w", m.cfg.Engine.Options.MotorDesignation, err)
	}

	orkData, err := openrocket.Load(m.cfg.Engine.Options.OpenRocketFile, m.cfg.Engine.External.OpenRocketVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load OpenRocket file '%s': %w", m.cfg.Engine.Options.OpenRocketFile, err)
	}
	return motorData, orkData, nil
}

// createAndInitializeSimInternal creates a new simulation instance and loads the rocket data into it.
func (m *Manager) createAndInitializeSimInternal(stores *storage.Stores, orkData *openrocket.OpenrocketDocument, motorData *thrustcurves.MotorData) (*simulation.Simulation, error) {
	// Create simulation, passing the provided stores
	sim, err := simulation.NewSimulation(m.cfg, m.log, stores) // This simulation is pkg/simulation.Simulation
	if err != nil {
		return nil, fmt.Errorf("failed to create simulation instance: %w", err)
	}

	// Load rocket data
	if err := sim.LoadRocket(orkData, motorData); err != nil {
		return nil, fmt.Errorf("failed to load rocket data into simulation: %w", err)
	}
	return sim, nil
}

// Run starts a new simulation if the manager is idle.
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
