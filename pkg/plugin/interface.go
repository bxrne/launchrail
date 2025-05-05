package plugin

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/zerodha/logf"
)

// SimulationPlugin defines the interface that all plugins must implement
type SimulationPlugin interface {
	// Initialize is called when the plugin is loaded
	Initialize(log logf.Logger, cfg *config.Config) error

	// Name returns the unique identifier of the plugin
	Name() string

	// Version returns the plugin version
	Version() string

	// BeforeSimStep is called before each simulation step
	BeforeSimStep(state *states.PhysicsState) error

	// AfterSimStep is called after each simulation step
	AfterSimStep(state *states.PhysicsState) error

	// Cleanup is called when the simulation ends
	Cleanup() error
}
