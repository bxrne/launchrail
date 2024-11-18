package simulation

import "github.com/bxrne/launchrail/pkg/components"

type Simulation struct {
	Rocket      *components.Rocket
	Environment Environment
}

func NewSimulation(rocket *components.Rocket, environment Environment) *Simulation {
	return &Simulation{
		Rocket:      rocket,
		Environment: environment,
	}
}

func (s *Simulation) Info() string {
	return s.Rocket.Info() + "\n" + s.Environment.Info()
}
