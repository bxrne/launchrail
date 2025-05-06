## pkg/atmosphere

### Responsibility
- Provides functions for calculating atmospheric properties.
- Implements models for atmospheric density, pressure, and temperature.

### Contract
- Defines `Model` interface for atmospheric data:
  - `GetAtmosphere(altitude float64) AtmosphereData`
  - `GetSpeedOfSound(altitude float64) float64`
- Default implementation: `ISAModel`.
- Plugins can override via `config.Setup.Plugins.Paths` by providing a custom `Model`.

### Scope
- Atmospheric property calculation functions.
- Atmospheric model definitions.

### Test Suite Overview
- Tests should cover atmospheric property calculations and ensure that the output is accurate.

### Decisions & Potential Gotchas
- Atmospheric models should be well-documented and validated.
- Calculation accuracy may need to be optimized for performance.
