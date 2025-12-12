# go-sqlfmt

## Project Overview

`go-sqlfmt` is a comprehensive SQL formatter written in Go. It is a port of [snowsql-formatter](https://github.com/Snowflake-Labs/snowsql-formatter) and serves as both a **CLI tool** and a **Go library**.

**Key Features:**
*   **Multi-Dialect Support:** Standard SQL, PostgreSQL, MySQL, SQLite, Oracle PL/SQL, IBM DB2, and Couchbase N1QL.
*   **Auto-Detection:** Infers dialect from file extensions and content.
*   **Colorized Output:** Supports ANSI color formatting for terminal display.
*   **Configurable:** Custom indentation, casing, and rule sets via `.sqlfmt.yaml`.
*   **Zero Dependencies (Runtime):** Lightweight and easy to distribute.

## Building and Running

The project uses `just` as a command runner.

### Key Commands

*   **Build Project:**
    ```bash
    just build      # Builds the module
    just build-cli  # Builds the CLI binary to bin/sqlfmt
    ```

*   **Run CLI (Dev):**
    ```bash
    go run . format testdata/input/standard_sql/basic_select.sql
    ```

*   **Testing:**
    ```bash
    just test               # Run all tests
    just test-coverage      # Run tests with coverage report
    just test-snapshots     # Run snapshot tests only
    just test-golden        # Run golden file tests (input vs expected output)
    ```

*   **Linting & Formatting:**
    ```bash
    just check      # Run all checks (format, lint, test)
    just fmt        # Format code using treefmt
    just lint       # Run golangci-lint
    just lint-fix   # Run linter with auto-fix
    ```

### Dependency Management

*   **Install Tools:** `just setup-deps` (Installs `golangci-lint`, `treefmt`, etc.)
*   **Update Deps:** `just deps` (Runs `go mod tidy`)

## Architecture & Conventions

### Directory Structure

*   `cmd/`: CLI command implementations (Cobra-based).
    *   `root.go`: Entry point and global flags.
    *   `format.go`: The `format` command logic.
*   `pkg/sqlfmt/`: Main library package.
    *   `format.go`: Public API (`Format`, `PrettyFormat`).
    *   `core/`: Core engine (`formatter.go`, `tokenizer.go`).
    *   `dialects/`: Dialect-specific implementations (e.g., `mysql.go`, `postgresql.go`).
*   `testdata/`:
    *   `input/`: Raw SQL files for testing.
    *   `golden/`: Expected formatted output.

### Code Style & Patterns

*   **Dialect Factory:** Uses a factory pattern in `dialects/registry.go` to instantiate the correct formatter based on the `Language` config.
*   **Configuration:** Uses a fluent builder pattern for `Config` (e.g., `NewConfig().WithIndent("  ")`).
*   **Testing Strategy:**
    *   **Golden Files:** Primary integration tests ensuring `input` matches `golden` output.
    *   **Snapshots:** Used for regression testing of internal components.
    *   **Table-Driven:** Standard Go practice for unit tests.

### Contribution Guidelines

*   **New Dialects:** Add a new file in `pkg/sqlfmt/dialects/`, implement the `Formatter` interface, and register it in `registry.go`.
*   **Validation:** Always run `just check` before committing to ensure linting and tests pass.
