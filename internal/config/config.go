package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

// Config represents the application configuration structure.
// INFO: All fields are mapped using mapstructure tags for Viper compatibility
type Config struct {
	App struct {
		Version string `mapstructure:"version"`
		License string `mapstructure:"license"`
		Repo    string `mapstructure:"repo"`
	} `mapstructure:"app"`
	Logs struct {
		File string `mapstructure:"file"`
	} `mapstructure:"logs"`
	Engine struct {
		TimeStepNS int `mapstructure:"timestep_ns"`
	} `mapstructure:"engine"`
}

// Global singleton instances for configuration management
// INFO: These variables are package-level to maintain singleton pattern
var (
	once     sync.Once
	instance *Config
	err      error
)

// LoadConfig loads the configuration from the specified file path and returns a singleton instance.
// It uses Viper for configuration management and ensures thread-safe initialization.
//
// INFO: This is the primary method for obtaining configuration throughout the application
//
// Parameters:
//   - configPath: Path to the configuration file
//
// Returns:
//   - *Config: Populated configuration instance
//   - error: Any error encountered during loading or unmarshaling
//
// WARN: Subsequent calls with different configPath values will not reload the config
func LoadConfig(configPath string) (*Config, error) {
	once.Do(func() {
		viper.SetConfigFile(configPath)

		if err = viper.ReadInConfig(); err != nil {
			err = fmt.Errorf("error reading config file: %w", err)
			return
		}

		instance = &Config{}
		if err = viper.Unmarshal(instance); err != nil {
			err = fmt.Errorf("unable to decode config into struct: %w", err)
		}
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}
