# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Just Commands (Primary)

- `just` - Default target (builds the module)
- `just test` - Run all tests with verbose output
- `just test-benchmarks` - Run benchmarks
- `just test-coverage` - Run tests with coverage report
- `just test-snapshots` - Run snapshot tests only
- `just test-golden` - Run golden file tests only
- `just update-snapshots` - Update snapshot test data
- `just lint` - Run golangci-lint with custom config
- `just lint-fix` - Run golangci-lint with automatic fixes
- `just fmt` - Format all code using treefmt
- `just build-cli` - Build the CLI binary to `bin/sqlfmt`
- `just install-cli` - Install the CLI globally
- `just check` - Run all checks (format, lint, test, tidy, generated)
- `just setup-deps` - Install all development tools
- `just help` - Show all available commands

### Direct Go Commands

- `go test ./...` - Run all tests
- `go test -bench=. ./...` - Run benchmarks
- `go test ./pkg/sqlfmt` - Run tests for the main package only
- `go build ./...` - Build the module
- `go build .` - Build the CLI binary
- `go mod tidy` - Clean up module dependencies

### CLI Commands

The project provides a fully-featured CLI tool with the following commands:

- `sqlfmt format [files...]` - Format SQL files or stdin with configurable options
- `sqlfmt pretty-format [files...]` - Format SQL with ANSI color formatting
- `sqlfmt pretty-print [files...]` - Format and print SQL with colors (stdout only)
- `sqlfmt validate [files...]` - Check if SQL files are properly formatted
- `sqlfmt dialects` - List all supported SQL dialects
- `sqlfmt --help` - Show help and available commands
- `sqlfmt --version` - Show version information

## Architecture

This is a Go library that formats SQL queries with support for multiple SQL dialects. The architecture is organized around a common `Formatter` interface with dialect-specific implementations.

### Project Structure

```
go-sqlfmt/
├── main.go                 # CLI entry point
├── cmd/                    # Cobra CLI commands
│   ├── root.go            # Root command and version info
│   ├── format.go          # Format command implementation
│   ├── pretty.go          # Pretty format/print commands
│   ├── validate.go        # Validation command
│   └── dialects.go        # List dialects command
├── pkg/sqlfmt/            # Main library package
│   ├── format.go          # Public API (Format, PrettyFormat, PrettyPrint)
│   ├── config.go          # Configuration system
│   ├── core/              # Internal core functionality
│   │   ├── formatter.go   # Core formatting logic
│   │   ├── tokenizer.go   # SQL tokenization engine
│   │   └── config.go      # Internal configuration types
│   ├── dialects/          # SQL dialect implementations
│   │   ├── registry.go    # Formatter factory and registry
│   │   ├── standard.go    # Standard SQL formatter
│   │   ├── postgresql.go  # PostgreSQL dialect
│   │   ├── db2.go         # IBM DB2 dialect
│   │   ├── plsql.go       # Oracle PL/SQL dialect
│   │   └── n1ql.go        # Couchbase N1QL dialect
│   ├── types/             # Type definitions
│   │   └── token_types.go # Token type constants
│   └── utils/             # Utility functions
│       ├── colors.go      # ANSI color formatting
│       ├── params.go      # Parameter handling
│       ├── indentation.go # Indentation utilities
│       ├── inline_block.go # Inline block handling
│       └── dedent.go      # Text dedentation
└── testdata/              # Test data files
    ├── input/             # Test input files by dialect
    └── golden/            # Expected output files by dialect
```

### Core Components

**Public API (`pkg/sqlfmt/format.go`)**

- `Format(query string, cfg ...*Config) string` - Format SQL query
- `PrettyFormat(query string, cfg ...*Config) string` - Format with ANSI colors
- `PrettyPrint(query string, cfg ...*Config)` - Format with colors and print

**Configuration System (`pkg/sqlfmt/config.go`)**

- Language constants (`StandardSQL`, `PostgreSQL`, `DB2`, `PLSQL`, `N1QL`)
- `Config` struct with fluent builder methods (`WithLang()`, `WithIndent()`, etc.)
- Color configuration and tokenizer customization

**Core Engine (`pkg/sqlfmt/core/`)**

- `formatter.go` - Main formatting logic and query processing
- `tokenizer.go` - SQL tokenization with dialect-specific rules
- `config.go` - Internal configuration interfaces

**Dialect System (`pkg/sqlfmt/dialects/`)**

Each dialect implements the `Formatter` interface with dialect-specific:

- Reserved word lists
- Tokenization rules
- Formatting behavior
- Special syntax handling

### Key Patterns

The library uses a factory pattern in `dialects.CreateFormatterForLanguage()` that selects the appropriate formatter based on the configured Language. All formatters implement the `Formatter` interface with a single `Format(string) string` method.

