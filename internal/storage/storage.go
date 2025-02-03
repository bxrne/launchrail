package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Storage is a service that writes csv's to disk
type Storage struct {
	baseDir  string
	dir      string
	mu       sync.RWMutex
	headers  []string
	filePath string
	writer   *csv.Writer
	file     *os.File
}

// NewStorage creates a new storage service
func NewStorage(baseDir, dir string) (*Storage, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	baseDir = filepath.Join(homeDir, baseDir)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	dir = filepath.Join(baseDir, dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Create the file with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filePath := filepath.Join(dir, fmt.Sprintf("simulation_%s.csv", timestamp))

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}

	return &Storage{
		baseDir:  baseDir,
		dir:      dir,
		filePath: filePath,
		file:     file,
		writer:   csv.NewWriter(file),
	}, nil
}

func (s *Storage) Init(headers []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.headers = headers
	if err := s.writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}
	s.writer.Flush()
	return nil
}

func (s *Storage) Write(data []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(data) != len(s.headers) {
		return fmt.Errorf("data length (%d) does not match headers length (%d)", len(data), len(s.headers))
	}

	// Write record and immediately flush to ensure it's written to disk
	if err := s.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}
	s.writer.Flush()

	// Check for flush errors
	if err := s.writer.Error(); err != nil {
		return fmt.Errorf("failed to flush data: %v", err)
	}

	return nil
}

func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.writer != nil {
		s.writer.Flush()
		if err := s.writer.Error(); err != nil {
			return fmt.Errorf("failed to flush on close: %v", err)
		}
	}

	if s.file != nil {
		if err := s.file.Sync(); err != nil {
			return fmt.Errorf("failed to sync file: %v", err)
		}
		return s.file.Close()
	}
	return nil
}

func (s *Storage) GetFilePath() string {
	return s.filePath
}
