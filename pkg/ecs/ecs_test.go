package ecs_test

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/entities"
	"github.com/bxrne/launchrail/pkg/ecs/systems"
	"github.com/bxrne/launchrail/pkg/ecs/types"
)

func TestNewWorld(t *testing.T) {
	world := ecs.NewWorld()
	if world == nil {
		t.Fatal("NewWorld() returned nil")
	}

	// Create entity to test initialization
	id := world.CreateEntity()
	if id != 0 {
		t.Errorf("First entity ID = %v, want 0", id)
	}
}

func TestWorld_CreateEntity(t *testing.T) {
	world := ecs.NewWorld()

	// Test sequential ID generation
	ids := make([]entities.EntityID, 3)
	for i := range ids {
		ids[i] = world.CreateEntity()
	}

	for i := 0; i < len(ids)-1; i++ {
		if ids[i+1] <= ids[i] {
			t.Errorf("Entity IDs not increasing: ids[%d]=%d, ids[%d]=%d",
				i, ids[i], i+1, ids[i+1])
		}
	}

	// Test concurrent entity creation
	var wg sync.WaitGroup
	numGoroutines := 100
	createdIDs := make([]entities.EntityID, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			createdIDs[index] = world.CreateEntity()
		}(i)
	}
	wg.Wait()

	// Verify all IDs are unique
	idMap := make(map[entities.EntityID]bool)
	for _, id := range createdIDs {
		if idMap[id] {
			t.Error("Duplicate entity ID created")
		}
		idMap[id] = true
	}
}

func TestWorld_AddComponent(t *testing.T) {
	world := ecs.NewWorld()
	entity := world.CreateEntity()

	tests := []struct {
		name        string
		component   components.Component
		entity      entities.EntityID
		wantErr     bool
		errContains string
	}{
		{
			name:      "add valid component",
			component: components.NewMockComponent("Test"),
			entity:    entity,
			wantErr:   false,
		},
		{
			name:        "add to non-existent entity",
			component:   components.NewMockComponent("Test"),
			entity:      99999,
			wantErr:     true,
			errContains: "does not exist",
		},
		{
			name: "add transform component",
			component: &components.TransformComponent{
				Position: types.Vector3{X: 1, Y: 2, Z: 3},
			},
			entity:  entity,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := world.AddComponent(tt.entity, tt.component)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddComponent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("AddComponent() error = %v, should contain %v", err, tt.errContains)
			}
		})
	}

	// Test concurrent component addition
	var wg sync.WaitGroup
	numGoroutines := 100
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			component := components.NewMockComponent(fmt.Sprintf("Test%d", index))
			_ = world.AddComponent(entity, component)
		}(i)
	}
	wg.Wait()
}

func TestWorld_GetComponent(t *testing.T) {
	world := ecs.NewWorld()
	entity := world.CreateEntity()
	component := components.NewMockComponent("Test")

	// Add component
	err := world.AddComponent(entity, component)
	if err != nil {
		t.Fatalf("Failed to add component: %v", err)
	}

	tests := []struct {
		name          string
		entity        entities.EntityID
		componentType string
		want          components.Component
		wantOk        bool
	}{
		{
			name:          "get existing component",
			entity:        entity,
			componentType: "Test",
			want:          component,
			wantOk:        true,
		},
		{
			name:          "get non-existent component type",
			entity:        entity,
			componentType: "NonExistent",
			want:          nil,
			wantOk:        false,
		},
		{
			name:          "get from non-existent entity",
			entity:        99999,
			componentType: "Test",
			want:          nil,
			wantOk:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := world.GetComponent(tt.entity, tt.componentType)
			if ok != tt.wantOk {
				t.Errorf("GetComponent() ok = %v, want %v", ok, tt.wantOk)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponent() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorld_AddSystem(t *testing.T) {
	world := ecs.NewWorld()
	system := systems.NewMockSystem(1)

	world.AddSystem(system)

	// Update world to verify system was added
	dt := 0.016
	world.Update(dt)

	if !system.GetUpdateCalled() {
		t.Error("System.Update was not called")
	}
	if system.GetUpdateDt() != dt {
		t.Errorf("System.Update called with dt = %v, want %v", system.GetUpdateDt(), dt)
	}
}

func TestWorld_Update(t *testing.T) {
	world := ecs.NewWorld()

	// Create pointers to systems instead of copying them
	mockSystems := []*systems.MockSystem{
		systems.NewMockSystem(1),
		systems.NewMockSystem(2),
		systems.NewMockSystem(3),
	}

	// Add the system pointers to the world
	for _, sys := range mockSystems {
		world.AddSystem(sys)
	}

	// Update world
	dt := 0.016
	world.Update(dt)

	// Verify all systems were updated - now checking the actual system instances
	for i, sys := range mockSystems {
		if !sys.GetUpdateCalled() {
			t.Errorf("System %d was not updated", i)
		}
		if sys.GetUpdateDt() != dt {
			t.Errorf("System %d updated with dt = %v, want %v", i, sys.GetUpdateDt(), dt)
		}
	}

	// Test concurrent updates
	var wg sync.WaitGroup
	numGoroutines := 10
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			world.Update(dt)
		}()
	}
	wg.Wait()
}
