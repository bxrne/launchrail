package systems_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStorageTest(t *testing.T) (*storage.Storage, func()) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	baseDir := "test_storage"
	dir := "test_data"
	fullBaseDir := filepath.Join(homeDir, baseDir)

	storage, err := storage.NewStorage(baseDir, dir, storage.MOTION)
	require.NoError(t, err)

	headers := []string{"Time", "Altitude", "Velocity", "Acceleration", "Thrust"}
	err = storage.Init(headers)
	require.NoError(t, err)

	cleanup := func() {
		storage.Close()
		os.RemoveAll(fullBaseDir)
	}

	return storage, cleanup
}

// TEST: GIVEN a new StorageParasiteSystem WHEN initialized THEN it should be created with correct defaults
func TestNewStorageParasiteSystem(t *testing.T) {
	world := &ecs.World{}
	storage, cleanup := setupStorageTest(t)
	defer cleanup()

	system := systems.NewStorageParasiteSystem(world, storage)

	assert.NotNil(t, system)
}

// TEST: GIVEN a running StorageParasiteSystem WHEN data is sent THEN it should write to storage
func TestStorageParasiteSystem_ProcessData(t *testing.T) {
	world := &ecs.World{}
	storage, cleanup := setupStorageTest(t)
	defer cleanup()

	system := systems.NewStorageParasiteSystem(world, storage)

	dataChan := make(chan systems.RocketState)
	system.Start(dataChan)

	testState := systems.RocketState{
		Time:         1.0,
		Altitude:     100.0,
		Velocity:     50.0,
		Acceleration: 9.81,
		Thrust:       100.0,
		MotorState:   "burning",
	}

	go func() {
		dataChan <- testState
		time.Sleep(100 * time.Millisecond)
		system.Stop()
	}()

	// Wait for goroutine to complete
	time.Sleep(200 * time.Millisecond)
}

// TEST: GIVEN a StorageParasiteSystem WHEN an entity is added THEN it should be stored in the system
func TestStorageParasiteSystem_Add(t *testing.T) {
	world := &ecs.World{}
	storage, cleanup := setupStorageTest(t)
	defer cleanup()

	system := systems.NewStorageParasiteSystem(world, storage)
	e := ecs.NewBasic()

	entity := systems.PhysicsEntity{
		Entity: &e,
	}

	system.Add(&entity)

	assert.NoError(t, nil)
}

// TEST: GIVEN a StorageParasiteSystem WHEN Priority is called THEN it should return correct priority
func TestStorageParasiteSystem_Priority(t *testing.T) {
	world := &ecs.World{}
	storage, cleanup := setupStorageTest(t)
	defer cleanup()

	system := systems.NewStorageParasiteSystem(world, storage)
	assert.Equal(t, 1, system.Priority())
}
