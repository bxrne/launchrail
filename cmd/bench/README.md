## cmd/bench

### Responsibility
- Provides a Go benchmark system for evaluating the performance of the simulation.
- Defines interfaces for benchmarks and benchmark suites.
- Implements example benchmarks for specific datasets.
- Provides a command-line interface for running benchmarks and reporting results.

### Scope
- Core interfaces (`Benchmark`, `BenchmarkResult`) and suite (`BenchmarkSuite`).
- Data structures (`FlightInfo`, etc.) and CSV loaders (`LoadFlightInfo`, etc.).
- Example benchmark implementation for `hipr-euroc24` dataset.
- Entry point, parses command-line flags, runs the suite, prints results, and exits with a status code for CI.

### Test Suite Overview
- Includes unit tests for data loading and benchmark comparisons.
- Uses `testify` for assertions.
- Tests cover data loading, comparison logic, and result reporting.

### Decisions & Potential Gotchas
- Uses CSV files for benchmark data.
- Comparison logic includes configurable tolerance for floating-point values.
- Handles zero expected values in comparisons by interpreting tolerance as absolute.
- Exits with a non-zero status code on failure, which is used for CI integration.
