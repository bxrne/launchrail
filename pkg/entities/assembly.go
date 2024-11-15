package entities

import "github.com/bxrne/launchrail/pkg/ork"

type Assembly struct {
	config ork.Openrocket
	Rocket Rocket
}

func (a *Assembly) Info() string {
	return a.Rocket.Name + " by " + a.Rocket.Designer
}

func NewAssembly(orkConfig ork.Openrocket, thrustCurvePath string) (*Assembly, error) {
	assembly := &Assembly{config: orkConfig}
	motor, err := NewSolidMotor(thrustCurvePath)
	if err != nil {
		return nil, err
	}

	assembly.Rocket = Rocket{
		Name:     orkConfig.Rocket.Name,
		Designer: orkConfig.Rocket.Designer,
		Motor:    motor,
	}

	return assembly, nil
}
