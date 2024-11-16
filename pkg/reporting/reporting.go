package output

import "github.com/bxrne/launchrail/pkg/types"

type FlightLogger interface {
	LogState(state types.State)
	GetTrajectory() []types.State
}
