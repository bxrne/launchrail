package openrocket_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/openrocket"
)

// TEST: GIVEN a DeploymentConfig struct WHEN calling the String method THEN return a string representation of the DeploymentConfig struct
func TestSchemaDeploymentConfigString(t *testing.T) {
	dc := &openrocket.DeploymentConfig{
		ConfigID:       "config",
		DeployEvent:    "event",
		DeployAltitude: 1.0,
		DeployDelay:    1.0,
	}

	expected := "DeploymentConfig{ConfigID=config, DeployEvent=event, DeployAltitude=1.00, DeployDelay=1.00}"
	if dc.String() != expected {
		t.Errorf("Expected %s, got %s", expected, dc.String())
	}
}

// TEST: GIVEN a Parachute struct WHEN calling the String method THEN return a string representation of the Parachute struct
func TestSchemaParachuteString(t *testing.T) {
	p := &openrocket.Parachute{
		Name:             "name",
		ID:               "id",
		AxialOffset:      openrocket.AxialOffset{},
		Position:         openrocket.Position{},
		PackedLength:     0.0,
		PackedRadius:     0.0,
		RadialPosition:   0.0,
		RadialDirection:  0.0,
		CD:               "0.0",
		Material:         openrocket.Material{},
		DeployEvent:      "event",
		DeployAltitude:   0.0,
		DeployDelay:      0.0,
		DeploymentConfig: openrocket.DeploymentConfig{},
		Diameter:         0.0,
		LineCount:        0,
		LineLength:       0.0,
		LineMaterial:     openrocket.LineMaterial{},
	}

	expected := "Parachute{Name=name, ID=id, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, CD=0.0, Material=Material{Type=, Density=0.00, Name=}, DeployEvent=event, DeployAltitude=0.00, DeployDelay=0.00, DeploymentConfig=DeploymentConfig{ConfigID=, DeployEvent=, DeployAltitude=0.00, DeployDelay=0.00}, Diameter=0.00, LineCount=0, LineLength=0.00, LineMaterial=LineMaterial{Type=, Density=0.00, Name=}}"
	if p.String() != expected {
		t.Errorf("Expected %s, got %s", expected, p.String())
	}
}
