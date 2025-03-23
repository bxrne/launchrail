package types

import "github.com/EngoEngine/ecs"

// Orientation represents a 3D orientation
type Orientation struct {
	ecs.BasicEntity
	Quat Quaternion
}

// Integrates the angular velocity to the orientation
func (o *Orientation) Integrate(angularVelocity Vector3, dt float64) {
	// Integrate quaternion
	o.Quat = *o.Quat.Multiply(&Quaternion{
		W: 0,
		X: angularVelocity.X * dt,
		Y: angularVelocity.Y * dt,
		Z: angularVelocity.Z * dt,
	}).Add(&o.Quat).Normalize()
}

// RotateVector rotates a vector by the orientation
func (o *Orientation) RotateVector(v Vector3) Vector3 {
	// Rotate vector by quaternion
	quat := o.Quat.Conjugate()
	quat = quat.Multiply(&Quaternion{
		W: 0,
		X: v.X,
		Y: v.Y,
		Z: v.Z,
	}).Multiply(&o.Quat)
	return Vector3{
		X: quat.X,
		Y: quat.Y,
		Z: quat.Z,
	}
}
