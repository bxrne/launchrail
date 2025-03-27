package openrocket

import (
	"encoding/xml"
	"fmt"

	"github.com/bxrne/launchrail/internal/config"
)

// TODO: Omit empty fields

// OpenrocketDocument represents the root of the XML document
type OpenrocketDocument struct {
	XMLName xml.Name       `xml:"openrocket"`
	Version string         `xml:"version,attr"`
	Creator string         `xml:"creator,attr"`
	Rocket  RocketDocument `xml:"rocket"`
}

// Bytes returns the byte representation of the OpenrocketDocument
func (o *OpenrocketDocument) Bytes() []byte {
	return []byte(fmt.Sprintf("%+v", o))
}

// Validate checks if the OpenrocketDocument is valid for this program
func (o *OpenrocketDocument) Validate(cfg *config.Engine) error {
	// Check if the OpenRocket version matches the expected version
	if fmt.Sprintf("OpenRocket %s", cfg.External.OpenRocketVersion) != o.Creator {
		return fmt.Errorf("OpenRocket version mismatch: expected %s, got %s", cfg.External.OpenRocketVersion, o.Creator)
	}

	// Check subcomponents structure (must be single sustainer only)
	if len(o.Rocket.Subcomponents.Stages) != 1 {
		return fmt.Errorf("expected 1 subcomponent, got %d", len(o.Rocket.Subcomponents.Stages))
	}

	if o.Rocket.Subcomponents.Stages[0].Name != "Sustainer" {
		return fmt.Errorf("expected Sustainer, got %s", o.Rocket.Subcomponents.Stages[0].Name)
	}

	return nil
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
	Subcomponents      Subcomponents      `xml:"subcomponents"`
}

// String returns full string representation of the RocketDocument
func (r *RocketDocument) String() string {
	return fmt.Sprintf("RocketDocument{Name=%s, ID=%s, AxialOffset=%s, Position=%s, Designer=%s, Revision=%s, MotorConfiguration=%s, ReferenceType=%s, Subcomponents={%s}}", r.Name, r.ID, r.AxialOffset.String(), r.Position.String(), r.Designer, r.Revision, r.MotorConfiguration.String(), r.ReferenceType, r.Subcomponents.String())
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
	var stages string
	for i, stage := range m.Stages {
		stages += stage.String()
		if i < len(m.Stages)-1 {
			stages += ", "
		}
	}

	return fmt.Sprintf("MotorConfiguration{ConfigID=%s, Default=%t, Stages=(%s)}", m.ConfigID, m.Default, stages)
}

// Subcomponents represents the subcomponents element of the XML document
type Subcomponents struct {
	XMLName xml.Name      `xml:"subcomponents"`
	Stages  []RocketStage `xml:"stage"`
}

// String returns full string representation of the Subcomponents
func (s *Subcomponents) String() string {
	var stages string
	for i, stage := range s.Stages {
		stages += stage.String()
		if i < len(s.Stages)-1 {
			stages += ", "
		}
	}

	return fmt.Sprintf("Subcomponents{Stages=(%s)}", stages)
}

// List returns a list of all subcomponents so we can iterate over them
func (s *Subcomponents) List() []RocketStage {
	return s.Stages
}

// RocketStage represents the stage subcomponent element of the XML document
type RocketStage struct {
	XMLName                xml.Name               `xml:"stage"`
	Name                   string                 `xml:"name"`
	ID                     string                 `xml:"id"`
	SustainerSubcomponents SustainerSubcomponents `xml:"subcomponents"`
}

// String returns full string representation of the RocketStage
func (r *RocketStage) String() string {
	return fmt.Sprintf("RocketStage{Name=%s, ID=%s, SustainerSubcomponents=%s}", r.Name, r.ID, r.SustainerSubcomponents.String())
}
