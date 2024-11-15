package ork_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/bxrne/launchrail/pkg/ork"
	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestZip(t *testing.T, content []byte) string {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "test.ork")

	// Create a buffer to write our zip to
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Add rocket.ork file to the zip
	f, err := w.Create("rocket.ork")
	require.NoError(t, err)

	if content != nil {
		_, err = f.Write(content)
		require.NoError(t, err)
	}

	err = w.Close()
	require.NoError(t, err)

	// Write the zip file
	err = os.WriteFile(zipPath, buf.Bytes(), 0644)
	require.NoError(t, err)

	return zipPath
}

// TEST: GIVEN a valid ork file WHEN DecompressTo is called THEN it should extract the rocket data correctly
func TestDecompressTo_ValidFile(t *testing.T) {
	logger := log.New(os.Stderr)
	input := createTestZip(t, []byte(`<openrocket>test</openrocket>`))
	output := filepath.Join(t.TempDir(), "output.xml")

	err := ork.DecompressTo(logger, input, output)
	assert.NoError(t, err)

	content, err := os.ReadFile(output)
	require.NoError(t, err)
	assert.Equal(t, "<openrocket>test</openrocket>", string(content))
}

// TEST: GIVEN a nonexistent file WHEN DecompressTo is called THEN it should return an error
func TestDecompressTo_NonexistentFile(t *testing.T) {
	logger := log.New(os.Stderr)
	err := ork.DecompressTo(logger, "nonexistent.ork", "output.xml")
	assert.Error(t, err)
}

// TEST: GIVEN an invalid zip file WHEN DecompressTo is called THEN it should return an error
func TestDecompressTo_InvalidZip(t *testing.T) {
	logger := log.New(os.Stderr)
	invalidFile := filepath.Join(t.TempDir(), "invalid.ork")
	require.NoError(t, os.WriteFile(invalidFile, []byte("not a zip file"), 0644))

	err := ork.DecompressTo(logger, invalidFile, "output.xml")
	assert.Error(t, err)
}

// TEST: GIVEN a zip file without rocket.ork WHEN DecompressTo is called THEN it should return an error
func TestDecompressTo_MissingRocketOrk(t *testing.T) {
	logger := log.New(os.Stderr)
	input := createTestZip(t, nil)

	err := ork.DecompressTo(logger, input, "output.xml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rocket.ork not found")
}

// TEST: GIVEN a valid ork file WHEN Decompress is called THEN it should parse the XML data correctly
func TestDecompress_ValidFile(t *testing.T) {
	input := createTestZip(t, []byte(`<openrocket version="1.0"><rocket><name>Test Rocket</name></rocket></openrocket>`))

	rocket, err := ork.Decompress(input)
	assert.NoError(t, err)
	assert.NotNil(t, rocket)
	assert.Equal(t, "1.0", rocket.Version)
	assert.Equal(t, "Test Rocket", rocket.Rocket.Name)
}

// TEST: GIVEN an invalid XML content WHEN Decompress is called THEN it should return an error
func TestDecompress_InvalidXML(t *testing.T) {
	input := createTestZip(t, []byte(`invalid xml content`))

	rocket, err := ork.Decompress(input)
	assert.Error(t, err)
	assert.Nil(t, rocket)
}

// TEST: GIVEN a nonexistent file WHEN Decompress is called THEN it should return an error
func TestDecompress_NonexistentFile(t *testing.T) {
	rocket, err := ork.Decompress("nonexistent.ork")
	assert.Error(t, err)
	assert.Nil(t, rocket)
}
