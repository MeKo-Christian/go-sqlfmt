# Testing Guide

This project uses a comprehensive multi-level testing approach to ensure reliability and maintainability across all SQL dialects and features.

## Test Structure and Organization

### Test Organization Overview

**Public API Tests (`pkg/sqlfmt/`)**

- `format_test.go` - Comprehensive table-driven tests for core formatting functionality
- `golden_test.go` - Golden file-based testing using external test data files
- `snapshot_test.go` - Snapshot-based regression testing with automatic test data management
- `tokenizer_test.go` - Public API tokenizer tests that validate tokenization through formatting
- `*_formatter_test.go` - Dialect-specific formatter tests (postgresql, mysql, sqlite, db2, n1ql, plsql)

**Internal Unit Tests (`pkg/sqlfmt/` subdirectories)**

- `core/tokenizer_test.go` - Direct unit tests for internal tokenizer functionality
- `utils/dedent_test.go` - Unit tests for utility functions

**CLI Integration Tests (`cmd/`)**

- `format_test.go` - Command-line interface integration tests
- `validate_test.go` - Validation command tests
- `dialects_test.go` - Dialect selection tests

### Testing Approaches

#### 1. Table-Driven Tests

Used extensively in `format_test.go` for systematic testing of formatting scenarios:

- Clear separation of input, expected output, and configuration
- Easy to add new test cases and maintain existing ones
- Covers core formatting functionality across all dialects

Example structure:

```go
func TestFormat(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        config   *Config
        expected string
    }{
        {
            name:     "Basic SELECT",
            input:    "select * from users",
            config:   NewDefaultConfig(),
            expected: "SELECT\n  *\nFROM\n  users",
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Format(tt.input, tt.config)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

#### 2. Golden File Testing

- **Input files**: `testdata/input/{dialect}/`
- **Expected output**: `testdata/golden/{dialect}/`
- Automatically discovers and tests all `.sql` files
- Excellent for comprehensive dialect testing and regression detection

Structure:

```
testdata/
├── input/
│   ├── postgresql/
│   │   ├── complex_queries.sql
│   │   ├── json_arrays.sql
│   │   ├── cte_window.sql
│   │   └── plpgsql_functions.sql
│   ├── mysql/
│   │   ├── json_operations.sql
│   │   ├── window_functions.sql
│   │   └── stored_routines.sql
│   └── sqlite/
│       ├── upsert_operations.sql
│       └── json_functions.sql
└── golden/
    ├── postgresql/
    │   ├── complex_queries.sql
    │   └── ...
    └── ...
```

#### 3. Snapshot Testing

- Uses `github.com/gkampitakis/go-snaps` library
- Stores test output in `__snapshots__/` directories
- Perfect for catching unexpected formatting changes
- Provides Jest-like snapshot testing experience

Snapshot files are automatically generated and managed:

```
pkg/sqlfmt/__snapshots__/
├── snapshot_test.snap
├── postgresql_formatter_test.snap
├── mysql_formatter_test.snap
└── ...
```

#### 4. Internal Unit Testing

- Direct testing of internal components without going through public API
- Faster test execution and more targeted debugging
- Tests core functionality like tokenization, formatting logic, and utilities

## Running Tests

### All Tests

```bash
# Recommended: runs with verbose output
just test

# Standard Go test execution
go test ./...

# With coverage
just test-coverage
```

### Specific Test Categories

```bash
# Public API tests only
go test ./pkg/sqlfmt

# Internal core tests
go test ./pkg/sqlfmt/core

# Utility tests
go test ./pkg/sqlfmt/utils

# CLI tests
go test ./cmd
```

### Tests by Pattern

```bash
# All formatting tests
go test ./pkg/sqlfmt -run TestFormat

# All tokenizer tests
go test ./pkg/sqlfmt -run TestTokenizer

# Dialect-specific tests
go test ./pkg/sqlfmt -run TestPostgreSQL
go test ./pkg/sqlfmt -run TestMySQL
go test ./pkg/sqlfmt -run TestSQLite
go test ./pkg/sqlfmt -run TestPLSQL
go test ./pkg/sqlfmt -run TestDB2
go test ./pkg/sqlfmt -run TestN1QL
```

### Specialized Test Types

```bash
# Golden file tests only
just test-golden

# Snapshot tests only
just test-snapshots

# Update snapshots after changes
just update-snapshots

# Performance benchmarks
just test-benchmarks
```

## Dialect-Specific Testing

### PostgreSQL

```bash
# Comprehensive PostgreSQL tests
go test ./pkg/sqlfmt -run TestPostgreSQL -v

# Specific PostgreSQL formatter tests
go test ./pkg/sqlfmt -run TestPostgreSQLFormatter -v

# Feature-specific test patterns
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Dollar"    # Dollar-quoted strings
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Cast"      # Type casting
go test ./pkg/sqlfmt -run "TestPostgreSQL.*JSON"      # JSON operations
go test ./pkg/sqlfmt -run "TestPostgreSQL.*CTE"       # Common Table Expressions
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Window"    # Window functions
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Array"     # Array operations
```

### MySQL

```bash
# All MySQL tests
go test ./pkg/sqlfmt -run TestMySQL -v

