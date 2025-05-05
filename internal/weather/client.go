package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const openMeteoURL = "https://api.open-meteo.com/v1/forecast"

// Client handles communication with the weather API.
type Client struct {
	HTTPClient *http.Client
}

// OpenMeteoResponse defines the structure for the relevant parts of the API response.
type OpenMeteoResponse struct {
	Latitude  float64     `json:"latitude"`
	Longitude float64     `json:"longitude"`
	Hourly    HourlyUnits `json:"hourly_units"`
	HourlyData HourlyData  `json:"hourly"`
}

// HourlyUnits defines the units for the hourly data.
type HourlyUnits struct {
	Time            string `json:"time"`
	WindSpeed10m    string `json:"windspeed_10m"`    // e.g., "km/h"
	WindDirection10m string `json:"winddirection_10m"` // e.g., "°"
}

// HourlyData contains the arrays of hourly forecast data.
type HourlyData struct {
	Time            []string  `json:"time"`             // ISO8601 time strings
	WindSpeed10m    []float64 `json:"windspeed_10m"`
	WindDirection10m []float64 `json:"winddirection_10m"`
}

// WindInfo holds the processed wind data for a specific time.
type WindInfo struct {
	Speed     float64 // meters per second
	Direction float64 // degrees
	Timestamp time.Time
}

// NewClient creates a new weather API client.
func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetWindData fetches wind data for the given coordinates.
// It currently returns the wind data for the *first* forecast hour.
// TODO: Add logic to select the most relevant forecast time.
func (c *Client) GetWindData(latitude, longitude float64) (*WindInfo, error) {
	// Construct the API request URL
	// Requesting hourly wind speed and direction at 10m for the next day.
	apiURL := fmt.Sprintf("%s?latitude=%.4f&longitude=%.4f&hourly=windspeed_10m,winddirection_10m&forecast_days=1",
		openMeteoURL, latitude, longitude)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create weather API request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute weather API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API request failed with status: %s", resp.Status)
	}

	var weatherData OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return nil, fmt.Errorf("failed to decode weather API response: %w", err)
	}

	// Basic validation: Check if we received hourly data
	if len(weatherData.HourlyData.Time) == 0 || len(weatherData.HourlyData.WindSpeed10m) == 0 || len(weatherData.HourlyData.WindDirection10m) == 0 {
		return nil, fmt.Errorf("incomplete hourly data received from weather API")
	}

	// --- Data Processing ---

	// Get the first hour's data
	firstTimeStr := weatherData.HourlyData.Time[0]
	firstSpeed := weatherData.HourlyData.WindSpeed10m[0]
	firstDirection := weatherData.HourlyData.WindDirection10m[0]

	// Parse the timestamp
	ts, err := time.Parse(time.RFC3339, firstTimeStr)
	if err != nil {
		// Log warning? Return error? For now, return error.
		return nil, fmt.Errorf("failed to parse timestamp from weather API: %w", err)
	}

	// Convert wind speed (assuming km/h from API) to m/s
	// Check units field if needed for robustness
	speedMS := firstSpeed * 1000 / 3600 // km/h to m/s

	windInfo := &WindInfo{
		Speed:     speedMS,
		Direction: firstDirection,
		Timestamp: ts,
	}

	return windInfo, nil
}
