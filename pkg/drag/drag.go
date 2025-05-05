package drag

import "math"

// DragCoefficient calculates the drag coefficient based on Mach number
// using approximations for different sonic regimes
func DragCoefficient(mach float64) float64 {
	switch {
	case mach < 0.8: // Subsonic
		// Base drag coefficient + slight increase with Mach
		return 0.2 + 0.1*mach
	case mach >= 0.8 && mach <= 1.2: // Transonic
		// Significant drag rise in transonic region
		return 0.5 + 0.5*math.Sin(math.Pi*(mach-0.8)/0.4)
	case mach > 1.2: // Supersonic
		// Gradual decrease in drag coefficient
		return 0.8 / (1 + math.Sqrt(mach))
	default:
		return 0.2
	}
}

// CalculateDragForce computes the drag force in Newtons
// velocity in m/s, density in kg/m³, reference area in m²
func CalculateDragForce(velocity, density, referenceArea float64, dragCoefficient float64) float64 {
	return 0.5 * density * math.Pow(velocity, 2) * referenceArea * dragCoefficient
}
