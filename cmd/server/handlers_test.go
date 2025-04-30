package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createMockRecord creates a dummy record directory and metadata file using RecordManager
func createMockRecord(rm *storage.RecordManager, baseDir, hash string, creationTime time.Time) (*storage.Record, error) {
	// Explicitly create the record directory structure first
	recordPath := filepath.Join(baseDir, hash)
	err := os.MkdirAll(recordPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create mock record dir %s: %w", recordPath, err)
	}

	// Write the specific CreationTime to metadata
	meta := storage.Metadata{CreationTime: creationTime}
	metaPath := filepath.Join(recordPath, storage.MetadataFileName) // Use created recordPath
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata for %s: %w", hash, err)
	}
	if err = os.WriteFile(metaPath, metaBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write metadata for %s: %w", hash, err)
	}

	// Create dummy data file to satisfy ListRecords check
	dataFilePath := filepath.Join(recordPath, storage.DataFileName)
	dataFile, err := os.Create(dataFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create dummy data file %s: %w", dataFilePath, err)
	}
	dataFile.Close()

	// Return a basic Record struct as expected by the signature
	// Note: This doesn't use rm.CreateRecord() as it wasn't creating the dir
	return &storage.Record{
		Hash:         hash,
		CreationTime: creationTime,
	}, nil
}