# Feature-specific MySQL tests
go test ./pkg/sqlfmt -run "TestMySQL.*JSON"      # JSON operations
go test ./pkg/sqlfmt -run "TestMySQL.*CTE"       # Common Table Expressions
go test ./pkg/sqlfmt -run "TestMySQL.*Window"    # Window functions
go test ./pkg/sqlfmt -run "TestMySQL.*Upsert"    # UPSERT operations
```

### SQLite

```bash
# All SQLite tests
go test ./pkg/sqlfmt -run TestSQLite -v

# Alternative command
just test-sqlite

# Feature-specific SQLite tests
go test ./pkg/sqlfmt -run "TestSQLite.*JSON"      # JSON operations
go test ./pkg/sqlfmt -run "TestSQLite.*UPSERT"    # UPSERT operations
go test ./pkg/sqlfmt -run "TestSQLite.*Window"    # Window functions
go test ./pkg/sqlfmt -run "TestSQLite.*CTE"       # Common Table Expressions
```

## Test Data Management

### Golden File Organization

Test data is organized by dialect:

- **Input**: `testdata/input/{dialect}/*.sql`
- **Expected**: `testdata/golden/{dialect}/*.sql`

### Adding New Test Data

1. **Create input file**: Add your SQL query to the appropriate dialect input directory
2. **Generate golden file**: Run the formatter and save the output as the expected result
3. **Verify formatting**: Ensure the output matches your expectations
4. **Run tests**: Verify the new test passes

### Updating Test Expectations

When formatting logic changes:

```bash
# Update golden files (manual process)
# 1. Review changes carefully
# 2. Update golden files with new expected output
# 3. Run tests to verify

# Update snapshots (automated)
just update-snapshots
```

## Continuous Integration

### Pre-commit Testing

Before committing changes, run:

```bash
# Full test suite
just check

# Or individual components
just lint          # Linting
just fmt           # Code formatting
just test          # All tests
```

### Test Coverage

```bash
# Generate coverage report
just test-coverage

# View coverage in browser (if available)
go tool cover -html=coverage.out
```

### Performance Testing

```bash
# Run benchmarks
just test-benchmarks

# Benchmark specific functions
go test -bench=BenchmarkFormat ./pkg/sqlfmt
go test -bench=BenchmarkTokenizer ./pkg/sqlfmt/core
```

## Adding Tests for New Features

### Integration Tests

1. **Add to `format_test.go`**: Include table-driven test cases for end-to-end functionality
2. **Test multiple dialects**: Ensure the feature works across relevant SQL dialects
3. **Test edge cases**: Include boundary conditions and error scenarios

### Unit Tests

1. **Core functionality**: Add tests in appropriate packages for isolated components
2. **Utility functions**: Test utility functions in `utils/` package
3. **Tokenization**: Add tokenizer tests for new syntax elements

### Golden Files

1. **Create comprehensive examples**: Add realistic SQL queries using the new feature
2. **Cover variations**: Include different syntax variations and combinations
3. **Test complex scenarios**: Include edge cases and advanced usage patterns

### Dialect-Specific Tests

When adding dialect-specific features:

1. **Create dialect test file**: Add `{dialect}_test.go` if it doesn't exist
2. **Add formatter tests**: Test the dialect-specific formatter implementation
3. **Update dialect documentation**: Include testing examples in dialect docs

## Test Maintenance

### Regular Maintenance Tasks

1. **Review failing tests**: Investigate and fix failing tests promptly
2. **Update expectations**: Update test expectations when behavior intentionally changes
3. **Clean up obsolete tests**: Remove tests for deprecated features
4. **Optimize slow tests**: Identify and optimize slow-running tests

### Snapshot Management

1. **Review snapshot changes**: Carefully review snapshot diffs before updating
2. **Clean obsolete snapshots**: Remove unused snapshots periodically
3. **Organize snapshots**: Keep snapshot files organized and well-named

### Performance Monitoring

1. **Monitor test duration**: Track test execution time to catch performance regressions
2. **Benchmark critical paths**: Regular benchmarking of core formatting functions
3. **Memory usage**: Monitor memory usage in tests to catch memory leaks

## Troubleshooting Tests

### Common Issues

**Tests failing after changes**:

1. Check if formatting behavior intentionally changed
2. Update test expectations if change is intended
3. Fix implementation if change is unintended

**Snapshot tests failing**:

1. Review snapshot diffs carefully
2. Update snapshots if changes are correct: `just update-snapshots`
3. Fix implementation if snapshots show unexpected changes

**Golden file tests failing**:

1. Compare actual vs expected output
2. Update golden files if formatting improved
3. Check for regressions if output seems wrong

**Performance issues**:

1. Identify slow tests with `go test -v`
2. Profile test execution if needed
3. Optimize test data or implementation

### Debugging Test Failures

```bash
# Run specific failing test with verbose output
go test ./pkg/sqlfmt -run TestSpecificCase -v

# Run with race detection
go test -race ./...

# Run with memory profiling
go test -memprofile=mem.prof ./pkg/sqlfmt
```

## Test Best Practices

1. **Clear test names**: Use descriptive test names that explain what's being tested
2. **Independent tests**: Ensure tests don't depend on each other
3. **Comprehensive coverage**: Cover both happy paths and edge cases
4. **Maintainable assertions**: Use clear, maintainable assertion patterns
5. **Fast execution**: Keep tests fast-running to encourage frequent execution
6. **Documentation**: Document complex test scenarios and their purposes

The comprehensive testing approach ensures reliability across all supported SQL dialects and provides confidence when making changes to the formatting engine.
