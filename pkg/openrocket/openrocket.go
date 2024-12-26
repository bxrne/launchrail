package openrocket

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func Load(filename string) (*OpenrocketDocument, error) {
	data, err := extractORK(filename)
	if err != nil {
		return nil, err
	}

	// marshal the .ork file content into a OpenrocketDocument struct
	var doc OpenrocketDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// extractORK extracts the .ork file content from the zip archive
func extractORK(filePath string) ([]byte, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	for _, f := range reader.File {
		if strings.HasSuffix(f.Name, ".ork") {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}
	return nil, fmt.Errorf("no .ork file found in the zip archive")
}
