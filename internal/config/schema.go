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
	Options struct {
		MotorDesignation string `mapstructure:"motor_designation"`
		OpenRocketFile   string `mapstructure:"openrocket_file"`
	} `mapstructure:"options"`
}

// Marshal to map structure for logging.
func (c *Config) String() map[string]string {
	marshalled := make(map[string]string)
	marshalled["app.name"] = c.App.Name
	marshalled["app.version"] = c.App.Version
	marshalled["logging.level"] = c.Logging.Level
	marshalled["options.motor_designation"] = c.Options.MotorDesignation
	marshalled["options.openrocket_file"] = c.Options.OpenRocketFile
	return marshalled
}
