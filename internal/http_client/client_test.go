package http_client_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/stretchr/testify/assert"
)

func TestDefaultHTTPClient_Post_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := &http_client.DefaultHTTPClient{}
	url := "https://postman-echo.com/post"
	contentType := "application/json"
	body := bytes.NewBuffer([]byte(`{"test":"data"}`))

	resp, err := client.Post(url, contentType, body)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
}
