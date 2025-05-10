package types_test

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TestNewMatrix3x3 tests the NewMatrix3x3 constructor.
func TestNewMatrix3x3(t *testing.T) {
	m := types.NewMatrix3x3([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9})
	assert.NotNil(t, m)
	assert.Equal(t, 1.0, m.M11)
	assert.Equal(t, 9.0, m.M33)

	// Test with incorrect number of elements
	nilMatrix := types.NewMatrix3x3([]float64{1, 2, 3}) // Fewer than 9
	assert.Nil(t, nilMatrix)

	nilMatrix = types.NewMatrix3x3([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) // More than 9
	assert.Nil(t, nilMatrix)

	// Test with empty slice
	nilMatrix = types.NewMatrix3x3([]float64{})
	assert.Nil(t, nilMatrix)
}

// TestIdentityMatrix tests the IdentityMatrix constructor.
func TestIdentityMatrix(t *testing.T) {
	m := types.IdentityMatrix()
	assert.NotNil(t, m)
	assert.Equal(t, 1.0, m.M11)
	assert.Equal(t, 0.0, m.M12)
	assert.Equal(t, 1.0, m.M22)
	assert.Equal(t, 1.0, m.M33)
}

// TestMatrix3x3_MultiplyVector tests the MultiplyVector method.
func TestMatrix3x3_MultiplyVector(t *testing.T) {
	m := types.NewMatrix3x3([]float64{
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	})
	v := &types.Vector3{X: 1, Y: 2, Z: 3}
	result := m.MultiplyVector(v)
	expected := &types.Vector3{X: 1*1 + 2*2 + 3*3, Y: 4*1 + 5*2 + 6*3, Z: 7*1 + 8*2 + 9*3} // 14, 32, 50
	assert.Equal(t, expected.X, result.X)
	assert.Equal(t, expected.Y, result.Y)
	assert.Equal(t, expected.Z, result.Z)
}

// TestMatrix3x3_Transpose tests the Transpose method.
func TestMatrix3x3_Transpose(t *testing.T) {
	m := types.NewMatrix3x3([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9})
	expected := types.NewMatrix3x3([]float64{1, 4, 7, 2, 5, 8, 3, 6, 9})
	result := m.Transpose()
	assert.Equal(t, expected, result)
}

// TestMatrix3x3_MultiplyMatrix tests the MultiplyMatrix method.
func TestMatrix3x3_MultiplyMatrix(t *testing.T) {
	m1 := types.NewMatrix3x3([]float64{1, 2, 0, 3, 4, 0, 0, 0, 1})
	m2 := types.NewMatrix3x3([]float64{5, 6, 0, 7, 8, 0, 1, 0, 1})

	// m1 * m2 = [[1*5+2*7+0*1, 1*6+2*8+0*0, 1*0+2*0+0*1],
	//            [3*5+4*7+0*1, 3*6+4*8+0*0, 3*0+4*0+0*1],
	//            [0*5+0*7+1*1, 0*6+0*8+1*0, 0*0+0*0+1*1]]
	//          = [[19, 22, 0],
	//             [43, 50, 0],
	//             [1,  0,  1]]
	expected := types.NewMatrix3x3([]float64{19, 22, 0, 43, 50, 0, 1, 0, 1})
	result := m1.MultiplyMatrix(m2)
	assert.Equal(t, expected, result)
}

// TestMatrix3x3_Inverse tests the Inverse method.
func TestMatrix3x3_Inverse(t *testing.T) {
	// Test case 1: Invertible matrix
	m := types.NewMatrix3x3([]float64{1, 2, 3, 0, 1, 4, 5, 6, 0})
	mInv := m.Inverse()
	assert.NotNil(t, mInv, "Inverse should exist for non-singular matrix")

	// If mInv is correct, m * mInv should be identity
	identity := m.MultiplyMatrix(mInv)
	expectedIdentity := types.IdentityMatrix()
	assert.InDelta(t, expectedIdentity.M11, identity.M11, 1e-9, "Identity M11")
	assert.InDelta(t, expectedIdentity.M12, identity.M12, 1e-9, "Identity M12")
	assert.InDelta(t, expectedIdentity.M13, identity.M13, 1e-9, "Identity M13")
	assert.InDelta(t, expectedIdentity.M21, identity.M21, 1e-9, "Identity M21")
	assert.InDelta(t, expectedIdentity.M22, identity.M22, 1e-9, "Identity M22")
	assert.InDelta(t, expectedIdentity.M23, identity.M23, 1e-9, "Identity M23")
	assert.InDelta(t, expectedIdentity.M31, identity.M31, 1e-9, "Identity M31")
	assert.InDelta(t, expectedIdentity.M32, identity.M32, 1e-9, "Identity M32")
	assert.InDelta(t, expectedIdentity.M33, identity.M33, 1e-9, "Identity M33")

	// Test case 2: Singular matrix (determinant is 0)
	singularM := types.NewMatrix3x3([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9}) // Rows are linearly dependent (R3 = 2*R2 - R1)
	singularMInv := singularM.Inverse()
	assert.Nil(t, singularMInv, "Inverse should not exist for singular matrix")
}

// TestRotationMatrixFromQuaternion tests conversion from Quaternion to Rotation Matrix.
func TestRotationMatrixFromQuaternion(t *testing.T) {
	// Test case 1: Identity quaternion (no rotation)
	q_identity := types.IdentityQuaternion()
	m_identity := types.RotationMatrixFromQuaternion(q_identity)
	expected_m_identity := types.IdentityMatrix()
	assert.Equal(t, expected_m_identity, m_identity, "Identity quaternion should result in identity matrix")

	// Test case 2: 90-degree rotation about Z-axis
	// q = cos(theta/2) + sin(theta/2)*(xi + yj + zk)
	// theta = 90deg = pi/2. theta/2 = pi/4.
	// cos(pi/4) = sin(pi/4) = sqrt(2)/2 approx 0.707106781
	val := math.Sqrt(2) / 2.0
	q_90z := &types.Quaternion{X: 0, Y: 0, Z: val, W: val}
	m_90z := types.RotationMatrixFromQuaternion(q_90z)
	// Expected rotation matrix for +90deg around Z:
	// [cos(90) -sin(90)  0]
	// [sin(90)  cos(90)  0]
	// [  0        0      1]
	// = [0 -1 0]
	//   [1  0 0]
	//   [0  0 1]
	expected_m_90z := types.NewMatrix3x3([]float64{
		0, -1, 0,
		1, 0, 0,
		0, 0, 1,
	})

	assert.InDelta(t, expected_m_90z.M11, m_90z.M11, 1e-9)
	assert.InDelta(t, expected_m_90z.M12, m_90z.M12, 1e-9)
	assert.InDelta(t, expected_m_90z.M13, m_90z.M13, 1e-9)
	assert.InDelta(t, expected_m_90z.M21, m_90z.M21, 1e-9)
	assert.InDelta(t, expected_m_90z.M22, m_90z.M22, 1e-9)
	assert.InDelta(t, expected_m_90z.M23, m_90z.M23, 1e-9)
	assert.InDelta(t, expected_m_90z.M31, m_90z.M31, 1e-9)
	assert.InDelta(t, expected_m_90z.M32, m_90z.M32, 1e-9)
	assert.InDelta(t, expected_m_90z.M33, m_90z.M33, 1e-9)
}

// TestTransformInertiaBodyToWorld tests the inertia tensor transformation.
func TestTransformInertiaBodyToWorld(t *testing.T) {
	// Simple case: Body frame aligned with world frame (Identity rotation)
	I_body := types.NewMatrix3x3([]float64{1, 0, 0, 0, 2, 0, 0, 0, 3}) // Diagonal inertia tensor
	R_identity := types.IdentityMatrix()
	I_world_identity := types.TransformInertiaBodyToWorld(I_body, R_identity)
	assert.Equal(t, I_body, I_world_identity, "With identity rotation, I_world should equal I_body")

	// Case: 90-degree rotation about Z-axis
	// R = [0 -1 0]
	//     [1  0 0]
	//     [0  0 1]
	// RT = [0  1 0]
	//      [-1 0 0]
	//      [0  0 1]
	// I_body = diag(Ixx, Iyy, Izz) = diag(1, 2, 3)
	// I_world = R * I_body * RT should swap Ixx and Iyy, and keep Izz
	// I_world_xx = Iyy_body, I_world_yy = Ixx_body, I_world_zz = Izz_body
	val := math.Sqrt(2) / 2.0
	q_90z := &types.Quaternion{X: 0, Y: 0, Z: val, W: val}
	R_90z := types.RotationMatrixFromQuaternion(q_90z)

	expected_I_world := types.NewMatrix3x3([]float64{2, 0, 0, 0, 1, 0, 0, 0, 3})
	I_world_90z := types.TransformInertiaBodyToWorld(I_body, R_90z)

	assert.InDelta(t, expected_I_world.M11, I_world_90z.M11, 1e-9)
	assert.InDelta(t, expected_I_world.M12, I_world_90z.M12, 1e-9)
	assert.InDelta(t, expected_I_world.M13, I_world_90z.M13, 1e-9)
	assert.InDelta(t, expected_I_world.M21, I_world_90z.M21, 1e-9)
	assert.InDelta(t, expected_I_world.M22, I_world_90z.M22, 1e-9)
	assert.InDelta(t, expected_I_world.M23, I_world_90z.M23, 1e-9)
	assert.InDelta(t, expected_I_world.M31, I_world_90z.M31, 1e-9)
	assert.InDelta(t, expected_I_world.M32, I_world_90z.M32, 1e-9)
	assert.InDelta(t, expected_I_world.M33, I_world_90z.M33, 1e-9)
}
