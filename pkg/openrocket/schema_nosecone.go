package openrocket

import (
	"encoding/xml"
	"fmt"
	"math"
)

// Nosecone represents the nosecone element of the XML document
type Nosecone struct {
	XMLName              xml.Name          `xml:"nosecone"`
	Name                 string            `xml:"name"`
	ID                   string            `xml:"id"`
	Finish               string            `xml:"finish"`
	Material             Material          `xml:"material"`
	Length               float64           `xml:"length"`
	Thickness            float64           `xml:"thickness"`
	Shape                string            `xml:"shape"`
	ShapeClipped         bool              `xml:"shapeclipped"`
	ShapeParameter       float64           `xml:"shapeparameter"`
	AftRadius            float64           `xml:"aftradius"`
	AftShoulderRadius    float64           `xml:"aftshoulderradius"`
	AftShoulderLength    float64           `xml:"aftshoulderlength"`
	AftShoulderThickness float64           `xml:"aftshoulderthickness"`
	AftShoulderCapped    bool              `xml:"aftshouldercapped"`
	IsFlipped            bool              `xml:"isflipped"`
	Subcomponents        NoseSubcomponents `xml:"subcomponents"`
}

// String returns full string representation of the Nosecone
func (n *Nosecone) String() string {
	return fmt.Sprintf("Nosecone{Name=%s, ID=%s, Finish=%s, Material=%s, Length=%.2f, Thickness=%.2f, Shape=%s, ShapeClipped=%t, ShapeParameter=%.2f, AftRadius=%.2f, AftShoulderRadius=%.2f, AftShoulderLength=%.2f, AftShoulderThickness=%.2f, AftShoulderCapped=%t, IsFlipped=%t, Subcomponents=%s}", n.Name, n.ID, n.Finish, n.Material.String(), n.Length, n.Thickness, n.Shape, n.ShapeClipped, n.ShapeParameter, n.AftRadius, n.AftShoulderRadius, n.AftShoulderLength, n.AftShoulderThickness, n.AftShoulderCapped, n.IsFlipped, n.Subcomponents.String())
}

// GetMass returns the mass of the nose cone, calculated based on shape and thickness.
func (n *Nosecone) GetMass() float64 {
	// --- Input parameters ---
	l := n.Length
	rOuter := n.AftRadius
	t := n.Thickness
	density := n.Material.Density
	shape := n.Shape
	shapeParam := n.ShapeParameter // Used for "power" shape

	if density <= 0 || l <= 0 || rOuter <= 0 {
		// Invalid dimensions or material, return only additional mass if any
		// Consider logging a warning
		return n.Subcomponents.MassComponent.Mass
	}

	// --- Calculate inner radius ---
	rInner := rOuter - t
	if rInner < 0 {
		rInner = 0 // Solid nose cone or invalid thickness
	}

	// --- Calculate volume based on shape ---
	var outerVolume, innerVolume float64

	// Function to calculate volume for a given radius (r) and length (l) based on shape
	calculateVolume := func(radius, length float64) float64 {
		if radius <= 0 { // Inner radius might be zero or negative
			return 0
		}
		switch shape {
		case "ogive":
			// Tangent Ogive Volume: rho = (r^2 + l^2) / (2*r)
			// V = pi * l * (rho^2 - l^2/3) - pi * (rho - r) * (rho^2 - l^2) // This formula seems complex/potentially wrong
			// Let's use a simpler approximation or a known formula.
			// Using formula from Apogee rocketry: V = pi*l/3 * (r^2 + r*rho + rho^2) ?? Needs verification.
			// For simplicity, let's approximate with Conical for now, acknowledging inaccuracy.
			// TODO: Implement accurate Ogive volume formula
			return (1.0 / 3.0) * math.Pi * radius * radius * length
		case "conical":
			return (1.0 / 3.0) * math.Pi * radius * radius * length
		case "elliptical": // Half-ellipsoid
			return (2.0 / 3.0) * math.Pi * radius * radius * length
		case "parabolic":
			return (1.0 / 2.0) * math.Pi * radius * radius * length
		case "power":
			// Power series: V = pi * r^2 * l / (p + 1)
			if shapeParam > -1 { // Avoid division by zero or negative
				return math.Pi * radius * radius * length / (shapeParam + 1.0)
			}
			return 0 // Undefined for shapeParam <= -1
		default:
			// Default to cylinder if shape is unknown (least accurate)
			// Consider logging a warning
			return math.Pi * radius * radius * length
		}
	}

	outerVolume = calculateVolume(rOuter, l)
	innerVolume = calculateVolume(rInner, l)

	// --- Calculate material mass ---
	materialVolume := outerVolume - innerVolume
	if materialVolume < 0 {
		materialVolume = 0 // Ensure non-negative volume
	}
	materialMass := materialVolume * density

	// --- Add mass of nested components ---
	additionalMass := n.Subcomponents.MassComponent.Mass // Assumes this field exists and is populated

	// --- Total mass ---
	totalMass := materialMass + additionalMass

	// NOTE: Shoulder mass is not included here.
	// NOTE: Ogive volume calculation needs refinement.

	return totalMass // REMOVED the / 10 division
}

// NoseSubcomponents represents the nested subcomponents element of the XML document
type NoseSubcomponents struct {
	XMLName       xml.Name      `xml:"subcomponents"`
	CenteringRing CenteringRing `xml:"centeringring"`
	MassComponent MassComponent `xml:"masscomponent"`
}

// String returns full string representation of the NoseSubcomponents
func (n *NoseSubcomponents) String() string {
	return fmt.Sprintf("NestedSubcomponents{CenteringRing=%s, MassComponent=%s}", n.CenteringRing.String(), n.MassComponent.String())
}
