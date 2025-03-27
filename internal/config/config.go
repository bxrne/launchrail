package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var (
	cfg *Config
)

// GetConfig returns the application configuration as a singleton
func GetConfig() (*Config, error) {

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %s", err)
	}

	return cfg, nil
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

	return nil
}
