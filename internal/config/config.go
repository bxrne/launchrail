package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Appearance Appearance

	Advanced Advanced
}

type Appearance struct {
	WindowWidth  float32 `toml:"window_width"`
	WindowHeight float32 `toml:"window_height"`
}

type Advanced struct {
	LogLevel int `toml:"log_level"`
}

func NewConfig() *Config {
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		panic(err)
	}
	return &config
}

func (c *Config) GetLogLevel() int {
	return c.Advanced.LogLevel
}

func (c *Config) GetWindowWidth() float32 {
	return c.Appearance.WindowWidth
}

func (c *Config) GetWindowHeight() float32 {
	return c.Appearance.WindowHeight
}
