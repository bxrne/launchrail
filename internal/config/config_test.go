package config_test

import (
	"os"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
)

// Helper to change directory and reset after test
func withWorkingDir(t *testing.T, dir string, testFunc func(cfg *config.Config, err error)) {
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

	// WARN: Run the test function with the configuration and handle its error within
	cfg, err := config.GetConfig()
	testFunc(cfg, err)
}

// TEST: GIVEN a valid configuration file WHEN GetConfig is called THEN no error is returned
func TestGetConfig(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		if cfg == nil {
			t.Error("Expected config to be non-nil")
		}
	})
}

// TEST: GIVEN an invalid config file WHEN GetConfig is called THEN the error 'failed to read config file' is returned
func TestGetConfigInvalidConfigFile(t *testing.T) {
	withWorkingDir(t, ".", func(cfg *config.Config, err error) {
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "failed to read config file:"
		if err.Error()[:len(expected)] != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a bad config file WHEN GetConfig is called THEN the error 'failed to unmarshal config' is returned
func TestGetConfigBadConfigFile(t *testing.T) {
	withWorkingDir(t, "../../testdata/config/bad", func(cfg *config.Config, err error) {
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "failed to unmarshal config"
		if err.Error()[:len(expected)] != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}

	})
}

// TEST: GIVEN a config WHEN another config is requested THEN the config is a singleton
func TestGetConfigSingleton(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg2, err := config.GetConfig()
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		if cfg != cfg2 {
			t.Error("Expected config to be a singleton")
		}
	})
}

// TEST: GIVEN a config with missing app.name WHEN Validate is called THEN an error is returned
func TestGetConfigMissingFields(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.App.Name = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "app.name is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing app.version WHEN Validate is called THEN an error is returned
func TestGetConfigMissingVersion(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.App.Version = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "app.version is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing app.base_dir WHEN Validate is called THEN an error is returned
func TestGetConfigMissingBaseDir(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.App.BaseDir = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "app.base_dir is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing logging.level WHEN Validate is called THEN an error is returned
func TestGetConfigMissingLoggingLevel(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Logging.Level = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "logging.level is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with external.openrocket_version WHEN Validate is called THEN no error is returned
func TestGetConfigExternalOpenRocketVersion(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.External.OpenRocketVersion = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "external.openrocket_version is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing options.motor_designation WHEN Validate is called THEN an error is returned
func TestGetConfigMissingMotorDesignation(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.MotorDesignation = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.motor_designation is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing options.openrocket_file WHEN Validate is called THEN an error is returned
func TestGetConfigMissingOpenRocketFile(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.OpenRocketFile = ""
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.openrocket_file is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing options.launchrail.length WHEN Validate is called THEN no error is returned
func TestGetConfigMissingLaunchrailLength(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchrail.Length = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchrail.length is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing options.launchrail.angle WHEN Validate is called THEN no error is returned
func TestGetConfigMissingLaunchrailAngle(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchrail.Angle = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchrail.angle is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing options.launchrail.orientation WHEN Validate is called THEN no error is returned
func TestGetConfigMissingLaunchrailOrientation(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchrail.Orientation = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchrail.orientation is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing options.launchsite.latitude WHEN Validate is called THEN no error is returned
func TestGetConfigMissingLaunchsiteLatitude(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchsite.Latitude = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchsite.latitude is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing options.launchsite.longitude WHEN Validate is called THEN no error is returned
func TestGetConfigMissingLaunchsiteLongitude(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchsite.Longitude = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchsite.longitude is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing options.launchsite.altitude WHEN Validate is called THEN no error is returned
func TestGetConfigMissingLaunchsiteAltitude(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchsite.Altitude = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchsite.altitude is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with invalid options.openrocket_file WHEN Validate is called THEN an error is returned
func TestGetConfigInvalidOpenRocketFile(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.OpenRocketFile = "invalid"
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.openrocket_file is invalid:"
		if err.Error()[:len(expected)] != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing simulation.step WHEN Validate is called THEN an error is returned
func TestGetConfigMissingSimulationStep(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Simulation.Step = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "simulation.step is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing simulation.max_time WHEN Validate is called THEN an error is returned
func TestGetConfigMissingSimulationMaxTime(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Simulation.MaxTime = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "simulation.max_time is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing atmosphere.isa_configuration.specific_gas_constant WHEN Validate is called THEN an error is returned
func TestGetConfigMissingISAConfigurationSpecificGasConstant(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchsite.Atmosphere.ISAConfiguration.SpecificGasConstant = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchsite.atmosphere.isa_configuration.specific_gas_constant is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing atmosphere.isa_configuration.gravitational_accel WHEN Validate is called THEN an error is returned
func TestGetConfigMissingISAConfigurationGravitationalAccel(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchsite.Atmosphere.ISAConfiguration.GravitationalAccel = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchsite.atmosphere.isa_configuration.gravitational_accel is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing atmosphere.isa_configuration.sea_level_density WHEN Validate is called THEN an error is returned
func TestGetConfigMissingISAConfigurationSeaLevelDensity(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchsite.Atmosphere.ISAConfiguration.SeaLevelDensity = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchsite.atmosphere.isa_configuration.sea_level_density is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing atmosphere.isa_configuration.sea_level_temperature WHEN Validate is called THEN an error is returned
func TestGetConfigMissingISAConfigurationSeaLevelTemperature(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchsite.Atmosphere.ISAConfiguration.SeaLevelTemperature = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchsite.atmosphere.isa_configuration.sea_level_temperature is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing atmosphere.isa_configuration.sea_level_pressure WHEN Validate is called THEN an error is returned
func TestGetConfigMissingISAConfigurationSeaLevelPressure(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchsite.Atmosphere.ISAConfiguration.SeaLevelPressure = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchsite.atmosphere.isa_configuration.sea_level_pressure is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing atmosphere.isa_configuration.ratio_specific_heats WHEN Validate is called THEN an error is returned
func TestGetConfigMissingISAConfigurationRatioSpecificHeats(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchsite.Atmosphere.ISAConfiguration.RatioSpecificHeats = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchsite.atmosphere.isa_configuration.ratio_specific_heats is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}

// TEST: GIVEN a config with missing atmosphere.isa_configuration.temperature_lapse_rate WHEN Validate is called THEN an error is returned
func TestGetConfigMissingISAConfigurationTemperatureLapseRate(t *testing.T) {
	withWorkingDir(t, "../..", func(cfg *config.Config, err error) {
		if err != nil {
			t.Errorf("Expected no error, got: %s", err)
		}

		cfg.Options.Launchsite.Atmosphere.ISAConfiguration.TemperatureLapseRate = 0
		err = cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}

		expected := "options.launchsite.atmosphere.isa_configuration.temperature_lapse_rate is required"
		if err.Error() != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	})
}
