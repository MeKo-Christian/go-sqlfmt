# go-sqlfmt

An SQL formatter written in Go, available as both a **library** and **CLI tool**.

This project is https://github.com/Snowflake-Labs/snowsql-formatter ported from javascript into Go with some enhancements, like being able to colorize the output.

There is support for [Standard SQL][], [Couchbase N1QL][], [IBM DB2][], [Oracle PL/SQL][], and [PostgreSQL][] dialects.

## Install

### As a Go Library

```shell
go get -u github.com/MeKo-Christian/go-sqlfmt
```

### As a CLI Tool

```shell
go install github.com/MeKo-Christian/go-sqlfmt/cmd/sqlfmt@latest
```

Or build from source:

```shell
git clone https://github.com/MeKo-Christian/go-sqlfmt.git
cd go-sqlfmt
just build-cli
# or: go build -o sqlfmt ./cmd/sqlfmt
```

## Usage

### CLI Usage

The CLI provides powerful formatting capabilities for SQL files and stdin:

#### Basic Examples

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

# Format with custom indentation
sqlfmt format --indent="    " query.sql  # 4 spaces
sqlfmt format --indent="\t" query.sql     # tabs
```

#### Advanced CLI Usage

```bash
# Format multiple files
sqlfmt format *.sql

# Validate formatting (useful for CI)
sqlfmt validate query.sql
sqlfmt validate --lang=postgresql *.sql

# List supported SQL dialects
sqlfmt dialects

# Get help
sqlfmt --help
sqlfmt format --help
```

#### CLI Options

| Flag              | Description                                      | Default           |
| ----------------- | ------------------------------------------------ | ----------------- |
| `--lang`          | SQL dialect (sql, postgresql, pl/sql, db2, n1ql) | `sql`             |
| `--indent`        | Indentation string                               | `"  "` (2 spaces) |
| `--write`         | Write result to file instead of stdout           | `false`           |
| `--color`         | Enable ANSI color formatting                     | `false`           |
| `--uppercase`     | Convert keywords to uppercase                    | `false`           |
| `--lines-between` | Lines between queries                            | `2`               |

### Library Usage

#### Basic Library Usage

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

Run PostgreSQL-focused tests:

```shell
go test ./sqlfmt -run TestPostgreSQL
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

You can also format with color:

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
go test ./sqlfmt                    # Public API tests
go test ./sqlfmt/internal/core      # Internal core tests
go test ./sqlfmt/internal/utils     # Utility tests
go test ./cmd/sqlfmt/cmd            # CLI tests
```

**Run tests by pattern:**

```bash
go test ./sqlfmt -run TestFormat        # Formatting tests
go test ./sqlfmt -run TestPostgreSQL    # PostgreSQL tests
go test ./sqlfmt -run TestTokenizer     # Tokenizer tests
go test ./sqlfmt -run TestGolden        # Golden file tests
go test ./sqlfmt -run TestSnapshot      # Snapshot tests
```

**Test coverage and performance:**

```bash
just test-coverage      # Generate coverage report
just test-benchmarks    # Run performance benchmarks
```

**Update snapshot tests:**

```bash
UPDATE_SNAPS=true go test ./sqlfmt -run TestSnapshot
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
2. **Add unit tests** in appropriate `internal/` packages for isolated components
3. **Add golden files** for new dialects or significant formatting changes
4. **Update snapshots** if output formatting changes: `UPDATE_SNAPS=true go test ./sqlfmt -run TestSnapshot`

### Development Commands

```bash
just                    # Build the project
just lint              # Run linting checks
just lint-fix          # Auto-fix linting issues
just fmt               # Format code
just check             # Run all checks (format, lint, test, tidy)
just setup-deps        # Install development dependencies
```

## Next Steps

- Add a `snowsql` dialect
- Add support for SnowSQL specific keywords and constructs

## License

[MIT](https://github.com/MeKo-Christian/go-sqlfmt/blob/master/LICENSE)

[standard sql]: https://en.wikipedia.org/wiki/SQL:2011
[couchbase n1ql]: http://www.couchbase.com/n1ql
[ibm db2]: https://www.ibm.com/analytics/us/en/technology/db2/
[oracle pl/sql]: http://www.oracle.com/technetwork/database/features/plsql/index.html
[postgresql]: https://www.postgresql.org/docs/
