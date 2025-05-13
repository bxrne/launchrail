# server

Backend web server for the LaunchRail application that handles HTTP requests, serves API endpoints for accessing simulation data, and manages simulation control functionalities.

## Notes
- Initializes and configures the Gin HTTP server based on configuration files
- Defines API routes and associates them with their respective handlers
- Manages server dependencies (config, data access layers)
- Implements record management through the storage package
- Provides simulation execution via HTTP API endpoints
- Serves web UI through templ templates
- Handles static assets for the frontend interface
- Exposes RESTful API for clients to interact with simulation data
- Offers visualization endpoints for plotting simulation data
- Includes pagination support for record listing endpoints
- Implements structured logging with zerolog
- Contains robust error handling for API responses