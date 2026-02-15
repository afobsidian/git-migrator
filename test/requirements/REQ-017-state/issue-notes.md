# Issue Notes: REQ-017-state Tests

## Summary
Tests for the REQ-017-state requirement are intermittently failing in CI or during full test suite runs. Failures are observed during both the build and execution phases. These failures are inconsistent and are resolved when the tests are run individually.

## Observations

1. **Failed Test Case:**
   - The `TestStateDatabase` suite consistently fails during build or execution with errors related to SQLite driver initialization or temporary build file cleanup.

2. **Environment-Specific Failures:**
   - The failures are not reproducible reliably on local environments but occur during CI or when running full test suite in parallel (`go test ./...`).

3. **SQLite Driver Issues:**
   - `modernc.org/sqlite` is being used as the SQLite driver.
   - Errors such as "disk I/O error" or "driver not found" were observed in earlier debugging steps.

4. **Cache/Concurrency Problems:**
   - The Go build system might be cleaning temporary files while another test depends on them.
   - `-count` and serial execution options (`-p 1`) reduced failures but did not resolve them fully.

## Resolution Attempts

### 1. **Run Tests Serially:**
- Command: `GOMAXPROCS=1 go test -p 1 ./...`
- Result: Reduced frequency of failures but did not eliminate them.

### 2. **Disable Go Cache:**
- Command: `go test -count=1 ./...`
- Result: Works for individual test files but not reliable for the full suite.

### 3. **Simplified Tests:**
- Added a basic database connection test (`simple_test.go`).
- Result: Passed consistently.

### 4. **Temporary Directory Isolation:**
- Used `t.TempDir()` to isolate database file creation.
- Result: No effect on failure frequency.

### 5. **Inspect SQLite Driver:**
- Reviewed `modernc.org/sqlite` for known issues with Go build caching.
- Result: No conclusive findings; driver functions as expected.

## Next Steps

1. Limit parallelism in CI for this test file:
   - Update CI pipelines to use `-count=1` and `-p 1` flags.

2. Investigate SQLite Driver:
   - Check if alternative SQLite drivers (e.g., `github.com/mattn/go-sqlite3`) are more reliable.

3. Defer to Maintainers:
   - Developers familiar with SQLite internals or Go's build cache should be consulted for further debugging.

4. Document Intermittency:
   - Update roadmap, README, and project documentation to reflect this intermittent issue.

## Workaround
Despite failures, tests pass reliably when executed individually. For now, they can be run as isolated cases during critical validations until a permanent fix is implemented.