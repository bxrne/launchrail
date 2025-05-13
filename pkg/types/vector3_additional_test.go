package types_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestVectorDot(t *testing.T) {
	// Test dot product of orthogonal vectors
	v1 := types.Vector3{X: 1, Y: 0, Z: 0}
	v2 := types.Vector3{X: 0, Y: 1, Z: 0}
	
	result := v1.Dot(v2)
	assert.Equal(t, 0.0, result)
	
	// Test dot product of parallel vectors
	v3 := types.Vector3{X: 2, Y: 3, Z: 4}
	v4 := types.Vector3{X: 2, Y: 3, Z: 4}
	
	result = v3.Dot(v4)
	expected := v3.X*v4.X + v3.Y*v4.Y + v3.Z*v4.Z
	assert.Equal(t, expected, result)
	
	// Test dot product of arbitrary vectors
	v5 := types.Vector3{X: 1, Y: 2, Z: 3}
	v6 := types.Vector3{X: 4, Y: 5, Z: 6}
	
	result = v5.Dot(v6)
	expected = v5.X*v6.X + v5.Y*v6.Y + v5.Z*v6.Z
	assert.Equal(t, expected, result)
}

func TestVectorNormalize(t *testing.T) {
	// Test with a non-zero vector
	v1 := types.Vector3{X: 3, Y: 4, Z: 0}
	norm := v1.Normalize()
	
	// The magnitude of the normalized vector should be 1
	mag := norm.Magnitude()
	assert.InDelta(t, 1.0, mag, 1e-10)
	
	// Check specific values
	expected := types.Vector3{X: 3.0/5.0, Y: 4.0/5.0, Z: 0}
	assert.InDelta(t, expected.X, norm.X, 1e-10)
	assert.InDelta(t, expected.Y, norm.Y, 1e-10)
	assert.InDelta(t, expected.Z, norm.Z, 1e-10)
	
	// Test with a zero vector
	v2 := types.Vector3{X: 0, Y: 0, Z: 0}
	norm = v2.Normalize()
	
	// The result should be a zero vector
	assert.Equal(t, 0.0, norm.X)
	assert.Equal(t, 0.0, norm.Y)
	assert.Equal(t, 0.0, norm.Z)
}
