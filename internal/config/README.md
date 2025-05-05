## internal/config

### Responsibility
- Loads and manages the application configuration.
- Provides a singleton instance of the configuration.
- Validates the configuration for missing or invalid fields.

### Scope
- Configuration struct definition.
- Loading configuration from YAML file using `spf13/viper`.
- Validation of configuration fields.

### Test Suite Overview
- Tests should cover configuration loading and validation.

### Decisions & Potential Gotchas
- Uses `spf13/viper` for configuration loading.
- Configuration file is located at `config.yaml`.
- Validation logic ensures that required fields are present and valid.
