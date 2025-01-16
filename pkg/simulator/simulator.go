package simulator

import (
	"fmt"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/ecs"
)

// Simulator represents the simulation environment
type Simulator struct {
	ecs       *ecs.ECS
	tick      time.Duration // Duration of each tick
	totalTime float64       // Cumulative time in seconds
	maxTime   float64       // Maximum time in seconds
}

// NewSimulator creates a new Simulator instance
func NewSimulator(cfg *config.Config, ecs *ecs.ECS) *Simulator {
	return &Simulator{
		ecs:     ecs,
		tick:    time.Duration(cfg.Simulation.Step) * time.Millisecond,
		maxTime: cfg.Simulation.MaxTime,
	}
}

// Run starts the simulation loop
func (s *Simulator) Run() {
	fmt.Println("Starting simulation...")

	// Reset total time
	s.totalTime = 0.0

	for {
		// Update the ECS with the current tick duration
		s.ecs.Update(float64(s.tick.Milliseconds()) / 1000.0) // Convert milliseconds to seconds
		s.totalTime += float64(s.tick.Milliseconds()) / 1000.0

		// Print the current state of the simulation
		fmt.Printf("Total Time: %.3f seconds\n", s.totalTime)

		// Optional: Sleep to maintain the desired tick rate
		time.Sleep(s.tick)
	}
}

// String returns a string representation of the Simulator
func (s *Simulator) String() string {
	return "Simulator with ECS: " + s.ecs.Describe()
}
