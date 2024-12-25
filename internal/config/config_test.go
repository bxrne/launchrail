package config_test

import (
	"os"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
)

// Helper to change directory and reset after test
func withWorkingDir(t *testing.T, dir string, testFunc func()) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %s", err)
	}

	// Change to the target directory
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory to %s: %s", dir, err)
	}

	// Ensure reset after test
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to reset directory: %s", err)
		}
	}()

	// Run the test function
	testFunc()
}

// TEST: GIVEN a valid configuration file WHEN GetConfig is called THEN no error is returned
func TestGetConfig(t *testing.T) {
	withWorkingDir(t, "../..", func() {
		cfg, err := config.GetConfig()
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		if cfg == nil {
			t.Error("Expected config to be non-nil")
		}
	})
}

// TEST: GIVEN a configuration file with missing app.name WHEN GetConfig is called THEN an error is returned
func TestValidateConfigMissingAppName(t *testing.T) {
	withWorkingDir(t, "../..", func() {
		cfg, err := config.GetConfig()
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.App.Name = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		if err.Error() != "app.name is required" {
			t.Errorf("Expected error to be 'app.name is required', got: %s", err)
		}

		cfg.App.Name = "launchrail-test" // Reset app.name
	})
}

// TEST: GIVEN a configuration file with missing app.version WHEN GetConfig is called THEN an error is returned
func TestValidateConfigMissingAppVersion(t *testing.T) {
	withWorkingDir(t, "../..", func() {
		cfg, err := config.GetConfig()
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.App.Version = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		if err.Error() != "app.version is required" {
			t.Errorf("Expected error to be 'app.version is required', got: %s", err)
		}

		cfg.App.Version = "0.0.0" // Reset app.version
	})
}

// TEST: GIVEN a configuration file with missing logging.level WHEN GetConfig is called THEN an error is returned
func TestValidateConfigMissingLoggingLevel(t *testing.T) {
	withWorkingDir(t, "../..", func() {
		cfg, err := config.GetConfig()
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Logging.Level = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		if err.Error() != "logging.level is required" {
			t.Errorf("Expected error to be 'logging.level is required', got: %s", err)
		}

		cfg.Logging.Level = "info" // Reset logging.level
	})
}

// TEST: GIVEN a configuration file with missing options.motor_designation WHEN GetConfig is called THEN an error is returned
func TestValidateConfigMissingMotorDesignation(t *testing.T) {
	withWorkingDir(t, "../..", func() {
		cfg, err := config.GetConfig()
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.MotorDesignation = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		if err.Error() != "options.motor_designation is required" {
			t.Errorf("Expected error to be 'options.motor_designation is required', got: %s", err)
		}

		cfg.Options.MotorDesignation = "A8" // Reset options.motor_designation
	})
}
