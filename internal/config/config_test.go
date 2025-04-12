package config_test

import (
	"os"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
)

// createValidConfig returns a fully valid configuration for testing
func createValidConfig() *config.Config {
	return &config.Config{
		Setup: config.Setup{
			App: config.App{
				Name:    "TestApp",
				Version: "1.0.0",
				BaseDir: "/tmp",
			},
			Logging: config.Logging{
				Level: "debug",
			},
			Plugins: config.Plugins{
				Paths: []string{"/opt/plugins"},
			},
		},
		Server: config.Server{
			Port: 8080,
		},
		Engine: config.Engine{
			External: config.External{
				OpenRocketVersion: "23.0",
			},
			Options: config.Options{
				MotorDesignation: "A8-3",
				OpenRocketFile:   "/tmp/rocket.ork",
				Launchrail: config.Launchrail{
					Length:      1.2,
					Angle:       5.0,
					Orientation: 90.0,
				},
				Launchsite: config.Launchsite{
					Latitude:  34.0522,
					Longitude: -118.2437,
					Altitude:  100.0,
					Atmosphere: config.Atmosphere{
						ISAConfiguration: config.ISAConfiguration{
							SpecificGasConstant:  287.058,
							GravitationalAccel:   9.807,
							SeaLevelDensity:      1.225,
							SeaLevelTemperature:  15.0,
							SeaLevelPressure:     101325.0,
							RatioSpecificHeats:   1.40,
							TemperatureLapseRate: 0.0065,
						},
					},
				},
			},
			Simulation: config.Simulation{
				Step:            0.01,
				MaxTime:         10.0,
				GroundTolerance: 0.1,
			},
		},
	}
}

// createInvalidConfig returns a valid config with the specified field set to an invalid value
func createInvalidConfig(invalidField string) *config.Config {
	cfg := createValidConfig()

	switch invalidField {
	case "app.name":
		cfg.Setup.App.Name = ""
	case "app.version":
		cfg.Setup.App.Version = ""
	case "app.base_dir":
		cfg.Setup.App.BaseDir = ""
	case "logging.level":
		cfg.Setup.Logging.Level = ""
	case "external.openrocket_version":
		cfg.Engine.External.OpenRocketVersion = ""
	case "options.motor_designation":
		cfg.Engine.Options.MotorDesignation = ""
	case "options.openrocket_file":
		cfg.Engine.Options.OpenRocketFile = ""
	case "options.launchrail.length":
		cfg.Engine.Options.Launchrail.Length = -1.0
	case "options.launchrail.angle":
		cfg.Engine.Options.Launchrail.Angle = 95.0
	case "options.launchrail.orientation":
		cfg.Engine.Options.Launchrail.Orientation = 370.0
	case "options.launchsite.latitude":
		cfg.Engine.Options.Launchsite.Latitude = 95.0
	case "options.launchsite.longitude":
		cfg.Engine.Options.Launchsite.Longitude = 190.0
	case "options.launchsite.altitude":
		cfg.Engine.Options.Launchsite.Altitude = -10.0
	case "options.launchsite.atmosphere.isa_configuration.specific_gas_constant":
		cfg.Engine.Options.Launchsite.Atmosphere.ISAConfiguration.SpecificGasConstant = 0.0
	case "options.launchsite.atmosphere.isa_configuration.gravitational_accel":
		cfg.Engine.Options.Launchsite.Atmosphere.ISAConfiguration.GravitationalAccel = 0.0
	case "options.launchsite.atmosphere.isa_configuration.sea_level_density":
		cfg.Engine.Options.Launchsite.Atmosphere.ISAConfiguration.SeaLevelDensity = 0.0
	case "options.launchsite.atmosphere.isa_configuration.sea_level_temperature":
		cfg.Engine.Options.Launchsite.Atmosphere.ISAConfiguration.SeaLevelTemperature = -300.0
	case "options.launchsite.atmosphere.isa_configuration.sea_level_pressure":
		cfg.Engine.Options.Launchsite.Atmosphere.ISAConfiguration.SeaLevelPressure = 0.0
	case "options.launchsite.atmosphere.isa_configuration.ratio_specific_heats":
		cfg.Engine.Options.Launchsite.Atmosphere.ISAConfiguration.RatioSpecificHeats = 0.9
	case "options.launchsite.atmosphere.isa_configuration.temperature_lapse_rate":
		cfg.Engine.Options.Launchsite.Atmosphere.ISAConfiguration.TemperatureLapseRate = 0.0
	case "simulation.step":
		cfg.Engine.Simulation.Step = 0.0
	case "simulation.max_time":
		cfg.Engine.Simulation.MaxTime = 0.0
	case "simulation.ground_tolerance":
		cfg.Engine.Simulation.GroundTolerance = -1.0
	case "plugins.paths":
		cfg.Setup.Plugins.Paths = []string{}
	case "server.port":
		cfg.Server.Port = 70000 // Invalid port number
	}

	return cfg
}

