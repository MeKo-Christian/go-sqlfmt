# go-sqlfmt

An SQL formatter written in Go, primarily a **CLI tool** with the library available as **pkg/sqlfmt**.

This project is https://github.com/Snowflake-Labs/snowsql-formatter ported from javascript into Go with some enhancements, like being able to colorize the output.

There is support for [Standard SQL][], [Couchbase N1QL][], [IBM DB2][], [Oracle PL/SQL][], and [PostgreSQL][] dialects.

## Install

### As a Go Library

```shell
go get -u github.com/MeKo-Christian/go-sqlfmt
```

### As a CLI Tool

```shell
go install github.com/MeKo-Christian/go-sqlfmt@latest
```

Or build from source:

```shell
git clone https://github.com/MeKo-Christian/go-sqlfmt.git
cd go-sqlfmt
just build-cli
# or: go build -o bin/sqlfmt .
```

## CLI Usage

The CLI provides powerful formatting capabilities for SQL files and stdin with multiple specialized commands:

### Available Commands

- `sqlfmt format [files...]` - Format SQL files or stdin with configurable options
- `sqlfmt pretty-format [files...]` - Format SQL with ANSI color formatting
- `sqlfmt pretty-print [files...]` - Format and print SQL with colors (stdout only)
- `sqlfmt validate [files...]` - Check if SQL files are properly formatted
- `sqlfmt dialects` - List all supported SQL dialects
- `sqlfmt --help` - Show help and available commands
- `sqlfmt --version` - Show version information

### Basic Examples

```bash
# Format a single file
sqlfmt format query.sql

# Format file in-place
sqlfmt format --write query.sql

# Format from stdin
cat query.sql | sqlfmt format -
echo "select * from users" | sqlfmt format -

# Format with specific dialect
sqlfmt format --lang=postgresql query.sql
sqlfmt format --lang=pl/sql query.sql

# Format with colors for terminal output
sqlfmt format --color query.sql
# or use the dedicated pretty commands:
sqlfmt pretty-format query.sql
sqlfmt pretty-print query.sql

# Format with custom indentation
sqlfmt format --indent="    " query.sql  # 4 spaces
sqlfmt format --indent="\t" query.sql     # tabs
```

### Advanced CLI Usage

```bash
# Format multiple files
sqlfmt format *.sql

# Format with colors and write to file
sqlfmt pretty-format --write query.sql

# Validate formatting (useful for CI)
sqlfmt validate query.sql
sqlfmt validate --lang=postgresql *.sql

# List supported SQL dialects
sqlfmt dialects

# Get help for specific commands
sqlfmt format --help
sqlfmt pretty-format --help
sqlfmt validate --help
```

### CLI Options

| Flag              | Description                                      | Default           | Available In          |
| ----------------- | ------------------------------------------------ | ----------------- | --------------------- |
| `--lang`          | SQL dialect (sql, postgresql, pl/sql, db2, n1ql) | `sql`             | All commands          |
| `--indent`        | Indentation string                               | `"  "` (2 spaces) | All commands          |
| `--write`         | Write result to file instead of stdout           | `false`           | format, pretty-format |
| `--color`         | Enable ANSI color formatting                     | `false`           | format only           |
| `--uppercase`     | Convert keywords to uppercase                    | `false`           | All commands          |
| `--lines-between` | Lines between queries                            | `2`               | All commands          |

**Note**: The `pretty-format` and `pretty-print` commands automatically enable color formatting. Use `pretty-print` when you only want stdout output, and `pretty-format` when you need the `--write` option.

## Library Usage

### Basic Library Usage

```go
package main

import (
    "fmt"
    "github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt"
)

func main() {
    query := `SELECT * FROM foo WHERE goo = 'taco'`
    fmt.Println(sqlfmt.Format(query))
}
```

This will output:

```sql
SELECT
  *
FROM
  foo
WHERE
  goo = 'taco'
```

### Config

You can use the `Config` to specify some formatting options:

```go
sqlfmt.Format(query, sqlfmt.NewDefaultConfig().WithLang(sqlfmt.N1QL))
```

Currently five SQL dialects are supported:

- **sql** - [Standard SQL][]
- **n1ql** - [Couchbase N1QL][]
- **db2** - [IBM DB2][]
- **pl/sql** - [Oracle PL/SQL][]
- **postgresql** - [PostgreSQL][]
- **mysql** - [MySQL][]
- **sqlite** - [SQLite][]

### PostgreSQL

Use the PostgreSQL dialect by setting the language to `sqlfmt.PostgreSQL`:

```go
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.PostgreSQL)
fmt.Println(sqlfmt.Format("SELECT 'a'::text AS casted", cfg))
```

PostgreSQL support includes:

- Dollar-quoted strings: `$$...$$`, `$tag$...$tag$`
- Common operators: type cast `::`, JSON `->`, `->>`, regex `~`, `~*`, `!~`, `!~*`
- PostgreSQL-style line comments: `-- comment`

