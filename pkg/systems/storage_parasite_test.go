package systems_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/types"
)

type mockStorage struct {
	records         [][]string
	storage.Storage // Embed the interface to ensure compliance
}

func (m *mockStorage) Read() ([]string, error) {
	// Mock implementation for Read method
	if len(m.records) > 0 {
		record := m.records[0]
		m.records = m.records[1:] // Simulate reading by removing the first record
		return record, nil
	}
	return nil, nil

}

func (m *mockStorage) Init() error {
	return nil
}

func (m *mockStorage) Write(record []string) error {
	m.records = append(m.records, record)
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestStorageParasiteSystem(t *testing.T) {
	world := &ecs.World{}
	mock := &mockStorage{}
	sys := systems.NewStorageParasiteSystem(world, mock, storage.MOTION)

	dataChan := make(chan *states.PhysicsState, 1)
	if err := sys.Start(dataChan); err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	defer sys.Stop()

	testState := &states.PhysicsState{
		Time:     1.23,
		Position: &types.Position{Vec: types.Vector3{Y: 100}},
		Velocity: &types.Velocity{Vec: types.Vector3{Y: 10}},
	}

	dataChan <- testState
}
