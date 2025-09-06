# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Just Commands (Primary)

- `just` - Default target (builds the module)
- `just test` - Run all tests with verbose output
- `just test-benchmarks` - Run benchmarks
- `just test-coverage` - Run tests with coverage report
- `just lint` - Run golangci-lint with custom config
- `just lint-fix` - Run golangci-lint with automatic fixes
- `just fmt` - Format all code using treefmt
- `just check` - Run all checks (format, lint, test, tidy, generated)
- `just setup-deps` - Install all development tools
- `just help` - Show all available commands

### Direct Go Commands

- `go test ./...` - Run all tests
- `go test -bench=. ./...` - Run benchmarks
- `go test ./sqlfmt` - Run tests for the main package only
- `go build ./...` - Build the module
- `go mod tidy` - Clean up module dependencies

## Architecture

This is a Go library that formats SQL queries with support for multiple SQL dialects. The architecture is organized around a common `Formatter` interface with dialect-specific implementations.

### Core Components

**Main Package (`sqlfmt/`)**

- `format.go` - Main entry points (`Format`, `PrettyFormat`, `PrettyPrint`) and formatter factory
- `config.go` - Configuration system with Language constants, Config struct, and builder methods

**Internal Architecture (`sqlfmt/internal/`)**

- `core/formatter.go` - Core formatting logic and main formatter struct
- `core/tokenizer.go` - SQL tokenization engine that breaks queries into tokens
- `core/config.go` - Internal configuration types and interfaces
- `dialects/` - SQL dialect implementations (standard, postgresql, db2, plsql, n1ql)
- `utils/` - Utility functions (colors, params, indentation, inline_block, dedent)
- `types/` - Type definitions (Token, TokenType, etc.)

**SQL Dialect Formatters**

- `standard_sql_formatter.go` - Default SQL formatter (public API wrapper)
- `n1ql_formatter.go` - Couchbase N1QL dialect
- `db2_formatter.go` - IBM DB2 dialect
- `pl_sql_formatter.go` - Oracle PL/SQL dialect
- `postgresql_formatter.go` - PostgreSQL dialect (Phase 1 & 2 complete)

### Key Patterns

The library uses a factory pattern in `getFormatter()` that selects the appropriate formatter based on the configured Language. All formatters implement the `Formatter` interface with a single `Format(string) string` method.

Configuration uses a fluent builder pattern where methods like `WithLang()`, `WithIndent()` return `*Config` for method chaining.

The tokenizer categorizes SQL elements into types defined in `token_types.go` and supports customization via `TokenizerConfig` for different SQL dialects.

## PostgreSQL Support Implementation

**Status**: PostgreSQL basic support (Phase 1 & 2) has been implemented. Currently supports all standard SQL formatting with PostgreSQL language recognition.

**Implementation Plan**: See `PLAN.md` for the comprehensive 15-phase PostgreSQL implementation roadmap.

**Usage**: Use `sqlfmt.PostgreSQL` as the language parameter to format PostgreSQL queries.

**Testing**: All PostgreSQL formatter tests can be run with `go test ./sqlfmt -run TestPostgreSQL`.

## Test Structure and Organization

The project uses a comprehensive multi-level testing approach combining integration tests, unit tests, and specialized testing patterns.

### Test Organization Overview

**Public API Tests (`sqlfmt/` package)**

- `format_test.go` - Comprehensive table-driven tests for core formatting functionality
- `golden_test.go` - Golden file-based testing using external test data files
- `snapshot_test.go` - Snapshot-based regression testing with automatic test data management
- `tokenizer_test.go` - Public API tokenizer tests that validate tokenization through formatting
- `*_formatter_test.go` - Dialect-specific formatter tests (postgresql, db2, n1ql, plsql)

**Internal Unit Tests (`sqlfmt/internal/` packages)**

- `core/tokenizer_test.go` - Direct unit tests for internal tokenizer functionality
- `utils/dedent_test.go` - Unit tests for utility functions
- Additional internal tests for critical components

**CLI Integration Tests (`cmd/sqlfmt/cmd/`)**

- `format_test.go` - Command-line interface integration tests
- `validate_test.go` - Validation command tests
- `dialects_test.go` - Dialect selection tests

### Testing Approaches

**1. Table-Driven Tests**

- Used extensively in `format_test.go` for systematic testing of formatting scenarios
- Clear separation of input, expected output, and configuration
- Easy to add new test cases and maintain existing ones

**2. Golden File Testing**

- Input files: `testdata/input/{dialect}/`
- Expected output: `testdata/golden/{dialect}/`
- Automatically discovers and tests all `.sql` files
- Excellent for comprehensive dialect testing and regression detection

**3. Snapshot Testing**

- Uses `github.com/gkampitakis/go-snaps` library
- Stores test output in `__snapshots__/` directories
- Perfect for catching unexpected formatting changes
- Update snapshots with: `UPDATE_SNAPS=true go test ./sqlfmt -run TestSnapshot`

**4. Internal Unit Testing**

- Direct testing of internal components without going through public API
- Faster test execution and more targeted debugging
- Tests core functionality like tokenization, formatting logic, and utilities

### Test Execution Commands

**Run All Tests**

```bash
just test                    # Run all tests with verbose output
go test ./...                # Standard Go test execution
```

**Run Specific Test Categories**

```bash
go test ./sqlfmt                           # Public API tests only
go test ./sqlfmt/internal/core             # Internal core tests
go test ./sqlfmt/internal/utils            # Utility tests
go test ./cmd/sqlfmt/cmd                   # CLI tests
```

**Run Tests by Pattern**

```bash
go test ./sqlfmt -run TestFormat           # All formatting tests
go test ./sqlfmt -run TestPostgreSQL       # PostgreSQL-specific tests
go test ./sqlfmt -run TestTokenizer        # Tokenizer tests
go test ./sqlfmt -run TestGolden           # Golden file tests
go test ./sqlfmt -run TestSnapshot         # Snapshot tests
```

**Test Coverage and Benchmarks**

```bash
just test-coverage           # Run tests with coverage report
just test-benchmarks         # Run performance benchmarks
```

### Adding New Tests

**For New Features**: Add tests to the appropriate level

- Integration tests in `format_test.go` for end-to-end functionality
- Unit tests in `internal/` packages for specific components
- Golden files for comprehensive dialect testing

**For Bug Fixes**: Add regression tests at the most appropriate level

- Unit tests for isolated component bugs
- Integration tests for formatting issues
- Snapshot tests to catch unexpected changes

**For New Dialects**: Follow the existing pattern

- Create `{dialect}_formatter_test.go` in main package
- Add golden files in `testdata/input/{dialect}/` and `testdata/golden/{dialect}/`
- Include integration tests in main test suite
