package types_test

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const delta = 1e-9 // Tolerance for float comparisons

// Helper function to assert matrix equality
func assertMatrixEqual(t *testing.T, expected, actual *types.Matrix3x3, msgAndArgs ...interface{}) {
	require.NotNil(t, actual, msgAndArgs...)
	assert.InDelta(t, expected.M11, actual.M11, delta, msgAndArgs...)
	assert.InDelta(t, expected.M12, actual.M12, delta, msgAndArgs...)
	assert.InDelta(t, expected.M13, actual.M13, delta, msgAndArgs...)
	assert.InDelta(t, expected.M21, actual.M21, delta, msgAndArgs...)
	assert.InDelta(t, expected.M22, actual.M22, delta, msgAndArgs...)
	assert.InDelta(t, expected.M23, actual.M23, delta, msgAndArgs...)
	assert.InDelta(t, expected.M31, actual.M31, delta, msgAndArgs...)
	assert.InDelta(t, expected.M32, actual.M32, delta, msgAndArgs...)
	assert.InDelta(t, expected.M33, actual.M33, delta, msgAndArgs...)
}

// Helper function to assert vector equality
func assertVectorEqual(t *testing.T, expected, actual *types.Vector3, msgAndArgs ...interface{}) {
	require.NotNil(t, actual, msgAndArgs...)
	assert.InDelta(t, expected.X, actual.X, delta, msgAndArgs...)
	assert.InDelta(t, expected.Y, actual.Y, delta, msgAndArgs...)
	assert.InDelta(t, expected.Z, actual.Z, delta, msgAndArgs...)
}

func TestNewMatrix3x3(t *testing.T) {
	m := types.NewMatrix3x3(1, 2, 3, 4, 5, 6, 7, 8, 9)
	expected := &types.Matrix3x3{1, 2, 3, 4, 5, 6, 7, 8, 9}
	assertMatrixEqual(t, expected, m, "NewMatrix3x3 creation")
}

func TestIdentityMatrix(t *testing.T) {
	m := types.IdentityMatrix()
	expected := &types.Matrix3x3{1, 0, 0, 0, 1, 0, 0, 0, 1}
	assertMatrixEqual(t, expected, m, "Identity matrix")
}

func TestMultiplyVector(t *testing.T) {
	m := types.NewMatrix3x3(
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	)
	v := &types.Vector3{X: 1, Y: 2, Z: 3}
	expected := &types.Vector3{X: 1*1 + 2*2 + 3*3, Y: 4*1 + 5*2 + 6*3, Z: 7*1 + 8*2 + 9*3} // (14, 32, 50)
	actual := m.MultiplyVector(v)
	assertVectorEqual(t, expected, actual, "Matrix-Vector multiplication")
}

func TestTranspose(t *testing.T) {
	m := types.NewMatrix3x3(1, 2, 3, 4, 5, 6, 7, 8, 9)
	expected := types.NewMatrix3x3(1, 4, 7, 2, 5, 8, 3, 6, 9)
	actual := m.Transpose()
	assertMatrixEqual(t, expected, actual, "Matrix transpose")
}

func TestMultiplyMatrix(t *testing.T) {
	m1 := types.NewMatrix3x3(1, 2, 0, 3, 4, 0, 0, 0, 1)
	m2 := types.NewMatrix3x3(5, 6, 0, 7, 8, 0, 1, 0, 1)
	// Expected = m1 * m2
	// [1*5+2*7+0*1, 1*6+2*8+0*0, 1*0+2*0+0*1] = [19, 22, 0]
	// [3*5+4*7+0*1, 3*6+4*8+0*0, 3*0+4*0+0*1] = [43, 50, 0]
	// [0*5+0*7+1*1, 0*6+0*8+1*0, 0*0+0*0+1*1] = [ 1,  0, 1]
	expected := types.NewMatrix3x3(19, 22, 0, 43, 50, 0, 1, 0, 1)
	actual := m1.MultiplyMatrix(m2)
	assertMatrixEqual(t, expected, actual, "Matrix-Matrix multiplication")
}

