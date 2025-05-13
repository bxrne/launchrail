package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Minimal valid base config structure to include in test content
const validBaseConfig = `
setup:
  app:
    name: BaseApp
    version: "1.0"
    base_dir: "."
  logging:
    level: debug
  plugins:
    paths: ["./plugins"]
server:
  port: 8000
  read_timeout_seconds: 15
  write_timeout_seconds: 15
  idle_timeout_seconds: 60
  static_dir: "{{STATIC_DIR}}"
engine:
  external:
    openrocket_version: "23.0"
  options:
    motor_designation: "A1-1"
    openrocket_file: "dummy.ork" # Assume exists relative to BaseDir
    launchrail:
      length: 1.0
      angle: 5.0
      orientation: 90.0
    launchsite:
      latitude: 0.0
      longitude: 0.0
      altitude: 0.0
      atmosphere:
        isa_configuration:
          specific_gas_constant: 287.0
          gravitational_accel: 9.8
          sea_level_density: 1.2
          sea_level_temperature: 15.0
          sea_level_pressure: 101325.0
          ratio_specific_heats: 1.4
          temperature_lapse_rate: 0.0065
  simulation:
    step: 0.01
    max_time: 10.0
    ground_tolerance: 0.1
`

// createTempConfig creates a temporary file with the given content for testing.
// It returns the file handle and a cleanup function.
func createTempConfig(t *testing.T, pattern string, content string) (*os.File, func()) {
	t.Helper()

	// Ensure the pattern includes the .yaml extension for Viper
	if !strings.HasSuffix(pattern, ".yaml") {
		pattern += "*.yaml"
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), pattern)
	require.NoError(t, err, "Failed to create temp config file")

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err, "Failed to write to temp config file")

	err = tmpFile.Close() // Close the file so viper can read it
	require.NoError(t, err, "Failed to close temp config file")

	cleanup := func() {
		// os.Remove(tmpFile.Name()) // t.TempDir() handles cleanup automatically
	}

	// Return the file (re-opened for reading if needed, though viper uses the path)
	// and the cleanup function.
	return tmpFile, cleanup
}

