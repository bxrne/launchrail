package thrustcurves

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/pkg/designation"
)

// NOTE: Assemble motor data from the ThrustCurve API.
func Load(designationString string, client http_client.HTTPClient) (*MotorData, error) {
	des, err := designation.New(designationString)

	if err != nil {
		return nil, fmt.Errorf("failed to create motor designation: %s", err)
	}

	props, err := getMotorProps(des, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get motor ID: %s", err)
	}

	curve, err := getMotorCurve(props.Results[0].MotorID, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get motor curve: %s", err)
	}

	return &MotorData{
		Designation:  des,
		ID:           props.Results[0].MotorID,
		Thrust:       curve,
		TotalImpulse: props.Results[0].TotalImpulse,
		BurnTime:     props.Results[0].BurnTime,
		AvgThrust:    props.Results[0].AvgThrust,
		TotalMass:    props.Results[0].TotalMass / 1000, // Convert grams to kg
		WetMass:      props.Results[0].WetMass / 1000,   // Convert grams to kg
		MaxThrust:    props.Results[0].MaxThrust,
	}, nil

}

// NOTE: Search for the motor ID using the designation via the ThrustCurve API.
func getMotorProps(designation designation.Designation, client http_client.HTTPClient) (SearchResponse, error) {
	url := "https://www.thrustcurve.org/api/v1/search.json"
	requestBody := map[string]interface{}{
		"designation": designation,
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return SearchResponse{}, err
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return SearchResponse{}, err
	}

	var searchResponse SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return SearchResponse{}, err
	}

	if len(searchResponse.Results) == 0 {
		return SearchResponse{}, fmt.Errorf("no results found for motor designation %s", designation)
	}

	return searchResponse, nil
}

// NOTE: Download the motor curve using the motor ID via the ThrustCurve API.
func getMotorCurve(id string, client http_client.HTTPClient) ([][]float64, error) {
	url := "https://www.thrustcurve.org/api/v1/download.json"
	requestBody := map[string]interface{}{
		"motorIds":   []string{id},
		"format":     "RASP",
		"license":    "PD",
		"data":       "samples",
		"maxResults": 1024,
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var downloadResponse DownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&downloadResponse); err != nil {
		return nil, err
	}

	if len(downloadResponse.Results) == 0 || len(downloadResponse.Results[0].Samples) == 0 {
		return nil, fmt.Errorf("no curve data found for motor ID %s", id)
	}

	curve := make([][]float64, len(downloadResponse.Results[0].Samples))
	for i, sample := range downloadResponse.Results[0].Samples {
		curve[i] = []float64{sample.Time, sample.Thrust}
	}

	return curve, nil
}
