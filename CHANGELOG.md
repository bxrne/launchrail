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
