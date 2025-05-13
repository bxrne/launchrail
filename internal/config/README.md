# config

Configuration management system that loads, validates and provides application settings from YAML files.

## Notes
- Defines structured configuration using Go structs
- Utilizes spf13/viper for loading configuration from YAML files
- Implements validation to ensure required fields and proper values
- Provides type-safe access to configuration parameters
- Uses a singleton pattern for consistent configuration access
- Supports different configuration sections: server, simulation, storage, engine
- Includes comprehensive test coverage for loading and validation
- Enables flexible configuration through predefined locations