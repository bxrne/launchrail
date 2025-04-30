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
	ID           ecs.BasicEntity
	Position     types.Vector3
	Radius       float64
	Length       float64
	Mass         float64
	Thickness    float64
	Density      float64 // Material density
	Finish       string  // Surface finish
	MaterialName string  // Name of material
	MaterialType string  // Type of material
	CrossSection float64 // Cross-sectional area
	SurfaceArea  float64 // Total surface area
	Volume       float64 // Volume of material
}

// NewBodytube creates a new bodytube instance
func NewBodytube(id ecs.BasicEntity, radius, length, mass, thickness float64) *Bodytube {
	return &Bodytube{
		ID:        id,
		Position:  types.Vector3{X: 0, Y: 0, Z: 0},
		Radius:    radius,
		Length:    length,
		Mass:      mass,
		Thickness: thickness,
	}
}

// NewBodytubeFromORK creates a new bodytube instance from an ORK Document
func NewBodytubeFromORK(id ecs.BasicEntity, orkData *openrocket.RocketDocument) (*Bodytube, error) {
	if orkData == nil || len(orkData.Subcomponents.Stages) == 0 {
		return nil, fmt.Errorf("invalid OpenRocket data: missing stages")
	}

	orkBodytube := orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube

	// Parse radius which may be in "auto X.XX" format
	radiusStr := orkBodytube.Radius
	var radius float64
	radiusStr = strings.TrimSpace(radiusStr)
	if strings.HasPrefix(radiusStr, "auto") {
		// Remove 'auto' and any whitespace
		trimmed := strings.TrimSpace(strings.TrimPrefix(radiusStr, "auto"))
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
	volume := crossSection * orkBodytube.Length * orkBodytube.Thickness

	// Calculate mass based on material density and volume
	mass := orkBodytube.Material.Density * volume

	return &Bodytube{
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
		Volume:       volume,
	}, nil
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
