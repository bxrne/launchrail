# launchrail (CLI Entry Point)

This package provides the primary command-line interface for Launchrail.

Responsibility:
- Parse flags and environment configuration.
- Initialize dependencies: storage, simulation manager, logger.
- Wire together components to run simulations.

Contract:
- Exposes `main()` which reads `--config` and `--benchdata` flags.
- Entrypoint for CI benchmarks via `cmd/bench` suite.

Test Suite Overview:
- No direct tests; behavior validated indirectly via `cmd/bench` benchmarks and `handlers_test.go`.

Decisions & Gotchas:
- Relies on external `config.yaml` for engine parameters.
- Errors during initialization abort the process with non-zero exit.
- Benchmarks in `cmd/bench` call this CLI under the hood; keep flags backward-compatible.
