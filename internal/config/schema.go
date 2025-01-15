package config

import "fmt"

// App represents the application configuration.
type App struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// Logging represents the logging configuration.
type Logging struct {
	Level string `mapstructure:"level"`
}

// External represents the external configuration.
type External struct {
	OpenRocketVersion string `mapstructure:"openrocket_version"`
}

// Launchrail represents the launchrail configuration.
type Launchrail struct {
	Length      float64 `mapstructure:"length"`
	Angle       float64 `mapstructure:"angle"`
	Orientation float64 `mapstructure:"orientation"`
}

// Launchsite represents the launchsite configuration.
type Launchsite struct {
	Latitude  float64 `mapstructure:"latitude"`
	Longitude float64 `mapstructure:"longitude"`
	Altitude  float64 `mapstructure:"altitude"`
}

// Options represents the application options.
type Options struct {
	MotorDesignation string     `mapstructure:"motor_designation"`
	OpenRocketFile   string     `mapstructure:"openrocket_file"`
	Launchrail       Launchrail `mapstructure:"launchrail"`
	Launchsite       Launchsite `mapstructure:"launchsite"`
}

// Config represents the overall application configuration.
type Config struct {
	App      App      `mapstructure:"app"`
	Logging  Logging  `mapstructure:"logging"`
	External External `mapstructure:"external"`
	Options  Options  `mapstructure:"options"`
}

// String returns the configuration as a map of strings, useful for testing.
func (c *Config) String() map[string]string {
	marshalled := make(map[string]string)
	marshalled["app.name"] = c.App.Name
	marshalled["app.version"] = c.App.Version
	marshalled["logging.level"] = c.Logging.Level
	marshalled["external.openrocket_version"] = c.External.OpenRocketVersion
	marshalled["options.motor_designation"] = c.Options.MotorDesignation
	marshalled["options.openrocket_file"] = c.Options.OpenRocketFile
	marshalled["options.launchrail.length"] = fmt.Sprintf("%.2f", c.Options.Launchrail.Length)
	marshalled["options.launchrail.angle"] = fmt.Sprintf("%.2f", c.Options.Launchrail.Angle)
	marshalled["options.launchrail.orientation"] = fmt.Sprintf("%.2f", c.Options.Launchrail.Orientation)
	marshalled["options.launchsite.latitude"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Latitude)
	marshalled["options.launchsite.longitude"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Longitude)
	marshalled["options.launchsite.altitude"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Altitude)
	return marshalled
}
