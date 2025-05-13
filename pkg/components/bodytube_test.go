package components_test

import (
	"encoding/xml"
	"fmt"
	"math"
	"testing"

	"github.com/EngoEngine/ecs"

	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a new Bodytube WHEN Type is called THEN "Bodytube" is returned
func TestBodytubeType(t *testing.T) {
	bt := components.Bodytube{}

	assert.Equal(t, "Bodytube", bt.Type(), "Type should return 'Bodytube'")
}

// TEST: GIVEN a new Bodytube WHEN String is called THEN a string representation is returned
func TestBodytubeString(t *testing.T) {
	bt := components.Bodytube{
		ID:           ecs.NewBasic(),
		Radius:       1.0,
		Length:       2.0,
		Mass:         0.5,
		Thickness:    0.1,
		MaterialName: "Test Material",
		Density:      1.2,
	}

	expected := fmt.Sprintf("Bodytube{ID: %d, Position: Vector3{X: 0.00, Y: 0.00, Z: 0.00}, Radius: 1.00, Length: 2.00, Mass: 0.50, Thickness: 0.10, Material: Test Material, Density: 1.20}", bt.ID.ID())
	assert.Equal(t, expected, bt.String(), "String representation should match expected format")
}

// TEST: GIVEN a new Bodytube WHEN SetID is called THEN the ID is updated
func TestBodytubeUpdate(t *testing.T) {
	bt := components.Bodytube{}

	// Ensure no error occurs on update (currently does nothing)
	err := bt.Update(0.016) // dt = 16ms
	assert.NoError(t, err, "Update should not return an error")
}

// TEST: GIVEN a new Bodytube WHEN Remove is called THEN the component is removed
func TestNewBodytube(t *testing.T) {
	id := ecs.NewBasic()
	radius := 1.0
	length := 2.0
	// mass := 0.5 // This is the target mass
	thickness := 0.1

	// Calculate required density for expectedMass = 0.5
	// Mass = Volume * Density  => Density = Mass / Volume
	// Volume = Pi * (R_outer^2 - R_inner^2) * L
	outerVolume := math.Pi * radius * radius * length
	innerRadius := radius - thickness
	if innerRadius < 0 {
		innerRadius = 0
	}
	innerVolume := math.Pi * innerRadius * innerRadius * length
	volume := outerVolume - innerVolume

	expectedMass := 0.5
	var density float64
	if volume > 1e-9 { // Avoid division by zero if volume is too small
		density = expectedMass / volume
	} else {
		// If volume is zero (or very small) and we expect non-zero mass, this test setup is problematic.
		// For this test, assume dimensions will yield a valid volume for the expected mass.
		// If expectedMass was 0, density could also be 0.
		density = 0
		if expectedMass > 1e-9 {
			t.Fatalf("Cannot achieve expected mass %.2f with near-zero volume (%.2e) for given dimensions", expectedMass, volume)
		}
	}

	bt := components.NewBodytube(id, radius, length, thickness, density)

	assert.NotNil(t, bt, "Bodytube should be created")
	assert.Equal(t, id, bt.ID, "Bodytube ID should match")
	assert.Equal(t, radius, bt.Radius, "Bodytube radius should match")
	assert.Equal(t, length, bt.Length, "Bodytube length should match")
	assert.InDelta(t, expectedMass, bt.Mass, 1e-9, "Bodytube mass should match")
	assert.Equal(t, thickness, bt.Thickness, "Bodytube thickness should match")
	assert.InDelta(t, density, bt.Density, 1e-9, "Bodytube density should match") // Also check density
}

// TEST: GIVEN a new Bodytube WHEN NewBodytubeFromORK is called THEN a new Bodytube is created
func TestBodytubeFromORK(t *testing.T) {
	id := ecs.NewBasic()
	orkDoc := &openrocket.OpenrocketDocument{
		XMLName: xml.Name{Local: "openrocket"},
		Rocket: openrocket.RocketDocument{
			XMLName: xml.Name{Local: "rocket"},
			Subcomponents: openrocket.Subcomponents{
				Stages: []openrocket.RocketStage{{
					SustainerSubcomponents: openrocket.SustainerSubcomponents{
						BodyTube: openrocket.BodyTube{
							Radius:    "auto 0.5",
							Length:    2.0,
							Thickness: 0.1,
							Material: openrocket.Material{
								Name:    "Test Material",
								Type:    "BULK",
								Density: 1.2,
							},
							Finish: "Smooth",
						},
					},
				}},
			},
		},
	}

	bt, err := components.NewBodytubeFromORK(id, orkDoc)
	assert.NoError(t, err)
	assert.NotNil(t, bt)
	assert.Equal(t, 0.5, bt.Radius)
	assert.Equal(t, 2.0, bt.Length)
	assert.Equal(t, 0.1, bt.Thickness)
	assert.Equal(t, "Test Material", bt.MaterialName)
	assert.Equal(t, 1.2, bt.Density)
}

// TEST: GIVEN a new Bodytube WHEN NewBodytubeFromORK is called THEN a new Bodytube is created with auto radius
func TestBodytubeGetters(t *testing.T) {
	bt := components.Bodytube{
		Radius:      0.5,
		Length:      2.0,
		Mass:        1.0,
		Density:     1.2,
		Volume:      0.8,
		SurfaceArea: 6.28,
	}

	tests := []struct {
		name     string
		getter   func() float64
		expected float64
	}{
		{"GetPlanformArea", bt.GetPlanformArea, math.Pi * 0.25},
		{"GetMass", bt.GetMass, 1.0},
		{"GetDensity", bt.GetDensity, 1.2},
		{"GetVolume", bt.GetVolume, 0.8},
		{"GetSurfaceArea", bt.GetSurfaceArea, 6.28},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.getter()
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

// TEST: GIVEN a new Bodytube WHEN NewBodytubeFromORK is called THEN a new Bodytube is created with auto radius
func TestBodytubeFromORKInvalidRadius(t *testing.T) {
	id := ecs.NewBasic()
	orkDoc := &openrocket.OpenrocketDocument{
		XMLName: xml.Name{Local: "openrocket"},
		Rocket: openrocket.RocketDocument{
			XMLName: xml.Name{Local: "rocket"},
			Subcomponents: openrocket.Subcomponents{
				Stages: []openrocket.RocketStage{{
					SustainerSubcomponents: openrocket.SustainerSubcomponents{
						BodyTube: openrocket.BodyTube{
							Radius: "invalid",
						},
					},
				}},
			},
		},
	}

	bt, err := components.NewBodytubeFromORK(id, orkDoc)
	assert.Error(t, err)
	assert.Nil(t, bt)
}
