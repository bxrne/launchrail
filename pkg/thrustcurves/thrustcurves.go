package thrustcurves

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/zerodha/logf"
)

// NOTE: Assemble motor data from the ThrustCurve API.
func Load(designationString string, client http_client.HTTPClient, log logf.Logger) (*MotorData, error) {
	des, err := designation.New(designationString)

	if err != nil {
		log.Error("Failed to create motor designation", "designation", designationString, "error", err)
		return nil, fmt.Errorf("failed to create motor designation: %s", err)
	}

	props, err := getMotorProps(des, client, log)
	if err != nil {
		// Error already logged in getMotorProps
		return nil, fmt.Errorf("failed to get motor props: %w", err) // Wrap error
	}

	if len(props.Results) == 0 {
		log.Warn("No search results from ThrustCurve.org for motor", "designation", designationString)
		return nil, fmt.Errorf("no search results from ThrustCurve.org for %s", designationString)
	}

	log.Info("Successfully fetched motor properties", "designation", designationString, "motorID", props.Results[0].MotorID, "results_count", len(props.Results))

	curve, err := getMotorCurve(props.Results[0].MotorID, client, log)
	if err != nil {
		// Error already logged in getMotorCurve
		return nil, fmt.Errorf("failed to get motor curve: %w", err) // Wrap error
	}

	finalMotorData := &MotorData{
		Designation:  des,
		ID:           props.Results[0].MotorID,
		Thrust:       curve,
		TotalImpulse: props.Results[0].TotalImpulse,
		BurnTime:     props.Results[0].BurnTime,
		AvgThrust:    props.Results[0].AvgThrust,
		TotalMass:    props.Results[0].TotalMass / 1000, // Convert grams to kg
		WetMass:      props.Results[0].WetMass / 1000,   // Convert grams to kg
		MaxThrust:    props.Results[0].MaxThrust,
	}

	log.Info("Motor data loaded successfully", "designation", designationString, "motorID", finalMotorData.ID)
	return finalMotorData, nil

}

// NOTE: Search for the motor ID using the designation via the ThrustCurve API.
func getMotorProps(designation designation.Designation, client http_client.HTTPClient, log logf.Logger) (SearchResponse, error) {
	url := "https://www.thrustcurve.org/api/v1/search.json"
	requestBody := map[string]interface{}{
		"designation": designation,
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Error("Failed to marshal search request body", "error", err, "url", url)
		return SearchResponse{}, fmt.Errorf("marshaling search request: %w", err)
	}

	log.Debug("Sending motor search request to ThrustCurve.org", "url", url, "body", string(requestBodyJSON))
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Error("Motor search request failed", "error", err, "url", url)
		return SearchResponse{}, fmt.Errorf("http post for search: %w", err)
	}
	defer resp.Body.Close()

	log.Debug("Received motor search response", "url", url, "status_code", resp.StatusCode)

	bodyBytes, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		log.Error("Failed to read motor search response body", "error", errRead, "url", url, "status_code", resp.StatusCode)
		return SearchResponse{}, fmt.Errorf("reading search response body: %w", errRead)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("Motor search request returned non-OK status", "url", url, "status_code", resp.StatusCode, "response_body", string(bodyBytes))
		return SearchResponse{}, fmt.Errorf("motor search API error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var searchResponse SearchResponse
	if err := json.Unmarshal(bodyBytes, &searchResponse); err != nil {
		log.Error("Failed to unmarshal motor search response", "error", err, "url", url, "status_code", resp.StatusCode, "response_body", string(bodyBytes))
		return SearchResponse{}, fmt.Errorf("unmarshaling search response: %w", err)
	}

	if len(searchResponse.Results) == 0 {
		log.Warn("No results found for motor designation in search response", "designation", designation, "response_body", string(bodyBytes))
		// Return empty response, Load function will handle the error message for the user
		return searchResponse, nil
	}

	log.Debug("Successfully unmarshalled motor search response", "designation", designation, "results_count", len(searchResponse.Results))
	return searchResponse, nil
}

// NOTE: Download the motor curve using the motor ID via the ThrustCurve API.
func getMotorCurve(id string, client http_client.HTTPClient, log logf.Logger) ([][]float64, error) {
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
		log.Error("Failed to marshal curve download request body", "error", err, "url", url, "motor_id", id)
		return nil, fmt.Errorf("marshaling curve request: %w", err)
	}

	log.Debug("Sending motor curve download request to ThrustCurve.org", "url", url, "motor_id", id, "body", string(requestBodyJSON))
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Error("Motor curve download request failed", "error", err, "url", url, "motor_id", id)
		return nil, fmt.Errorf("http post for curve: %w", err)
	}
	defer resp.Body.Close()

	log.Debug("Received motor curve download response", "url", url, "motor_id", id, "status_code", resp.StatusCode)

	bodyBytes, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		log.Error("Failed to read motor curve download response body", "error", errRead, "url", url, "motor_id", id, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("reading curve response body: %w", errRead)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("Motor curve download request returned non-OK status", "url", url, "motor_id", id, "status_code", resp.StatusCode, "response_body", string(bodyBytes))
		return nil, fmt.Errorf("motor curve API error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var downloadResponse DownloadResponse
	if err := json.Unmarshal(bodyBytes, &downloadResponse); err != nil {
		log.Error("Failed to unmarshal motor curve download response", "error", err, "url", url, "motor_id", id, "status_code", resp.StatusCode, "response_body", string(bodyBytes))
		return nil, fmt.Errorf("unmarshaling curve response: %w", err)
	}

	if len(downloadResponse.Results) == 0 || len(downloadResponse.Results[0].Samples) == 0 {
		log.Warn("No curve data found for motor ID in download response", "motor_id", id, "response_body", string(bodyBytes))
		return nil, fmt.Errorf("no curve data found for motor ID %s", id)
	}

	curve := make([][]float64, len(downloadResponse.Results[0].Samples))
	for i, sample := range downloadResponse.Results[0].Samples {
		curve[i] = []float64{sample.Time, sample.Thrust}
	}

	log.Debug("Successfully unmarshalled motor curve data", "motor_id", id, "samples_count", len(curve))
	return curve, nil
}
