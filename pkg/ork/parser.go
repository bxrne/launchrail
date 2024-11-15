package ork

import (
	"archive/zip"
	"encoding/xml"

	"github.com/pkg/errors"
)

func openOrkFile(input string) (*zip.ReadCloser, error) {
	r, err := zip.OpenReader(input)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// INFO: OpenRocket files .ork are actually xml but zipped
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
