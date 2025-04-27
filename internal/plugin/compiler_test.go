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

// Mock logger for tests
var testLogger = logf.New(logf.Opts{})

// TestCompileAllPlugins covers various scenarios
func TestCompileAllPlugins(t *testing.T) {
	// Ensure 'go' command exists for actual compilation tests
	_, err := exec.LookPath("go")
	if err != nil {
		t.Skip("Skipping compilation tests: 'go' command not found in PATH")
	}

	// Test case: Successful compilation
	t.Run("SuccessfulCompilation", func(t *testing.T) {
		sourceDir := t.TempDir()
		outputDir := t.TempDir()

		// Create a valid dummy plugin
		dummyContent := `
package main

func main() {}
// Export a symbol to be a valid plugin
var V int
`
		_ = createDummyPlugin(t, sourceDir, "testplugin", dummyContent)

		err := plugin.CompileAllPlugins(sourceDir, outputDir, testLogger)
		assert.NoError(t, err)

		// Check if the compiled plugin exists
		expectedOutputPath := filepath.Join(outputDir, "testplugin.so")
		_, err = os.Stat(expectedOutputPath)
		assert.NoError(t, err, "Compiled plugin file should exist")
	})

	// Test case: Source directory does not exist
	t.Run("SourceDirNotExist", func(t *testing.T) {
		nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")
		outputDir := t.TempDir()
		err := plugin.CompileAllPlugins(nonExistentDir, outputDir, testLogger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read plugins source directory")
	})

	// Test case: Source directory is empty
	t.Run("SourceDirEmpty", func(t *testing.T) {
		sourceDir := t.TempDir()
		outputDir := t.TempDir()
		err := plugin.CompileAllPlugins(sourceDir, outputDir, testLogger)
		assert.NoError(t, err) // Should succeed with 0 plugins compiled
	})

	// Test case: Source directory contains subdirs but no .go files
	t.Run("SourceDirNoGoFiles", func(t *testing.T) {
		sourceDir := t.TempDir()
		outputDir := t.TempDir()
		_ = createDummyPlugin(t, sourceDir, "noplugin", "") // Create dir, but no file
		// Create a non-go file
		filePath := filepath.Join(sourceDir, "noplugin", "readme.txt")
		err := os.WriteFile(filePath, []byte("hello"), 0644)
		require.NoError(t, err)

		err = plugin.CompileAllPlugins(sourceDir, outputDir, testLogger)
		assert.NoError(t, err) // Should succeed, finding 0 plugins to compile
	})

	// Test case: Compilation fails (invalid Go code)
	t.Run("CompilationFailure", func(t *testing.T) {
		sourceDir := t.TempDir()
		outputDir := t.TempDir()

		// Create a plugin with invalid Go code
		dummyContent := `package main this is invalid`
		_ = createDummyPlugin(t, sourceDir, "badplugin", dummyContent)

		err := plugin.CompileAllPlugins(sourceDir, outputDir, testLogger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin(s) failed to compile")

		// Check that the output file was NOT created
		expectedOutputPath := filepath.Join(outputDir, "badplugin.so")
		_, err = os.Stat(expectedOutputPath)
		assert.Error(t, err, "Compiled plugin file should NOT exist on failure")
		assert.True(t, os.IsNotExist(err))
	})

	// Test case: Multiple plugins, one fails
	t.Run("PartialCompilationFailure", func(t *testing.T) {
		sourceDir := t.TempDir()
		outputDir := t.TempDir()

		// Valid plugin
		validContent := `package main
var V int`
		_ = createDummyPlugin(t, sourceDir, "goodplugin", validContent)

		// Invalid plugin
		invalidContent := `package main func ???`
		_ = createDummyPlugin(t, sourceDir, "badplugin2", invalidContent)

		err := plugin.CompileAllPlugins(sourceDir, outputDir, testLogger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "1 plugin(s) failed to compile") // Check error message count

		// Check the good one exists
		goodPluginName := "goodplugin"
		expectedGoodOutputPath := filepath.Join(outputDir, goodPluginName+".so")
		_, err = os.Stat(expectedGoodOutputPath)
		assert.NoError(t, err, "Good plugin should exist")

		// Check the bad one does not exist
		badOutputPath := filepath.Join(outputDir, "badplugin2.so")
		_, err = os.Stat(badOutputPath)
		assert.Error(t, err, "Bad plugin should NOT exist")
		assert.True(t, os.IsNotExist(err))
	})
}

// TestcheckDirForGoFiles checks the helper function
func TestCheckDirForGoFiles(t *testing.T) {
	t.Run("DirWithGoFile", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "main.go")
		err := os.WriteFile(filePath, []byte("package main"), 0644)
		require.NoError(t, err)

		hasGo, err := plugin.CheckDirForGoFiles(dir)
		assert.NoError(t, err)
		assert.True(t, hasGo)
	})

	t.Run("DirWithoutGoFile", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "readme.txt")
		err := os.WriteFile(filePath, []byte("hello"), 0644)
		require.NoError(t, err)

		hasGo, err := plugin.CheckDirForGoFiles(dir)
		assert.NoError(t, err)
		assert.False(t, hasGo)
	})

	t.Run("EmptyDir", func(t *testing.T) {
		dir := t.TempDir()
		hasGo, err := plugin.CheckDirForGoFiles(dir)
		assert.NoError(t, err)
		assert.False(t, hasGo)
	})

	t.Run("NonExistentDir", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "nonexistent")
		_, err := plugin.CheckDirForGoFiles(dir)
		assert.Error(t, err)
	})
}
