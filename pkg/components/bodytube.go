package components

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/EngoEngine/ecs"

	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/types"
)

// Bodytube represents the bodytube entity of a rocket
type Bodytube struct {
	ID            ecs.BasicEntity
	Position      types.Vector3
	Radius        float64
	Length        float64
	Mass          float64
	Thickness     float64
	Density       float64 // Material density
	Finish        string  // Surface finish
	MaterialName  string  // Name of material
	MaterialType  string  // Type of material
	CrossSection  float64 // Cross-sectional area
	SurfaceArea   float64 // Total surface area
	Volume        float64 // Volume of material
	CenterOfMass  types.Vector3
	InertiaTensor types.Matrix3x3
}

// NewBodytube creates a new bodytube instance
func NewBodytube(id ecs.BasicEntity, radius, length, thickness, density float64) *Bodytube {
	bt := &Bodytube{
		ID:        id,
		Position:  types.Vector3{X: 0, Y: 0, Z: 0},
		Radius:    radius,
		Length:    length,
		Thickness: thickness,
		Density:   density,
	}
	bt.calculateAndSetProperties()
	return bt
}

// NewBodytubeFromORK creates a new bodytube instance from an ORK Document
func NewBodytubeFromORK(id ecs.BasicEntity, orkData *openrocket.OpenrocketDocument) (*Bodytube, error) {
	if orkData == nil || orkData.Rocket.XMLName.Local == "" || len(orkData.Rocket.Subcomponents.Stages) == 0 {
		return nil, fmt.Errorf("invalid OpenRocket data: missing rocket, stages, or nil orkData")
	}

	orkBodytube := orkData.Rocket.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube

	// Parse radius which may be in "auto X.XX" format
	radiusStr := orkBodytube.Radius
	var radius float64
	radiusStr = strings.TrimSpace(radiusStr)
	if strings.HasPrefix(radiusStr, "auto") {
		// Remove 'auto' and any whitespace
		trimmed := strings.TrimSpace(strings.TrimPrefix(radiusStr, "auto"))
		// Trim surrounding quotes if present
		trimmed = strings.Trim(trimmed, "\"")
		parsed, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid BodyTube radius '%s': %v", radiusStr, err)
		}
		radius = parsed
	} else {
		parsed, err := strconv.ParseFloat(radiusStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid BodyTube radius '%s': %v", radiusStr, err)
		}
		radius = parsed
	}

	// Calculate areas and volume
	crossSection := math.Pi * radius * radius
	circumference := 2 * math.Pi * radius
	surfaceArea := circumference * orkBodytube.Length
	// Calculate material volume for a hollow cylinder
	innerRadius := radius - orkBodytube.Thickness
	// Ensure innerRadius is not negative, which could happen with bad data
	if innerRadius < 0 {
		innerRadius = 0
	}
	materialVolume := math.Pi * (math.Pow(radius, 2) - math.Pow(innerRadius, 2)) * orkBodytube.Length

	// Calculate mass based on material density and volume
	mass := orkBodytube.Material.Density * materialVolume

	bt := &Bodytube{
		ID:           id,
		Position:     types.Vector3{X: 0, Y: 0, Z: 0},
		Radius:       radius,
		Length:       orkBodytube.Length,
		Mass:         mass,
		Thickness:    orkBodytube.Thickness,
		Density:      orkBodytube.Material.Density,
		Finish:       orkBodytube.Finish,
		MaterialName: orkBodytube.Material.Name,
		MaterialType: orkBodytube.Material.Type,
		CrossSection: crossSection,
		SurfaceArea:  surfaceArea,
		Volume:       materialVolume, // Store the material volume
	}
	bt.calculateAndSetProperties()
	return bt, nil
}

// String returns a string representation of the bodytube
func (b *Bodytube) String() string {
	return fmt.Sprintf("Bodytube{ID: %d, Position: %v, Radius: %.2f, Length: %.2f, Mass: %.2f, Thickness: %.2f, Material: %s, Density: %.2f}",
		b.ID.ID(), b.Position, b.Radius, b.Length, b.Mass, b.Thickness, b.MaterialName, b.Density)
}

// Update updates the bodytube (currently does nothing)
func (b *Bodytube) Update(dt float64) error {
	// INFO: Empty, just meeting interface requirements
	return nil
}

