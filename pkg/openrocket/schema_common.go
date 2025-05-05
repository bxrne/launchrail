package openrocket

import (
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
)

// INFO: Contains the 'utility' elements that are not component specific

// RadiusOffset represents the radius offset element of the XML document
type RadiusOffset struct {
	XMLName xml.Name `xml:"radiusoffset"`
	Method  string   `xml:"method,attr"`
	Value   float64  `xml:",chardata"`
}

// String returns full string representation of the radiusoffset
func (r *RadiusOffset) String() string {
	return fmt.Sprintf("RadiusOffset{Method=%s, Value=%.2f}", r.Method, r.Value)
}

// AngleOffset represents the angle offset element of the XML document
type AngleOffset struct {
	XMLName xml.Name `xml:"angleoffset"`
	Method  string   `xml:"method,attr"`
	Value   float64  `xml:",chardata"`
}

// String returns full string representation of the angleoffset
func (a *AngleOffset) String() string {
	return fmt.Sprintf("AngleOffset{Method=%s, Value=%.2f}", a.Method, a.Value)
}

// AxialOffset is a calculated offset from the center of the component to the center of the rocket
type AxialOffset struct {
	XMLName xml.Name `xml:"axialoffset"`
	Method  string   `xml:"method,attr"`
	Value   float64  `xml:",chardata"`
}

// String returns full string representation of the AxialOffset
func (a *AxialOffset) String() string {
	return fmt.Sprintf("AxialOffset{Method=%s, Value=%.2f}", a.Method, a.Value)
}

// Position represents the position of the component in the rocket
type Position struct {
	XMLName xml.Name `xml:"position"`
	Value   float64  `xml:",chardata"`
	Type    string   `xml:"type,attr"`
}

// String returns full string representation of the Position
func (p *Position) String() string {
	return fmt.Sprintf("Position{Value=%.2f, Type=%s}", p.Value, p.Type)
}

// CenteringRing represents the centering ring element of the XML document
type CenteringRing struct {
	XMLName            xml.Name    `xml:"centeringring"`
	Name               string      `xml:"name"`
	ID                 string      `xml:"id"`
	InstanceCount      int         `xml:"instancecount"`
	InstanceSeparation float64     `xml:"instanceseparation"`
	AxialOffset        AxialOffset `xml:"axialoffset"`
	Position           Position    `xml:"position"`
	Material           Material    `xml:"material"`
	Length             float64     `xml:"length"`
	RadialPosition     float64     `xml:"radialposition"`
	RadialDirection    float64     `xml:"radialdirection"`
	OuterRadius        string      `xml:"outerradius"` // WARN: May be 'auto'
	InnerRadius        string      `xml:"innerradius"` // WARN: May be 'auto'
}

// GetMass calculates the mass of the centering ring(s).
// NOTE: Cannot calculate mass if OuterRadius or InnerRadius is "auto".
// Returns 0 if radii are "auto" or invalid.
func (c *CenteringRing) GetMass() float64 {
	if c == nil {
		return 0.0
	}
	if c.Material.Density <= 0 || c.Length <= 0 || c.InstanceCount <= 0 {
		return 0.0
	}

	outerR, errOuter := parseRadius(c.OuterRadius)
	innerR, errInner := parseRadius(c.InnerRadius)

	if errOuter != nil || errInner != nil {
		// Cannot calculate without context for "auto"
		// fmt.Printf("Info: Cannot calculate CenteringRing mass for '%s': Radius is 'auto'.\n", c.Name)
		return 0.0
	}

	if outerR <= 0 || innerR < 0 || outerR <= innerR {
		// Invalid dimensions
		// fmt.Printf("Warning: Invalid CenteringRing dimensions for '%s': OuterRadius (%.4f), InnerRadius (%.4f).\n", c.Name, outerR, innerR)
		return 0.0
	}

	// Area of the annulus = pi * (R_outer^2 - R_inner^2)
	area := math.Pi * (outerR*outerR - innerR*innerR)
	volumePerRing := area * c.Length
	massPerRing := volumePerRing * c.Material.Density

	totalMass := massPerRing * float64(c.InstanceCount)

	if math.IsNaN(totalMass) || totalMass < 0 {
		fmt.Printf("Warning: Invalid mass (%.4f) calculated for CenteringRing '%s', returning 0.\n", totalMass, c.Name)
		return 0.0
	}

	return totalMass
}

