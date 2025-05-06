package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	v *viper.Viper // Package-level viper instance, used by Validate via ConfigFileUsed()
)

// GetConfig reads configuration, validates it, resolves paths, and returns a new instance.
func GetConfig() (*Config, error) {
	v = viper.New() // Initialize package-level viper instance for this call's context

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")     // Look for config in the current directory
	v.AddConfigPath("..")    // Look for config in the parent directory
	v.AddConfigPath("../..") // Look for config in the grandparent directory

	// Attempt to find and read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read mandatory config file: %w", err)
	}

	currentCfg := &Config{}
	if err := v.Unmarshal(currentCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	var configFileDir string
	configFilePath := v.ConfigFileUsed()
	if configFilePath != "" {
		configFileDir = filepath.Dir(configFilePath)
	} else {
		// Fallback if viper doesn't provide the config file path (e.g., if config is set programmatically without a file)
		cwd, err := os.Getwd()
		if err != nil {
			// This is a more critical fallback failure, as we can't determine a base dir.
			return nil, fmt.Errorf("could not determine config file directory: viper.ConfigFileUsed() is empty and os.Getwd() failed: %w", err)
		}
		configFileDir = cwd
		// Potentially log a warning here if this path is taken in a non-test environment.
	}

	if err := currentCfg.Validate(configFileDir); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return currentCfg, nil
}

// App represents the application configuration.
type App struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// Logging represents the logging configuration.
type Logging struct {
	Level string `mapstructure:"level"`
}

// External represents the external configuration.
type External struct {
	OpenRocketVersion string `mapstructure:"openrocket_version"`
}

// Plugins represents runtime plugins to enrich the simulation
type Plugins struct {
	Paths []string `mapstructure:"paths"`
}

// Launchrail represents the launchrail configuration.
type Launchrail struct {
	Length      float64 `mapstructure:"length"`
	Angle       float64 `mapstructure:"angle"`
	Orientation float64 `mapstructure:"orientation"`
}

// Launchsite represents the launchsite configuration.
type Launchsite struct {
	Latitude   float64    `mapstructure:"latitude"`
	Longitude  float64    `mapstructure:"longitude"`
	Altitude   float64    `mapstructure:"altitude"`
	Atmosphere Atmosphere `mapstructure:"atmosphere"`
}

// Atmosphere represents the atmosphere configuration.
type Atmosphere struct {
	ISAConfiguration ISAConfiguration `mapstructure:"isa_configuration"`
}

// ISAConfiguration represents the ISA configuration.
type ISAConfiguration struct {
	SpecificGasConstant  float64 `mapstructure:"specific_gas_constant"`
	GravitationalAccel   float64 `mapstructure:"gravitational_accel"`
	SeaLevelDensity      float64 `mapstructure:"sea_level_density"`
	SeaLevelTemperature  float64 `mapstructure:"sea_level_temperature"`
	SeaLevelPressure     float64 `mapstructure:"sea_level_pressure"`
	RatioSpecificHeats   float64 `mapstructure:"ratio_specific_heats"`
	TemperatureLapseRate float64 `mapstructure:"temperature_lapse_rate"`
}

// Options represents the application options.
type Options struct {
	MotorDesignation string     `mapstructure:"motor_designation"`
	OpenRocketFile   string     `mapstructure:"openrocket_file"`
	Launchrail       Launchrail `mapstructure:"launchrail"`
	Launchsite       Launchsite `mapstructure:"launchsite"`
}

// Simulation represents the simulation configuration.
type Simulation struct {
	Step            float64 `mapstructure:"step"`
	MaxTime         float64 `mapstructure:"max_time"`
	GroundTolerance float64 `mapstructure:"ground_tolerance"` // Add ground tolerance
}

// Engine represents the engine configuration (simulation specific).
type Engine struct {
	External   External   `mapstructure:"external"`
	Options    Options    `mapstructure:"options"`
	Simulation Simulation `mapstructure:"simulation"`
}

// Setup represents the setup configuration.
type Setup struct {
	App     App     `mapstructure:"app"`
	Logging Logging `mapstructure:"logging"`
	Plugins Plugins `mapstructure:"plugins"`
}

// Server represents the server configuration.
type Server struct {
	Port int `mapstructure:"port"`
}

