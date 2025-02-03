package components

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/openrocket"
)

// TrapezoidFinset represents a trapezoidal fin
type TrapezoidFinset struct {
	ecs.BasicEntity
	RootChord  float64
	TipChord   float64
	Span       float64
	SweepAngle float64
	Position   Position
	Mass       float64
}

// GetMass returns the mass of the finset
func (f *TrapezoidFinset) GetMass() float64 {
	return f.Mass
}

// NewTrapezoidFinsetFromORK creates a new TrapezoidFinset component from OpenRocket data
func NewTrapezoidFinsetFromORK(basic ecs.BasicEntity, ork *openrocket.RocketDocument) *TrapezoidFinset {
	stage := ork.Subcomponents.Stages[0]
	finset := stage.SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinset

	return &TrapezoidFinset{
		BasicEntity: basic,
		RootChord:   finset.RootChord,
		TipChord:    finset.TipChord,
		Span:        finset.Height,
		SweepAngle:  finset.SweepLength,
		Mass:        finset.GetMass(),
		Position: Position{
			X: finset.AxialOffset.Value,
			Y: 0,
			Z: 0,
		},
	}
}

// calculateFinMass calculates mass based on material density and dimensions
func CalculateFinMass(fin *openrocket.TrapezoidFinset) float64 {
	area := (fin.RootChord + fin.TipChord) * fin.Height / 2
	volume := area * fin.Thickness
	density := fin.Material.Density
	return volume * density
}

// GetPlanformArea returns the planform area of the finset
func (f *TrapezoidFinset) GetPlanformArea() float64 {
	return (f.RootChord + f.TipChord) * f.Span / 2
}
