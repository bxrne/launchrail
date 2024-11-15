package ork

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/pkg/errors"
)

type orkFile struct {
	file       *os.File
	zipReader  *zip.Reader
	rocketFile *zip.File
}

func openOrkFile(path string) (*orkFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "opening file")
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, errors.Wrap(err, "getting file info")
	}

	reader, err := zip.NewReader(file, info.Size())
	if err != nil {
		file.Close()
		return nil, errors.Wrap(err, "creating zip reader")
	}

	var rocketFile *zip.File
	for _, f := range reader.File {
		if f.Name == "rocket.ork" {
			rocketFile = f
			break
		}
	}

	if rocketFile == nil {
		file.Close()
		return nil, errors.New("rocket.ork not found in archive")
	}

	return &orkFile{
		file:       file,
		zipReader:  reader,
		rocketFile: rocketFile,
	}, nil
}

func (o *orkFile) Close() error {
	return o.file.Close()
}

func (o *orkFile) readRocketData() ([]byte, error) {
	rc, err := o.rocketFile.Open()
	if err != nil {
		return nil, errors.Wrap(err, "opening rocket.ork")
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

// DecompressTo extracts rocket.ork data to the specified output file
func DecompressTo(logger *log.Logger, filePath string, outputFilePath string) error {
	ork, err := openOrkFile(filePath)
	if err != nil {
		logger.Error("failed to open ork file", "error", err)
		return err
	}
	defer ork.Close()

	xmlData, err := ork.readRocketData()
	if err != nil {
		logger.Error("failed to read rocket data", "error", err)
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputFilePath), 0755); err != nil {
		logger.Error("failed to create output directory", "error", err)
		return errors.Wrap(err, "creating output directory")
	}

	if err := os.WriteFile(outputFilePath, xmlData, 0644); err != nil {
		logger.Error("failed to write output file", "error", err)
		return errors.Wrap(err, "writing output file")
	}

	logger.Info("successfully decompressed file",
		"input", filePath,
		"output", outputFilePath)
	return nil
}

// Decompress extracts and parses the rocket.ork file into an Openrocket struct
func Decompress(filePath string) (*Openrocket, error) {
	ork, err := openOrkFile(filePath)
	if err != nil {
		return nil, err
	}
	defer ork.Close()

	xmlData, err := ork.readRocketData()
	if err != nil {
		return nil, err
	}

	var rocket Openrocket
	if err := xml.Unmarshal(xmlData, &rocket); err != nil {
		return nil, errors.Wrap(err, "unmarshaling XML")
	}

	return &rocket, nil
}
