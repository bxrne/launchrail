# launchrail

[![Build and Test](https://github.com/bxrne/launchrail/actions/workflows/build_and_test.yml/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/build_and_test.yml)

+ [![cov](https://bxrne.github.io/launchrail/badges/coverage.svg)](https://github.com/bxrne/launchrail/actions)

Experimenting with the idea of creating a 6-DOF High Power Rocketry simulator. This is a work in progress and is not yet functional.


## Roadmap

**Layer 1: Core Simulation (Proof of Concept)**

- **3DOF Physics Engine**: Initial implementation will focus on a 3-degree-of-freedom simulation (x, y, z position) with basic drag and thrust calculations.
- **Parachute and Event Handling**: Implement a basic event system to deploy parachutes at predefined altitudes.
- **Basic CLI Interface**: Before the GUI, a command-line interface will allow users to load configurations and run simulations.

*This will provide a working core simulation and basic usability, proving the essential functionality of the physics and event systems.*

**Layer 2: Enhanced Physics and GUI Integration (Minimum viable product)**

- **6DOF Physics Expansion**: Extend the physics engine to support 6-degree-of-freedom (3D position plus orientation).
- **GUI Implementation**: Build the desktop GUI for easier user interaction, allowing users to visualise the rocket's trajectory in real-time.
- **Event-Driven Improvements**: Expand the event system to include more complex triggers like velocity changes and time-based events.

*With these features, the system will support more complex and realistic simulations, significantly improving usability.*

**Layer 3: External Integration and Advanced Features**

- **OpenRocket/RASAero Integration**: Add seamless import of configurations from OpenRocket and RASAero to reduce human error.
- **Simulation Optimisations**: Optimise the simulation engine for faster performance, using better integration algorithms (e.g., Runge-Kutta).
- **Advanced Visualisations**: Add more advanced 3D visualisations of rocket flight and post-simulation analysis tools.

*This layer focuses on enhancing accuracy, user experience, and integration with existing rocketry tools.*


