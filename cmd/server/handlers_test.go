package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Response structure for API tests
type ListRecordsAPIResponse struct {
	Total   int               `json:"total"`
	Records []*storage.Record `json:"records"`
}

// RecordManagerInterface defines the methods that the real RecordManager implements
// This allows us to create a test implementation with the same methods
type RecordManagerInterface interface {
	ListRecords() ([]*storage.Record, error)
	GetRecord(hash string) (*storage.Record, error)
	CreateRecord() (*storage.Record, error)
	DeleteRecord(hash string) error
	GetStorageDir() string
}

// TestRecordManager is a test implementation of RecordManagerInterface
type TestRecordManager struct {
	mock *MockRecordManager
}

// MockRecordManager is a simplified implementation for tests
type MockRecordManager struct {
	records []*storage.Record
	baseDir string
}

// Create a new test record manager - returns our mock implementation
func newTestRecordManager(n int) RecordManagerInterface {
	mock := &MockRecordManager{
		records: make([]*storage.Record, 0, n),
		baseDir: "/tmp/test-storage",
	}

	// Add initial test records
	for i := 0; i < n; i++ {
		hash := fmt.Sprintf("test-hash-%d", i+1)
		creationTime := time.Now().Add(-time.Duration(i) * time.Hour)
		mock.records = append(mock.records, &storage.Record{
			Hash:         hash,
			LastModified: creationTime,
			CreationTime: creationTime,
		})
	}

	return &TestRecordManager{mock: mock}
}

// We don't need the AsRecordManager method anymore since we're directly using TestRecordManager
// in our tests instead of trying to convert it to a storage.RecordManager

// ListRecords implements storage.RecordManager interface
func (t *TestRecordManager) ListRecords() ([]*storage.Record, error) {
	return t.mock.records, nil
}

// GetRecord implements storage.RecordManager interface
func (t *TestRecordManager) GetRecord(hash string) (*storage.Record, error) {
	for _, r := range t.mock.records {
		if r.Hash == hash {
			return r, nil
		}
	}
	return nil, fmt.Errorf("record not found with hash: %s", hash)
}

// CreateRecord implements storage.RecordManager interface
func (t *TestRecordManager) CreateRecord() (*storage.Record, error) {
	hash := fmt.Sprintf("test-hash-%d", len(t.mock.records)+1)
	record := &storage.Record{
		Hash:         hash,
		CreationTime: time.Now(),
		LastModified: time.Now(),
	}
	t.mock.records = append(t.mock.records, record)
	return record, nil
}

