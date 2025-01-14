package http_client_test

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/bxrne/launchrail/internal/http_client"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a valid configuration file WHEN GetConfig is called THEN no error is returned
func TestMockHTTPClient_Post(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		contentType string
		body        *bytes.Buffer
		mockResp    *http.Response
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful request",
			url:         "http://example.com",
			contentType: "application/json",
			body:        bytes.NewBuffer([]byte(`{"key":"value"}`)),
			mockResp:    &http.Response{StatusCode: http.StatusOK},
			mockErr:     nil,
			wantErr:     false,
		},
		{
			name:        "failed request",
			url:         "http://example.com",
			contentType: "application/json",
			body:        bytes.NewBuffer([]byte(`{"key":"value"}`)),
			mockResp:    nil,
			mockErr:     errors.New("network error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(http_client.MockHTTPClient)
			mockClient.On("Post", tt.url, tt.contentType, tt.body).
				Return(tt.mockResp, tt.mockErr)

			resp, err := mockClient.Post(tt.url, tt.contentType, tt.body)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResp, resp)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
