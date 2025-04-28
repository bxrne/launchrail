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
		// Ensure sourcePath and outputPath are absolute for reliable relative path calculation
		sourcePath, err := filepath.Abs(filepath.Join(pluginsSourceDir, pluginName))
		if err != nil {
			logger.Warn("Failed to get absolute path for source, skipping", "dir", pluginName, "error", err)
			continue
		}
		outputPath, err := filepath.Abs(filepath.Join(pluginsOutputDir, pluginName+".so"))
        if err != nil {
			logger.Warn("Failed to get absolute path for output, skipping", "dir", pluginName, "error", err)
			continue
		}

		// Basic check: does the directory contain Go files?
		hasGoFiles, err := CheckDirForGoFiles(sourcePath)
		if err != nil {
			logger.Warn("Error checking directory for Go files, skipping", "dir", sourcePath, "error", err)
			continue
		}
		if !hasGoFiles {
			logger.Debug("Directory does not contain Go files, skipping", "dir", sourcePath)
			continue
		}

		// Calculate the relative path for the output file from the source directory
		relOutputPath, err := filepath.Rel(sourcePath, outputPath)
		if err != nil {
			err := fmt.Errorf("failed to calculate relative output path for plugin '%s': %w", pluginName, err)
			logger.Error("Plugin path calculation error", "name", pluginName, "source", sourcePath, "output", outputPath, "error", err)
			compileErrors = append(compileErrors, err)
			continue
		}

		logger.Info("Compiling plugin", "name", pluginName, "source", sourcePath, "output", outputPath, "relativeOutput", relOutputPath)
        // Use "." for source path argument since cmd.Dir is the source directory
		cmd := exec.Command(goExecutable, "build", "-buildmode=plugin", "-o", relOutputPath, ".")
		cmd.Dir = sourcePath   // Set working directory to the plugin source directory
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

// CheckDirForGoFiles checks if a directory contains any .go files.
func CheckDirForGoFiles(dirPath string) (bool, error) {
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
