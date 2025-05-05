package weather_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bxrne/launchrail/internal/weather"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWindData_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("latitude"), "52.5200")
		assert.Contains(t, r.URL.Query().Get("longitude"), "13.4100")
		assert.Equal(t, "windspeed_10m,winddirection_10m", r.URL.Query().Get("hourly"))
		assert.Equal(t, "1", r.URL.Query().Get("forecast_days"))

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{
			"latitude": 52.52,
			"longitude": 13.41,
			"hourly_units": {
				"time": "iso8601",
				"windspeed_10m": "km/h",
				"winddirection_10m": "°"
			},
			"hourly": {
				"time": ["2024-01-01T00:00:00Z", "2024-01-01T01:00:00Z"],
				"windspeed_10m": [18.0, 19.0],
				"winddirection_10m": [270.0, 280.0]
			}
		}`)
	}))
	defer server.Close()

	// Override the global constant for testing
	originalURL := weather.OpenMeteoURL
	weather.OpenMeteoURL = server.URL
	defer func() { weather.OpenMeteoURL = originalURL }()

	client := weather.NewClient()
	windInfo, err := client.GetWindData(52.52, 13.41)

	require.NoError(t, err)
	require.NotNil(t, windInfo)

	assert.InDelta(t, 5.0, windInfo.Speed, 0.001) // 18 km/h = 5 m/s
	assert.InDelta(t, 270.0, windInfo.Direction, 0.001)
	expectedTime, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	assert.Equal(t, expectedTime, windInfo.Timestamp)
}

func TestGetWindData_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{"error": "server broke"}`)
	}))
	defer server.Close()

	originalURL := weather.OpenMeteoURL
	weather.OpenMeteoURL = server.URL
	defer func() { weather.OpenMeteoURL = originalURL }()

	client := weather.NewClient()
	windInfo, err := client.GetWindData(52.52, 13.41)

	require.Error(t, err)
	assert.Nil(t, windInfo)
	assert.Contains(t, err.Error(), "weather API request failed with status: 500 Internal Server Error")
}

func TestGetWindData_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `this is not json`)
	}))
	defer server.Close()

	originalURL := weather.OpenMeteoURL
	weather.OpenMeteoURL = server.URL
	defer func() { weather.OpenMeteoURL = originalURL }()

	client := weather.NewClient()
	windInfo, err := client.GetWindData(52.52, 13.41)

	require.Error(t, err)
	assert.Nil(t, windInfo)
	assert.Contains(t, err.Error(), "failed to decode weather API response")
}

func TestGetWindData_IncompleteData(t *testing.T) {
	testCases := []struct {
		name         string
		responseData string
	}{
		{
			name:         "Missing time",
			responseData: `{"hourly": {"windspeed_10m": [1.0], "winddirection_10m": [1.0]}}`,
		},
		{
			name:         "Missing speed",
			responseData: `{"hourly": {"time": ["2024-01-01T00:00:00Z"], "winddirection_10m": [1.0]}}`,
		},
		{
			name:         "Missing direction",
			responseData: `{"hourly": {"time": ["2024-01-01T00:00:00Z"], "windspeed_10m": [1.0]}}`,
		},
		{
			name:         "Empty time array",
			responseData: `{"hourly": {"time": [], "windspeed_10m": [1.0], "winddirection_10m": [1.0]}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, tc.responseData)
			}))
			defer server.Close()

			originalURL := weather.OpenMeteoURL
			weather.OpenMeteoURL = server.URL
			defer func() { weather.OpenMeteoURL = originalURL }()

			client := weather.NewClient()
			windInfo, err := client.GetWindData(52.52, 13.41)

			require.Error(t, err)
			assert.Nil(t, windInfo)
			assert.Contains(t, err.Error(), "incomplete hourly data received")
		})
	}
}

func TestGetWindData_InvalidTimestamp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{
			"hourly": {
				"time": ["invalid-time-format"],
				"windspeed_10m": [18.0],
				"winddirection_10m": [270.0]
			}
		}`)
	}))
	defer server.Close()

	originalURL := weather.OpenMeteoURL
	weather.OpenMeteoURL = server.URL
	defer func() { weather.OpenMeteoURL = originalURL }()

	client := weather.NewClient()
	windInfo, err := client.GetWindData(52.52, 13.41)

	require.Error(t, err)
	assert.Nil(t, windInfo)
	assert.Contains(t, err.Error(), "failed to parse timestamp")
}
