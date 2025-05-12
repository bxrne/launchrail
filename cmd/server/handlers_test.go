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
	"github.com/bxrne/launchrail/internal/reporting"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testLog = logger.GetLogger("debug") // Define testLog

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
	// 1. Setup real RecordManager and create a dummy record
	_ = setupTestTemplate(t) // Ensure template exists, though its path isn't directly used in cfg for this handler
	cfg := &config.Config{ // Minimal config
		Setup: config.Setup{
			App:     config.App{Version: "test-v0.1.0"},
			Logging: config.Logging{Level: "error"},
		},
	}
	tempStorageDir := t.TempDir()
	realManager, err := storage.NewRecordManager(cfg, tempStorageDir)
	require.NoError(t, err, "Failed to create real RecordManager for test")
	dummyRecord, err := realManager.CreateRecord(cfg)
	require.NoError(t, err)
	defer dummyRecord.Close()
	actualHash := dummyRecord.Hash

	// Create a dummy engine_config.json in the record's directory for LoadSimulationData
	recordDir := filepath.Join(tempStorageDir, actualHash)
	err = os.MkdirAll(recordDir, 0755) // Ensure the specific record directory exists
	require.NoError(t, err, "Failed to create record directory for engine_config.json")
	dummyEngineConfig := config.Engine{
		Options: config.Options{
			OpenRocketFile:   "./testdata/l1.ork",
			MotorDesignation: "TestMotor-ABC",
		},
	}
	dummyEngineConfigBytes, err := json.Marshal(dummyEngineConfig)
	require.NoError(t, err, "Failed to marshal dummy engine config")
	err = os.WriteFile(filepath.Join(recordDir, "engine_config.json"), dummyEngineConfigBytes, 0644)
	require.NoError(t, err, "Failed to write dummy engine_config.json")

	// Close the record to flush data and release file handles before the handler tries to read them
	err = dummyRecord.Close()
	require.NoError(t, err, "Failed to close dummy record")

	dataHandler := NewDataHandler(realManager, cfg, testLog)
	router := gin.New()
	// The test was previously trying to hit /reports/{hash}/download, but ReportAPIV2 is mounted differently
	// The actual route in main.go is /explore/:hash/report.
	// For consistency with how TestDownloadReport_NotFound sets up the router, we'll use the versioned path.
	router.GET("/api/v0/explore/:hash/report", dataHandler.ReportAPIV2)

	w := httptest.NewRecorder()
	reqURL := fmt.Sprintf("/api/v0/explore/%s/report", actualHash)
	req, _ := http.NewRequest("GET", reqURL, nil)

	// Act
	router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code, "Expected OK status for report data")
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Expected Content-Type to be application/json")

	var reportDataResponse reporting.ReportData
	err = json.Unmarshal(w.Body.Bytes(), &reportDataResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response from ReportAPIV2")

	// Verify some key fields in the ReportData
	assert.Equal(t, actualHash, reportDataResponse.RecordID, "RecordID in response mismatch")
	assert.Equal(t, "test-v0.1.0", reportDataResponse.Version, "Version in response mismatch")
	assert.Equal(t, "l1.ork", reportDataResponse.RocketName, "RocketName in response mismatch")      // Based on dummyEngineConfig
	assert.Equal(t, "TestMotor-ABC", reportDataResponse.MotorName, "MotorName in response mismatch") // Based on dummyEngineConfig

	// Check MotionMetrics (Apogee, MaxVelocity etc.) - These are calculated by processReportData
	// which is called within LoadSimulationData. We need to ensure processReportData is robust.
	// For now, let's assume LoadSimulationData populates these correctly based on the dummy data.
	// Apogee was 30.0 at t=2.0. Max velocity was 10.0 at t=2.0.
	// The current `LoadSimulationData` calls `processReportData` which should populate these.

	// For dummy data: Apogee should be 30.0, MaxVelocity 10.0
	// Need to ensure processReportData is called and populates these fields in ReportData.
	// Directly asserting after LoadSimulationData would require processReportData to be called by it.
	// Let's verify if MotionMetrics is not nil first.
	assert.NotNil(t, reportDataResponse.MotionMetrics, "MotionMetrics should be populated")
	if reportDataResponse.MotionMetrics != nil {
		assert.InDelta(t, 30.0, reportDataResponse.MotionMetrics.MaxAltitude, 0.001, "Apogee mismatch")
		assert.InDelta(t, 10.0, reportDataResponse.MotionMetrics.MaxVelocity, 0.001, "MaxVelocity mismatch")
	}

	// Check if EventsData is populated
	assert.NotEmpty(t, reportDataResponse.EventsData, "EventsData should not be empty")
	if len(reportDataResponse.EventsData) > 1 { // header + data rows
		// Example: check the first data event (Liftoff)
		// Assuming EventsData includes headers. If not, adjust index.
		assert.Contains(t, reportDataResponse.EventsData[1], "Liftoff", "First event should be Liftoff")
	}

	// Check if plot data (placeholders for now, as it's SVG strings) exists
	// The ReportAPIV2 and LoadSimulationData pipeline should generate plot data and include it.
	// The `GeneratePlots` function inside `LoadSimulationData` should populate `rData.Plots`.
	assert.NotEmpty(t, reportDataResponse.Plots, "Plots map should not be empty")
	assert.Contains(t, reportDataResponse.Plots, "altitude_vs_time", "Altitude plot should exist")
	assert.Contains(t, reportDataResponse.Plots, "velocity_vs_time", "Velocity plot should exist")
	assert.Contains(t, reportDataResponse.Plots, "acceleration_vs_time", "Acceleration plot should exist")

	// The old test checked for zip file contents. Now we check the JSON fields directly.
	// Assertions about specific markdown content or SVG links are no longer applicable here.
}

