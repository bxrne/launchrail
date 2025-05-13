## v0.8.0 (2025-05-13)

### Feat

- implement efficiency factors in motor thrust calculations and add CSV output
- add CSV output files and fix motor mass calculation formatting
- improve aerodynamic modeling with realistic ISA atmosphere and enhanced drag calculations
- enhance aerodynamic model with angle of attack and transonic drag calculations
- implement multi-layer ISA model and improve parachute drag effects
- add parachute status debug logging in aerodynamics calculation
- implement rocket simulation with event tracking and data output
- add burnout and parachute events, rename component methods for clarity
- enhance record listing with 16/64 hex support and improved logging
- add config parameter to storage initialization for consistent logging
- add nanosecond precision to simulation hash and improve record deletion logging
- update OpenRocket XML parsing to support full document structure
- add parachute drag calculation and improve apogee detection logic
- implement rocket center of mass and inertia tensor calculation
- refactor motor FSM to handle coasting state and XML schema updates
- implement inertia tensor calculation for trapezoidal finset component
- add center of mass calculation for trapezoid finset component
- implement gravity and thrust forces in RK4 integration with improved mass tracking
- add timestamps to log filenames for better organization
- standardize log file paths to ~/.launchrail/logs directory across all commands
- **server, logger, entities, simulation**: implement file logging with ANSI color stripping and Gin middleware
- **rocket, simulation, matrix**: enhance mass calculation logging in rocket entity, refactor matrix creation to accept slice input, and improve matrix-related tests for robustness
- **tests, simulation**: add new tests for rocket entity creation, simulation run conditions, and matrix operations; introduce SonarLint configuration for improved code quality

### Fix

- remove thrust scaling factor to use actual motor values in physics simulation
- format code and remove trailing whitespace across multiple files
- **logger**: improved coverage

### Refactor

- improve physics system code organization and add comprehensive tests
- simplify motor mass calculation and remove efficiency factors
- embed BasicEntity in PhysicsState and standardize System interface
- adjust drag coefficient and simplify logging messages for clarity
- move thrust force calculation to RK4 integrator to avoid double-counting
- restructure ISA model with layer-based atmospheric calculations and improved error handling
- split rules system into smaller, focused methods for better maintainability
- update RecordManager to use logger dependency and deterministic hashing
- remove system directory filtering and update test suite to use CreateRecordWithConfig
- extract simulation exit conditions into dedicated shouldStopSimulation method
- improve rocket physics initialization and component handling
- improve parachute component OpenRocket parsing and test coverage
- improve error handling and code organization in rocket entity creation
- update wind effect plugin to use force-based calculations and improve logging
- replace mass parameter with density in bodytube component
- add ground collision handling and rocket initialization on launch rail
- **tests, matrix**: remove redundant matrix and vector equality assertion functions to simplify test code
- **aerodynamics, physics**: disable aerodynamic moments for debugging and remove ground collision handling from PhysicsSystem
- **simulation, physics**: implement inertia tensor calculations and enhance RK4 integration for angular motion updates
- **simulation, physics, aerodynamics**: enhance inertia tensor calculations, integrate RK4 for state updates, and streamline force accumulation

## v0.7.3 (2025-05-07)

### Fix

- **bench, server, logger, plugins**: update CSV event loading to handle 3-column format and improve logger initialization

## v0.7.2 (2025-05-06)

### Refactor

- **bench, atmosphere, blackscholes**: simplify BlackScholes turbulence model and add atmosphere Model interface
- **plugins**: implement geometric brownian motion for turbulence simulation

## v0.7.1 (2025-05-06)

### Fix

- **bench**: simplify LoadEventInfo by reusing loadCSV helper and improving error messages

## v0.7.0 (2025-05-06)

### Feat

- **reporting**: Implement report data API with SVG plot generation and JSON response format

### Fix

