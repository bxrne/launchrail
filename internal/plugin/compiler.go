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

		err := compileSinglePlugin(goExecutable, sourcePath, pluginsOutputDir, pluginName, logger)
		if err != nil {
			logger.Error("Plugin compilation failed", "name", pluginName, "error", err)
			compileErrors = append(compileErrors, fmt.Errorf("plugin '%s': %w", pluginName, err))
		} else {
			// Only log success if compileSinglePlugin didn't return an error (and didn't skip).
			// We are in the else block, meaning err IS nil from compileSinglePlugin.
			// compileSinglePlugin returns nil on success OR skip, so logging here is correct
			// for actual successes, as skips are handled internally in compileSinglePlugin.
			// The previous redundant check `if err == nil` is removed.
			logger.Info("Successfully compiled plugin", "name", pluginName)
			compiledCount++
		}
	}

	logger.Info("Plugin compilation finished", "compiled", compiledCount, "errors", len(compileErrors))

	if len(compileErrors) > 0 {
		// Combine errors (or return the first one, depending on desired behavior)
		// For now, let's just return a generic error indicating failures occurred.
		// return fmt.Errorf("%d plugin(s) failed to compile", len(compileErrors))
		// Combine errors for a more detailed final error message
		var errorMessages []string
		for _, e := range compileErrors {
			errorMessages = append(errorMessages, e.Error())
		}
		return fmt.Errorf("%d plugin(s) failed to compile:\n - %s",
			len(compileErrors), strings.Join(errorMessages, "\n - "))
	}

	return nil
}

// compileSinglePlugin handles the compilation logic for a single plugin.
// It returns an error if compilation fails or an issue occurs before compilation.
// It returns nil if compilation is successful OR if the directory is skipped (e.g., no .go files).
func compileSinglePlugin(goExecutable, sourcePath, pluginsOutputDir, pluginName string, logger logf.Logger) error {
	// Ensure sourcePath and outputPath are absolute for reliable relative path calculation
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute source path: %w", err)
	}
	absOutputPath, err := filepath.Abs(filepath.Join(pluginsOutputDir, pluginName+".so"))
	if err != nil {
		return fmt.Errorf("failed to get absolute output path: %w", err)
	}

	// Basic check: does the directory contain Go files?
	hasGoFiles, err := CheckDirForGoFiles(absSourcePath)
	if err != nil {
		// Log the warning here, but return the error for aggregation
		logger.Warn("Error checking directory for Go files, skipping plugin", "dir", absSourcePath, "error", err)
		return fmt.Errorf("error checking directory for Go files: %w", err)
	}
	if !hasGoFiles {
		logger.Debug("Directory does not contain Go files, skipping", "dir", absSourcePath)
		return nil // Not an error, just skip
	}

	// Calculate the relative path for the output file from the source directory
	relOutputPath, err := filepath.Rel(absSourcePath, absOutputPath)
	if err != nil {
		// Log the error here, but return it for aggregation
		logger.Error("Plugin path calculation error", "name", pluginName, "source", absSourcePath, "output", absOutputPath, "error", err)
		return fmt.Errorf("failed to calculate relative output path: %w", err)
	}

	logger.Info("Compiling plugin", "name", pluginName, "source", absSourcePath, "output", absOutputPath, "relativeOutput", relOutputPath)
	// Use "." for source path argument since cmd.Dir is the source directory
	cmd := exec.Command(goExecutable, "build", "-buildmode=plugin", "-o", relOutputPath, ".")
	cmd.Dir = absSourcePath // Set working directory to the plugin source directory
	cmd.Stderr = os.Stderr  // Pipe build errors to main stderr
	cmd.Stdout = os.Stdout  // Pipe build output to main stdout

	if err := cmd.Run(); err != nil {
		// Don't need to log here, as the calling function will log the aggregated error
		return fmt.Errorf("compilation command failed: %w", err)
	}

	return nil // Success
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
