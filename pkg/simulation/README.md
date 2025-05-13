# simulation

Core simulation framework that orchestrates the physics-based rocket flight simulation process.

## Notes
- Implements numerical integration methods (RK4) for flight dynamics
- Orchestrates system updates in the appropriate sequence
- Manages simulation time step and advancement
- Detects and handles flight events (launch, burnout, apogee)
- Provides interfaces for system integration and data collection
- Implements different simulation modes (standard, real-time)
- Ensures consistent physics calculations across simulation systems
- Handles simulation initialization and termination conditions