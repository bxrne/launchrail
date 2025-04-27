# Launchrail

[![Build and Test](https://github.com/bxrne/launchrail/actions/workflows/build_test.yaml/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/build_test.yaml)
[![Lint and Vet](https://github.com/bxrne/launchrail/actions/workflows/lint_vet.yaml/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/lint_vet.yaml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=bxrne_launchrail&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=bxrne_launchrail) [![Coverage](https://sonarcloud.io/api/project_badges/measure?project=bxrne_launchrail&metric=coverage)](https://sonarcloud.io/summary/new_code?id=bxrne_launchrail) [![Benchmark](https://github.com/bxrne/launchrail/actions/workflows/benchmark.yaml/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/benchmark.yaml) [![CodeQL](https://github.com/bxrne/launchrail/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/github-code-scanning/codeql) [![Dependabot Updates](https://github.com/bxrne/launchrail/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/dependabot/dependabot-updates) [![Docker Build & Publish](https://github.com/bxrne/launchrail/actions/workflows/docker_publish.yaml/badge.svg)](https://github.com/bxrne/launchrail/actions/workflows/docker_publish.yaml) [![Go Reference](https://pkg.go.dev/badge/github.com/bxrne/launchrail.svg)](https://pkg.go.dev/github.com/bxrne/launchrail)


Launchrail is an open-source 6DOF High-Powered Rocket Simulator. It leverages [OpenRocket](http://openrocket.info/) design files for configuration and [ThrustCurve](https://www.thrustcurve.org/) API for motor curves. The project also explores using financial algorithms to model atmospheric turbulence.

**Repository:** [https://github.com/bxrne/launchrail](https://github.com/bxrne/launchrail)

---

## 🚀 Getting Started

```sh
git clone https://github.com/bxrne/launchrail.git
cd launchrail

go run ./cmd/launchrail
# For hot reload (development)
air
```

---

## 🐳 Docker Usage

You can run Launchrail as a Docker container, either by building locally or pulling from [GHCR](https://github.com/bxrne/launchrail/pkgs/container/launchrail).

### Build and Run Locally

```sh
DOCKER_BUILDKIT=1 docker build -t launchrail:latest .
docker run --rm -it -p 8080:8080 launchrail:latest
```

### Pull and Run from GHCR

```sh
docker pull ghcr.io/bxrne/launchrail:latest
# Or pull a specific version
docker pull ghcr.io/bxrne/launchrail:<tag>
docker run --rm -it -p 8080:8080 ghcr.io/bxrne/launchrail:latest
```

---

## 🧑‍💻 Contributing & Git Workflow

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for full guidelines.

- **Clone and branch:**
  ```sh
  git clone https://github.com/bxrne/launchrail.git
  git checkout -b my-feature-branch
  ```
- **Run tests:**
  ```sh
  go test ./... -v
  ```
- **Lint:**
  ```sh
  golangci-lint run ./...
  ```
- **Commit using Commitizen:**
  ```sh
  cz commit
  ```
  This project uses [Commitizen](https://commitizen-tools.github.io/commitizen/) and [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for semantic versioning (see `.cz.toml`).
  Semantic versioning is enforced for all releases and PRs. Please use `cz commit` to ensure proper versioning and changelogs.

---

## 🧪 Testing

Run all tests:
```sh
go test ./... -v
```

---

## 🛠️ Built With

- [Go](https://golang.org/) — The Go Programming Language
- [OpenRocket](http://openrocket.info/) — Model rocket design and simulation
- [ThrustCurve](https://www.thrustcurve.org/) — Model rocket motor database

---

## 📦 License

This project is licensed under the GNU General Public License v3.0 — see the [LICENSE](LICENSE) file for details.
