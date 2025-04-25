package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/bxrne/launchrail/pkg/plugin"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/zerodha/logf"
)

// BlackScholesPlugin implements the atmospheric turbulence model
type BlackScholesPlugin struct {
	log                   logf.Logger
	rng                   *rand.Rand // Random number generator
	turbulenceIntensity   float64    // σ (sigma) - Represents magnitude of random fluctuations
	deterministicForcesMu float64    // μ (mu) - Represents deterministic forces (e.g., drag, gravity). Currently unused in this simplified model.
	// TODO: Potentially add configuration for which state variable (e.g., Velocity.X, Velocity.Y) is affected
}

// Initialize is called when the plugin is loaded
func (p *BlackScholesPlugin) Initialize(log logf.Logger) error {
	p.log = log
	p.log.Info("Initializing Black-Scholes turbulence plugin")

	// Seed the random number generator
	seed := time.Now().UnixNano()
	p.rng = rand.New(rand.NewSource(seed))
	p.log.Debug("Random number generator seeded", "seed", seed)

	// TODO: Initialize parameters (e.g., load from config file)
	p.turbulenceIntensity = 0.05 // Example initial value for turbulence intensity (adjust based on desired effect)
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

// AfterSimStep applies stochastic turbulence effects based on Black-Scholes inspiration
func (p *BlackScholesPlugin) AfterSimStep(state *states.PhysicsState) error {
	// TODO: Get the actual simulation time step (dt) if available. Assuming 1.0 for now.
	dt := 1.0 // Placeholder for simulation time step

	// Generate a random fluctuation based on a normal distribution (Gaussian noise)
	// The standard deviation is proportional to the turbulence intensity (σ)
	// and scaled by sqrt(dt) which is characteristic of Wiener processes (Brownian motion)
	// often associated with financial models like Black-Scholes.
	// We also make it proportional to the current velocity magnitude as turbulence often scales with speed.

	// Calculate current speed (magnitude of velocity vector)
	// Assumes Vector3 has Magnitude() method, using Euclidean distance as fallback
	speed := math.Sqrt(state.Velocity.Vec.X*state.Velocity.Vec.X + state.Velocity.Vec.Y*state.Velocity.Vec.Y + state.Velocity.Vec.Z*state.Velocity.Vec.Z)

	// Calculate standard deviation for the noise term
	// Adjust the scaling factor (e.g., 1.0) as needed for desired simulation behavior
	stdDev := p.turbulenceIntensity * speed * 1.0 * math.Sqrt(dt)

	// Generate random noise components for each velocity axis
	noiseX := p.rng.NormFloat64() * stdDev
	noiseY := p.rng.NormFloat64() * stdDev
	noiseZ := p.rng.NormFloat64() * stdDev

	// Apply the noise to the velocity state
	// This simulates the random buffeting effect of turbulence
	state.Velocity.Vec.X += noiseX
	state.Velocity.Vec.Y += noiseY
	state.Velocity.Vec.Z += noiseZ

	// Optional: Log the applied effect for debugging
	// p.log.Debug("Applied Black-Scholes inspired turbulence", "noiseX", noiseX, "noiseY", noiseY, "noiseZ", noiseZ, "stateVel", state.Velocity)

	// Note: The deterministicForcesMu (μ) is not directly used here.
	// The base simulation loop should handle deterministic forces like gravity/drag.
	// This plugin *adds* the stochastic turbulence component on top of that.
	// A more complex model might have μ influence σ or the drift of the stochastic process.

	return nil
}

// Cleanup is called when the simulation ends
func (p *BlackScholesPlugin) Cleanup() error {
	p.log.Info("Cleaning up Black-Scholes turbulence plugin")
	return nil
}

// Export the plugin instance
// The symbol name 'Plugin' is conventional for go plugin loading
var Plugin plugin.SimulationPlugin = &BlackScholesPlugin{}

func main() {
	// This main function is required for the Go plugin build process,
	// but it won't be executed when loaded as a plugin.
}
