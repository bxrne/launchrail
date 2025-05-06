# cmd/bench

`cmd/bench` hosts the benchmarking CLI for Launchrail.

Responsibility

- Provide a `launchrail bench` command to run performance benchmarks.
- Load CSV flight and benchmark datasets via `LoadFlightInfo` and related helpers.
- Compare actual vs expected metrics with configurable tolerances.
- Exit with status code 0 on success, 1 on any failure (for CI integration).

Contract

- Exposes `main.go` and `benchmark.go` defining `Benchmark` and `BenchmarkSuite`.
- Implements `BenchmarkSuite.Run()` to orchestrate data loads and result comparison.

Test Suite Overview

- `main_test.go` (using `testify`) covers:
  - CSV data loading (`datastore_test.go`).
  - Example hipr-euroc24 benchmark logic (`hipr_euroc24_test.go`).
  - Exit-code behavior for pass/fail.

Decisions & Gotchas

- Zero expected values switch tolerance logic to absolute difference.
- Tolerance scaling chosen to balance noise vs strictness; may need tuning per dataset.
- CSV schema changes must be mirrored in `datastore.go` loaders.
