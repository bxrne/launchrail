# Contributing to Launchrail

Thank you for your interest in contributing to Launchrail! We welcome contributions from the community to help make this project better.

## How to Contribute

1. **Fork the repository** and clone your fork locally:
   ```sh
   git clone https://github.com/bxrne/launchrail.git
   cd launchrail
   ```
2. **Create a new branch** for your feature or bugfix:
   ```sh
   git checkout -b my-feature-branch
   ```
3. **Install dependencies** and ensure the project builds:
   ```sh
   go build ./...
   ```
4. **Write tests** for your changes and run them:
   ```sh
   go test ./... -v
   ```
5. **Lint your code** before submitting:
   ```sh
   golangci-lint run ./...
   ```
6. **Use Conventional Commits with Commitizen** for all commits:
   ```sh
   cz commit
   ```
   - This project uses [Commitizen](https://commitizen-tools.github.io/commitizen/) and [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) to automate semantic versioning (see `.cz.toml`).
   - Install Commitizen: `pip install commitizen` or `npm install -g commitizen`

7. **Push your branch** and open a Pull Request (PR) against `main`.
   - Fill out the PR template and describe your changes clearly.
   - Reference any related issues in your PR description.

## Code Style and Quality
- Follow Go best practices and idiomatic Go style.
- Keep cognitive complexity low; prefer readable, maintainable code.
- Write clear comments and documentation for public APIs.
- All code must pass tests and linting before review.

## Docker & CI
- If your change affects Docker or CI, ensure the Docker build passes:
  ```sh
  DOCKER_BUILDKIT=1 docker build -t launchrail:latest .
  ```
- The project publishes images to [GitHub Container Registry (GHCR)](https://github.com/bxrne/launchrail/pkgs/container/launchrail).

---

Thank you for helping make Launchrail awesome!
