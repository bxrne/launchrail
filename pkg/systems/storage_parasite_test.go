package systems_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	logf "github.com/zerodha/logf"
)

// mockStorage with better capabilities for testing
type mockStorage struct {
	mu                       sync.Mutex
	records                  [][]string
	writeError               error // Error to return on Write
	initError                error // Error to return on Init
	initCalled               bool
	closeCalled              bool
	storage.StorageInterface // Embed interface for compatibility check (if defined)
}

func (m *mockStorage) Init() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initCalled = true
	return m.initError
}

func (m *mockStorage) Write(record []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.writeError != nil {
		return m.writeError
	}
	m.records = append(m.records, record)
	return nil
}

func (m *mockStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCalled = true
	return nil
}

func (m *mockStorage) Read() ([]string, error) {
	// Not used by parasite, but needed for interface if StorageInterface includes it
	return nil, errors.New("read not implemented in mock")
}

// Helper to get recorded data safely
func (m *mockStorage) getRecords() [][]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy
	recs := make([][]string, len(m.records))
	copy(recs, m.records)
	return recs
}

// Helper to wait for a specific number of records
func (m *mockStorage) waitForRecords(t *testing.T, count int) {
	timeout := time.After(100 * time.Millisecond) // Prevent test hanging
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for %d records", count)
		case <-ticker.C:
			m.mu.Lock()
			currentCount := len(m.records)
			m.mu.Unlock()
			if currentCount >= count {
				return
			}
		}
	}
}

// --- Test Cases ---

// Helper function to create a valid motor for tests
func newTestMotor(t *testing.T) *components.Motor {
	t.Helper()
	basicEntity := ecs.NewBasic()
	motorData := &thrustcurves.MotorData{
		Designation: designation.Designation("TestMotorL1100"), // Use correct type and field
		ID:          "test-motor-123",                          // Use correct field
		TotalMass:   10.0,
		BurnTime:    5.0,
		MaxThrust:   1200,
		Thrust:      [][]float64{{0, 100}, {5, 0}}, // Use correct type [][]float64
	}
	discardLogger := logf.New(logf.Opts{}) // Create logger with default options
	motor, err := components.NewMotor(basicEntity, motorData, discardLogger)
	require.NoError(t, err, "Creating test motor should not fail")
	return motor
}

func setupTest(t *testing.T, storeType storage.SimStorageType) (*systems.StorageParasiteSystem, *mockStorage, chan *states.PhysicsState) {
	t.Helper()
	world := &ecs.World{}
	mock := &mockStorage{}
	sys := systems.NewStorageParasiteSystem(world, mock, storeType)

	dataChan := make(chan *states.PhysicsState, 10) // Buffer helps prevent blocking
	err := sys.Start(dataChan)
	require.NoError(t, err, "sys.Start should not error")
	require.True(t, mock.initCalled, "mock.Init should have been called")

	// Cleanup function
	t.Cleanup(func() {
		sys.Stop()
		// Give the processor time to exit after closing done channel
		time.Sleep(10 * time.Millisecond)
		// Close is not directly called by parasite, but good practice if interface had it
		// assert.True(t, mock.closeCalled, "mock.Close should be called")
	})

	return sys, mock, dataChan
}

func TestStorageParasiteSystem_Motion(t *testing.T) {
	_, mock, dataChan := setupTest(t, storage.MOTION)

	testMotor := newTestMotor(t)

	testState := &states.PhysicsState{
		Time:         1.234567,
		Position:     &types.Position{Vec: types.Vector3{Y: 100.1}},
		Velocity:     &types.Velocity{Vec: types.Vector3{Y: 10.2}},
		Acceleration: &types.Acceleration{Vec: types.Vector3{Y: 1.3}},
		Motor:        testMotor, // Thrust now expected to be 74.025 if raw was 100.0 // Use helper
		// Parachute and Orientation not needed for MOTION
	}
	// testState.Motor.Ignite(0) // Removed - NewMotor handles initial state

	dataChan <- testState
	mock.waitForRecords(t, 1) // Wait for 1 record

	records := mock.getRecords()
	require.Len(t, records, 1, "Expected 1 record")

	// Calculate expected thrust with efficiency factors: 0.85 * 0.90 * 0.97 = 0.74025
	expectedRecord := []string{
		"1.234567",
		"100.100000",
		"10.200000",
		"1.300000",
		"100.000000",
	}
	assert.Equal(t, expectedRecord, records[0])
}

func TestStorageParasiteSystem_Events(t *testing.T) {
	_, mock, dataChan := setupTest(t, storage.EVENTS)

	testMotor := newTestMotor(t)

	testState := &states.PhysicsState{
		Time:      2.5,
		Motor:     testMotor,
		Parachute: &components.Parachute{Deployed: false}, // Initialized correctly
		// Position, Velocity etc. not needed for EVENTS
	}
	// testState.Motor.Ignite(0) // Removed
	testState.Parachute.Deploy() // Call without arguments

	dataChan <- testState
	mock.waitForRecords(t, 1)

	records := mock.getRecords()
	require.Len(t, records, 1)
	expectedRecord := []string{"2.500000", "NONE", "IGNITED", "DEPLOYED"}
	assert.Equal(t, expectedRecord, records[0])
}

