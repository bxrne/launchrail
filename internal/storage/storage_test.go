package storage_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTest uses t.TempDir() to create an isolated temporary directory.
func setupTest(t *testing.T) (string, string, func()) {
	baseDir := t.TempDir() // Use temporary directory provided by the test framework.
	dir := "test_dir"
	cleanup := func() {
		// Increase delay to ensure the filesystem has time to flush data.
		time.Sleep(500 * time.Millisecond)
		// t.TempDir() is automatically cleaned up.
	}
	return baseDir, dir, cleanup
}

// TEST: GIVEN a base dir WHEN creating MOTION storage THEN no error is returned
func TestNewStorageMotion(t *testing.T) {
	baseDir, _, cleanup := setupTest(t)
	defer cleanup()
	cfg := &config.Config{Setup: config.Setup{Logging: config.Logging{Level: "error"}}}

	recordDir := filepath.Join(baseDir, "test_record")
	s, err := storage.NewStorage(recordDir, storage.MOTION, cfg)
	require.NoError(t, err)

	// Close the storage before cleanup
	require.NoError(t, s.Close())

	expectedBaseDir := baseDir // baseDir is already absolute.
	expectedDir := filepath.Join(expectedBaseDir, "test_record")

	_, err = os.Stat(expectedBaseDir)
	assert.NoError(t, err)
	_, err = os.Stat(expectedDir)
	assert.NoError(t, err)
}

// TEST: GIVEN a base dir WHEN creating EVENTS storage THEN no error is returned
func TestNewStorageEvents(t *testing.T) {
	baseDir, _, cleanup := setupTest(t)
	defer cleanup()
	cfg := &config.Config{Setup: config.Setup{Logging: config.Logging{Level: "error"}}}

	recordDir := filepath.Join(baseDir, "test_record")
	s, err := storage.NewStorage(recordDir, storage.EVENTS, cfg)
	require.NoError(t, err)

	// Close the storage before cleanup
	require.NoError(t, s.Close())

	expectedBaseDir := baseDir
	expectedDir := filepath.Join(expectedBaseDir, "test_record")

	_, err = os.Stat(expectedBaseDir)
	assert.NoError(t, err)
	_, err = os.Stat(expectedDir)
	assert.NoError(t, err)
}

// TEST: GIVEN a storage WHEN calling Init THEN the CSV file is created with headers
func TestInit(t *testing.T) {
	baseDir, _, cleanup := setupTest(t)
	defer cleanup()
	cfg := &config.Config{Setup: config.Setup{Logging: config.Logging{Level: "error"}}}

	recordDir := filepath.Join(baseDir, "test_init")
	s, err := storage.NewStorage(recordDir, storage.MOTION, cfg)
	require.NoError(t, err)

	err = s.Init()
	require.NoError(t, err)

	// Close to flush and release the file
	require.NoError(t, s.Close())

	// Remove sleep as it's no longer needed since we properly close the file
	fullDir := recordDir
	files, err := os.ReadDir(fullDir)
	require.NoError(t, err)
	require.NotEmpty(t, files, "expected at least one file in %s", fullDir)

	filePath := filepath.Join(fullDir, files[0].Name())
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	readHeaders, err := reader.Read()
	require.NoError(t, err)
	assert.Equal(t, storage.StorageHeaders[storage.MOTION], readHeaders)
}

// TEST: GIVEN a storage WHEN writing valid data THEN data is appended
func TestWrite(t *testing.T) {
	baseDir, _, cleanup := setupTest(t)
	defer cleanup()
	cfg := &config.Config{Setup: config.Setup{Logging: config.Logging{Level: "error"}}}

	recordDir := filepath.Join(baseDir, "test_write")
	s, err := storage.NewStorage(recordDir, storage.MOTION, cfg)
	require.NoError(t, err)

	err = s.Init()
	require.NoError(t, err)

	data := []string{"Value1", "Value2", "Value3", "Value4", "Value5"}
	err = s.Write(data)
	require.NoError(t, err)

	require.NoError(t, s.Close())

	// Remove sleep as it's no longer needed
	fullDir := recordDir
	files, err := os.ReadDir(fullDir)
	require.NoError(t, err)
	require.NotEmpty(t, files, "expected at least one file in %s", fullDir)

	filePath := filepath.Join(fullDir, files[0].Name())
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read() // read headers
	require.NoError(t, err)

	readData, err := reader.Read()
	require.NoError(t, err)
	assert.Equal(t, data, readData)
}