Configuration uses a fluent builder pattern where methods like `WithLang()`, `WithIndent()` return `*Config` for method chaining.

The tokenizer categorizes SQL elements into types defined in `types/token_types.go` and supports customization via `TokenizerConfig` for different SQL dialects.

## Supported SQL Dialects

- **Standard SQL** (`sql`, `standard`) - ANSI SQL with common formatting rules
- **PostgreSQL** (`postgresql`, `postgres`) - PostgreSQL-specific formatting and keywords (see `PLAN-POSTGRESQL.md` for implementation roadmap)
- **PL/SQL** (`pl/sql`, `plsql`, `oracle`) - Oracle PL/SQL with procedural extensions
- **DB2** (`db2`) - IBM DB2 SQL dialect
- **N1QL** (`n1ql`) - Couchbase N1QL (SQL for JSON)

### Future Dialects

- **MySQL** - MySQL dialect support is planned (see `PLAN-MYSQL.md` for implementation roadmap)

## Test Structure and Organization

The project uses a comprehensive multi-level testing approach combining integration tests, unit tests, and specialized testing patterns.

### Test Organization Overview

**Public API Tests (`pkg/sqlfmt/`)**

- `format_test.go` - Comprehensive table-driven tests for core formatting functionality
- `golden_test.go` - Golden file-based testing using external test data files
- `snapshot_test.go` - Snapshot-based regression testing with automatic test data management
- `tokenizer_test.go` - Public API tokenizer tests that validate tokenization through formatting
- `*_formatter_test.go` - Dialect-specific formatter tests (postgresql, db2, n1ql, plsql)

**Internal Unit Tests (`pkg/sqlfmt/` subdirectories)**

- `core/tokenizer_test.go` - Direct unit tests for internal tokenizer functionality
- `utils/dedent_test.go` - Unit tests for utility functions

**CLI Integration Tests (`cmd/`)**

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
- Update snapshots with: `just update-snapshots`

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
go test ./pkg/sqlfmt                       # Public API tests only
go test ./pkg/sqlfmt/core                  # Internal core tests
go test ./pkg/sqlfmt/utils                 # Utility tests
go test ./cmd                              # CLI tests
```

**Run Tests by Pattern**

```bash
go test ./pkg/sqlfmt -run TestFormat             # All formatting tests
go test ./pkg/sqlfmt -run TestPostgreSQL         # PostgreSQL-specific tests
go test ./pkg/sqlfmt -run TestPostgreSQLFormatter # PostgreSQL formatter tests
go test ./pkg/sqlfmt -run TestTokenizer          # Tokenizer tests
go test ./pkg/sqlfmt -run TestMySQL              # MySQL-specific tests
go test ./pkg/sqlfmt -run TestSQLite             # SQLite-specific tests
just test-golden                                 # Golden file tests
just test-snapshots                              # Snapshot tests
```

**PostgreSQL-Specific Testing**

```bash
# Run comprehensive PostgreSQL tests
go test ./pkg/sqlfmt -run TestPostgreSQL -v

# Run specific PostgreSQL formatter tests
go test ./pkg/sqlfmt -run TestPostgreSQLFormatter -v

# Test PostgreSQL golden files (complex queries)
just test-golden
# Test data locations:
# - Input: testdata/input/postgresql/*.sql
# - Expected: testdata/golden/postgresql/*.sql

# Update PostgreSQL snapshots if output changes
just update-snapshots
# or: UPDATE_SNAPS=true go test ./pkg/sqlfmt -run TestSnapshot

# PostgreSQL feature-specific test patterns
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Dollar"    # Dollar-quoted strings
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Cast"      # Type casting
go test ./pkg/sqlfmt -run "TestPostgreSQL.*JSON"      # JSON operations
go test ./pkg/sqlfmt -run "TestPostgreSQL.*CTE"       # Common Table Expressions
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Window"    # Window functions
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Array"     # Array operations
```

**Test Coverage and Benchmarks**

```bash
just test-coverage           # Run tests with coverage report
just test-benchmarks         # Run performance benchmarks
```

### Adding New Tests

**For New Features**: Add tests to the appropriate level

- Integration tests in `format_test.go` for end-to-end functionality
- Unit tests in `core/` or `utils/` packages for specific components
- Golden files for comprehensive dialect testing

**For Bug Fixes**: Add regression tests at the most appropriate level

- Unit tests for isolated component bugs
- Integration tests for formatting issues
- Snapshot tests to catch unexpected changes

**For New Dialects**: Follow the existing pattern

- Create `{dialect}_formatter_test.go` in main package
- Add golden files in `testdata/input/{dialect}/` and `testdata/golden/{dialect}/`
- Include integration tests in main test suite

# important-instruction-reminders

Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (\*.md) or README files. Only create documentation files if explicitly requested by the User.
