package config_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
)

// TEST: GIVEN a valid config WHEN String is called THEN returns a map of strings
func TestConfig_String(t *testing.T) {
	cfg := &config.Config{
		Setup: config.Setup{
			App: config.App{
				Name:    "TestApp",
				Version: "1.0.0",
				BaseDir: "/tmp",
			},
			Logging: config.Logging{
				Level: "debug",
			},
			Plugins: config.Plugins{
				Paths: []string{"/opt/plugins"},
			},
		},
		Server: config.Server{
			Port: 8080,
		},
		Engine: config.Engine{
			External: config.External{
				OpenRocketVersion: "23.0",
			},
			Options: config.Options{
				MotorDesignation: "A8-3",
				OpenRocketFile:   "/tmp/rocket.ork",
				Launchrail: config.Launchrail{
					Length:      1.2,
					Angle:       5.0,
					Orientation: 90.0,
				},
				Launchsite: config.Launchsite{
					Latitude:  34.0522,
					Longitude: -118.2437,
					Altitude:  100.0,
					Atmosphere: config.Atmosphere{
						ISAConfiguration: config.ISAConfiguration{
							SpecificGasConstant:  287.058,
							GravitationalAccel:   9.807,
							SeaLevelDensity:      1.225,
							SeaLevelTemperature:  15.0,
							SeaLevelPressure:     101325.0,
							RatioSpecificHeats:   1.40,
							TemperatureLapseRate: 0.0065,
						},
					},
				},
			},
			Simulation: config.Simulation{
				Step:            0.01,
				MaxTime:         10.0,
				GroundTolerance: 0.1,
			},
		},
	}

	expected := map[string]string{
		"app.name":                       "TestApp",
		"app.version":                    "1.0.0",
		"app.base_dir":                   "/tmp",
		"logging.level":                  "debug",
		"external.openrocket_version":    "23.0",
		"options.motor_designation":      "A8-3",
		"options.openrocket_file":        "/tmp/rocket.ork",
		"options.launchrail.length":      "1.20",
		"options.launchrail.angle":       "5.00",
		"options.launchrail.orientation": "90.00",
		"options.launchsite.latitude":    "34.0522",
		"options.launchsite.longitude":   "-118.2437",
		"options.launchsite.altitude":    "100.00",
		"options.launchsite.atmosphere.isa_configuration.specific_gas_constant":  "287.058",
		"options.launchsite.atmosphere.isa_configuration.gravitational_accel":    "9.807",
		"options.launchsite.atmosphere.isa_configuration.sea_level_density":      "1.225",
		"options.launchsite.atmosphere.isa_configuration.sea_level_temperature":  "15.00",
		"options.launchsite.atmosphere.isa_configuration.sea_level_pressure":     "101325.00",
		"options.launchsite.atmosphere.isa_configuration.ratio_specific_heats":   "1.40",
		"options.launchsite.atmosphere.isa_configuration.temperature_lapse_rate": "0.01",
		"simulation.step":             "0.0100",
		"simulation.max_time":         "10.00",
		"simulation.ground_tolerance": "0.10",
		"plugins.paths":               "[/opt/plugins]",
		"server.port":                 "8080",
	}

	actual := cfg.String()

	if actual["app.name"] != expected["app.name"] {
		t.Errorf("String() = %v, want %v", actual, expected)
	}
}

// TEST: GIVEN a valid config WHEN Bytes is called THEN returns a byte array
func TestConfig_Bytes(t *testing.T) {
	cfg := &config.Config{
		Setup: config.Setup{
			App: config.App{
				Name:    "TestApp",
				Version: "1.0.0",
				BaseDir: "/tmp",
			},
			Logging: config.Logging{
				Level: "debug",
			},
			Plugins: config.Plugins{
				Paths: []string{"/opt/plugins"},
			},
		},
		Server: config.Server{
			Port: 8080,
		},
		Engine: config.Engine{
			External: config.External{
				OpenRocketVersion: "23.0",
			},
			Options: config.Options{
				MotorDesignation: "A8-3",
				OpenRocketFile:   "/tmp/rocket.ork",
				Launchrail: config.Launchrail{
					Length:      1.2,
					Angle:       5.0,
					Orientation: 90.0,
				},
				Launchsite: config.Launchsite{
					Latitude:  34.0522,
					Longitude: -118.2437,
					Altitude:  100.0,
					Atmosphere: config.Atmosphere{
						ISAConfiguration: config.ISAConfiguration{
							SpecificGasConstant:  287.058,
							GravitationalAccel:   9.807,
							SeaLevelDensity:      1.225,
							SeaLevelTemperature:  15.0,
							SeaLevelPressure:     101325.0,
							RatioSpecificHeats:   1.40,
							TemperatureLapseRate: 0.0065,
						},
					},
				},
			},
			Simulation: config.Simulation{
				Step:            0.01,
				MaxTime:         10.0,
				GroundTolerance: 0.1,
			},
		},
	}

	bytes := cfg.Bytes()
	if len(bytes) == 0 {
		t.Errorf("Bytes() returned an empty byte array")
	}
}
