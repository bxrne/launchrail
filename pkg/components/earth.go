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

// INFO: Implementing the list.Item interface
func (e Earth) Title() string { return e.String() }
func (e Earth) Description() string {
	switch e {
	case FlatEarth:
		return "No elevation data, constant gravity, no curvature"
	case SphericalEarth:
		return "No elevation data, constant gravity, spherical curvature"
	case TopographicalEarth:
		return "Elevation data, variable gravity, spherical curvature"
	default:
		return ""
	}
}
func (e Earth) FilterValue() string { return e.String() }