// String returns full string representation of the CenteringRing
func (c *CenteringRing) String() string {
	return fmt.Sprintf("CenteringRing{Name=%s, ID=%s, InstanceCount=%d, InstanceSeparation=%.2f, AxialOffset=%s, Position=%s, Material=%s, Length=%.2f, RadialPosition=%.2f, OuterRadius=%s, InnerRadius=%s}", c.Name, c.ID, c.InstanceCount, c.InstanceSeparation, c.AxialOffset.String(), c.Position.String(), c.Material.String(), c.Length, c.RadialPosition, c.OuterRadius, c.InnerRadius)
}

// MassComponent represents the mass component element of the XML document
type MassComponent struct {
	XMLName         xml.Name    `xml:"masscomponent"`
	Name            string      `xml:"name"`
	ID              string      `xml:"id"`
	AxialOffset     AxialOffset `xml:"axialoffset"`
	Position        Position    `xml:"position"`
	PackedLength    float64     `xml:"packedlength"`
	PackedRadius    float64     `xml:"packedradius"`
	RadialPosition  float64     `xml:"radialposition"`
	RadialDirection float64     `xml:"radialdirection"`
	Mass            float64     `xml:"mass"`
	Type            string      `xml:"masscomponenttype"`
}

// GetMass returns the pre-defined mass of the component.
func (m *MassComponent) GetMass() float64 {
	if m == nil {
		return 0.0
	}
	if m.Mass < 0 {
		fmt.Printf("Warning: Negative mass (%.4f) specified for MassComponent '%s'.\n", m.Mass, m.Name)
	}
	return m.Mass
}

// String returns full string representation of the MassComponent
func (m *MassComponent) String() string {
	return fmt.Sprintf("MassComponent{Name=%s, ID=%s, AxialOffset=%s, Position=%s, PackedLength=%.2f, PackedRadius=%.2f, RadialPosition=%.2f, RadialDirection=%.2f, Mass=%.2f, Type=%s}", m.Name, m.ID, m.AxialOffset.String(), m.Position.String(), m.PackedLength, m.PackedRadius, m.RadialPosition, m.RadialDirection, m.Mass, m.Type)
}

// Shockcord represents the shockcord element of the XML document
type Shockcord struct {
	XMLName         xml.Name    `xml:"shockcord"`
	Name            string      `xml:"name"`
	ID              string      `xml:"id"`
	AxialOffset     AxialOffset `xml:"axialoffset"`
	Position        Position    `xml:"position"`
	PackedLength    float64     `xml:"packedlength"`
	PackedRadius    float64     `xml:"packedradius"`
	RadialPosition  float64     `xml:"radialposition"`
	RadialDirection float64     `xml:"radialdirection"`
	CordLength      float64     `xml:"cordlength"`
	Material        Material    `xml:"material"`
}

// GetMass calculates the mass of the shockcord.
// NOTE: Assumes Material.Density is LINEAR density (e.g., kg/m).
// If it's volumetric (kg/m^3), this calculation needs cord diameter.
func (s *Shockcord) GetMass() float64 {
	if s == nil { // Add nil check
		return 0.0
	}
	if s.Material.Density <= 0 || s.CordLength <= 0 {
		return 0.0
	}

	mass := s.CordLength * s.Material.Density

	if math.IsNaN(mass) || mass < 0 {
		fmt.Printf("Warning: Invalid mass (%.4f) calculated for Shockcord '%s', returning 0.\n", mass, s.Name)
		return 0.0
	}
	return mass
}

// String returns full string representation of the shockcord
func (s *Shockcord) String() string {
	return fmt.Sprintf("Shockcord{Name=%s, ID=%s, AxialOffset=%s, Position=%s, PackedLength=%.2f, PackedRadius=%.2f, RadialPosition=%.2f, RadialDirection=%.2f, CordLength=%.2f, Material=%s}", s.Name, s.ID, s.AxialOffset.String(), s.Position.String(), s.PackedLength, s.PackedRadius, s.RadialPosition, s.RadialDirection, s.CordLength, s.Material.String())
}

// parseRadius is a helper to convert radius string to float64.
// Returns error if input is "auto" or not a valid float.
func parseRadius(radiusStr string) (float64, error) {
	if radiusStr == "auto" {
		return 0, fmt.Errorf("radius is 'auto'")
	}
	val, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid radius value: %s", radiusStr)
	}
	return val, nil
}