// createValidConfig returns a fully valid configuration for testing
func createValidConfig(staticDir string) config.Config {
	tmpDir := os.TempDir()
	designFilePath := filepath.Join(tmpDir, "design.ork")
	dataDirPath := filepath.Join(tmpDir, "bench_data")
	benchmarkDataDirPath := filepath.Join(tmpDir, "benchdata") // Directory for benchmark data

	// Ensure dummy files/dirs exist for validation within this helper
	_ = os.WriteFile(designFilePath, []byte("dummy"), 0644)
	_ = os.Mkdir(dataDirPath, 0755)
	_ = os.Mkdir(benchmarkDataDirPath, 0755) // Create dummy benchmark data dir
	// No need for explicit cleanup here if tests use t.TempDir() or handle it

	return config.Config{
		Setup: config.Setup{
			App: config.App{
				Name:    "TestApp",
				Version: "1.0.0",
			},
			Logging: config.Logging{
				Level: "debug",
			},
			Plugins: config.Plugins{
				Paths: []string{"/opt/plugins"},
			},
		},
		Server: config.Server{
			Port:         8080,
			ReadTimeout:  15,
			WriteTimeout: 15,
			IdleTimeout:  60,
		},
		Engine: config.Engine{
			External: config.External{
				OpenRocketVersion: "23.0",
			},
			Options: config.Options{
				MotorDesignation: "A8-3",
				OpenRocketFile:   designFilePath, // Use the created path
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
		Benchmarks: map[string]config.BenchmarkEntry{
			"test-bench": {
				Name:             "Test Benchmark",
				Description:      "Detailed description of Test Benchmark",
				DesignFile:       designFilePath,
				DataDir:          dataDirPath,
				Enabled:          true,
				MotorDesignation: "M1297", // Corrected field
			},
		},
	}
}

// createInvalidConfig returns a valid config with the specified field set to an invalid value
func createInvalidConfig(invalidField string, staticDir string) *config.Config {
	cfg := createValidConfig(staticDir)

	switch invalidField {
	case "app.name":
		cfg.Setup.App.Name = ""
	case "app.version":
		cfg.Setup.App.Version = ""
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
	case "benchmarks.test-bench.name":
		bench := cfg.Benchmarks["test-bench"]
		bench.Name = ""
		cfg.Benchmarks["test-bench"] = bench
	case "benchmarks.test-bench.design_file":
		bench := cfg.Benchmarks["test-bench"]
		bench.DesignFile = ""
		cfg.Benchmarks["test-bench"] = bench
	case "benchmarks.test-bench.data_dir":
		bench := cfg.Benchmarks["test-bench"]
		bench.DataDir = ""
		cfg.Benchmarks["test-bench"] = bench
	case "benchmarks.test-bench.enabled":
		bench := cfg.Benchmarks["test-bench"]
		bench.Enabled = false
		cfg.Benchmarks["test-bench"] = bench
	}

	return &cfg
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
func TestConfig_Validate_Valid(t *testing.T) {
	tempDir := t.TempDir()
	configContent := strings.ReplaceAll(validBaseConfig, "{{STATIC_DIR}}", tempDir)
	configFile := filepath.Join(tempDir, "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))
	// Create dummy.ork as referenced by the YAML config
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "dummy.ork"), []byte("dummy"), 0644))

	v := viper.New()
	v.SetConfigFile(configFile)
	err := v.ReadInConfig()
	require.NoError(t, err, "ReadInConfig failed")

	var cfg config.Config
	err = v.Unmarshal(&cfg)
	require.NoError(t, err, "Unmarshal failed")

	err = cfg.Validate(tempDir)
	assert.NoError(t, err, "Validate() should not return an error for valid config")
}

// TEST: GIVEN a valid config file WHEN GetConfig is called THEN returns a valid config
func TestGetConfig_ValidConfig(t *testing.T) {
	// Refactored: Test the core logic GetConfig performs, but with a controlled temp file.
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	dummyORK := filepath.Join(tempDir, "dummy.ork")
	dummyPluginDir := filepath.Join(tempDir, "plugins")
	dummyBenchmarkDataDir := filepath.Join(tempDir, "benchdata") // Directory for benchmark data

	// Create dummy files/dirs needed by the valid base config
	require.NoError(t, os.WriteFile(dummyORK, []byte("dummy"), 0644))
	require.NoError(t, os.Mkdir(dummyPluginDir, 0755))
	require.NoError(t, os.Mkdir(dummyBenchmarkDataDir, 0755)) // Create benchmark data dir

	// Minimal valid content + base requirements + benchmark config
	validContent := fmt.Sprintf(`
setup:
  app:
    name: TestAppFromGetConfig
    version: "1.1"
    base_dir: %q
  logging:
    level: debug
  plugins:
    paths: [%q]
server:
  port: 9999
  read_timeout_seconds: 15
  write_timeout_seconds: 15
  idle_timeout_seconds: 60
  static_dir: %q
engine:
  external:
    openrocket_version: "23.0"
  options:
    motor_designation: "A1-1"
    openrocket_file: %q
    launchrail:
      length: 1.2
      angle: 5
      orientation: 90
    launchsite:
      latitude: 0
      longitude: 0
      altitude: 0
      atmosphere:
        isa_configuration:
          specific_gas_constant: 287.058
          gravitational_accel: 9.80665
          sea_level_density: 1.225
          sea_level_temperature: 288.15
          sea_level_pressure: 101325.0
          ratio_specific_heats: 1.4
          temperature_lapse_rate: 0.0065
  simulation:
    step: 0.01
    max_time: 60
    ground_tolerance: 0.1
`,
		tempDir, dummyPluginDir, tempDir, dummyORK)

	// Substitute a real static_dir
	validContent = strings.ReplaceAll(validContent, "{{STATIC_DIR}}", tempDir)
	validContent = strings.ReplaceAll(validContent, "{{STATIC_DIR}}", tempDir)
	require.NoError(t, os.WriteFile(configFile, []byte(validContent), 0644))
	// Create dummy.ork as referenced by the YAML config
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "dummy.ork"), []byte("dummy"), 0644))

	// Perform the steps GetConfig would do
	v := viper.New()
	v.SetConfigFile(configFile)

	err := v.ReadInConfig()
	require.NoError(t, err, "ReadInConfig failed")

	var cfg config.Config
	err = v.Unmarshal(&cfg)
	require.NoError(t, err, "Unmarshal failed")

	err = cfg.Validate(tempDir) // Pass the tempDir as the configFileDir
	require.NoError(t, err, "Validate failed")

	// Assert loaded values
	assert.Equal(t, "TestAppFromGetConfig", cfg.Setup.App.Name)
	assert.Equal(t, "1.1", cfg.Setup.App.Version)
	assert.Equal(t, 9999, cfg.Server.Port)
}

