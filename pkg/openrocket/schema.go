package openrocket

import (
	"encoding/xml"
	"fmt"
)

// OpenrocketDocument represents the root of the XML document
type OpenrocketDocument struct {
	XMLName xml.Name       `xml:"openrocket"`
	Version string         `xml:"version,attr"`
	Creator string         `xml:"creator,attr"`
	Rocket  RocketDocument `xml:"rocket"`
}

// Describe returns a string representation of the OpenrocketDocument
func (o *OpenrocketDocument) Describe() string {
	return fmt.Sprintf("Version=%s, Creator=%s, Rocket=%s", o.Version, o.Creator, o.Rocket.Name)
}

// String returns full string representation of the OpenrocketDocument
func (o *OpenrocketDocument) String() string {
	return fmt.Sprintf("OpenrocketDocument{Version=%s, Creator=%s, Rocket=%s}", o.Version, o.Creator, o.Rocket.String())
}

// RocketDocument represents the rocket element of the XML document
type RocketDocument struct {
	XMLName            xml.Name           `xml:"rocket"`
	Name               string             `xml:"name"`
	ID                 string             `xml:"id"`
	AxialOffset        AxialOffset        `xml:"axialoffset"`
	Position           Position           `xml:"position"`
	Designer           string             `xml:"designer"`
	Revision           string             `xml:"revision"`
	MotorConfiguration MotorConfiguration `xml:"motorconfiguration"`
	ReferenceType      string             `xml:"referencetype"`
}

// String returns full string representation of the RocketDocument
func (r *RocketDocument) String() string {
	return fmt.Sprintf("RocketDocument{Name=%s, ID=%s, AxialOffset=%s, Position=%s, Designer=%s, Revision=%s, MotorConfiguration=%s, ReferenceType=%s}", r.Name, r.ID, r.AxialOffset.String(), r.Position.String(), r.Designer, r.Revision, r.MotorConfiguration.String(), r.ReferenceType)
}

// AxialOffset represents the axial offset element of the XML document
type AxialOffset struct {
	XMLName xml.Name `xml:"axialoffset"`
	Method  string   `xml:"method,attr"`
	Value   float64  `xml:",chardata"`
}

// String returns full string representation of the AxialOffset
func (a *AxialOffset) String() string {
	return fmt.Sprintf("AxialOffset{Method=%s, Value=%f}", a.Method, a.Value)
}

// Position represents the position element of the XML document
type Position struct {
	XMLName xml.Name `xml:"position"`
	Value   float64  `xml:",chardata"`
	Type    string   `xml:"type,attr"`
}

// String returns full string representation of the Position
func (p *Position) String() string {
	return fmt.Sprintf("Position{Value=%f, Type=%s}", p.Value, p.Type)
}

// Stage represents motor configuration stages
type Stage struct {
	XMLName xml.Name `xml:"stage"`
	Number  int      `xml:"number,attr"`
	Active  bool     `xml:"active,attr"`
}

// String returns full string representation of the Stage
func (s *Stage) String() string {
	return fmt.Sprintf("Stage{Number=%d, Active=%t}", s.Number, s.Active)
}

// MotorConfiguration represents the motor configuration element of the XML document
type MotorConfiguration struct {
	XMLName  xml.Name `xml:"motorconfiguration"`
	ConfigID string   `xml:"configid,attr"`
	Default  bool     `xml:"default,attr"`
	Stages   []Stage  `xml:"stage"`
}

// String returns full string representation of the MotorConfiguration
func (m *MotorConfiguration) String() string {
	return fmt.Sprintf("MotorConfiguration{ConfigID=%s, Default=%t, Stages=%v}", m.ConfigID, m.Default, m.Stages)
}
