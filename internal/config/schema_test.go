package config_test

import (
	"reflect"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
)

// TEST: GIVEN a populated configuration WHEN String is called THEN a map of configuration values is returned
func TestConfigString(t *testing.T) {
	cfg := config.Config{
		App: config.App{
			Name:    "launchrail-test",
			Version: "0.0.0",
			BaseDir: "/tmp",
		},
		Logging: config.Logging{
			Level: "info",
		},
		External: config.External{
			OpenRocketVersion: "15.03",
		},
		Options: config.Options{
			MotorDesignation: "G80-7T",
			OpenRocketFile:   "test/fixtures/rocket.ork",
			Launchrail: config.Launchrail{
				Length:      0.00,
				Angle:       0.00,
				Orientation: 0.00,
			},
			Launchsite: config.Launchsite{
				Latitude:  0.00,
				Longitude: 0.00,
				Altitude:  0.00,
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						SpecificGasConstant:  287.05,
						GravitationalAccel:   9.81,
						SeaLevelDensity:      1.225,
						SeaLevelTemperature:  288.15,
						SeaLevelPressure:     101325.0,
						RatioSpecificHeats:   1.4,
						TemperatureLapseRate: -0.0065,
					},
				},
			},
		},
		Plugins: config.Plugins{
			Paths: []string{"fake_plugin.so"},
		},
		Simulation: config.Simulation{
			Step:    0.00,
			MaxTime: 0.00,
		},
	}

	expected := map[string]string{
		"app.name":                       "launchrail-test",
		"app.version":                    "0.0.0",
		"app.base_dir":                   "/tmp",
		"logging.level":                  "info",
		"external.openrocket_version":    "15.03",
		"options.motor_designation":      "G80-7T",
		"options.openrocket_file":        "test/fixtures/rocket.ork",
		"options.launchrail.length":      "0.00",
		"options.launchrail.angle":       "0.00",
		"options.launchrail.orientation": "0.00",
		"options.launchsite.latitude":    "0.00",
		"options.launchsite.longitude":   "0.00",
		"options.launchsite.altitude":    "0.00",
		"options.launchsite.atmosphere.isa_configuration.specific_gas_constant":  "287.05",
		"options.launchsite.atmosphere.isa_configuration.gravitational_accel":    "9.81",
		"options.launchsite.atmosphere.isa_configuration.sea_level_density":      "1.225",
		"options.launchsite.atmosphere.isa_configuration.sea_level_temperature":  "288.15",
		"options.launchsite.atmosphere.isa_configuration.sea_level_pressure":     "101325.00",
		"options.launchsite.atmosphere.isa_configuration.ratio_specific_heats":   "1.40",
		"options.launchsite.atmosphere.isa_configuration.temperature_lapse_rate": "-0.01",
		"simulation.step":             "0.00",
		"simulation.max_time":         "0.00",
		"simulation.ground_tolerance": "0.00",
		"plugins.paths":               "fake_plugin.so",
	}

	actual := cfg.String()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

// TEST: GIVEN a configuration WHEN Bytes is called THEN a byte representation of the configuration is returned
func TestConfigBytes(t *testing.T) {
	cfg := config.Config{
		App: config.App{
			Name:    "launchrail-test",
			Version: "0.0.0",
			BaseDir: "/tmp",
		},
		Logging: config.Logging{
			Level: "info",
		},
		External: config.External{
			OpenRocketVersion: "15.03",
		},
		Options: config.Options{
			MotorDesignation: "G80-7T",
			OpenRocketFile:   "test/fixtures/rocket.ork",
			Launchrail: config.Launchrail{
				Length:      0.00,
				Angle:       0.00,
				Orientation: 0.00,
			},
			Launchsite: config.Launchsite{
				Latitude:  0.00,
				Longitude: 0.00,
				Altitude:  0.00,
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						SpecificGasConstant:  287.05,
						GravitationalAccel:   9.81,
						SeaLevelDensity:      1.225,
						SeaLevelTemperature:  288.15,
						SeaLevelPressure:     101325.0,
						RatioSpecificHeats:   1.4,
						TemperatureLapseRate: -0.0065,
					},
				},
			},
		},
		Simulation: config.Simulation{
			Step:    0.00,
			MaxTime: 0.00,
		},
		Plugins: config.Plugins{
			Paths: []string{"fake_plugin.so"},
		},
	}

	expected := "&{App:{Name:launchrail-test Version:0.0.0 BaseDir:/tmp} Logging:{Level:info} External:{OpenRocketVersion:15.03} Options:{MotorDesignation:G80-7T OpenRocketFile:test/fixtures/rocket.ork Launchrail:{Length:0 Angle:0 Orientation:0} Launchsite:{Latitude:0 Longitude:0 Altitude:0 Atmosphere:{ISAConfiguration:{SpecificGasConstant:287.05 GravitationalAccel:9.81 SeaLevelDensity:1.225 SeaLevelTemperature:288.15 SeaLevelPressure:101325 RatioSpecificHeats:1.4 TemperatureLapseRate:-0.0065}}}} Simulation:{Step:0 MaxTime:0 GroundTolerance:0} Plugins:{Paths:[fake_plugin.so]}}"
	actual := string(cfg.Bytes())
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}
