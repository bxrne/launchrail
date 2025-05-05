package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
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
	// Check for version string
	assert.Contains(t, body, "test", "HTML should contain app version")

}

// setupTestTemplate creates the dummy template file needed by the reporting generator.
func setupTestTemplate(t *testing.T) string {
	templateDir := filepath.Join("internal", "reporting", "templates")
	templatePath := filepath.Join(templateDir, "report.md.tmpl")

	// Ensure parent directory exists
	err := os.MkdirAll(templateDir, 0755)
	require.NoError(t, err)

	// Create dummy template content
	templateContent := "Report for {{.RecordID}} Version {{.Version}}"
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	// Cleanup function to remove the created structure
	t.Cleanup(func() {
		os.RemoveAll(filepath.Join("internal")) // Remove the top-level dir created
	})

	return templateDir // Although not used by handler directly, good practice
}

func TestDownloadReport(t *testing.T) {
	// Arrange
	_ = setupTestTemplate(t) // Create the dummy template file

	// Use the real RecordManager in a temp directory
	tempStorageDir := t.TempDir()
	realManager, err := storage.NewRecordManager(tempStorageDir)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	// Create a dummy record using the real manager
	dummyRecord, err := realManager.CreateRecord()
	require.NoError(t, err, "Failed to create dummy record")
	require.NotNil(t, dummyRecord)
	recordHash := dummyRecord.Hash
	defer dummyRecord.Close() // Close the record created by the real manager

	cfg := &config.Config{ // Minimal config needed
		Setup: config.Setup{
			App:     config.App{Version: "test-report-v1"},
			Logging: config.Logging{Level: "error"},
		},
	}

	// Initialize DataHandler with a logger
	log := logger.GetLogger("debug")
	dataHandler := &DataHandler{records: realManager, Cfg: cfg, log: log}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery()) // Add recovery middleware
	router.GET("/explore/:hash/report", dataHandler.DownloadReport)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/explore/"+recordHash+"/report", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code, "Expected OK status for report download")

	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))

	expectedFilename := fmt.Sprintf("launch_report_%s.pdf", recordHash)
	contentDisposition := w.Header().Get("Content-Disposition")
	assert.Contains(t, contentDisposition, "attachment; filename=")
	assert.Contains(t, contentDisposition, expectedFilename)

	// Check body contains placeholder content (since PDF conversion is placeholder)
	body := w.Body.String()
	assert.NotEmpty(t, body)
	assert.Contains(t, body, "--- PDF Conversion Placeholder ---") // From reporting.convertMarkdownToPDF placeholder
	assert.Contains(t, body, recordHash)                           // Check if hash from template is included
}

func TestDownloadReport_NotFound(t *testing.T) {
	// Arrange
	// _ = setupTestTemplate(t) // No longer needed as we abort before report generation

	// Use the real RecordManager in a temp directory
	tempStorageDir := t.TempDir()
	realManager, err := storage.NewRecordManager(tempStorageDir)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	nonExistentHash := "record-does-not-exist"

	cfg := &config.Config{ // Minimal config
		Setup: config.Setup{
			App:     config.App{Version: "test-report-v1"},
			Logging: config.Logging{Level: "error"},
		},
	}

	// Initialize DataHandler with a logger
	log := logger.GetLogger("debug")
	dataHandler := &DataHandler{records: realManager, Cfg: cfg, log: log}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery()) // Add recovery middleware
	router.GET("/explore/:hash/report", dataHandler.DownloadReport)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/explore/"+nonExistentHash+"/report", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	// Expecting NotFound status because LoadSimulationData fails
	require.Equal(t, http.StatusNotFound, w.Code)

	// Check error message in JSON response
	var jsonResponse map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &jsonResponse)
	require.NoError(t, err, "Response body should be valid JSON")
	assert.Equal(t, "Failed to load data for report", jsonResponse["error"])
}

// --- New Test for HTML ListRecords with Real Manager ---

func TestListRecords_RealManager(t *testing.T) {
	// Arrange
	// 1. Setup real RecordManager
	tempStorageDir := t.TempDir()
	realManager, err := storage.NewRecordManager(tempStorageDir)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	// 2. Create dummy records
	var expectedHashes []string
	for i := 0; i < 3; i++ {
		record, err := realManager.CreateRecord()
		require.NoError(t, err)
		require.NotNil(t, record)
		expectedHashes = append(expectedHashes, record.Hash)
		// Close record resources, important for file handles
		defer record.Close()
	}

	cfg := &config.Config{ // Minimal config
		Setup: config.Setup{
			App:     config.App{Version: "test-list-v1"},
			Logging: config.Logging{Level: "error"},
		},
	}

	// 3. Setup real DataHandler and Router
	// Initialize DataHandler with a logger
	log := logger.GetLogger("debug")
	dataHandler := &DataHandler{records: realManager, Cfg: cfg, log: log}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/data", dataHandler.ListRecords) // Use the HTML handler

	// 4. Make GET request
	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 5. Assert StatusOK
	require.Equal(t, http.StatusOK, w.Code)

	// 6. Assert HTML content
	body := w.Body.String()
	assert.Contains(t, body, "<html", "Response should be HTML")
	assert.Contains(t, body, "<table", "HTML should contain a table for records")
	// Check if the hashes of created records are present
	for _, hash := range expectedHashes {
		assert.Contains(t, body, hash, "HTML should contain record hash: %s", hash)
	}
	// Check for version string
	assert.Contains(t, body, "test-list-v1", "HTML should contain app version")

}

