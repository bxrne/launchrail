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

func configFromCtx(c *gin.Context) (*config.Config, error) {
	yamlData, err := c.GetRawData()
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	v := viper.New()
	v.SetConfigType("yaml")

	if err := v.ReadConfig(bytes.NewBuffer(yamlData)); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	var simConfig config.Config
	if err := v.Unmarshal(&simConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := simConfig.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &simConfig, nil
}

func main() {
	r := gin.Default()

	r.POST("/run", func(c *gin.Context) {
		simConfig, err := configFromCtx(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		go runSim(simConfig)

		c.JSON(http.StatusAccepted, gin.H{"message": "Simulation started"})
	})

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
