package config

import (
	"fmt"
	"sync"

	"github.com/bxrne/launchrail/internal/logger"
	"github.com/spf13/viper"
)

// INFO: Singleton configuration instance.
var (
	once     sync.Once
	instance *Config
)

// GetConfig returns the singleton instance of the configuration.
func GetConfig(name string) *Config {
	log := logger.GetLogger()
	once.Do(func() {
		v := viper.New()
		v.SetConfigName(name)
		v.SetConfigType("yaml")
		v.AddConfigPath(".")

		if err := v.ReadInConfig(); err != nil {
			log.Fatal("Failed to read configuration", "error", err)
		}

		if err := v.Unmarshal(&instance); err != nil {
			log.Fatal("Failed to unmarshal configuration", "error", err)
		}

		if err := validateConfig(instance); err != nil {
			log.Fatal("Failed to validate configuration", "error", err)
		}

		log.Info("Configuration loaded", "config", instance.String())
	})

	return instance
}

// validateConfig validates the configuration struct.
func validateConfig(cfg *Config) error {
	log := logger.GetLogger()

	if cfg.App.Name == "" {
		err := fmt.Errorf("app.name is required")
		log.Fatal("Failed to validate configuration", "error", err)
		return err
	}

	if cfg.App.Version == "" {
		err := fmt.Errorf("app.version is required")
		log.Fatal("Failed to validate configuration", "error", err)
		return err
	}

	if cfg.Logging.Level == "" {
		err := fmt.Errorf("logging.level is required")
		log.Fatal("Failed to validate configuration", "error", err)
		return err
	}

	if cfg.Options.MotorDesignation == "" {
		err := fmt.Errorf("options.motor_designation is required")
		log.Fatal("Failed to validate configuration", "error", err)
		return err
	}

	return nil
}

// WARNING: ResetConfig resets the singleton instance for testing purposes.
func ResetConfig() {
	once = sync.Once{}
	instance = nil
	log := logger.GetLogger()
	log.Info("Configuration reset")
}
