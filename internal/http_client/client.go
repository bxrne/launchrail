package http_client

import (
	"bytes"
	"net/http"
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

// NewHTTPClient creates a new HTTPClient.
func NewHTTPClient() HTTPClient {
	return &DefaultHTTPClient{}
}
