package components_test

import (
	"fmt"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/types"

	"github.com/stretchr/testify/assert"
)

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

func TestNosecone_String(t *testing.T) {
	id := ecs.NewBasic()
	radius := 1.0
	length := 2.0
	mass := 0.5
	shapeParameter := 0.1

	nosecone := components.NewNosecone(id, radius, length, mass, shapeParameter)

	expected := fmt.Sprintf("Nosecone{ID: %d, Position: Vector3{X: 0.00, Y: 0.00, Z: 0.00}, Radius: 1.00, Length: 2.00, Mass: 0.50, ShapeParameter: 0.10}", nosecone.ID.ID())
	assert.Equal(t, expected, nosecone.String(), "String representation should match expected format")
}

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
