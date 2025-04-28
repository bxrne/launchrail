package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary directory for testing storage
func setupTestStorage(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "launchrail-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	// RecordManager operates directly in the baseDir, no need for subdirs here.
	return tempDir
}

// createMockRecord creates a mock record directory with minimal structure and a specific creation time.
func createMockRecord(t *testing.T, baseDir, hash string, creationTime time.Time) {
	t.Helper()
	recordPath := filepath.Join(baseDir, hash)
	err := os.MkdirAll(recordPath, 0755)
	require.NoError(t, err, "Failed to create mock record dir %s", recordPath)

	// Create the metadata file with the specified creation time
	meta := storage.Metadata{CreationTime: creationTime}
	metaFilePath := filepath.Join(recordPath, storage.MetadataFileName)
	metaFile, err := os.Create(metaFilePath)
	require.NoError(t, err, "Failed to create metadata file %s", metaFilePath)
	defer metaFile.Close()
	err = json.NewEncoder(metaFile).Encode(meta)
	require.NoError(t, err, "Failed to encode metadata to %s", metaFilePath)

	// Create a dummy data file to make the record appear valid in ListRecords
	dataFilePath := filepath.Join(recordPath, storage.DataFileName)
	dataFile, err := os.Create(dataFilePath)
	require.NoError(t, err, "Failed to create dummy data file %s", dataFilePath)
	dataFile.Close()

	// Setting directory ModTime is no longer the primary mechanism, but might still be needed
	// if metadata reading fails. Keep it for robustness, but rely on CreationTime.
	// Note: Chtimes might still have cross-platform issues, but it's a fallback now.
	if runtime.GOOS != "windows" { // Chtimes not fully supported on Windows
		err = os.Chtimes(recordPath, creationTime, creationTime)
		// Log warning instead of failing the test if Chtimes fails
		if err != nil {
			t.Logf("Warning: could not set times for %s: %v", recordPath, err)
		}
	}
}

// Helper function to set up the Gin engine and handler for testing
func setupTestServer(t *testing.T, storagePath string) (*gin.Engine, *DataHandler) {
	// Minimal config for testing
	testCfg := &config.Config{
		Setup: config.Setup{
			App: config.App{
				BaseDir: storagePath,
				Version: "test-v0.1",
			},
		},
	}

	h, err := NewDataHandler(testCfg)
	require.NoError(t, err, "Failed to create DataHandler")

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Register routes needed for testing
	r.GET("/data", h.ListRecords)
	// Add other routes as needed for tests

	return r, h
}

