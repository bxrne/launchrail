package openrocket

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/zerodha/logf"
)

// Package-level logger for openrocket parsing/calculation issues
var pkgLogger = logf.New(logf.Opts{
	Level:  logf.WarnLevel,
	Writer: os.Stderr,
})

// SustainerSubcomponents represents the sustainer subcomponents element of the XML document
type SustainerSubcomponents struct {
	XMLName  xml.Name `xml:"subcomponents"`
	Nosecone Nosecone `xml:"nosecone"`
	BodyTube BodyTube `xml:"bodytube"`
}

// String returns full string representation of the SustainerSubcomponents
func (s *SustainerSubcomponents) String() string {
	return fmt.Sprintf("SustainerSubcomponents{Nosecone=%s, BodyTube=%s}", s.Nosecone.String(), s.BodyTube.String())
}

// BodyTube represents the body tube element of the XML document
type BodyTube struct {
	XMLName       xml.Name              `xml:"bodytube"`
	Name          string                `xml:"name"`
	ID            string                `xml:"id"`
	Finish        string                `xml:"finish"`
	Material      Material              `xml:"material"`
	Length        float64               `xml:"length"`
	Thickness     float64               `xml:"thickness"`
	Radius        string                `xml:"radius"` // WARN: May be 'auto' and num
	Subcomponents BodyTubeSubcomponents `xml:"subcomponents"`
}

// GetMass returns the mass of the bodytube material itself (excluding subcomponents).
// It logs warnings if the mass cannot be calculated due to invalid radius or dimensions.
func (b *BodyTube) GetMass() float64 {
	// --- Input parameters ---
	length := b.Length
	thickness := b.Thickness
	density := b.Material.Density
	radiusStr := b.Radius

	// --- Determine Outer Radius ---
	outerRadius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		// Failed to parse - likely "auto" or invalid format.
		// Cannot calculate mass without a numeric radius derived from context.
		pkgLogger.Warn(fmt.Sprintf("Cannot calculate BodyTube mass for ID=%s: Radius is '%s', requires numeric value.", b.ID, radiusStr))
		return 0.0 // Return 0 mass, acknowledging incompleteness
	}

	// --- Validate inputs ---
	if density <= 0 || length <= 0 || outerRadius <= 0 {
		// Invalid dimensions or material
		pkgLogger.Warn(fmt.Sprintf("Invalid dimensions/material for BodyTube ID=%s: density=%.4f, length=%.4f, outerRadius=%.4f", b.ID, density, length, outerRadius))
		return 0.0
	}

	// --- Calculate Inner Radius ---
	innerRadius := outerRadius - thickness
	if innerRadius < 0 {
		innerRadius = 0 // Treat as solid rod if thickness >= outerRadius
	}

	// --- Calculate Volume & Mass ---
	// Volume = pi * (R_outer^2 - R_inner^2) * L
	materialVolume := math.Pi * (outerRadius*outerRadius - innerRadius*innerRadius) * length
	materialMass := materialVolume * density

	// Note: Mass of subcomponents (fins, parachutes, etc.) inside this tube
	// are handled separately by the caller (e.g., calculateTotalMass).

	return materialMass
}

// String returns full string representation of the BodyTube
func (b *BodyTube) String() string {
	return fmt.Sprintf("BodyTube{Name=%s, ID=%s, Finish=%s, Material=%s, Length=%.2f, Thickness=%.2f, Radius=%s, Subcomponents=%s}", b.Name, b.ID, b.Finish, b.Material.String(), b.Length, b.Thickness, b.Radius, b.Subcomponents.String())

}

// BodyTubeSubcomponents represents the nested subcomponents element of the XML document
type BodyTubeSubcomponents struct {
	XMLName          xml.Name          `xml:"subcomponents"`
	InnerTube        InnerTube         `xml:"innertube"`
	TrapezoidFinsets []TrapezoidFinset `xml:"trapezoidfinset"` // Changed to slice
	Parachute        Parachute         `xml:"parachute"`
	CenteringRings   []CenteringRing   `xml:"centeringring"`
	Shockcord        Shockcord         `xml:"shockcord"`
}

