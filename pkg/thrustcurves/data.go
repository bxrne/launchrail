package thrustcurves

import (
	"github.com/bxrne/launchrail/pkg/designation"
)

// MotorData represents the motor data loaded from the ThrustCurve API
type MotorData struct {
	Designation designation.Designation
	ID          string
	Thrust      [][]float64
}

// GetMass returns the mass of the motor
func (m *MotorData) GetMass() float64 {
	return 0.0 // TODO: Fix
}

// SearchResponse represents the response from the ThrustCurve search API
type SearchResponse struct {
	Results []struct {
		MotorID string `json:"motorId"`
	} `json:"results"`
}

// DownloadResponse represents the response from the ThrustCurve download API
type DownloadResponse struct {
	Results []struct {
		Samples []struct {
			Time   float64 `json:"time"`
			Thrust float64 `json:"thrust"`
		} `json:"samples"`
	} `json:"results"`
}
