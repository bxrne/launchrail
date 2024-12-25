package http_client

import (
	"bytes"
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of HTTPClient.
type MockHTTPClient struct {
	mock.Mock
}

// Post makes an HTTP POST request.
func (m *MockHTTPClient) Post(url, contentType string, body *bytes.Buffer) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).(*http.Response), args.Error(1)
}
