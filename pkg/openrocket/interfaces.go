package openrocket

// MassProvider defines an interface for components that can provide their mass.
// This helps in aggregating mass from various component types.
type MassProvider interface {
	GetMass() float64
}
