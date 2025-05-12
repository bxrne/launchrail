package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/diff"
	"github.com/bxrne/launchrail/pkg/openrocket"
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
func NewRecord(baseDir string, hash string, appCfg *config.Config) (*Record, error) {
	recordDir := baseDir // Use the baseDir directly as it already includes the hash

	motionStore, err := NewStorage(recordDir, MOTION, appCfg) // Use recordDir and appCfg
	if err != nil {
		return nil, err
	}
	if err := motionStore.Init(); err != nil {
		motionStore.Close()
		return nil, fmt.Errorf("failed to initialize motion storage: %w", err)
	}

	eventsStore, err := NewStorage(recordDir, EVENTS, appCfg) // Use recordDir and appCfg
	if err != nil {
		motionStore.Close()
		return nil, err
	}
	if err := eventsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, fmt.Errorf("failed to initialize events storage: %w", err)
	}

	dynamicsStore, err := NewStorage(recordDir, DYNAMICS, appCfg) // Use recordDir and appCfg
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
	appCfg  *config.Config // Added to store application-level config
	records map[string]*Record
	stopCh  chan struct{}
}

// GetStorageDir returns the base directory for the record manager.
func (rm *RecordManager) GetStorageDir() string {
	return rm.baseDir
}

// NewRecordManager creates a new RecordManager.
// It requires the application config, the base directory for records, and a logger.
func NewRecordManager(cfg *config.Config, baseDir string, log *logf.Logger) (*RecordManager, error) {
	// Ensure cfg is not nil to prevent panic when accessing cfg.Setup.Logging.Level
	if cfg == nil {
		return nil, fmt.Errorf("application configuration (cfg) cannot be nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Validate and potentially make baseDir absolute
	if !filepath.IsAbs(baseDir) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, baseDir)
	}

	// Ensure base directory exists
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return nil, err
	}

	// Create RecordManager instance
	rm := &RecordManager{
		baseDir: baseDir,
		appCfg:  cfg,
		log:     log,
		records: make(map[string]*Record),
		stopCh:  make(chan struct{}),
	}

	return rm, nil
}

// CreateRecord creates a new record with a unique hash based on the current time
func (rm *RecordManager) CreateRecord(cfg *config.Config) (*Record, error) {
	pc, file, line, _ := runtime.Caller(1)
	rm.log.Info("CreateRecord called (Note: uses time-based hash)", "file", file, "line", line, "caller", runtime.FuncForPC(pc).Name())
	rm.mu.Lock()
	defer rm.mu.Unlock()

	ork, err := openrocket.Load(cfg.Engine.Options.OpenRocketFile, cfg.Engine.External.OpenRocketVersion)
	if err != nil {
		return nil, err
	}

	hash := diff.CombinedHash(cfg.Bytes(), ork.Bytes())

	// Use rm.baseDir as the root for record directories.
	// rm.baseDir should already be an absolute path, created by NewRecordManager.
	// MkdirAll is still good practice in case baseDir was removed externally or for subdirs.
	err = os.MkdirAll(rm.baseDir, 0755) // Ensure baseDir itself exists
	if err != nil {
		return nil, fmt.Errorf("failed to ensure base directory exists %s: %w", rm.baseDir, err)
	}

	// Create hash-named directory inside rm.baseDir
	recordDir := filepath.Join(rm.baseDir, hash)
	if err := os.MkdirAll(recordDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create record directory %s: %w", recordDir, err)
	}

	// Create record with its storage services
	return NewRecord(recordDir, hash, cfg)
}

