package main

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/bxrne/launchrail/internal/config"
	logger "github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/plot_transformer"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/templates/pages"
	"github.com/gin-gonic/gin"
	"github.com/zerodha/logf"
)

// runSim takes the global config and a record manager and runs a simulation
func runSim(cfg *config.Config, recordManager *storage.RecordManager, log *logf.Logger) error {
	// Use the old logger for the simulation package
	oldLog := log
	oldLog.Info("Starting simulation run")

	// Create simulation manager using the old logger
	simManager := simulation.NewManager(cfg, *oldLog)

	// Create a record for this simulation run *before* initializing the manager
	record, err := recordManager.CreateRecord()
	if err != nil {
		oldLog.Error("Failed to create simulation record", "Error", err)
		return fmt.Errorf("failed to create simulation record: %w", err)
	}
	oldLog.Info("Simulation record created", "path", record.Path)

	// Create the Stores object from the record's initialized stores
	stores := &storage.Stores{
		Motion:   record.Motion,
		Events:   record.Events,
		Dynamics: record.Dynamics,
	}

	// Initialize the simulation manager with the stores from the record
	if err := simManager.Initialize(stores); err != nil {
		oldLog.Error("Failed to initialize simulation manager", "Error", err)
		// Attempt to clean up the created record directory if initialization fails
		cleanupErr := os.RemoveAll(record.Path)
		if cleanupErr != nil {
			oldLog.Error("Failed to cleanup record directory after init failure", "path", record.Path, "cleanupError", cleanupErr)
		}
		return fmt.Errorf("failed to initialize simulation manager: %w", err)
	}

	// Defer closing the record only if creation succeeded
	defer func() {
		if cerr := record.Close(); cerr != nil {
			oldLog.Error("Failed to close simulation record", "Error", cerr)
			if err == nil {
				err = cerr
			}
		}
	}()

	// Defer closing the manager
	defer func() {
		if cerr := simManager.Close(); cerr != nil {
			oldLog.Error("Failed to close simulation manager", "Error", cerr)
			// Don't overwrite the original error if there was one
			if err == nil {
				err = cerr
			}
		}
	}()

	// Run the simulation
	if err = simManager.Run(); err != nil {
		oldLog.Error("Simulation run failed", "Error", err)
		return fmt.Errorf("simulation run failed: %w", err)
	}

	oldLog.Info("Simulation run completed successfully")
	return nil
}

