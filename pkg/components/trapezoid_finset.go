package components

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/types"
)

// TrapezoidFinset represents a trapezoidal fin
type TrapezoidFinset struct {
	ecs.BasicEntity
	RootChord  float64
	TipChord   float64
	Span       float64
	SweepAngle float64
	Position   types.Position
	Mass       float64
}

// GetMass returns the mass of the finset
func (f *TrapezoidFinset) GetMass() float64 {
	return f.Mass
}

// NewTrapezoidFinsetFromORK creates a new TrapezoidFinset component from OpenRocket data
func NewTrapezoidFinsetFromORK(basic ecs.BasicEntity, ork *openrocket.RocketDocument) *TrapezoidFinset {
	stage := ork.Subcomponents.Stages[0]
	if len(stage.SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets) == 0 {
		// Log or handle the case where no fin sets are defined
		// For now, returning nil or a default/empty finset might be options
		return nil // Or handle error appropriately
	}
	finset := stage.SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets[0] // Taking the first finset

	return &TrapezoidFinset{
		BasicEntity: basic,
		RootChord:   finset.RootChord,
		TipChord:    finset.TipChord,
		Span:        finset.Height,
		SweepAngle:  finset.SweepLength,
		Mass:        finset.GetMass(),
		Position: types.Position{
			Vec: types.Vector3{
				X: finset.AxialOffset.Value,
				Y: 0,
				Z: 0,
			},
		},
	}
}

// GetPlanformArea returns the planform area of the finset
func (f *TrapezoidFinset) GetPlanformArea() float64 {
	return (f.RootChord + f.TipChord) * f.Span / 2
}
