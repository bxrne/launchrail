package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bxrne/launchrail/internal/logger"
	"github.com/zerodha/logf"
)

// SimStorageType is the type of storage service (MOTION, EVENTS, etc.)
type SimStorageType string

const (
	// MOTION storage SimStorageType
	MOTION SimStorageType = "MOTION"
	// EVENTS storage SimStorageType
	EVENTS SimStorageType = "EVENTS"
	// DYNAMICS storage SimStorageType
	DYNAMICS SimStorageType = "DYNAMICS"
)

// StorageHeaders is a map of columns for storage types
var StorageHeaders = map[SimStorageType][]string{
	MOTION: {
		"time", "altitude", "velocity", "acceleration", "thrust",
	},
	EVENTS: {
		"time", "event_name", "motor_status", "parachute_status",
	},
	DYNAMICS: {
		"time", "position_x", "position_y", "position_z", "velocity_x", "velocity_y", "velocity_z", "acceleration_x", "acceleration_y", "acceleration_z", "orientation_x", "orientation_y", "orientation_z", "orientation_w",
	},
}

// Storage is a service that writes csv's to disk
type Storage struct {
	recordDir string
	store     SimStorageType
	mu        sync.RWMutex
	filePath  string
	writer    *csv.Writer
	file      *os.File
	log       *logf.Logger
	closed    bool
}

// Stores is a collection of storage services
type Stores struct {
	Motion   *Storage
	Events   *Storage
	Dynamics *Storage
}

// NewStorage creates a new storage service for a specific store type within a given record directory.
func NewStorage(recordDir string, store SimStorageType) (*Storage, error) {
	// Ensure the record directory path is absolute
	absRecordDir, err := filepath.Abs(recordDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for record directory %s: %w", recordDir, err)
	}

	// Ensure the record directory exists
	if err := os.MkdirAll(absRecordDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create record directory %s: %w", absRecordDir, err)
	}

	// Construct the specific file path within the record directory
	// Always use uppercase for consistency with SimStorageType constants
	filePath := filepath.Join(absRecordDir, fmt.Sprintf("%s.csv", strings.ToUpper(string(store))))

	// Open file in read/write mode with append flag.
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create/open file %s: %w", filePath, err)
	}

	return &Storage{
		recordDir: absRecordDir,
		store:     store,
		filePath:  filePath,
		file:      file,
		writer:    csv.NewWriter(file),
		closed:    false,
	}, nil
}

// Init ensures the header row is written if the file is new/empty.
// It is non-destructive if the file already appears to be initialized.
func (s *Storage) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log = logger.GetLogger("debug") // Ensure logger is always initialized

	// Check file size to determine if it's new or empty
	fileInfo, err := s.file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats for %s: %w", s.filePath, err)
	}

	if fileInfo.Size() == 0 {
		// File is empty, proceed to write headers
		// Truncate is not strictly necessary for a 0-size file but ensures clean state if somehow not 0 but intended to be new.
		if err := s.file.Truncate(0); err != nil {
			return fmt.Errorf("failed to truncate empty file %s: %w", s.filePath, err)
		}
		if _, err := s.file.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek to beginning of empty file %s: %w", s.filePath, err)
		}

		headers := StorageHeaders[s.store]
		s.log.Debug(fmt.Sprintf("Initializing storage with headers for %s: headers=%v", s.filePath, headers))
		if err := s.writer.Write(headers); err != nil {
			s.log.Error(fmt.Sprintf("failed to write headers to %s: %v", s.filePath, err))
			// Return the error to indicate Init failure
			return fmt.Errorf("failed to write headers to %s: %w", s.filePath, err)
		}
		s.writer.Flush()
		if err := s.writer.Error(); err != nil {
			s.log.Error(fmt.Sprintf("failed to flush headers to %s: %v", s.filePath, err))
			// Return the error to indicate Init failure
			return fmt.Errorf("failed to flush headers to %s: %w", s.filePath, err)
		}
		s.log.Debug(fmt.Sprintf("Successfully initialized headers for %s", s.filePath))
	} else {
		// File is not empty, assume it's already initialized with headers.
		// We need to ensure the CSV writer is ready for subsequent writes, even if we didn't write headers now.
		// For csv.NewWriter, it's generally fine as it writes to the underlying io.Writer (s.file).
		// However, the file pointer might be at the end. For reading, ReadAll seeks to start.
		// For writing, we want to append, which os.O_APPEND handles at OS level.
		// Ensure the writer is reset or aware of the file state if necessary, though typically it appends.
		s.log.Debug(fmt.Sprintf("Storage file %s already exists and is not empty, assuming initialized.", s.filePath))
	}

	return nil
}

// Write writes a record to the storage service.
func (s *Storage) Write(data []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	headers := StorageHeaders[s.store]
	s.log.Debug(fmt.Sprintf("Writing to %s: headers=%v, data=%v", s.filePath, headers, data))
	if len(data) != len(headers) {
		s.log.Error(fmt.Sprintf("data length (%d) does not match headers length (%d)", len(data), len(headers)))
		// It's important to return an error here if this condition is problematic
		return fmt.Errorf("data length (%d) does not match headers length (%d)", len(data), len(headers))
	}

	if err := s.writer.Write(data); err != nil {
		s.log.Error(fmt.Sprintf("failed to write data: %v", err))
	}

	if err := s.writer.Error(); err != nil {
		s.log.Error(fmt.Sprintf("failed to flush data: %v", err))
	}
	s.writer.Flush()

	return nil
}

// Close closes the storage service.
func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	var firstErr error

	// Helper to capture and log the first error encountered
	// It ensures that subsequent operations are still attempted.
	setErr := func(err error, context string) {
		if err != nil {
			if s.log != nil { // s.log might be nil if Init() was never called or failed
				s.log.Error(fmt.Sprintf("error during %s: %v", context, err))
			}
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", context, err)
			}
		}
	}

	if s.writer != nil {
		s.writer.Flush() // Attempt to flush any buffered data
		setErr(s.writer.Error(), "csv writer flush/error")
	}

	if s.file != nil {
		setErr(s.file.Sync(), "file sync")
		setErr(s.file.Close(), "file close")
	}

	s.closed = true
	return firstErr
}

// GetFilePath returns the file path of the storage service.
func (s *Storage) GetFilePath() string {
	return s.filePath
}

// ReadAll reads all data from the storage file
func (s *Storage) ReadAll() ([][]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Seek to the beginning of the file
	if _, err := s.file.Seek(0, 0); err != nil {
		s.log.Error(fmt.Sprintf("failed to seek to beginning: %v", err))
	}

	reader := csv.NewReader(s.file)
	allData, err := reader.ReadAll()
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to read CSV data: %v", err))
	}

	// Ensure there is at least one row (headers)
	if len(allData) == 0 {
		s.log.Error("no data found in storage")
	}

	return allData, nil
}

// ReadHeadersAndData reads the headers and data separately from the storage file
func (s *Storage) ReadHeadersAndData() ([]string, [][]string, error) {
	allData, err := s.ReadAll()
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to read all data: %v", err))
	}

	// Separate headers and data
	headers := allData[0]
	data := allData[1:] // Skip the first row (headers)

	return headers, data, nil
}

type StorageInterface interface {
	Init() error
	Write([]string) error
	Close() error
}
