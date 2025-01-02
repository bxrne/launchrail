package openrocket_test

import (
	"github.com/bxrne/launchrail/pkg/openrocket"
	"testing"
)

// TEST: GIVEN a RadiusOffset struct WHEN calling the String method THEN return a string representation of the RadiusOffset struct
func TestSchemaRadiusOffsetString(t *testing.T) {
	r := &openrocket.RadiusOffset{
		Method: "method1",
		Value:  1.0,
	}

	expected := "RadiusOffset{Method=method1, Value=1.00}"
	if r.String() != expected {
		t.Errorf("Expected %s, got %s", expected, r.String())
	}
}

// TEST: GIVEN a AngleOffset struct WHEN calling the String method THEN return a string representation of the AngleOffset struct
func TestSchemaAngleOffsetString(t *testing.T) {
	a := &openrocket.AngleOffset{
		Method: "method1",
		Value:  1.0,
	}

	expected := "AngleOffset{Method=method1, Value=1.00}"
	if a.String() != expected {
		t.Errorf("Expected %s, got %s", expected, a.String())
	}
}

// TEST: GIVEN a AxialOffset struct WHEN calling the String method THEN return a string representation of the AxialOffset struct
func TestSchemaAxialOffsetString(t *testing.T) {
	a := &openrocket.AxialOffset{
		Method: "method1",
		Value:  1.0,
	}

	expected := "AxialOffset{Method=method1, Value=1.00}"
	if a.String() != expected {
		t.Errorf("Expected %s, got %s", expected, a.String())
	}
}

// TEST: GIVEN a Position struct WHEN calling the String method THEN return a string representation of the Position struct
func TestSchemaPositionString(t *testing.T) {
	p := &openrocket.Position{
		Value: 1.0,
		Type:  "type1",
	}

	expected := "Position{Value=1.00, Type=type1}"
	if p.String() != expected {
		t.Errorf("Expected %s, got %s", expected, p.String())
	}
}

// TEST: GIVEN a CenteringRing struct WHEN calling the String method THEN return a string representation of the CenteringRing struct
func TestSchemaCenteringRingString(t *testing.T) {
	c := &openrocket.CenteringRing{
		Name:               "name",
		ID:                 "id",
		InstanceCount:      1,
		InstanceSeparation: 1.0,
		AxialOffset:        openrocket.AxialOffset{},
		Position:           openrocket.Position{},
		Material:           openrocket.Material{},
		Length:             1.0,
		RadialPosition:     1.0,
		OuterRadius:        "auto",
		InnerRadius:        "auto",
	}

	expected := "CenteringRing{Name=name, ID=id, InstanceCount=1, InstanceSeparation=1.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=1.00, RadialPosition=1.00, OuterRadius=auto, InnerRadius=auto}"
	if c.String() != expected {
		t.Errorf("Expected %s, got %s", expected, c.String())
	}
}

// TEST: GIVEN a MassComponent struct WHEN calling the String method THEN return a string representation of the MassComponent struct
func TestSchemaMassComponentString(t *testing.T) {
	m := &openrocket.MassComponent{
		Name:            "name",
		ID:              "id",
		AxialOffset:     openrocket.AxialOffset{},
		Position:        openrocket.Position{},
		PackedLength:    1.0,
		PackedRadius:    1.0,
		RadialPosition:  1.0,
		RadialDirection: 1.0,
		Mass:            1.0,
		Type:            "type1",
	}

	expected := "MassComponent{Name=name, ID=id, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=1.00, PackedRadius=1.00, RadialPosition=1.00, RadialDirection=1.00, Mass=1.00, Type=type1}"
	if m.String() != expected {
		t.Errorf("Expected %s, got %s", expected, m.String())
	}
}
