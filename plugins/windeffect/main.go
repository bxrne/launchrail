package main

import (
	"math"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/zerodha/logf"
)

type WindEffectPlugin struct {
	log       logf.Logger
	windSpeed float64
}

var Plugin WindEffectPlugin

func (p *WindEffectPlugin) Initialize(log logf.Logger, cfg *config.Config) error {
	p.log = log
	p.windSpeed = 5.0 // m/s base wind speed
	return nil
}

func (p *WindEffectPlugin) Name() string {
	return "WindEffect"
}

func (p *WindEffectPlugin) Version() string {
	return "1.0.0"
}

func (p *WindEffectPlugin) BeforeSimStep(state *states.PhysicsState) error {
	// Calculate the intended acceleration component from the "wind"
	// This simulates a wind that varies sinusoidally with time,
	// and its strength is p.windSpeed.
	// If p.windSpeed is 5.0, this means the wind can cause up to 5 m/s^2 of acceleration.
	accelerationX := p.windSpeed * math.Sin(state.Time)

	// Calculate the force to apply: F = m * a
	// state.Mass.Value holds the scalar mass of the rocket.
	forceX := state.Mass.Value * accelerationX

	// Apply this force to the X component of the accumulated forces.
	// The physics engine will then use this accumulated force to update velocity.
	state.AccumulatedForce.X += forceX
	p.log.Debug("WindEffect applied force", "time", state.Time, "wind_accel_x", accelerationX, "mass", state.Mass.Value, "applied_force_x", forceX, "total_accum_force_x", state.AccumulatedForce.X)
	return nil
}

func (p *WindEffectPlugin) AfterSimStep(state *states.PhysicsState) error {
	return nil
}

func (p *WindEffectPlugin) Cleanup() error {
	return nil
}

// main is required for Go plugins.
func main() {}