Notes and current limitations:

- Placeholders: named (`@foo`, `:foo`) and `?` indexed placeholders work. `$1`-style placeholders are planned but not yet supported.
- PL/pgSQL blocks are recognized via dollar-quoting; additional PL/pgSQL formatting improvements are planned.

For detailed PostgreSQL implementation progress and roadmap, see `PLAN-POSTGRESQL.md`.

Run PostgreSQL-focused tests:

```shell
go test ./pkg/sqlfmt -run TestPostgreSQL
```

### SQLite

Use the SQLite dialect by setting the language to `sqlfmt.SQLite`:

```go
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.SQLite)
fmt.Println(sqlfmt.Format("SELECT data->>'name' FROM users WHERE id = ?", cfg))
```

SQLite support includes:

**Comments**: Standard SQL comments `-- comment` and `/* block comment */` are supported.

**Identifier Quoting**: SQLite supports multiple identifier quoting styles for compatibility:
- `"double-quoted"` identifiers (standard SQL)
- `` `backtick-quoted` `` identifiers (MySQL-style)
- `[bracket-quoted]` identifiers (SQL Server-style)

**Placeholders**: All SQLite parameter binding styles are supported with 1-based indexing:
- `?` - Anonymous positional parameter
- `?NNN` - Numbered parameter (e.g., `?1`, `?2`)
- `:name` - Named parameter with colon prefix
- `@name` - Named parameter with at-sign prefix  
- `$name` - Named parameter with dollar prefix

**JSON Operations** (SQLite 3.38+):
- `->` - JSON path extraction
- `->>` - JSON value extraction as text
- Example: `SELECT profile->>'name', settings->'theme' FROM users`

**LIMIT Clauses**: Both SQLite LIMIT syntax styles are supported:
- `LIMIT n OFFSET m` (standard)
- `LIMIT m, n` (MySQL-compatible)

**UPSERT Operations** (SQLite 3.24+):
- `INSERT ... ON CONFLICT(...) DO UPDATE SET ...`
- `INSERT ... ON CONFLICT(...) DO NOTHING`
- `INSERT OR REPLACE INTO ...`

**Advanced Features**:
- Common Table Expressions (CTE) with `WITH` and `WITH RECURSIVE`
- Window functions with `OVER`, `PARTITION BY`, frame specifications
- Generated columns: `GENERATED ALWAYS AS (expr) [VIRTUAL|STORED]`
- Table constraints: `WITHOUT ROWID`, `STRICT` tables
- Triggers: `CREATE TRIGGER ... BEGIN ... END` with proper indentation
- Views: `CREATE VIEW ... AS ...`
- `PRAGMA` statements with minimal formatting

**Pattern Matching**:
- `LIKE` for SQL-standard pattern matching
- `GLOB` for Unix shell-style pattern matching

**NULL Handling**:
- `IS NULL`, `IS NOT NULL` (standard)
- `IS DISTINCT FROM`, `IS NOT DISTINCT FROM` (SQLite 3.39+)

**Concatenation**: String concatenation using `||` operator.

**Blob Literals**: Binary data literals using `X'hexstring'` format.

Notes and current limitations:

- `REGEXP` operator support depends on user-defined functions and is treated as a standard operator
- `PRAGMA` values are preserved as-is without validation or reformatting
- Unicode identifiers are preserved without case coercion inside quotes
- Requires SQLite 3.24+ for UPSERT, 3.25+ for window functions, 3.38+ for JSON operators

Run SQLite-focused tests:

```shell
go test ./pkg/sqlfmt -run TestSQLite
just test-sqlite
```

Config options available are:

- Language (SQL Dialect)
- Indentation
- Lines between queries
- Make reserved words uppercase
- Add parameters
- Add coloring config
- Add tokenizing config

### Colored Output

The CLI provides built-in color support through dedicated commands:

```bash
# Format with colors for terminal viewing
sqlfmt pretty-format query.sql
sqlfmt pretty-print query.sql

# Or use the --color flag with format command
sqlfmt format --color query.sql
```

For library usage, you can format with colors programmatically:

```go
fmt.Println(sqlfmt.PrettyFormat(query))
```

Or use `PrettyPrint` to have it print for you:

```go
sqlfmt.PrettyPrint(query)
```

