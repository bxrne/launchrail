package types_test

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs/types"
)

// TEST: GIVEN two vectors WHEN Add is called THEN the sum of the vectors is returned.
func TestVector3_Add(t *testing.T) {
	tests := []struct {
		name     string
		v1       types.Vector3
		v2       types.Vector3
		expected types.Vector3
	}{
		{
			name:     "add positive vectors",
			v1:       types.Vector3{1, 2, 3},
			v2:       types.Vector3{4, 5, 6},
			expected: types.Vector3{5, 7, 9},
		},
		{
			name:     "add negative vectors",
			v1:       types.Vector3{-1, -2, -3},
			v2:       types.Vector3{-4, -5, -6},
			expected: types.Vector3{-5, -7, -9},
		},
		{
			name:     "add mixed sign vectors",
			v1:       types.Vector3{1, -2, 3},
			v2:       types.Vector3{-4, 5, -6},
			expected: types.Vector3{-3, 3, -3},
		},
		{
			name:     "add zero vector",
			v1:       types.Vector3{1, 2, 3},
			v2:       types.Vector3{0, 0, 0},
			expected: types.Vector3{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Add(tt.v2)
			if result != tt.expected {
				t.Errorf("Add() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TEST: GIVEN a vector and a scalar WHEN Scale is called THEN the scaled vector is returned.
func TestVector3_Scale(t *testing.T) {
	tests := []struct {
		name     string
		v        types.Vector3
		scalar   float64
		expected types.Vector3
	}{
		{
			name:     "scale by positive",
			v:        types.Vector3{1, 2, 3},
			scalar:   2,
			expected: types.Vector3{2, 4, 6},
		},
		{
			name:     "scale by negative",
			v:        types.Vector3{1, 2, 3},
			scalar:   -2,
			expected: types.Vector3{-2, -4, -6},
		},
		{
			name:     "scale by zero",
			v:        types.Vector3{1, 2, 3},
			scalar:   0,
			expected: types.Vector3{0, 0, 0},
		},
		{
			name:     "scale negative vector",
			v:        types.Vector3{-1, -2, -3},
			scalar:   2,
			expected: types.Vector3{-2, -4, -6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Scale(tt.scalar)
			if result != tt.expected {
				t.Errorf("Scale() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TEST: GIVEN a vector WHEN Magnitude is called THEN the length of the vector is returned.
func TestVector3_Magnitude(t *testing.T) {
	tests := []struct {
		name     string
		v        types.Vector3
		expected float64
	}{
		{
			name:     "unit vector",
			v:        types.Vector3{1, 0, 0},
			expected: 1.0,
		},
		{
			name:     "zero vector",
			v:        types.Vector3{0, 0, 0},
			expected: 0.0,
		},
		{
			name:     "right angle triangle",
			v:        types.Vector3{3, 4, 0},
			expected: 5.0,
		},
		{
			name:     "negative components",
			v:        types.Vector3{-1, -1, -1},
			expected: math.Sqrt(3),
		},
	}

	const epsilon = 1e-10 // Small value for floating-point comparison

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Magnitude()
			if math.Abs(result-tt.expected) > epsilon {
				t.Errorf("Magnitude() = %v, want %v", result, tt.expected)
			}
		})
	}
}