- **server**: Add os import for filepath operations
- **reporting, bench**: Add flight event summaries and refactor event processing
- **templates**: Update API version handling and upgrade templ from v0.3.857 to v0.3.865
- **report, storage, server**: Add report generation with environmental conditions and flight metrics
- **bench**: Update event mapping and parsing for EUROC24 ground truth data format
- **bench**: Update CSV parser to handle new ignored column in event data files
- **windeffect**: Add config param to WindEffectPlugin and support 4-column CSV format in datastore

### Refactor

- **reporting**: move report template to external file and add asset copying
- **reporting, bench, server**: Refactor report generation to use markdown and improve event data handling

## v0.6.1 (2025-05-06)

### Fix

- **storage**: Fix log formatting and add error return in Write method
- **storage**: Replace log package with logf logger and update error handling

## v0.6.0 (2025-05-06)

### Feat

- **weather**: Add config to plugin system and implement weather client for wind data
- **reporting**: Add report generation and download functionality with PDF placeholder

### Fix

- **weather, plugin, blackscholes, simulation**: Update plugin initialization to pass config and add weather client tests
- **weather, blackscholes**: Add weather client tests and update plugin initialization interface
- **server**: Add HTMX response to delete handler with updated record list
- **server**: Remove debug logs and fix HTMX header casing in delete record handler
- **bench**: Update CSV parsing to handle NaN/Inf and adjust column indices for events/states
- **bench, openrocket, systems**: Add motor designation field and improve test coverage for data loading and benchmarks
- **bench**: Improve event time comparison handling and store benchmark results in user home dir

### Refactor

- **server, entities, openrocket**: Refactor mass calculation and add unit tests for component mass calculations
- **server**: Add request header logging and initialize DataHandler with logger in tests
- **server, reporting, storage, pages**: Add error handling and report generation for record deletion and downloads
- **bench, config**: Refactor benchmark system with new Run interface and improved data handling

## v0.5.0 (2025-05-05)

### Feat

- **bench, storage, simulation, systems, types**: Add event tracking and simulation data output for rocket dynamics

### Fix

- **systems**: Remove debug logging and enable angular velocity reset in ground collisions
- **systems**: Refactor physics system with improved force calculations and error handling
- **systems, simulation**: Add logging to aerodynamics system and remove motor update from physics loop

## v0.4.2 (2025-05-03)

### Refactor

- **entities**: Clean up rocket mass calculation and fix code formatting

## v0.4.1 (2025-05-01)

### Fix

- *****: Refactor rocket mass calculation and improve component initialization

## v0.4.0 (2025-05-01)

### Feat

- **cmd/bench**: Benchmarking added

## v0.3.8 (2025-04-29)

### Refactor

- move simulation output path from CLI flag to config.yaml

## v0.3.7 (2025-04-29)

### Fix

- **config, benchmark.yaml**: Update paths to relative and bump Go version to 1.23 in CI workflow

## v0.3.6 (2025-04-29)

### Fix

- **benchmark.yaml**: Update bench CLI flag from --benchdata to --resultsdir in benchmark workflow

## v0.3.5 (2025-04-29)

### Fix

- **bench**: Add CLI flags and integrate simulation with benchmark comparison

### Refactor

- improve benchmark config and add multi-benchmark support with result formatting

## v0.3.4 (2025-04-28)

### Refactor

- **openrocket**: schema types and add logging for mass calculations

## v0.3.3 (2025-04-28)

### Fix

- **server**: Refactor API pagination from page/size to limit/offset
- **server**: Add version parsing safety checks and refactor API version handling

### Refactor

- **plugin**: Remove unused test code and add missing go.work.sum dependency
- **plugin**: extract plugin compilation logic into separate function
- **server**: Remove unused parseOrDefaultInt helper function from handlers
- **server,-plugin-compilation,-storage-mgr**: Refactor record sorting to use CreationTime and add unit tests for handlers

## v0.3.2 (2025-04-27)

### Fix

- **rocket-entity,-airframe-openrocket-schema**: Add massProvider interface and fix formatting in rocket mass calculations

## v0.3.1 (2025-04-27)

### Refactor

- **rocket-entity**: Refactor mass calculation into helper functions for better code organization

## v0.3.0 (2025-04-27)

