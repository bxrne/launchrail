package openrocket_test

import (
	"github.com/bxrne/launchrail/pkg/openrocket"
	"testing"
)

// TEST: GIVEN a Nosecone struct WHEN calling the String method THEN return a string representation of the Nosecone struct
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
		Subcomponents:        openrocket.NoseSubcomponents{},
	}

	expected := "Nosecone{Name=name, ID=id, Finish=finish, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Shape=shape, ShapeClipped=false, ShapeParameter=0.00, AftRadius=0.00, AftShoulderRadius=0.00, AftShoulderLength=0.00, AftShoulderThickness=0.00, AftShoulderCapped=false, IsFlipped=false, Subcomponents=NestedSubcomponents{CenteringRing=CenteringRing{Name=, ID=, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=, InnerRadius=}, MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}}"
	if nc.String() != expected {
		t.Errorf("Expected %s, got %s", expected, nc.String())
	}
}

// TEST: GIVEN a NoseSubcomponents struct WHEN calling the String method THEN return a string representation of the NoseSubcomponents struct
func TestSchemaNoseSubcomponentsString(t *testing.T) {
	ns := &openrocket.NoseSubcomponents{
		CenteringRing: openrocket.CenteringRing{},
		MassComponent: openrocket.MassComponent{},
	}

	expected := "NestedSubcomponents{CenteringRing=CenteringRing{Name=, ID=, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=, InnerRadius=}, MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}"
	if ns.String() != expected {
		t.Errorf("Expected %s, got %s", expected, ns.String())
	}
}
