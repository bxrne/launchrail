# states

State management components that track and update physical properties of simulation entities during flight.

## Notes
- Defines data structures for entity state representation
- Tracks physical properties like position, velocity, orientation
- Maintains acceleration, mass, and other key flight parameters
- Implements state transition logic for different flight phases
- Provides thread-safe access to entity state information
- Supports serialization of state data for storage and analysis
- Enables efficient querying of current and historical state
- Integrates with physics systems for consistent state updates