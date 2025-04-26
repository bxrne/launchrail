package main

import (
	"fmt"
	"math"
	"net/http"
	"os"
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
	var err error
	log := logger.GetLogger(cfg.Setup.Logging.Level)
	log.Info("Starting simulation run")

	// Initialize the simulation manager
	simManager := simulation.NewManager(cfg, log)
	// Defer closing the manager
	defer func() {
		if cerr := simManager.Close(); cerr != nil {
			log.Error("Failed to close simulation manager", "Error", cerr)
			// Don't overwrite the original error if there was one
			if err == nil {
				err = cerr
			}
		}
	}()

	if err = simManager.Initialize(); err != nil {
		log.Error("Failed to initialize simulation manager", "Error", err)
		return fmt.Errorf("failed to initialize simulation manager: %w", err)
	}

	// Only now create a new record for the simulation
	record, err := recordManager.CreateRecord()
	if err != nil {
		log.Error("Failed to create record", "Error", err)
		return fmt.Errorf("failed to create simulation record: %w", err)
	}
	// Defer closing the record only if creation succeeded
	defer func() {
		if cerr := record.Close(); cerr != nil {
			log.Error("Failed to close simulation record", "Error", cerr)
			if err == nil {
				err = cerr
			}
		}
	}()

	// Run the simulation
	if err = simManager.Run(); err != nil {
		log.Error("Simulation run failed", "Error", err)
		return fmt.Errorf("simulation run failed: %w", err)
	}

	log.Info("Simulation run completed successfully")
	return nil
}

