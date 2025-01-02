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
	FilletMaterial FilletMaterial `xml:"filletmaterial"`
	RootChord      float64        `xml:"rootchord"`
	TipChord       float64        `xml:"tipchord"`
	SweepLength    float64        `xml:"sweeplength"`
	Height         float64        `xml:"height"`
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

	return fmt.Sprintf("TrapezoidFinset{Name=%s, ID=%s, InstanceCount=%d, FinCount=%d, RadiusOffset=%s, AngleOffset=%s, Rotation=%.2f, AxialOffset=%s, Position=%s, Finish=%s, Material=%s, Thickness=%.2f, CrossSection=%s, Cant=%.2f, TabHeight=%.2f, TabLength=%.2f, TabPositions=(%s), FilletRadius=%.2f, RootChord=%.2f, TipChord=%.2f, SweepLength=%.2f, Height=%.2f}", t.Name, t.ID, t.InstanceCount, t.FinCount, t.RadiusOffset.String(), t.AngleOffset.String(), t.Rotation, t.AxialOffset.String(), t.Position.String(), t.Finish, t.Material.String(), t.Thickness, t.CrossSection, t.Cant, t.TabHeight, t.TabLength, tabPosition, t.FilletRadius, t.RootChord, t.TipChord, t.SweepLength, t.Height)
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
