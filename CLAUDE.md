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
- `formatter.go` - Core formatting logic shared across dialects
- `tokenizer.go` - SQL tokenization engine that breaks queries into tokens

**SQL Dialect Formatters**

- `standard_sql_formatter.go` - Default SQL formatter
- `n1ql_formatter.go` - Couchbase N1QL dialect
- `db2_formatter.go` - IBM DB2 dialect
- `pl_sql_formatter.go` - Oracle PL/SQL dialect
- `postgresql_formatter.go` - PostgreSQL dialect (Phase 1 & 2 complete)

**Supporting Systems**

- `colors.go` - ANSI color formatting for terminal output
- `params.go` - Parameter replacement (named/indexed placeholders)
- `indentation.go` - Indentation management
- `inline_block.go` - Inline block detection
- `dedent.go` - Dedentation utilities

### Key Patterns

The library uses a factory pattern in `getFormatter()` that selects the appropriate formatter based on the configured Language. All formatters implement the `Formatter` interface with a single `Format(string) string` method.

Configuration uses a fluent builder pattern where methods like `WithLang()`, `WithIndent()` return `*Config` for method chaining.

The tokenizer categorizes SQL elements into types defined in `token_types.go` and supports customization via `TokenizerConfig` for different SQL dialects.

## PostgreSQL Support Implementation

**Status**: PostgreSQL basic support (Phase 1 & 2) has been implemented. Currently supports all standard SQL formatting with PostgreSQL language recognition.

**Implementation Plan**: See `PLAN.md` for the comprehensive 15-phase PostgreSQL implementation roadmap.

**Usage**: Use `sqlfmt.PostgreSQL` as the language parameter to format PostgreSQL queries.

**Testing**: All PostgreSQL formatter tests can be run with `go test ./sqlfmt -run TestPostgreSQL`.
