package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// GetConfig returns the application configuration
func GetConfig() (*Config, error) {
	var cfg *Config
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		return nil, errors.New("failed to read config file")
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errors.New("failed to unmarshal config")
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
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
