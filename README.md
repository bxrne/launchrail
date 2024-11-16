# launchrail

[![Build and Test](https://github.com/bxrne/launchrail/actions/workflows/ci.yaml/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/ci.yaml) [![codecov](https://codecov.io/gh/bxrne/launchrail/graph/badge.svg?token=HDTJQK087F)](https://codecov.io/gh/bxrne/launchrail) [![Go Reference](https://pkg.go.dev/badge/github.com/bxrne/launchrail.svg)](https://pkg.go.dev/github.com/bxrne/launchrail)

> ”Can the Black-Scholes model be adapted to simulate risk-neutral baseline trajectories for high powered rocket launches?”

Final Year Project for BSc component of MSc (3+1) Immersive Software Engineering at the University of Limerick. Aiming to improve the process of design iteration for sounding and research rockets by basing the simulation on the Black-Scholes model which allows creating a risk neutral baseline trajectory for the rocket by basing market volatility on atmospheric turbulence and predicting the price of the stock (altitude) at a future time (apogee) especially useful for optimising to competition requirements. This is motivated by my work at ULAS-HiPR where we are a student team developing High Powered Rockets for competitions and to further our skills. We mainly design our rockets in OpenRocket and simulate with RocketPy but it is was difficult to maintain and slower to iterate on design based on results.

## Usage

```bash
git clone https://github.com/bxrne/launchrail.git
cd launchrail

go mod download
go test ./...

go run ./cmd
```

## Configuration

Application can be configured via the `config.yaml` file in the root directory. The configuration file is in YAML format and contains the following fields:

Note: Go allows using `testdata` directory for testing purposes, so the configuration file is loaded from the `testdata` directory in testing.

```yaml
app:
  version: "0.0.1"
  license: "GNU GPL v3"
  repo: "https://github.com/bxrne/launchrail"
logs:
  file: "launchrail.log"
```

## Literature Review (WIP)

There is a lot of literature in aerospace simulation, notably for aerospace vehicle simulation and analysis under 6 degrees of freedom [1] and for sounding rockets the simulation and analysis of their results such as landing zones and construction error risk impact on stability [2]. The intersection of options pricing and aerospace design is the integrating of economic principles of cost, return, risk, flexibility, and other concepts from Black-Scholes option pricing theory and the multiplicative binomial process [3]. There is a gap to produce a risk-neutral baseline simulation method for sub orbital rockets that will improve mission and vehicle design via comparison in simulation.

## Aim

This project will deliver an open source desktop application to launch simulations based on Rocket designs by adapting the Black-Scholes model to a simulation framework to allow users to bring their physical designs and simulate the rocket flying at the expected launch site under 6 degrees of freedom and then analyse and playback different flight characteristics (stability, thrust etc). There are not many resources in this space so I will be open sourcing it under GNU GPL-3.0. The Black-Scholes simulation framework will be made atomically available as a module for general applications. The desktop app will comprise of a multithreaded physics engine written in Go and an Electron based frontend written in Typescript.

## Plan

### Foundational work

- **Literature Review**
  - Define work done in this vein already
  - Expand search scope to further refine foundational work
- **User Requirements**
  - Investigate functional requirements
  - Investigate HCI and accessibility requirements
  - Define deliverable qualities (non-functional requirements)

### Proof of Concept

- **3DOF Physics Engine**: Initial implementation will focus on a 3-degree-of-freedom simulation (x, y, z position) with basic drag and thrust calculations. Adapted Black Scholes model implemented for deviation analysis
- **OpenRocket/RASAero Integration**: Import physical design and aerodynamic performance from related tools
- **Basic TUI Interface**: Before the GUI, a command-line interface will allow users to load configurations and run simulations.

### MVP

- **6DOF Physics Expansion**: Extend the physics engine to support 6-degree-of-freedom (3D position plus orientation).
- **Parachute and Event Handling**: Implement an event system to handle rail clearance, apogee, motor burnout and CATO as well as deploying parachutes with lag.

### Release v0.1

- **GUI Implementation**: Build the desktop GUI for easier user interaction, allowing users to visualise the rocket's trajectory in real-time.
  - GUI design and development will go through iterations against these harnesses:
    - Usability: Nielsen’s 10 Usability Heuristics [4]
    - Accessibility: WCAG 2.2 (Web Content Accessibility Guidelines) [5]
- **Advanced Visualisations**: Add more advanced 3D visualisations of rocket flight and post-simulation analysis tools such as playback.

### Future Work

- **User Interviews:** Iterate further on design and flow of the application driven from user interviews within the domain
- **Plugin System:** Embed a plugin system to allow the tool to be furthered via open-source, likely embedding Lua as the Vim → Neovim move did.

## References

1. Zipfel, P.H. (2005) *Advanced Six Degrees of Freedom Aerospace Simulation and Analysis in C++*, Reston, Va: American Institute of Aeronautics & Astronautics AIAA.
2. Trevisi, F., Poli, M., Pezzato, M., Di Iorio, E., Madonna, A., Bressanin, N., & Debei, S. (2017). Simulation of a sounding rocket flight's dynamic. In Proceedings of the 2017 IEEE International Workshop on Metrology for AeroSpace (METROAEROSPACE) (pp. 296-300). IEEE.
3. Gray, A. A., Arabshahi, P., Lamassoure, E., Okino, C., & Andringa, J. (2005). A real options framework for space mission design. In Proceedings of the 2005 IEEE Aerospace Conference, Vols 1-4 (pp. 137-146). IEEE.
4. Nielsen, J. and Mack, R.L. (1994) *Usability Inspection Methods*.
5. W3C (2023). *Web Content Accessibility Guidelines (WCAG) 2.2*. [online] www.w3.org. Available at: [https://www.w3.org/TR/WCAG22/](https://www.w3.org/TR/WCAG22/).