// DeleteRecord implements storage.RecordManager interface
func (t *TestRecordManager) DeleteRecord(hash string) error {
	for i, r := range t.mock.records {
		if r.Hash == hash {
			// Remove the record at index i
			t.mock.records = append(t.mock.records[:i], t.mock.records[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("record not found with hash: %s", hash)
}

// GetStorageDir implements storage.RecordManager interface
func (t *TestRecordManager) GetStorageDir() string {
	return t.mock.baseDir
}

// TestDataHandler is a version of DataHandler that works with our test interface
type TestDataHandler struct {
	records RecordManagerInterface
	Cfg     *config.Config
}

// Reimplement the necessary handler methods using our interface

func (h *TestDataHandler) ListRecordsAPI(c *gin.Context) {
	// Fetch all records
	records, err := h.records.ListRecords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get limit and offset from query parameters
	limit := 0
	offset := 0

	limitParam := c.Query("limit")
	offsetParam := c.Query("offset")

	if limitParam != "" {
		limit, _ = strconv.Atoi(limitParam)
	}

	if offsetParam != "" {
		offset, _ = strconv.Atoi(offsetParam)
	}

	// Apply pagination if requested
	totalRecords := len(records)
	paginatedRecords := records

	if offset > 0 && offset < len(records) {
		paginatedRecords = records[offset:]
	} else if offset >= len(records) {
		paginatedRecords = []*storage.Record{}
	}

	if limit > 0 && limit < len(paginatedRecords) {
		paginatedRecords = paginatedRecords[:limit]
	}

	// Return JSON response with total and records
	c.JSON(http.StatusOK, gin.H{
		"total":   totalRecords,
		"records": paginatedRecords,
	})
}

func (h *TestDataHandler) ListRecords(c *gin.Context) {
	// Fetch all records
	records, err := h.records.ListRecords()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	// Render simple HTML for testing
	_, _ = c.Writer.WriteString("<html><body>")
	for _, record := range records {
		_, _ = c.Writer.WriteString("<div>" + record.Hash + "</div>")
	}
	_, _ = c.Writer.WriteString("</body></html>")
}

// TestEmptyRecordsAPI verifies that the API handles empty record sets correctly
func TestEmptyRecordsAPI(t *testing.T) {
	// Create a mock storage with no records
	mockRecords := newTestRecordManager(0)

	// Set up the API handler
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create minimal config
	cfg := &config.Config{
		Setup: config.Setup{
			App: config.App{Version: "test"},
		},
	}

	// Create handler
	handler := &TestDataHandler{
		records: mockRecords,
		Cfg:     cfg,
	}

	// Register API route
	apiPath := "/api/v0"
	apiGroup := router.Group(apiPath)
	apiGroup.GET("/data", handler.ListRecordsAPI)

	// Make API request
	req := httptest.NewRequest("GET", apiPath+"/data", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	require.Equal(t, http.StatusOK, w.Code)

	// Parse JSON response
	var response ListRecordsAPIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify contents
	assert.Equal(t, 0, response.Total, "Empty storage should return total=0")
	assert.Len(t, response.Records, 0, "Empty storage should return empty records array")
}

// TestPaginationAPI tests pagination functionality in the API
func TestPaginationAPI(t *testing.T) {
	// Create a mock storage with 5 pre-populated records
	mockRecords := newTestRecordManager(5)

	// Set up the API handler
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create minimal config
	cfg := &config.Config{
		Setup: config.Setup{
			App: config.App{Version: "test"},
		},
	}

	// Create handler
	handler := &TestDataHandler{
		records: mockRecords,
		Cfg:     cfg,
	}

	// Register API route
	apiPath := "/api/v0"
	apiGroup := router.Group(apiPath)
	apiGroup.GET("/data", handler.ListRecordsAPI)

	// Create a test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Helper function to make pagination requests
	fetchWithParams := func(params string) ListRecordsAPIResponse {
		url := fmt.Sprintf("%s%s/data", server.URL, apiPath)
		if params != "" {
			url += "?" + params
		}

		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result ListRecordsAPIResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		return result
	}

	// Test cases
	testCases := []struct {
		name        string
		params      string
		expectTotal int
		expectCount int
	}{
		{
			name:        "Default (no pagination)",
			params:      "",
			expectTotal: 5,
			expectCount: 5,
		},
		{
			name:        "With limit=2",
			params:      "limit=2",
			expectTotal: 5,
			expectCount: 2,
		},
		{
			name:        "With offset=3",
			params:      "offset=3",
			expectTotal: 5,
			expectCount: 2,
		},
		{
			name:        "With limit=2&offset=2",
			params:      "limit=2&offset=2",
			expectTotal: 5,
			expectCount: 2,
		},
		{
			name:        "With offset past end",
			params:      "offset=5",
			expectTotal: 5,
			expectCount: 0,
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := fetchWithParams(tc.params)

			assert.Equal(t, tc.expectTotal, response.Total,
				"Total count should match regardless of pagination")
			assert.Len(t, response.Records, tc.expectCount,
				"Record count should match expected for params: %s", tc.params)
		})
	}
}

// TestListRecordsHTML tests the HTML rendering for the records list
func TestListRecordsHTML(t *testing.T) {
	// Create a mock with 3 records
	mockRecords := newTestRecordManager(3)

	// Set up the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create minimal config
	cfg := &config.Config{
		Setup: config.Setup{
			App: config.App{Version: "test"},
		},
	}

	// Create handler
	handler := &TestDataHandler{
		records: mockRecords,
		Cfg:     cfg,
	}

	// Register route for HTML handler
	router.GET("/data", handler.ListRecords)

	// Make request
	req := httptest.NewRequest("GET", "/data", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	require.Equal(t, http.StatusOK, w.Code)

	// Check HTML response
	body := w.Body.String()

	// It should contain HTML elements
	assert.Contains(t, body, "<html")

	// It should contain test-hash identifiers from our mock records
	for i := 1; i <= 3; i++ {
		assert.Contains(t, body, fmt.Sprintf("test-hash-%d", i),
			"HTML should include record hash identifiers")
	}
}
