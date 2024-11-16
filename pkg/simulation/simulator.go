package simulation

import (
	"github.com/bxrne/launchrail/pkg/entities"
	"github.com/bxrne/launchrail/pkg/physics"
	"github.com/bxrne/launchrail/pkg/types"
)

type SimulationConfig struct {
	Mode         types.SimulationMode
	TimeStep     float64
	MaxTime      float64
	InitialState types.State
	Assembly     entities.Assembly
	Environment  entities.Environment
}

type Simulator struct {
	config SimulationConfig
	state  types.State
	forces physics.ForceCalculator
}

func NewSimulator(config SimulationConfig) *Simulator {
	return &Simulator{
		config: config,
		state:  config.InitialState,
	}
}

func (s *Simulator) Step() error {
	// TODO: Implementation of state change and time management
	return nil
}
