package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/templates/pages"
	"github.com/gin-gonic/gin"
)

// runSim starts the simulation with the given configuration
func runSim(cfg *config.Config, recordManager *storage.RecordManager) error {
	log := logger.GetLogger(cfg.Setup.Logging.Level)

	// Create a new record for the simulation
	record, err := recordManager.CreateRecord()
	if err != nil {
		log.Fatal("Failed to create record", "Error", err)
	}
	defer record.Close()

	// Initialize the simulation manager
	simManager := simulation.NewManager(cfg, log)
	defer simManager.Close()

	if err := simManager.Initialize(); err != nil {
		log.Fatal("Failed to initialize simulation", "Error", err)
	}

	// Run the simulation
	if err := simManager.Run(); err != nil {
		log.Fatal("Simulation failed", "Error", err)
	}

	return nil
}

// configFromCtx reads the request body and parses it into a config.Config and validates it
func configFromCtx(c *gin.Context, currentCfg *config.Config) (*config.Config, error) {
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
			App:     currentCfg.Setup.App,
			Logging: currentCfg.Setup.Logging,
			Plugins: config.Plugins{
				Paths: []string{pluginPaths},
			},
		},
		Server: currentCfg.Server,
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

// Helper function to render templ components in Gin
func render(c *gin.Context, component templ.Component) {
	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}
	// Initialize GetLogger
	log := logger.GetLogger(cfg.Setup.Logging.Level)
	log.Info("Config loaded", "Name", cfg.Setup.App.Name, "Version", cfg.Setup.App.Version, "Message", "Starting server")

	r := gin.Default()
	r.SetTrustedProxies(nil)

	dataHandler, err := NewDataHandler(".launchrail")
	if err != nil {
		log.Fatal(err.Error())
	}

	// Serve static files
	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")
	r.StaticFile("/robots.txt", "./static/robots.txt")
	r.StaticFile("/manifest.json", "./static/manifest.json")

	// Data routes
	r.GET("/data", dataHandler.ListRecords)
	r.GET("/data/:hash/:type", dataHandler.GetRecordData)

	// Landing page
	r.GET("/", func(c *gin.Context) {
		render(c, pages.Index())
	})

	r.GET("/explore/:hash", func(c *gin.Context) {
		hash := c.Param("hash")
		record, err := dataHandler.records.GetRecord(hash)
		if err != nil {
			log.Error("Failed to get record", "Error", err)
			render(c, pages.ErrorPage("Record not found"))
			return
		}
		defer record.Close()

		// Ensure storage objects are not nil
		if record.Motion == nil || record.Events == nil || record.Dynamics == nil {
			render(c, pages.ErrorPage("Record storage is not properly initialized"))
			return
		}

		// Read headers and data from storage
		motionHeaders, motionData, err := record.Motion.ReadHeadersAndData()
		if err != nil {
			render(c, pages.ErrorPage("Failed to read motion data: "+err.Error()))
			return
		}

		dynamicsHeaders, dynamicsData, err := record.Dynamics.ReadHeadersAndData()
		if err != nil {
			render(c, pages.ErrorPage("Failed to read dynamics data: "+err.Error()))
			return
		}

		eventsHeaders, eventsData, err := record.Events.ReadHeadersAndData()
		if err != nil {
			render(c, pages.ErrorPage("Failed to read events data: "+err.Error()))
			return
		}

		// Render the explorer page with templ
		data := pages.ExplorerData{
			MotionHeaders:   motionHeaders,
			MotionData:      motionData,
			DynamicsHeaders: dynamicsHeaders,
			DynamicsData:    dynamicsData,
			EventsHeaders:   eventsHeaders,
			EventsData:      eventsData,
		}

		render(c, pages.Explorer(data))
	})

	r.POST("/run", func(c *gin.Context) {
		simConfig, err := configFromCtx(c, cfg)
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

	portStr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err := r.Run(portStr); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}

	log.Info("Server started", "Port", portStr)
}
