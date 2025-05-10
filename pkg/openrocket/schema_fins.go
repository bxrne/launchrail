package openrocket

import (
	"encoding/xml"
	"fmt"
)

// TrapezoidFinset represents the trapezoid finset element of the XML document
type TrapezoidFinset struct {
	XMLName        xml.Name       `xml:"trapezoidfinset"`
	Name           string         `xml:"name"`
	ID             string         `xml:"id"`
	InstanceCount  int            `xml:"instancecount"`
	FinCount       int            `xml:"fincount"`
	RadiusOffset   RadiusOffset   `xml:"radiusoffset"`
	AngleOffset    AngleOffset    `xml:"angleoffset"`
	Rotation       float64        `xml:"rotation"`
	AxialOffset    AxialOffset    `xml:"axialoffset"`
	Position       Position       `xml:"position"`
	Finish         string         `xml:"finish"`
	Material       Material       `xml:"material"`
	Thickness      float64        `xml:"thickness"`
	CrossSection   string         `xml:"crosssection"`
	Cant           float64        `xml:"cant"`
	TabHeight      float64        `xml:"tabheight"`
	TabLength      float64        `xml:"tablength"`
	TabPositions   []TabPosition  `xml:"tabposition"`
	FilletRadius   float64        `xml:"filletradius"`
	FilletMaterial Material       `xml:"filletmaterial"`
	RootChord      float64        `xml:"rootchord"`
	TipChord       float64        `xml:"tipchord"`
	SweepLength    float64        `xml:"sweeplength"`
	Height         float64        `xml:"height"`
	Subcomponents  struct {
		Fillets []Fillet `xml:"fillet"`
	} `xml:"subcomponents"`
}

// Fillet represents a fillet subcomponent within a fin set.
type Fillet struct {
	XMLName     xml.Name    `xml:"fillet"`
	Name        string      `xml:"name"`
	ID          string      `xml:"id"`
	AxialOffset AxialOffset `xml:"axialoffset"`
	Position    Position    `xml:"position"`
	Length      float64     `xml:"length"`
	Radius      float64     `xml:"radius"`
	Material    Material    `xml:"material"` // Or could be a simpler surface material type
}

// String returns a string representation of the Fillet.
func (f *Fillet) String() string {
	return fmt.Sprintf("Fillet{Name='%s', ID='%s', Length=%.3f, Radius=%.3f, Material='%s'}", 
		f.Name, f.ID, f.Length, f.Radius, f.Material.Name)
}

// GetMass calculates mass based on material density and dimensions
func (t *TrapezoidFinset) GetMass() float64 {
	area := (t.RootChord + t.TipChord) * t.Height / 2
	volume := area * t.Thickness
	density := t.Material.Density
	return volume * density
}

// String returns full string representation of the TrapezoidFinset
func (t *TrapezoidFinset) String() string {
	var tabPosition string
	for i, tp := range t.TabPositions {
		tabPosition += tp.String()
		if i < len(t.TabPositions)-1 {
			tabPosition += ", "
		}
	}

	var fillets string
	for i, f := range t.Subcomponents.Fillets {
		fillets += f.String()
		if i < len(t.Subcomponents.Fillets)-1 {
			fillets += ", "
		}
	}

	return fmt.Sprintf("TrapezoidFinset{Name=%s, ID=%s, InstanceCount=%d, FinCount=%d, RadiusOffset=%s, AngleOffset=%s, Rotation=%.2f, AxialOffset=%s, Position=%s, Finish=%s, Material=%s, Thickness=%.2f, CrossSection=%s, Cant=%.2f, TabHeight=%.2f, TabLength=%.2f, TabPositions=(%s), FilletRadius=%.2f, FilletMaterial=%s, RootChord=%.2f, TipChord=%.2f, SweepLength=%.2f, Height=%.2f, Subcomponents={Fillets=(%s)}}", t.Name, t.ID, t.InstanceCount, t.FinCount, t.RadiusOffset.String(), t.AngleOffset.String(), t.Rotation, t.AxialOffset.String(), t.Position.String(), t.Finish, t.Material.String(), t.Thickness, t.CrossSection, t.Cant, t.TabHeight, t.TabLength, tabPosition, t.FilletRadius, t.FilletMaterial.Name, t.RootChord, t.TipChord, t.SweepLength, t.Height, fillets)
}

// TabPosition represents the tabposition element of the XML document
type TabPosition struct {
	RelativeTo string  `xml:"relativeto,attr"`
	Value      float64 `xml:",chardata"`
}

// String returns full string representation of the tabposition
func (t *TabPosition) String() string {
	return fmt.Sprintf("TabPosition{RelativeTo=%s, Value=%.2f}", t.RelativeTo, t.Value)
}
