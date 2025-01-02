package openrocket

import (
	"encoding/xml"
	"fmt"
)

// MotorMount represents the motor mount element of the XML document
type MotorMount struct {
	XMLName        xml.Name       `xml:"motormount"`
	IgnitionEvent  string         `xml:"ignitionevent"`
	IgnitionDelay  float64        `xml:"ignitiondelay"`
	Overhang       float64        `xml:"overhang"`
	Motor          Motor          `xml:"motor"`
	IgnitionConfig IgnitionConfig `xml:"ignitionconfiguration"` // WARN: This duplicates field above but let's keep it for now
}

// String returns full string representation of the motormount
func (m *MotorMount) String() string {
	return fmt.Sprintf("MotorMount{IgnitionEvent=%s, IgnitionDelay=%.2f, Overhang=%.2f, Motor=%s, IgnitionConfig=%s}", m.IgnitionEvent, m.IgnitionDelay, m.Overhang, m.Motor.String(), m.IgnitionConfig.String())
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

// IgnitionConfig represents the ignition configuration element of the XML document
type IgnitionConfig struct {
	XMLName       xml.Name `xml:"ignitionconfiguration"`
	ConfigID      string   `xml:"configid,attr"`
	IgnitionEvent string   `xml:"ignitionevent"`
	IgnitionDelay float64  `xml:"ignitiondelay"`
}

// String returns full string representation of the ignition clusterconfiguration
func (i *IgnitionConfig) String() string {
	return fmt.Sprintf("IgnitionConfig{ConfigID=%s, IgnitionEvent=%s, IgnitionDelay=%.2f}", i.ConfigID, i.IgnitionEvent, i.IgnitionDelay)
}
