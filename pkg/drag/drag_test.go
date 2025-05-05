package drag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDragCoefficient(t *testing.T) {
	tests := []struct {
		name     string
		mach     float64
		expected float64
	}{
		{"Subsonic Low", 0.3, 0.23},
		{"Subsonic High", 0.7, 0.27},
		{"Transonic Start", 0.8, 0.5},
		{"Transonic Peak", 1.0, 1.0},
		{"Transonic End", 1.2, 0.5},
		{"Supersonic Low", 1.5, 0.36},
		{"Supersonic High", 3.0, 0.29},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := DragCoefficient(tt.mach)
			assert.InDelta(t, tt.expected, actual, 0.01, "unexpected drag coefficient")
		})
	}
}

func TestCalculateDragForce(t *testing.T) {
	tests := []struct {
		name            string
		velocity        float64
		density         float64
		referenceArea   float64
		dragCoefficient float64
		expected        float64
	}{
		{"Low Speed", 10.0, 1.225, 0.1, 0.5, 3.0625},
		{"High Speed", 300.0, 1.225, 0.1, 0.5, 2756.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := CalculateDragForce(tt.velocity, tt.density, tt.referenceArea, tt.dragCoefficient)
			assert.InDelta(t, tt.expected, actual, 0.01, "unexpected drag force")
		})
	}
}
