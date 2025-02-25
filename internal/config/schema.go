package config

import "fmt"

// App represents the application configuration.
type App struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	BaseDir string `mapstructure:"base_dir"`
}

// Logging represents the logging configuration.
type Logging struct {
	Level string `mapstructure:"level"`
}

// External represents the external configuration.
type External struct {
	OpenRocketVersion string `mapstructure:"openrocket_version"`
}

// Plugins represents runtime plugins to enrich the simulation
type Plugins struct {
	Paths []string `mapstructure:"paths"`
}

// Launchrail represents the launchrail configuration.
type Launchrail struct {
	Length      float64 `mapstructure:"length"`
	Angle       float64 `mapstructure:"angle"`
	Orientation float64 `mapstructure:"orientation"`
}

// Launchsite represents the launchsite configuration.
type Launchsite struct {
	Latitude   float64    `mapstructure:"latitude"`
	Longitude  float64    `mapstructure:"longitude"`
	Altitude   float64    `mapstructure:"altitude"`
	Atmosphere Atmosphere `mapstructure:"atmosphere"`
}

// Atmosphere represents the atmosphere configuration.
type Atmosphere struct {
	ISAConfiguration ISAConfiguration `mapstructure:"isa_configuration"`
}

// ISAConfiguration represents the ISA configuration.
type ISAConfiguration struct {
	SpecificGasConstant  float64 `mapstructure:"specific_gas_constant"`
	GravitationalAccel   float64 `mapstructure:"gravitational_accel"`
	SeaLevelDensity      float64 `mapstructure:"sea_level_density"`
	SeaLevelTemperature  float64 `mapstructure:"sea_level_temperature"`
	SeaLevelPressure     float64 `mapstructure:"sea_level_pressure"`
	RatioSpecificHeats   float64 `mapstructure:"ratio_specific_heats"`
	TemperatureLapseRate float64 `mapstructure:"temperature_lapse_rate"`
}

// Options represents the application options.
type Options struct {
	MotorDesignation string     `mapstructure:"motor_designation"`
	OpenRocketFile   string     `mapstructure:"openrocket_file"`
	Launchrail       Launchrail `mapstructure:"launchrail"`
	Launchsite       Launchsite `mapstructure:"launchsite"`
}

// Simulation represents the simulation configuration.
type Simulation struct {
	Step            float64 `mapstructure:"step"`
	MaxTime         float64 `mapstructure:"max_time"`
	GroundTolerance float64 `mapstructure:"ground_tolerance"` // Add ground tolerance
}

// Config represents the overall application configuration.
type Config struct {
	App        App        `mapstructure:"app"`
	Logging    Logging    `mapstructure:"logging"`
	External   External   `mapstructure:"external"`
	Options    Options    `mapstructure:"options"`
	Simulation Simulation `mapstructure:"simulation"`
	Plugins    Plugins    `mapstructure:"plugins"`
}

// String returns the configuration as a map of strings, useful for testing.
func (c *Config) String() map[string]string {
	marshalled := make(map[string]string)
	marshalled["app.name"] = c.App.Name
	marshalled["app.version"] = c.App.Version
	marshalled["logging.level"] = c.Logging.Level
	marshalled["app.base_dir"] = c.App.BaseDir
	marshalled["external.openrocket_version"] = c.External.OpenRocketVersion
	marshalled["options.motor_designation"] = c.Options.MotorDesignation
	marshalled["options.openrocket_file"] = c.Options.OpenRocketFile
	marshalled["options.launchrail.length"] = fmt.Sprintf("%.2f", c.Options.Launchrail.Length)
	marshalled["options.launchrail.angle"] = fmt.Sprintf("%.2f", c.Options.Launchrail.Angle)
	marshalled["options.launchrail.orientation"] = fmt.Sprintf("%.2f", c.Options.Launchrail.Orientation)
	marshalled["options.launchsite.latitude"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Latitude)
	marshalled["options.launchsite.longitude"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Longitude)
	marshalled["options.launchsite.altitude"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Altitude)
	marshalled["options.launchsite.atmosphere.isa_configuration.specific_gas_constant"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Atmosphere.ISAConfiguration.SpecificGasConstant)
	marshalled["options.launchsite.atmosphere.isa_configuration.gravitational_accel"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Atmosphere.ISAConfiguration.GravitationalAccel)
	marshalled["options.launchsite.atmosphere.isa_configuration.sea_level_density"] = fmt.Sprintf("%.3f", c.Options.Launchsite.Atmosphere.ISAConfiguration.SeaLevelDensity)
	marshalled["options.launchsite.atmosphere.isa_configuration.sea_level_temperature"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Atmosphere.ISAConfiguration.SeaLevelTemperature)
	marshalled["options.launchsite.atmosphere.isa_configuration.sea_level_pressure"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Atmosphere.ISAConfiguration.SeaLevelPressure)
	marshalled["options.launchsite.atmosphere.isa_configuration.ratio_specific_heats"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Atmosphere.ISAConfiguration.RatioSpecificHeats)
	marshalled["options.launchsite.atmosphere.isa_configuration.temperature_lapse_rate"] = fmt.Sprintf("%.2f", c.Options.Launchsite.Atmosphere.ISAConfiguration.TemperatureLapseRate)
	marshalled["simulation.step"] = fmt.Sprintf("%.2f", c.Simulation.Step)
	marshalled["simulation.max_time"] = fmt.Sprintf("%.2f", c.Simulation.MaxTime)
	marshalled["simulation.ground_tolerance"] = fmt.Sprintf("%.2f", c.Simulation.GroundTolerance)

	if len(c.Plugins.Paths) > 0 {
		marshalled["plugins.paths"] = c.Plugins.Paths[0]
	} else {
		marshalled["plugins.paths"] = ""
	}

	return marshalled
}

// Bytes returns the configuration as bytes
func (c *Config) Bytes() []byte {
	return []byte(fmt.Sprintf("%+v", c))
}
