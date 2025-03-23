package types

import "math"

// Quaternion represents a 3D quaternion
type Quaternion struct {
	X, Y, Z, W float64
}

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

// Normalize normalizes a quaternion using the square root of the sum of squares
func (q *Quaternion) Normalize() *Quaternion {
	mag := math.Sqrt(q.Magnitude())
	if mag == 0 {
		return q // Return the original quaternion if the magnitude is zero.
	}
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

// RotateVector rotates a vector by the quaternion
func (q *Quaternion) RotateVector(v *Vector3) *Vector3 {
	// Rotate vector by quaternion
	quat := q.Conjugate()
	quat = quat.Multiply(&Quaternion{
		W: 0,
		X: v.X,
		Y: v.Y,
		Z: v.Z,
	}).Multiply(q)
	return &Vector3{
		X: quat.X,
		Y: quat.Y,
		Z: quat.Z,
	}
}