// configFromCtx reads the request body and parses it into a config.Config and validates it
func configFromCtx(c *gin.Context, currentCfg *config.Config, log *logf.Logger) (*config.Config, error) {
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
		log.Warn("Failed to parse numeric form field", "error", firstParseErr)
		return nil, firstParseErr
	}

	// Validate required fields
	if motorDesignation == "" || openRocketFile == "" || openRocketVersion == "" {
		log.Warn("Validation failed: A required field is empty", "motorDesignation", motorDesignation, "openRocketFile", openRocketFile, "openRocketVersion", openRocketVersion)
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

	// We no longer need to initialize the manager here, as it's done
	// with the correct storage.Stores instance inside runSim.
	// m := simulation.NewManager(&simConfig, *logger.GetLogger(currentCfg.Setup.Logging.Level))
	// if err := m.Initialize(); err != nil {
	// 	log.Warn("Manager initialization failed within configFromCtx", "error", err)
	// 	return nil, fmt.Errorf("failed to initialize simulation manager: %w", err)
	// }
	log.Debug("Manager initialized successfully within configFromCtx")

	return &simConfig, nil
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
	log := logger.GetLogger("debug")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Warn("Failed to load config", "error", err)
		return
	}
	log.Info("Config loaded", "Name", cfg.Setup.App.Name, "Version", cfg.Setup.App.Version, "Message", "Starting server")

	r := gin.Default()
	err = r.SetTrustedProxies(nil)
	if err != nil {
		log.Warn("Failed to set trusted proxies", "Error", err)
		os.Exit(1) // Exit on fatal error
	}

	// Get user's home directory
	usr, err := user.Current()
	if err != nil {
		log.Warn("Failed to get user's home directory", "error", err)
		os.Exit(1) // Exit on fatal error
	}
	baseDir := filepath.Join(usr.HomeDir, ".launchrail") // Use ~/.launchrail

	// Initialize RecordManager
	recordManager, err := storage.NewRecordManager(baseDir)
	if err != nil {
		log.Warn("Failed to initialize record manager", "error", err)
		os.Exit(1) // Exit on fatal error
	}
	dataHandler := &DataHandler{records: recordManager, Cfg: cfg, log: log} // Pass cfg and logger here

	// Serve static files (CSS, JS)
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
	majorVersion := "0" // Default if split fails or version is invalid
	if parts := strings.Split(cfg.Setup.App.Version, "."); len(parts) > 0 {
		majorVersion = parts[0]
	}
	apiVersion := fmt.Sprintf("/api/v%s", majorVersion)
	api := r.Group(apiVersion)
	{
		api.POST("/run", dataHandler.handleSimRun) // handleSimRun uses dataHandler.records
		api.GET("/data", dataHandler.ListRecordsAPI)
		api.GET("/explore/:hash", dataHandler.GetExplorerData)
		api.GET("/spec", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/swagger/spec")
		})
		api.GET("/explore/:hash/report", dataHandler.ReportAPIV2) // New endpoint for structured report data
	}

	// Web routes
	r.GET("/data", dataHandler.ListRecords)
	r.GET("/data/:hash/:type", dataHandler.GetRecordData)
	r.DELETE("/data/:hash", dataHandler.DeleteRecord)
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
		if page < 1 {                        // Default to page 1 if not specified, zero, negative, or parse error
			page = 1
		}

		// Update pagination data
		explorerData.Pagination.CurrentPage = page

		render(c, pages.Explorer(explorerData, cfg.Setup.App.Version))
	})
	r.GET("/explore/:hash/json", dataHandler.GetExplorerData)
	r.GET("/explore/:hash/report", dataHandler.ReportAPIV2)

	r.POST("/plot", func(c *gin.Context) {
		hash := c.PostForm("hash")
		source := c.PostForm("source")
		xAxis := c.PostForm("xAxis")
		yAxis := c.PostForm("yAxis")
		zAxis := c.PostForm("zAxis")

		// ... logic to get record using dataHandler.records.GetRecord(hash) ...
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

		// Use plot_transformer for all plotting transformation logic
		plotData, plotLayout, err := plot_transformer.TransformForPlot(headers, rows, source, xAxis, yAxis, zAxis)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"plotData": plotData, "plotLayout": plotLayout})
	})

	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": cfg.Setup.App.Version,
			"apiPath": apiVersion,
		})
	})

	log.Info("Server started", "Port", cfg.Server.Port)
	portStr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err := r.Run(portStr); err != nil {
		log.Warn("Failed to start server", "error", err)
	}
}

// handleSimRun handles API requests to start simulations (now a method of DataHandler)
func (h *DataHandler) handleSimRun(c *gin.Context) {
	h.log.Info("handleSimRun invoked", "time", time.Now().Format(time.RFC3339), "remote_addr", c.ClientIP())

	// Pass the handler's config (h.Cfg) to configFromCtx
	simConfig, err := configFromCtx(c, h.Cfg, h.log)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use the existing record manager from the DataHandler instance (h.records)
	if err := runSim(simConfig, h.records.(*storage.RecordManager), h.log); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Simulation started"})
}

// Deprecated: Use plot_transformer.TransformRowsToFloat64 instead.
func convertToFloat64(data [][]string) [][]float64 {
	return plot_transformer.TransformRowsToFloat64(data)
}

func calculateTotalPages(total int, perPage int) int {
	return int(math.Ceil(float64(total) / float64(perPage)))
}
