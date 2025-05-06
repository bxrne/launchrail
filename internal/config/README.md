# internal/config

## Description

The `internal/config` package is responsible for managing the application's configuration. It loads settings from a configuration file, validates them, and provides a structured way for other parts of the application to access these settings.

### Key Components

* **Configuration Structs (`config.go`):**
  * Defines Go structs (e.g., `Config`, `ServerConfig`, `SimulationConfig`, `StorageConfig`, `PluginConfig`, `LogConfig`, etc.) that map to the structure of the configuration file.
  * These structs hold all configurable parameters for different aspects of the LaunchRail application.

* **Loading Mechanism (`config.go`):**
  * Utilizes the `spf13/viper` library to read configuration data from a YAML file (typically named `config.yaml` or specified via environment variables/flags).
  * Provides functions like `Load()` or `NewConfig()` to initialize and populate the configuration structs.
  * Often implements a singleton pattern or a central access point to ensure consistent configuration across the application.

* **Validation (`config.go`):**
  * Includes logic to validate the loaded configuration. This ensures that required fields are present, values are within acceptable ranges or formats, and inter-dependencies between settings are consistent.
  * Validation might occur during the loading process or via an explicit `Validate()` method.

* **Testing (`config_test.go`):**
  * Contains unit tests to verify:
    * Successful loading of valid configuration files.
    * Correct parsing of different configuration values and structures.
    * Proper handling of missing or malformed configuration files.
    * Effectiveness of validation rules (e.g., ensuring errors are reported for invalid settings).

### Core Functionalities

* **Centralized Configuration:** Provides a single source of truth for all application settings.
* **Type-Safe Access:** Allows other packages to access configuration parameters in a type-safe manner through the defined structs.
* **Flexibility:** Supports configuration via files, and potentially environment variables or command-line flags (capabilities often provided by `viper`).
* **Robustness:** Ensures the application starts with a valid and coherent configuration through validation.

### Technical Details

* **Library:** `spf13/viper` for configuration management.
* **Format:** Typically YAML, but `viper` can support other formats.
* **Default Location:** Often expects a `config.yaml` in a predefined location or the working directory, but this can be overridden.