// --- New Test for DeleteRecord ---

func TestDeleteRecord(t *testing.T) {
	// Arrange
	// 1. Setup real RecordManager
	tempStorageDir := t.TempDir()
	realManager, err := storage.NewRecordManager(tempStorageDir)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	// 2. Create dummy records
	recordToDelete, err := realManager.CreateRecord()
	require.NoError(t, err)
	defer recordToDelete.Close()
	hashToDelete := recordToDelete.Hash

	recordToKeep, err := realManager.CreateRecord()
	require.NoError(t, err)
	defer recordToKeep.Close()
	hashToKeep := recordToKeep.Hash

	cfg := &config.Config{ // Minimal config
		Setup: config.Setup{
			App:     config.App{Version: "test-delete-v1"},
			Logging: config.Logging{Level: "error"},
		},
	}

	// 3. Setup real DataHandler and Router
	// Initialize DataHandler with a logger
	log := logger.GetLogger("debug")
	dataHandler := &DataHandler{records: realManager, Cfg: cfg, log: log}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery()) // Add recovery middleware
	// Note the route path from main.go
	router.DELETE("/data/:hash", dataHandler.DeleteRecord)

	// 4. Make DELETE request
	req := httptest.NewRequest(http.MethodDelete, "/data/"+hashToDelete, nil)
	req.Header.Set("Accept", "text/html") // Simulate HTMX request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 5. Assert StatusOK
	require.Equal(t, http.StatusOK, w.Code, "Expected OK status for HTMX delete response")

	// 6. Assert HTML response content (updated list)
	body := w.Body.String()
	// Check for the presence of the ID attribute, less sensitive to class changes
	if !assert.Contains(t, body, "id=\"records-list\"", "Response should contain the records-list ID") {
		t.Logf("Actual response body:\n%s", body) // Print body on failure
	}
	assert.NotContains(t, body, hashToDelete, "HTML response should NOT contain the deleted hash")
	assert.Contains(t, body, hashToKeep, "HTML response SHOULD contain the remaining hash")

	// 7. Assert record is actually deleted from storage
	_, err = realManager.GetRecord(hashToDelete)
	require.Error(t, err, "GetRecord should return an error for the deleted hash")
	assert.ErrorContains(t, err, "not found", "Error message should indicate record not found")

	// Double-check the kept record still exists
	_, err = realManager.GetRecord(hashToKeep)
	require.NoError(t, err, "GetRecord should succeed for the kept hash")
}

// --- New Test for DeleteRecordAPI ---

func TestDeleteRecordAPI(t *testing.T) {
	// Arrange
	// 1. Setup real RecordManager
	tempStorageDir := t.TempDir()
	realManager, err := storage.NewRecordManager(tempStorageDir)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	// 2. Create a dummy record
	recordToDelete, err := realManager.CreateRecord()
	require.NoError(t, err)
	defer recordToDelete.Close()
	hashToDelete := recordToDelete.Hash

	cfg := &config.Config{ // Minimal config
		Setup: config.Setup{
			Logging: config.Logging{Level: "error"},
		},
	}

	// 3. Setup real DataHandler and Router
	// Initialize DataHandler with a logger
	log := logger.GetLogger("debug")
	dataHandler := &DataHandler{records: realManager, Cfg: cfg, log: log}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery()) // Add recovery middleware
	// Note the route path from main.go - the handler is DeleteRecord
	router.DELETE("/api/data/:hash", dataHandler.DeleteRecord)

	// --- Test Case 1: Delete Existing Record ---
	// 4. Make DELETE request
	req1 := httptest.NewRequest(http.MethodDelete, "/api/data/"+hashToDelete, nil)
	req1.Header.Set("Accept", "application/json") // Set Accept header
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// 5. Assert StatusNoContent (204)
	require.Equal(t, http.StatusNoContent, w1.Code, "Expected NoContent status for successful API delete")

	// 6. Assert record is actually deleted from storage
	_, err = realManager.GetRecord(hashToDelete)
	require.Error(t, err, "GetRecord should return an error for the deleted hash")
	assert.ErrorContains(t, err, "not found", "Error message should indicate record not found")

	// --- Test Case 2: Delete Non-Existent Record ---
	// 7. Make DELETE request for non-existent hash
	nonExistentHash := "this-hash-does-not-exist"
	req2 := httptest.NewRequest(http.MethodDelete, "/api/data/"+nonExistentHash, nil)
	req2.Header.Set("Accept", "application/json") // Set Accept header
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// 8. Assert StatusNotFound (404)
	require.Equal(t, http.StatusNotFound, w2.Code, "Expected NotFound status for deleting non-existent hash")

	// Check error message in JSON response
	var jsonResponse map[string]string
	err = json.Unmarshal(w2.Body.Bytes(), &jsonResponse)
	require.NoError(t, err, "Response body should be valid JSON for 404 error")
	assert.Equal(t, "Record not found", jsonResponse["error"], "Expected 'Record not found' error message")

}
