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
	"github.com/zerodha/logf"
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
	log     *logf.Logger
}

// GetStorageDir returns the base directory for the record manager.
func (rm *RecordManager) GetStorageDir() string {
	return rm.baseDir
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

	log := logger.GetLogger("info") // Or a more appropriate level

	return &RecordManager{
		baseDir: baseDir,
		log:     log,
	}, nil
}

// CreateRecord creates a new record with a unique hash based on the current time
// DEPRECATED: Use CreateRecordWithConfig for deterministic hashing based on simulation parameters
func (rm *RecordManager) CreateRecord() (*Record, error) {
	pc, file, line, _ := runtime.Caller(1)
	log := logger.GetLogger("")
	log.Warn("DEPRECATED: CreateRecord called without config data. This may cause duplicate records.", "file", file, "line", line, "caller", runtime.FuncForPC(pc).Name())
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	// Use a fixed prefix with the current day (without time) to make simulations run on the same day 
	// have the same hash, avoiding duplicate result sets
	currentDate := time.Now().Format("2006-01-02")
	hashInput := fmt.Sprintf("simulation-%s", currentDate)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(hashInput)))

	// Use rm.baseDir as the root for record directories.
	// rm.baseDir should already be an absolute path, created by NewRecordManager.
	// MkdirAll is still good practice in case baseDir was removed externally or for subdirs.
	err := os.MkdirAll(rm.baseDir, 0755) // Ensure baseDir itself exists
	if err != nil {
		return nil, fmt.Errorf("failed to ensure base directory exists %s: %w", rm.baseDir, err)
	}

	// Create hash-named directory inside rm.baseDir
	recordDir := filepath.Join(rm.baseDir, hash)
	if err := os.MkdirAll(recordDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create record directory %s: %w", recordDir, err)
	}

	// Create record with its storage services
	return NewRecord(recordDir, hash)
}

// CreateRecordWithConfig creates a new record with a hash derived from configuration and OpenRocket data
func (rm *RecordManager) CreateRecordWithConfig(configData []byte, orkData []byte) (*Record, error) {
	pc, file, line, _ := runtime.Caller(1)
	log := logger.GetLogger("")
	log.Info("CreateRecordWithConfig called", "file", file, "line", line, "caller", runtime.FuncForPC(pc).Name())
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	// Hash the configuration and OpenRocket data
	combinedData := append(configData, orkData...)
	// Add a prefix to ensure consistency in hash generation
	combinedData = append([]byte("launchrail-config-"), combinedData...)
	hashInput := sha256.Sum256(combinedData)
	hash := fmt.Sprintf("%x", hashInput)
	
	// Truncate to a reasonable length (first 16 characters)
	hash = hash[:16]

	// Use rm.baseDir as the root for record directories.
	// rm.baseDir should already be an absolute path, created by NewRecordManager.
	// MkdirAll is still good practice in case baseDir was removed externally or for subdirs.
	err := os.MkdirAll(rm.baseDir, 0755) // Ensure baseDir itself exists
	if err != nil {
		return nil, fmt.Errorf("failed to ensure base directory exists %s: %w", rm.baseDir, err)
	}

	// Create hash-named directory inside rm.baseDir
	hashDir := filepath.Join(rm.baseDir, hash)
	err = os.MkdirAll(hashDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create hash directory %s: %w", hashDir, err)
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
	// Validate the hash to prevent directory traversal
	if strings.Contains(hash, "/") || strings.Contains(hash, "\\") || strings.Contains(hash, "..") {
		return fmt.Errorf("invalid hash: contains forbidden characters")
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Use rm.baseDir to construct the path to the record directory
	recordPath := filepath.Join(rm.baseDir, hash)

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

	// Use rm.baseDir instead of hardcoding to ~/.launchrail
	entries, err := os.ReadDir(rm.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", rm.baseDir, err)
	}

	var records []*Record
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Construct recordPath relative to rm.baseDir
		recordPath := filepath.Join(rm.baseDir, entry.Name())
		info, err := os.Stat(recordPath)
		if err != nil {
			rm.log.Warn("Failed to stat record, skipping", "path", recordPath, "error", err)
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

	// Construct path using rm.baseDir
	recordPath := filepath.Join(rm.baseDir, hash)
	if _, err := os.Stat(recordPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("record not found") // Consistent error message with loadRecord
		}
		return nil, fmt.Errorf("failed to stat record path %s: %w", recordPath, err)
	}

	// Initialize storage services for the record
	motionStore, err := NewStorage(recordPath, MOTION) // Use recordPath
	if err != nil {
		return nil, fmt.Errorf("failed to create motion storage for %s: %w", hash, err) // More context for error
	}
	if err := motionStore.Init(); err != nil {
		motionStore.Close()
		return nil, fmt.Errorf("failed to initialize motion storage for %s: %w", hash, err)
	}

	eventsStore, err := NewStorage(recordPath, EVENTS) // Use recordPath
	if err != nil {
		motionStore.Close() // Clean up previously successful store
		return nil, fmt.Errorf("failed to create events storage for %s: %w", hash, err)
	}
	if err := eventsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, fmt.Errorf("failed to initialize events storage for %s: %w", hash, err)
	}

	dynamicsStore, err := NewStorage(recordPath, DYNAMICS) // Use recordPath
	if err != nil {
		motionStore.Close()
		eventsStore.Close() // Clean up previously successful stores
		return nil, fmt.Errorf("failed to create dynamics storage for %s: %w", hash, err)
	}
	if err := dynamicsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		dynamicsStore.Close()
		return nil, fmt.Errorf("failed to initialize dynamics storage for %s: %w", hash, err)
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
	// Construct path using rm.baseDir
	recordPath := filepath.Join(rm.baseDir, hash)
	info, err := os.Stat(recordPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("record %s not found at %s", hash, recordPath)
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
	if err := motionStore.Init(); err != nil {
		motionStore.Close()
		return nil, fmt.Errorf("failed to initialize motionStore for %s: %w", hash, err)
	}
	eventsStore, err := NewStorage(recordPath, EVENTS)
	if err != nil {
		motionStore.Close()
		return nil, fmt.Errorf("failed to init events storage for %s: %w", hash, err)
	}
	if err := eventsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, fmt.Errorf("failed to initialize eventsStore for %s: %w", hash, err)
	}
	dynamicsStore, err := NewStorage(recordPath, DYNAMICS)
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, fmt.Errorf("failed to init dynamics storage for %s: %w", hash, err)
	}
	if err := dynamicsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		dynamicsStore.Close()
		return nil, fmt.Errorf("failed to initialize dynamicsStore for %s: %w", hash, err)
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
