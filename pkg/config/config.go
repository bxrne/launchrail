package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Version string `mapstructure:"version"`
		License string `mapstructure:"license"`
		Repo    string `mapstructure:"repo"`
	} `mapstructure:"app"`
	Logs struct {
		File string `mapstructure:"file"`
	} `mapstructure:"logs"`
}

var (
	once     sync.Once
	instance *Config
	err      error
)

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
