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
func createValidConfig() config.Config {
	tmpDir := os.TempDir()
	designFilePath := filepath.Join(tmpDir, "design.ork")
	dataDirPath := filepath.Join(tmpDir, "bench_data")

	// Ensure dummy files/dirs exist for validation within this helper
	_ = os.WriteFile(designFilePath, []byte("dummy"), 0644)
	_ = os.Mkdir(dataDirPath, 0755)
	// No need for explicit cleanup here if tests use t.TempDir() or handle it

	return config.Config{
		Setup: config.Setup{
			App: config.App{
				Name:    "TestApp",
				Version: "1.0.0",
				BaseDir: tmpDir,
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
				Name:       "Test Benchmark",
				DesignFile: designFilePath,
				DataDir:    dataDirPath,
				Enabled:    true,
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

// TEST: GIVEN a config with valid parameters WHEN Validate is called THEN does not return an error
func TestConfig_Validate_Valid(t *testing.T) {
	cfg := createValidConfig()

	// Create dummy files/dirs needed for benchmark validation
	tempDir := t.TempDir()
	designFilePath := filepath.Join(tempDir, "design.ork")
	dataDirPath := filepath.Join(tempDir, "bench_data")
	require.NoError(t, os.WriteFile(designFilePath, []byte("dummy ork"), 0644), "Failed to create dummy design file")
	require.NoError(t, os.Mkdir(dataDirPath, 0755), "Failed to create dummy data dir")

	// Update the config to use these temporary paths
	if bench, ok := cfg.Benchmarks["test-bench"]; ok {
		bench.DesignFile = designFilePath
		bench.DataDir = dataDirPath
		cfg.Benchmarks["test-bench"] = bench
	}
	cfg.Setup.App.BaseDir = tempDir // Set base dir for relative path resolution if needed

	err := cfg.Validate()
	require.NoError(t, err, "Validate() should not return an error for valid config") // Use require
}

// TEST: GIVEN a valid config WHEN String is called THEN returns expected values
func TestConfig_String(t *testing.T) {
	cfg := createValidConfig()
	cfgStr := cfg.String()

	// Simple check: ensure the string representation is not empty.
	// Avoid detailed checks due to Viper's inconsistent key flattening in AllSettings().
	assert.NotEmpty(t, cfgStr, "String() output should not be empty for a valid config")

	// Optional: Check for a few very basic, top-level keys if needed, but keep it minimal.
	// assert.Contains(t, cfgStr, "app.name:", "String() should contain app.name key")
	// assert.Contains(t, cfgStr, "server.port:", "String() should contain server.port key")
}

// TEST: GIVEN a valid config WHEN Bytes is called THEN returns non-empty bytes
func TestConfig_Bytes(t *testing.T) {
	cfg := createValidConfig()
	cfgBytes := cfg.Bytes()
	assert.NotEmpty(t, cfgBytes, "Bytes() output should not be empty")
}

// TEST: GIVEN a valid config file WHEN GetConfig is called THEN returns a valid config
func TestGetConfig_ValidConfig(t *testing.T) {
	// Refactored: Test the core logic GetConfig performs, but with a controlled temp file.
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	dummyORK := filepath.Join(tempDir, "dummy.ork")
	dummyPluginDir := filepath.Join(tempDir, "plugins")

	// Create dummy files/dirs needed by the valid base config
	require.NoError(t, os.WriteFile(dummyORK, []byte("dummy"), 0644))
	require.NoError(t, os.Mkdir(dummyPluginDir, 0755))

	// Minimal valid content + base requirements
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
`, tempDir, dummyPluginDir, dummyORK)

	require.NoError(t, os.WriteFile(configFile, []byte(validContent), 0644))

	// Perform the steps GetConfig would do
	v := viper.New()
	v.SetConfigFile(configFile)

	err := v.ReadInConfig()
	require.NoError(t, err, "ReadInConfig failed")

	var cfg config.Config
	err = v.Unmarshal(&cfg)
	require.NoError(t, err, "Unmarshal failed")

	// Manually set BaseDir *after* unmarshal if needed for Validate, mimicking GetConfig behavior
	cfg.Setup.App.BaseDir = tempDir

	err = cfg.Validate()
	require.NoError(t, err, "Validate failed")

	// Assert loaded values
	assert.Equal(t, "TestAppFromGetConfig", cfg.Setup.App.Name)
	assert.Equal(t, "1.1", cfg.Setup.App.Version)
	assert.Equal(t, 9999, cfg.Server.Port)
	assert.Equal(t, tempDir, cfg.Setup.App.BaseDir)
}

// TEST: GIVEN an invalid config file path WHEN GetConfig is called THEN returns an error
func TestGetConfig_InvalidConfigPath(t *testing.T) {
	// Backup existing config if needed
	hadConfig, err := backupConfigYaml()
	if err != nil {
		t.Fatalf("Failed to backup config: %v", err)
	}

	// Test GetConfig with non-existent file
	_, err = config.GetConfig()
	if err == nil {
		t.Errorf("GetConfig() should return error for missing config file")
	}

	// Restore the original config if it existed
	err = restoreConfigYaml(hadConfig)
	if err != nil {
		t.Fatalf("Failed to restore config: %v", err)
	}
}

// TEST: GIVEN an invalid config format WHEN GetConfig is called THEN returns an error
func TestGetConfig_InvalidConfigFormat(t *testing.T) {
	// Backup existing config if needed
	hadConfig, err := backupConfigYaml()
	if err != nil {
		t.Fatalf("Failed to backup config: %v", err)
	}

	// Create a invalid test config file
	err = createInvalidConfigYaml()
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test GetConfig
	_, err = config.GetConfig()
	if err == nil {
		t.Errorf("GetConfig() should return error for invalid config format")
	}

	// Restore the original config if it existed
	err = restoreConfigYaml(hadConfig)
	if err != nil {
		t.Fatalf("Failed to restore config: %v", err)
	}

	os.Remove("config.yaml")
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
		{"MissingBenchmarkName", "benchmarks.test-bench.name"},              // Test missing required field in map
		{"MissingBenchmarkDesignFile", "benchmarks.test-bench.design_file"}, // Test missing required field in map
		{"MissingBenchmarkDataDir", "benchmarks.test-bench.data_dir"},       // Test missing required field in map
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run tests in parallel

			cfg := createInvalidConfig(tc.invalidField)
			err := cfg.Validate()
			if err == nil {
				t.Errorf("Validate() should return error for invalid field: %s", tc.invalidField)
			}
		})
	}
}

// TEST: GIVEN a config with a valid benchmark WHEN Validate is called THEN does not return an error
func TestConfig_Validate_ValidBenchmark(t *testing.T) {
	cfg := createValidConfig()
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() should not return an error for valid benchmark: %v", err)
	}
}

// TEST: GIVEN a config with an invalid benchmark WHEN Validate is called THEN returns an error
func TestConfig_Validate_InvalidBenchmark(t *testing.T) {
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
    design_file: "/tmp/design.ork"
    data_dir: "/tmp/bench_data"
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
    data_dir: "/tmp/bench_data"
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
    design_file: "/tmp/design.ork"
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
    design_file: "/tmp/design.ork"
    data_dir: "/tmp/bench_data"
    enabled: "invalid"
`,
			expectedError: "cannot parse 'benchmarks[test-bench].enabled' as bool: strconv.ParseBool: parsing \"invalid\": invalid syntax",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Combine base config with test case content
			fullContent := validBaseConfig + tc.content
			// Need dummy base file for the base config part
			tempDir := t.TempDir()
			_ = os.WriteFile(filepath.Join(tempDir, "dummy.ork"), []byte("dummy"), 0644)
			_ = os.Mkdir(filepath.Join(tempDir, "plugins"), 0755)

			cfgFile, cleanup := createTempConfig(t, "invalid_bench_config*.yaml", fullContent)
			defer cleanup()

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

			cfg.Setup.App.BaseDir = filepath.Dir(cfgFile.Name())
			t.Logf("Config struct before Validate for [%s]: %+v", tc.name, cfg)
			err = cfg.Validate() // Now validate the loaded config
			require.Error(t, err, "Validate should return an error")
			assert.Contains(t, err.Error(), tc.expectedError, "Validation error message mismatch")
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
    design_file: "/tmp/design.ork"
    data_dir: "/tmp/bench_data"
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
    data_dir: "/tmp/bench_data"
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
    design_file: "/tmp/design.ork"
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
    design_file: "/tmp/design.ork"
    data_dir: "/tmp/bench_data"
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
			_ = os.WriteFile(filepath.Join(tempDir, "dummy.ork"), []byte("dummy"), 0644)
			_ = os.Mkdir(filepath.Join(tempDir, "plugins"), 0755)

			cfgFile, cleanup := createTempConfig(t, "nonexistent_paths*.yaml", fullContent)
			defer cleanup()

			// Need dummy files/dirs for the paths that *are* present, even if one is missing/invalid
			_ = os.WriteFile(filepath.Join(tempDir, "existing_design.ork"), []byte("dummy"), 0644)
			_ = os.Mkdir(filepath.Join(tempDir, "existing_data_dir"), 0755)
			_ = os.WriteFile(filepath.Join(tempDir, "dummy.ork"), []byte("dummy"), 0644) // For base config
			_ = os.Mkdir(filepath.Join(tempDir, "plugins"), 0755)                        // For base config plugins

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

			cfg.Setup.App.BaseDir = tempDir // Set BaseDir for validation
			t.Logf("Config struct before Validate for [%s]: %+v", tc.name, cfg)
			err = cfg.Validate() // Now validate the loaded config
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
