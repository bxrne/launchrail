# launchrail

[![gobuild](https://github.com/bxrne/launchrail/actions/workflows/gobuild.yml/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/gobuild.yml)
> ”Can the Black-Scholes model be adapted to simulate risk-neutral baseline trajectories for high powered rocket launches?”

Final Year Project for BSc component of MSc (3+1) Immersive Software Engineering at the University of Limerick.
Aiming to improve the process of design iteration for sounding and research rockets by basing the simulation on the Black-Scholes model which allows creating a risk neutral baseline trajectory for the rocket by basing market volatility on atmospheric turbulence and predicting the price of the stock (altitude) at a future time (apogee) especially useful for optimising to competition requirements. This is motivated by my work at ULAS-HiPR where we are a student team developing High Powered Rockets for competitions and to further our skills. We mainly design our rockets in OpenRocket and simulate with RocketPy but it is was difficult to maintain and slower to iterate on design based on results.


## Start

```bash
git clone https://github.com/bxrne/launchrail.git
cd launchrail

go mod tidy
go test ./pkg/...

go run ./cmd
```

