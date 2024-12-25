package thrustcurves

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bxrne/launchrail/pkg/designation"
)

// NOTE: Assemble motor data from the ThrustCurve API.
func Load(raw_designation string) (*MotorData, error) {
	designation, err := designation.New(raw_designation)
	if err != nil {
		return nil, fmt.Errorf("failed to create motor designation: %s", err)
	}

	valid, err := designation.Validate()
	if !valid {
		return nil, fmt.Errorf("invalid motor designation: %s", designation)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to validate motor designation: %s", err)
	}

	id, err := getMotorID(designation)
	if err != nil {
		return nil, fmt.Errorf("failed to get motor ID: %s", err)
	}

	curve, err := getMotorCurve(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get motor curve: %s", err)
	}

	return &MotorData{
		Designation: designation,
		ID:          id,
		Thrust:      curve,
	}, nil
}

// NOTE: Search for the motor ID using the designation via the ThrustCurve API.
func getMotorID(designation designation.Designation) (string, error) {
	url := "https://www.thrustcurve.org/api/v1/search.json"
	requestBody := map[string]interface{}{
		"designation": designation,
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var searchResponse SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return "", err
	}

	if len(searchResponse.Results) == 0 {
		return "", fmt.Errorf("no motor found for designation %s", designation)
	}

	return searchResponse.Results[0].MotorID, nil
}

// NOTE: Download the motor curve using the motor ID via the ThrustCurve API.
func getMotorCurve(id string) ([][]float64, error) {
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

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBodyJSON))
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
