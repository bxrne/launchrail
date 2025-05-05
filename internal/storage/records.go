package storage

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bxrne/launchrail/internal/logger"
)

// No JSON metadata or data files are produced per record; only CSVs

type Record struct {
	Name         string    `json:"name"`
	Hash         string    `json:"hash"`
	LastModified time.Time `json:"lastModified"` // Keep for potential compatibility, but prioritize CreationTime
	CreationTime time.Time `json:"creationTime"` // More reliable timestamp
	Path         string
	Motion       *Storage
	Events       *Storage
	Dynamics     *Storage
}

// NewRecord creates a new simulation record with associated storage services
func NewRecord(baseDir string, hash string) (*Record, error) {
	recordDir := baseDir // Use the baseDir directly as it already includes the hash

	motionStore, err := NewStorage(recordDir, MOTION) // Use recordDir
	if err != nil {
		return nil, err
	}
	if err := motionStore.Init(); err != nil {
		motionStore.Close()
		return nil, fmt.Errorf("failed to initialize motion storage: %w", err)
	}

	eventsStore, err := NewStorage(recordDir, EVENTS) // Use recordDir
	if err != nil {
		motionStore.Close()
		return nil, err
	}
	if err := eventsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, fmt.Errorf("failed to initialize events storage: %w", err)
	}

	dynamicsStore, err := NewStorage(recordDir, DYNAMICS) // Use recordDir
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, err
	}
	if err := dynamicsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		dynamicsStore.Close()
		return nil, fmt.Errorf("failed to initialize dynamics storage: %w", err)
	}

	return &Record{
		Hash:         hash,
		Name:         hash,
		LastModified: time.Now(),
		Path:         recordDir, // Store the path
		Motion:       motionStore,
		Events:       eventsStore,
		Dynamics:     dynamicsStore,
	}, nil
}

// Close closes all associated storage services
func (r *Record) Close() error {
	var errs []error
	if err := r.Motion.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.Events.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.Dynamics.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close one or more stores: %v", errs)
	}
	return nil
}

// RecordManager manages simulation records
type RecordManager struct {
	baseDir string
	mu      sync.RWMutex
}

func NewRecordManager(baseDir string) (*RecordManager, error) {
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

	return &RecordManager{
		baseDir: baseDir,
	}, nil
}

// CreateRecord creates a new record with a unique hash
func (rm *RecordManager) CreateRecord() (*Record, error) {
	pc, file, line, _ := runtime.Caller(1)
	log := logger.GetLogger("")
	log.Info("CreateRecord called", "file", file, "line", line, "caller", runtime.FuncForPC(pc).Name())
	rm.mu.Lock()
	defer rm.mu.Unlock()
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(time.Now().String())))

	// Ensure .launchrail directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	launchrailDir := filepath.Join(homeDir, ".launchrail")
	err = os.MkdirAll(launchrailDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create .launchrail directory: %w", err)
	}

	// Create hash-named directory inside .launchrail
	hashDir := filepath.Join(launchrailDir, hash)
	err = os.MkdirAll(hashDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create hash directory: %w", err)
	}

	record, err := NewRecord(hashDir, hash)
	if err != nil {
		return nil, err
	}

	// Initialize storage
	if err := record.Motion.Init(); err != nil {
		record.Close()
		return nil, err
	}
	if err := record.Events.Init(); err != nil {
		record.Close()
		return nil, err
	}
	if err := record.Dynamics.Init(); err != nil {
		record.Close()
		return nil, err
	}

	// Load the newly created record information
	newRecord, err := rm.loadRecord(hash)
	if err != nil {
		return nil, err
	}

	return newRecord, nil
}

