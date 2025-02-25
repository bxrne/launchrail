package storage_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bxrne/launchrail/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (string, string, func()) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	baseDir := "test_base"
	dir := "test_dir"
	fullBaseDir := filepath.Join(homeDir, baseDir)

	cleanup := func() {
		os.RemoveAll(fullBaseDir)
	}

	return baseDir, dir, cleanup
}

// TEST: GIVEN a base directory and a directory name WHEN NewStorage for MOTION data is called THEN a new storage instance is created
func TestNewStorageMotion(t *testing.T) {
	baseDir, dir, cleanup := setupTest(t)
	defer cleanup()

	_, err := storage.NewStorage(baseDir, dir, storage.MOTION)
	require.NoError(t, err)

	homeDir, _ := os.UserHomeDir()
	expectedBaseDir := filepath.Join(homeDir, baseDir)
	expectedDir := filepath.Join(expectedBaseDir, dir)

	_, err = os.Stat(expectedBaseDir)
	assert.NoError(t, err)
	_, err = os.Stat(expectedDir)
	assert.NoError(t, err)
}

// TEST: GIVEN a base directory and a directory name WHEN NewStorage for EVENTS data is called THEN a new storage instance is created
func TestNewStorageEvents(t *testing.T) {
	baseDir, dir, cleanup := setupTest(t)
	defer cleanup()

	_, err := storage.NewStorage(baseDir, dir, storage.EVENTS)
	require.NoError(t, err)

	homeDir, _ := os.UserHomeDir()
	expectedBaseDir := filepath.Join(homeDir, baseDir)
	expectedDir := filepath.Join(expectedBaseDir, dir)

	_, err = os.Stat(expectedBaseDir)
	assert.NoError(t, err)
	_, err = os.Stat(expectedDir)
	assert.NoError(t, err)
}

// TEST: GIVEN a base directory and a directory name WHEN NewStorage is called THEN a new storage instance is created
func TestInit(t *testing.T) {
	baseDir, dir, cleanup := setupTest(t)
	defer cleanup()

	s, err := storage.NewStorage(baseDir, dir, storage.MOTION)
	require.NoError(t, err)

	headers := []string{"Column1", "Column2", "Column3"}
	err = s.Init(headers)
	require.NoError(t, err)

	// Ensure the file is closed after writing
	err = s.Close()
	require.NoError(t, err)

	// Add a small delay to ensure the file system has enough time to flush the data
	time.Sleep(100 * time.Millisecond)

	homeDir, _ := os.UserHomeDir()
	fullDir := filepath.Join(homeDir, baseDir, dir)

	files, err := os.ReadDir(fullDir)
	require.NoError(t, err)

	filePath := filepath.Join(fullDir, files[0].Name())
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	readHeaders, err := reader.Read()
	require.NoError(t, err)
	assert.Equal(t, headers, readHeaders)
}

// TEST: GIVEN a base directory and a directory name WHEN NewStorage is called THEN a new storage instance is created
func TestWrite(t *testing.T) {
	baseDir, dir, cleanup := setupTest(t)
	defer cleanup()

	s, err := storage.NewStorage(baseDir, dir, storage.MOTION)
	require.NoError(t, err)

	headers := []string{"Column1", "Column2", "Column3"}
	err = s.Init(headers)
	require.NoError(t, err)

	data := []string{"Value1", "Value2", "Value3"}
	err = s.Write(data)
	require.NoError(t, err)

	// Ensure the file is closed after writing
	err = s.Close()
	require.NoError(t, err)

	// Add a small delay to ensure the file system has enough time to flush the data
	time.Sleep(100 * time.Millisecond)

	homeDir, _ := os.UserHomeDir()
	fullDir := filepath.Join(homeDir, baseDir, dir)

	files, err := os.ReadDir(fullDir)
	require.NoError(t, err)

	filePath := filepath.Join(fullDir, files[0].Name())
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read()
	require.NoError(t, err)

	readData, err := reader.Read()
	require.NoError(t, err)
	assert.Equal(t, data, readData)
}

// TEST: GIVEN a base directory and a directory name WHEN NewStorage is called THEN a new storage instance is created
func TestWriteInvalidData(t *testing.T) {
	baseDir, dir, cleanup := setupTest(t)
	defer cleanup()

	s, err := storage.NewStorage(baseDir, dir, storage.MOTION)
	require.NoError(t, err)

	headers := []string{"Column1", "Column2"}
	err = s.Init(headers)
	require.NoError(t, err)

	data := []string{"Value1", "Value2", "Value3"}
	err = s.Write(data)
	require.Error(t, err)
	assert.EqualError(t, err, "data length (3) does not match headers length (2)")
}
