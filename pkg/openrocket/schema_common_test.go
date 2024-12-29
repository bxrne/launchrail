package openrocket_test

import (
	"github.com/bxrne/launchrail/pkg/openrocket"
	"testing"
)

// TEST: GIVEN an AxialOffset struct WHEN calling the String method THEN return a string representation of the AxialOffset struct
func TestSchemaAxialOffsetString(t *testing.T) {
	ao := openrocket.AxialOffset{
		Method: "absolute",
		Value:  0.0,
	}
	expected := "AxialOffset{Method=absolute, Value=0.00}"
	if ao.String() != expected {
		t.Errorf("Expected %s, got %s", expected, ao.String())
	}
}

// TEST: GIVEN a Position struct WHEN calling the String method THEN return a string representation of the Position struct
func TestSchemaPositionString(t *testing.T) {
	p := openrocket.Position{
		Value: 0.0,
		Type:  "absolute",
	}
	expected := "Position{Value=0.00, Type=absolute}"
	if p.String() != expected {
		t.Errorf("Expected %s, got %s", expected, p.String())
	}
}

// TEST: GIVEN a Material struct WHEN calling the String method THEN return a string representation of the Material struct
func TestSchemaMaterialString(t *testing.T) {
	m := openrocket.Material{
		Type:    "type",
		Density: 0.0,
		Name:    "name",
	}
	expected := "Material{Type=type, Density=0.00, Name=name}"
	if m.String() != expected {
		t.Errorf("Expected %s, got %s", expected, m.String())
	}
}

// TEST: GIVEN a CenteringRing struct WHEN calling the String method THEN return a string representation of the CenteringRing struct
func TestSchemaCenteringRingString(t *testing.T) {
	cr := openrocket.CenteringRing{
		Name:               "name",
		ID:                 "id",
		InstanceCount:      0,
		InstanceSeparation: 0.0,
		AxialOffset:        openrocket.AxialOffset{},
		Position:           openrocket.Position{},
		Material:           openrocket.Material{},
		Length:             0.0,
		RadialPosition:     0.0,
		RadialDirection:    0.0,
		OuterRadius:        "auto",
		InnerRadius:        "auto",
	}
	expected := "CenteringRing{Name=name, ID=id, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=auto, InnerRadius=auto}"
	if cr.String() != expected {
		t.Errorf("Expected %s, got %s", expected, cr.String())
	}
}

// TEST: GIVEN a MassComponent struct WHEN calling the String method THEN return a string representation of the MassComponent struct
func TestSchemaMassComponentString(t *testing.T) {
	mc := openrocket.MassComponent{
		Name:            "name",
		ID:              "id",
		AxialOffset:     openrocket.AxialOffset{},
		Position:        openrocket.Position{},
		PackedLength:    0.0,
		PackedRadius:    0.0,
		RadialPosition:  0.0,
		RadialDirection: 0.0,
		Mass:            0.0,
		Type:            "type",
	}

	expected := "MassComponent{Name=name, ID=id, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=type}"
	if mc.String() != expected {
		t.Errorf("Expected %s, got %s", expected, mc.String())
	}

}
