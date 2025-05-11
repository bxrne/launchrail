package main

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

func TestWindEffectPlugin_Initialize(t *testing.T) {
	p := &WindEffectPlugin{}
	logger := logf.New(logf.Opts{})
	cfg := &config.Config{}
	err := p.Initialize(logger, cfg)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}
}

func TestWindEffectPlugin_Name(t *testing.T) {
	p := &WindEffectPlugin{}
	if p.Name() != "WindEffect" {
		t.Errorf("Expected name to be WindEffect, got %s", p.Name())
	}
}

func TestWindEffectPlugin_Version(t *testing.T) {
	p := &WindEffectPlugin{}
	if p.Version() != "1.0.0" {
		t.Errorf("Expected version to be 1.0.0, got %s", p.Version())
	}
}

func TestWindEffectPlugin_BeforeSimStep(t *testing.T) {
	p := &WindEffectPlugin{}
	logger := logf.New(logf.Opts{})
	cfg := &config.Config{} // Initialize with an empty config or a relevant one if needed by Initialize
	if err := p.Initialize(logger, cfg); err != nil {
		t.Fatalf("Plugin Initialize failed: %v", err)
	}

	initialVelocityX := 10.0
	state := &states.PhysicsState{
		Time:             0, // math.Sin(0) is 0, so force will be 0
		Velocity:         &types.Velocity{Vec: types.Vector3{X: initialVelocityX, Y: 0, Z: 0}},
		Mass:             &types.Mass{Value: 2.0}, // Example mass
		AccumulatedForce: types.Vector3{X: 0, Y: 0, Z: 0},
	}

	err := p.BeforeSimStep(state)
	if err != nil {
		t.Errorf("BeforeSimStep failed: %v", err)
	}

	// With Time = 0, math.Sin(state.Time) = 0, so accelerationX = 0, forceX = 0
	expectedForceX := p.windSpeed * math.Sin(state.Time) * state.Mass.Value
	if math.Abs(state.AccumulatedForce.X-expectedForceX) > 0.0001 {
		t.Errorf("Expected AccumulatedForce.X to be %f, got %f", expectedForceX, state.AccumulatedForce.X)
	}

	// Velocity should not be directly changed by BeforeSimStep
	if math.Abs(state.Velocity.Vec.X-initialVelocityX) > 0.0001 {
		t.Errorf("Expected Velocity.X to be unchanged (%f), got %f", initialVelocityX, state.Velocity.Vec.X)
	}

	// Test with Time that gives non-zero force
	state.Time = math.Pi / 2 // math.Sin(Pi/2) is 1
	state.AccumulatedForce.X = 0 // Reset accumulated force for new test case
	err = p.BeforeSimStep(state)
	if err != nil {
		t.Errorf("BeforeSimStep failed for Time=Pi/2: %v", err)
	}
	expectedForceXPiOver2 := p.windSpeed * math.Sin(state.Time) * state.Mass.Value // windSpeed * 1 * mass
	if math.Abs(state.AccumulatedForce.X-expectedForceXPiOver2) > 0.0001 {
		t.Errorf("Expected AccumulatedForce.X to be %f for Time=Pi/2, got %f", expectedForceXPiOver2, state.AccumulatedForce.X)
	}
}

func TestWindEffectPlugin_AfterSimStep(t *testing.T) {
	p := &WindEffectPlugin{}
	state := &states.PhysicsState{}

	err := p.AfterSimStep(state)
	if err != nil {
		t.Errorf("AfterSimStep failed: %v", err)
	}
}