You can even use a custom coloring config (if you supply a color config, you don't need to use the `Pretty` functions):

```go
clr := sqlfmt.NewDefaultColorConfig()
clr.ReservedWordFormatOptions = []sqlfmt.ANSIFormatOption{
    sqlfmt.ColorBrightGreen, sqlfmt.FormatUnderline,
}
sqlfmt.Format(query, sqlfmt.NewDefaultConfig().WithColorConfig(clr))
```

### Placeholders replacement

#### Named Placeholders

```go
query := "SELECT * FROM tbl WHERE foo = @foo"
sqlfmt.Format(query, sqlfmt.NewDefaultConfig().WithParams(
    sqlfmt.NewMapParams(map[string]string{
        "foo": "'bar'",
    }),
))
```

#### Indexed Placeholders

```go
query := "SELECT * FROM tbl WHERE foo = ?"
sqlfmt.Format(query, sqlfmt.NewDefaultConfig().WithParams(
    sqlfmt.NewListParams([]string{"'bar'"}),
))
```

Both result in:

```sql
SELECT
  *
FROM
  tbl
WHERE
  foo = 'bar'
```

### Tokenizer customization

If for some reason you want things to be tokenized differently, that can be adjusted too.

```go
stdCfg := sqlfmt.NewStandardSQLTokenizerConfig()
stdCfg.ReservedTopLevelWords = append(stdCfg.ReservedTopLevelWords, "BONUS")
sqlfmt.Format(query, sqlfmt.NewDefaultConfig().WithTokenizerConfig(stdCfg))
```

## Testing

This project uses a comprehensive multi-level testing approach to ensure reliability and maintainability.

### Test Structure

**Public API Tests** - Integration testing through the public API:

- Table-driven tests for systematic formatting scenarios
- Golden file testing for comprehensive dialect coverage
- Snapshot testing for regression detection
- Dialect-specific formatter testing

**Internal Unit Tests** - Direct testing of internal components:

- Core tokenization logic testing
- Utility function testing
- Isolated component testing for faster debugging

**CLI Integration Tests** - Command-line interface testing:

- Format command testing with various options
- Pretty format/print command testing
- Validation command testing
- Dialect selection testing

### Running Tests

**Run all tests:**

```bash
just test           # Recommended: runs with verbose output
go test ./...       # Standard Go test execution
```

**Run specific test categories:**

```bash
go test ./pkg/sqlfmt                 # Public API tests
go test ./pkg/sqlfmt/core            # Internal core tests
go test ./pkg/sqlfmt/utils           # Utility tests
go test ./cmd                        # CLI tests
```

**Run tests by pattern:**

```bash
go test ./pkg/sqlfmt -run TestFormat        # Formatting tests
go test ./pkg/sqlfmt -run TestPostgreSQL    # PostgreSQL tests
go test ./pkg/sqlfmt -run TestTokenizer     # Tokenizer tests
just test-golden                            # Golden file tests
just test-snapshots                         # Snapshot tests
```

**Test coverage and performance:**

```bash
just test-coverage      # Generate coverage report
just test-benchmarks    # Run performance benchmarks
```

**Update snapshot tests:**

```bash
just update-snapshots   # Update all snapshots
# or: UPDATE_SNAPS=true go test ./pkg/sqlfmt -run TestSnapshot
```

### Testing Approaches

1. **Table-Driven Tests** - Systematic test cases with clear input/output separation
2. **Golden File Testing** - External test data files for comprehensive dialect testing
3. **Snapshot Testing** - Automatic regression detection with managed test data
4. **Internal Unit Testing** - Fast, targeted testing of individual components

Test data is organized in:

- `testdata/input/{dialect}/` - Input SQL files
- `testdata/golden/{dialect}/` - Expected formatted output
- `__snapshots__/` - Automatic snapshot test data

## Contributing

Create a branch and open a pull request!

### Adding Tests for New Features

When contributing new features:

1. **Add integration tests** in `format_test.go` for end-to-end functionality
2. **Add unit tests** in appropriate packages for isolated components
3. **Add golden files** for new dialects or significant formatting changes
4. **Update snapshots** if output formatting changes: `just update-snapshots`

### Development Commands

```bash
just                    # Build the project
just build-cli          # Build CLI binary to bin/sqlfmt
just install-cli        # Install CLI globally
just lint               # Run linting checks
just lint-fix           # Auto-fix linting issues
just fmt                # Format code
just check              # Run all checks (format, lint, test, tidy)
just setup-deps         # Install development dependencies
```

## Next Steps

- Add a `snowsql` dialect
- Add support for SnowSQL specific keywords and constructs
- Add MySQL dialect support (see `PLAN-MYSQL.md` for implementation roadmap)

## License

[MIT](https://github.com/MeKo-Christian/go-sqlfmt/blob/master/LICENSE)

[standard sql]: https://en.wikipedia.org/wiki/SQL:2011
[couchbase n1ql]: http://www.couchbase.com/n1ql
[ibm db2]: https://www.ibm.com/analytics/us/en/technology/db2/
[oracle pl/sql]: http://www.oracle.com/technetwork/database/features/plsql/index.html
[postgresql]: https://www.postgresql.org/docs/
[mysql]: https://dev.mysql.com/doc/
[sqlite]: https://www.sqlite.org/docs.html
