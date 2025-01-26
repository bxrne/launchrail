package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

// Storage is a service that writes csv's to disk
type Storage struct {
	baseDir string
	dir     string
	mu      sync.RWMutex
	headers []string
}

// NewStorage creates a new storage service
func NewStorage(baseDir, dir string) (*Storage, error) {
	// Create the base directory if it doesn't exist
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		err := os.Mkdir(baseDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	// Create the directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return &Storage{
		baseDir: baseDir,
		dir:     dir,
	}, nil
}

func (s *Storage) Init(headers []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.headers = headers

	// Create the file in dir with timestamp as name
	timestamp := time.Now().Format("2006-01-02T15:04:05")
	filename := fmt.Sprintf("%s/%s.csv", s.dir, timestamp)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	// Write the headers to the filename
	writer := csv.NewWriter(file)
	writer.Write(headers)
	writer.Flush()
	file.Close()

	return nil
}

func (s *Storage) Write(data []string) error {
	// Append the data to the file as long as it meets the headers
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(data) != len(s.headers) {
		return fmt.Errorf("data length does not match headers length")
	}

	// Open the file in dir with timestamp as name
	timestamp := time.Now().Format("2006-01-02T15:04:05")
	filename := fmt.Sprintf("%s/%s.csv", s.dir, timestamp)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}

	// Write the data to the filename
	writer := csv.NewWriter(file)

	writer.Write(data)
	writer.Flush()

	file.Close()

	return nil
}
