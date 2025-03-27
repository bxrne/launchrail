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

// Engine represents the engine configuration (simulation specific).
type Engine struct {
	External   External   `mapstructure:"external"`
	Options    Options    `mapstructure:"options"`
	Simulation Simulation `mapstructure:"simulation"`
}

// Setup represents the setup configuration.
type Setup struct {
	App     App     `mapstructure:"app"`
	Logging Logging `mapstructure:"logging"`
	Plugins Plugins `mapstructure:"plugins"`
}

// Server represents the server configuration.
type Server struct {
	Port int `mapstructure:"port"`
}

// Config represents the overall application configuration.
type Config struct {
	Setup  Setup  `mapstructure:"setup"`
	Server Server `mapstructure:"server"`
	Engine Engine `mapstructure:"engine"`
}

// String returns the configuration as a map of strings, useful for testing.
func (c *Config) String() map[string]string {
	marshalled := make(map[string]string)

	// Setup Config
	marshalled["app.name"] = c.Setup.App.Name
	marshalled["app.version"] = c.Setup.App.Version
	marshalled["app.base_dir"] = c.Setup.App.BaseDir
	marshalled["logging.level"] = c.Setup.Logging.Level

	// Engine -> External
	marshalled["external.openrocket_version"] = c.Engine.External.OpenRocketVersion

	// Engine -> Options
	marshalled["options.motor_designation"] = c.Engine.Options.MotorDesignation
	marshalled["options.openrocket_file"] = c.Engine.Options.OpenRocketFile

	// Launchrail
	marshalled["options.launchrail.length"] = fmt.Sprintf("%.2f", c.Engine.Options.Launchrail.Length)
	marshalled["options.launchrail.angle"] = fmt.Sprintf("%.2f", c.Engine.Options.Launchrail.Angle)
	marshalled["options.launchrail.orientation"] = fmt.Sprintf("%.2f", c.Engine.Options.Launchrail.Orientation)

	// Launchsite
	marshalled["options.launchsite.latitude"] = fmt.Sprintf("%.4f", c.Engine.Options.Launchsite.Latitude)
	marshalled["options.launchsite.longitude"] = fmt.Sprintf("%.4f", c.Engine.Options.Launchsite.Longitude)
	marshalled["options.launchsite.altitude"] = fmt.Sprintf("%.2f", c.Engine.Options.Launchsite.Altitude)

	// Atmosphere
	isa := c.Engine.Options.Launchsite.Atmosphere.ISAConfiguration
	marshalled["options.launchsite.atmosphere.isa_configuration.specific_gas_constant"] = fmt.Sprintf("%.3f", isa.SpecificGasConstant)
	marshalled["options.launchsite.atmosphere.isa_configuration.gravitational_accel"] = fmt.Sprintf("%.3f", isa.GravitationalAccel)
	marshalled["options.launchsite.atmosphere.isa_configuration.sea_level_density"] = fmt.Sprintf("%.3f", isa.SeaLevelDensity)
	marshalled["options.launchsite.atmosphere.isa_configuration.sea_level_temperature"] = fmt.Sprintf("%.2f", isa.SeaLevelTemperature)
	marshalled["options.launchsite.atmosphere.isa_configuration.sea_level_pressure"] = fmt.Sprintf("%.2f", isa.SeaLevelPressure)
	marshalled["options.launchsite.atmosphere.isa_configuration.ratio_specific_heats"] = fmt.Sprintf("%.2f", isa.RatioSpecificHeats)
	marshalled["options.launchsite.atmosphere.isa_configuration.temperature_lapse_rate"] = fmt.Sprintf("%.2f", isa.TemperatureLapseRate)

	// Simulation
	marshalled["simulation.step"] = fmt.Sprintf("%.4f", c.Engine.Simulation.Step)
	marshalled["simulation.max_time"] = fmt.Sprintf("%.2f", c.Engine.Simulation.MaxTime)
	marshalled["simulation.ground_tolerance"] = fmt.Sprintf("%.2f", c.Engine.Simulation.GroundTolerance)

	// Plugins - Store full list as comma-separated values
	if len(c.Setup.Plugins.Paths) > 0 {
		marshalled["plugins.paths"] = fmt.Sprintf("%v", c.Setup.Plugins.Paths)
	} else {
		marshalled["plugins.paths"] = ""
	}

	// Server Port
	marshalled["server.port"] = fmt.Sprintf("%d", c.Server.Port)

	return marshalled
}

// Bytes returns the configuration as bytes
func (c *Config) Bytes() []byte {
	return []byte(fmt.Sprintf("%+v", c))
}
