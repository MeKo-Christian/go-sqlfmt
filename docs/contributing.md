# Contributing Guide

Thank you for your interest in contributing to go-sqlfmt! This guide will help you set up your development environment and understand the contribution process.

## Getting Started

### Prerequisites

- Go 1.19 or higher
- Git
- [Just](https://github.com/casey/just) command runner (recommended)

### Setting up the Development Environment

1. **Fork and clone the repository**:

   ```bash
   git clone https://github.com/your-username/go-sqlfmt.git
   cd go-sqlfmt
   ```

2. **Install development dependencies**:

   ```bash
   just setup-deps
   ```

3. **Verify the setup**:
   ```bash
   just check
   ```

## Development Commands

### Just Commands (Recommended)

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

### CLI Testing Commands

The project provides a fully-featured CLI tool. You can test your changes using:

- `sqlfmt format [files...]` - Format SQL files or stdin with configurable options
- `sqlfmt pretty-format [files...]` - Format SQL with ANSI color formatting
- `sqlfmt pretty-print [files...]` - Format and print SQL with colors (stdout only)
- `sqlfmt validate [files...]` - Check if SQL files are properly formatted
- `sqlfmt dialects` - List all supported SQL dialects
- `sqlfmt --help` - Show help and available commands
- `sqlfmt --version` - Show version information

## Project Architecture

### Overview

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
│   │   ├── mysql.go       # MySQL dialect
│   │   ├── sqlite.go      # SQLite dialect
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

### Key Components

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

## Contributing Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-description
```

### 2. Make Changes

- Write your code following the existing patterns
- Add or update tests for your changes
- Ensure code is properly formatted and linted

### 3. Test Your Changes

```bash
# Run all checks
just check

# Run specific tests
just test
just lint
just fmt

# Test your changes with the CLI
just build-cli
./bin/sqlfmt format your-test.sql
```

### 4. Update Documentation

- Update relevant documentation files in `docs/`
- Add examples for new features
- Update CLAUDE.md if adding new test commands

### 5. Commit and Push

```bash
git add .
git commit -m "feat: add new feature"
git push origin feature/your-feature-name
```

### 6. Create Pull Request

- Open a pull request on GitHub
- Provide a clear description of your changes
- Link any related issues
- Ensure all CI checks pass

## Adding Tests for New Features

### Integration Tests

When contributing new features:

1. **Add integration tests** in `format_test.go` for end-to-end functionality
2. **Add unit tests** in appropriate packages for isolated components
3. **Add golden files** for new dialects or significant formatting changes
4. **Update snapshots** if output formatting changes: `just update-snapshots`

### Test Structure

The project uses multiple testing approaches:

1. **Table-Driven Tests** - Systematic test cases with clear input/output separation
2. **Golden File Testing** - External test data files for comprehensive dialect testing
3. **Snapshot Testing** - Automatic regression detection with managed test data
4. **Internal Unit Testing** - Fast, targeted testing of individual components

### Adding Tests

For new features:

```bash
# Add test cases to appropriate _test.go files
# Add golden files in testdata/input/{dialect}/ and testdata/golden/{dialect}/
# Run tests
just test

# Update snapshots if needed
just update-snapshots
```

## SQL Dialect Support

### Supported Dialects

- **Standard SQL** (`sql`, `standard`) - ANSI SQL with common formatting rules
- **PostgreSQL** (`postgresql`, `postgres`) - PostgreSQL-specific formatting and keywords
- **MySQL** (`mysql`) - MySQL-specific features and syntax
- **SQLite** (`sqlite`) - SQLite-specific features and syntax
- **PL/SQL** (`pl/sql`, `plsql`, `oracle`) - Oracle PL/SQL with procedural extensions
- **DB2** (`db2`) - IBM DB2 SQL dialect
- **N1QL** (`n1ql`) - Couchbase N1QL (SQL for JSON)

### Adding a New Dialect

1. **Create the dialect file**: `pkg/sqlfmt/dialects/{dialect}.go`
2. **Implement the Formatter interface**:

   ```go
   type YourDialectFormatter struct {
       cfg *Config
   }

   func (f *YourDialectFormatter) Format(query string) string {
       return core.FormatQuery(f.cfg, f.tokenOverride, query)
   }
   ```

3. **Create tokenizer configuration**: Define reserved words, operators, etc.
4. **Add to factory**: Update `CreateFormatterForLanguage()` in registry
5. **Add language constant**: Update `config.go` with the new language
6. **Add tests**: Create comprehensive test coverage
7. **Add documentation**: Create `docs/dialects/{dialect}.md`

## Code Style Guidelines

### General Principles

1. **Follow existing patterns**: Maintain consistency with existing code
2. **Write clear, readable code**: Prefer clarity over cleverness
3. **Add comments for complex logic**: Especially in tokenizer and formatter code
4. **Use meaningful variable names**: Make code self-documenting

### Go-Specific Guidelines

1. **Follow standard Go formatting**: Use `gofmt` (handled by `just fmt`)
2. **Handle errors properly**: Always check and handle errors appropriately
3. **Write idiomatic Go**: Use Go conventions and patterns
4. **Add godoc comments**: Document exported functions and types

### Testing Guidelines

1. **Write comprehensive tests**: Cover happy paths and edge cases
2. **Use table-driven tests**: For systematic test coverage
3. **Test at appropriate levels**: Unit tests for components, integration tests for features
4. **Keep tests fast**: Avoid slow-running tests where possible

## Performance Considerations

### Optimization Guidelines

1. **Profile before optimizing**: Use Go's built-in profiling tools
2. **Test with large queries**: Ensure formatter handles large SQL files efficiently
3. **Memory usage**: Be mindful of memory allocations in hot paths
4. **Benchmark critical paths**: Add benchmarks for performance-critical code

### Running Benchmarks

```bash
# Run all benchmarks
just test-benchmarks

# Run specific benchmarks
go test -bench=BenchmarkFormat ./pkg/sqlfmt
go test -bench=BenchmarkTokenizer ./pkg/sqlfmt/core
```

## Release Process

### Version Management

The project follows semantic versioning (semver):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Pre-release Checklist

1. **All tests pass**: `just check`
2. **Documentation updated**: Update relevant docs
3. **Changelog updated**: Document changes
4. **Performance verified**: Run benchmarks
5. **Examples work**: Test documented examples

## Getting Help

### Community

- **GitHub Issues**: Report bugs or request features
- **GitHub Discussions**: Ask questions or discuss ideas
- **Pull Requests**: Contribute code changes

### Development Questions

When asking for help:

1. **Provide context**: What are you trying to achieve?
2. **Share code**: Include relevant code snippets
3. **Include errors**: Share any error messages
4. **Describe expected behavior**: What should happen?

### Code Review Process

Pull requests will be reviewed for:

1. **Functionality**: Does it work as intended?
2. **Testing**: Are there adequate tests?
3. **Code quality**: Is the code clear and maintainable?
4. **Performance**: Any performance implications?
5. **Documentation**: Is it properly documented?

## License

By contributing to go-sqlfmt, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

Thank you for contributing to go-sqlfmt! Your contributions help make SQL formatting better for everyone.
