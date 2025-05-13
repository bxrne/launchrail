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
func NewNoseconeFromORK(id ecs.BasicEntity, orkData *openrocket.OpenrocketDocument) *Nosecone {
	if orkData == nil || orkData.Rocket.XMLName.Local == "" || len(orkData.Rocket.Subcomponents.Stages) == 0 {
		// Optionally log an error or return nil if critical data is missing
		// For now, proceeding assuming valid structure for simplicity, but real-world might need robust error handling.
		// If we return nil, the caller (e.g., initComponentsFromORK) needs to handle it.
		// Returning an empty/default Nosecone might also be an option if that's preferable to nil.
		return nil // Or handle error appropriately
	}
	orkNosecone := orkData.Rocket.Subcomponents.Stages[0].SustainerSubcomponents.Nosecone

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

// GetPosition returns the global position of the nosecone's reference point (e.g., its tip or base attachment).
func (n *Nosecone) GetPosition() types.Vector3 {
	// TODO: Ensure n.Position is correctly set during NewNoseconeFromORK relative to a common rocket origin.
	// For now, it's initialized to {0,0,0} in NewNoseconeFromORK, which implies its tip is at the rocket's origin if not updated.
	return n.Position
}

// GetCenterOfMassLocal returns the center of mass of the nosecone relative to its own reference point (Position).
// Assumes solid cone, tip at Z=0, base at Z=n.Length. CG is 3/4 Length from the tip.
func (n *Nosecone) GetCenterOfMassLocal() types.Vector3 {
	if n.Length == 0 {
		// Consider logging a warning if appropriate for zero-length nosecone
		return types.Vector3{X: 0, Y: 0, Z: 0}
	}
	// CG for a solid cone is 1/4 height from the base, or 3/4 height from the tip.
	// Assuming tip is at Z=0 and base is at Z=n.Length in local coordinates.
	return types.Vector3{X: 0, Y: 0, Z: (3.0 * n.Length) / 4.0}
}

// GetInertiaTensorLocal returns the inertia tensor of the nosecone about its own center of mass,
// in local coordinates (Z-axis along cone height).
func (n *Nosecone) GetInertiaTensorLocal() types.Matrix3x3 {
	if n.Mass <= 1e-9 { // Effectively zero mass
		return types.Matrix3x3{}
	}
	if n.Length == 0 || n.Radius == 0 {
		// Consider logging a warning for zero dimensions
		return types.Matrix3x3{}
	}

	mass := n.Mass
	height := n.Length // h for cone formula
	radius := n.Radius // R for cone formula

	// Inertia tensor for a solid cone about its CG (origin at CG, Z-axis along height):
	// Ixx_cg = Iyy_cg = mass * ( (3/20)*R^2 + (3/80)*h^2 )
	// Izz_cg = (3/10) * mass * R^2

	ixxIyyCg := mass * ((3.0/20.0)*radius*radius + (3.0/80.0)*height*height)
	izzCg := (3.0 / 10.0) * mass * radius * radius

	return types.Matrix3x3{
		M11: ixxIyyCg, M12: 0, M13: 0,
		M21: 0, M22: ixxIyyCg, M23: 0,
		M31: 0, M32: 0, M33: izzCg,
	}
}
