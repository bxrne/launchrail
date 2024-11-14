package ork

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"os"

	"github.com/charmbracelet/log"
	"github.com/pkg/errors"
)

// WARNING: The .ork is actually a zip file with the extension changed to .ork
func DecompressTo(log *log.Logger, filePath string, outputFilePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("could not open file: %s", err)
		return err
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Errorf("error closing file: %s", cerr)
		}
	}()

	info, err := file.Stat()
	if err != nil {
		log.Errorf("could not get file info: %s", err)
		return err
	}

	reader, err := zip.NewReader(file, info.Size())
	if err != nil {
		log.Errorf("could not open as zip file: %s", err)
		return err
	}

	for _, f := range reader.File {
		if f.Name == "rocket.ork" { // WARNING: File now exists in the zip archive (weird)
			rc, err := f.Open()
			if err != nil {
				log.Errorf("could not open 'rocket.ork': %s", err)
				return err
			}
			defer func() {
				if cerr := rc.Close(); cerr != nil {
					log.Errorf("error closing zip file entry: %s", cerr)
				}
			}()

			xmlData, err := io.ReadAll(rc)
			if err != nil {
				log.Errorf("could not read XML data: %s", err)
				return err
			}

			err = os.WriteFile(outputFilePath, xmlData, 0644)
			if err != nil {
				log.Errorf("could not write XML data: %s", err)
				return err
			}

			log.Infof("Decompressed %s to %s", filePath, outputFilePath)

			return nil
		}
	}

	log.Errorf("'rocket.ork' not found in the zip archive")
	return errors.Errorf("'rocket.ork' not found in the zip archive")
}

func Decompress(log *log.Logger, filePath string) (*Openrocket, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("could not open file: %s", err)
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Errorf("error closing file: %s", cerr)
		}
	}()

	info, err := file.Stat()
	if err != nil {
		log.Errorf("could not get file info: %s", err)
		return nil, err
	}

	reader, err := zip.NewReader(file, info.Size())
	if err != nil {
		log.Errorf("could not open as zip file: %s", err)
		return nil, err
	}

	for _, f := range reader.File {
		if f.Name == "rocket.ork" { // WARNING: File now exists in the zip archive (weird)
			rc, err := f.Open()
			if err != nil {
				log.Errorf("could not open 'rocket.ork': %s", err)
				return nil, err
			}
			defer func() {
				if cerr := rc.Close(); cerr != nil {
					log.Errorf("error closing zip file entry: %s", cerr)
				}
			}()

			xmlData, err := io.ReadAll(rc)
			if err != nil {
				log.Errorf("could not read XML data: %s", err)
				return nil, err
			}

			var rocket Openrocket
			err = xml.Unmarshal(xmlData, &rocket)
			if err != nil {
				log.Errorf("could not unmarshal XML data: %s", err)
				return nil, err
			}

			return &rocket, nil
		}
	}

	log.Errorf("'rocket.ork' not found in the zip archive")
	return nil, errors.Errorf("'rocket.ork' not found in the zip archive")

}
