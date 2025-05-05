package main

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

func TestWindEffectPlugin_Initialize(t *testing.T) {
	p := &WindEffectPlugin{}
	logger := logf.New(logf.Opts{})

	err := p.Initialize(logger)
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
	state := &states.PhysicsState{
		Time:     0, // This will make sin(time) = 1
		Velocity: &types.Velocity{Vec: types.Vector3{X: 10.0}},
	}

	err := p.BeforeSimStep(state)
	if err != nil {
		t.Errorf("BeforeSimStep failed: %v", err)
	}

	expectedVelocity := 10.0 // Original velocity + (windSpeed * sin(time) * 0.1)
	if math.Abs(state.Velocity.Vec.X-expectedVelocity) > 0.0001 {
		t.Errorf("Expected velocity to be %f, got %f", expectedVelocity, state.Velocity.Vec.X)
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
