package simulator

import (
	"fmt"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/ecs"
)

// Simulator represents the simulation environment
type Simulator struct {
	ecs         *ecs.ECS
	maxTime     float64 // Maximum time in seconds
	elapsedTime float64 // Elapsed time in Seconds
	timeStep    float64 // Time step in seconds
}

// NewSimulator creates a new Simulator instance
func NewSimulator(cfg *config.Config, ecs *ecs.ECS) *Simulator {
	return &Simulator{
		ecs:         ecs,
		maxTime:     cfg.Simulation.MaxTime,
		timeStep:    cfg.Simulation.Step,
		elapsedTime: 0.0,
	}
}

// Run starts the simulation loop
func (s *Simulator) Run() error {
	// Run for all available steps until remainder/none
	for (s.elapsedTime + s.timeStep) < s.maxTime {
		s.elapsedTime += s.timeStep
		err := s.ecs.Update(s.timeStep)
		if err != nil {
			return err
		}

	}

	// Update the ECS with the remaining timeStep
	err := s.ecs.Update(s.maxTime - s.elapsedTime)
	if err != nil {
		return err
	}

	return nil
}

// String returns a string representation of the Simulator
func (s *Simulator) String() string {
	return fmt.Sprintf("Simulator{MaxTime: %.2f, ElapsedTime: %.2f, TimeStep: %.2f, ECS: %s}", s.maxTime, s.elapsedTime, s.timeStep, s.ecs.Describe())
}

// Describe returns a string representation of the simulator
func (s *Simulator) Describe() string {
	return fmt.Sprintf("MaxTime=%.2f, TimeStep=%.2f", s.maxTime, s.timeStep)
}
