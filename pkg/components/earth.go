package components

type Earth int

const (
	FlatEarth Earth = iota
	SphericalEarth
	TopographicalEarth
)

func (e Earth) String() string {
	switch e {
	case FlatEarth:
		return "Flat Earth"
	case SphericalEarth:
		return "Spherical Earth"
	case TopographicalEarth:
		return "Topographical Earth"
	default:
		return "Unknown Earth Model"
	}
}
