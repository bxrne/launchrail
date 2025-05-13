# openrocket

Implementation of OpenRocket file parsing and integration for importing rocket designs into the simulation.

## Notes
- Parses OpenRocket (.ork) file format for rocket specifications
- Converts OpenRocket components to simulation-compatible models
- Preserves physical and aerodynamic properties from designs
- Maps OpenRocket parameters to LaunchRail component structure
- Handles different component types (nose cones, fins, body tubes)
- Ensures dimensional consistency during import process
- Supports validation of imported designs against constraints
- Provides data structures for representing OpenRocket schemas