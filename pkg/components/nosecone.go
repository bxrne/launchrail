package components

import (
	"fmt"
	"math"

	"github.com/EngoEngine/ecs"

	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/types"
)

// Nosecone represents the nosecone entity of a rocket
type Nosecone struct {
	ID                ecs.BasicEntity
	Position          types.Vector3
	Radius            float64
	Length            float64
	Mass              float64
	ShapeParameter    float64
	Thickness         float64
	Shape             string
	Finish            string
	MaterialName      string
	MaterialType      string
	Density           float64
	Volume            float64
	SurfaceArea       float64
	AftShoulderRadius float64
	AftShoulderLength float64
	AftShoulderCapped bool
	ShapeClipped      bool
	IsFlipped         bool
}

// NewNosecone creates a new nosecone instance
func NewNosecone(id ecs.BasicEntity, radius, length, mass, shapeParameter float64) *Nosecone {
	return &Nosecone{
		ID:             id,
		Position:       types.Vector3{X: 0, Y: 0, Z: 0},
		Radius:         radius,
		Length:         length,
		Mass:           mass,
		ShapeParameter: shapeParameter,
	}
}

// NewNoseconeFromORK creates a new nosecone instance from an ORK Document
func NewNoseconeFromORK(id ecs.BasicEntity, orkData *openrocket.RocketDocument) *Nosecone {
	orkNosecone := orkData.Subcomponents.Stages[0].SustainerSubcomponents.Nosecone

	// Calculate volume and surface area (simplified approximation)
	baseArea := math.Pi * orkNosecone.AftRadius * orkNosecone.AftRadius
	slantHeight := math.Sqrt(orkNosecone.Length*orkNosecone.Length + orkNosecone.AftRadius*orkNosecone.AftRadius)
	lateralArea := math.Pi * orkNosecone.AftRadius * slantHeight
	// volume := (baseArea * orkNosecone.Length) / 3.0 // This was solid volume, removed as unused
	surfaceArea := lateralArea + baseArea

	// Calculate total mass including any additional mass components
	materialVolume := lateralArea * orkNosecone.Thickness
	bodyMass := materialVolume * orkNosecone.Material.Density
	additionalMass := orkNosecone.Subcomponents.MassComponent.Mass
	totalMass := bodyMass + additionalMass

	return &Nosecone{
		ID:                id,
		Position:          types.Vector3{X: 0, Y: 0, Z: 0},
		Radius:            orkNosecone.AftRadius,
		Length:            orkNosecone.Length,
		Mass:              totalMass,
		ShapeParameter:    orkNosecone.ShapeParameter,
		Thickness:         orkNosecone.Thickness,
		Shape:             orkNosecone.Shape,
		Finish:            orkNosecone.Finish,
		MaterialName:      orkNosecone.Material.Name,
		MaterialType:      orkNosecone.Material.Type,
		Density:           orkNosecone.Material.Density,
		Volume:            materialVolume,
		SurfaceArea:       surfaceArea,
		AftShoulderRadius: orkNosecone.AftShoulderRadius,
		AftShoulderLength: orkNosecone.AftShoulderLength,
		AftShoulderCapped: orkNosecone.AftShoulderCapped,
		ShapeClipped:      orkNosecone.ShapeClipped,
		IsFlipped:         orkNosecone.IsFlipped,
	}
}

// String returns a string representation of the Nosecone
func (n *Nosecone) String() string {
	return fmt.Sprintf("Nosecone{ID: %d, Position: %v, Radius: %.2f, Length: %.2f, Mass: %.2f, Shape: %s, Material: %s, Density: %.2f}",
		n.ID.ID(), n.Position, n.Radius, n.Length, n.Mass, n.Shape, n.MaterialName, n.Density)
}

// Update updates the nosecone (currently does nothing)
func (n *Nosecone) Update(dt float64) error {
	// INFO: Empty, just meeting interface requirements
	return nil
}

// Type returns the type of the component
func (n *Nosecone) Type() string {
	return "Nosecone"
}

// GetPlanformArea returns the planform area of the nosecone
func (n *Nosecone) GetPlanformArea() float64 {
	return math.Pi * n.Radius * n.Radius
}

// GetMass returns the mass of the nosecone
func (n *Nosecone) GetMass() float64 {
	return n.Mass
}

// GetVolume returns the volume of the nosecone material
func (n *Nosecone) GetVolume() float64 {
	return n.Volume
}

// GetSurfaceArea returns the surface area of the nosecone
func (n *Nosecone) GetSurfaceArea() float64 {
	return n.SurfaceArea
}

// GetDensity returns the material density
func (n *Nosecone) GetDensity() float64 {
	return n.Density
}

// GetPosition returns the nosecone's reference position (e.g., its tip) in rocket coordinates.
func (nc *Nosecone) GetPosition() types.Vector3 {
	// TODO: Ensure nc.Position is correctly set during NewNoseconeFromORK relative to a common rocket origin.
	return nc.Position
}

// GetCenterOfMassLocal returns the nosecone's center of mass relative to its Position (tip).
// This is a simplified placeholder. Actual CG depends on shape and wall thickness.
func (nc *Nosecone) GetCenterOfMassLocal() types.Vector3 {
	// Assuming Position is the tip and the nosecone extends along the positive X-axis.
	// For many common shapes (cone, ogive), the CG is roughly 0.4 to 0.5 * Length from the tip.
	// Using 0.5 * Length as a rough placeholder. A more precise calculation is needed.
	// Note: OpenRocket typically defines component positions from the tip of the nosecone downwards (positive X).
	// So, if nc.Position is the tip (0,0,0 in local frame), CG is at (+X_cg_local, 0, 0).
	localCgX := 0.5 * nc.Length 
	// TODO: Implement accurate local CG calculation based on nc.Shape, nc.Thickness, etc.
	// log.Warn("Nosecone.GetCenterOfMassLocal() returning placeholder.")
	return types.Vector3{X: localCgX, Y: 0, Z: 0}
}

// GetInertiaTensorLocal returns the nosecone's inertia tensor about its local CG, aligned with rocket axes.
// This is a placeholder, returning a zero matrix.
func (nc *Nosecone) GetInertiaTensorLocal() types.Matrix3x3 {
	// TODO: Implement accurate inertia tensor calculation for the nosecone based on its shape, mass distribution.
	// log.Warn("Nosecone.GetInertiaTensorLocal() returning placeholder (zero matrix).")
	return types.Matrix3x3{}
}
