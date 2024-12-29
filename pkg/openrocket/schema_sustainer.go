package openrocket

import (
	"encoding/xml"
	"fmt"
)

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

// Nosecone represents the nosecone element of the XML document
type Nosecone struct {
	XMLName              xml.Name            `xml:"nosecone"`
	Name                 string              `xml:"name"`
	ID                   string              `xml:"id"`
	Finish               string              `xml:"finish"`
	Material             Material            `xml:"material"`
	Length               float64             `xml:"length"`
	Thickness            float64             `xml:"thickness"`
	Shape                string              `xml:"shape"`
	ShapeClipped         bool                `xml:"shapeclipped"`
	ShapeParameter       float64             `xml:"shapeparameter"`
	AftRadius            float64             `xml:"aftradius"`
	AftShoulderRadius    float64             `xml:"aftshoulderradius"`
	AftShoulderLength    float64             `xml:"aftshoulderlength"`
	AftShoulderThickness float64             `xml:"aftshoulderthickness"`
	AftShoulderCapped    bool                `xml:"aftshouldercapped"`
	IsFlipped            bool                `xml:"isflipped"`
	Subcomponents        NestedSubcomponents `xml:"subcomponents"`
}

// String returns full string representation of the Nosecone
func (n *Nosecone) String() string {
	return fmt.Sprintf("Nosecone{Name=%s, ID=%s, Finish=%s, Material=%s, Length=%.2f, Thickness=%.2f, Shape=%s, ShapeClipped=%t, ShapeParameter=%.2f, AftRadius=%.2f, AftShoulderRadius=%.2f, AftShoulderLength=%.2f, AftShoulderThickness=%.2f, AftShoulderCapped=%t, IsFlipped=%t, Subcomponents=%s}", n.Name, n.ID, n.Finish, n.Material.String(), n.Length, n.Thickness, n.Shape, n.ShapeClipped, n.ShapeParameter, n.AftRadius, n.AftShoulderRadius, n.AftShoulderLength, n.AftShoulderThickness, n.AftShoulderCapped, n.IsFlipped, n.Subcomponents.String())
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

// NestedSubcomponents represents the nested subcomponents element of the XML document
type NestedSubcomponents struct {
	XMLName       xml.Name      `xml:"subcomponents"`
	CenteringRing CenteringRing `xml:"centeringring"`
	MassComponent MassComponent `xml:"masscomponent"`
}

// String returns full string representation of the NestedSubcomponents
func (n *NestedSubcomponents) String() string {
	return fmt.Sprintf("NestedSubcomponents{CenteringRing=%s, MassComponent=%s}", n.CenteringRing.String(), n.MassComponent.String())
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

// BodyTube represents the body tube element of the XML document
type BodyTube struct {
	XMLName   xml.Name `xml:"bodytube"`
	Name      string   `xml:"name"`
	ID        string   `xml:"id"`
	Finish    string   `xml:"finish"`
	Material  Material `xml:"material"`
	Length    float64  `xml:"length"`
	Thickness float64  `xml:"thickness"`
	Radius    string   `xml:"radius"` // WARN: May be 'auto' and num
}

// String returns full string representation of the BodyTube
func (b *BodyTube) String() string {
	return fmt.Sprintf("BodyTube{Name=%s, ID=%s, Finish=%s, Material=%s, Length=%.2f, Thickness=%.2f, Radius=%s}", b.Name, b.ID, b.Finish, b.Material.String(), b.Length, b.Thickness, b.Radius)
}