// BenchmarkEntry defines the configuration for a single benchmark.
type BenchmarkEntry struct {
	Name             string `mapstructure:"name" validate:"required"`
	Description      string `mapstructure:"description"` // Added missing field
	DesignFile       string `mapstructure:"design_file" validate:"required,file"`
	DataDir          string `mapstructure:"data_dir" validate:"required,dir"`
	MotorDesignation string `mapstructure:"motor_designation"` // Added motor designation
	Enabled          bool   `mapstructure:"enabled" validate:"boolean"`
}

// Config represents the overall application configuration.
type Config struct {
	Setup      Setup                     `mapstructure:"setup"`
	Server     Server                    `mapstructure:"server"`
	Engine     Engine                    `mapstructure:"engine"`
	Benchmarks map[string]BenchmarkEntry `mapstructure:"benchmarks"`
}

// ToMap converts the configuration to a map of strings.
func (c *Config) ToMap() map[string]string {
	marshalled := make(map[string]string)

	// Setup Config
	marshalled["app.name"] = c.Setup.App.Name
	marshalled["app.version"] = c.Setup.App.Version
	marshalled["logging.level"] = c.Setup.Logging.Level

	// Engine -> External
	marshalled["external.openrocket_version"] = c.Engine.External.OpenRocketVersion

	// Engine -> Options
	marshalled["options.motor_designation"] = c.Engine.Options.MotorDesignation
	marshalled["options.openrocket_file"] = c.Engine.Options.OpenRocketFile

	// Launchrail
	marshalled["options.launchrail.length"] = fmt.Sprintf("%.2f", c.Engine.Options.Launchrail.Length)
	marshalled["options.launchrail.angle"] = fmt.Sprintf("%.2f", c.Engine.Options.Launchrail.Angle)
	marshalled["options.launchrail.orientation"] = fmt.Sprintf("%.2f", c.Engine.Options.Launchrail.Orientation)

	// Launchsite
	marshalled["options.launchsite.latitude"] = fmt.Sprintf("%.4f", c.Engine.Options.Launchsite.Latitude)
	marshalled["options.launchsite.longitude"] = fmt.Sprintf("%.4f", c.Engine.Options.Launchsite.Longitude)
	marshalled["options.launchsite.altitude"] = fmt.Sprintf("%.2f", c.Engine.Options.Launchsite.Altitude)

	// Atmosphere
	isa := c.Engine.Options.Launchsite.Atmosphere.ISAConfiguration
	marshalled["options.launchsite.atmosphere.isa_configuration.specific_gas_constant"] = fmt.Sprintf("%.3f", isa.SpecificGasConstant)
	marshalled["options.launchsite.atmosphere.isa_configuration.gravitational_accel"] = fmt.Sprintf("%.3f", isa.GravitationalAccel)
	marshalled["options.launchsite.atmosphere.isa_configuration.sea_level_density"] = fmt.Sprintf("%.3f", isa.SeaLevelDensity)
	marshalled["options.launchsite.atmosphere.isa_configuration.sea_level_temperature"] = fmt.Sprintf("%.2f", isa.SeaLevelTemperature)
	marshalled["options.launchsite.atmosphere.isa_configuration.sea_level_pressure"] = fmt.Sprintf("%.2f", isa.SeaLevelPressure)
	marshalled["options.launchsite.atmosphere.isa_configuration.ratio_specific_heats"] = fmt.Sprintf("%.2f", isa.RatioSpecificHeats)
	marshalled["options.launchsite.atmosphere.isa_configuration.temperature_lapse_rate"] = fmt.Sprintf("%.2f", isa.TemperatureLapseRate)

	// Simulation
	marshalled["simulation.step"] = fmt.Sprintf("%.4f", c.Engine.Simulation.Step)
	marshalled["simulation.max_time"] = fmt.Sprintf("%.2f", c.Engine.Simulation.MaxTime)
	marshalled["simulation.ground_tolerance"] = fmt.Sprintf("%.2f", c.Engine.Simulation.GroundTolerance)

	// Plugins - Store full list as comma-separated values
	if len(c.Setup.Plugins.Paths) > 0 {
		marshalled["plugins.paths"] = fmt.Sprintf("%v", c.Setup.Plugins.Paths)
	} else {
		marshalled["plugins.paths"] = ""
	}

	// Server Port
	marshalled["server.port"] = fmt.Sprintf("%d", c.Server.Port)

	for tag, benchmark := range c.Benchmarks {
		marshalled["benchmarks."+tag+".name"] = benchmark.Name
		marshalled["benchmarks."+tag+".description"] = benchmark.Description
		marshalled["benchmarks."+tag+".design_file"] = benchmark.DesignFile
		marshalled["benchmarks."+tag+".data_dir"] = benchmark.DataDir
		marshalled["benchmarks."+tag+".motor_designation"] = benchmark.MotorDesignation
		marshalled["benchmarks."+tag+".enabled"] = fmt.Sprintf("%v", benchmark.Enabled)
	}

	return marshalled
}

