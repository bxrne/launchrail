package plugin

import (
	"fmt"
	"plugin"

	pluginapi "github.com/bxrne/launchrail/pkg/plugin"
	"github.com/zerodha/logf"
)

type Manager struct {
	plugins []pluginapi.SimulationPlugin
	log     logf.Logger
}

func NewManager(log logf.Logger) *Manager {
	return &Manager{
		plugins: make([]pluginapi.SimulationPlugin, 0),
		log:     log,
	}
}

func (m *Manager) GetPlugins() []pluginapi.SimulationPlugin {
	return m.plugins
}

func (m *Manager) LoadPlugin(pluginPath string) error {

	// Load the plugin
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %w", pluginPath, err)
	}

	// Look up the plugin symbol
	symPlugin, err := plug.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("plugin %s does not export 'Plugin' symbol: %w", pluginPath, err)
	}

	// Assert the plugin implements our interface
	simulationPlugin, ok := symPlugin.(pluginapi.SimulationPlugin)
	if !ok {
		return fmt.Errorf("plugin %s does not implement SimulationPlugin interface", pluginPath)
	}

	// Initialize the plugin
	if err := simulationPlugin.Initialize(m.log); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", pluginPath, err)
	}

	m.plugins = append(m.plugins, simulationPlugin)
	return nil
}
