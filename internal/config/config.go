package config

import (
	"fmt"
	"sync"

	"github.com/bxrne/launchrail/internal/logger"
	"github.com/spf13/viper"
)

var (
	once     sync.Once
	instance *Config
)

// GetConfig returns the singleton instance of the configuration.
func GetConfig() *Config {
	log := logger.GetLogger()

	once.Do(func() {
		v := viper.New()
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")

		// v.AutomaticEnv()

		if err := v.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Msg("Failed to read configuration")
		}

		if err := v.Unmarshal(&instance); err != nil {
			log.Fatal().Err(err).Msg("Failed to unmarshal configuration")
		}

		if err := validateConfig(instance); err != nil {
			log.Fatal().Err(err).Msg("Failed to validate configuration")
		}
	})
	return instance
}

// validateConfig validates the configuration struct.
func validateConfig(cfg *Config) error {
	log := logger.GetLogger()

	if cfg.App.Name == "" {
		err := fmt.Errorf("app.name is required")
		log.Error().Err(err).Msg("Validation error")
		return err
	}

	if cfg.App.Version == "" {
		err := fmt.Errorf("app.version is required")
		log.Error().Err(err).Msg("Validation error")
		return err
	}

	if cfg.Logging.Level == "" {
		err := fmt.Errorf("logging.level is required")
		log.Error().Err(err).Msg("Validation error")
		return err
	}

	return nil
}

// ResetConfig resets the singleton instance for testing purposes.
func ResetConfig() {
	once = sync.Once{}
	instance = nil
}
