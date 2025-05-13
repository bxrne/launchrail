# simulation

Core simulation engine that manages and executes rocket flight simulations with physics-based modeling.

## Notes
- Orchestrates the entire simulation workflow
- Initializes and manages the Entity Component System (ECS)
- Implements numerical integration methods for physics calculations
- Handles event detection and processing (e.g., motor burnout, apogee)
- Maintains simulation state and time progression
- Integrates with storage systems for recording simulation results
- Provides diagnostic tools for simulation debugging
- Supports different simulation modes (standard, batch, real-time)