// setupTestEngine initializes a test Gin engine and DataHandler using a specific storage path.
func setupTestEngine(t *testing.T, storagePath string) (*gin.Engine, *DataHandler) {
	t.Helper()
	// Minimal config needed for DataHandler initialization
	cfg := &config.Config{
		Setup: config.Setup{
			App: config.App{
				Version: "0.0.1",
			},
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	// Initialize RecordManager for the test
	// Use the provided storagePath
	recordManager, err := storage.NewRecordManager(storagePath)
	require.NoError(t, err)

	// Initialize DataHandler directly
	dataHandler := &DataHandler{records: recordManager, Cfg: cfg}

	// Setup only the routes needed for TestListRecords (HTML handler)
	r.GET("/data", dataHandler.ListRecords)
	// Add other HTML routes if needed by other tests later
	// r.GET("/data/:hash/:type", dataHandler.GetRecordData)
	// r.DELETE("/data/:hash", dataHandler.DeleteRecord)

	// No need for rm.Close() here, temp dir cleanup handles storage
	// t.Cleanup(func() { rm.Close() })

	return r, dataHandler
}

// setupTestAPIServer initializes a test HTTP server and DataHandler for API endpoint tests
func setupTestAPIServer(t *testing.T) (*httptest.Server, *DataHandler, string) {
	// Create a minimal config with a temporary directory for this test run
	cfg := &config.Config{
		Setup: config.Setup{
			App: config.App{
				Version: "0.0.1",
			},
		},
		Server: config.Server{
			Port: 8080, // Example port, not actually used by httptest.Server
		},
	}

	// Initialize RecordManager for the test
	tempDir := t.TempDir()
	recordManager, err := storage.NewRecordManager(tempDir)
	require.NoError(t, err)

	// Initialize DataHandler directly
	dataHandler := &DataHandler{records: recordManager, Cfg: cfg}

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// *** ADD: Create mock records in the handler's BaseDir ***
	storageDir := tempDir // Get the correct temp dir
	// Create a temporary RM pointing to the same dir for setup purposes
	setupRM, err := storage.NewRecordManager(storageDir)
	require.NoError(t, err, "Failed to create setup RecordManager")

	baseTime := time.Now()
	// Create 5 records, newest first for default API sort
	for i := 0; i < 5; i++ {
		creationTime := baseTime.Add(time.Duration(i) * time.Second * -1)
		recordHash := fmt.Sprintf("record%d", 4-i) // record4, record3, ... record0
		// Pass baseDir to createMockRecord
		_, err = createMockRecord(setupRM, storageDir, recordHash, creationTime)
		require.NoError(t, err, "Failed to create mock record %s", recordHash)
	}
	// We don't need setupRM anymore
	// *** END ADD ***

	// --- API Versioning Setup --- (Similar to main.go)
	majorVersion := "0" // Default if split fails or version is invalid
	if parts := strings.Split(cfg.Setup.App.Version, "."); len(parts) > 0 {
		majorVersion = parts[0]
	}
	apiBasePath := fmt.Sprintf("/api/v%s", majorVersion)
	apiGroup := r.Group(apiBasePath)
	{
		// Register the API endpoint correctly within the versioned group
		apiGroup.GET("/data", dataHandler.ListRecordsAPI)
		// Add other API routes needed for testing here
		// apiGroup.GET("/explore/:hash", dataHandler.GetExplorerData)
		// apiGroup.DELETE("/data/:hash", dataHandler.DeleteRecord)
		// apiGroup.GET("/data/:hash/:type", dataHandler.GetRecordData)
	}

	// Debug: Print registered routes
	log.Printf("Registered routes for API test server:")
	for _, route := range r.Routes() {
		log.Printf("  %s %s -> %s", route.Method, route.Path, route.Handler)
	}

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close) // Ensure server is closed after test
	// No need for rm.Close() here either

	return srv, dataHandler, apiBasePath // Return the calculated base path string
}

func TestListRecords_Empty(t *testing.T) {
	// Use setupTestEngine as this tests the HTML rendering handler
	r, _ := setupTestEngine(t, t.TempDir()) // No storage path needed anymore

	req, _ := http.NewRequest("GET", "/data", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Check for the actual message rendered by the template
	assert.Contains(t, w.Body.String(), "No records found.", "Expected empty state message")
}

func TestListRecords_WithData(t *testing.T) {
	// Use setupTestEngine for HTML handler test

	// Create a single temp directory for this test
	tempDir := t.TempDir()

	// Pass the tempDir to the test engine setup
	r, _ := setupTestEngine(t, tempDir)
	// Create a temp RM linked to the handler's storage dir
	// Use the same tempDir for the record creation RM
	rm, err := storage.NewRecordManager(tempDir)
	require.NoError(t, err)

	// Create mock records in the temp directory used by the handler
	now := time.Now()
	baseTime := now.Add(-5 * time.Minute)
	for i := 0; i < 5; i++ {
		creationTime := baseTime.Add(time.Duration(i) * time.Minute)
		recordHash := fmt.Sprintf("hash%d", i)
		// Pass baseDir to createMockRecord
		// Use the same tempDir here as well
		_, err = createMockRecord(rm, tempDir, recordHash, creationTime) // Use rm
		require.NoError(t, err)
	}

	// --- Test Case 2: Sorting Descending (Default) ---
	t.Run("ListWithRecordsDefaultSort", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/data", nil) // Default page 1
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		assert.NotContains(t, body, "No records found.")
		for i := 0; i < 5; i++ {
			recordHash := fmt.Sprintf("hash%d", i)
			assert.Contains(t, body, recordHash)
		}

		// Check order (newest first by default)
		for i := 4; i > 0; i-- {
			posI := strings.Index(body, fmt.Sprintf("hash%d", i))
			posPrev := strings.Index(body, fmt.Sprintf("hash%d", i-1))
			assert.True(t, posI < posPrev, "Expected record %d before record %d", i, i-1)
		}
	})

	// --- Test Case 3: Sorting Ascending ---
	t.Run("ListWithRecordsSortAsc", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/data?sort=time_asc", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		for i := 0; i < 5; i++ {
			recordHash := fmt.Sprintf("hash%d", i)
			assert.Contains(t, body, recordHash)
		}

		// Check order (oldest first)
		for i := 0; i < 4; i++ {
			posI := strings.Index(body, fmt.Sprintf("hash%d", i))
			posNext := strings.Index(body, fmt.Sprintf("hash%d", i+1))
			assert.True(t, posI < posNext, "Expected record %d before record %d", i, i+1)
		}
	})

	// --- Test Case 4: Pagination (assuming default ItemsPerPage=15) ---
	t.Run("Pagination", func(t *testing.T) {
		// Create 16 records total to test pagination (15 on page 1, 1 on page 2)
		// Create the initial 5 records from the setup
		baseTime := now.Add(-5 * time.Minute)
		for i := 0; i < 5; i++ {
			creationTime := baseTime.Add(time.Duration(i) * time.Minute)
			recordHash := fmt.Sprintf("hash%d", i)
			// Pass baseDir to createMockRecord
			_, err = createMockRecord(rm, tempDir, recordHash, creationTime) // Use rm
			require.NoError(t, err)
		}

		// Add 11 more records (total 16)
		for i := 6; i <= 16; i++ {
			// Add a slight offset to ensure distinct creation times
			// Pass baseDir to createMockRecord
			_, err = createMockRecord(rm, tempDir, fmt.Sprintf("page-hash%d", i), now.Add(time.Duration(i)*time.Second)) // Use rm
			require.NoError(t, err)
		}

		// Add a small sleep in case of filesystem delay affecting ModTime reading
		time.Sleep(100 * time.Millisecond)

		// Request Page 1 (Default sort: newest first)
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/data?page=1", nil)
		r.ServeHTTP(w1, req1)

		assert.Equal(t, http.StatusOK, w1.Code)
		body1 := w1.Body.String()
		// Page 1 should contain the 15 newest records: page-hash16 down to hash3-newest
		assert.Contains(t, body1, "page-hash16", "Expected newest page-hash on page 1") // Newest of the loop
		assert.Contains(t, body1, "page-hash6", "Expected oldest page-hash on page 1")  // Oldest of the loop
		assert.Contains(t, body1, "hash4", "Expected hash4 on page 1")                  // 14th newest
		assert.Contains(t, body1, "hash3", "Expected hash3 on page 1")                  // 15th newest
		assert.NotContains(t, body1, "hash0", "Oldest record should not be on page 1")  // 16th newest (should be on page 2)

		// Check pagination rendering for Page 1
		assert.Contains(t, body1, `class="Link--secondary mx-1 px-2 color-fg-accent"`, "Expected page 1 link to be active")
		assert.Contains(t, body1, `>1<`, "Expected page number 1 link text")
		assert.Contains(t, body1, `>2<`, "Expected page number 2 link text")
		assert.NotContains(t, body1, `>3<`, "Expected page number 3 link text (should not exist)")

		// Request Page 2
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/data?page=2", nil)
		r.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
		body2 := w2.Body.String()
		// Page 2 should contain only the 16th record (oldest)
		assert.NotContains(t, body2, "page-hash16", "Newest record should not be on page 2")
		assert.NotContains(t, body2, "hash3", "hash3 should not be on page 2")
		assert.NotContains(t, body2, "hash4", "hash4 should not be on page 2")
		assert.Contains(t, body2, "hash0", "Oldest record (hash0) should be on page 2")

		// Check pagination rendering for Page 2
		assert.Contains(t, body2, `>1<`, "Expected page number 1 link text on page 2")
		assert.Contains(t, body2, `class="Link--secondary mx-1 px-2 color-fg-accent"`, "Expected page 2 link to be active") // Page 2 link active
		assert.Contains(t, body2, `>2<`, "Expected page number 2 link text")
		assert.NotContains(t, body2, `>3<`, "Expected page number 3 link text (should not exist)")
	})

}

