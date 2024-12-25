package config

// Config represents the application configuration.
type Config struct {
	App struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
	} `mapstructure:"app"`
	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`
	Dev struct {
		MotorDesignation string `mapstructure:"motor_designation"`
	} `mapstructure:"dev"`
}

// Marshal to map structure for logging.
func (c *Config) String() map[string]string {
	marshalled := make(map[string]string)
	marshalled["app.name"] = c.App.Name
	marshalled["app.version"] = c.App.Version
	marshalled["logging.level"] = c.Logging.Level
	marshalled["dev.motor_designation"] = c.Dev.MotorDesignation
	return marshalled
}
