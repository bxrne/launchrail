## internal/logger

### Responsibility
- Provides a centralized logging facility for the application.
- Uses `zerodha/logf` for structured logging.
- Configures logging level and output format.

### Scope
- Logger initialization and configuration.
- Logging functions for different levels (e.g., Info, Debug, Error).

### Test Suite Overview
- Tests should cover logging output and level configuration.

### Decisions & Potential Gotchas
- Uses `zerodha/logf` for structured logging.
- Logging level is configurable via the configuration file.
- Output format can be customized.
