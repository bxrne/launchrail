## cmd/server

### Responsibility
- Implements the Launchrail web server.
- Handles HTTP requests and responses.
- Serves static files and templates.
- Provides API endpoints for simulation data and control.

### Scope
- Gin web framework.
- Route handling for static files, templates, and API endpoints.
- Data handling for simulation records.
- Template rendering using `templ`.

### Test Suite Overview
- Lacks a comprehensive test suite.
- Tests should cover route handling, API endpoints, and data handling.

### Decisions & Potential Gotchas
- Uses `gin` for web framework.
- Uses `templ` for template rendering.
- Static files are served from the `static` directory.
- Template files are located in the `templates` directory.
