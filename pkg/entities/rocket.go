package entities

import (
	"github.com/bxrne/launchrail/internal/openrocket"
	"github.com/bxrne/launchrail/pkg/components"
)

type Rocket struct {
	Name  string
	Nose  *components.Nose
	Fins  *components.Fins
	Motor *components.Motor
}

func NewRocket(orkConfig *openrocket.Openrocket, motor *components.Motor) (*Rocket, error) {
	name := orkConfig.Rocket.Name

	nose, err := components.NewNose(orkConfig)
	if err != nil {
		return nil, err
	}

	fins, err := components.NewFins(orkConfig)
	if err != nil {
		return nil, err
	}

	return &Rocket{
		Name:  name,
		Nose:  nose,
		Fins:  fins,
		Motor: motor,
	}, nil

}
