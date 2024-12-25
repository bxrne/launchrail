package http_client_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/stretchr/testify/assert"
)

// MockHTTPClient is a mock implementation of the HTTPClient interface for testing.
type MockHTTPClient struct {
	Response *http.Response
	Err      error
}

func (m *MockHTTPClient) Post(url, contentType string, body *bytes.Buffer) (*http.Response, error) {
	return m.Response, m.Err
}

// TEST: GIVEN a DefaultHTTPClient WHEN making a POST request THEN the request is successful.
func TestDefaultHTTPClient_Post(t *testing.T) {
	client := &http_client.DefaultHTTPClient{}

	// Mock server
	server := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, `{"key":"value"}`, string(body))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"response":"ok"}`))
	})
	testServer := http.NewServeMux()
	testServer.Handle("/test", server)
	httpServer := &http.Server{
		Addr:    "127.0.0.1:8081",
		Handler: testServer,
	}
	go httpServer.ListenAndServe()
	defer httpServer.Close()

	// Request
	reqBody := bytes.NewBuffer([]byte(`{"key":"value"}`))
	resp, err := client.Post("http://127.0.0.1:8081/test", "application/json", reqBody)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"response":"ok"}`, string(respBody))
}

// TEST: GIVEN a MockHTTPClient WHEN making a POST request THEN the request is successful.
func TestMockHTTPClient_Post(t *testing.T) {
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"response":"mocked"}`)),
	}
	mockClient := &MockHTTPClient{
		Response: mockResp,
		Err:      nil,
	}

	reqBody := bytes.NewBuffer([]byte(`{"key":"value"}`))
	resp, err := mockClient.Post("http://mock.url/test", "application/json", reqBody)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"response":"mocked"}`, string(respBody))
}
