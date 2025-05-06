# cmd/launchrail

## Description

The `cmd/launchrail` directory contains the main entry point for the LaunchRail application Its primary responsibility is to initialize and execute rocket flight simulations based on user-provided configurations and command-line arguments

### Key Features

* **Main Application Entry Point** The `main.go` file houses the `main` function, which orchestrates the application's startup, simulation execution, and shutdown
* **Command-Line Interface (CLI) Handling** It parses command-line arguments to control simulation parameters and application behavior
* **Configuration Management** Loads simulation and application settings, potentially using libraries like `spf13/viper`, from configuration files
* **Simulation Initialization & Execution** Sets up the simulation environment, initializes the simulation manager (likely from the `internal/simulation` package), and triggers the simulation run
* **Output Management** Manages the creation and use of an output directory (e.g., `~/.launchrail`) where simulation results, logs, and other artifacts are stored
* **Plugin Compilation** As per memory `15c1ebe0-b84a-4000-ad18-e76a2a9f3b57`, the simulation initialization process, likely triggered from here, includes the automatic compilation of Go plugins

### Workflow

1 The application is started via the `launchrail` executable
2 Command-line arguments are parsed
3 Application and simulation configurations are loaded
4 The logging system is initialized
5 The simulation manager is initialized, which includes compiling any Go plugins
6 The simulation is executed
7 Results are written to the designated output directory
8 The application exits
