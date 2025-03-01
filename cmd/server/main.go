package main

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/simulation"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// Load config
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	log := logger.GetLogger(cfg)

	// Initialize simulation manager
	simManager := simulation.NewManager(cfg, log)
	if err := simManager.Initialize(); err != nil {
		log.Fatal("Failed to initialize simulation", "error", err)
	}

	// Setup HTTP server
	r := gin.Default()

	// Define routes
	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": simManager.GetStatus(),
		})
	})

	r.POST("/simulate", func(c *gin.Context) {
		go simManager.Run() // Run simulation asynchronously
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Simulation started",
		})
	})

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server", "error", err)
	}
}
