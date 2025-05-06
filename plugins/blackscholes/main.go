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
	rng                   *rand.Rand // Random number generator
	turbulenceIntensity   float64    // σ (sigma) - magnitude of random fluctuations
	deterministicForcesMu float64    // μ (mu) - deterministic drift component
	cfg                   *config.Config // config for simulation parameters
}

var Plugin BlackScholesPlugin

// Initialize is called when the plugin is loaded
func (p *BlackScholesPlugin) Initialize(log logf.Logger, cfg *config.Config) error {
	// TODO: Use cfg if needed to load parameters like turbulenceIntensity
	p.log = log
	p.cfg = cfg
	p.log.Info("Initializing Black-Scholes turbulence plugin")

	// Seed the random number generator
	seed := time.Now().UnixNano()
	p.rng = rand.New(rand.NewSource(seed))
	p.log.Debug("Random number generator seeded", "seed", seed)

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
	// Retrieve simulation timestep
	dt := p.cfg.Engine.Simulation.Step

	// Current velocity components
	vx := state.Velocity.Vec.X
	vy := state.Velocity.Vec.Y
	vz := state.Velocity.Vec.Z

	// Parameters
	sigma := p.turbulenceIntensity
	mu := p.deterministicForcesMu

	// Compute increments via Geometric Brownian Motion
	inc := func(v float64) float64 {
		return (mu-0.5*sigma*sigma)*dt + sigma*math.Sqrt(dt)*p.rng.NormFloat64()
	}

	// Update velocities (override additive noise)
	state.Velocity.Vec.X = vx * math.Exp(inc(vx))
	state.Velocity.Vec.Y = vy * math.Exp(inc(vy))
	state.Velocity.Vec.Z = vz * math.Exp(inc(vz))

	return nil
}

// Cleanup is called when the simulation ends
func (p *BlackScholesPlugin) Cleanup() error {
	p.log.Info("Cleaning up Black-Scholes turbulence plugin")
	return nil
}
