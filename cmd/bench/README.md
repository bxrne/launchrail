# bench

Benchmarking CLI for Launchrail that provides performance testing capabilities and metrics comparison.

## Notes
- Provides benchmarking functionality for testing LaunchRail's performance
- Loads CSV flight and benchmark datasets via specialized helpers
- Compares actual vs expected metrics with configurable tolerances
- Exits with status code 0 on success, 1 on any failure (for CI integration)
- Implements `BenchmarkSuite.Run()` to orchestrate data loads and result comparison
- Reads configuration directly from config files, not CLI flags
- Creates timestamped output directories in the user's home directory
- Zero expected values switch tolerance logic to absolute difference
- Tolerance scaling chosen to balance noise vs strictness; may need tuning per dataset
- CSV schema changes must be mirrored in `datastore.go` loaders