// Bytes returns the configuration as bytes
func (c *Config) Bytes() []byte {
	return []byte(fmt.Sprintf("%+v", c))
}

// Validate checks the configuration for errors and resolves relative paths.
// configFileDir is the directory containing the configuration file, used as a base for relative paths.
func (cfg *Config) Validate(configFileDir string) error {
	// Check for required fields
	if cfg.Setup.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if cfg.Setup.App.Version == "" {
		return fmt.Errorf("app.version is required")
	}
	if cfg.Setup.Logging.Level == "" {
		return fmt.Errorf("logging.level is required")
	}
	if cfg.Engine.External.OpenRocketVersion == "" {
		return fmt.Errorf("external.openrocket_version is required")
	}
	if cfg.Engine.Options.MotorDesignation == "" {
		return fmt.Errorf("options.motor_designation is required")
	}
	if cfg.Engine.Options.OpenRocketFile == "" {
		return fmt.Errorf("options.openrocket_file is required")
	}

	// Get the directory of the config file
	// Resolve OpenRocketFile path relative to config file's directory
	optionsFilePath := cfg.Engine.Options.OpenRocketFile
	if !filepath.IsAbs(optionsFilePath) {
		optionsFilePath = filepath.Join(configFileDir, optionsFilePath)
	}
	if _, err := os.Stat(optionsFilePath); os.IsNotExist(err) {
		return fmt.Errorf("options.openrocket_file path does not exist: %s (resolved from %s)", optionsFilePath, cfg.Engine.Options.OpenRocketFile)
	} else if err != nil {
		return fmt.Errorf("error checking options.openrocket_file path '%s': %w", optionsFilePath, err)
	}
	cfg.Engine.Options.OpenRocketFile = optionsFilePath // Update with resolved absolute path

	// Validate Launchrail
	if cfg.Engine.Options.Launchrail.Length <= 0 {
		return fmt.Errorf("options.launchrail.length must be greater than zero")
	}
	if cfg.Engine.Options.Launchrail.Angle < -90 || cfg.Engine.Options.Launchrail.Angle > 90 {
		return fmt.Errorf("options.launchrail.angle must be between -90 and 90 degrees")
	}
	if cfg.Engine.Options.Launchrail.Orientation < 0 || cfg.Engine.Options.Launchrail.Orientation > 360 {
		return fmt.Errorf("options.launchrail.orientation must be between 0 and 360 degrees")
	}

	// Launchsite
	if cfg.Engine.Options.Launchsite.Latitude < -90 || cfg.Engine.Options.Launchsite.Latitude > 90 {
		return fmt.Errorf("options.launchsite.latitude must be between -90 and 90")
	}
	if cfg.Engine.Options.Launchsite.Longitude < -180 || cfg.Engine.Options.Launchsite.Longitude > 180 {
		return fmt.Errorf("options.launchsite.longitude must be between -180 and 180")
	}
	if cfg.Engine.Options.Launchsite.Altitude < 0 {
		return fmt.Errorf("options.launchsite.altitude must be non-negative")
	}

	// Atmosphere
	isa := cfg.Engine.Options.Launchsite.Atmosphere.ISAConfiguration
	if isa.SpecificGasConstant <= 0 {
		return fmt.Errorf("options.launchsite.atmosphere.isa_configuration.specific_gas_constant is required")
	}
	if isa.GravitationalAccel <= 0 {
		return fmt.Errorf("options.launchsite.atmosphere.isa_configuration.gravitational_accel is required")
	}
	if isa.SeaLevelDensity <= 0 {
		return fmt.Errorf("options.launchsite.atmosphere.isa_configuration.sea_level_density is required")
	}
	if isa.SeaLevelTemperature < -273.15 {
		return fmt.Errorf("options.launchsite.atmosphere.isa_configuration.sea_level_temperature must be above absolute zero")
	}
	if isa.SeaLevelPressure <= 0 {
		return fmt.Errorf("options.launchsite.atmosphere.isa_configuration.sea_level_pressure is required")
	}
	if isa.RatioSpecificHeats <= 1 {
		return fmt.Errorf("options.launchsite.atmosphere.isa_configuration.ratio_specific_heats must be greater than 1")
	}
	if isa.TemperatureLapseRate == 0 {
		return fmt.Errorf("options.launchsite.atmosphere.isa_configuration.temperature_lapse_rate is required")
	}

	// Simulation
	if cfg.Engine.Simulation.Step <= 0 {
		return fmt.Errorf("simulation.step must be greater than zero")
	}
	if cfg.Engine.Simulation.MaxTime <= 0 {
		return fmt.Errorf("simulation.max_time must be greater than zero")
	}
	if cfg.Engine.Simulation.GroundTolerance < 0 {
		return fmt.Errorf("simulation.ground_tolerance must be non-negative")
	}

	// Plugins
	if len(cfg.Setup.Plugins.Paths) == 0 {
		return fmt.Errorf("plugins.paths must contain at least one valid path")
	}
	// Optionally validate each plugin path exists, similar to benchmark files
	for i, p := range cfg.Setup.Plugins.Paths {
		pluginPath := p
		if !filepath.IsAbs(pluginPath) {
			pluginPath = filepath.Join(configFileDir, pluginPath)
		}
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			return fmt.Errorf("plugins.paths[%d] path does not exist: %s (resolved from %s)", i, pluginPath, p)
		} else if err != nil {
			return fmt.Errorf("error checking plugins.paths[%d] path '%s': %w", i, pluginPath, err)
		}
	}

	// Server
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}

	// Benchmarks
	for tag, benchmark := range cfg.Benchmarks {
		if benchmark.Name == "" {
			return fmt.Errorf("benchmark '%s': benchmark.name is required", tag)
		}
		if benchmark.DesignFile == "" {
			return fmt.Errorf("benchmark '%s': benchmark.design_file is required", tag)
		}
		if benchmark.DataDir == "" {
			return fmt.Errorf("benchmark '%s': benchmark.data_dir is required", tag)
		}
		if benchmark.MotorDesignation == "" { // Added validation
			return fmt.Errorf("benchmark '%s': benchmark.motor_designation is required", tag)
		}

		// Resolve DesignFile path relative to config file's directory
		designFilePath := benchmark.DesignFile
		if !filepath.IsAbs(designFilePath) {
			designFilePath = filepath.Join(configFileDir, designFilePath)
		}
		if _, err := os.Stat(designFilePath); os.IsNotExist(err) {
			return fmt.Errorf("benchmark '%s' designFile path does not exist: %s (resolved from %s)", tag, designFilePath, benchmark.DesignFile)
		} else if err != nil {
			return fmt.Errorf("error checking benchmark '%s' designFile path '%s': %w", tag, designFilePath, err)
		}

		// Resolve DataDir path relative to config file's directory
		dataDirPath := benchmark.DataDir
		if !filepath.IsAbs(dataDirPath) {
			dataDirPath = filepath.Join(configFileDir, dataDirPath)
		}
		if stat, err := os.Stat(dataDirPath); os.IsNotExist(err) {
			return fmt.Errorf("benchmark '%s' dataDir path does not exist: %s (resolved from %s)", tag, dataDirPath, benchmark.DataDir)
		} else if err != nil {
			return fmt.Errorf("error checking benchmark '%s' dataDir path '%s': %w", tag, dataDirPath, err)
		} else if !stat.IsDir() {
			return fmt.Errorf("benchmark '%s' dataDir path is not a directory: %s (resolved from %s)", tag, dataDirPath, benchmark.DataDir)
		}
	}

	return nil
}
