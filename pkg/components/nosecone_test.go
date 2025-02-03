package components_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/types"

	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a new Nosecone WHEN Type is called THEN "Nosecone" is returned
func TestNewNosecone(t *testing.T) {
	id := ecs.NewBasic()
	radius := 1.0
	length := 2.0
	mass := 0.5
	shapeParameter := 0.1

	nosecone := components.NewNosecone(id, radius, length, mass, shapeParameter)

	assert.NotNil(t, nosecone, "Nosecone should be created")
	assert.Equal(t, id, nosecone.ID, "Nosecone ID should match")
	assert.Equal(t, types.Vector3{X: 0, Y: 0, Z: 0}, nosecone.Position, "Nosecone position should be zeroed")
	assert.Equal(t, radius, nosecone.Radius, "Nosecone radius should match")
	assert.Equal(t, length, nosecone.Length, "Nosecone length should match")
	assert.Equal(t, mass, nosecone.Mass, "Nosecone mass should match")
	assert.Equal(t, shapeParameter, nosecone.ShapeParameter, "Nosecone shape parameter should match")
}

// TEST: GIVEN a new Nosecone WHEN Type is called THEN "Nosecone" is returned
func TestNosecone_String(t *testing.T) {
	nosecone := components.Nosecone{
		ID:           ecs.NewBasic(),
		Position:     types.Vector3{X: 0, Y: 0, Z: 0},
		Radius:       1.0,
		Length:       2.0,
		Mass:         0.5,
		Shape:        "ogive",
		MaterialName: "Test Material",
		Density:      1.2,
	}

	expected := fmt.Sprintf("Nosecone{ID: %d, Position: Vector3{X: 0.00, Y: 0.00, Z: 0.00}, Radius: 1.00, Length: 2.00, Mass: 0.50, Shape: ogive, Material: Test Material, Density: 1.20}", nosecone.ID.ID())
	assert.Equal(t, expected, nosecone.String(), "String representation should match expected format")
}

// TEST: GIVEN a new Nosecone WHEN Update is called THEN the component is updated
func TestNosecone_Update(t *testing.T) {
	id := ecs.NewBasic()
	radius := 1.0
	length := 2.0
	mass := 0.5
	shapeParameter := 0.1

	nosecone := components.NewNosecone(id, radius, length, mass, shapeParameter)

	// Ensure no error occurs on update (currently does nothing)
	err := nosecone.Update(0.016) // dt = 16ms
	assert.NoError(t, err, "Update should not return an error")
}

// TEST: GIVEN a new Nosecone WHEN Remove is called THEN the component is removed
func TestNoseconeFromORK(t *testing.T) {
	id := ecs.NewBasic()
	orkDoc := &openrocket.RocketDocument{
		Subcomponents: openrocket.Subcomponents{
			Stages: []openrocket.RocketStage{{
				SustainerSubcomponents: openrocket.SustainerSubcomponents{
					Nosecone: openrocket.Nosecone{
						AftRadius:      0.5,
						Length:         2.0,
						Thickness:      0.1,
						Shape:          "ogive",
						ShapeParameter: 0.5,
						Material: openrocket.Material{
							Name:    "Test Material",
							Type:    "BULK",
							Density: 1.2,
						},
						Finish:            "Smooth",
						AftShoulderRadius: 0.4,
						AftShoulderLength: 0.3,
						Subcomponents: openrocket.NoseSubcomponents{
							MassComponent: openrocket.MassComponent{
								Mass: 0.1,
							},
						},
					},
				},
			}},
		},
	}

	nc := components.NewNoseconeFromORK(id, orkDoc)
	assert.NotNil(t, nc)
	assert.Equal(t, 0.5, nc.Radius)
	assert.Equal(t, 2.0, nc.Length)
	assert.Equal(t, "ogive", nc.Shape)
	assert.Equal(t, "Test Material", nc.MaterialName)
	assert.Equal(t, 1.2, nc.Density)
}

// TEST: GIVEN a new Nosecone WHEN Getters are called THEN the expected values are returned
func TestNoseconeGetters(t *testing.T) {
	nc := components.Nosecone{
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
		{"GetPlanformArea", nc.GetPlanformArea, math.Pi * 0.25},
		{"GetMass", nc.GetMass, 1.0},
		{"GetDensity", nc.GetDensity, 1.2},
		{"GetVolume", nc.GetVolume, 0.8},
		{"GetSurfaceArea", nc.GetSurfaceArea, 6.28},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.getter()
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}
