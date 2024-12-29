package openrocket_test

import (
	"github.com/bxrne/launchrail/pkg/openrocket"
	"testing"
)

// TEST: GIVEN a NestedSubcomponents struct WHEN calling the String method THEN return a string representation of the NestedSubcomponents struct
func TestSchemaNestedSubcomponentsString(t *testing.T) {
	ns := openrocket.NestedSubcomponents{
		CenteringRing: openrocket.CenteringRing{},
		MassComponent: openrocket.MassComponent{},
	}

	expected := "NestedSubcomponents{CenteringRing=CenteringRing{Name=, ID=, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=, InnerRadius=}, MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}"
	if ns.String() != expected {
		t.Errorf("Expected %s, got %s", expected, ns.String())
	}
}

// TEST: GIVEN a SustainerSubcomponents struct WHEN calling the String method THEN return a string representation of the SustainerSubcomponents struct
func TestSchemaSustainerSubcomponentsString(t *testing.T) {
	ss := openrocket.SustainerSubcomponents{
		Nosecone: openrocket.Nosecone{},
		BodyTube: openrocket.BodyTube{},
	}

	expected := "SustainerSubcomponents{Nosecone=Nosecone{Name=, ID=, Finish=, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Shape=, ShapeClipped=false, ShapeParameter=0.00, AftRadius=0.00, AftShoulderRadius=0.00, AftShoulderLength=0.00, AftShoulderThickness=0.00, AftShoulderCapped=false, IsFlipped=false, Subcomponents=NestedSubcomponents{CenteringRing=CenteringRing{Name=, ID=, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=, InnerRadius=}, MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}}, BodyTube=BodyTube{Name=, ID=, Finish=, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Radius=}}"
	if ss.String() != expected {
		t.Errorf("Expected %s, got %s", expected, ss.String())
	}
}

// TEST: GIVEN a Nosecone struct WHEN calling the String method THEN return a string representation of the Nosecone struct
func TestSchemaNoseconeString(t *testing.T) {
	n := openrocket.Nosecone{
		Name:                 "name",
		ID:                   "id",
		Finish:               "finish",
		Material:             openrocket.Material{},
		Length:               0.0,
		Thickness:            0.0,
		Shape:                "shape",
		ShapeClipped:         false,
		ShapeParameter:       0.0,
		AftRadius:            0.0,
		AftShoulderRadius:    0.0,
		AftShoulderLength:    0.0,
		AftShoulderThickness: 0.0,
		AftShoulderCapped:    false,
		IsFlipped:            false,
		Subcomponents:        openrocket.NestedSubcomponents{},
	}

	expected := "Nosecone{Name=name, ID=id, Finish=finish, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Shape=shape, ShapeClipped=false, ShapeParameter=0.00, AftRadius=0.00, AftShoulderRadius=0.00, AftShoulderLength=0.00, AftShoulderThickness=0.00, AftShoulderCapped=false, IsFlipped=false, Subcomponents=NestedSubcomponents{CenteringRing=CenteringRing{Name=, ID=, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=, InnerRadius=}, MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}}"
	if n.String() != expected {
		t.Errorf("Expected %s, got %s", expected, n.String())
	}
}

// TEST: GIVEN a BodyTube struct WHEN calling the String method THEN return a string representation of the BodyTube struct
func TestSchemaBodyTubeString(t *testing.T) {
	bt := openrocket.BodyTube{
		Name:      "name",
		ID:        "id",
		Finish:    "finish",
		Material:  openrocket.Material{},
		Length:    0.0,
		Thickness: 0.0,
		Radius:    "auto 0.0",
	}

	expected := "BodyTube{Name=name, ID=id, Finish=finish, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Radius=auto 0.0}"
	if bt.String() != expected {
		t.Errorf("Expected %s, got %s", expected, bt.String())
	}
}