### Feat

- **benchmark,-rocket-components,-openrocket-parsing**: Added more accurrate vehicle mass calculation to simulation and updated benchmark and openrocket code
- **benchmark-framework-and-@ULAS-HiPR-initial-dataset-from-EuRoc24-competition**: Added various sensor dumps from the flight, which were sourced from COTS parts

### Fix

- **benchmark**: Update csv format and improve float comparison with rel. tolerance

### Refactor

- **rocket-entity**: Refactor motor mass calculation into separate helper function
- **benchmark-and-ulas-hipr-euroc24**: Tidied benchmark comparison and tolerance logic

## v0.2.1 (2025-04-27)

### Refactor

- **plugin_manager**: Removed unused methods

## v0.2.0 (2025-04-27)

### Feat

- **plugins**: add plugin compilation and change logger to interface type

### Fix

- **plugin-compilation**: Use search for go exec in PATH instead of assumption that it was available

## v0.1.3 (2025-04-27)

### Fix

- **storage**: Add strings import for filepath manipulation in storage package

## v0.1.2 (2025-04-27)

### Fix

- **storage**: Fixed use of RecordsList via templ

## v0.1.1 (2025-04-27)

### Fix

- Add strings package import to storage module files

## v0.1.0 (2025-04-27)

### Feat

- **ui**: Add link styling and improve table layout with pagination enhancements
- **data**: Add filtering and sorting options to simulation records table
- **ui**: Implement dark theme styles and enhance footer/navbar layout
- **explorer**: Enhance plot functionality with download option and update tracking
- **explorer**: Add endpoint and frontend integration for fetching and plotting JSON data
- **explorer**: Refactor data structures and update rendering logic for motion, dynamics, and events
- **explorer**: Add plotting functionality with dynamic axis selection
- **storage**: Add ReadHeadersAndData method to read headers and data separately
- **explorer**: Implement tabbed interface for viewing motion, dynamics, and events records
- **explore**: Add explore route and template for detailed record viewing
- **index.html**: Add response handling for simulation start feedback
- **server**: Added data views and handlers to index simulations
- implement validation for LogParasiteSystem and add unit tests
- add entity validation and unit tests for PhysicsSystem
- add LaunchRailSystem methods and corresponding tests
- Refactor PhysicsSystem to use pointers for entities; improve performance and memory efficiency
- Refactor force application and state update logic in PhysicsSystem; enhance ground collision handling and improve code readability
- Refactor Motor update logic for improved thrust and mass calculations; enhance burnout handling and add unit tests
- Improve error handling in Motor state transitions during ignition and burnout
- Improve error handling in NewBodytubeFromORK and NewRocketEntity functions; add unit tests for rocket entity creation
- Add unit tests for ISAModel including temperature, atmosphere, and speed of sound calculations
- Add temperature calculation methods and improve speed of sound handling in aerodynamic system
- Enhance landing detection and add LogParasite and StorageParasite systems with tests
- Add unit tests for AerodynamicSystem functionality and create launch rail test file
- Refactor entity type from physicsEntity to PhysicsEntity across systems
- Simplify landing detection logic in physics system
- Refactor motor update logic and enhance FSM state management
- Add unit tests for FlightStats functionality
- Improve motor thrust calculation and update flight stats handling
- Enhance motor and aerodynamic systems with configuration support and improved calculations
- Refactor system update methods to return error and enhance interface consistency
- Add validation for atmosphere configuration parameters and improve test assertions
- Refactor motor state management and enhance physics calculations for lift and drag
- Enhance simulation components with new methods and structural improvements
- Add atmosphere configuration to launchsite and update ISA model calculations
- Update simulation parameters for improved stability and performance
- Add launch rail system to simulation and update rocket motion constraints
- Implement flight statistics tracking and rules system for simulation
- Add mass calculation to TrapezoidFinset and improve motor state handling
- add mock components and systems for ECS testing
- implement ECS architecture with entities, components, and systems

### Fix

