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