// CreateRecordWithConfig creates a new record with a hash derived from configuration and OpenRocket data
func (rm *RecordManager) CreateRecordWithConfig(configData []byte, orkData []byte) (*Record, error) {
	pc, file, line, _ := runtime.Caller(1)
	rm.log.Info("CreateRecordWithConfig called", "file", file, "line", line, "caller", runtime.FuncForPC(pc).Name()) // Uses rm.log which depends on logf
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Use diff.CombinedHash to generate the hash from config and ORK data
	hash := diff.CombinedHash(configData, orkData)

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

	record, err := NewRecord(hashDir, hash, rm.appCfg)
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
	if _, err := os.Stat(recordPath); err != nil {
		if os.IsNotExist(err) {
			return ErrRecordNotFound // Use the defined sentinel error
		}
		return fmt.Errorf("failed to check record existence: %w", err)
	}

	// Directory exists, proceed with deletion
	rm.log.Debug("Attempting to delete record directory", "recordPath", recordPath, "baseDir", rm.baseDir)
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

	rm.log.Debug("ListRecords called", "baseDir", rm.baseDir) // Log baseDir

	entries, err := os.ReadDir(rm.baseDir)
	if err != nil { // Error reading base directory
		rm.log.Error("ListRecords: Failed to read base directory", "baseDir", rm.baseDir, "error", err)
		return nil, fmt.Errorf("failed to read directory %s: %w", rm.baseDir, err)
	}

	// Regex to match a 16 or 64-character hexadecimal string
	hashRegex := regexp.MustCompile(`^([0-9a-f]{16}|[0-9a-f]{64})$`)

	var records []*Record
	for _, entry := range entries {
		entryName := entry.Name()
		rm.log.Debug("ListRecords: Processing entry", "name", entryName, "isDir", entry.IsDir())

		if !entry.IsDir() { // Skip if not a directory
			rm.log.Debug("ListRecords: Skipping non-directory entry", "name", entryName)
			continue
		}

		if !hashRegex.MatchString(entryName) { // Skip if name doesn't match 16/64 hex pattern
			rm.log.Debug("ListRecords: Skipping entry with non-matching name", "name", entryName)
			continue
		}
		rm.log.Debug("ListRecords: Entry name matched hash regex", "name", entryName)

		// Passed checks, create basic record struct
		recordPath := filepath.Join(rm.baseDir, entryName)
		rm.log.Debug("ListRecords: Attempting to stat record path", "path", recordPath)
		info, err := os.Stat(recordPath)
		if err != nil { // Error stating the specific record directory
			rm.log.Warn("ListRecords: Failed to stat record path, skipping", "path", recordPath, "error", err)
			continue // Skip this entry
		}
		rm.log.Debug("ListRecords: Successfully stated record path", "path", recordPath)

		creationTime := info.ModTime() // Note: Using ModTime for CreationTime

		record := &Record{
			Hash:         entryName,
			Name:         entryName,
			LastModified: info.ModTime(),
			CreationTime: creationTime, // Uses ModTime
			Path:         recordPath,   // Added Path field? Assumed based on common practice
			Motion:       nil,          // Explicitly nil
			Events:       nil,          // Explicitly nil
			Dynamics:     nil,          // Explicitly nil
		}
		rm.log.Debug("ListRecords: Appending record", "hash", entryName)
		records = append(records, record)
	}

	rm.log.Debug("ListRecords finished", "record_count", len(records))
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
	motionStore, err := NewStorage(recordPath, MOTION, rm.appCfg) // Pass rm.appCfg
	if err != nil {
		return nil, fmt.Errorf("failed to create motion storage for %s: %w", hash, err) // More context for error
	}
	if err := motionStore.Init(); err != nil {
		motionStore.Close()
		return nil, fmt.Errorf("failed to initialize motion storage for %s: %w", hash, err)
	}

	eventsStore, err := NewStorage(recordPath, EVENTS, rm.appCfg) // Pass rm.appCfg
	if err != nil {
		motionStore.Close() // Clean up previously successful store
		return nil, fmt.Errorf("failed to create events storage for %s: %w", hash, err)
	}
	if err := eventsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, fmt.Errorf("failed to initialize events storage for %s: %w", hash, err)
	}

	dynamicsStore, err := NewStorage(recordPath, DYNAMICS, rm.appCfg) // Pass rm.appCfg
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
	motionStore, err := NewStorage(recordPath, MOTION, rm.appCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init motion storage for %s: %w", hash, err)
	}
	if err := motionStore.Init(); err != nil {
		motionStore.Close()
		return nil, fmt.Errorf("failed to initialize motionStore for %s: %w", hash, err)
	}
	eventsStore, err := NewStorage(recordPath, EVENTS, rm.appCfg)
	if err != nil {
		motionStore.Close()
		return nil, fmt.Errorf("failed to init events storage for %s: %w", hash, err)
	}
	if err := eventsStore.Init(); err != nil {
		motionStore.Close()
		eventsStore.Close()
		return nil, fmt.Errorf("failed to initialize eventsStore for %s: %w", hash, err)
	}
	dynamicsStore, err := NewStorage(recordPath, DYNAMICS, rm.appCfg)
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
