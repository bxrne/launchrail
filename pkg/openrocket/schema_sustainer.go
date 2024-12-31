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

// String returns full string representation of the BodyTube
func (b *BodyTube) String() string {
	return fmt.Sprintf("BodyTube{Name=%s, ID=%s, Finish=%s, Material=%s, Length=%.2f, Thickness=%.2f, Radius=%s, Subcomponents=%s}", b.Name, b.ID, b.Finish, b.Material.String(), b.Length, b.Thickness, b.Radius, b.Subcomponents.String())

}

// BodyTubeSubcomponents represents the nested subcomponents element of the XML document
type BodyTubeSubcomponents struct {
	XMLName   xml.Name  `xml:"subcomponents"`
	InnerTube InnerTube `xml:"innertube"`
}

// String returns full string representation of the BodyTubeSubcomponents
func (b *BodyTubeSubcomponents) String() string {
	return fmt.Sprintf("NestedSubcomponents{InnerTube=%s}", b.InnerTube.String())
}

// InnerTube represents the inner tube element of the XML document
type InnerTube struct {
	XMLName              xml.Name    `xml:"innertube"`
	Name                 string      `xml:"name"`
	ID                   string      `xml:"id"`
	AxialOffset          AxialOffset `xml:"axialoffset"`
	Position             Position    `xml:"position"`
	Material             Material    `xml:"material"`
	Length               float64     `xml:"length"`
	RadialPosition       float64     `xml:"radialposition"`
	RadialDirection      float64     `xml:"radialdirection"`
	OuterRadius          float64     `xml:"outerradius"`
	Thickness            float64     `xml:"thickness"`
	ClusterConfiguration string      `xml:"clusterconfiguration"`
	ClusterScale         float64     `xml:"clusterscale"`
	ClusterRotation      float64     `xml:"clusterrotation"`
	MotorMount           MotorMount  `xml:"motormount"`
}

// String returns full string representation of the innertube
func (i *InnerTube) String() string {
	return fmt.Sprintf("InnerTube{Name=%s, ID=%s, AxialOffset=%s, Position=%s, Material=%s, Length=%.2f, RadialPosition=%.2f, RadialDirection=%.2f, OuterRadius=%.2f, Thickness=%.2f, ClusterConfiguration=%s, ClusterScale=%.2f, ClusterRotation=%.2f, MotorMount=%s}", i.Name, i.ID, i.AxialOffset.String(), i.Position.String(), i.Material.String(), i.Length, i.RadialPosition, i.RadialDirection, i.OuterRadius, i.Thickness, i.ClusterConfiguration, i.ClusterScale, i.ClusterRotation, i.MotorMount.String())
}

// MotorMount represents the motor mount element of the XML document
type MotorMount struct {
	XMLName       xml.Name `xml:"motormount"`
	IgnitionEvent string   `xml:"ignitionevent"`
	IgnitionDelay float64  `xml:"ignitiondelay"`
	Overhang      float64  `xml:"overhang"`
	Motor         Motor    `xml:"motor"`
}

// String returns full string representation of the motormount
func (m *MotorMount) String() string {
	return fmt.Sprintf("MotorMount{IgnitionEvent=%s, IgnitionDelay=%.2f, Overhang=%.2f, Motor=%s}", m.IgnitionEvent, m.IgnitionDelay, m.Overhang, m.Motor.String())
}

// Motor represents the motor element of the XML document
type Motor struct {
	XMLName      xml.Name `xml:"motor"`
	ConfigID     string   `xml:"configid,attr"`
	Type         string   `xml:"type"`
	Manufacturer string   `xml:"manufacturer"`
	Digest       string   `xml:"digest"`
	Designation  string   `xml:"designation"`
	Diameter     float64  `xml:"diameter"`
	Length       float64  `xml:"length"`
	Delay        string   `xml:"delay"`
}

// String returns full string representation of the motormount
func (m *Motor) String() string {
	return fmt.Sprintf("Motor{ConfigID=%s, Type=%s, Manufacturer=%s, Digest=%s, Designation=%s, Diameter=%.2f, Length=%.2f, Delay=%s}", m.ConfigID, m.Type, m.Manufacturer, m.Digest, m.Designation, m.Diameter, m.Length, m.Delay)
}
