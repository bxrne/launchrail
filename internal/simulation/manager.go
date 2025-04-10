package simulation

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/diff"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/simulation"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/zerodha/logf"
)

type Manager struct {
	cfg     *config.Config
	log     *logf.Logger
	sim     *simulation.Simulation
	status  SimulationStatus
	simHash string
	stores  *storage.Stores
}

type SimulationStatus string

const (
	StatusIdle     SimulationStatus = "idle"
	StatusRunning  SimulationStatus = "running"
	StatusComplete SimulationStatus = "complete"
	StatusError    SimulationStatus = "error"
)

func NewManager(cfg *config.Config, log *logf.Logger) *Manager {
	return &Manager{
		cfg:    cfg,
		log:    log,
		status: StatusIdle,
	}
}

func (m *Manager) Initialize() error {
	// Load motor data
	motorData, err := thrustcurves.Load(m.cfg.Engine.Options.MotorDesignation, http_client.NewHTTPClient())
	if err != nil {
		return err
	}
	m.log.Debug("Motor data loaded", "Designation", motorData.Designation)

	// Load OpenRocket data
	orkData, err := openrocket.Load(m.cfg.Engine.Options.OpenRocketFile, m.cfg.Engine.External.OpenRocketVersion)
	if err != nil {
		return err
	}
	m.log.Debug("OpenRocket data loaded", "Version", orkData.Version)

	// Generate simulation hash
	m.simHash = diff.CombinedHash(orkData.Bytes(), m.cfg.Bytes())

	// Initialize storages
	if err := m.initializeStorages(); err != nil {
		return err
	}

	// Create simulation
	m.sim, err = simulation.NewSimulation(m.cfg, m.log, m.stores)
	if err != nil {
		return err
	}

	// Load rocket data
	if err := m.sim.LoadRocket(&orkData.Rocket, motorData); err != nil {
		return err
	}

	m.status = StatusIdle
	return nil
}

func (m *Manager) initializeStorages() error {
	// Initialize motion storage
	motionStorage, err := storage.NewStorage(m.cfg.Setup.App.BaseDir, m.simHash, storage.MOTION)
	if err != nil {
		return err
	}

	if err := motionStorage.Init(); err != nil {
		return err
	}

	// Initialize events storage
	eventsStorage, err := storage.NewStorage(m.cfg.Setup.App.BaseDir, m.simHash, storage.EVENTS)
	if err != nil {
		return err
	}

	if err := eventsStorage.Init(); err != nil {
		return err
	}

	dynamicsStorage, err := storage.NewStorage(m.cfg.Setup.App.BaseDir, m.simHash, storage.DYNAMICS)
	if err != nil {
		return err
	}

	if err := dynamicsStorage.Init(); err != nil {
		return err
	}

	m.stores = &storage.Stores{
		Motion:   motionStorage,
		Events:   eventsStorage,
		Dynamics: dynamicsStorage,
	}

	return nil
}

func (m *Manager) Run() error {
	m.status = StatusRunning

	if err := m.sim.Run(); err != nil {
		m.status = StatusError
		return err
	}

	m.status = StatusComplete
	m.log.Info("Simulation completed successfully")
	return nil
}

func (m *Manager) Close() error {
	if m.stores != nil {
		if m.stores.Motion != nil {
			if err := m.stores.Motion.Close(); err != nil {
				return err
			}
		}
		if m.stores.Events != nil {
			if err := m.stores.Events.Close(); err != nil {
				return err
			}
		}
		if m.stores.Dynamics != nil {
			if err := m.stores.Dynamics.Close(); err != nil {
				return err
			}
		}
	}

	m.log.Info("Simulation manager closed", "hash", m.simHash)

	return nil
}

func (m *Manager) GetStatus() SimulationStatus {
	return m.status
}