// TEST: GIVEN a config with a valid benchmark WHEN Validate is called THEN does not return an error
func TestConfig_Validate_ValidBenchmark(t *testing.T) {
	tempDir := t.TempDir()
	configContent := strings.ReplaceAll(validBaseConfig, "{{STATIC_DIR}}", tempDir)
	configFile := filepath.Join(tempDir, "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))
	// Create dummy.ork as referenced by the YAML config
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "dummy.ork"), []byte("dummy"), 0644))

	v := viper.New()
	v.SetConfigFile(configFile)
	err := v.ReadInConfig()
	require.NoError(t, err, "ReadInConfig failed")

	var cfg config.Config
	err = v.Unmarshal(&cfg)
	require.NoError(t, err, "Unmarshal failed")

	err = cfg.Validate(tempDir)
	assert.NoError(t, err, "Validate() should not return an error for valid benchmark")
}

// TEST: GIVEN a config with an invalid benchmark WHEN Validate is called THEN returns an error
func TestConfig_Validate_InvalidBenchmark(t *testing.T) {
	baseCfg := createValidConfig(t.TempDir())
	baseCfg.Benchmarks = make(map[string]config.BenchmarkEntry) // Clear existing benchmarks
	tempDir := t.TempDir()                                      // Base directory for resolving relative paths in tests

	// Create dummy files/dirs that ARE expected to exist for other parts of the config
	dummyOrkPath := filepath.Join(tempDir, "main_rocket.ork")
	_ = os.WriteFile(dummyOrkPath, []byte("dummy data"), 0644)
	baseCfg.Engine.Options.OpenRocketFile = dummyOrkPath

	dummyPluginDir := filepath.Join(tempDir, "test_plugins")
	_ = os.Mkdir(dummyPluginDir, 0755)
	baseCfg.Setup.Plugins.Paths = []string{dummyPluginDir}

	tests := []struct {
		name          string
		content       string
		expectedError string
	}{
		{
			name: "MissingBenchmarkName",
			content: `
benchmarks:
  test-bench:
    design_file: "./existing_design.ork"
    data_dir: "./existing_data_dir"
    enabled: true
`,
			expectedError: "benchmark 'test-bench': benchmark.name is required",
		},
		{
			name: "MissingBenchmarkDesignFile",
			content: `
benchmarks:
  test-bench:
    name: "Test Benchmark"
    data_dir: "./existing_data_dir"
    enabled: true
`,
			expectedError: "benchmark 'test-bench': benchmark.design_file is required",
		},
		{
			name: "MissingBenchmarkDataDir",
			content: `
benchmarks:
  test-bench:
    name: "Test Benchmark"
    design_file: "./existing_design.ork"
    enabled: true
`,
			expectedError: "benchmark 'test-bench': benchmark.data_dir is required",
		},
		{
			name: "InvalidBenchmarkEnabled",
			content: `
benchmarks:
  test-bench:
    name: "Test Benchmark"
    design_file: "./existing_design.ork"
    data_dir: "./existing_data_dir"
    enabled: "invalid"
`,
			expectedError: "cannot parse 'benchmarks[test-bench].enabled' as bool: strconv.ParseBool: parsing \"invalid\": invalid syntax",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Combine base config with test case content
			fullContent := validBaseConfig + tc.content
			// Need dummy base file AND plugin dir for the base config part
			tempDir := t.TempDir()
			baseOrkPath := filepath.Join(tempDir, "dummy_base.ork") // Base ORK for the main config part
			_ = os.WriteFile(baseOrkPath, []byte("dummy base ork"), 0644)
			basePluginDir := filepath.Join(tempDir, "base_plugins")
			_ = os.Mkdir(basePluginDir, 0755)

			// Update fullContent to use these absolute paths for the base part if necessary,
			// or ensure GetConfig (when it unmarshals) will resolve them correctly based on configFileDir.
			// For simplicity, let's assume the `validBaseConfig` string uses relative paths like './dummy.ork'
			// and the test below will use `tempDir` as `configFileDir`.

			cfgFile, cleanup := createTempConfig(t, "nonexistent_paths*.yaml", fullContent)
			defer cleanup()

			// Need dummy files/dirs for the paths that *are* present in the tc.content (benchmark part)
			// These are typically relative in tc.content, so they'll be resolved against tempDir.
			_ = os.WriteFile(filepath.Join(tempDir, "existing_design.ork"), []byte("dummy benchmark design"), 0644)
			_ = os.Mkdir(filepath.Join(tempDir, "existing_data_dir"), 0755)

			v := viper.New()
			v.SetConfigFile(cfgFile.Name())
			err := v.ReadInConfig() // Read the specific temp config
			require.NoError(t, err, "ReadInConfig should succeed for structurally valid YAML, got: %v", err)

			var cfg config.Config // Declare cfg INSIDE the closure
			err = v.Unmarshal(&cfg)

			// Handle InvalidBenchmarkEnabled specifically: expect Unmarshal error
			if strings.Contains(tc.name, "InvalidBenchmarkEnabled") {
				require.Error(t, err, "Unmarshal should fail for invalid boolean syntax string in test [%s]", tc.name)
				assert.Contains(t, err.Error(), tc.expectedError, "Expected Unmarshal syntax error for invalid boolean string in test [%s]", tc.name)
				return // Test passes here
			}

			// For other test cases, Unmarshal MUST succeed
			require.NoError(t, err, "Unmarshal failed unexpectedly for test case [%s]: %v", tc.name, err)

			t.Logf("Config struct before Validate for [%s]: %+v", tc.name, cfg)
			// For Validate, configFileDir is the directory relative to which paths in config are resolved.
			// Since we're constructing paths with tempDir, use tempDir.
			configFileDir := filepath.Dir(cfgFile.Name())

			// Critical: Ensure the unmarshaled cfg has its base paths (OpenRocketFile, PluginDirs) updated
			// to reflect the files created in tempDir if they were defined relatively in validBaseConfig.
			// If `validBaseConfig` has e.g. `openrocket_file: ./dummy.ork`
			// and `plugins_paths: [./plugins]`
			// then we must update them after unmarshal before Validate, or ensure Validate resolves them correctly.
			// The `GetConfig` function itself handles this by passing `configFileDir` to `Validate`.
			// Here, we mimic that behavior.
			// Let's assume `validBaseConfig` uses relative paths `dummy.ork` and `plugins`
			cfg.Engine.Options.OpenRocketFile = baseOrkPath   // Override with the one we created
			cfg.Setup.Plugins.Paths = []string{basePluginDir} // Override with the one we created

			err = cfg.Validate(configFileDir)
			require.Error(t, err, "Validate should return an error")

			// Check if the error is about non-existent paths or missing required fields
			if strings.Contains(tc.name, "Path") {
				assert.Regexp(t, `benchmark 'test-bench'.*(does not exist|is required)`, err.Error(), "Error should be about non-existent path or missing field")
			} else {
				assert.Contains(t, err.Error(), tc.expectedError, "Validation error message mismatch for required fields")
			}
		})
	}
}

