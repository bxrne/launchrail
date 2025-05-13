package types_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestIdentityMatrix3x3(t *testing.T) {
	// Test the IdentityMatrix3x3 function
	identity := types.IdentityMatrix3x3()
	
	// Check that diagonal elements are 1
	assert.Equal(t, 1.0, identity.M11)
	assert.Equal(t, 1.0, identity.M22)
	assert.Equal(t, 1.0, identity.M33)
	
	// Check that off-diagonal elements are 0
	assert.Equal(t, 0.0, identity.M12)
	assert.Equal(t, 0.0, identity.M13)
	assert.Equal(t, 0.0, identity.M21)
	assert.Equal(t, 0.0, identity.M23)
	assert.Equal(t, 0.0, identity.M31)
	assert.Equal(t, 0.0, identity.M32)
}

func TestMatrixAdd(t *testing.T) {
	m1 := types.Matrix3x3{
		M11: 1, M12: 2, M13: 3,
		M21: 4, M22: 5, M23: 6,
		M31: 7, M32: 8, M33: 9,
	}
	
	m2 := types.Matrix3x3{
		M11: 9, M12: 8, M13: 7,
		M21: 6, M22: 5, M23: 4,
		M31: 3, M32: 2, M33: 1,
	}
	
	result := m1.Add(m2)
	
	// Each element should be the sum of the corresponding elements
	assert.Equal(t, 10.0, result.M11)
	assert.Equal(t, 10.0, result.M12)
	assert.Equal(t, 10.0, result.M13)
	assert.Equal(t, 10.0, result.M21)
	assert.Equal(t, 10.0, result.M22)
	assert.Equal(t, 10.0, result.M23)
	assert.Equal(t, 10.0, result.M31)
	assert.Equal(t, 10.0, result.M32)
	assert.Equal(t, 10.0, result.M33)
}

func TestMatrixSubtract(t *testing.T) {
	m1 := types.Matrix3x3{
		M11: 10, M12: 10, M13: 10,
		M21: 10, M22: 10, M23: 10,
		M31: 10, M32: 10, M33: 10,
	}
	
	m2 := types.Matrix3x3{
		M11: 5, M12: 4, M13: 3,
		M21: 2, M22: 1, M23: 0,
		M31: -1, M32: -2, M33: -3,
	}
	
	result := m1.Subtract(m2)
	
	// Each element should be the difference of the corresponding elements
	assert.Equal(t, 5.0, result.M11)
	assert.Equal(t, 6.0, result.M12)
	assert.Equal(t, 7.0, result.M13)
	assert.Equal(t, 8.0, result.M21)
	assert.Equal(t, 9.0, result.M22)
	assert.Equal(t, 10.0, result.M23)
	assert.Equal(t, 11.0, result.M31)
	assert.Equal(t, 12.0, result.M32)
	assert.Equal(t, 13.0, result.M33)
}

func TestMatrixMultiplyScalar(t *testing.T) {
	m := types.Matrix3x3{
		M11: 1, M12: 2, M13: 3,
		M21: 4, M22: 5, M23: 6,
		M31: 7, M32: 8, M33: 9,
	}
	
	scalar := 2.0
	result := m.MultiplyScalar(scalar)
	
	// Each element should be multiplied by the scalar
	assert.Equal(t, 2.0, result.M11)
	assert.Equal(t, 4.0, result.M12)
	assert.Equal(t, 6.0, result.M13)
	assert.Equal(t, 8.0, result.M21)
	assert.Equal(t, 10.0, result.M22)
	assert.Equal(t, 12.0, result.M23)
	assert.Equal(t, 14.0, result.M31)
	assert.Equal(t, 16.0, result.M32)
	assert.Equal(t, 18.0, result.M33)
}
