# plugins/blackscholes

Provides a stochastic turbulence model inspired by the Black-Scholes (GBM) process and can override the default ISA atmosphere.

## Responsibility

- Implements a custom `Model` for atmospheric perturbations.
- Registers as a plugin under `config.Setup.Plugins.Paths` to replace default `pkg/atmosphere.ISAModel`.
- Applies Gaussian noise scaled by velocity and time step to simulate turbulence.

## Contract

- Exports `Plugin BlackScholesPlugin` satisfying `pluginapi.SimulationPlugin`:
  - `Initialize(logf.Logger, *config.Config) error`
  - `Name() string`  → returns "blackscholes".
  - `Version() string` → plugin version.
  - `BeforeSimStep(*states.PhysicsState) error` (noop).
  - `AfterSimStep(*states.PhysicsState) error` → applies turbulence.
  - `Cleanup() error` (cleanup hook).
- Implements additive noise:
  ```go
  if σ==0 { return }
  stdDev = σ * |v| * √Δt
  vᵢ += N(0,1) * stdDev
  ```
  where σ = `turbulenceIntensity` from config, Δt = `cfg.Engine.Simulation.Step`.

## Test Suite Overview

- `main_test.go` covers:
  - Default initialization values.
  - `AfterSimStep` preserves velocity when σ=0.
  - Noise scaling increases with higher σ.
  - Deterministic RNG seed for repeatable tests when `cfg==nil`.

## Decisions & Gotchas

- **Override mechanism**: Plugins override ISA when loaded before systems instantiate; ensure `Simulation.NewSimulation` injects plugin atmosphere.
- **Gaussian noise**: Additive noise (not multiplicative GBM) ensures zero-intensity behavior.
- **Deterministic tests**: Uses fixed RNG seed if config is nil.
- **Performance**: Noise calc is lightweight; can tweak σ via config.

---
*For details on registering plugins and overriding models, see `internal/plugin/manager.go` and `pkg/atmosphere.Model`.*