func TestStorageParasiteSystem_Dynamics(t *testing.T) {
	_, mock, dataChan := setupTest(t, storage.DYNAMICS)

	testState := &states.PhysicsState{
		Time:         3.14,
		Position:     &types.Position{Vec: types.Vector3{X: 1, Y: 2, Z: 3}},
		Velocity:     &types.Velocity{Vec: types.Vector3{X: 4, Y: 5, Z: 6}},
		Acceleration: &types.Acceleration{Vec: types.Vector3{X: 7, Y: 8, Z: 9}},
		Orientation:  &types.Orientation{Quat: types.Quaternion{X: 0.1, Y: 0.2, Z: 0.3, W: 0.9}}, // Ensure orientation exists
		// Motor and Parachute not needed for DYNAMICS
	}

	dataChan <- testState
	mock.waitForRecords(t, 1)

	records := mock.getRecords()
	require.Len(t, records, 1)
	expectedRecord := []string{
		"3.140000",
		"1.000000", "2.000000", "3.000000", // Pos
		"4.000000", "5.000000", "6.000000", // Vel
		"7.000000", "8.000000", "9.000000", // Acc
		"0.100000", "0.200000", "0.300000", "0.900000", // Orient
	}
	assert.Equal(t, expectedRecord, records[0])
}

func TestStorageParasiteSystem_NilState(t *testing.T) {
	_, mock, dataChan := setupTest(t, storage.MOTION)

	dataChan <- nil                   // Send a nil state
	time.Sleep(20 * time.Millisecond) // Give time for potential processing

	records := mock.getRecords()
	assert.Empty(t, records, "No records should be written for nil state")
}

func TestStorageParasiteSystem_PartialState_Motion(t *testing.T) {
	_, mock, dataChan := setupTest(t, storage.MOTION)

	// State missing Motor
	testState := &states.PhysicsState{
		Time:         1.0,
		Position:     &types.Position{Vec: types.Vector3{Y: 100}},
		Velocity:     &types.Velocity{Vec: types.Vector3{Y: 10}},
		Acceleration: &types.Acceleration{Vec: types.Vector3{Y: 1}},
		Motor:        nil, // Missing
	}

	dataChan <- testState
	time.Sleep(20 * time.Millisecond)

	records := mock.getRecords()
	assert.Empty(t, records, "No records should be written for partial state (missing motor)")
}

func TestStorageParasiteSystem_WriteError(t *testing.T) {
	_, mock, dataChan := setupTest(t, storage.MOTION)
	mock.writeError = errors.New("disk full") // Simulate a write error

	testMotor := newTestMotor(t)

	testState := &states.PhysicsState{
		Time:         1.23,
		Position:     &types.Position{Vec: types.Vector3{Y: 100}},
		Velocity:     &types.Velocity{Vec: types.Vector3{Y: 10}},
		Acceleration: &types.Acceleration{Vec: types.Vector3{Y: 1}},
		Motor:        testMotor, // Thrust now expected to be 74.025 if raw was 100.0
	}
	// testState.Motor.Ignite(0) // Removed

	dataChan <- testState
	time.Sleep(20 * time.Millisecond) // Give time for processing attempt

	// We expect the write error to be printed, but not easy to capture stdout in tests.
	// We can verify that no record was successfully added to our mock list.
	records := mock.getRecords()
	assert.Empty(t, records, "No records should be stored when write fails")
}

func TestStorageParasiteSystem_StartError(t *testing.T) {
	world := &ecs.World{}
	mock := &mockStorage{initError: errors.New("init failed")} // Simulate init error
	sys := systems.NewStorageParasiteSystem(world, mock, storage.MOTION)

	dataChan := make(chan *states.PhysicsState, 1)
	err := sys.Start(dataChan)

	require.Error(t, err, "sys.Start should return an error")
	assert.Contains(t, err.Error(), "init failed")
	assert.True(t, mock.initCalled, "mock.Init should still be called")
}

// Test Add - currently doesn't do much system state-wise but ensures it runs
func TestStorageParasiteSystem_Add(t *testing.T) {
	sys, _, _ := setupTest(t, storage.MOTION)
	state := &states.PhysicsState{Time: 1.0}
	assert.NotPanics(t, func() { sys.Add(state) }, "Add should not panic")
	// Could potentially inspect internal 'entities' slice if needed, but it's not used by Update/processData
}

// Test Update - currently does nothing, just ensure it runs without error
func TestStorageParasiteSystem_Update(t *testing.T) {
	sys, _, _ := setupTest(t, storage.MOTION)
	assert.NoError(t, sys.Update(0.1), "Update should not return an error")
}