func TestListRecords(t *testing.T) {
	storageDir := setupTestStorage(t)
	defer os.RemoveAll(storageDir) // Clean up temp dir

	r, _ := setupTestServer(t, storageDir)

	// --- Test Case 1: Empty list --- 
	t.Run("EmptyList", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/data", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Check for the actual message rendered by the template
		assert.Contains(t, w.Body.String(), "No records found.", "Expected empty state message")
	})

	// --- Create Mock Records --- 
	now := time.Now()
	record1Hash := "hash1-oldest"
	record2Hash := "hash2-middle"
	record3Hash := "hash3-newest"

	createMockRecord(t, storageDir, record1Hash, now.Add(-3*time.Minute)) // Oldest
	createMockRecord(t, storageDir, record2Hash, now.Add(-2*time.Minute)) // Middle
	createMockRecord(t, storageDir, record3Hash, now.Add(-1*time.Minute)) // Newest

	// --- Test Case 2: Sorting Descending (Default) ---
	t.Run("ListWithRecordsDefaultSort", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/data", nil) // Default page 1
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		assert.NotContains(t, body, "No records found.")
		assert.Contains(t, body, record1Hash)
		assert.Contains(t, body, record2Hash)
		assert.Contains(t, body, record3Hash)

		// Check order (newest first by default)
		pos3 := strings.Index(body, record3Hash)
		pos2 := strings.Index(body, record2Hash)
		pos1 := strings.Index(body, record1Hash)

		assert.True(t, pos3 < pos2, "Expected newest (%s) before middle (%s)", record3Hash, record2Hash)
		assert.True(t, pos2 < pos1, "Expected middle (%s) before oldest (%s)", record2Hash, record1Hash)
	})

	// --- Test Case 3: Sorting Ascending --- 
	t.Run("ListWithRecordsSortAsc", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/data?sort=time_asc", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, record1Hash)
		assert.Contains(t, body, record2Hash)
		assert.Contains(t, body, record3Hash)

		// Check order (oldest first)
		pos1 := strings.Index(body, record1Hash)
		pos2 := strings.Index(body, record2Hash)
		pos3 := strings.Index(body, record3Hash)

		assert.True(t, pos1 < pos2, "Expected oldest (%s) before middle (%s)", record1Hash, record2Hash)
		assert.True(t, pos2 < pos3, "Expected middle (%s) before newest (%s)", record2Hash, record3Hash)
	})

	// --- Test Case 4: Pagination (assuming default ItemsPerPage=15) ---
	t.Run("Pagination", func(t *testing.T) {
		// Create 16 records total to test pagination (15 on page 1, 1 on page 2)
		storageDir := setupTestStorage(t) // Use the test setup's storage dir
		defer os.RemoveAll(storageDir)

		now := time.Now() // Use a consistent 'now'
		// Create the initial 3 records from the setup
		record1Hash := "hash1-oldest"
		record2Hash := "hash2-middle"
		record3Hash := "hash3-newest"
		createMockRecord(t, storageDir, record1Hash, now.Add(-3*time.Minute))
		createMockRecord(t, storageDir, record2Hash, now.Add(-2*time.Minute))
		createMockRecord(t, storageDir, record3Hash, now.Add(-1*time.Minute))

		// Add 13 more records (total 16)
		for i := 4; i <= 16; i++ {
			// Add a slight offset to ensure distinct creation times
			createMockRecord(t, storageDir, fmt.Sprintf("page-hash%d", i), now.Add(time.Duration(i)*time.Second))
		}

		// Add a small sleep in case of filesystem delay affecting ModTime reading
		time.Sleep(100 * time.Millisecond)

		// Setup server using the same storageDir
		r, _ := setupTestServer(t, storageDir) // Correct assignment: setupTestServer returns 2 values

		// Request Page 1 (Default sort: newest first)
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/data?page=1", nil)
		r.ServeHTTP(w1, req1)

		assert.Equal(t, http.StatusOK, w1.Code)
		body1 := w1.Body.String()
		// Page 1 should contain the 15 newest records: page-hash16 down to hash3-newest
		assert.Contains(t, body1, "page-hash16", "Expected newest page-hash on page 1") // Newest of the loop
		assert.Contains(t, body1, "page-hash4", "Expected oldest page-hash on page 1")   // Oldest of the loop
		assert.Contains(t, body1, record3Hash, "Expected hash3-newest on page 1")       // 14th newest
		assert.Contains(t, body1, record2Hash, "Expected hash2-middle on page 1")       // 15th newest
		assert.NotContains(t, body1, record1Hash, "Oldest record should not be on page 1") // 16th newest (should be on page 2)

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
		assert.NotContains(t, body2, record2Hash, "hash2-middle should not be on page 2")
		assert.NotContains(t, body2, record3Hash, "hash3-newest should not be on page 2")
		assert.Contains(t, body2, record1Hash, "Oldest record (hash1-oldest) should be on page 2")

		// Check pagination rendering for Page 2
		assert.Contains(t, body2, `>1<`, "Expected page number 1 link text on page 2")
		assert.Contains(t, body2, `class="Link--secondary mx-1 px-2 color-fg-accent"`, "Expected page 2 link to be active") // Page 2 link active
		assert.Contains(t, body2, `>2<`, "Expected page number 2 link text")
		assert.NotContains(t, body2, `>3<`, "Expected page number 3 link text (should not exist)")
	})

}

// TODO: Add tests for other handlers like DeleteRecord, GetRecordData, ListRecordsAPI, etc.
