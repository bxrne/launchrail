package plugin_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/plugin"
	"github.com/stretchr/testify/require"
	"github.com/zerodha/logf"
)

// TEST: GIVEN a logger WHEN NewManager is called THEN a new Manager is returned
func TestNewManager(t *testing.T) {
	l := logf.Logger{}
	m := plugin.NewManager(l)
	require.NotNil(t, m)
	require.Empty(t, m.GetPlugins())
}

// GIVEN a plugin manager WHEN LoadPlugin is called with a invalid path THEN the plugin manager returns an error
func TestLoadPluginInvalidPath(t *testing.T) {
	l := logf.Logger{}
	m := plugin.NewManager(l)
	err := m.LoadPlugin("invalid/path")
	require.Error(t, err)
}

// GIVEN a plugin manager WHEN GetPlugins is called THEN the plugin manager returns the loaded plugins
func TestGetPlugins(t *testing.T) {
	l := logf.Logger{}
	m := plugin.NewManager(l)
	require.Empty(t, m.GetPlugins())
}
