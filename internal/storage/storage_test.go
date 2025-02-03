package storage_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

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

// TEST: GIVEN a base directory and a directory name WHEN NewStorage is called THEN a new storage instance is created
func TestNewStorage(t *testing.T) {
	baseDir, dir, cleanup := setupTest(t)
	defer cleanup()

	_, err := storage.NewStorage(baseDir, dir)
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

	s, err := storage.NewStorage(baseDir, dir)
	require.NoError(t, err)

	headers := []string{"Column1", "Column2", "Column3"}
	err = s.Init(headers)
	require.NoError(t, err)

	homeDir, _ := os.UserHomeDir()
	fullDir := filepath.Join(homeDir, baseDir, dir)

	files, err := os.ReadDir(fullDir)
	require.NoError(t, err)
	require.Len(t, files, 1)

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

	s, err := storage.NewStorage(baseDir, dir)
	require.NoError(t, err)

	headers := []string{"Column1", "Column2", "Column3"}
	err = s.Init(headers)
	require.NoError(t, err)

	data := []string{"Value1", "Value2", "Value3"}
	err = s.Write(data)
	require.NoError(t, err)

	homeDir, _ := os.UserHomeDir()
	fullDir := filepath.Join(homeDir, baseDir, dir)

	files, err := os.ReadDir(fullDir)
	require.NoError(t, err)
	require.Len(t, files, 1)

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

	s, err := storage.NewStorage(baseDir, dir)
	require.NoError(t, err)

	headers := []string{"Column1", "Column2"}
	err = s.Init(headers)
	require.NoError(t, err)

	data := []string{"Value1", "Value2", "Value3"}
	err = s.Write(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data length does not match headers length")
}