// TEST: GIVEN a config with a valid benchmark WHEN Validate is called THEN does not return an error
func TestConfig_Validate_ValidBenchmarkNonExistentPaths(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedError string
	}{
		{
			name: "MissingBenchmarkName",
			content: `
benchmarks:
  test-bench:
    design_file: "./existing_design.ork"
    data_dir: "./existing_data_dir"
    enabled: true
`,
			expectedError: "benchmark 'test-bench': benchmark.name is required",
		},
		{
			name: "MissingBenchmarkDesignFile",
			content: `
benchmarks:
  test-bench:
    name: "Test Benchmark"
    data_dir: "./existing_data_dir"
    enabled: true
`,
			expectedError: "benchmark 'test-bench': benchmark.design_file is required",
		},
		{
			name: "MissingBenchmarkDataDir",
			content: `
benchmarks:
  test-bench:
    name: "Test Benchmark"
    design_file: "./existing_design.ork"
    enabled: true
`,
			expectedError: "benchmark 'test-bench': benchmark.data_dir is required",
		},
		{
			name: "InvalidBenchmarkEnabled",
			content: `
benchmarks:
  test-bench:
    name: "Test Benchmark"
    design_file: "./existing_design.ork"
    data_dir: "./existing_data_dir"
    enabled: "invalid"
`,
			expectedError: "cannot parse 'benchmarks[test-bench].enabled' as bool: strconv.ParseBool: parsing \"invalid\": invalid syntax",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Combine base config with test case content
			fullContent := validBaseConfig + tc.content
			// Need dummy base file AND plugin dir for the base config part
			tempDir := t.TempDir()
			baseOrkPath := filepath.Join(tempDir, "dummy_base.ork") // Base ORK for the main config part
			_ = os.WriteFile(baseOrkPath, []byte("dummy base ork"), 0644)
			basePluginDir := filepath.Join(tempDir, "base_plugins")
			_ = os.Mkdir(basePluginDir, 0755)

			// Update fullContent to use these absolute paths for the base part if necessary,
			// or ensure GetConfig (when it unmarshals) will resolve them correctly based on configFileDir.
			// For simplicity, let's assume the `validBaseConfig` string uses relative paths like './dummy.ork'
			// and the test below will use `tempDir` as `configFileDir`.

			cfgFile, cleanup := createTempConfig(t, "nonexistent_paths*.yaml", fullContent)
			defer cleanup()

			// Need dummy files/dirs for the paths that *are* present in the tc.content (benchmark part)
			// These are typically relative in tc.content, so they'll be resolved against tempDir.
			_ = os.WriteFile(filepath.Join(tempDir, "existing_design.ork"), []byte("dummy benchmark design"), 0644)
			_ = os.Mkdir(filepath.Join(tempDir, "existing_data_dir"), 0755)

			v := viper.New()
			v.SetConfigFile(cfgFile.Name())
			err := v.ReadInConfig() // Read the specific temp config
			require.NoError(t, err, "ReadInConfig should succeed for structurally valid YAML, got: %v", err)

			var cfg config.Config // Declare cfg INSIDE the closure
			err = v.Unmarshal(&cfg)

			// Handle InvalidBenchmarkEnabled specifically: expect Unmarshal error
			if strings.Contains(tc.name, "InvalidBenchmarkEnabled") {
				require.Error(t, err, "Unmarshal should fail for invalid boolean syntax string in test [%s]", tc.name)
				assert.Contains(t, err.Error(), tc.expectedError, "Expected Unmarshal syntax error for invalid boolean string in test [%s]", tc.name)
				return // Test passes here
			}

			// For other test cases, Unmarshal MUST succeed
			require.NoError(t, err, "Unmarshal failed unexpectedly for test case [%s]: %v", tc.name, err)

			t.Logf("Config struct before Validate for [%s]: %+v", tc.name, cfg)
			// For Validate, configFileDir is the directory relative to which paths in config are resolved.
			// Since we're constructing paths with tempDir, use tempDir.
			configFileDir := filepath.Dir(cfgFile.Name())

			// Critical: Ensure the unmarshaled cfg has its base paths (OpenRocketFile, PluginDirs) updated
			// to reflect the files created in tempDir if they were defined relatively in validBaseConfig.
			// If `validBaseConfig` has e.g. `openrocket_file: ./dummy.ork`
			// and `plugins_paths: [./plugins]`
			// then we must update them after unmarshal before Validate, or ensure Validate resolves them correctly.
			// The `GetConfig` function itself handles this by passing `configFileDir` to `Validate`.
			// Here, we mimic that behavior.
			// Let's assume `validBaseConfig` uses relative paths `dummy.ork` and `plugins`
			cfg.Engine.Options.OpenRocketFile = baseOrkPath   // Override with the one we created
			cfg.Setup.Plugins.Paths = []string{basePluginDir} // Override with the one we created

			err = cfg.Validate(configFileDir)
			require.Error(t, err, "Validate should return an error")

			// Check if the error is about non-existent paths or missing required fields
			if strings.Contains(tc.name, "Path") {
				assert.Regexp(t, `benchmark 'test-bench'.*(does not exist|is required)`, err.Error(), "Error should be about non-existent path or missing field")
			} else {
				assert.Contains(t, err.Error(), tc.expectedError, "Validation error message mismatch for required fields")
			}
		})
	}
}

