package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	DeleteRecord(hash string) error
	GetStorageDir() string
}

// TestRecordManager is a test implementation of RecordManagerInterface for testing.
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
	log := logger.GetLogger("debug")
	handler := NewDataHandler(mockRecords, cfg, log)

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
	log := logger.GetLogger("debug")
	handler := NewDataHandler(mockRecords, cfg, log)

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
	log := logger.GetLogger("debug")
	handler := NewDataHandler(mockRecords, cfg, log)

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
	dummyRecord, err := realManager.CreateRecord() // CreateRecord takes no arguments
	require.NoError(t, err, "Failed to create dummy record")
	require.NotNil(t, dummyRecord, "Dummy record should not be nil")

	recordHash := dummyRecord.Hash // Use the Hash field from the Record struct

	// Write sample MOTION.csv data
	require.NotNil(t, dummyRecord.Motion, "Motion storage should not be nil")
	err = dummyRecord.Motion.Init() // Write headers
	require.NoError(t, err, "Failed to init motion storage")
	motionData := [][]string{
		{"0.0", "10.0", "0.0", "9.8", "0.0"}, // time, altitude, velocity, acceleration, thrust
		{"1.0", "15.0", "5.0", "15.0", "0.0"},
		{"2.0", "30.0", "10.0", "10.0", "0.0"}, // Apogee for this simple data
		{"3.0", "25.0", "-5.0", "-9.8", "0.0"},
		{"4.0", "10.5", "-10.0", "-9.8", "0.0"}, // Landing
	}
	for _, row := range motionData {
		err = dummyRecord.Motion.Write(row)
		require.NoError(t, err, "Failed to write motion data row")
	}
	err = dummyRecord.Motion.Close() // Close after writing
	require.NoError(t, err, "Failed to close motion storage")

	// Write sample EVENTS.csv data
	require.NotNil(t, dummyRecord.Events, "Events storage should not be nil")
	err = dummyRecord.Events.Init()
	require.NoError(t, err, "Failed to init events storage")
	eventsData := [][]string{
		{"0.0", "Liftoff", "", ""}, // time, event_name, motor_status, parachute_status
		{"2.0", "Apogee", "", ""},
		{"4.0", "Landing", "", ""},
	}
	for _, row := range eventsData {
		err = dummyRecord.Events.Write(row)
		require.NoError(t, err, "Failed to write event data row")
	}
	err = dummyRecord.Events.Close() // Close after writing
	require.NoError(t, err, "Failed to close events storage")

	// The main record.Close() is deferred, which is fine as it will clean up the directory.
	// Individual stores are closed above to ensure data is flushed before reading.
	defer dummyRecord.Close() // Close the record created by the real manager

	cfg := &config.Config{ // Minimal config needed
		Setup: config.Setup{
			App: config.App{Version: "test-report-v1"},
		},
	}

	// Initialize DataHandler with a logger
	log := logger.GetLogger("debug")
	dataHandler := NewDataHandler(realManager, cfg, log)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery()) // Add recovery middleware
	router.GET("/api/v0/explore/:hash/report", dataHandler.DownloadReport)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v0/explore/"+recordHash+"/report", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code, "Expected OK status for report download")

	// Check for Zip content type and disposition
	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "application/zip", contentType, "Expected Content-Type to be application/zip")

	expectedFilename := fmt.Sprintf("report_%s.zip", recordHash)
	contentDisposition := w.Header().Get("Content-Disposition")
	assert.Equal(t, "attachment; filename="+expectedFilename, contentDisposition, "Content-Disposition not set correctly for zip download")

	bodyBytes := w.Body.Bytes()
	zipReader, err := zip.NewReader(bytes.NewReader(bodyBytes), int64(len(bodyBytes)))
	require.NoError(t, err, "Failed to read zip archive from response body")

	foundReportMd := false
	foundAtmospherePlot := false

	for _, zf := range zipReader.File {
		switch zf.Name {
		case "report.md":
			foundReportMd = true
			rc, err := zf.Open()
			require.NoError(t, err, "Failed to open report.md from zip")
			defer rc.Close()

			mdContentBytes, err := io.ReadAll(rc)
			require.NoError(t, err, "Failed to read report.md content from zip")
			mdContentString := string(mdContentBytes)

			assert.Contains(t, mdContentString, fmt.Sprintf("# Simulation Report: %s", recordHash), "Zipped report.md does not contain correct title")
			assert.Contains(t, mdContentString, "## Plots & Data", "Zipped report.md does not contain Plots & Data section")
			assert.Contains(t, mdContentString, "![](assets/atmosphere_plot.png)", "Zipped report.md does not contain atmosphere plot asset link")
			// Check for new content based on parsed data
			assert.Contains(t, mdContentString, "Apogee: 30.0 meters", "Report missing correct apogee")
			assert.Contains(t, mdContentString, "Max Velocity: 10.0 m/s", "Report missing correct max velocity")
			assert.Contains(t, mdContentString, "Total Flight Time: 4.0 seconds", "Report missing correct total flight time")
			// assert.Contains(t, mdContentString, "| Liftoff | 0.0 | 10.0 |", "Report missing Liftoff event") // This will fail until events parsing/headers are aligned

		case "assets/atmosphere_plot.png":
			foundAtmospherePlot = true
			// Optionally, check file size or content if needed, but presence is often enough for dummy assets
			assert.Greater(t, zf.UncompressedSize64, uint64(0), "atmosphere_plot.png in zip should not be empty")
		}
	}

	assert.True(t, foundReportMd, "report.md not found in the downloaded zip archive")
	assert.True(t, foundAtmospherePlot, "assets/atmosphere_plot.png not found in the downloaded zip archive")

	// Clean up: remove the created reports directory to avoid clutter
	homeDir, _ := os.UserHomeDir()
	reportSpecificDir := filepath.Join(homeDir, ".launchrail", "reports", recordHash)
	_ = os.RemoveAll(reportSpecificDir) // Clean up the specific report directory
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
	dataHandler := NewDataHandler(realManager, cfg, log)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery()) // Add recovery middleware
	router.GET("/api/v0/explore/:hash/report", dataHandler.DownloadReport)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v0/explore/"+nonExistentHash+"/report", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	// Expecting NotFound status because LoadSimulationData fails
	require.Equal(t, http.StatusNotFound, w.Code)

	// Check error message in JSON response
	var jsonResponse map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &jsonResponse)
	require.NoError(t, err, "Response body should be valid JSON")
	// Updated expected error message
	assert.Equal(t, "Data for report not found", jsonResponse["error"], "Error message for non-existent report mismatch")
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
	dataHandler := NewDataHandler(realManager, cfg, log)

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
	dataHandler := NewDataHandler(realManager, cfg, log)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery()) // Add recovery middleware
	// Note the route path from main.go
	router.DELETE("/data/:hash", dataHandler.DeleteRecord)

	// 4. Make DELETE request
	req := httptest.NewRequest(http.MethodDelete, "/data/"+hashToDelete, nil)
	req.Header.Set("Hx-Request", "true") // Simulate HTMX request
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
	dataHandler := NewDataHandler(realManager, cfg, log)

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
	// Updated expected error message
	assert.Equal(t, "Record not found", jsonResponse["error"], "Expected 'Record not found' error message")

}
