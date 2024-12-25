package thrustcurves_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoadMotor_ValidResponse(t *testing.T) {
	mockHTTP := new(http_client.MockHTTPClient)

	mockSearchResponse := `{"results":[{"motorId":"motor123"}]}`
	mockHTTP.On("Post", "https://www.thrustcurve.org/api/v1/search.json", "application/json", mock.Anything).
		Return(&http.Response{Body: io.NopCloser(bytes.NewBufferString(mockSearchResponse))}, nil)

	mockDownloadResponse := `{"results":[{"samples":[{"time":0.1,"thrust":10.0},{"time":0.2,"thrust":20.0}]}]}`
	mockHTTP.On("Post", "https://www.thrustcurve.org/api/v1/download.json", "application/json", mock.Anything).
		Return(&http.Response{Body: io.NopCloser(bytes.NewBufferString(mockDownloadResponse))}, nil)

	motorData, err := thrustcurves.Load("269H110-14A", mockHTTP, &designation.DefaultDesignationValidator{})
	assert.NoError(t, err)
	assert.Equal(t, "motor123", motorData.ID)
	assert.Equal(t, [][]float64{{0.1, 10.0}, {0.2, 20.0}}, motorData.Thrust)
}

func TestLoadMotor_InvalidDesignation(t *testing.T) {
	mockHTTP := new(http_client.MockHTTPClient)

	motorData, err := thrustcurves.Load("<invalid>", mockHTTP, &designation.DefaultDesignationValidator{})
	assert.Error(t, err)
	assert.Nil(t, motorData)
}
