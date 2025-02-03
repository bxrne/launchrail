package components_test

import (
	"fmt"
	"testing"

	"github.com/EngoEngine/ecs"

	"github.com/bxrne/launchrail/pkg/components"
	"github.com/stretchr/testify/assert"
)

func TestBodytubeType(t *testing.T) {
	bt := components.Bodytube{}

	assert.Equal(t, "Bodytube", bt.Type(), "Type should return 'Bodytube'")
}

func TestBodytubeString(t *testing.T) {
	bt := components.Bodytube{
		ID:        ecs.NewBasic(),
		Radius:    1.0,
		Length:    2.0,
		Mass:      0.5,
		Thickness: 0.1,
	}

	expected := fmt.Sprintf("Bodytube{ID: %d, Position: Vector3{X: 0.00, Y: 0.00, Z: 0.00}, Radius: 1.00, Length: 2.00, Mass: 0.50, Thickness: 0.10}", bt.ID.ID())
	assert.Equal(t, expected, bt.String(), "String representation should match expected format")
}

func TestBodytubeUpdate(t *testing.T) {
	bt := components.Bodytube{}

	// Ensure no error occurs on update (currently does nothing)
	err := bt.Update(0.016) // dt = 16ms
	assert.NoError(t, err, "Update should not return an error")
}

func TestNewBodytube(t *testing.T) {
	id := ecs.NewBasic()
	radius := 1.0
	length := 2.0
	mass := 0.5
	thickness := 0.1

	bt := components.NewBodytube(id, radius, length, mass, thickness)

	assert.NotNil(t, bt, "Bodytube should be created")
	assert.Equal(t, id, bt.ID, "Bodytube ID should match")
	assert.Equal(t, radius, bt.Radius, "Bodytube radius should match")
	assert.Equal(t, length, bt.Length, "Bodytube length should match")
	assert.Equal(t, mass, bt.Mass, "Bodytube mass should match")
	assert.Equal(t, thickness, bt.Thickness, "Bodytube thickness should match")
}
