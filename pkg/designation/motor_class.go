package designation

// DetermineMotorClass classifies motors based on total impulse following NAR/TRA standards
// Returns a letter designation (or other description) based on the NAR/TRA classification system
func DetermineMotorClass(totalImpulse float64) string {
	// NAR/TRA motor classification based on total impulse (N·s)
	// https://www.nar.org/standards-and-testing-committee/nar-standards/

	motorClasses := []struct {
		class      string
		minImpulse float64
		maxImpulse float64
	}{
		{"A", 1.26, 2.5},
		{"B", 2.51, 5.0},
		{"C", 5.01, 10.0},
		{"D", 10.01, 20.0},
		{"E", 20.01, 40.0},
		{"F", 40.01, 80.0},
		{"G", 80.01, 160.0},
		{"H", 160.01, 320.0},
		{"I", 320.01, 640.0},
		{"J", 640.01, 1280.0},
		{"K", 1280.01, 2560.0},
		{"L", 2560.01, 5120.0},
		{"M", 5120.01, 10240.0},
		{"N", 10240.01, 20480.0},
		{"O", 20480.01, 40960.0},
	}

	// Micro motors (less than A)
	if totalImpulse <= 1.25 && totalImpulse > 0 {
		return "1/4A-1/2A" // Micro motor range
	}

	// Zero impulse case
	if totalImpulse <= 0 {
		return "Unknown"
	}

	// Check standard motor classes
	for _, class := range motorClasses {
		if totalImpulse >= class.minImpulse && totalImpulse <= class.maxImpulse {
			return class.class
		}
	}

	// Beyond O class (experimental)
	if totalImpulse > 40960.0 {
		return "P+" // Experimental/custom high-power motor
	}

	return "Unknown" // Default if no match found
}
