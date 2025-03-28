package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// StorageType is the type of storage service (MOTION, EVENTS, etc.)
type StorageType string

const (
	// MOTION storage StorageType
	MOTION StorageType = "MOTION"
	// EVENTS storage StorageType
	EVENTS StorageType = "EVENTS"
	// DYNAMICS storage StorageType
	DYNAMICS StorageType = "DYNAMICS"
)

// StorageHeaders is a map of columns for storage types
var StorageHeaders = map[StorageType][]string{
	MOTION: {
		"time", "altitude", "velocity", "acceleration", "thrust",
	},
	EVENTS: {
		"time", "motor_status", "parachute_status",
	},
	DYNAMICS: {
		"time", "position_x", "position_y", "position_z", "velocity_x", "velocity_y", "velocity_z", "acceleration_x", "acceleration_y", "acceleration_z", "orientation_x", "orientation_y", "orientation_z", "orientation_w",
	},
}

// Storage is a service that writes csv's to disk
type Storage struct {
	baseDir  string
	dir      string
	store    StorageType
	mu       sync.RWMutex
	filePath string
	writer   *csv.Writer
	file     *os.File
}

// Stores is a collection of storage services
type Stores struct {
	Motion   *Storage
	Events   *Storage
	Dynamics *Storage
}

// NewStorage creates a new storage service.
// If the provided baseDir is not absolute, it is prepended with the user's home directory.
func NewStorage(baseDir string, dir string, store StorageType) (*Storage, error) {
	// If baseDir is not absolute, prepend the user's home directory.
	if !filepath.IsAbs(baseDir) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		baseDir = filepath.Join(homeDir, baseDir)
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	dir = filepath.Join(baseDir, dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dir, fmt.Sprintf("%s.csv", store))

	// Open file in read/write mode with append flag.
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create/open file: %v", err)
	}

	return &Storage{
		baseDir:  baseDir,
		dir:      dir,
		store:    store,
		filePath: filePath,
		file:     file,
		writer:   csv.NewWriter(file),
	}, nil
}

// Init initializes the storage service with headers.
func (s *Storage) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Truncate file before writing headers
	if err := s.file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate file: %v", err)
	}

	// Reset file pointer to beginning
	if _, err := s.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to beginning: %v", err)
	}

	// Write headers
	headers := StorageHeaders[s.store]
	if err := s.writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}
	s.writer.Flush()
	if err := s.writer.Error(); err != nil {
		return fmt.Errorf("failed to flush headers: %v", err)
	}

	return nil
}

// Write writes a record to the storage service.
func (s *Storage) Write(data []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(data) != len(StorageHeaders[s.store]) {
		return fmt.Errorf("data length (%d) does not match headers length (%d)", len(data), len(StorageHeaders[s.store]))
	}

	if err := s.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	if err := s.writer.Error(); err != nil {
		return fmt.Errorf("failed to flush data: %v", err)
	}
	s.writer.Flush()

	return nil
}

// Close closes the storage service.
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

// GetFilePath returns the file path of the storage service.
func (s *Storage) GetFilePath() string {
	return s.filePath
}
