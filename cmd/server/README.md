# cmd/server

## Description

The `cmd/server` directory implements the backend web server for the LaunchRail application. It is responsible for handling HTTP requests, serving API endpoints for accessing simulation data, and potentially managing simulation control functionalities. The server is built using the Gin web framework.

### Key Components

* **Main Server (`main.go`):**
  * Initializes and configures the Gin HTTP server.
  * Defines API routes and associates them with their respective handlers.
  * Manages server dependencies, such as application configuration (`config.Config`) and data access layers (e.g., `DataHandler` for interacting with `internal/storage`).
  * May handle serving static assets or UI templates (the previous README mentioned `templ` for templates and `static` and `templates` directories, though these are not directly within this package's file listing, their serving would be configured here).

* **Request Handlers (`handlers.go`):**
  * Contains the logic for processing incoming HTTP requests for various API endpoints.
  * Includes handlers like `ListRecordsAPI` for retrieving lists of simulation records, with support for pagination using `limit` and `offset` query parameters. Responses for record listings include a `total` count of available records (as per memory `92749e42-4c6b-4a4e-96c0-43f1cb9fd624`).
  * Interacts with a `DataHandler` to fetch or modify data.

* **Handler Tests (`handlers_test.go`):**
  * Provides unit tests for the request handlers to ensure their correctness (as detailed in memory `0aa1abbe-9f85-4b0c-a7a5-3c5529a0fe3e`).
  * Uses `net/http/httptest` and `testify` for creating test servers and making assertions.
  * Includes setup functions like `setupTestServer` to initialize a test Gin engine and `DataHandler` with temporary storage for isolated testing.

### Core Functionalities

* **API Endpoints:** Exposes a RESTful API for clients to interact with simulation data.
  * Example: Listing simulation records with pagination.
* **Data Management:** Facilitates access to simulation records stored by the system.
* **Web Framework:** Utilizes the Gin framework for robust and efficient request routing and handling.

### Technical Details

* **Framework:** [Gin Web Framework](https://gin-gonic.com/)
* **Testing:** `net/http/httptest`, `github.com/stretchr/testify`
* **Data Interaction:** Through a `DataHandler` interface, likely interacting with components from the `internal/storage` package.
