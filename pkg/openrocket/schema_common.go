package openrocket

import (
	"encoding/xml"
	"fmt"
)

// Contains the 'utility' or repeated schema

// AxialOffset represents the axial offset element of the XML document
type AxialOffset struct {
	XMLName xml.Name `xml:"axialoffset"`
	Method  string   `xml:"method,attr"`
	Value   float64  `xml:",chardata"`
}

// String returns full string representation of the AxialOffset
func (a *AxialOffset) String() string {
	return fmt.Sprintf("AxialOffset{Method=%s, Value=%.2f}", a.Method, a.Value)
}

// Position represents the position element of the XML document
type Position struct {
	XMLName xml.Name `xml:"position"`
	Value   float64  `xml:",chardata"`
	Type    string   `xml:"type,attr"`
}

// String returns full string representation of the Position
func (p *Position) String() string {
	return fmt.Sprintf("Position{Value=%.2f, Type=%s}", p.Value, p.Type)
}

// Material represents the material element of the XML document
type Material struct {
	XMLName xml.Name `xml:"material"`
	Type    string   `xml:"type,attr"`
	Density float64  `xml:"density,attr"`
	Name    string   `xml:",chardata"`
}

// String returns full string representation of the Material
func (m *Material) String() string {
	return fmt.Sprintf("Material{Type=%s, Density=%.2f, Name=%s}", m.Type, m.Density, m.Name)
}

// FilletMaterial represents the fillet material element of the XML document (XMLName is the only delta from Material)
type FilletMaterial struct {
	XMLName xml.Name `xml:"filletmaterial"`
	Type    string   `xml:"type,attr"`
	Density float64  `xml:"density,attr"`
	Name    string   `xml:",chardata"`
}

// String returns full string representation of the filletmaterial
func (f *FilletMaterial) String() string {
	return fmt.Sprintf("FilletMaterial{Type=%s, Density=%.2f, Name=%s}", f.Type, f.Density, f.Name)
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

// String returns full string representation of the MassComponent
func (m *MassComponent) String() string {
	return fmt.Sprintf("MassComponent{Name=%s, ID=%s, AxialOffset=%s, Position=%s, PackedLength=%.2f, PackedRadius=%.2f, RadialPosition=%.2f, RadialDirection=%.2f, Mass=%.2f, Type=%s}", m.Name, m.ID, m.AxialOffset.String(), m.Position.String(), m.PackedLength, m.PackedRadius, m.RadialPosition, m.RadialDirection, m.Mass, m.Type)
}

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
