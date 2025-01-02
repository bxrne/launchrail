package openrocket

import (
	"encoding/xml"
	"fmt"
)

// Material represents the material element of the XML document
type Material struct {
	XMLName xml.Name `xml:"material"`
	Type    string   `xml:"type,attr"`
	Density float64  `xml:"density,attr"`
	Name    string   `xml:",chardata"`
}

// String returns full string representation of the Material
func (m *Material) String() string {
	return fmt.Sprintf("Material{Type=%s, Density=%.2f, Name=%s}", m.Type, m.Density, m.Name)
}

// FilletMaterial represents the fillet material element of the XML document (XMLName is the only delta from Material)
type FilletMaterial struct {
	XMLName xml.Name `xml:"filletmaterial"`
	Type    string   `xml:"type,attr"`
	Density float64  `xml:"density,attr"`
	Name    string   `xml:",chardata"`
}

// String returns full string representation of the filletmaterial
func (f *FilletMaterial) String() string {
	return fmt.Sprintf("FilletMaterial{Type=%s, Density=%.2f, Name=%s}", f.Type, f.Density, f.Name)
}

// LineMaterial is the same as Material with a diff XMLName
type LineMaterial struct {
	XMLName xml.Name `xml:"linematerial"`
	Type    string   `xml:"type,attr"`
	Density float64  `xml:"density,attr"`
	Name    string   `xml:",chardata"`
}

// String returns full string representation of the LineMaterial
func (l *LineMaterial) String() string {
	return fmt.Sprintf("LineMaterial{Type=%s, Density=%.2f, Name=%s}", l.Type, l.Density, l.Name)
}