// createConfigYaml creates a sample config.yaml file for testing GetConfig
func createConfigYaml() error {
	configContent := `
setup:
  app:
    name: "TestApp"
    version: "1.0.0"
    base_dir: "/tmp"
  logging:
    level: "debug"
  plugins:
    paths:
      - "/opt/plugins"
server:
  port: 8080
engine:
  external:
    openrocket_version: "23.0"
  options:
    motor_designation: "A8-3"
    openrocket_file: "/tmp/rocket.ork"
    launchrail:
      length: 1.2
      angle: 5.0
      orientation: 90.0
    launchsite:
      latitude: 34.0522
      longitude: -118.2437
      altitude: 100.0
      atmosphere:
        isa_configuration:
          specific_gas_constant: 287.058
          gravitational_accel: 9.807
          sea_level_density: 1.225
          sea_level_temperature: 15.0
          sea_level_pressure: 101325.0
          ratio_specific_heats: 1.4
          temperature_lapse_rate: 0.0065
  simulation:
    step: 0.01
    max_time: 10.0
    ground_tolerance: 0.1
`
	return os.WriteFile("config.yaml", []byte(configContent), 0644)
}

// createInvalidConfigYaml creates an invalid YAML file for testing error paths
func createInvalidConfigYaml() error {
	configContent := `
this is not valid yaml
setup:
  app:
    name: TestApp
`
	return os.WriteFile("config.yaml", []byte(configContent), 0644)
}