// TEST: GIVEN multiple records WHEN listing API records with pagination THEN correct slice is returned
func TestListRecordsAPIPagination(t *testing.T) {
	// Use setupTestAPIServer as this tests the API endpoint
	srv, _, apiBasePath := setupTestAPIServer(t) // Uses setup with 5 records

	// Helper function to make request and decode response
	fetchAndDecode := func(queryParams string) (ListRecordsAPIResponse, *http.Response) {
		url := srv.URL + apiBasePath + "/data" // Construct full API endpoint URL
		if queryParams != "" {
			url += "?" + queryParams
		}
		req, _ := http.NewRequest("GET", url, nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp ListRecordsAPIResponse
		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		require.NoError(t, err)
		resp.Body.Close()
		return apiResp, resp
	}

	// Define expected order (newest first by default from setupTestServer)
	// Hashes are record4, record3, record2, record1, record0
	expectedHashes := []string{"record4", "record3", "record2", "record1", "record0"}

	// --- Test Cases ---

	// Case 1: No parameters (default limit/offset) -> should return all 5, newest first
	t.Run("NoParams", func(t *testing.T) {
		respData, _ := fetchAndDecode("")
		assert.Equal(t, 5, respData.Total)
		require.Len(t, respData.Records, 5)
		assert.Equal(t, expectedHashes[0], respData.Records[0].Hash)
		assert.Equal(t, expectedHashes[4], respData.Records[4].Hash)
		// Quick check of full order
		returnedHashes := make([]string, len(respData.Records))
		for i, r := range respData.Records {
			returnedHashes[i] = r.Hash
		}
		assert.Equal(t, expectedHashes, returnedHashes)
	})

	// Case 2: limit=2 -> should return 2 newest
	t.Run("Limit2", func(t *testing.T) {
		respData, _ := fetchAndDecode("limit=2")
		assert.Equal(t, 5, respData.Total)
		require.Len(t, respData.Records, 2)
		assert.Equal(t, expectedHashes[0], respData.Records[0].Hash) // record4
		assert.Equal(t, expectedHashes[1], respData.Records[1].Hash) // record3
	})

	// Case 3: limit=3&offset=1 -> should return 2nd, 3rd, 4th newest
	t.Run("Limit3Offset1", func(t *testing.T) {
		respData, _ := fetchAndDecode("limit=3&offset=1")
		assert.Equal(t, 5, respData.Total)
		require.Len(t, respData.Records, 3)
		assert.Equal(t, expectedHashes[1], respData.Records[0].Hash) // record3
		assert.Equal(t, expectedHashes[2], respData.Records[1].Hash) // record2
		assert.Equal(t, expectedHashes[3], respData.Records[2].Hash) // record1
	})

	// Case 4: limit=2&offset=4 -> should return the oldest one
	t.Run("Limit2Offset4", func(t *testing.T) {
		respData, _ := fetchAndDecode("limit=2&offset=4")
		assert.Equal(t, 5, respData.Total)
		require.Len(t, respData.Records, 1)
		assert.Equal(t, expectedHashes[4], respData.Records[0].Hash) // record0
	})

	// Case 5: offset=5 -> should return empty list
	t.Run("OffsetPastEnd", func(t *testing.T) {
		respData, _ := fetchAndDecode("offset=5")
		assert.Equal(t, 5, respData.Total)
		require.Len(t, respData.Records, 0)
	})

	// Case 6: limit=0 -> should behave like no limit (return all)
	t.Run("LimitZero", func(t *testing.T) {
		respData, _ := fetchAndDecode("limit=0")
		assert.Equal(t, 5, respData.Total)
		require.Len(t, respData.Records, 5)
	})

	// Case 7: Invalid limit/offset -> should use defaults (return all)
	t.Run("InvalidParams", func(t *testing.T) {
		respData, _ := fetchAndDecode("limit=-1&offset=abc")
		assert.Equal(t, 5, respData.Total)
		require.Len(t, respData.Records, 5) // Defaults applied
	})
}

// Helper struct to decode the JSON response from ListRecordsAPI
type ListRecordsAPIResponse struct {
	Total   int               `json:"total"` // Add Total field
	Records []*storage.Record `json:"records"`
}
