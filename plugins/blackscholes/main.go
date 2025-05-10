package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/zerodha/logf"
)

// BlackScholesPlugin implements the atmospheric turbulence model
type BlackScholesPlugin struct {
	log                   logf.Logger
	rng                   *rand.Rand     // Random number generator
	turbulenceIntensity   float64        // σ (sigma) - magnitude of random fluctuations
	deterministicForcesMu float64        // μ (mu) - deterministic drift component
	cfg                   *config.Config // config for simulation parameters
}

var Plugin BlackScholesPlugin

// Initialize is called when the plugin is loaded
func (p *BlackScholesPlugin) Initialize(log logf.Logger, cfg *config.Config) error {
	// TODO: Use cfg if needed to load parameters like turbulenceIntensity
	p.log = log
	p.cfg = cfg
	p.log.Info("Initializing Black-Scholes turbulence plugin")

	// Seed RNG: deterministic for tests (nil cfg), otherwise time-based
	if cfg == nil {
		p.rng = rand.New(rand.NewSource(1))
	} else {
		seed := time.Now().UnixNano()
		p.rng = rand.New(rand.NewSource(seed))
	}
	p.log.Debug("Random number generator seeded", "seed" /* omitted */)

	// TODO: Initialize parameters (e.g., load from config file)
	p.turbulenceIntensity = 0.05   // Example initial value for turbulence intensity (adjust based on desired effect)
	p.deterministicForcesMu = 9.81 // Example initial value (gravity) - currently unused here

	return nil
}

// Name returns the unique identifier of the plugin
func (p *BlackScholesPlugin) Name() string {
	return "blackscholes"
}

// Version returns the plugin version
func (p *BlackScholesPlugin) Version() string {
	return "0.1.1" // Incremented version
}

// BeforeSimStep is called before each simulation step
func (p *BlackScholesPlugin) BeforeSimStep(state *states.PhysicsState) error {
	// No action needed before step in this simple model
	return nil
}

// AfterSimStep applies stochastic turbulence effects based on Black-Scholes model
func (p *BlackScholesPlugin) AfterSimStep(state *states.PhysicsState) error {
	// No change if zero intensity
	if p.turbulenceIntensity == 0 {
		return nil
	}
	// Determine timestep
	dt := 1.0
	if p.cfg != nil {
		dt = p.cfg.Engine.Simulation.Step
	}
	// Compute current speed magnitude
	vx, vy, vz := state.Velocity.Vec.X, state.Velocity.Vec.Y, state.Velocity.Vec.Z
	speed := math.Sqrt(vx*vx + vy*vy + vz*vz)
	// Standard deviation for noise
	stdDev := p.turbulenceIntensity * speed * math.Sqrt(dt)
	// Generate noise and apply
	state.Velocity.Vec.X += p.rng.NormFloat64() * stdDev
	state.Velocity.Vec.Y += p.rng.NormFloat64() * stdDev
	state.Velocity.Vec.Z += p.rng.NormFloat64() * stdDev

	return nil
}

// Cleanup is called when the simulation ends
func (p *BlackScholesPlugin) Cleanup() error {
	p.log.Info("Cleaning up Black-Scholes turbulence plugin")
	return nil
}

// main is required for Go plugins.
func main() {}
