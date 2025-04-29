package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	cfg *Config // Singleton instance of Config
)

// GetConfig reads configuration, validates it, resolves paths, and returns the singleton instance.
func GetConfig() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".") // Look for config in the current directory

	// Attempt to find and read the config file
	if err := v.ReadInConfig(); err != nil {
		// Config file MUST exist now if we rely on it solely for paths etc.
		return nil, fmt.Errorf("failed to read mandatory config file: %w", err)
	}

	// Unmarshal the config read from the file into the cfg struct
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// --- Resolve Simulation Output Directory ---
	// Precedence: Config File > Default (No flag consideration)
	outputDir := ""
	if cfg.Setup.App.SimulationOutputDir != "" {
		// Use config file value
		outputDir = cfg.Setup.App.SimulationOutputDir
		fmt.Printf("Using simulation output directory from config file: %s\n", outputDir) // Debug/Info log
	} else {
		// Use default only if config file value wasn't set
		outputDir = filepath.Join(cfg.Setup.App.BaseDir, "cli_run")
		fmt.Printf("Simulation output directory not specified in config, using default: %s\n", outputDir) // Debug/Info log
	}

	// Ensure the path is absolute
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for output directory '%s': %w", outputDir, err)
	}

	// Store the resolved path
	cfg.Setup.App.ResolvedSimulationOutputDir = absOutputDir
	fmt.Printf("Resolved absolute output directory: %s\n", cfg.Setup.App.ResolvedSimulationOutputDir) // Debug/Info log
    // --- End Resolve Simulation Output Directory ---

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

// App represents the application configuration.
type App struct {
	Name                      string `mapstructure:"name"`
	Version                   string `mapstructure:"version"`
	BaseDir                   string `mapstructure:"base_dir"`
	SimulationOutputDir       string `mapstructure:"simulation_output_dir,omitempty"` // Optional output dir from YAML
	ResolvedSimulationOutputDir string `mapstructure:"-"`                         // Not from YAML, resolved absolute path
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
	Name       string `mapstructure:"name" validate:"required"`
	DesignFile string `mapstructure:"design_file" validate:"required,file"`
	DataDir    string `mapstructure:"data_dir" validate:"required,dir"`
	Enabled    bool   `mapstructure:"enabled" validate:"boolean"`
}

// BenchmarkConfig holds global settings for the benchmark tool.
type BenchmarkConfig struct {
	// SimulationResultsDir specifies the directory containing the actual simulation output files.
	// This is the directory the benchmark tool reads from.
	SimulationResultsDir string `mapstructure:"simulation_results_dir"`
	// DefaultBenchmarkTag specifies a tag to run if none is provided externally (e.g., via env var in future).
	// If empty, all enabled benchmarks are run.
	DefaultBenchmarkTag string `mapstructure:"default_benchmark_tag"`
	// MarkdownOutputPath specifies the file path to write the benchmark summary in Markdown format.
	// If empty, no Markdown file is written.
	MarkdownOutputPath string `mapstructure:"markdown_output_path"`
}

// Config represents the overall application configuration.
type Config struct {
	Setup     Setup     `mapstructure:"setup"`
	Server    Server    `mapstructure:"server"`
	Engine    Engine    `mapstructure:"engine"`
	Benchmarks map[string]BenchmarkEntry `mapstructure:"benchmarks"`
	Benchmark BenchmarkConfig `mapstructure:"benchmark"` // New benchmark settings
}

// String returns the configuration as a map of strings, useful for testing.
func (c *Config) String() map[string]string {
	marshalled := make(map[string]string)

	// Setup Config
	marshalled["app.name"] = c.Setup.App.Name
	marshalled["app.version"] = c.Setup.App.Version
	marshalled["app.base_dir"] = c.Setup.App.BaseDir
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

	// Benchmark
	marshalled["benchmark.simulation_results_dir"] = c.Benchmark.SimulationResultsDir
	marshalled["benchmark.default_benchmark_tag"] = c.Benchmark.DefaultBenchmarkTag
	marshalled["benchmark.markdown_output_path"] = c.Benchmark.MarkdownOutputPath

	return marshalled
}

// Bytes returns the configuration as bytes
func (c *Config) Bytes() []byte {
	return []byte(fmt.Sprintf("%+v", c))
}

// Validate checks the config for missing or invalid fields.
func (cfg *Config) Validate() error {
	// App Config
	if cfg.Setup.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if cfg.Setup.App.Version == "" {
		return fmt.Errorf("app.version is required")
	}
	if cfg.Setup.App.BaseDir == "" {
		return fmt.Errorf("app.base_dir is required")
	}

	// Logging
	if cfg.Setup.Logging.Level == "" {
		return fmt.Errorf("logging.level is required")
	}

	// External
	if cfg.Engine.External.OpenRocketVersion == "" {
		return fmt.Errorf("external.openrocket_version is required")
	}

	// Options
	if cfg.Engine.Options.MotorDesignation == "" {
		return fmt.Errorf("options.motor_designation is required")
	}
	if cfg.Engine.Options.OpenRocketFile == "" {
		return fmt.Errorf("options.openrocket_file is required")
	}
	if _, err := os.Stat(cfg.Engine.Options.OpenRocketFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("options.openrocket_file is invalid: %s", err)
	}

	// Launchrail
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
		// Removed os.Stat check that incorrectly used BaseDir here
		if benchmark.DataDir == "" {
			return fmt.Errorf("benchmark '%s': benchmark.data_dir is required", tag)
		}

		// Check if DesignFile exists (relative to project root or absolute)
		// Use the path directly as specified in config.yaml
		if _, err := os.Stat(benchmark.DesignFile); os.IsNotExist(err) {
			return fmt.Errorf("benchmark '%s' designFile path does not exist: %s", tag, benchmark.DesignFile)
		} else if err != nil {
			return fmt.Errorf("error checking benchmark '%s' designFile path '%s': %w", tag, benchmark.DesignFile, err)
		}

		// Check if DataDir exists (relative to project root or absolute)
		// Use the path directly as specified in config.yaml
		if stat, err := os.Stat(benchmark.DataDir); os.IsNotExist(err) {
			return fmt.Errorf("benchmark '%s' dataDir path does not exist: %s", tag, benchmark.DataDir)
		} else if err != nil {
			return fmt.Errorf("error checking benchmark '%s' dataDir path '%s': %w", tag, benchmark.DataDir, err)
		} else if !stat.IsDir() {
			return fmt.Errorf("benchmark '%s' dataDir path is not a directory: %s", tag, benchmark.DataDir)
		}
	}

	// Benchmark
	if cfg.Benchmark.SimulationResultsDir == "" {
		return fmt.Errorf("benchmark.simulation_results_dir is required")
	}
	if cfg.Benchmark.DefaultBenchmarkTag == "" {
		return fmt.Errorf("benchmark.default_benchmark_tag is required")
	}
	if cfg.Benchmark.MarkdownOutputPath == "" {
		return fmt.Errorf("benchmark.markdown_output_path is required")
	}

	return nil
}
