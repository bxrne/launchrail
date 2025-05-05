## internal/http_client

### Responsibility
- Provides a wrapper around the standard `net/http` client.
- Handles HTTP requests and responses with custom configuration.
- Implements retry logic and error handling.

### Scope
- HTTP client creation and configuration.
- Request execution and response handling.
- Retry logic for failed requests.

### Test Suite Overview
- Tests should cover request execution, response handling, and retry logic.

### Decisions & Potential Gotchas
- Uses the standard `net/http` client.
- Retry logic is configurable.
- Error handling should be robust.
