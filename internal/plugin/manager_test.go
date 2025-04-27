package plugin_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/plugin"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stretchr/testify/require"
	"github.com/zerodha/logf"
)

// compileTestPlugin creates and compiles a simple dummy plugin for testing.
// It returns the path to the compiled .so file and a cleanup function.
func compileTestPlugin(t *testing.T, pluginName string) (string, func()) {
	t.Helper()

	// Find go executable
	goExecutable, err := exec.LookPath("go")
	require.NoError(t, err, "go executable not found")

	// Create temp source directory
	sourceDir := t.TempDir()

	// Create plugin subdirectory
	pluginSourceDir := filepath.Join(sourceDir, pluginName)
	err = os.Mkdir(pluginSourceDir, 0755)
	require.NoError(t, err)

	// Find project root by looking for go.mod
	projectRoot, err := findProjectRoot(".")
	require.NoError(t, err, "Failed to find project root")

	// Copy go.mod and go.sum to the temp plugin source directory
	for _, filename := range []string{"go.mod", "go.sum"} {
		srcPath := filepath.Join(projectRoot, filename)
		dstPath := filepath.Join(pluginSourceDir, filename)
		input, err := os.ReadFile(srcPath)
		require.NoError(t, err, "Failed to read %s from project root", filename)
		err = os.WriteFile(dstPath, input, 0644)
		require.NoError(t, err, "Failed to write %s to temp dir", filename)
		// Verify the file was written
		_, err = os.Stat(dstPath)
		require.NoError(t, err, "Failed to stat copied %s in temp dir", filename)
	}

	// Write dummy plugin code
	pluginCodePath := filepath.Join(pluginSourceDir, "plugin.go")
	// Generate code that exports a 'Plugin' symbol implementing the pkg/plugin.Plugin interface
	pluginCode := fmt.Sprintf(`package main

import "github.com/bxrne/launchrail/pkg/states"

// DummyPlugin satisfies the plugin.Plugin interface for testing.
type DummyPlugin struct{}

// Name returns the plugin name.
func (p *DummyPlugin) Name() string {
	return "%s"
}
// Execute is a dummy implementation.
func (p *DummyPlugin) Execute(state *states.State) error {
	return nil
}
// Plugin is the exported symbol required by the plugin manager.
var Plugin = &DummyPlugin{}

func main() {} // Required for compilation`, pluginName)
	err = os.WriteFile(pluginCodePath, []byte(pluginCode), 0644)
	require.NoError(t, err, "Failed to write dummy plugin code")

	// Run go mod tidy in the temp dir to ensure dependencies are recognized
	tidyCmd := exec.Command(goExecutable, "mod", "tidy")
	tidyCmd.Dir = pluginSourceDir
	tidyOutput, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed in temp dir: %s", string(tidyOutput))

	// Create temp output directory
	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, pluginName+".so")

	// Compile the plugin
	cmd := exec.Command(goExecutable, "build", "-buildmode=plugin", "-o", outputPath, "plugin.go") // Use relative path
	cmd.Dir = pluginSourceDir // Set working directory
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Plugin compilation failed: %s", string(output))

	// Check if the .so file was created
	_, err = os.Stat(outputPath)
	require.NoError(t, err, "Compiled plugin .so file not found")

	cleanup := func() {
		// Temp dirs are cleaned up automatically by t.TempDir()
	}

	return outputPath, cleanup
}

// findProjectRoot searches upwards from the startPath for a directory containing go.mod
func findProjectRoot(startPath string) (string, error) {
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", startPath, err)
	}

	for {
		goModPath := filepath.Join(currentPath, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod
			return currentPath, nil
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached the root directory without finding go.mod
			break
		}
		currentPath = parentPath
	}

	return "", fmt.Errorf("go.mod not found in any parent directory of %s", startPath)
}

// TEST: GIVEN a logger WHEN NewManager is called THEN a new Manager is returned
func TestNewManager(t *testing.T) {
	l := logf.Logger{}
	m := plugin.NewManager(l)
	require.NotNil(t, m)
	require.Empty(t, m.GetPlugins())
}

// TEST: GIVEN a plugin manager WHEN LoadPlugin is called with an invalid path THEN the plugin manager returns an error
func TestLoadPluginInvalidPath(t *testing.T) {
	l := logf.Logger{}
	m := plugin.NewManager(l)
	err := m.LoadPlugin("invalid/path/plugin.so")
	require.Error(t, err)
}

// TEST: GIVEN a valid compiled plugin WHEN LoadPlugin is called THEN the plugin is loaded successfully
func TestLoadPluginSuccess(t *testing.T) {
	pluginPath, cleanup := compileTestPlugin(t, "testplugin1")
	defer cleanup()

	l := logf.Logger{}
	m := plugin.NewManager(l)

	err := m.LoadPlugin(pluginPath)
	require.NoError(t, err)

	loadedPlugins := m.GetPlugins()
	require.Len(t, loadedPlugins, 1, "Expected 1 plugin to be loaded")
}

// TEST: GIVEN a plugin manager WHEN GetPlugins is called THEN the plugin manager returns the loaded plugins
func TestGetPlugins(t *testing.T) {
	l := logf.Logger{}
	m := plugin.NewManager(l)
	require.Empty(t, m.GetPlugins())
}