// TEST: GIVEN a storage WHEN writing invalid data THEN an error is returned
func TestWriteInvalidData(t *testing.T) {
	baseDir, _, cleanup := setupTest(t)
	defer cleanup()
	cfg := &config.Config{Setup: config.Setup{Logging: config.Logging{Level: "error"}}}

	recordDir := filepath.Join(baseDir, "test_invalid_data")
	s, err := storage.NewStorage(recordDir, storage.MOTION, cfg)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	err = s.Init()
	require.NoError(t, err)

	data := []string{"Value1", "Value2", "Value3"}
	err = s.Write(data)
	require.Error(t, err)
	assert.EqualError(t, err, "data length (3) does not match headers length (5)")
}

// TEST: GIVEN a storage with data WHEN calling ReadAll THEN data is returned
func TestReadAll(t *testing.T) {
	baseDir, _, cleanup := setupTest(t)
	defer cleanup()
	cfg := &config.Config{Setup: config.Setup{Logging: config.Logging{Level: "error"}}}

	recordDir := filepath.Join(baseDir, "test_read_all")
	s, err := storage.NewStorage(recordDir, storage.MOTION, cfg)
	require.NoError(t, err)
	require.NoError(t, s.Init())

	data := []string{"1", "2", "3", "4", "5"}
	require.NoError(t, s.Write(data))
	require.NoError(t, s.Close())

	s2, err := storage.NewStorage(recordDir, storage.MOTION, cfg)
	require.NoError(t, err)
	defer s2.Close()

	rows, err := s2.ReadAll()
	require.NoError(t, err)
	require.Len(t, rows, 2) // headers + data row
}

// TEST: GIVEN a storage with data WHEN calling ReadHeadersAndData THEN headers and rows are returned
func TestReadHeadersAndData(t *testing.T) {
	baseDir, _, cleanup := setupTest(t)
	defer cleanup()
	cfg := &config.Config{Setup: config.Setup{Logging: config.Logging{Level: "error"}}}

	recordDir := filepath.Join(baseDir, "test_read_headers_and_data")
	s, err := storage.NewStorage(recordDir, storage.EVENTS, cfg)
	require.NoError(t, err)
	require.NoError(t, s.Init())

	data := []string{"0.1", "Liftoff", "ACTIVE", "NONE"} // Provide 4 values
	require.NoError(t, s.Write(data))
	require.NoError(t, s.Close())

	s2, err := storage.NewStorage(recordDir, storage.EVENTS, cfg)
	require.NoError(t, err)
	defer s2.Close()

	headers, rows, err := s2.ReadHeadersAndData()
	require.NoError(t, err)
	require.Len(t, headers, len(storage.StorageHeaders[storage.EVENTS]))
	require.Len(t, rows, 1)
}

// TEST: GIVEN a storage WHEN calling GetFilePath THEN the correct path is returned
func TestGetFilePath(t *testing.T) {
	baseDir, _, cleanup := setupTest(t)
	defer cleanup()
	cfg := &config.Config{Setup: config.Setup{Logging: config.Logging{Level: "error"}}}

	recordDir := filepath.Join(baseDir, "test_get_file_path")
	s, err := storage.NewStorage(recordDir, storage.DYNAMICS, cfg)
	require.NoError(t, err)
	assert.Contains(t, s.GetFilePath(), "DYNAMICS.csv")
	require.NoError(t, s.Close())
}
