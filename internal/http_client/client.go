package http_client

import (
	"bytes"
	"net/http"

	"github.com/stretchr/testify/mock"
)

// HTTPClient is an interface for making HTTP requests.
type HTTPClient interface {
	Post(url, contentType string, body *bytes.Buffer) (*http.Response, error)
}

// DefaultHTTPClient is the default implementation of HTTPClient.
type DefaultHTTPClient struct{}

// Post makes an HTTP POST request.
func (c *DefaultHTTPClient) Post(url, contentType string, body *bytes.Buffer) (*http.Response, error) {
	return http.Post(url, contentType, body)
}

// MockHTTPClient is a mock implementation of HTTPClient.
type MockHTTPClient struct {
	mock.Mock
}

// Post makes an HTTP POST request.
func (m *MockHTTPClient) Post(url, contentType string, body *bytes.Buffer) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).(*http.Response), args.Error(1)
}
