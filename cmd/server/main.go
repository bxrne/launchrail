package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func runSim(cfg *config.Config) {
	log := logger.GetLogger(cfg)

	simManager := simulation.NewManager(cfg, log)
	if err := simManager.Initialize(); err != nil {
		log.Fatal("Failed to initialize simulation", "error", err)
	}
	if err := simManager.Run(); err != nil {
		log.Error("Simulation failed", "error", err)
	}
}

func main() {
	r := gin.Default()

	r.POST("/run", func(c *gin.Context) {
		// Read raw YAML body
		yamlData, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to read request body: %v", err)})
			return
		}

		// Create new Viper instance for this request
		v := viper.New()
		v.SetConfigType("yaml")

		// Load YAML from request body
		if err := v.ReadConfig(bytes.NewBuffer(yamlData)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to parse YAML: %v", err)})
			return
		}

		// Create new config instance
		var simConfig config.Config
		if err := v.Unmarshal(&simConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to unmarshal config: %v", err)})
			return
		}

		// Validate configuration
		if err := simConfig.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Run simulation
		go runSim(&simConfig)

		c.JSON(http.StatusAccepted, gin.H{"message": "Simulation started"})
	})

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
