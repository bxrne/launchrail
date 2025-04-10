package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/gin-gonic/gin"
)

// runSim starts the simulation with the given configuration
func runSim(cfg *config.Config, recordManager *storage.RecordManager) error {
	log := logger.GetLogger(cfg)

	// Create a new record for the simulation
	record, err := recordManager.CreateRecord()
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}
	defer record.Close()

	// Initialize the simulation manager
	simManager := simulation.NewManager(cfg, log)
	if err := simManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize simulation: %w", err)
	}

	// Run the simulation
	if err := simManager.Run(); err != nil {
		return fmt.Errorf("simulation failed: %w", err)
	}

	return nil
}

// configFromCtx reads the request body and parses it into a config.Config and validates it
func configFromCtx(c *gin.Context) (*config.Config, error) {
	// Extracting form values
	motorDesignation := c.PostForm("motor-designation")
	openRocketFile := c.PostForm("openrocket-file")
	launchrailLength := c.PostForm("launchrail-length")
	launchrailAngle := c.PostForm("launchrail-angle")
	launchrailOrientation := c.PostForm("launchrail-orientation")
	latitude := c.PostForm("latitude")
	longitude := c.PostForm("longitude")
	altitude := c.PostForm("altitude")
	openRocketVersion := c.PostForm("openrocket-version")
	simulationStep := c.PostForm("simulation-step")
	maxTime := c.PostForm("max-time")
	groundTolerance := c.PostForm("ground-tolerance")
	specificGasConstant := c.PostForm("specific-gas-constant")
	gravitationalAccel := c.PostForm("gravitational-accel")
	seaLevelDensity := c.PostForm("sea-level-density")
	seaLevelTemperature := c.PostForm("sea-level-temperature")
	seaLevelPressure := c.PostForm("sea-level-pressure")
	ratioSpecificHeats := c.PostForm("ratio-specific-heats")
	temperatureLapseRate := c.PostForm("temperature-lapse-rate")
	pluginPaths := c.PostForm("plugin-paths")

	// Validate required fields
	if motorDesignation == "" || openRocketFile == "" || launchrailLength == "" ||
		launchrailAngle == "" || launchrailOrientation == "" || latitude == "" ||
		longitude == "" || altitude == "" || openRocketVersion == "" ||
		simulationStep == "" || maxTime == "" || groundTolerance == "" ||
		specificGasConstant == "" || gravitationalAccel == "" || seaLevelDensity == "" ||
		seaLevelTemperature == "" || seaLevelPressure == "" || ratioSpecificHeats == "" ||
		temperatureLapseRate == "" || pluginPaths == "" {
		return nil, fmt.Errorf("all fields are required")
	}

	// Create the config.Config struct
	simConfig := config.Config{
		Setup: config.Setup{
			App: config.App{
				Name:    "Launchrail",
				Version: "1.0",
				BaseDir: "./",
			},
			Logging: config.Logging{
				Level: "info",
			},
			Plugins: config.Plugins{
				Paths: []string{pluginPaths},
			},
		},
		Server: config.Server{
			Port: 8080, // Set your desired port
		},
		Engine: config.Engine{
			External: config.External{
				OpenRocketVersion: openRocketVersion,
			},
			Options: config.Options{
				MotorDesignation: motorDesignation,
				OpenRocketFile:   openRocketFile,
				Launchrail: config.Launchrail{
					Length:      parseFloat(launchrailLength),
					Angle:       parseFloat(launchrailAngle),
					Orientation: parseFloat(launchrailOrientation),
				},
				Launchsite: config.Launchsite{
					Latitude:  parseFloat(latitude),
					Longitude: parseFloat(longitude),
					Altitude:  parseFloat(altitude),
					Atmosphere: config.Atmosphere{
						ISAConfiguration: config.ISAConfiguration{
							SpecificGasConstant:  parseFloat(specificGasConstant),
							GravitationalAccel:   parseFloat(gravitationalAccel),
							SeaLevelDensity:      parseFloat(seaLevelDensity),
							SeaLevelTemperature:  parseFloat(seaLevelTemperature),
							SeaLevelPressure:     parseFloat(seaLevelPressure),
							RatioSpecificHeats:   parseFloat(ratioSpecificHeats),
							TemperatureLapseRate: parseFloat(temperatureLapseRate),
						},
					},
				},
			},
			Simulation: config.Simulation{
				Step:            parseFloat(simulationStep),
				MaxTime:         parseFloat(maxTime),
				GroundTolerance: parseFloat(groundTolerance),
			},
		},
	}

	// Validate the configuration
	if err := simConfig.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &simConfig, nil
}

// Helper function to parse float values from strings
func parseFloat(value string) float64 {
	result, _ := strconv.ParseFloat(value, 64)
	return result
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*html")

	dataHandler, err := NewDataHandler(".launchrail")
	if err != nil {
		log.Fatal(err)
	}
	// Data routes
	r.GET("/data", dataHandler.ListRecords)
	r.GET("/data/:hash/:type", dataHandler.GetRecordData)
	// Landing page (pun intended)
	r.GET("/", func(c *gin.Context) {
		c.File("templates/index.html")
	})

	r.GET("/explore/:hash", func(c *gin.Context) {
		hash := c.Param("hash")
		record, err := dataHandler.records.GetRecord(hash)
		if err != nil {
			c.HTML(http.StatusNotFound, "partials/error.html", gin.H{
				"error": "Record not found",
			})
			return
		}

		motionData, _ := record.Motion.ReadAll()
		dynamicsData, _ := record.Dynamics.ReadAll()
		eventsData, _ := record.Events.ReadAll()

		c.HTML(http.StatusOK, "explorer.html", gin.H{
			"MotionHeaders":   motionData[0],
			"MotionData":      motionData[1:],
			"DynamicsHeaders": dynamicsData[0],
			"DynamicsData":    dynamicsData[1:],
			"EventsHeaders":   eventsData[0],
			"EventsData":      eventsData[1:],
		})
	})

	r.POST("/run", func(c *gin.Context) {
		simConfig, err := configFromCtx(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := runSim(simConfig, dataHandler.records); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"message": "Simulation started"})
	})

	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
