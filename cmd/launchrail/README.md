## cmd/launchrail

### Responsibility
- Serves as the main entry point for the Launchrail application.
- Handles command-line arguments and configuration loading.
- Initializes and runs the simulation.
- Manages the output directory for simulation results.

### Scope
- Main function that initializes the simulation.
- Construction of the output directory.
- Management of the simulation run process.

### Test Suite Overview
- Currently lacks a comprehensive test suite.
- Tests should cover argument parsing, configuration loading, and simulation initialization.

### Decisions & Potential Gotchas
- Uses `spf13/viper` for configuration loading.
- Output directory is hardcoded to `~/.launchrail`.
- Error handling for configuration loading and simulation initialization needs to be robust.
