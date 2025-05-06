# cmd/bench

## Description

The `cmd/bench` directory provides a Go-based benchmarking system designed to evaluate and validate the performance and accuracy of the LaunchRail simulation core. This system is crucial for continuous integration (CI) and ensuring the reliability of simulation results against known datasets.

### Key Features:

*   **Benchmark Suite:** Defines a `BenchmarkSuite` along with `Benchmark` and `BenchmarkResult` interfaces (`benchmark.go`) to structure and execute a collection of benchmarks.
*   **Data Handling:** Includes data structures like `FlightInfo` and utilities for loading benchmark data from CSV files (e.g., `LoadFlightInfo` in `datastore.go`).
*   **Example Implementation:** Contains an example benchmark, `hipr_euroc24.go`, which demonstrates how to implement specific benchmark tests against datasets like `hipr-euroc24`. This includes comparison logic with configurable tolerances (e.g., `compareFloat`) and analysis functions (e.g., `findApogee`, `findMaxVelocity`).
*   **Command-Line Interface:** The `main.go` file serves as the entry point. It parses command-line arguments (such as `--benchdata` to specify the location of benchmark data files), runs the defined benchmark suite, prints a summary of the results, and generates a benchmark hash (SHA1 of tag and timestamp, truncated to 8 hex characters).
*   **CI Integration:** The program exits with a status code of 0 for a successful run (all benchmarks pass) and 1 if any benchmark fails, facilitating easy integration into CI/CD pipelines.
*   **Testing:** Unit tests (e.g., `datastore_test.go`, `hipr_euroc24_test.go`) are provided using the `testify` library to ensure the correctness of data loading and comparison logic.

### How it Works:

1.  Benchmark data, typically in CSV format, is loaded.
2.  Simulations or calculations are performed based on the benchmark case.
3.  Actual results are compared against expected values defined in the benchmark data.
4.  Comparisons for floating-point numbers use a configurable tolerance. If an expected value is zero, the tolerance is treated as an absolute value.
5.  Each metric within a benchmark is reported as PASS or FAIL.
6.  An overall status for the suite is determined.
