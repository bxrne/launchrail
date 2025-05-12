package storage_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	storage "github.com/bxrne/launchrail/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zerodha/logf"
)

// Dummy data for CreateRecordWithConfig
var (
	dummyConfigData = []byte(`{"test": "config"}`)
	dummyOrkData    = []byte(`<testork/>`)
)

// TEST: GIVEN a record manager WHEN we create and close a record THEN no error is returned
func TestNewRecordAndClose(t *testing.T) {
	cfg := &config.Config{}
	log := logf.New(logf.Opts{Level: logf.ErrorLevel})
	tempDir := t.TempDir()
	rm, err := storage.NewRecordManager(cfg, tempDir, &log)
	require.NoError(t, err)

	// Use CreateRecordWithConfig
	rec, err := rm.CreateRecordWithConfig(dummyConfigData, dummyOrkData)
	assert.NoError(t, err)
	assert.NotNil(t, rec)
	assert.NoError(t, rec.Close())
}

// TEST: GIVEN a record manager WHEN multiple records exist THEN they are listed
func TestRecordManagerListRecords(t *testing.T) {
	cfg := &config.Config{}
	log := logf.New(logf.Opts{Level: logf.ErrorLevel})
	tempDir := t.TempDir()
	rm, err := storage.NewRecordManager(cfg, tempDir, &log)
	require.NoError(t, err)

	// Use CreateRecordWithConfig
	_, err = rm.CreateRecordWithConfig(dummyConfigData, dummyOrkData)
	assert.NoError(t, err)

	records, err := rm.ListRecords()
	require.NoError(t, err)
	require.NotEmpty(t, records)
}

// TEST: GIVEN a record manager WHEN we retrieve a record by hash THEN the record is returned
func TestRecordManagerGetRecord(t *testing.T) {
	cfg := &config.Config{}
	log := logf.New(logf.Opts{Level: logf.ErrorLevel})
	tempDir := t.TempDir()
	rm, err := storage.NewRecordManager(cfg, tempDir, &log)
	require.NoError(t, err)

	// Use CreateRecordWithConfig
	rec, err := rm.CreateRecordWithConfig(dummyConfigData, dummyOrkData)
	assert.NoError(t, err)
	assert.NotNil(t, rec)

	gotRec, err := rm.GetRecord(rec.Hash)
	assert.NoError(t, err)
	assert.NoError(t, gotRec.Close())
}

// TEST: GIVEN a record manager WHEN we delete a record THEN the record is removed
func TestRecordManagerDeleteRecord(t *testing.T) {
	cfg := &config.Config{}
	log := logf.New(logf.Opts{Level: logf.ErrorLevel})
	tempDir := t.TempDir()
	rm, err := storage.NewRecordManager(cfg, tempDir, &log)
	require.NoError(t, err)

	// Use CreateRecordWithConfig
	rec, err := rm.CreateRecordWithConfig(dummyConfigData, dummyOrkData)
	assert.NoError(t, err)
	assert.NotNil(t, rec)

	err = rm.DeleteRecord(rec.Hash)
	assert.NoError(t, err)

	_, err = rm.GetRecord(rec.Hash)
	assert.Error(t, err)
}
