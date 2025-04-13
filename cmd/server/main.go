package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

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

// Helper function to parse int values from strings
func parseInt(value string, defaultValue int) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return result
}

// Helper function to find the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to render templ components in Gin
func render(c *gin.Context, component templ.Component) {
	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		err_err := c.AbortWithError(http.StatusInternalServerError, err)
		if err_err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template"})
		}
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
	err = r.SetTrustedProxies(nil)
	if err != nil {
		log.Fatal("Failed to set trusted proxies", "Error", err)
	}

	dataHandler, err := NewDataHandler(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Serve static files
	r.Static("/static", "./static")

	// Serve Swagger UI
	r.Static("/docs", "./swagger-ui")

	// Ensure `/api/spec` is registered only once
	r.GET("/api/spec", func(c *gin.Context) {
		c.File("./swagger-ui/openapi.yaml") // Ensure the file exists at this path
	})

	// Data routes
	r.GET("/data", dataHandler.ListRecords)
	r.GET("/data/:hash/:type", dataHandler.GetRecordData)
	r.DELETE("/data/:hash", dataHandler.DeleteRecord)

	// Landing page
	r.GET("/", func(c *gin.Context) {
		render(c, pages.Index(cfg.Setup.App.Version))
	})

	r.GET("/explore/:hash", func(c *gin.Context) {
		hash := c.Param("hash")
		table := c.Query("table")
		if table == "" {
			table = "motion" // Default to motion table
		}

		record, err := dataHandler.records.GetRecord(hash)
		if err != nil {
			render(c, pages.ErrorPage("Record not found"))
			return
		}
		defer record.Close()

		// Load all headers first
		motionHeaders, motionData, err := record.Motion.ReadHeadersAndData()
		if err != nil {
			render(c, pages.ErrorPage("Failed to read motion data"))
			return
		}

		dynamicsHeaders, dynamicsData, err := record.Dynamics.ReadHeadersAndData()
		if err != nil {
			render(c, pages.ErrorPage("Failed to read dynamics data"))
			return
		}

		eventsHeaders, eventsData, err := record.Events.ReadHeadersAndData()
		if err != nil {
			render(c, pages.ErrorPage("Failed to read events data"))
			return
		}

		// Create explorer data structure
		explorerData := pages.ExplorerData{
			Hash:  hash,
			Table: table,
			Headers: pages.ExplorerHeaders{
				Motion:   motionHeaders,
				Dynamics: dynamicsHeaders,
				Events:   eventsHeaders,
			},
			Data: pages.ExplorerDataContent{
				Motion:   convertToFloat64(motionData),
				Dynamics: convertToFloat64(dynamicsData),
				Events:   eventsData,
			},
			Pagination: pages.Pagination{
				CurrentPage: parseInt(c.Query("page"), 1),
				TotalPages:  calculateTotalPages(len(motionData), 15),
			},
		}

		render(c, pages.Explorer(explorerData, cfg.Setup.App.Version))
	})

	r.GET("/explore/:hash/json", dataHandler.GetExplorerData)

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

	r.POST("/plot", func(c *gin.Context) {
		hash := c.PostForm("hash")
		source := c.PostForm("source")
		xAxis := c.PostForm("xAxis")
		yAxis := c.PostForm("yAxis")
		zAxis := c.PostForm("zAxis")

		record, err := dataHandler.records.GetRecord(hash)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found"})
			return
		}
		defer record.Close()

		// Get the correct headers/data
		var headers []string
		var rows [][]string
		switch source {
		case "motion":
			headers, rows, err = record.Motion.ReadHeadersAndData()
		case "dynamics":
			headers, rows, err = record.Dynamics.ReadHeadersAndData()
		case "events":
			headers, rows, err = record.Events.ReadHeadersAndData()
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read data"})
			return
		}

		// Convert rows to float64 array unless events
		var floatData [][]float64
		if source != "events" {
			floatData = make([][]float64, len(rows))
			for i, row := range rows {
				floatData[i] = make([]float64, len(row))
				for j, val := range row {
					floatData[i][j], _ = strconv.ParseFloat(val, 64)
				}
			}
		}

		// Find field indices
		xIndex, yIndex, zIndex := -1, -1, -1
		for i, h := range headers {
			if h == xAxis {
				xIndex = i
			}
			if h == yAxis {
				yIndex = i
			}
			if h == zAxis {
				zIndex = i
			}
		}

		if xIndex < 0 || yIndex < 0 || (zAxis != "" && zIndex < 0) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid axes"})
			return
		}

		// Build the response data as if we were going to feed Plotly
		// Return rows as strings if events, otherwise as floats
		var xData, yData, zData []interface{}
		if source == "events" {
			for _, row := range rows {
				xData = append(xData, row[xIndex])
				yData = append(yData, row[yIndex])
				if zIndex >= 0 {
					zData = append(zData, row[zIndex])
				}
			}
		} else {
			for _, row := range floatData {
				xData = append(xData, row[xIndex])
				yData = append(yData, row[yIndex])
				if zIndex >= 0 {
					zData = append(zData, row[zIndex])
				}
			}
		}

		plotLayout := map[string]interface{}{
			"title": fmt.Sprintf("%s vs %s%s", yAxis, xAxis, func() string {
				if zAxis != "" {
					return " vs " + zAxis
				}
				return ""
			}()),
			"xaxis": map[string]string{"title": xAxis},
			"yaxis": map[string]string{"title": yAxis},
		}

		plotData := []map[string]interface{}{
			{
				"x": xData,
				"y": yData,
				"type": func() string {
					if zAxis != "" {
						return "scatter3d"
					}
					return "scatter"
				}(),
				"mode": "markers",
			},
		}
		if zAxis != "" {
			plotData[0]["z"] = zData
		}

		c.JSON(http.StatusOK, gin.H{"plotData": plotData, "plotLayout": plotLayout})
	})

	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": cfg.Setup.App.Version,
			"apiPath": fmt.Sprintf("/api/v%s", strings.Split(cfg.Setup.App.Version, ".")[0]),
		})
	})

	// Group API routes under versioned prefix
	api := r.Group(fmt.Sprintf("/api/v%s", strings.Split(cfg.Setup.App.Version, ".")[0]))
	{
		api.POST("/run", handleSimRun)
		api.GET("/data", dataHandler.ListRecordsAPI)
		api.GET("/explore/:hash", dataHandler.GetExplorerData)
	}

	log.Info("Server started", "Port", cfg.Server.Port)
	portStr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err := r.Run(portStr); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

// handleSimRun handles API requests to start simulations
func handleSimRun(c *gin.Context) {
	cfg, err := config.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to load config: %v", err)})
		return
	}

	simConfig, err := configFromCtx(c, cfg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataHandler, err := NewDataHandler(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := runSim(simConfig, dataHandler.records); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Simulation started"})
}

func convertToFloat64(data [][]string) [][]float64 {
	result := make([][]float64, len(data))
	for i, row := range data {
		result[i] = make([]float64, len(row))
		for j, val := range row {
			result[i][j], _ = strconv.ParseFloat(val, 64)
		}
	}
	return result
}

func calculateTotalPages(total int, perPage int) int {
	return int(math.Ceil(float64(total) / float64(perPage)))
}
