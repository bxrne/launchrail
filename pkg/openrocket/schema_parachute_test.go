package openrocket_test

import (
	"math"
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

// TEST: GIVEN a Parachute struct WHEN calling the GetMass method THEN return the calculated mass
func TestParachuteGetMass(t *testing.T) {
	tests := []struct {
		name      string
		parachute *openrocket.Parachute
		wantMass  float64
	}{
		{
			name: "Valid Calculation",
			parachute: &openrocket.Parachute{
				Name:         "Test Chute",
				Diameter:     1.0,                                // 1m diameter
				Material:     openrocket.Material{Density: 0.05}, // 50g/m^2 areal density
				LineCount:    8,
				LineLength:   1.5,
				LineMaterial: openrocket.LineMaterial{Density: 0.002}, // 2g/m linear density
			},
			wantMass: (math.Pi * 0.5 * 0.5 * 0.05) + (8 * 1.5 * 0.002), // Canopy Area * Canopy Density + Lines * Line Length * Line Density
		},
		{
			name: "Zero Diameter",
			parachute: &openrocket.Parachute{
				Name:         "Zero Diameter Chute",
				Diameter:     0.0,
				Material:     openrocket.Material{Density: 0.05},
				LineCount:    8,
				LineLength:   1.5,
				LineMaterial: openrocket.LineMaterial{Density: 0.002},
			},
			wantMass: 0.0,
		},
		{
			name: "Zero Canopy Density",
			parachute: &openrocket.Parachute{
				Name:         "Zero Density Chute",
				Diameter:     1.0,
				Material:     openrocket.Material{Density: 0.0},
				LineCount:    8,
				LineLength:   1.5,
				LineMaterial: openrocket.LineMaterial{Density: 0.002},
			},
			wantMass: 0.0, // Function returns 0 if canopy density is <= 0
		},
		{
			name: "Zero Line Length",
			parachute: &openrocket.Parachute{
				Name:         "Zero Line Length Chute",
				Diameter:     1.0,
				Material:     openrocket.Material{Density: 0.05},
				LineCount:    8,
				LineLength:   0.0,
				LineMaterial: openrocket.LineMaterial{Density: 0.002},
			},
			wantMass: math.Pi * 0.5 * 0.5 * 0.05, // Only canopy mass
		},
		{
			name: "Zero Line Count",
			parachute: &openrocket.Parachute{
				Name:         "Zero Line Count Chute",
				Diameter:     1.0,
				Material:     openrocket.Material{Density: 0.05},
				LineCount:    0,
				LineLength:   1.5,
				LineMaterial: openrocket.LineMaterial{Density: 0.002},
			},
			wantMass: math.Pi * 0.5 * 0.5 * 0.05, // Only canopy mass
		},
		{
			name: "Zero Line Density",
			parachute: &openrocket.Parachute{
				Name:         "Zero Line Density Chute",
				Diameter:     1.0,
				Material:     openrocket.Material{Density: 0.05},
				LineCount:    8,
				LineLength:   1.5,
				LineMaterial: openrocket.LineMaterial{Density: 0.0},
			},
			wantMass: math.Pi * 0.5 * 0.5 * 0.05, // Only canopy mass
		},
		{
			name: "Negative Line Density (triggers warning)",
			parachute: &openrocket.Parachute{
				Name:         "Negative Line Density Chute",
				Diameter:     0.1, // Small diameter to make line mass dominant
				Material:     openrocket.Material{Density: 0.01},
				LineCount:    8,
				LineLength:   1.0,
				LineMaterial: openrocket.LineMaterial{Density: -1.0}, // Negative density
			},
			wantMass: 0.0, // Should return 0 due to negative mass check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMass := tt.parachute.GetMass()
			// Use a small tolerance for floating point comparison
			if math.Abs(gotMass-tt.wantMass) > 1e-9 {
				t.Errorf("Parachute.GetMass() = %v, want %v", gotMass, tt.wantMass)
			}
		})
	}
}
