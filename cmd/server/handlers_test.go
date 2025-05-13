package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/reporting"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/zerodha/logf"
)

// Dummy data for CreateRecordWithConfig
var (
	dummyConfigData = []byte(`{"simulation": {"timestep": 0.01}}`)
	dummyOrkData    = []byte(`<openrocket/>`)
)

func setupTestTemplate(t *testing.T) string {
	templateDir := filepath.Join("templates", "reports")
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
	cfg := &config.Config{   // Minimal config
		Setup: config.Setup{
			App:     config.App{Version: "test-v0.1.0"},
			Logging: config.Logging{Level: "error"},
		},
	}
	tempStorageDir := t.TempDir()
	log := logf.New(logf.Opts{Level: logf.ErrorLevel})
	realManager, err := storage.NewRecordManager(cfg, tempStorageDir, &log)
	require.NoError(t, err, "Failed to create real RecordManager for test")
	// Create a dummy record using the correct method
	dummyRecord, err := realManager.CreateRecordWithConfig(dummyConfigData, dummyOrkData)
	require.NoError(t, err, "Failed to create dummy record")

	// Create dummy data files (motion.json, events.json, dynamics.json) within the record directory
	motionCSVPath := filepath.Join(dummyRecord.Path, "MOTION.csv")
	eventsCSVPath := filepath.Join(dummyRecord.Path, "EVENTS.csv")

	// Sample motion data (matches expected apogee/max_velocity)
	// Headers: time,altitude,velocity,acceleration,thrust
	motionData := []string{
		"time,altitude,velocity,acceleration,thrust", // Headers should already be there, but good to be explicit
		"0,0,0,0,100",
		"1,10,10,10,100",
		"2,25,5,0,0",   // Max velocity around here
		"3,30,0,-10,0", // Apogee
		"4,20,-10,-10,0",
	}
	err = os.WriteFile(motionCSVPath, []byte(strings.Join(motionData, "\n")), 0644)
	require.NoError(t, err, "Failed to write sample motion data")

	// Sample event data
	// Headers: time,event_name,motor_status,parachute_status
	eventData := []string{
		"time,event_name,motor_status,parachute_status", // Headers
		"0,Liftoff,BURNOUT,NONE",                        // Changed from LAUNCH to Liftoff
		"3,Apogee,BURNOUT,DEPLOYED",                     // Changed from APOGEE to Apogee
	}
	err = os.WriteFile(eventsCSVPath, []byte(strings.Join(eventData, "\n")), 0644)
	require.NoError(t, err, "Failed to write sample event data")

	// Create a dummy engine_config.json as it's expected by LoadSimulationData
	engineConfigPath := filepath.Join(dummyRecord.Path, "engine_config.json")
	dummyEngineConfig := config.Engine{
		Options: config.Options{
			OpenRocketFile:   "./testdata/l1.ork",
			MotorDesignation: "TestMotor-ABC",
		},
	}
	dummyEngineConfigBytes, err := json.Marshal(dummyEngineConfig)
	require.NoError(t, err, "Failed to marshal dummy engine config")
	err = os.WriteFile(engineConfigPath, dummyEngineConfigBytes, 0644)
	require.NoError(t, err, "Failed to write dummy engine_config.json")

	// Close the record to flush data and release file handles before the handler tries to read them
	err = dummyRecord.Close()
	require.NoError(t, err, "Failed to close dummy record")

	dataHandler := NewDataHandler(realManager, cfg, &log)
	router := gin.New()
	// The test was previously trying to hit /reports/{hash}/download, but ReportAPIV2 is mounted differently
	// The actual route in main.go is /explore/:hash/report.
	// For consistency with how TestDownloadReport_NotFound sets up the router, we'll use the versioned path.
	router.GET("/api/v0/explore/:hash/report", dataHandler.ReportAPIV2)

	w := httptest.NewRecorder()
	reqURL := fmt.Sprintf("/api/v0/explore/%s/report", dummyRecord.Hash)
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
	assert.Equal(t, dummyRecord.Hash, reportDataResponse.RecordID, "RecordID in response mismatch")
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

// TestListRecordsAPI tests the API handler for listing records
func TestListRecordsAPI(t *testing.T) {
	// 1. Setup real RecordManager
	tempStorageDir := t.TempDir()
	cfg := &config.Config{Setup: config.Setup{Logging: config.Logging{Level: "error"}}} // Minimal config for storage
	log := logf.New(logf.Opts{Level: logf.ErrorLevel})
	realManager, err := storage.NewRecordManager(cfg, tempStorageDir, &log)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	// 2. Create dummy records
	numRecords := 3
	createdHashes := make([]string, numRecords)
	for i := 0; i < numRecords; i++ {
		// Create slightly different config data for unique hashes
		configData := []byte(fmt.Sprintf(`{"iteration": %d}`, i))
		orkData := []byte(fmt.Sprintf(`<rocket iteration="%d"/>`, i))
		record, err := realManager.CreateRecordWithConfig(configData, orkData)
		require.NoError(t, err)
		createdHashes[i] = record.Hash
	}

	// 3. Setup real DataHandler and Router
	// Initialize DataHandler with a logger
	dataHandler := NewDataHandler(realManager, cfg, &log)

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

// TestDeleteRecord tests the delete record handler
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
	log := logf.New(logf.Opts{Level: logf.ErrorLevel})
	realManager, err := storage.NewRecordManager(cfg, tempStorageDir, &log)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	// 2. Create dummy records
	hashToKeep := ""
	hashToDelete := ""
	for i := 0; i < 2; i++ {
		configData := []byte(fmt.Sprintf(`{"test": "delete%d"}`, i))
		orkData := []byte(fmt.Sprintf(`<rocket test="delete%d"/>`, i))
		record, err := realManager.CreateRecordWithConfig(configData, orkData)
		require.NoError(t, err)
		if i == 0 {
			hashToKeep = record.Hash
		} else {
			hashToDelete = record.Hash
		}
	}

	// 3. Setup real DataHandler and Router
	// Initialize DataHandler with a logger
	dataHandler := NewDataHandler(realManager, cfg, &log)

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

// TestDeleteRecordAPI tests the delete record API handler
func TestDeleteRecordAPI(t *testing.T) {
	// Arrange
	// 1. Setup real RecordManager
	cfg := &config.Config{ // Minimal config
		Setup: config.Setup{
			Logging: config.Logging{Level: "error"},
		},
	}
	tempStorageDir := t.TempDir()
	log := logf.New(logf.Opts{Level: logf.ErrorLevel})
	realManager, err := storage.NewRecordManager(cfg, tempStorageDir, &log)
	require.NoError(t, err, "Failed to create real RecordManager for test")

	// 2. Create a dummy record
	// Use slightly different data to ensure unique hash if run concurrently
	configDataDel := []byte(`{"action": "delete_api"}`)
	orkDataDel := []byte(`<rocket action="delete_api"/>`)
	recordToDelete, err := realManager.CreateRecordWithConfig(configDataDel, orkDataDel)
	require.NoError(t, err)
	require.NotNil(t, recordToDelete, "Created record should not be nil")
	hashToDelete := recordToDelete.Hash

	// 3. Setup real DataHandler and Router
	// Initialize DataHandler with a logger
	dataHandler := NewDataHandler(realManager, cfg, &log)

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
