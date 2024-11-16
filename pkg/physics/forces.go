package physics

import "github.com/bxrne/launchrail/pkg/types"

type ForceCalculator interface {
	CalculateForces(state types.State, atmosphere Atmosphere) AeroForces
}
