package config

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	once sync.Once
	cfg  *Config
)

// GetConfig returns the application configuration as a singleton
func GetConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		cfg = nil
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		cfg = nil
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	if err := cfg.Validate(); err != nil {
		cfg = nil
		return nil, fmt.Errorf("failed to validate config: %s", err)
	}

	if cfg == nil {
		return nil, errors.New("failed to load configuration")
	}

	return cfg, nil
}

// Reset resets the configuration singleton, useful for testing
func Reset() {
	cfg = nil
}

// Validate checks the config to error on empty field
func (cfg *Config) Validate() error {
	if cfg.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	if cfg.App.Version == "" {
		return fmt.Errorf("app.version is required")
	}

	if cfg.Logging.Level == "" {
		return fmt.Errorf("logging.level is required")
	}

	if cfg.External.OpenRocketVersion == "" {
		return fmt.Errorf("external.openrocket_version is required")
	}

	if cfg.Options.MotorDesignation == "" {
		return fmt.Errorf("options.motor_designation is required")
	}

	if cfg.Options.OpenRocketFile == "" {
		return fmt.Errorf("options.openrocket_file is required")
	}

	if _, err := os.Stat(cfg.Options.OpenRocketFile); err != nil {
		return fmt.Errorf("options.openrocket_file is invalid: %s", err)
	}

	if cfg.Options.Launchrail.Length == 0 {
		return fmt.Errorf("options.launchrail.length is required")
	}

	if cfg.Options.Launchrail.Angle == 0 {
		return fmt.Errorf("options.launchrail.angle is required")
	}

	if cfg.Options.Launchrail.Orientation == 0 {
		return fmt.Errorf("options.launchrail.orientation is required")
	}

	if cfg.Options.Launchsite.Latitude == 0 {
		return fmt.Errorf("options.launchsite.latitude is required")
	}

	if cfg.Options.Launchsite.Longitude == 0 {
		return fmt.Errorf("options.launchsite.longitude is required")
	}

	if cfg.Options.Launchsite.Altitude == 0 {
		return fmt.Errorf("options.launchsite.altitude is required")
	}

	return nil
}