// configFromCtx reads the request body and parses it into a config.Config and validates it
func configFromCtx(c *gin.Context, currentCfg *config.Config) (*config.Config, error) {
	// Extracting form values
	motorDesignation := c.PostForm("motor-designation")
	openRocketFile := c.PostForm("openrocket-file")
	launchrailLengthStr := c.PostForm("launchrail-length")
	launchrailAngleStr := c.PostForm("launchrail-angle")
	launchrailOrientationStr := c.PostForm("launchrail-orientation")
	latitudeStr := c.PostForm("latitude")
	longitudeStr := c.PostForm("longitude")
	altitudeStr := c.PostForm("altitude")
	openRocketVersion := c.PostForm("openrocket-version")
	simulationStepStr := c.PostForm("simulation-step")
	maxTimeStr := c.PostForm("max-time")
	groundToleranceStr := c.PostForm("ground-tolerance")
	specificGasConstantStr := c.PostForm("specific-gas-constant")
	gravitationalAccelStr := c.PostForm("gravitational-accel")
	seaLevelDensityStr := c.PostForm("sea-level-density")
	seaLevelTemperatureStr := c.PostForm("sea-level-temperature")
	seaLevelPressureStr := c.PostForm("sea-level-pressure")
	ratioSpecificHeatsStr := c.PostForm("ratio-specific-heats")
	temperatureLapseRateStr := c.PostForm("temperature-lapse-rate")
	pluginPaths := c.PostForm("plugin-paths")

	log := logger.GetLogger(currentCfg.Setup.Logging.Level)
	log.Debug("Received plugin-paths from form", "value", pluginPaths)

	// Helper for checking parse errors
	var firstParseErr error
	checkParse := func(err error) {
		if err != nil && firstParseErr == nil {
			firstParseErr = err
		}
	}

	// Parse numeric fields
	launchrailLength, err := parseFloat(launchrailLengthStr, "launchrail-length")
	checkParse(err)
	launchrailAngle, err := parseFloat(launchrailAngleStr, "launchrail-angle")
	checkParse(err)
	launchrailOrientation, err := parseFloat(launchrailOrientationStr, "launchrail-orientation")
	checkParse(err)
	latitude, err := parseFloat(latitudeStr, "latitude")
	checkParse(err)
	longitude, err := parseFloat(longitudeStr, "longitude")
	checkParse(err)
	altitude, err := parseFloat(altitudeStr, "altitude")
	checkParse(err)
	simulationStep, err := parseFloat(simulationStepStr, "simulation-step")
	checkParse(err)
	maxTime, err := parseFloat(maxTimeStr, "max-time")
	checkParse(err)
	groundTolerance, err := parseFloat(groundToleranceStr, "ground-tolerance")
	checkParse(err)
	specificGasConstant, err := parseFloat(specificGasConstantStr, "specific-gas-constant")
	checkParse(err)
	gravitationalAccel, err := parseFloat(gravitationalAccelStr, "gravitational-accel")
	checkParse(err)
	seaLevelDensity, err := parseFloat(seaLevelDensityStr, "sea-level-density")
	checkParse(err)
	seaLevelTemperature, err := parseFloat(seaLevelTemperatureStr, "sea-level-temperature")
	checkParse(err)
	seaLevelPressure, err := parseFloat(seaLevelPressureStr, "sea-level-pressure")
	checkParse(err)
	ratioSpecificHeats, err := parseFloat(ratioSpecificHeatsStr, "ratio-specific-heats")
	checkParse(err)
	temperatureLapseRate, err := parseFloat(temperatureLapseRateStr, "temperature-lapse-rate")
	checkParse(err)

	// Return the first parsing error encountered
	if firstParseErr != nil {
		log.Error("Failed to parse numeric form field", "error", firstParseErr)
		return nil, firstParseErr
	}

	// Validate required fields
	if motorDesignation == "" || openRocketFile == "" || openRocketVersion == "" {
		log.Error("Validation failed: A required field is empty", "motorDesignation", motorDesignation, "openRocketFile", openRocketFile, "openRocketVersion", openRocketVersion)
		return nil, fmt.Errorf("required string fields (motor, ork file/version) cannot be empty")
	}

	// Create plugin paths slice explicitly
	var parsedPluginPaths []string
	if pluginPaths != "" {
		parts := strings.Split(pluginPaths, ",")
		parsedPluginPaths = make([]string, 0, len(parts))
		for _, p := range parts {
			trimmedPath := strings.TrimSpace(p)
			if trimmedPath != "" {
				parsedPluginPaths = append(parsedPluginPaths, trimmedPath)
			}
		}
	}
	log.Debug("Parsed plugin paths explicitly", "paths", parsedPluginPaths)

	// Create the config.Config struct
	simConfig := config.Config{
		Setup: config.Setup{
			App:     currentCfg.Setup.App,
			Logging: currentCfg.Setup.Logging,
			Plugins: config.Plugins{
				Paths: parsedPluginPaths,
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
					Length:      launchrailLength,
					Angle:       launchrailAngle,
					Orientation: launchrailOrientation,
				},
				Launchsite: config.Launchsite{
					Latitude:  latitude,
					Longitude: longitude,
					Altitude:  altitude,
					Atmosphere: config.Atmosphere{
						ISAConfiguration: config.ISAConfiguration{
							SpecificGasConstant:  specificGasConstant,
							GravitationalAccel:   gravitationalAccel,
							SeaLevelDensity:      seaLevelDensity,
							SeaLevelTemperature:  seaLevelTemperature,
							SeaLevelPressure:     seaLevelPressure,
							RatioSpecificHeats:   ratioSpecificHeats,
							TemperatureLapseRate: temperatureLapseRate,
						},
					},
				},
			},
			Simulation: config.Simulation{
				Step:            simulationStep,
				MaxTime:         maxTime,
				GroundTolerance: groundTolerance,
			},
		},
	}

	// After parsing POST data into newCfg, ensure consistency by calling Manager.Initialize():
	m := simulation.NewManager(&simConfig, logger.GetLogger(currentCfg.Setup.Logging.Level))
	// Initialize the manager to set up stores & apply config consistently
	if err := m.Initialize(); err != nil {
		log.Error("Manager initialization failed within configFromCtx", "error", err)
		return nil, fmt.Errorf("failed to initialize simulation manager: %w", err)
	}
	log.Debug("Manager initialized successfully within configFromCtx")

	// Validate the configuration
	log.Debug("Attempting to validate simConfig")
	if err := simConfig.Validate(); err != nil {
		log.Error("simConfig.Validate() failed", "error", err)
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}
	log.Debug("simConfig.Validate() succeeded")

	return &simConfig, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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

	// Documentation & API spec routes
	docs := r.Group("/docs")
	{
		// Serve Swagger UI files
		docs.Static("/swagger", "./docs/swagger-ui")

		// Serve the API documentation page
		docs.GET("", func(c *gin.Context) {
			render(c, pages.API(cfg.Setup.App.Version))
		})

		// Serve OpenAPI spec
		docs.GET("/openapi", func(c *gin.Context) {
			specFile := "./docs/openapi.yaml"
			spec, err := os.ReadFile(specFile)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read OpenAPI spec"})
				return
			}
			// Replace version placeholder
			specStr := strings.ReplaceAll(string(spec), "${VERSION}", cfg.Setup.App.Version)
			c.Data(http.StatusOK, "application/yaml", []byte(specStr))
		})
	}

	// API endpoints group with version from config
	apiVersion := fmt.Sprintf("/api/v%s", strings.Split(cfg.Setup.App.Version, ".")[0])
	api := r.Group(apiVersion)
	{
		api.POST("/run", dataHandler.handleSimRun)
		api.GET("/data", dataHandler.ListRecordsAPI)
		api.GET("/explore/:hash", dataHandler.GetExplorerData)
		api.GET("/spec", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/swagger/spec")
		})
	}

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
				CurrentPage: 1,
				TotalPages:  calculateTotalPages(len(motionData), 15),
			},
		}

		pageStr := c.Query("page")
		page, _ := parseInt(pageStr, "page") // Ignore error, check value instead
		if page < 1 { // Default to page 1 if not specified, zero, negative, or parse error
			page = 1
		}

		// Update pagination data
		explorerData.Pagination.CurrentPage = page

		render(c, pages.Explorer(explorerData, cfg.Setup.App.Version))
	})

	r.GET("/explore/:hash/json", dataHandler.GetExplorerData)

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

	log.Info("Server started", "Port", cfg.Server.Port)
	portStr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err := r.Run(portStr); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

// handleSimRun handles API requests to start simulations (now a method of DataHandler)
func (h *DataHandler) handleSimRun(c *gin.Context) {
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

	// Use the existing record manager from the DataHandler instance (h.records)
	if err := runSim(simConfig, h.records); err != nil {
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
