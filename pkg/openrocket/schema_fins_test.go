package openrocket_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/openrocket"
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
		TabPositions: []openrocket.TabPosition{
			{RelativeTo: "RootChord", Value: 0.01},
			{RelativeTo: "TipChord", Value: 0.02},
		},
		FilletRadius: 0.0,
		FilletMaterial: openrocket.FilletMaterial{
			Type:    "",
			Density: 0.0,
			Name:    "",
		},
		RootChord:   0.0,
		TipChord:    0.0,
		SweepLength: 0.0,
		Height:      0.0,
	}

	expected := "TrapezoidFinset{Name=name, ID=id, InstanceCount=0, FinCount=0, RadiusOffset=RadiusOffset{Method=, Value=0.00}, AngleOffset=AngleOffset{Method=, Value=0.00}, Rotation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Finish=finish, Material=Material{Type=, Density=0.00, Name=}, Thickness=0.00, CrossSection=cross, Cant=0.00, TabHeight=0.00, TabLength=0.00, TabPositions=(TabPosition{RelativeTo=RootChord, Value=0.01}, TabPosition{RelativeTo=TipChord, Value=0.02}), FilletRadius=0.00, FilletMaterial=FilletMaterial{Type=, Density=0.00, Name=}, RootChord=0.00, TipChord=0.00, SweepLength=0.00, Height=0.00, Subcomponents={Fillets=()}}"
	if tf.String() != expected {
		t.Errorf("Expected %s, got %s", expected, tf.String())
	}
}

// TEST: GIVEN a TrapezoidFinset struct with dimensions and density WHEN calling the GetMass method THEN return the calculated mass
func TestSchemaTrapezoidFinsetGetMass(t *testing.T) {
	tf := &openrocket.TrapezoidFinset{
		RootChord: 0.1,   // 10 cm
		TipChord:  0.05,  // 5 cm
		Height:    0.15,  // 15 cm
		Thickness: 0.003, // 3 mm
		Material: openrocket.Material{
			Density: 1200, // Example density kg/m^3
		},
	}

	// Expected calculation: area = (0.1 + 0.05) * 0.15 / 2 = 0.01125 m^2
	// volume = area * thickness = 0.01125 * 0.003 = 0.00003375 m^3
	// mass = volume * density = 0.00003375 * 1200 = 0.0405 kg
	expectedMass := 0.0405
	calculatedMass := tf.GetMass()

	// Use a small tolerance for floating point comparison
	tolerance := 1e-9
	if calculatedMass < expectedMass-tolerance || calculatedMass > expectedMass+tolerance {
		t.Errorf("Expected mass %f, got %f", expectedMass, calculatedMass)
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

// TEST: GIVEN a Fillet struct WHEN calling the String method THEN return a string representation of the Fillet struct
func TestSchemaFilletString(t *testing.T) {
	fillet := &openrocket.Fillet{
		Name:   "testFillet",
		ID:     "fillet1",
		Length: 0.05,
		Radius: 0.025,
		Material: openrocket.Material{
			Name:    "CardboardMaterial",
			Type:    "BULK",
			Density: 600.0,
		},
	}

	expected := "Fillet{Name='testFillet', ID='fillet1', Length=0.050, Radius=0.025, Material='CardboardMaterial'}"
	if fillet.String() != expected {
		t.Errorf("Expected %s, got %s", expected, fillet.String())
	}
}
