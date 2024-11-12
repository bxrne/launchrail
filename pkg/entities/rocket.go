package entities

import "github.com/bxrne/launchrail/pkg/ork"

type Rocket struct {
	Name     string
	Designer string
	Motor    *SolidMotor
}

type Assembly struct {
	config ork.Openrocket
	rocket Rocket
}

func (a *Assembly) Info() string {
	return a.rocket.Name + " by " + a.rocket.Designer
}

func NewRocket(orkConfig ork.Openrocket, thrustCurvePath string) (*Assembly, error) {
	assembly := &Assembly{config: orkConfig}
	motor, err := NewSolidMotor(thrustCurvePath)
	if err != nil {
		return nil, err
	}

	assembly.rocket = Rocket{
		Name:     orkConfig.Rocket.Name,
		Designer: orkConfig.Rocket.Designer,
		Motor:    motor,
	}

	return assembly, nil
}
