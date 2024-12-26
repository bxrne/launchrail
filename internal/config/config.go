package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// GetConfig returns the singleton instance of the configuration.
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

	if cfg.Options.MotorDesignation == "" {
		return fmt.Errorf("options.motor_designation is required")
	}

	if cfg.Options.OpenRocketFile == "" {
		return fmt.Errorf("options.openrocket_file is required")
	}

	if _, err := os.Stat(cfg.Options.OpenRocketFile); err != nil {
		return fmt.Errorf("options.openrocket_file is invalid: %s", err)
	}

	return nil
}
