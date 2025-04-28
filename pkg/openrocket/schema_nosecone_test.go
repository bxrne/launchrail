package openrocket_test

import (
	"encoding/xml"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"testing"
)

func TestNosecone_GetMass_Conical(t *testing.T) {
	nc := &openrocket.Nosecone{
		Length:    10,
		AftRadius: 2,
		Thickness: 0.2,
		Shape:     "conical",
		Material:  openrocket.Material{Density: 1.0},
	}
	got := nc.GetMass()
	if got <= 0 {
		t.Errorf("Expected positive mass for conical, got %v", got)
	}
}

func TestNosecone_GetMass_Ogive(t *testing.T) {
	nc := &openrocket.Nosecone{
		Length:    10,
		AftRadius: 2,
		Thickness: 0.2,
		Shape:     "ogive",
		Material:  openrocket.Material{Density: 1.0},
	}
	got := nc.GetMass()
	if got <= 0 {
		t.Errorf("Expected positive mass for ogive, got %v", got)
	}
}

func TestNosecone_GetMass_Invalid(t *testing.T) {
	nc := &openrocket.Nosecone{
		Length:    0,
		AftRadius: 0,
		Thickness: 0,
		Shape:     "conical",
		Material:  openrocket.Material{Density: 1.0},
	}
	got := nc.GetMass()
	if got != 0 {
		t.Errorf("Expected zero mass for invalid dimensions, got %v", got)
	}
}

func TestSchemaNoseconeString(t *testing.T) {
	nc := &openrocket.Nosecone{
		Name:                 "name",
		ID:                   "id",
		Finish:               "finish",
		Material:             openrocket.Material{},
		Thickness:            0.0,
		Length:               0.0,
		Shape:                "shape",
		ShapeClipped:         false,
		ShapeParameter:       0.0,
		AftRadius:            0.0,
		AftShoulderRadius:    0.0,
		AftShoulderLength:    0.0,
		AftShoulderThickness: 0.0,
		AftShoulderCapped:    false,
		IsFlipped:            false,
		Subcomponents: openrocket.NoseSubcomponents{
			XMLName: xml.Name{Local: "subcomponents"},
			MassComponent: openrocket.MassComponent{
				XMLName:     xml.Name{Local: "masscomponent"},
				Name:        "Payload",
				ID:          "masscomp-1",
				Mass:        0.1,
				Type:        "part",
			},
		},
	}

	expected := "Nosecone{Name=name, ID=id, Finish=finish, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Shape=shape, ShapeClipped=false, ShapeParameter=0.00, AftRadius=0.00, AftShoulderRadius=0.00, AftShoulderLength=0.00, AftShoulderThickness=0.00, AftShoulderCapped=false, IsFlipped=false, Subcomponents=NoseSubcomponents{MassComponent=MassComponent{Name=Payload, ID=masscomp-1, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.10, Type=part}}}"
	if nc.String() != expected {
		t.Errorf("Expected %s, got %s", expected, nc.String())
	}
}

func TestSchemaNoseSubcomponentsString(t *testing.T) {
	ns := &openrocket.NoseSubcomponents{
		XMLName: xml.Name{Local: "subcomponents"},
		MassComponent: openrocket.MassComponent{
			XMLName: xml.Name{Local: "masscomponent"},
		},
	}

	expected := "NoseSubcomponents{MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}"
	if ns.String() != expected {
		t.Errorf("Expected %s, got %s", expected, ns.String())
	}
}
