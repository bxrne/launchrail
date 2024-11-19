package components

import "github.com/bxrne/launchrail/internal/openrocket"

type Rocket struct {
	Design *openrocket.Openrocket
	Motor  *SolidMotor
}

func NewRocket(orkConfig *openrocket.Openrocket, motor *SolidMotor) *Rocket {
	return &Rocket{Design: orkConfig, Motor: motor}
}

func (r *Rocket) Info() string {
	return r.Design.Rocket.Name + " by " + r.Design.Rocket.Designer
}