// backupConfigYaml backs up an existing config.yaml if it exists
func backupConfigYaml() (bool, error) {
	if _, err := os.Stat("config.yaml"); err == nil {
		err = os.Rename("config.yaml", "config.yaml.bak")
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// restoreConfigYaml restores a backed up config.yaml if it exists
func restoreConfigYaml(exists bool) error {
	if exists {
		return os.Rename("config.yaml.bak", "config.yaml")
	}
	return nil
}

// TEST: GIVEN a config with valid parameters WHEN Validate is called THEN does not return an error
func TestConfig_Validate_Valid(t *testing.T) {
	cfg := createValidConfig()
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() should not return an error for valid config: %v", err)
	}
}

// TEST: GIVEN a valid config WHEN String is called THEN returns expected values
func TestConfig_String(t *testing.T) {
	cfg := createValidConfig()
	result := cfg.String()

	// Check a subset of values to verify they're correct
	expectedValues := map[string]string{
		"app.name":                    "TestApp",
		"app.version":                 "1.0.0",
		"app.base_dir":                "/tmp",
		"logging.level":               "debug",
		"options.launchrail.length":   "1.20",
		"options.launchsite.latitude": "34.0522",
		"simulation.step":             "0.0100",
		"plugins.paths":               "[/opt/plugins]",
		"server.port":                 "8080",
	}

	for key, expected := range expectedValues {
		if result[key] != expected {
			t.Errorf("String()[%s] = %v, want %v", key, result[key], expected)
		}
	}
}

// TEST: GIVEN a valid config WHEN Bytes is called THEN returns non-empty bytes
func TestConfig_Bytes(t *testing.T) {
	cfg := createValidConfig()
	bytes := cfg.Bytes()
	if len(bytes) == 0 {
		t.Errorf("Bytes() returned an empty byte array")
	}
}

// TEST: GIVEN a valid config file WHEN GetConfig is called THEN returns a valid config
func TestGetConfig_ValidConfig(t *testing.T) {
	// Backup existing config if needed
	hadConfig, err := backupConfigYaml()
	if err != nil {
		t.Fatalf("Failed to backup config: %v", err)
	}
	defer restoreConfigYaml(hadConfig)

	// Create a test config file
	err = createConfigYaml()
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove("config.yaml")

	// Test GetConfig
	cfg, err := config.GetConfig()
	if err != nil {
		t.Errorf("GetConfig() returned error: %v", err)
	}
	if cfg == nil {
		t.Errorf("GetConfig() returned nil config")
	}

	// Verify a few values from the config
	if cfg.Setup.App.Name != "TestApp" {
		t.Errorf("GetConfig() returned config with incorrect app name: %s", cfg.Setup.App.Name)
	}
}

// TEST: GIVEN an invalid config file path WHEN GetConfig is called THEN returns an error
func TestGetConfig_InvalidConfigPath(t *testing.T) {
	// Backup existing config if needed
	hadConfig, err := backupConfigYaml()
	if err != nil {
		t.Fatalf("Failed to backup config: %v", err)
	}
	defer restoreConfigYaml(hadConfig)

	// Test GetConfig with non-existent file
	_, err = config.GetConfig()
	if err == nil {
		t.Errorf("GetConfig() should return error for missing config file")
	}
}

// TEST: GIVEN an invalid config format WHEN GetConfig is called THEN returns an error
func TestGetConfig_InvalidConfigFormat(t *testing.T) {
	// Backup existing config if needed
	hadConfig, err := backupConfigYaml()
	if err != nil {
		t.Fatalf("Failed to backup config: %v", err)
	}
	defer restoreConfigYaml(hadConfig)

	// Create a invalid test config file
	err = createInvalidConfigYaml()
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove("config.yaml")

	// Test GetConfig
	_, err = config.GetConfig()
	if err == nil {
		t.Errorf("GetConfig() should return error for invalid config format")
	}
}

// TABLE TEST: GIVEN configs with various invalid fields WHEN Validate is called THEN returns appropriate errors
func TestConfig_Validate_InvalidFields(t *testing.T) {
	testCases := []struct {
		name         string
		invalidField string
	}{
		{"MissingAppName", "app.name"},
		{"MissingAppVersion", "app.version"},
		{"MissingBaseDir", "app.base_dir"},
		{"MissingLoggingLevel", "logging.level"},
		{"MissingOpenrocketVersion", "external.openrocket_version"},
		{"MissingMotorDesignation", "options.motor_designation"},
		{"MissingOpenrocketFile", "options.openrocket_file"},
		{"InvalidLaunchrailLength", "options.launchrail.length"},
		{"InvalidLaunchrailAngle", "options.launchrail.angle"},
		{"InvalidLaunchrailOrientation", "options.launchrail.orientation"},
		{"InvalidLaunchsiteLatitude", "options.launchsite.latitude"},
		{"InvalidLaunchsiteLongitude", "options.launchsite.longitude"},
		{"InvalidLaunchsiteAltitude", "options.launchsite.altitude"},
		{"InvalidSpecificGasConstant", "options.launchsite.atmosphere.isa_configuration.specific_gas_constant"},
		{"InvalidGravitationalAccel", "options.launchsite.atmosphere.isa_configuration.gravitational_accel"},
		{"InvalidSeaLevelDensity", "options.launchsite.atmosphere.isa_configuration.sea_level_density"},
		{"InvalidSeaLevelTemperature", "options.launchsite.atmosphere.isa_configuration.sea_level_temperature"},
		{"InvalidSeaLevelPressure", "options.launchsite.atmosphere.isa_configuration.sea_level_pressure"},
		{"InvalidRatioSpecificHeats", "options.launchsite.atmosphere.isa_configuration.ratio_specific_heats"},
		{"InvalidTemperatureLapseRate", "options.launchsite.atmosphere.isa_configuration.temperature_lapse_rate"},
		{"InvalidSimulationStep", "simulation.step"},
		{"InvalidSimulationMaxTime", "simulation.max_time"},
		{"InvalidGroundTolerance", "simulation.ground_tolerance"},
		{"EmptyPluginsPaths", "plugins.paths"},
		{"InvalidServerPort", "server.port"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createInvalidConfig(tc.invalidField)
			err := cfg.Validate()
			if err == nil {
				t.Errorf("Validate() should return error for invalid field: %s", tc.invalidField)
			}
		})
	}
}
