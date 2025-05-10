package types

// Matrix3x3 represents a 3x3 matrix.
// Components are row-major: M11, M12, M13 are the first row.
type Matrix3x3 struct {
	M11, M12, M13 float64 // Row 1
	M21, M22, M23 float64 // Row 2
	M31, M32, M33 float64 // Row 3
}

// NewMatrix3x3 creates a new matrix from a slice of 9 elements (row-major).
// Returns nil if the elements slice does not contain exactly 9 values.
func NewMatrix3x3(elements []float64) *Matrix3x3 {
	if len(elements) != 9 {
		// Consider logging an error here as well, e.g., using a package-level logger
		// For now, returning nil to indicate failure.
		return nil
	}
	return &Matrix3x3{
		M11: elements[0], M12: elements[1], M13: elements[2],
		M21: elements[3], M22: elements[4], M23: elements[5],
		M31: elements[6], M32: elements[7], M33: elements[8],
	}
}

// IdentityMatrix returns an identity matrix.
func IdentityMatrix() *Matrix3x3 {
	return &Matrix3x3{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}

// IdentityMatrix3x3 returns an identity matrix.
func IdentityMatrix3x3() Matrix3x3 {
	return Matrix3x3{
		M11: 1, M12: 0, M13: 0,
		M21: 0, M22: 1, M23: 0,
		M31: 0, M32: 0, M33: 1,
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

// Add returns the sum of two matrices (m + other).
func (m Matrix3x3) Add(other Matrix3x3) Matrix3x3 {
	return Matrix3x3{
		M11: m.M11 + other.M11, M12: m.M12 + other.M12, M13: m.M13 + other.M13,
		M21: m.M21 + other.M21, M22: m.M22 + other.M22, M23: m.M23 + other.M23,
		M31: m.M31 + other.M31, M32: m.M32 + other.M32, M33: m.M33 + other.M33,
	}
}

// Subtract returns the difference of two matrices (m - other).
func (m Matrix3x3) Subtract(other Matrix3x3) Matrix3x3 {
	return Matrix3x3{
		M11: m.M11 - other.M11, M12: m.M12 - other.M12, M13: m.M13 - other.M13,
		M21: m.M21 - other.M21, M22: m.M22 - other.M22, M23: m.M23 - other.M23,
		M31: m.M31 - other.M31, M32: m.M32 - other.M32, M33: m.M33 - other.M33,
	}
}

// MultiplyScalar returns the matrix scaled by a scalar.
func (m Matrix3x3) MultiplyScalar(s float64) Matrix3x3 {
	return Matrix3x3{
		M11: m.M11 * s, M12: m.M12 * s, M13: m.M13 * s,
		M21: m.M21 * s, M22: m.M22 * s, M23: m.M23 * s,
		M31: m.M31 * s, M32: m.M32 * s, M33: m.M33 * s,
	}
}

// Determinant calculates the determinant of the matrix.
func (m Matrix3x3) Determinant() float64 {
	return m.M11*(m.M22*m.M33-m.M23*m.M32) -
		m.M12*(m.M21*m.M33-m.M23*m.M31) +
		m.M13*(m.M21*m.M32-m.M22*m.M31)
}

// Inverse calculates the inverse of the matrix.
// Returns nil if the matrix is singular (determinant is zero or very close to zero).
func (m Matrix3x3) Inverse() *Matrix3x3 {
	det := m.Determinant()
	// Use a small epsilon for singularity check to handle floating point inaccuracies
	if det > -1e-9 && det < 1e-9 {
		return nil // Singular matrix
	}

	invDet := 1.0 / det
	adj := Matrix3x3{
		M11: (m.M22*m.M33 - m.M23*m.M32) * invDet,
		M12: (m.M13*m.M32 - m.M12*m.M33) * invDet,
		M13: (m.M12*m.M23 - m.M13*m.M22) * invDet,

		M21: (m.M23*m.M31 - m.M21*m.M33) * invDet,
		M22: (m.M11*m.M33 - m.M13*m.M31) * invDet,
		M23: (m.M13*m.M21 - m.M11*m.M23) * invDet,

		M31: (m.M21*m.M32 - m.M22*m.M31) * invDet,
		M32: (m.M12*m.M31 - m.M11*m.M32) * invDet,
		M33: (m.M11*m.M22 - m.M12*m.M21) * invDet,
	}
	return &adj
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
