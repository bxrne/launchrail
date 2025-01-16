package thrustcurves

import (
	"github.com/bxrne/launchrail/pkg/designation"
)

// MotorData represents the motor data loaded from the ThrustCurve API
type MotorData struct {
	Designation  designation.Designation
	ID           string
	Thrust       [][]float64
	TotalImpulse float64
	BurnTime     float64
	AvgThrust    float64
	TotalMass    float64
	WetMass      float64
	MaxThrust    float64
}

// GetMass returns the mass of the motor
func (m *MotorData) GetMass() float64 {
	return 0.0 // TODO: Fix
}

// SearchResponse represents the response from the ThrustCurve search API
type SearchResponse struct {
	Results []struct {
		MotorID      string  `json:"motorId"`
		AvgThrust    float64 `json:"avgThrustN"`
		MaxThrust    float64 `json:"maxThrustN"`
		TotalImpulse float64 `json:"totImpulseNs"`
		BurnTime     float64 `json:"burnTimeS"`
		TotalMass    float64 `json:"totalWeightG"`
		WetMass      float64 `json:"propWeightG"`
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
