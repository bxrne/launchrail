package config_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
)

// TEST: GIVEN an empty config WHEN Validate is called THEN returns an error
func TestConfig_Validate_Empty(t *testing.T) {
	cfg := &config.Config{}

	err := cfg.Validate()
	if err == nil {
		t.Errorf("Validate() should return an error for empty config")
	}
}

// TEST: GIVEN a config with missing app name WHEN Validate is called THEN returns an error
func TestConfig_Validate_MissingAppName(t *testing.T) {
	cfg := &config.Config{
		Setup: config.Setup{
			App: config.App{
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

	err := cfg.Validate()
	if err == nil {
		t.Errorf("Validate() should return an error when app name is missing")
	}
}

// TEST: GIVEN a config with invalid launchrail length WHEN Validate is called THEN returns an error
func TestConfig_Validate_InvalidLaunchrailLength(t *testing.T) {
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
					Length:      -1.0, // Invalid length
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

	err := cfg.Validate()
	if err == nil {
		t.Errorf("Validate() should return an error for invalid launchrail length")
	}
}

// TEST: GIVEN a config with valid parameters WHEN Validate is called THEN does not return an error
func TestConfig_Validate_Valid(t *testing.T) {
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

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() should not return an error for valid config: %v", err)
	}
}
