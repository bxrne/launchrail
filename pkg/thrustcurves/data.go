package thrustcurves

import (
	"github.com/bxrne/launchrail/pkg/designation"
)

// MotorData represents the motor data loaded from the ThrustCurve API
type MotorData struct {
	Designation  designation.Designation
	ID           string
	Thrust       [][]float64 // [[time, thrust], ...]
	TotalImpulse float64     // Newton-seconds
	BurnTime     float64     // Seconds
	AvgThrust    float64     // Newtons
	TotalMass    float64     // Kg
	WetMass      float64     // Kg
	MaxThrust    float64     // Newtons
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