func TestInverse(t *testing.T) {
	t.Run("Invertible Matrix", func(t *testing.T) {
		// Matrix from example: [[1, 2, 3], [0, 1, 4], [5, 6, 0]]
		// Det = 1(0-24) - 2(0-20) + 3(0-5) = -24 + 40 - 15 = 1
		m := types.NewMatrix3x3(1, 2, 3, 0, 1, 4, 5, 6, 0)
		inv := m.Inverse()
		require.NotNil(t, inv, "Inverse should exist")

		// Check M * M_inv = Identity
		identity := m.MultiplyMatrix(inv)
		assertMatrixEqual(t, types.IdentityMatrix(), identity, "M * M_inv should be Identity")
	})

	t.Run("Singular Matrix", func(t *testing.T) {
		// Matrix with det=0: [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
		m := types.NewMatrix3x3(1, 2, 3, 4, 5, 6, 7, 8, 9)
		inv := m.Inverse()
		assert.Nil(t, inv, "Inverse of singular matrix should be nil")
	})
}

func TestRotationMatrixFromQuaternion(t *testing.T) {
	// Test case 1: Identity quaternion -> Identity matrix
	q_ident := types.IdentityQuaternion()
	m_ident := types.RotationMatrixFromQuaternion(q_ident)
	assertMatrixEqual(t, types.IdentityMatrix(), m_ident, "Identity quaternion")

	// Test case 2: 90 deg rotation about Z axis
	// q = cos(pi/4) + k*sin(pi/4) = sqrt(2)/2 + k*sqrt(2)/2
	angle := math.Pi / 2.0
	q_90z := types.NewQuaternion(0, 0, math.Sin(angle/2.0), math.Cos(angle/2.0))
	m_90z := types.RotationMatrixFromQuaternion(q_90z)
	// Expected Rotation Matrix Rz(90) = [[0, -1, 0], [1, 0, 0], [0, 0, 1]]
	expected_m_90z := types.NewMatrix3x3(
		0, -1, 0,
		1, 0, 0,
		0, 0, 1,
	)
	assertMatrixEqual(t, expected_m_90z, m_90z, "90 deg Z rotation")

	// Test rotation of vector (1,0,0) -> (0,1,0)
	v := &types.Vector3{X: 1, Y: 0, Z: 0}
	rotated_v := m_90z.MultiplyVector(v)
	expected_rotated_v := &types.Vector3{X: 0, Y: 1, Z: 0}
	assertVectorEqual(t, expected_rotated_v, rotated_v, "Vector rotated by 90 deg Z")
}

func TestTransformInertiaBodyToWorld(t *testing.T) {
	// Simple diagonal body inertia
	I_body := types.NewMatrix3x3(1, 0, 0, 0, 2, 0, 0, 0, 3)

	// Rotation: 90 deg about Z axis (same as previous test)
	angle := math.Pi / 2.0
	q_90z := types.NewQuaternion(0, 0, math.Sin(angle/2.0), math.Cos(angle/2.0))
	R := types.RotationMatrixFromQuaternion(q_90z) // [[0, -1, 0], [1, 0, 0], [0, 0, 1]]

	// Expected I_world = R * I_body * R_transpose
	// R_transpose = [[0, 1, 0], [-1, 0, 0], [0, 0, 1]]
	// I_body * R_transpose = [[1,0,0],[0,2,0],[0,0,3]] * [[0,1,0],[-1,0,0],[0,0,1]]
	//                   = [[0, 1, 0], [-2, 0, 0], [0, 0, 3]]
	// R * (I_body * R_transpose) = [[0,-1,0],[1,0,0],[0,0,1]] * [[0,1,0],[-2,0,0],[0,0,3]]
	//                         = [[2, 0, 0], [0, 1, 0], [0, 0, 3]]
	expected_I_world := types.NewMatrix3x3(2, 0, 0, 0, 1, 0, 0, 0, 3)

	actual_I_world := types.TransformInertiaBodyToWorld(I_body, R)
	assertMatrixEqual(t, expected_I_world, actual_I_world, "Inertia tensor transformation")
}
