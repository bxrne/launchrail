# launchrail

Final Year Project for BSc component of MSc (3+1) Immersive Software Engineering at the University of Limerick.
Aiming to improve the process of design iteration for sounding and research rockets by shortening the feedback loop by simulating rocket flights in a 6DOF physics simulation at a location with a given weather profile all from a `.ork` file.

This is motivated by my work at ULAS-HiPR where we are a student team developing High Powered Rockets for competitions and to further our skills. We mainly design our rockets in OpenRocket and simulate with RocketPy but it is was difficult to maintain and slower to iterate on design based on results.

> Currently working on parsing `.ork` files into structs that can be used to generate a simulation.

## Start

```bash
git clone https://github.com/bxrne/launchrail.git
cd launchrail

go mod tidy
go test ./...

go run ./cmd
```

## Internal packages

- `logger` - Singleton styled logging
- `ork` - Parse `.ork` files into structs mappable to our bare formats.


