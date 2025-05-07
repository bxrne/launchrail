package types

import (
	"math"
)

// Matrix3x3 represents a 3x3 matrix.
// Components are row-major: M11, M12, M13 are the first row.
type Matrix3x3 struct {
	M11, M12, M13 float64 // Row 1
	M21, M22, M23 float64 // Row 2
	M31, M32, M33 float64 // Row 3
}

// NewMatrix3x3 creates a new matrix from components.
func NewMatrix3x3(m11, m12, m13, m21, m22, m23, m31, m32, m33 float64) *Matrix3x3 {
	return &Matrix3x3{m11, m12, m13, m21, m22, m23, m31, m32, m33}
}

// IdentityMatrix returns an identity matrix.
func IdentityMatrix() *Matrix3x3 {
	return &Matrix3x3{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}

// MultiplyVector multiplies the matrix by a column vector v and returns the resulting vector.
// result = M * v
func (m *Matrix3x3) MultiplyVector(v *Vector3) *Vector3 {
	return &Vector3{
		X: m.M11*v.X + m.M12*v.Y + m.M13*v.Z,
		Y: m.M21*v.X + m.M22*v.Y + m.M23*v.Z,
		Z: m.M31*v.X + m.M32*v.Y + m.M33*v.Z,
	}
}

// Transpose returns the transpose of the matrix.
func (m *Matrix3x3) Transpose() *Matrix3x3 {
	return &Matrix3x3{
		m.M11, m.M21, m.M31,
		m.M12, m.M22, m.M32,
		m.M13, m.M23, m.M33,
	}
}

// MultiplyMatrix multiplies this matrix by another matrix 'other' (M * other).
func (m *Matrix3x3) MultiplyMatrix(other *Matrix3x3) *Matrix3x3 {
	return &Matrix3x3{
		M11: m.M11*other.M11 + m.M12*other.M21 + m.M13*other.M31,
		M12: m.M11*other.M12 + m.M12*other.M22 + m.M13*other.M32,
		M13: m.M11*other.M13 + m.M12*other.M23 + m.M13*other.M33,

		M21: m.M21*other.M11 + m.M22*other.M21 + m.M23*other.M31,
		M22: m.M21*other.M12 + m.M22*other.M22 + m.M23*other.M32,
		M23: m.M21*other.M13 + m.M22*other.M23 + m.M23*other.M33,

		M31: m.M31*other.M11 + m.M32*other.M21 + m.M33*other.M31,
		M32: m.M31*other.M12 + m.M32*other.M22 + m.M33*other.M32,
		M33: m.M31*other.M13 + m.M32*other.M23 + m.M33*other.M33,
	}
}

// Inverse computes the inverse of a 3x3 matrix.
// Returns nil if the matrix is singular (determinant is zero).
func (m *Matrix3x3) Inverse() *Matrix3x3 {
	det := m.M11*(m.M22*m.M33-m.M23*m.M32) -
		m.M12*(m.M21*m.M33-m.M23*m.M31) +
		m.M13*(m.M21*m.M32-m.M22*m.M31)

	if math.Abs(det) < 1e-9 { // Consider very small determinant as singular
		return nil
	}

	invDet := 1.0 / det
	inv := &Matrix3x3{}

	inv.M11 = (m.M22*m.M33 - m.M23*m.M32) * invDet
	inv.M12 = (m.M13*m.M32 - m.M12*m.M33) * invDet
	inv.M13 = (m.M12*m.M23 - m.M13*m.M22) * invDet
	inv.M21 = (m.M23*m.M31 - m.M21*m.M33) * invDet
	inv.M22 = (m.M11*m.M33 - m.M13*m.M31) * invDet
	inv.M23 = (m.M13*m.M21 - m.M11*m.M23) * invDet
	inv.M31 = (m.M21*m.M32 - m.M22*m.M31) * invDet
	inv.M32 = (m.M12*m.M31 - m.M11*m.M32) * invDet
	inv.M33 = (m.M11*m.M22 - m.M12*m.M21) * invDet

	return inv
}

// RotationMatrixFromQuaternion converts a Quaternion to a 3x3 rotation matrix.
func RotationMatrixFromQuaternion(q *Quaternion) *Matrix3x3 {
	x, y, z, w := q.X, q.Y, q.Z, q.W

	xx, yy, zz := x*x, y*y, z*z
	xy, xz, yz := x*y, x*z, y*z
	wx, wy, wz := w*x, w*y, w*z

	return &Matrix3x3{
		M11: 1 - 2*(yy+zz),
		M12: 2 * (xy - wz),
		M13: 2 * (xz + wy),

		M21: 2 * (xy + wz),
		M22: 1 - 2*(xx+zz),
		M23: 2 * (yz - wx),

		M31: 2 * (xz - wy),
		M32: 2 * (yz + wx),
		M33: 1 - 2*(xx+yy),
	}
}

// TransformInertiaBodyToWorld transforms an inertia tensor from body frame to world frame.
// I_world = R * I_body * R_transpose
// where R is the rotation matrix from body to world.
func TransformInertiaBodyToWorld(inertiaBody *Matrix3x3, rotationMatrixBodyToWorld *Matrix3x3) *Matrix3x3 {
	R := rotationMatrixBodyToWorld
	RT := R.Transpose()
	// Temp = I_body * R_transpose
	temp := inertiaBody.MultiplyMatrix(RT)
	// I_world = R * Temp
	iWorld := R.MultiplyMatrix(temp)
	return iWorld
}
