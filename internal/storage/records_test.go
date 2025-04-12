package storage_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/storage"
	"github.com/stretchr/testify/require"
)

// TEST: GIVEN a record manager WHEN we create and close a record THEN no error is returned
func TestNewRecordAndClose(t *testing.T) {
	rm, err := storage.NewRecordManager(t.TempDir())
	require.NoError(t, err)

	rec, err := rm.CreateRecord()
	require.NoError(t, err)
	require.NoError(t, rec.Close())
}

// TEST: GIVEN a record manager WHEN multiple records exist THEN they are listed
func TestRecordManagerListRecords(t *testing.T) {
	rm, err := storage.NewRecordManager(t.TempDir())
	require.NoError(t, err)

	_, err = rm.CreateRecord()
	require.NoError(t, err)

	records, err := rm.ListRecords()
	require.NoError(t, err)
	require.NotEmpty(t, records)
}

// TEST: GIVEN a record manager WHEN we retrieve a record by hash THEN the record is returned
func TestRecordManagerGetRecord(t *testing.T) {
	rm, err := storage.NewRecordManager(t.TempDir())
	require.NoError(t, err)

	rec, err := rm.CreateRecord()
	require.NoError(t, err)

	gotRec, err := rm.GetRecord(rec.Hash)
	require.NoError(t, err)
	require.NoError(t, gotRec.Close())
}
