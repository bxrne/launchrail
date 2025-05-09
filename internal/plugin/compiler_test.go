package plugin_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bxrne/launchrail/internal/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zerodha/logf"
)

// Helper function to create a dummy plugin source directory
func createDummyPlugin(t *testing.T, baseDir, pluginName, content string) string {
	t.Helper()
	pluginDir := filepath.Join(baseDir, pluginName)
	err := os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)

	if content != "" {
		filePath := filepath.Join(pluginDir, "plugin.go")
		err = os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		// Add a go.mod file to make it a valid module for compilation
		goModPath := filepath.Join(pluginDir, "go.mod")
		goModContent := fmt.Sprintf("module %s\n\ngo 1.23\n", pluginName) // Use plugin name for module path
		err = os.WriteFile(goModPath, []byte(goModContent), 0644)
		require.NoError(t, err)
	}
	return pluginDir
}

// TestCheckDirForGoFiles checks the helper function
func TestCheckDirForGoFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkdirtest-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Case 1: Directory with .go files
	dirWithGo := filepath.Join(tmpDir, "withgo")
	require.NoError(t, os.Mkdir(dirWithGo, 0755))
	_, err = os.Create(filepath.Join(dirWithGo, "main.go"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(dirWithGo, "other.go"))
	require.NoError(t, err)

	// Case 2: Directory with non-.go files
	dirWithOther := filepath.Join(tmpDir, "withother")
	require.NoError(t, os.Mkdir(dirWithOther, 0755))
	_, err = os.Create(filepath.Join(dirWithOther, "script.sh"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(dirWithOther, "README.md"))
	require.NoError(t, err)

	// Case 3: Empty directory
	dirEmpty := filepath.Join(tmpDir, "empty")
	require.NoError(t, os.Mkdir(dirEmpty, 0755))

	// Case 4: Non-existent directory path
	dirNonExistent := filepath.Join(tmpDir, "nonexistent")

	// --- Run Tests ---

	// Test Case 1: Should find .go files
	hasGo, err := plugin.CheckDirForGoFiles(dirWithGo)
	assert.NoError(t, err, "Test Case 1 failed: checking dir with .go files")
	assert.True(t, hasGo, "Test Case 1 failed: expected true for dir with .go files")

	// Test Case 2: Should not find .go files
	hasGo, err = plugin.CheckDirForGoFiles(dirWithOther)
	assert.NoError(t, err, "Test Case 2 failed: checking dir with other files")
	assert.False(t, hasGo, "Test Case 2 failed: expected false for dir with other files")

	// Test Case 3: Should not find .go files in empty dir
	hasGo, err = plugin.CheckDirForGoFiles(dirEmpty)
	assert.NoError(t, err, "Test Case 3 failed: checking empty dir")
	assert.False(t, hasGo, "Test Case 3 failed: expected false for empty dir")

	// Test Case 4: Should return an error for non-existent dir
	_, err = plugin.CheckDirForGoFiles(dirNonExistent)
	assert.Error(t, err, "Test Case 4 failed: expected error for non-existent dir")

	// Test Case 5: Path is a file, not a directory
	filePath := filepath.Join(dirWithGo, "main.go")
	_, err = plugin.CheckDirForGoFiles(filePath)
	assert.Error(t, err, "Test Case 5 failed: expected error when path is a file")
}

func TestCompileAllPlugins(t *testing.T) {
	logger := logf.New(logf.Opts{})

	// Need 'go' command for these tests
	goExec, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("Skipping CompileAllPlugins tests: 'go' executable not found in PATH: %v", err)
	}
	t.Logf("Using go executable: %s", goExec)

	// --- Test Setup ---
	testBaseDir, err := os.MkdirTemp("", "compilepluginstest-")
	require.NoError(t, err)
	defer os.RemoveAll(testBaseDir)

	pluginsSourceDir := filepath.Join(testBaseDir, "plugins-src")
	pluginsOutputDir := filepath.Join(testBaseDir, "plugins-out")
	require.NoError(t, os.Mkdir(pluginsSourceDir, 0755))
	require.NoError(t, os.Mkdir(pluginsOutputDir, 0755))

	// Plugin 1: Valid
	validPluginContent := `
package main

import "fmt"

var V int

func F() { fmt.Printf("Hello, number %d\n", V) }
`
	dir_a := createDummyPlugin(t, pluginsSourceDir, "plugina", validPluginContent)
	require.Equal(t, dir_a, filepath.Join(pluginsSourceDir, "plugina"))

	// Plugin 2: Invalid syntax
	invalidPluginContent := `
package main

func F() { fmt.Println("Invalid plugin }
`
	dir_b := createDummyPlugin(t, pluginsSourceDir, "pluginb", invalidPluginContent)
	require.Equal(t, dir_b, filepath.Join(pluginsSourceDir, "pluginb"))

	// Plugin 3: Valid (different name)
	dir_c := createDummyPlugin(t, pluginsSourceDir, "pluginc", validPluginContent)
	require.Equal(t, dir_c, filepath.Join(pluginsSourceDir, "pluginc"))

	// Directory 4: No .go files
	noGoDir := filepath.Join(pluginsSourceDir, "nogodir")
	require.NoError(t, os.Mkdir(noGoDir, 0755))
	_, err = os.Create(filepath.Join(noGoDir, "readme.txt"))
	require.NoError(t, err)

	// File 5: A file, not a directory
	_, err = os.Create(filepath.Join(pluginsSourceDir, "notadir.txt"))
	require.NoError(t, err)

	// --- Run CompileAllPlugins ---
	err = plugin.CompileAllPlugins(pluginsSourceDir, pluginsOutputDir, logger)

	// --- Assertions ---
	// Expecting an error because pluginb failed
	assert.Error(t, err, "Expected an error due to pluginb failing compilation")
	if err != nil { // Avoid nil panic if assertion fails
		assert.Contains(t, err.Error(), "plugin 'pluginb': compilation command failed", "Error message should mention pluginb failure")
		assert.Contains(t, err.Error(), "1 plugin(s) failed to compile", "Error message should mention count of failures")
	}

	// Check output directory for compiled plugins
	compiledPluginA := filepath.Join(pluginsOutputDir, "plugina.so")
	_, errA := os.Stat(compiledPluginA)
	assert.NoError(t, errA, "Valid plugina.so should exist")

	compiledPluginB := filepath.Join(pluginsOutputDir, "pluginb.so")
	_, errB := os.Stat(compiledPluginB)
	assert.Error(t, errB, "Invalid pluginb.so should NOT exist")
	assert.True(t, os.IsNotExist(errB), "Error for pluginb.so should be 'not exist'")

	compiledPluginC := filepath.Join(pluginsOutputDir, "pluginc.so")
	_, errC := os.Stat(compiledPluginC)
	assert.NoError(t, errC, "Valid pluginc.so should exist")

	// Check that skipped dirs/files didn't produce output
	compiledNogo := filepath.Join(pluginsOutputDir, "nogodir.so")
	_, errNogo := os.Stat(compiledNogo)
	assert.Error(t, errNogo, "nogodir.so should not exist")

	compiledNotadir := filepath.Join(pluginsOutputDir, "notadir.txt.so") // Unlikely name, but check
	_, errNotadir := os.Stat(compiledNotadir)
	assert.Error(t, errNotadir, "notadir.txt.so should not exist")

}

// TODO: Add tests for CompileAllPlugins edge cases:
// - Source directory doesn't exist
// - Output directory doesn't exist (should it be created? Currently assumes exists)
// - Go executable not found (mock exec.LookPath? More complex)
// - Path calculation errors (e.g., permissions, complex relative paths - harder to test reliably)
