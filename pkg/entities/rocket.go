package entities

import "github.com/bxrne/launchrail/pkg/physics"

type Rocket struct {
	Name       string
	Designer   string
	Motor      *SolidMotor
	AeroCoeffs *physics.AeroCoefficients
	AeroForces *physics.AeroForces
}