// DeleteRecord deletes a record by Hash
func (rm *RecordManager) DeleteRecord(hash string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	launchrailDir := filepath.Join(homeDir, ".launchrail")
	recordPath := filepath.Join(launchrailDir, hash)

	// Check if the record directory exists first
	if _, err := os.Stat(recordPath); os.IsNotExist(err) {
		// Directory does not exist, return specific error
		return ErrRecordNotFound // Use the defined sentinel error
	} else if err != nil {
		// Some other error occurred during stat (e.g., permissions)
		return fmt.Errorf("failed to check record existence: %w", err)
	}

	// Directory exists, proceed with deletion
	if err := os.RemoveAll(recordPath); err != nil {
		// Error during deletion
		return fmt.Errorf("failed to delete record directory: %w", err)
	}

	return nil // Deletion successful
}

// ListRecords lists all existing valid records in the base directory.
func (rm *RecordManager) ListRecords() ([]*Record, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	launchrailDir := filepath.Join(homeDir, ".launchrail")
	entries, err := os.ReadDir(launchrailDir)
	if err != nil {
		return nil, err
	}

	var records []*Record
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		recordPath := filepath.Join(launchrailDir, entry.Name())
		info, err := os.Stat(recordPath)
		if err != nil {
			continue // Skip invalid records
		}

		creationTime := info.ModTime()

		// Load the record without creating a new one
		record := &Record{
			Hash:         entry.Name(),
			Name:         entry.Name(), // Or potentially load from metadata if stored
			LastModified: info.ModTime(),
			CreationTime: creationTime,
			Motion:       nil,
			Events:       nil,
			Dynamics:     nil,
		}
		records = append(records, record)
	}

	return records, nil
}

// GetRecord retrieves an existing record by hash without creating a new one.
func (rm *RecordManager) GetRecord(hash string) (*Record, error) {
	// Validate the hash to ensure it is a valid directory name
	if strings.Contains(hash, "/") || strings.Contains(hash, "\\") || strings.Contains(hash, "..") {
		return nil, fmt.Errorf("invalid hash value")
	}

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	launchrailDir := filepath.Join(homeDir, ".launchrail")
	recordPath := filepath.Join(launchrailDir, hash)
	if _, err := os.Stat(recordPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("record not found")
		}
		return nil, fmt.Errorf("failed to stat record path: %w", err)
	}

	// Initialize storage services for the record
	motionStore, err := NewStorage(recordPath, MOTION) // Use recordPath
	if err != nil {
		return nil, err
	}

	eventsStore, err := NewStorage(recordPath, EVENTS) // Use recordPath
	if err != nil {
		motionStore.Close()
		return nil, err
	}

	dynamicsStore, err := NewStorage(recordPath, DYNAMICS) // Use recordPath
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, err
	}

	// Get last modified time
	info, err := os.Stat(recordPath)
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		dynamicsStore.Close()
		return nil, err
	}

	creationTime := info.ModTime()

	return &Record{
		Hash:         hash,
		Name:         hash,
		LastModified: info.ModTime(),
		CreationTime: creationTime,
		Motion:       motionStore,
		Events:       eventsStore,
		Dynamics:     dynamicsStore,
	}, nil
}

// loadRecord loads a record by hash and initializes its CSV stores.
func (rm *RecordManager) loadRecord(hash string) (*Record, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	recordPath := filepath.Join(homeDir, ".launchrail", hash)
	info, err := os.Stat(recordPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("record %s not found", hash)
		}
		return nil, fmt.Errorf("failed to stat record directory %s: %w", recordPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", recordPath)
	}
	creationTime := info.ModTime()
	motionStore, err := NewStorage(recordPath, MOTION)
	if err != nil {
		return nil, fmt.Errorf("failed to init motion storage for %s: %w", hash, err)
	}
	eventsStore, err := NewStorage(recordPath, EVENTS)
	if err != nil {
		motionStore.Close()
		return nil, fmt.Errorf("failed to init events storage for %s: %w", hash, err)
	}
	dynamicsStore, err := NewStorage(recordPath, DYNAMICS)
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, fmt.Errorf("failed to init dynamics storage for %s: %w", hash, err)
	}
	return &Record{
		Hash:         hash,
		Name:         hash,
		LastModified: info.ModTime(),
		CreationTime: creationTime,
		Path:         recordPath,
		Motion:       motionStore,
		Events:       eventsStore,
		Dynamics:     dynamicsStore,
	}, nil
}
