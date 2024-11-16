package physics

import "github.com/bxrne/launchrail/pkg/types"

type AeroCoefficients struct {
	CD            float64 // Drag coefficient
	ReferenceArea float64

	// INFO: 6DOF specific coefficients
	CL *float64       // Lift coefficient
	CM *float64       // Moment coefficient
	CP *types.Vector3 // Center of pressure
}

type AeroForces struct {
	Drag   types.Vector3
	Lift   types.Vector3
	Moment *types.Vector3 `json:",omitempty"` // INFO: Only for 6DOF
}
