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

// AxialOffset represents the axial offset element of the XML document
type AxialOffset struct {
	XMLName xml.Name `xml:"axialoffset"`
	Method  string   `xml:"method,attr"`
	Value   float64  `xml:",chardata"`
}

// Position represents the position element of the XML document
type Position struct {
	XMLName xml.Name `xml:"position"`
	Value   float64  `xml:",chardata"`
	Type    string   `xml:"type,attr"`
}

// Stage represents motor configuration stages
type Stage struct {
	XMLName xml.Name `xml:"stage"`
	Number  int      `xml:"number,attr"`
	Active  bool     `xml:"active,attr"`
}

// MotorConfiguration represents the motor configuration element of the XML document
type MotorConfiguration struct {
	XMLName  xml.Name `xml:"motorconfiguration"`
	ConfigID string   `xml:"configid,attr"`
	Default  bool     `xml:"default,attr"`
	Stages   []Stage  `xml:"stage"`
}
