package storage

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Record struct {
	Hash      string    `json:"hash"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Motion    *Storage  `json:"-"`
	Events    *Storage  `json:"-"`
	Dynamics  *Storage  `json:"-"`
}

// NewRecord creates a new simulation record with associated storage services
func NewRecord(baseDir string, hash string) (*Record, error) {
	motionStore, err := NewStorage(baseDir, hash, MOTION)
	if err != nil {
		return nil, err
	}

	eventsStore, err := NewStorage(baseDir, hash, EVENTS)
	if err != nil {
		motionStore.Close()
		return nil, err
	}

	dynamicsStore, err := NewStorage(baseDir, hash, DYNAMICS)
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, err
	}

	return &Record{
		Hash:      hash,
		Name:      hash,
		Timestamp: time.Now(),
		Motion:    motionStore,
		Events:    eventsStore,
		Dynamics:  dynamicsStore,
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
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Generate unique hash
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(time.Now().String())))

	record, err := NewRecord(rm.baseDir, hash)
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

	return record, nil
}

// ListRecords lists all existing records in the base directory with their last modified time.
func (rm *RecordManager) ListRecords() ([]*Record, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	entries, err := os.ReadDir(rm.baseDir)
	if err != nil {
		return nil, err
	}

	var records []*Record
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		recordPath := filepath.Join(rm.baseDir, entry.Name())
		info, err := os.Stat(recordPath)
		if err != nil {
			continue // Skip invalid records
		}

		// Load the record without creating a new one
		record := &Record{
			Hash:      entry.Name(),
			Name:      entry.Name(),
			Timestamp: info.ModTime(),
		}
		records = append(records, record)
	}

	return records, nil
}

// GetRecord retrieves an existing record by hash without creating a new one.
func (rm *RecordManager) GetRecord(hash string) (*Record, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	recordPath := filepath.Join(rm.baseDir, hash)
	if _, err := os.Stat(recordPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("record not found")
	}

	// Initialize storage services for the record
	motionStore, err := NewStorage(rm.baseDir, hash, MOTION)
	if err != nil {
		return nil, err
	}

	eventsStore, err := NewStorage(rm.baseDir, hash, EVENTS)
	if err != nil {
		motionStore.Close()
		return nil, err
	}

	dynamicsStore, err := NewStorage(rm.baseDir, hash, DYNAMICS)
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, err
	}

	return &Record{
		Hash:      hash,
		Name:      hash,
		Timestamp: time.Now(),
		Motion:    motionStore,
		Events:    eventsStore,
		Dynamics:  dynamicsStore,
	}, nil
}
