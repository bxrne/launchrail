package storage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bxrne/launchrail/internal/logger"
)

const (
	MetadataFileName = "record_meta.json" // File to store reliable timestamp
	DataFileName     = "SIMULATION.json"  // Core data file for a record
)

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

type Metadata struct {
	CreationTime time.Time `json:"creationTime"`
}

// NewRecord creates a new simulation record with associated storage services
func NewRecord(baseDir string, hash string) (*Record, error) {
	motionStore, err := NewStorage(baseDir, hash, "motion")
	if err != nil {
		return nil, err
	}

	eventsStore, err := NewStorage(baseDir, hash, "events")
	if err != nil {
		motionStore.Close()
		return nil, err
	}

	dynamicsStore, err := NewStorage(baseDir, hash, "dynamics")
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, err
	}

	return &Record{
		Hash:         hash,
		Name:         hash,
		LastModified: time.Now(),
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

	// Create metadata file with creation time
	meta := Metadata{CreationTime: time.Now()}
	metaFilePath := filepath.Join(rm.baseDir, hash, MetadataFileName)
	metaFile, err := os.Create(metaFilePath)
	if err != nil {
		// Attempt cleanup if metadata creation fails
		_ = os.RemoveAll(filepath.Join(rm.baseDir, hash))
		return nil, fmt.Errorf("failed to create metadata file %s: %w", metaFilePath, err)
	}
	defer metaFile.Close()

	if err := json.NewEncoder(metaFile).Encode(meta); err != nil {
		// Attempt cleanup if encoding fails
		_ = os.RemoveAll(filepath.Join(rm.baseDir, hash))
		return nil, fmt.Errorf("failed to encode metadata to %s: %w", metaFilePath, err)
	}

	// Create an empty data file to mark the record as structurally valid initially
	dataFilePath := filepath.Join(rm.baseDir, hash, DataFileName)
	dataFile, err := os.Create(dataFilePath)
	if err != nil {
		_ = os.RemoveAll(filepath.Join(rm.baseDir, hash)) // Attempt cleanup
		return nil, fmt.Errorf("failed to create empty data file %s: %w", dataFilePath, err)
	}
	dataFile.Close()

	// Load the newly created record information
	newRecord, err := rm.loadRecord(hash)
	if err != nil {
		return nil, err
	}

	return newRecord, nil
}

// loadRecord loads record details from disk.
// Assumes the caller holds the appropriate lock.
func (rm *RecordManager) loadRecord(hash string) (*Record, error) {
	recordPath := filepath.Join(rm.baseDir, hash)
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

	// Prioritize reading CreationTime from metadata file
	creationTime := info.ModTime() // Fallback to directory ModTime
	metaFilePath := filepath.Join(recordPath, MetadataFileName)
	metaFile, err := os.Open(metaFilePath)
	if err == nil { // If metadata file exists and is readable
		var meta Metadata
		if json.NewDecoder(metaFile).Decode(&meta) == nil {
			creationTime = meta.CreationTime // Use timestamp from file
		}
		metaFile.Close()
	} // Ignore errors reading metadata, just use fallback

	// Check for the existence of core data files
	dataPath := filepath.Join(recordPath, DataFileName)
	_, err = os.Stat(dataPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("record %s is incomplete: missing %s", hash, DataFileName)
	}

	// Initialize storage handlers for the loaded record
	motionStore, err := NewStorage(rm.baseDir, hash, "motion")
	if err != nil {
		return nil, fmt.Errorf("failed to init motion storage for %s: %w", hash, err)
	}
	eventsStore, err := NewStorage(rm.baseDir, hash, "events")
	if err != nil {
		motionStore.Close() // Close already opened store
		return nil, fmt.Errorf("failed to init events storage for %s: %w", hash, err)
	}
	dynamicsStore, err := NewStorage(rm.baseDir, hash, "dynamics")
	if err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, fmt.Errorf("failed to init dynamics storage for %s: %w", hash, err)
	}

	return &Record{
		Hash:         hash,
		Name:         hash, // Or potentially load from metadata if stored
		LastModified: info.ModTime(), // Still store ModTime for potential other uses
		CreationTime: creationTime, // Use the reliable timestamp
		Motion:       motionStore,
		Events:       eventsStore,
		Dynamics:     dynamicsStore,
	}, nil
}

// DeleteRecord deletes a record by Hash
func (rm *RecordManager) DeleteRecord(hash string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	recordPath := filepath.Join(rm.baseDir, hash)
	if err := os.RemoveAll(recordPath); err != nil {
		return fmt.Errorf("failed to delete record: %v", err)
	}

	return nil
}

// ListRecords lists all existing valid records in the base directory.
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

		// Prioritize reading CreationTime from metadata file
		creationTime := info.ModTime() // Fallback to directory ModTime
		lastModified := info.ModTime()
		metaFilePath := filepath.Join(recordPath, MetadataFileName)
		metaFile, err := os.Open(metaFilePath)
		if err == nil { // If metadata file exists and is readable
			var meta Metadata
			if json.NewDecoder(metaFile).Decode(&meta) == nil {
				creationTime = meta.CreationTime // Use timestamp from file
			}
			metaFile.Close()
		} // Ignore errors reading metadata, just use fallback

		// Check if core data file exists to consider the record valid
		dataPath := filepath.Join(recordPath, DataFileName)
		_, dataStatErr := os.Stat(dataPath)
		if dataStatErr != nil {
			continue // Skip incomplete records
		}

		// Load the record without creating a new one
		record := &Record{
			Hash:         entry.Name(),
			Name:         entry.Name(), // Or potentially load from metadata if stored
			LastModified: lastModified,
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

	recordPath := filepath.Join(rm.baseDir, hash)
	if _, err := os.Stat(recordPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("record not found")
	}

	// Initialize storage services for the record
	motionStore, err := NewStorage(rm.baseDir, hash, "motion")
	if err != nil {
		return nil, err
	}

	eventsStore, err := NewStorage(rm.baseDir, hash, "events")
	if err != nil {
		motionStore.Close()
		return nil, err
	}

	dynamicsStore, err := NewStorage(rm.baseDir, hash, "dynamics")
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

	// Prioritize reading CreationTime from metadata file
	creationTime := info.ModTime() // Fallback to directory ModTime
	metaFilePath := filepath.Join(recordPath, MetadataFileName)
	metaFile, err := os.Open(metaFilePath)
	if err == nil { // If metadata file exists and is readable
		var meta Metadata
		if json.NewDecoder(metaFile).Decode(&meta) == nil {
			creationTime = meta.CreationTime // Use timestamp from file
		}
		metaFile.Close()
	} // Ignore errors reading metadata, just use fallback

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
