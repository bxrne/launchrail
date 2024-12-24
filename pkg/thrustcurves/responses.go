package thrustcurves

type SearchResponse struct {
	Results []struct {
		MotorID string `json:"motorId"`
	} `json:"results"`
}

type DownloadResponse struct {
	Results []struct {
		Samples []struct {
			Time   float64 `json:"time"`
			Thrust float64 `json:"thrust"`
		} `json:"samples"`
	} `json:"results"`
}
