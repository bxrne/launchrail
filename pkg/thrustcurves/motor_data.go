package thrustcurves

import (
	"github.com/bxrne/launchrail/pkg/designation"
)

type MotorData struct {
	Designation designation.Designation
	ID          string
	Thrust      [][]float64
}
