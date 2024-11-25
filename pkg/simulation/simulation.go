package simulation

import (
	"fmt"
	"time"

	"github.com/bxrne/launchrail/pkg/entities"
)

type Simulation struct {
	rocket      *entities.Rocket
	environment Environment
	timeStep    time.Duration
	elapsedTime time.Duration
}

func NewSimulation(rocket *entities.Rocket, environment Environment, timeStepNS int) *Simulation {
	return &Simulation{
		rocket:      rocket,
		environment: environment,
		timeStep:    time.Duration(timeStepNS),
	}
}

func (s *Simulation) Run() error {
	err := s.rocket.Motor.Update(s.timeStep)
	if err != nil {
		return fmt.Errorf("motor update error: %v", err)
	}

	// TODO: Implement rocket flight physics calculations

	s.elapsedTime += s.timeStep

	return nil
}

func (s *Simulation) Info() string {
	return fmt.Sprintf("%s\n%s\nElapsed Time: %s\nMotor State:\n%s",
		s.rocket.Info(),
		s.environment.Info(),
		s.elapsedTime,
		s.rocket.Motor.String(),
	)
}
