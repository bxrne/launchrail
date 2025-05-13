package systems_test

import (
	"io"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/atmosphere"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

func TestAerodynamicSystem_CalculateDrag(t *testing.T) {
	testLogger := logf.New(logf.Opts{Writer: io.Discard})
	isa := atmosphere.NewISAModel(&config.ISAConfiguration{
		SeaLevelDensity:      1.225,
		SeaLevelPressure:     101325,
		SeaLevelTemperature:  288.15,
		SpecificGasConstant:  287.058,
		GravitationalAccel:   9.80665,
		RatioSpecificHeats:   1.4,
		TemperatureLapseRate: -0.0065,
	})
	system := systems.NewAerodynamicSystem(&ecs.World{}, isa, testLogger)

	// Create test entity
	basicEntity := ecs.BasicEntity{}
	entity := &states.PhysicsState{
		BasicEntity: basicEntity,
		Position:    &types.Position{Vec: types.Vector3{Y: 0}},   // Sea level
		Velocity:    &types.Velocity{Vec: types.Vector3{Y: 100}}, // 100 m/s upward
		Nosecone:    &components.Nosecone{Radius: 0.1},
		Bodytube:    &components.Bodytube{Radius: 0.1, Length: 1.0},
		// Parachute remains nil for this test (body drag case)
	}

	dragForce := system.CalculateDrag(entity)

	// Verify drag force is negative (opposing motion)
	if dragForce.Y >= 0 {
		t.Errorf("Expected negative drag force, got %f", dragForce.Y)
	}
}

func TestAerodynamicSystem_Update(t *testing.T) {
	world := &ecs.World{}
	testLogger := logf.New(logf.Opts{Writer: io.Discard})
	isa := atmosphere.NewISAModel(&config.ISAConfiguration{
		SeaLevelDensity:      1.225,
		SeaLevelPressure:     101325,
		SeaLevelTemperature:  288.15,
		SpecificGasConstant:  287.058,
		GravitationalAccel:   9.80665,
		RatioSpecificHeats:   1.4,
		TemperatureLapseRate: -0.0065,
	})
	system := systems.NewAerodynamicSystem(world, isa, testLogger)

	// Create a basic ECS entity
	basicEntity := ecs.BasicEntity{}

	// Create PhysicsState and link to basicEntity
	entity1 := &states.PhysicsState{
		BasicEntity:         basicEntity,
		Position:            &types.Position{},
		Velocity:            &types.Velocity{Vec: types.Vector3{Y: 100}},
		Acceleration:        &types.Acceleration{},
		Mass:                &types.Mass{Value: 1.0},
		Nosecone:            &components.Nosecone{Radius: 0.1},
		Bodytube:            &components.Bodytube{Radius: 0.1, Length: 1.0},
		Orientation:         &types.Orientation{},
		AngularVelocity:     &types.Vector3{},
		AngularAcceleration: &types.Vector3{},
		AccumulatedForce:    types.Vector3{},
		AccumulatedMoment:   types.Vector3{},
	}

	system.Add(entity1)

	system.Update(0.01) // 10ms timestep

	// Verify that forces were accumulated
	if entity1.AccumulatedForce.Magnitude() == 0 {
		t.Error("Expected non-zero accumulated force after update")
	}
}