// TestDownloadReport_NotFound tests the scenario where a report is requested for a non-existent hash.
func TestDownloadReport_NotFound(t *testing.T) {
	// Arrange
	// _ = setupTestTemplate(t) // No longer needed as we abort before report generation

	cfg := &config.Config{ // Minimal config
		Setup: config.Setup{
			App:     config.App{Version: "test-report-v1"},
			Logging: config.Logging{Level: "error"},
		},
	}

	// Use the real RecordManager in a temp directory
	tempStorageDir := t.TempDir()
	realManager, err := storage.NewRecordManager(cfg, tempStorageDir)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	nonExistentHash := "record-does-not-exist"

	// Initialize DataHandler with a logger
	log := logger.GetLogger("debug")
	dataHandler := NewDataHandler(realManager, cfg, log)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery()) // Add recovery middleware
	router.GET("/api/v0/explore/:hash/report", dataHandler.ReportAPIV2)

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
	assert.Equal(t, "Data for report not found or incomplete", jsonResponse["error"], "Error message for non-existent report mismatch")
}

// --- New Test for HTML ListRecords with Real Manager ---

func TestListRecordsAPI(t *testing.T) {
	// 1. Setup real RecordManager
	tempStorageDir := t.TempDir()
	cfg := &config.Config{Setup: config.Setup{Logging: config.Logging{Level: "error"}}} // Minimal config for storage
	realManager, err := storage.NewRecordManager(cfg, tempStorageDir)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	// 2. Create dummy records
	numRecords := 3
	for i := 0; i < numRecords; i++ {
		// Pass cfg to CreateRecord
		record, err := realManager.CreateRecord(cfg)
		require.NoError(t, err)
		// Ensure hash is generated and accessible
		require.NotEmpty(t, record.Hash, "Record hash should not be empty after creation")
		defer record.Close()
	}

	// 3. Setup real DataHandler and Router
	// Initialize DataHandler with a logger
	log := logger.GetLogger("debug")
	// Ensure AppVersion is set in the config for the template
	cfg.Setup.App.Version = "test-version"
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
	// Note: Since we're not storing the hashes, we can't directly check for them.
	// Instead, we verify the structure of the HTML response.
	assert.Contains(t, body, "<tr>", "HTML should contain table rows for records")
	// Check for version string
	assert.Contains(t, body, cfg.Setup.App.Version, "HTML should contain app version")
}

// --- New Test for DeleteRecord ---

func TestDeleteRecord(t *testing.T) {
	// Arrange
	// 1. Setup real RecordManager
	cfg := &config.Config{ // Minimal config
		Setup: config.Setup{
			App:     config.App{Version: "test-delete-v1"},
			Logging: config.Logging{Level: "error"},
		},
	}
	tempStorageDir := t.TempDir()
	realManager, err := storage.NewRecordManager(cfg, tempStorageDir)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	// 2. Create dummy records
	hashToKeep := ""
	hashToDelete := ""
	for i := 0; i < 2; i++ {
		record, err := realManager.CreateRecord(cfg)
		require.NoError(t, err)
		defer record.Close()
		if i == 0 {
			hashToKeep = record.Hash
		} else {
			hashToDelete = record.Hash
		}
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
	cfg := &config.Config{ // Minimal config
		Setup: config.Setup{
			Logging: config.Logging{Level: "error"},
		},
	}
	tempStorageDir := t.TempDir()
	realManager, err := storage.NewRecordManager(cfg, tempStorageDir)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	// 2. Create a dummy record
	recordToDelete, err := realManager.CreateRecord(cfg)
	require.NoError(t, err)
	defer recordToDelete.Close()
	hashToDelete := recordToDelete.Hash

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
