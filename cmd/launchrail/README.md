# launchrail

Primary command-line interface for Launchrail that serves as the entry point for running simulations.

## Notes
- Loads configuration from files through the config package
- Initializes dependencies: storage, simulation manager, logger
- Wires together components to run simulations
- Creates run-specific directories using a hash of configuration and OpenRocket file
- Sets up multiple storage components (motion, events, dynamics)
- Doesn't use CLI flags - configuration is read from config files
- Errors during initialization abort the process with non-zero exit
- Creates structured file storage for simulation results
- Outputs SHA256 hash on successful simulation completion