- **docker_publish**: skip docker build/push when no version bump occurs
- **docs**: Remove logo from README and delete git-czrc config file
- handle invalid mass values and exclude docs from sonar analysis
- **workspace**: Update dependencies and add missing hash values in go.work.sum
- **motor, parachute, aero, rules**: Better state validation and testing
- **pagination**: Update pagination links to use query parameters for navigation
- **sonar**: Update exclusions to include static files in sonar-project.properties
- **routes**: Update endpoint from /explorer to /explore for JSON data fetching fix(explorer): Correct hidden input value syntax in plot form refactor(explorer_templ): Improve string handling for record hash in template
- **handler**: Normalize JSON keys in GetExplorerData response to lowercase
- **server**: Moved log to before blocking call
- **server**: Linting complaints on unchecked errs
- **templates**: Better navbar and pages
- **js**: Removed template literals
- **css**: Duplicate selector fixed
- **server/storage/templates**: Fixed last modified time typo with time.Now(), added Delete record
- **server/templates**: Switched from html to templ with proper SOCs
- **server**: Load all html files to depth of one subfolder
- **server**: Propagated the app config to the submission to the simulation, hash and server itself.
- **server**: Tidy up server simManager call hanging resource
- **rules**: Simplify landing detection logic in processRules function
- **logger**: Instatiate via level, coverage upped too
- **config**: move into one file
- **records**: Enhance GetRecord to properly initialize storage services and handle errors
- **records**: Update ListRecords to correctly list records with last modified time
- **server**: Enhance runSim to handle record creation and error reporting
- **records**: Update ListRecords to set Timestamp based on last modified time
- **server**: Remove goroutine from simulation start for synchronous execution
- **index.html**: Add expands for certain details
- **server**: form based sim start
- **data-template**: Show dtime as main heading instead of hash
- handle error during storage initialization in StorageParasiteSystem

### Refactor

- **explorer**: Enhance pagination and layout for better usability
- **storage**: Extract data handling logic into separate methods in StorageParasiteSystem
- **storage**: Implement StorageInterface and update StorageParasiteSystem to use it
- **data**: Remove unused data view template to streamline codebase
- **systems**: Removed unused Priority property
- update config structure to use Engine and Setup types
- update orientation integration method and remove unused functions
- remove unused stability force calculation and applyForce method from PhysicsSystem
- add dynamics storage to test setup and ensure proper cleanup
- remove obsolete test files for LogParasiteSystem, StorageParasiteSystem, AerodynamicSystem, and LaunchRailSystem
- apply wind effect in BeforeSimStep and update state velocity handling
- integrate dynamics parasite system into simulation and enhance entity management
- improve error handling in storage initialization and closure methods
- ensure data is flushed correctly in storage write method and enhance log output with additional state information
- enhance rocket entity dynamics by initializing angular properties and improving aerodynamic moment calculations
- enhance rocket entity state handling and integrate angular velocity updates in physics calculations
- remove redundant core system update logic and streamline state propagation to rocket entity
- streamline simulation update process by reusing physics state and ensuring proper order of system and plugin execution
- enhance rocket entity initialization and validation, improve physics state handling, and add error checks in force calculations
- change entity storage to pointer slices in aerodynamic, launch rail, log parasite, and storage parasite systems for improved performance
- update WindEffectPlugin to use PhysicsState for consistency across systems
- replace PhysicsEntity with PhysicsState across systems for improved clarity and consistency
- simplify designation handling by removing validator interface and updating related code
- optimize TestWorld_Update by using pointers for mock systems
- rename parameter in Load function for clarity
- move MockHTTPClient implementation to a separate file and add unit tests for its Post method
- introduce HTTP client and designation validator interfaces; update thrustcurve loading logic
- remove thrustcurves_test.go file and associated mock HTTP client tests
- reorganize thrustcurves package; introduce MotorData type and API interaction functions
- introduce MotorData and Designation types; update motor data loading logic
- enhance motor data loading with error handling and rm logging
- update sonar exclusions to improve test coverage analysis
- remove unused logger output function and clean up tests
- improve configuration loading and validation; add logger integration
