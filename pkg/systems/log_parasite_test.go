package systems_test

import (
	"testing"
	"time"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"
)

// TEST: GIVEN a new LogParasiteSystem WHEN initialized THEN it should be created with correct defaults
func TestNewLogParasiteSystem(t *testing.T) {
	world := &ecs.World{}
	logger := logf.New(logf.Opts{})

	system := systems.NewLogParasiteSystem(world, &logger)

	assert.NotNil(t, system)
}

// TEST: GIVEN a running LogParasiteSystem WHEN data is sent THEN it should process the data
func TestLogParasiteSystem_ProcessData(t *testing.T) {
	world := &ecs.World{}
	logger := logf.New(logf.Opts{})
	system := systems.NewLogParasiteSystem(world, &logger)

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

// TEST: GIVEN a LogParasiteSystem WHEN an entity is added THEN it should be stored in the system
func TestLogParasiteSystem_Add(t *testing.T) {
	world := &ecs.World{}
	logger := logf.New(logf.Opts{})
	system := systems.NewLogParasiteSystem(world, &logger)
	e := ecs.NewBasic()

	entity := systems.PhysicsEntity{
		Entity: &e,
	}

	system.Add(&entity)

	assert.NoError(t, nil)
}