// Type returns the type of the component
func (b *Bodytube) Type() string {
	return "Bodytube"
}

// GetPlanformArea returns the planform area of the bodytube
func (b *Bodytube) GetPlanformArea() float64 {
	return math.Pi * b.Radius * b.Radius
}

// GetMass returns the mass of the bodytube
func (b *Bodytube) GetMass() float64 {
	return b.Mass
}

// GetDensity returns the material density of the bodytube
func (b *Bodytube) GetDensity() float64 {
	return b.Density
}

// GetVolume returns the volume of the bodytube material
func (b *Bodytube) GetVolume() float64 {
	return b.Volume
}

// GetSurfaceArea returns the total surface area of the bodytube
func (b *Bodytube) GetSurfaceArea() float64 {
	return b.SurfaceArea
}

// calculateAndSetProperties calculates the mass, volume, surface area, cross-section,
// center of mass, and inertia tensor for the body tube.
func (bt *Bodytube) calculateAndSetProperties() {
	// Calculate Volume and Mass
	outerVolume := math.Pi * bt.Radius * bt.Radius * bt.Length
	innerRadius := bt.Radius - bt.Thickness
	if innerRadius < 0 {
		innerRadius = 0 // Cannot have negative inner radius
	}
	innerVolume := math.Pi * innerRadius * innerRadius * bt.Length
	bt.Volume = outerVolume - innerVolume
	bt.Mass = bt.Volume * bt.Density

	// Calculate Surface Area and Cross Section (simplified)
	bt.SurfaceArea = 2 * math.Pi * bt.Radius * bt.Length // Outer surface area
	bt.CrossSection = math.Pi * bt.Radius * bt.Radius    // Outer cross-sectional area

	// Calculate Center of Mass
	// For a uniform cylinder, the CM is at its geometric center.
	// Assuming bt.Position is the base of the bodytube along the rocket's main axis (e.g., X or Y).
	// The CM would be at Length/2 along that axis, relative to bt.Position if bt.Position is one end.
	// If bt.Position is already the geometric center from ORK, then CM relative to that is (0,0,0).
	// For now, assume component's local CM is at its geometric center (0,0,0) relative to its `Position` field.
	// The `Position` field itself will be relative to rocket origin.
	bt.CenterOfMass = types.Vector3{X: 0, Y: 0, Z: 0} // Relative to Bodytube's own Position reference

	// Calculate Inertia Tensor (about bt.CenterOfMass, aligned with principal axes)
	// Assuming X is the longitudinal axis of the rocket/bodytube.
	// For a hollow cylinder:
	// I_xx (roll axis) = 0.5 * M * (R_outer^2 + R_inner^2)
	// I_yy = I_zz (pitch/yaw axes) = (1/12) * M * (3*(R_outer^2 + R_inner^2) + L^2)

	if bt.Mass <= 1e-9 { // Avoid issues if mass is zero
		bt.InertiaTensor = types.Matrix3x3{}
		return
	}

	rOuterSq := bt.Radius * bt.Radius
	rInnerSq := innerRadius * innerRadius
	lengthSq := bt.Length * bt.Length

	ixx := 0.5 * bt.Mass * (rOuterSq + rInnerSq)
	iyy := (1.0 / 12.0) * bt.Mass * (3.0*(rOuterSq+rInnerSq) + lengthSq)
	izz := iyy // Due to symmetry for a cylinder about Y and Z axes

	// Assuming principal axes align with rocket body axes (X longitudinal)
	// If not, rotation would be needed.
	bt.InertiaTensor = types.Matrix3x3{
		M11: ixx, M12: 0, M13: 0,
		M21: 0, M22: iyy, M23: 0,
		M31: 0, M32: 0, M33: izz,
	}
}

// GetInertiaTensor returns the inertia tensor of the body tube.
// This is about its own CM and aligned with its principal axes (assumed to be rocket body axes).
func (bt *Bodytube) GetInertiaTensor() types.Matrix3x3 {
	return bt.InertiaTensor
}

// GetCenterOfMass returns the Center of Mass of the body tube relative to its own Position reference point.
func (bt *Bodytube) GetCenterOfMass() types.Vector3 {
	return bt.CenterOfMass
}
