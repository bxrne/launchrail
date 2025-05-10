package types

import "math"

// Quaternion represents a 3D quaternion
type Quaternion struct {
	X, Y, Z, W float64
}

const epsilon = 1e-12 // Small threshold for magnitude checks

// NewQuaternion creates a new Quaternion
func NewQuaternion(x, y, z, w float64) *Quaternion {
	return &Quaternion{
		X: x,
		Y: y,
		Z: z,
		W: w,
	}
}

// IdentityQuaternion returns the identity quaternion
func IdentityQuaternion() *Quaternion {
	return &Quaternion{
		X: 0,
		Y: 0,
		Z: 0,
		W: 1,
	}
}

// Add adds two quaternions together
func (q *Quaternion) Add(q2 *Quaternion) *Quaternion {
	return &Quaternion{
		X: q.X + q2.X,
		Y: q.Y + q2.Y,
		Z: q.Z + q2.Z,
		W: q.W + q2.W,
	}
}

// Subtract subtracts one quaternion from another
func (q *Quaternion) Subtract(q2 *Quaternion) *Quaternion {
	return &Quaternion{
		X: q.X - q2.X,
		Y: q.Y - q2.Y,
		Z: q.Z - q2.Z,
		W: q.W - q2.W,
	}
}

// Multiply multiplies two quaternions together
func (q *Quaternion) Multiply(q2 *Quaternion) *Quaternion {
	return &Quaternion{
		X: q.W*q2.X + q.X*q2.W + q.Y*q2.Z - q.Z*q2.Y,
		Y: q.W*q2.Y + q.Y*q2.W + q.Z*q2.X - q.X*q2.Z,
		Z: q.W*q2.Z + q.Z*q2.W + q.X*q2.Y - q.Y*q2.X,
		W: q.W*q2.W - q.X*q2.X - q.Y*q2.Y - q.Z*q2.Z,
	}
}

// Scale scales a quaternion by a scalar
func (q *Quaternion) Scale(scalar float64) *Quaternion {
	return &Quaternion{
		X: q.X * scalar,
		Y: q.Y * scalar,
		Z: q.Z * scalar,
		W: q.W * scalar,
	}
}

// Normalize normalizes a quaternion
func (q *Quaternion) Normalize() *Quaternion {
	magnitudeSquared := q.Magnitude()

	// Check for invalid magnitude squared (negative or NaN)
	if magnitudeSquared < 0 || math.IsNaN(magnitudeSquared) {
		// Quaternion is invalid, return identity for stability
		return &Quaternion{W: 1, X: 0, Y: 0, Z: 0}
	}

	// Check for near-zero magnitude
	if magnitudeSquared <= epsilon {
		// Near-zero magnitude, treat as identity.
		return &Quaternion{W: 1, X: 0, Y: 0, Z: 0}
	}

	// Calculate magnitude
	mag := math.Sqrt(magnitudeSquared)

	// Final check on calculated magnitude (should be redundant if above checks are good)
	if mag <= epsilon || math.IsNaN(mag) || math.IsInf(mag, 0) {
		// If somehow mag is still invalid, return identity
		return &Quaternion{W: 1, X: 0, Y: 0, Z: 0}
	}

	// Normalize components
	return &Quaternion{
		X: q.X / mag,
		Y: q.Y / mag,
		Z: q.Z / mag,
		W: q.W / mag,
	}
}

// Magnitude returns the sum of squares of the quaternion's components
func (q *Quaternion) Magnitude() float64 {
	return q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W
}

// Integrate increments q by the angular velocity (radians/sec) over dt
func (q *Quaternion) Integrate(omega Vector3, dt float64) *Quaternion {
	// Small-angle approximation
	halfDt := dt * 0.5
	mag := math.Sqrt(omega.X*omega.X + omega.Y*omega.Y + omega.Z*omega.Z)
	if mag == 0 {
		return q
	}

	// Axis components
	ax := omega.X / mag
	ay := omega.Y / mag
	az := omega.Z / mag

	thetaOverTwo := mag * halfDt
	sinTerm := math.Sin(thetaOverTwo)

	dq := Quaternion{
		W: math.Cos(thetaOverTwo),
		X: ax * sinTerm,
		Y: ay * sinTerm,
		Z: az * sinTerm,
	}

	// q = q * dq (post multiplication)
	newQ := q.Multiply(&dq).Normalize()
	q.X, q.Y, q.Z, q.W = newQ.X, newQ.Y, newQ.Z, newQ.W
	return q
}

// Conjugate returns the conjugate of the quaternion
func (q *Quaternion) Conjugate() *Quaternion {
	return &Quaternion{
		X: -q.X,
		Y: -q.Y,
		Z: -q.Z,
		W: q.W,
	}
}

// Inverse returns the inverse of the quaternion.
// For a unit quaternion (which orientation quaternions should be after normalization),
// the inverse is its conjugate.
func (q *Quaternion) Inverse() *Quaternion {
	// Ensure the quaternion is normalized before taking conjugate as inverse.
	// This handles non-unit quaternions gracefully by effectively making them unit first.
	// However, for performance, if q is known to be unit, just Conjugate() could be called.
	// For safety in general use, normalizing first is better.
	// The Normalize method already handles zero/invalid magnitude by returning identity.
	normalizedQ := q.Normalize()
	return normalizedQ.Conjugate()
}

// RotateVector rotates a vector v by the quaternion q.
// Ensures q is normalized. Returns original v if q or v is invalid, or if q normalizes to identity.
func (q *Quaternion) RotateVector(v *Vector3) *Vector3 {
	// Check for invalid vector components
	if v == nil || math.IsNaN(v.X) || math.IsNaN(v.Y) || math.IsNaN(v.Z) ||
		math.IsInf(v.X, 0) || math.IsInf(v.Y, 0) || math.IsInf(v.Z, 0) {
		return v // Return original vector if v is invalid
	}

	// Check for invalid quaternion components
	if q == nil || math.IsNaN(q.X) || math.IsNaN(q.Y) || math.IsNaN(q.Z) || math.IsNaN(q.W) ||
		math.IsInf(q.X, 0) || math.IsInf(q.Y, 0) || math.IsInf(q.Z, 0) || math.IsInf(q.W, 0) {
		return v // Return original vector if q is invalid
	}

	// Normalize the quaternion (handles zero/invalid magnitude by returning identity)
	qNorm := q.Normalize()

	// If normalization results in identity, no rotation occurs (or q was invalid)
	if qNorm.IsIdentity() {
		return v
	}

	// Represent the vector as a pure quaternion
	p := &Quaternion{X: v.X, Y: v.Y, Z: v.Z, W: 0}

	// Compute the rotated vector using normalized quaternion: qNorm * p * qNorm.Conjugate()
	qConj := qNorm.Conjugate()
	rotatedP := qNorm.Multiply(p).Multiply(qConj)

	// Extract the vector part
	return &Vector3{X: rotatedP.X, Y: rotatedP.Y, Z: rotatedP.Z}
}

// IsIdentity checks if the quaternion is the identity quaternion
func (q *Quaternion) IsIdentity() bool {
	return math.Abs(q.X) <= epsilon && math.Abs(q.Y) <= epsilon && math.Abs(q.Z) <= epsilon && math.Abs(q.W-1) <= epsilon
}
