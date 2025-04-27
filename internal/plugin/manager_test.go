package plugin_test

import (
	"testing"

	logger "github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/plugin"

	"github.com/stretchr/testify/require"
)

// TEST: GIVEN a logger WHEN NewManager is called THEN a new Manager is returned
func TestNewManager(t *testing.T) {
	l := logger.GetLogger("debug")
	m := plugin.NewManager(*l)
	require.NotNil(t, m)
	require.Empty(t, m.GetPlugins())
}

// TEST: GIVEN a plugin manager WHEN LoadPlugin is called with an invalid path THEN the plugin manager returns an error
func TestLoadPluginInvalidPath(t *testing.T) {
	l := logger.GetLogger("debug")
	m := plugin.NewManager(*l)
	err := m.LoadPlugin("invalid/path/plugin.so")
	require.Error(t, err)
}

// TEST: GIVEN a valid compiled plugin WHEN LoadPlugin is called THEN the plugin is loaded successfully
func TestLoadPluginSuccess(t *testing.T) {
	l := logger.GetLogger("debug")

	// INFO: Add any built-in plugins here for confidence
	m := plugin.NewManager(*l)
	err := m.LoadPlugin("../../plugins/windeffect.so")
	require.NoError(t, err)
	require.NotEmpty(t, m.GetPlugins())

	err = m.LoadPlugin("../../plugins/blackscholes.so")
	require.NoError(t, err)
	require.NotEmpty(t, m.GetPlugins())
}

// TEST: GIVEN a plugin manager WHEN GetPlugins is called THEN the plugin manager returns the loaded plugins
func TestGetPlugins(t *testing.T) {
	l := logger.GetLogger("debug")
	m := plugin.NewManager(*l)
	require.Empty(t, m.GetPlugins())
}
