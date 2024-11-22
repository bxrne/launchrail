// Package openrocket provides functionality to parse and interact with OpenRocket (.ork) files.
// INFO: OpenRocket files are XML documents compressed in ZIP format.
package openrocket

import (
	"archive/zip"
	"encoding/xml"

	"github.com/pkg/errors"
)

// openOrkFile opens an OpenRocket file and returns a zip reader.
// WARN: Caller must ensure to close the returned ReadCloser.
func openOrkFile(input string) (*zip.ReadCloser, error) {
	r, err := zip.OpenReader(input)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Decompress takes a filepath to an OpenRocket file (.ork) and returns the parsed rocket data.
// INFO: The function handles both unzipping and XML parsing in one step.
func Decompress(filePath string) (*Openrocket, error) {
	ork_rc, err := openOrkFile(filePath)
	if err != nil {
		return nil, err
	}
	defer ork_rc.Close()

	var rocket Openrocket
	for _, f := range ork_rc.File {
		if f.Name == "rocket.ork" {
			rc, err := f.Open()
			if err != nil {
				return nil, errors.Wrap(err, "opening rocket.ork")
			}
			defer rc.Close()

			err = xml.NewDecoder(rc).Decode(&rocket)
			if err != nil {
				return nil, errors.Wrap(err, "decoding rocket.ork")
			}
		}
	}

	return &rocket, nil

}
