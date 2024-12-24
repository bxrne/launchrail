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
