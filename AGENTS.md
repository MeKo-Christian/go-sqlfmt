# Repository Guidelines

## Project Structure & Module Organization

- Root module: `github.com/MeKo-Christian/go-sqlfmt` (Go 1.23.2).
- Core package lives in `sqlfmt/`:
  - Formatting: `format.go`, `formatter.go`, dialects `standard_sql_formatter.go`, `n1ql_formatter.go`, `db2_formatter.go`, `pl_sql_formatter.go`.
  - Tokenization & helpers: `tokenizer.go`, `token_types.go`, `indentation.go`, `inline_block.go`, `dedent.go`.
  - Config & params: `config.go`, `params.go`, color utilities `colors.go`.
  - Tests: `sqlfmt/*_test.go` cover formatters, tokenizer, and pretty printing.

## Build, Test, and Development Commands

- Init deps: `go mod tidy` — sync and vendor module metadata.
- Run tests: `go test -v ./...` — executes all package tests.
- Benchmarks: `go test -bench=. ./...` — runs formatter/tokenizer benchmarks.
- Lint (local):
  - With Trunk: `trunk check` (auto-fix: `trunk check --fix`).
  - Direct: `golangci-lint run` and `gofmt -l -w .`.

## Coding Style & Naming Conventions

- Formatting: `gofmt`/`go fmt` (tabs, standard Go formatting). CI enforces gofmt and golangci-lint.
- Naming: exported `UpperCamelCase`, unexported `lowerCamelCase`; test files `*_test.go` with `TestXxx` funcs.
- Files: keep dialect-specific logic in `*_formatter.go`; shared utilities in clearly scoped files (e.g., `tokenizer.go`).

## Testing Guidelines

- Frameworks: Go `testing` with `stretchr/testify/require` for assertions.
- Scope: add unit tests for new tokens, rules, or dialect behaviors; include negative cases.
- Conventions: table-driven tests preferred; place under `sqlfmt/` mirroring the code under test.
- Run: `go test -v ./...` (ensure benchmarks still pass or document regressions).

## Commit & Pull Request Guidelines

- Commits: imperative, concise subject (e.g., `sqlfmt: handle BETWEEN ranges`), include rationale in body when changing formatting rules.
- PRs: clear description, linked issues, before/after examples for formatting changes, and tests; update README if API/config changes.
- CI: Go 1.23.2; PRs must pass tests, benchmarks, and lint (`golangci-lint`).

## Security & Configuration Tips

- Do not commit secrets; this repo runs linters (`trufflehog`, `osv-scanner`) via Trunk/CI.
- Keep `.go-version` and `go.mod` aligned (current: 1.23.2).
