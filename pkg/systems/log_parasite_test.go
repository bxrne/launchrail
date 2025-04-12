package systems_test

import (
	"testing"
	"time"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

func createTestMotor() (*components.Motor, *thrustcurves.MotorData) {
	// Create simple thrust curve data
	thrustData := [][]float64{
		{0.0, 10.0}, // Initial thrust
		{1.0, 10.0}, // Constant thrust for 1 second
		{2.0, 0.0},  // Burnout
	}

	motorData := &thrustcurves.MotorData{
		Thrust:    thrustData,
		TotalMass: 10.0, // Initial mass
		BurnTime:  2.0,  // 2 second burn
		AvgThrust: 10.0, // Average thrust
	}

	logger := logf.New(logf.Opts{})
	motor, err := components.NewMotor(ecs.NewBasic(), motorData, logger)

	if err != nil {
		panic(err)
	}

	return motor, motorData
}

func TestLogParasiteSystem(t *testing.T) {
	world := &ecs.World{}
	logger := logf.New(logf.Opts{})
	system := systems.NewLogParasiteSystem(world, &logger)

	// Test initialization
	if system == nil {
		t.Fatal("Failed to create LogParasiteSystem")
	}

	// Test data processing
	dataChan := make(chan *states.PhysicsState)
	system.Start(dataChan)
	defer system.Stop()

	// Create mock components
	motor, _ := createTestMotor()

	parachute := components.NewParachute(ecs.NewBasic(), 10.0, 5.0, 4, components.ParachuteTriggerApogee)

	// Send test data with all required components
	testState := &states.PhysicsState{
		Time:         1.0,
		Position:     &types.Position{Vec: types.Vector3{X: 0, Y: 10, Z: 0}},
		Velocity:     &types.Velocity{Vec: types.Vector3{X: 0, Y: 20, Z: 0}},
		Acceleration: &types.Acceleration{Vec: types.Vector3{X: 0, Y: -9.81, Z: 0}},
		Orientation:  &types.Orientation{Quat: types.Quaternion{X: 0, Y: 0, Z: 0, W: 1}},
		Motor:        motor,
		Parachute:    parachute,
	}

	// Send data with timeout to avoid blocking
	select {
	case dataChan <- testState:
	case <-time.After(time.Second):
		t.Error("Timeout sending test data")
	}

	// Test adding entities
	system.Add(testState)
}
