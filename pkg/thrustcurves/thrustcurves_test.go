package thrustcurves_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock HTTP Client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Post(url, contentType string, body *bytes.Buffer) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestGetMotorID_ValidResponse(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	thrustcurves.SetHTTPClient(mockHTTP)

	mockResponse := `{"results":[{"motorId":"motor123"}]}`
	mockHTTP.On("Post", "https://www.thrustcurve.org/api/v1/search.json", "application/json", mock.Anything).
		Return(http.Response{Body: io.NopCloser(bytes.NewBufferString(mockResponse))}, nil)

	id, err := thrustcurves.getMotorID("269H110-14A")
	assert.NoError(t, err)
	assert.Equal(t, "motor123", id)
}

func TestGetMotorCurve_ValidResponse(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	thrustcurves.SetHTTPClient(mockHTTP)

	mockResponse := `{"results":[{"samples":[{"time":0.1,"thrust":10.0},{"time":0.2,"thrust":20.0}]}]}`
	mockHTTP.On("Post", "https://www.thrustcurve.org/api/v1/download.json", "application/json", mock.Anything).
		Return(http.Response{Body: io.NopCloser(bytes.NewBufferString(mockResponse))}, nil)

	curve, err := thrustcurves.getMotorCurve("motor123")
	assert.NoError(t, err)
	assert.Equal(t, [][]float64{{0.1, 10.0}, {0.2, 20.0}}, curve)
}
