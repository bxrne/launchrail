package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zerodha/logf"
)

// CompilePlugins is a function variable that points to CompileAllPlugins by default.
// It can be overridden in tests to mock plugin compilation.
var CompilePlugins = CompileAllPlugins

// CompileAllPlugins finds Go plugin source directories and compiles them.
func CompileAllPlugins(pluginsSourceDir, pluginsOutputDir string, logger logf.Logger) error {
	logger.Info("Starting plugin compilation", "sourceDir", pluginsSourceDir, "outputDir", pluginsOutputDir)

	goExecutable, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("could not find 'go' executable in PATH: %w", err)
	}
	logger.Debug("Found 'go' executable", "path", goExecutable)

	entries, err := os.ReadDir(pluginsSourceDir)
	if err != nil {
		return fmt.Errorf("failed to read plugins source directory '%s': %w", pluginsSourceDir, err)
	}

	var compileErrors []error
	compiledCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginName := entry.Name()
		sourcePath := filepath.Join(pluginsSourceDir, pluginName)
		outputPath := filepath.Join(pluginsOutputDir, pluginName+".so")

		// Basic check: does the directory contain Go files?
		hasGoFiles, err := checkDirForGoFiles(sourcePath)
		if err != nil {
			logger.Warn("Error checking directory for Go files, skipping", "dir", sourcePath, "error", err)
			continue
		}
		if !hasGoFiles {
			logger.Debug("Directory does not contain Go files, skipping", "dir", sourcePath)
			continue
		}

		logger.Info("Compiling plugin", "name", pluginName, "source", sourcePath, "output", outputPath)
		cmd := exec.Command(goExecutable, "build", "-buildmode=plugin", "-o", outputPath, sourcePath)
		cmd.Stderr = os.Stderr // Pipe build errors to main stderr
		cmd.Stdout = os.Stdout // Pipe build output to main stdout

		if err := cmd.Run(); err != nil {
			compileErr := fmt.Errorf("failed to compile plugin '%s': %w", pluginName, err)
			logger.Error("Plugin compilation failed", "name", pluginName, "error", err)
			compileErrors = append(compileErrors, compileErr)
		} else {
			logger.Info("Successfully compiled plugin", "name", pluginName, "output", outputPath)
			compiledCount++
		}
	}

	logger.Info("Plugin compilation finished", "compiled", compiledCount, "errors", len(compileErrors))

	if len(compileErrors) > 0 {
		// Combine errors (or return the first one, depending on desired behavior)
		// For now, let's just return a generic error indicating failures occurred.
		return fmt.Errorf("%d plugin(s) failed to compile", len(compileErrors))
	}

	return nil
}

// checkDirForGoFiles checks if a directory contains any .go files.
func checkDirForGoFiles(dirPath string) (bool, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true, nil
		}
	}
	return false, nil
}