// String returns full string representation of the BodyTubeSubcomponents
func (b *BodyTubeSubcomponents) String() string {
	var trapezoidFinsetStr string
	for i, tf := range b.TrapezoidFinsets {
		trapezoidFinsetStr += tf.String()
		if i < len(b.TrapezoidFinsets)-1 {
			trapezoidFinsetStr += ", "
		}
	}

	var centeringRings string
	for _, cr := range b.CenteringRings {
		centeringRings += cr.String()
		if cr != b.CenteringRings[len(b.CenteringRings)-1] {
			centeringRings += ", "
		}
	}

	return fmt.Sprintf("BodyTubeSubcomponents{InnerTube=%s, TrapezoidFinset=(%s), Parachute=%s, CenteringRings=(%s), Shockcord=%s}", b.InnerTube.String(), trapezoidFinsetStr, b.Parachute.String(), centeringRings, b.Shockcord.String())
}

// InnerTube represents the inner tube element of the XML document
type InnerTube struct {
	XMLName              xml.Name               `xml:"innertube"`
	Name                 string                 `xml:"name"`
	ID                   string                 `xml:"id"`
	AxialOffset          AxialOffset            `xml:"axialoffset"`
	Position             Position               `xml:"position"`
	Material             Material               `xml:"material"`
	Length               float64                `xml:"length"`
	RadialPosition       float64                `xml:"radialposition"`
	RadialDirection      float64                `xml:"radialdirection"`
	OuterRadius          float64                `xml:"outerradius"`
	Thickness            float64                `xml:"thickness"`
	ClusterConfiguration string                 `xml:"clusterconfiguration"`
	ClusterScale         float64                `xml:"clusterscale"`
	ClusterRotation      float64                `xml:"clusterrotation"`
	MotorMount           MotorMount             `xml:"motormount"`    // Direct motor mount info
	Subcomponents        InnerTubeSubcomponents `xml:"subcomponents"` // Contains nested components like another motor mount? (Schema unclear)
}

// GetMass calculates the mass of the inner tube material itself (excluding subcomponents like motor).
func (i *InnerTube) GetMass() float64 {
	// --- Input parameters ---
	length := i.Length
	thickness := i.Thickness
	density := i.Material.Density
	outerRadius := i.OuterRadius // InnerTube uses numeric OuterRadius

	if density <= 0 || length <= 0 || outerRadius <= 0 || thickness <= 0 || thickness >= outerRadius {
		// Invalid dimensions or material
		pkgLogger.Warn(fmt.Sprintf("Invalid dimensions/material for InnerTube ID=%s: density=%.4f, length=%.4f, outerRadius=%.4f", i.ID, density, length, outerRadius))
		return 0.0
	}

	// --- Calculate Inner Radius ---
	innerRadius := outerRadius - thickness

	// --- Calculate Volume & Mass ---
	// Volume = pi * (R_outer^2 - R_inner^2) * L
	materialVolume := math.Pi * (outerRadius*outerRadius - innerRadius*innerRadius) * length
	materialMass := materialVolume * density

	// Note: Mass of subcomponents (e.g., motor mount, internal components)
	// are handled separately by the caller.
	if math.IsNaN(materialMass) || materialMass < 0 {
		fmt.Printf("Warning: Invalid mass (%.4f) calculated for InnerTube '%s' material, returning 0.\n", materialMass, i.Name)
		return 0.0
	}

	return materialMass
}

// String returns full string representation of the InnerTube
func (i *InnerTube) String() string {
	return fmt.Sprintf("InnerTube{Name=%s, ID=%s, AxialOffset=%s, Position=%s, Material=%s, Length=%.2f, RadialPosition=%.2f, RadialDirection=%.2f, OuterRadius=%.2f, Thickness=%.2f, ClusterConfiguration=%s, ClusterScale=%.2f, ClusterRotation=%.2f, MotorMount=%s, Subcomponents=%s}", i.Name, i.ID, i.AxialOffset.String(), i.Position.String(), i.Material.String(), i.Length, i.RadialPosition, i.RadialDirection, i.OuterRadius, i.Thickness, i.ClusterConfiguration, i.ClusterScale, i.ClusterRotation, i.MotorMount.String(), i.Subcomponents.String())
}

// InnerTubeSubcomponents represents the nested subcomponents for an InnerTube.
// Note: The exact structure and purpose within InnerTube needs clarification from OR schema/usage.
// It seems redundant with the direct MotorMount field.
type InnerTubeSubcomponents struct {
	XMLName    xml.Name   `xml:"subcomponents"`
	MotorMount MotorMount `xml:"motormount"` // Often empty if motor is directly in InnerTube?
}

// String returns a string representation of InnerTubeSubcomponents.
func (s *InnerTubeSubcomponents) String() string {
	return fmt.Sprintf("InnerTubeSubcomponents{MotorMount=%s}", s.MotorMount.String())
}
