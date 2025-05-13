package components_test

import (
	"encoding/xml"
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
	orkDoc := &openrocket.OpenrocketDocument{
		XMLName: xml.Name{Local: "openrocket"},
		Rocket: openrocket.RocketDocument{
			XMLName: xml.Name{Local: "rocket"},
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

	// Test GetPosition
	t.Run("GetPosition", func(t *testing.T) {
		nc.Position = types.Vector3{X: 1.0, Y: 2.0, Z: 3.0}
		pos := nc.GetPosition()
		assert.Equal(t, types.Vector3{X: 1.0, Y: 2.0, Z: 3.0}, pos, "GetPosition should return the current position")
	})
}

// TestNoseconeCenterOfMassAndInertia tests the GetCenterOfMassLocal and GetInertiaTensorLocal methods
func TestNoseconeCenterOfMassAndInertia(t *testing.T) {
	// Regular case
	t.Run("regular nose cone", func(t *testing.T) {
		nc := components.Nosecone{
			Radius: 0.5,
			Length: 2.0,
			Mass:   1.0,
		}

		// Get center of mass
		cm := nc.GetCenterOfMassLocal()
		expectedCM := types.Vector3{X: 0, Y: 0, Z: 1.5} // 3/4 of length from tip
		assert.Equal(t, expectedCM, cm, "Center of mass should be 3/4 length from tip")

		// Get inertia tensor
		itensor := nc.GetInertiaTensorLocal()
		// Expected values based on formulas for a cone
		expectedIxxIyy := 1.0 * ((3.0/20.0)*0.5*0.5 + (3.0/80.0)*2.0*2.0)
		expectedIzz := (3.0 / 10.0) * 1.0 * 0.5 * 0.5

		assert.InDelta(t, expectedIxxIyy, itensor.M11, 0.001, "Ixx should match expected value")
		assert.InDelta(t, expectedIxxIyy, itensor.M22, 0.001, "Iyy should match expected value")
		assert.InDelta(t, expectedIzz, itensor.M33, 0.001, "Izz should match expected value")
		assert.InDelta(t, 0.0, itensor.M12, 0.001, "Off-diagonal elements should be zero")
		assert.InDelta(t, 0.0, itensor.M13, 0.001, "Off-diagonal elements should be zero")
		assert.InDelta(t, 0.0, itensor.M23, 0.001, "Off-diagonal elements should be zero")
	})

	// Edge case: Zero length
	t.Run("zero length", func(t *testing.T) {
		nc := components.Nosecone{
			Radius: 0.5,
			Length: 0.0, // Zero length
			Mass:   1.0,
		}

		// Get center of mass
		cm := nc.GetCenterOfMassLocal()
		expectedCM := types.Vector3{X: 0, Y: 0, Z: 0} // Special handling for zero length
		assert.Equal(t, expectedCM, cm, "Center of mass for zero length should be at origin")

		// Get inertia tensor
		itensor := nc.GetInertiaTensorLocal()
		// For zero length or radius, should get a zero tensor
		expectedTensor := types.Matrix3x3{} // All zeros

		assert.Equal(t, expectedTensor, itensor, "Inertia tensor for zero length should be zero")
	})

	// Edge case: Zero mass
	t.Run("zero mass", func(t *testing.T) {
		nc := components.Nosecone{
			Radius: 0.5,
			Length: 2.0,
			Mass:   0.0, // Zero mass
		}

		// For zero mass, inertia tensor should be zero
		itensor := nc.GetInertiaTensorLocal()
		expectedTensor := types.Matrix3x3{} // All zeros

		assert.Equal(t, expectedTensor, itensor, "Inertia tensor for zero mass should be zero")
	})

	// Edge case: Zero radius
	t.Run("zero radius", func(t *testing.T) {
		nc := components.Nosecone{
			Radius: 0.0, // Zero radius
			Length: 2.0,
			Mass:   1.0,
		}

		// Get inertia tensor
		itensor := nc.GetInertiaTensorLocal()
		// For zero radius, should get a zero tensor
		expectedTensor := types.Matrix3x3{} // All zeros

		assert.Equal(t, expectedTensor, itensor, "Inertia tensor for zero radius should be zero")
	})
}
