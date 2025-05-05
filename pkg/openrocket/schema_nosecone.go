package openrocket

import (
	"encoding/xml"
	"fmt"
	"math"
)

// Nosecone represents the nosecone element of the XML document
type Nosecone struct {
	XMLName              xml.Name              `xml:"nosecone"`
	Name                 string                `xml:"name"`
	ID                   string                `xml:"id"`
	Finish               string                `xml:"finish"`
	Material             Material              `xml:"material"`
	Length               float64               `xml:"length"`
	Thickness            float64               `xml:"thickness"`
	Shape                string                `xml:"shape"`
	ShapeClipped         bool                  `xml:"shapeclipped"`
	ShapeParameter       float64               `xml:"shapeparameter"`
	AftRadius            float64               `xml:"aftradius"`
	AftShoulderRadius    float64               `xml:"aftshoulderradius"`
	AftShoulderLength    float64               `xml:"aftshoulderlength"`
	AftShoulderThickness float64               `xml:"aftshoulderthickness"`
	AftShoulderCapped    bool                  `xml:"aftshouldercapped"`
	IsFlipped            bool                  `xml:"isflipped"`
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
			// Tangent Ogive Volume Calculation (derived from integral)
			if radius <= 0 {
				return 0 // Avoid division by zero and invalid geometry
			}
			r_sq := radius * radius
			l_sq := length * length
			// Avoid division by zero if radius is extremely small, though caught above.
			// Check for length being zero as well.
			if length == 0 {
			    return 0
			}

			R := (r_sq + l_sq) / (2.0 * radius) // Ogive radius (rho)
			h := (r_sq - l_sq) / (2.0 * radius) // x-offset of circle center
			R_sq := R * R

			// arcsin argument must be between -1 and 1.
			// Check for R being zero or very small, although unlikely with l > 0.
			if R == 0 {
			    return 0
			}
			asinArg := length / R
			if asinArg > 1.0 {
				// Handle potential floating point inaccuracies or invalid geometry (l > R)
				asinArg = 1.0
			} else if asinArg < -1.0 {
				// Should not happen with l>0, R>0
				asinArg = -1.0
			}

			// Ensure asinArg is valid before calling arcsin
			if math.IsNaN(asinArg) {
				fmt.Printf("Warning: Invalid asin argument (NaN) for ogive volume calculation (radius=%.4f, length=%.4f). Returning 0.\n", radius, length)
				return 0.0
			}

			volume := math.Pi * (R_sq*length - l_sq*length/3.0 + h*R_sq*math.Asin(asinArg))

			// Final safety check for calculated volume
			if math.IsNaN(volume) || volume < 0 {
				fmt.Printf("Warning: Invalid volume (%.4f) calculated for ogive (radius=%.4f, length=%.4f). Returning 0.\n", volume, radius, length)
				return 0.0
			}
			return volume
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

	return totalMass // REMOVED the / 10 division
}

// NoseSubcomponents represents the nested subcomponents element of the XML document
type NoseSubcomponents struct {
	XMLName       xml.Name      `xml:"subcomponents"`
	MassComponent MassComponent `xml:"masscomponent"`
}

// String returns full string representation of the NoseSubcomponents
func (n *NoseSubcomponents) String() string {
	return fmt.Sprintf("NoseSubcomponents{MassComponent=%s}", n.MassComponent.String())
}
