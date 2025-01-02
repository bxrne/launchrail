package openrocket

import (
	"encoding/xml"
	"fmt"
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
