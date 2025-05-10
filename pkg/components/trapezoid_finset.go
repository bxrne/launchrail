package components

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/types"
	"log"
	"math"
)

// TrapezoidFinset represents a trapezoidal fin set component
// All properties are for the entire set unless otherwise specified.
// Individual fin properties are used to calculate set properties.
type TrapezoidFinset struct {
	ecs.BasicEntity
	RootChord     float64            // Root chord of a single fin
	TipChord      float64            // Tip chord of a single fin
	Span          float64            // Span (height) of a single fin
	SweepDistance float64            // Sweep distance (or length) of the leading edge of a single fin
	Thickness     float64            // Thickness of a single fin
	FinCount      int                // Number of fins in the set
	Material      openrocket.Material // Material of the fins
	Position      types.Vector3      // Axial position of the fin set's attachment point (e.g., leading edge of root chord of fin 0)
	Mass          float64            // Total mass of the entire fin set (all fins)
	CenterOfMass  types.Vector3      // CM of the entire fin set, relative to rocket origin
}

// calculateAndSetCenterOfMass calculates the center of mass for the entire fin set
// and updates the CenterOfMass field.
// The Position field is assumed to be the attachment point of the fin set (e.g., leading edge of root chord of fin 0).
func (f *TrapezoidFinset) calculateAndSetCenterOfMass() {
	if (f.RootChord + f.TipChord) == 0 { // Avoid division by zero
		log.Printf("Warning: Sum of RootChord and TipChord is zero for finset. Cannot calculate CM. Defaulting CM to Position.")
		f.CenterOfMass = f.Position
		return
	}

	// Calculate the x-coordinate of a single fin's CG relative to its root chord's leading edge.
	// Formula: (SweepDistance * (RootChord + 2*TipChord) + (RootChord^2 + RootChord*TipChord + TipChord^2)) / (3 * (RootChord + TipChord))
	xCgLocalNum := (f.SweepDistance * (f.RootChord + 2*f.TipChord)) + (math.Pow(f.RootChord, 2) + f.RootChord*f.TipChord + math.Pow(f.TipChord, 2))
	xCgLocalDen := 3 * (f.RootChord + f.TipChord)
	xCgLocal := xCgLocalNum / xCgLocalDen

	// The CM of the fin set (assuming symmetrical placement)
	// X-coordinate is the attachment point's X + local fin CG's X.
	// Y and Z coordinates are assumed to be the same as the attachment point's Y and Z (typically 0 for centerline mounting).
	f.CenterOfMass.X = f.Position.X + xCgLocal
	f.CenterOfMass.Y = f.Position.Y
	f.CenterOfMass.Z = f.Position.Z
}

// GetMass returns the total mass of the finset
func (f *TrapezoidFinset) GetMass() float64 {
	return f.Mass
}

// NewTrapezoidFinsetFromORK creates a new TrapezoidFinset component from OpenRocket data
func NewTrapezoidFinsetFromORK(basic ecs.BasicEntity, ork *openrocket.RocketDocument) *TrapezoidFinset {
	stage := ork.Subcomponents.Stages[0]
	if len(stage.SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets) == 0 {
		log.Println("Warning: No trapezoid fin sets found in ORK data for component creation.")
		return nil 
	}
	// Assuming we're building the component from the first finset defined
	orkFinset := stage.SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets[0] 

	// Mass of a single fin from ORK data (which already considers its geometry and material density)
	singleFinMass := orkFinset.GetMass() 
	totalMass := singleFinMass * float64(orkFinset.FinCount)

	if totalMass <= 0 {
		log.Printf("Warning: Calculated total mass for finset '%s' is zero or negative (%.4f). Check ORK finset data.", orkFinset.Name, totalMass)
		// Potentially return nil or a component with zero mass if this is an error state
	}

	finsetComponent := &TrapezoidFinset{
		BasicEntity:   basic,
		RootChord:     orkFinset.RootChord,
		TipChord:      orkFinset.TipChord,
		Span:          orkFinset.Height,
		SweepDistance: orkFinset.SweepLength, 
		Thickness:     orkFinset.Thickness,
		FinCount:      orkFinset.FinCount,
		Material:      orkFinset.Material,
		Mass:          totalMass, 
		Position: types.Vector3{ 
			X: orkFinset.Position.Value, 
			Y: 0, 
			Z: 0, 
		},
	}

	// Calculate and set the center of mass after all properties are populated
	finsetComponent.calculateAndSetCenterOfMass()

	return finsetComponent
}

// GetPlanformArea returns the planform area of a single fin in the set
func (f *TrapezoidFinset) GetPlanformArea() float64 {
	return (f.RootChord + f.TipChord) * f.Span / 2
}