// TEST: GIVEN a valid config WHEN ToMap is called THEN returns a map with correct stringified values
func TestConfig_ToMap(t *testing.T) {
	tc := createValidConfig(t.TempDir()) // Use the existing helper to get a valid config

	// Create dummy files/dirs needed by the valid config for path resolution if ToMap relies on it
	// (Although ToMap primarily just stringifies existing fields, good practice if validation is implicitly part of it)
	tempDir := t.TempDir()
	dummyOrkFile, err := os.Create(filepath.Join(tempDir, "test.ork"))
	require.NoError(t, err)
	dummyOrkFile.Close()
	tc.Engine.Options.OpenRocketFile = dummyOrkFile.Name()

	dummyPluginDir := filepath.Join(tempDir, "plugins")
	err = os.Mkdir(dummyPluginDir, 0755)
	require.NoError(t, err)
	tc.Setup.Plugins.Paths = []string{dummyPluginDir}

	// Create dummy benchmark files/dirs
	dummyBenchDesignFile, err := os.Create(filepath.Join(tempDir, "bench_design.ork"))
	require.NoError(t, err)
	dummyBenchDesignFile.Close()
	dummyBenchDataDir := filepath.Join(tempDir, "bench_data_dir")
	err = os.Mkdir(dummyBenchDataDir, 0755)
	require.NoError(t, err)

	for k, bench := range tc.Benchmarks {
		bench.DesignFile = dummyBenchDesignFile.Name()
		bench.DataDir = dummyBenchDataDir
		tc.Benchmarks[k] = bench
	}

	configMap := tc.ToMap()

	assert.NotEmpty(t, configMap, "Map should not be empty")

	// Check a few key values
	assert.Equal(t, "TestApp", configMap["app.name"])
	assert.Equal(t, "1.0.0", configMap["app.version"])
	assert.Equal(t, "debug", configMap["logging.level"])
	assert.Equal(t, "8080", configMap["server.port"])
	assert.Equal(t, "23.0", configMap["external.openrocket_version"])
	assert.Equal(t, "A8-3", configMap["options.motor_designation"])
	assert.Equal(t, dummyOrkFile.Name(), configMap["options.openrocket_file"])

	// Launchrail
	assert.Equal(t, "1.20", configMap["options.launchrail.length"])
	assert.Equal(t, "5.00", configMap["options.launchrail.angle"])
	assert.Equal(t, "90.00", configMap["options.launchrail.orientation"])

	// Launchsite
	assert.Equal(t, "34.0522", configMap["options.launchsite.latitude"])
	assert.Equal(t, "-118.2437", configMap["options.launchsite.longitude"])
	assert.Equal(t, "100.00", configMap["options.launchsite.altitude"])

	// Atmosphere
	assert.Equal(t, "287.058", configMap["options.launchsite.atmosphere.isa_configuration.specific_gas_constant"])
	assert.Equal(t, "9.807", configMap["options.launchsite.atmosphere.isa_configuration.gravitational_accel"])
	assert.Equal(t, "1.225", configMap["options.launchsite.atmosphere.isa_configuration.sea_level_density"])
	assert.Equal(t, "15.00", configMap["options.launchsite.atmosphere.isa_configuration.sea_level_temperature"])
	assert.Equal(t, "101325.00", configMap["options.launchsite.atmosphere.isa_configuration.sea_level_pressure"])
	assert.Equal(t, "1.40", configMap["options.launchsite.atmosphere.isa_configuration.ratio_specific_heats"])
	assert.Equal(t, "0.01", configMap["options.launchsite.atmosphere.isa_configuration.temperature_lapse_rate"])

	// Simulation
	assert.Equal(t, "0.0100", configMap["simulation.step"])
	assert.Equal(t, "10.00", configMap["simulation.max_time"])
	assert.Equal(t, "0.10", configMap["simulation.ground_tolerance"])

	// Plugins
	// ToMap stores plugins.paths as a Go-syntax string representation of a slice, e.g., "[/path/to/plugin1 /path/to/plugin2]"
	expectedPluginPathString := fmt.Sprintf("[%s]", dummyPluginDir) // For a single path
	assert.Equal(t, expectedPluginPathString, configMap["plugins.paths"])

	// Benchmarks - check one entry
	assert.Equal(t, "Test Benchmark", configMap["benchmarks.test-bench.name"])
	assert.Equal(t, "Detailed description of Test Benchmark", configMap["benchmarks.test-bench.description"])
	assert.Equal(t, dummyBenchDesignFile.Name(), configMap["benchmarks.test-bench.design_file"])
	assert.Equal(t, dummyBenchDataDir, configMap["benchmarks.test-bench.data_dir"])
	assert.Equal(t, "M1297", configMap["benchmarks.test-bench.motor_designation"])
	assert.Equal(t, "true", configMap["benchmarks.test-bench.enabled"])

	// Test ToMap with empty plugin paths
	tcNoPlugins := createValidConfig(t.TempDir())
	tcNoPlugins.Setup.Plugins.Paths = []string{}
	configMapNoPlugins := tcNoPlugins.ToMap()
	assert.Equal(t, "", configMapNoPlugins["plugins.paths"], "plugins.paths should be empty string if no paths")
}

