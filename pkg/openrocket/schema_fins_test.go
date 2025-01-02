package openrocket_test

import (
	"github.com/bxrne/launchrail/pkg/openrocket"
	"testing"
)

// TEST: GIVEN a TrapezoidFinset struct WHEN calling the String method THEN return a string representation of the TrapezoidFinset struct
func TestSchemaTrapezoidFinsetString(t *testing.T) {
	tf := &openrocket.TrapezoidFinset{
		Name:          "name",
		ID:            "id",
		InstanceCount: 0,
		FinCount:      0,
		RadiusOffset:  openrocket.RadiusOffset{},
		AngleOffset:   openrocket.AngleOffset{},
		Rotation:      0.0,
		AxialOffset:   openrocket.AxialOffset{},
		Position:      openrocket.Position{},
		Finish:        "finish",
		Material:      openrocket.Material{},
		Thickness:     0.0,
		CrossSection:  "cross",
		Cant:          0.0,
		TabHeight:     0.0,
		TabLength:     0.0,
		TabPositions:  []openrocket.TabPosition{},
		FilletRadius:  0.0,
		RootChord:     0.0,
		TipChord:      0.0,
		SweepLength:   0.0,
		Height:        0.0,
	}

	expected := "TrapezoidFinset{Name=name, ID=id, InstanceCount=0, FinCount=0, RadiusOffset=RadiusOffset{Method=, Value=0.00}, AngleOffset=AngleOffset{Method=, Value=0.00}, Rotation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Finish=finish, Material=Material{Type=, Density=0.00, Name=}, Thickness=0.00, CrossSection=cross, Cant=0.00, TabHeight=0.00, TabLength=0.00, TabPositions=(), FilletRadius=0.00, RootChord=0.00, TipChord=0.00, SweepLength=0.00, Height=0.00}"
	if tf.String() != expected {
		t.Errorf("Expected %s, got %s", expected, tf.String())
	}
}

// TEST: GIVEN a TabPosition struct WHEN calling the String method THEN return a string representation of the TabPosition struct
func TestSchemaTabPositionString(t *testing.T) {
	tp := &openrocket.TabPosition{
		RelativeTo: "relative",
		Value:      0.0,
	}

	expected := "TabPosition{RelativeTo=relative, Value=0.00}"
	if tp.String() != expected {
		t.Errorf("Expected %s, got %s", expected, tp.String())
	}
}
