package ork_test

import (
	"archive/zip"
	"bytes"
	"os"
	"testing"

	"github.com/bxrne/launchrail/pkg/ork"
	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a valid .ork file WHEN DecompressTo is called THEN it should extract the XML data correctly
func TestDecompressTo_ValidFile(t *testing.T) {
	inputFilePath := "../../testdata/ULR3.ork"
	outputFile, err := os.CreateTemp("", "output_*.xml")
	assert.NoError(t, err)
	defer os.Remove(outputFile.Name())

	logger := log.NewWithOptions(nil, log.Options{})
	err = ork.DecompressTo(logger, inputFilePath, outputFile.Name())
	assert.NoError(t, err)

	_, err = os.ReadFile(outputFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, outputFile)
}

// TEST: GIVEN a non-existent .ork file WHEN DecompressTo is called THEN it should return an error
func TestDecompressTo_FileNotFound(t *testing.T) {
	logger := log.NewWithOptions(nil, log.Options{})
	err := ork.DecompressTo(logger, "non_existent_file.ork", "output.xml")
	assert.Error(t, err)
}

// TEST: GIVEN a .ork file without rocket.ork WHEN DecompressTo is called THEN it should return an error
func TestDecompressTo_MissingRocketFile(t *testing.T) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	_, err := zipWriter.Create("otherfile.txt")
	assert.NoError(t, err)
	err = zipWriter.Close()
	assert.NoError(t, err)

	tmpFile, err := os.CreateTemp("", "test_*.ork")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(buf.Bytes())
	assert.NoError(t, err)
	tmpFile.Close()

	logger := log.NewWithOptions(nil, log.Options{})
	err = ork.DecompressTo(logger, tmpFile.Name(), "output.xml")
	assert.Error(t, err)
}

// TEST: GIVEN a valid .ork file WHEN Decompress is called THEN it should return the correct Openrocket struct
func TestDecompress_ValidFile(t *testing.T) {
	inputFilePath := "../../testdata/ULR3.ork"

	orkData, err := ork.Decompress(inputFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, orkData)
	assert.Equal(t, "ULR3", orkData.Rocket.Name)
}

// TEST: GIVEN a non-existent .ork file WHEN Decompress is called THEN it should return an error
func TestDecompress_FileNotFound(t *testing.T) {
	_, err := ork.Decompress("non_existent_file.ork")
	assert.Error(t, err)
}

// TEST: GIVEN a .ork file without rocket.ork WHEN Decompress is called THEN it should return an error
func TestDecompress_MissingRocketFile(t *testing.T) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	_, err := zipWriter.Create("otherfile.txt")
	assert.NoError(t, err)
	err = zipWriter.Close()
	assert.NoError(t, err)

	tmpFile, err := os.CreateTemp("", "test_*.ork")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(buf.Bytes())
	assert.NoError(t, err)
	tmpFile.Close()

	_, err = ork.Decompress(tmpFile.Name())
	assert.Error(t, err)
}
