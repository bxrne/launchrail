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

func TestNewStorage(t *testing.T) {
	baseDir := "./test_base"
	dir := "./test_dir"
	defer os.RemoveAll(baseDir)
	defer os.RemoveAll(dir)

	_, err := storage.NewStorage(baseDir, dir)

	require.NoError(t, err)

	// Check directories were created
	_, err1 := os.Stat(baseDir)
	_, err2 := os.Stat(dir)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func TestNewStorageExistingDirectories(t *testing.T) {
	baseDir := "./test_base_existing"
	dir := "./test_dir_existing"

	// Create directories first
	err1 := os.MkdirAll(baseDir, os.ModePerm)
	err2 := os.MkdirAll(dir, os.ModePerm)
	require.NoError(t, err1)
	require.NoError(t, err2)

	defer os.RemoveAll(baseDir)
	defer os.RemoveAll(dir)

	_, err := storage.NewStorage(baseDir, dir)

	require.NoError(t, err)
}

func TestInit(t *testing.T) {
	baseDir := "./test_base_init"
	dir := "./test_dir_init"
	defer os.RemoveAll(baseDir)
	defer os.RemoveAll(dir)

	s, err := storage.NewStorage(baseDir, dir)
	require.NoError(t, err)

	headers := []string{"Column1", "Column2", "Column3"}
	err = s.Init(headers)
	require.NoError(t, err)

	// Find the CSV file
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Open and check the file
	filePath := filepath.Join(dir, files[0].Name())
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	readHeaders, err := reader.Read()
	require.NoError(t, err)
	assert.Equal(t, headers, readHeaders)
}

func TestWrite(t *testing.T) {
	baseDir := "./test_base_write"
	dir := "./test_dir_write"
	defer os.RemoveAll(baseDir)
	defer os.RemoveAll(dir)

	s, err := storage.NewStorage(baseDir, dir)
	require.NoError(t, err)

	headers := []string{"Column1", "Column2", "Column3"}
	err = s.Init(headers)
	require.NoError(t, err)

	// Write valid data
	data := []string{"Value1", "Value2", "Value3"}
	err = s.Write(data)
	require.NoError(t, err)

	// Find the CSV file
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Open and check the file
	filePath := filepath.Join(dir, files[0].Name())
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip headers
	_, err = reader.Read()
	require.NoError(t, err)

	// Read data row
	readData, err := reader.Read()
	require.NoError(t, err)
	assert.Equal(t, data, readData)
}

func TestWriteInvalidData(t *testing.T) {
	baseDir := "./test_base_invalid"
	dir := "./test_dir_invalid"
	defer os.RemoveAll(baseDir)
	defer os.RemoveAll(dir)

	s, err := storage.NewStorage(baseDir, dir)
	require.NoError(t, err)

	headers := []string{"Column1", "Column2"}
	err = s.Init(headers)
	require.NoError(t, err)

	// Write invalid data (wrong length)
	data := []string{"Value1", "Value2", "Value3"}
	err = s.Write(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data length does not match headers length")
}
