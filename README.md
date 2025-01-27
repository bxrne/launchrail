# launchrail

[![Build and Test](https://github.com/bxrne/launchrail/actions/workflows/build_test.yaml/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/build_test.yaml) [![Lint and Vet](https://github.com/bxrne/launchrail/actions/workflows/lint_vet.yaml/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/lint_vet.yaml) [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=bxrne_launchrail&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=bxrne_launchrail) [![Coverage](https://sonarcloud.io/api/project_badges/measure?project=bxrne_launchrail&metric=coverage)](https://sonarcloud.io/summary/new_code?id=bxrne_launchrail) [![Go Reference](https://pkg.go.dev/badge/github.com/bxrne/launchrail.svg)](https://pkg.go.dev/github.com/bxrne/launchrail)

Launchrail is an open-source GNU General Public License v3.0 (GPL-3.0) project that aims to create a 6DOF High-Powered Rocket Simulator. The project aims to leverage [OpenRocket](http://openrocket.info/) design files to reduce config friction and uses [ThrustCurve](https://www.thrustcurve.org/) API for motor curves via designation. The project is also testing the question of whether atmospheric turbulence can be better modelled by using financial algorithms over the standard ISA model.

## Getting Started

```bash
git clone https://github.com/bxrne/launchrail.git
cd launchrail

go run ./cmd/launchrail
air # for hot reload (dev)
```

### Testing

Run locally with the command below, runs on change for PRs and on main push (see [build and test CI](.github/workflows/build_test.yaml)).

```bash
go test ./... -v 
```


## Built With

- [Go](https://golang.org/) - The Go Programming Language
- [OpenRocket](http://openrocket.info/) - OpenRocket is a free, fully featured model rocket simulator that allows you to design and simulate your rockets before actually building and flying them.
- [ThrustCurve](https://www.thrustcurve.org/) - ThrustCurve is a comprehensive model rocket motor database.


## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
