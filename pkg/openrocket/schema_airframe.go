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
	XMLName         xml.Name        `xml:"subcomponents"`
	InnerTube       InnerTube       `xml:"innertube"`
	TrapezoidFinset TrapezoidFinset `xml:"trapezoidfinset"`
	Parachute       Parachute       `xml:"parachute"`
}

// String returns full string representation of the BodyTubeSubcomponents
func (b *BodyTubeSubcomponents) String() string {
	return fmt.Sprintf("BodyTubeSubcomponents{InnerTube=%s, TrapezoidFinset=%s, Parachute=%s}", b.InnerTube.String(), b.TrapezoidFinset.String(), b.Parachute.String())
}

// InnerTube represents the inner tube element of the XML document
type InnerTube struct {
	XMLName              xml.Name          `xml:"innertube"`
	Name                 string            `xml:"name"`
	ID                   string            `xml:"id"`
	AxialOffset          AxialOffset       `xml:"axialoffset"`
	Position             Position          `xml:"position"`
	Material             Material          `xml:"material"`
	Length               float64           `xml:"length"`
	RadialPosition       float64           `xml:"radialposition"`
	RadialDirection      float64           `xml:"radialdirection"`
	OuterRadius          float64           `xml:"outerradius"`
	Thickness            float64           `xml:"thickness"`
	ClusterConfiguration string            `xml:"clusterconfiguration"`
	ClusterScale         float64           `xml:"clusterscale"`
	ClusterRotation      float64           `xml:"clusterrotation"`
	MotorMount           MotorMount        `xml:"motormount"`
	Subcomponents        NoseSubcomponents `xml:"subcomponents"` // TODO: Refactor naming here
}

// String returns full string representation of the innertube
func (i *InnerTube) String() string {
	return fmt.Sprintf("InnerTube{Name=%s, ID=%s, AxialOffset=%s, Position=%s, Material=%s, Length=%.2f, RadialPosition=%.2f, RadialDirection=%.2f, OuterRadius=%.2f, Thickness=%.2f, ClusterConfiguration=%s, ClusterScale=%.2f, ClusterRotation=%.2f, MotorMount=%s, Subcomponents=%s}", i.Name, i.ID, i.AxialOffset.String(), i.Position.String(), i.Material.String(), i.Length, i.RadialPosition, i.RadialDirection, i.OuterRadius, i.Thickness, i.ClusterConfiguration, i.ClusterScale, i.ClusterRotation, i.MotorMount.String(), i.Subcomponents.String())
}