// TEST: GIVEN a valid config WHEN Bytes is called THEN returns a non-empty byte slice
func TestConfig_Bytes(t *testing.T) {
	tc := createValidConfig(t.TempDir()) // Use the existing helper

	// Create dummy files/dirs needed by the valid config
	tempDir := t.TempDir()
	dummyOrkFile, err := os.Create(filepath.Join(tempDir, "test.ork"))
	require.NoError(t, err)
	dummyOrkFile.Close()
	tc.Engine.Options.OpenRocketFile = dummyOrkFile.Name()

	dummyPluginDir := filepath.Join(tempDir, "plugins")
	err = os.Mkdir(dummyPluginDir, 0755)
	require.NoError(t, err)
	tc.Setup.Plugins.Paths = []string{dummyPluginDir}

	// Create dummy benchmark files/dirs
	dummyBenchDesignFile, err := os.Create(filepath.Join(tempDir, "bench_design.ork"))
	require.NoError(t, err)
	dummyBenchDesignFile.Close()
	dummyBenchDataDir := filepath.Join(tempDir, "bench_data_dir")
	err = os.Mkdir(dummyBenchDataDir, 0755)
	require.NoError(t, err)

	for k, bench := range tc.Benchmarks {
		bench.DesignFile = dummyBenchDesignFile.Name()
		bench.DataDir = dummyBenchDataDir
		tc.Benchmarks[k] = bench
	}

	configBytes := tc.Bytes()

	assert.NotEmpty(t, configBytes, "Byte slice should not be empty")

	// Optionally, check if some key substrings are present
	// This can be brittle if the %+v format changes, but can catch major issues.
	configString := string(configBytes)
	assert.Contains(t, configString, "TestApp", "Byte slice string should contain app name")
	assert.Contains(t, configString, "8080", "Byte slice string should contain server port")
	assert.Contains(t, configString, "A8-3", "Byte slice string should contain motor designation")

}
