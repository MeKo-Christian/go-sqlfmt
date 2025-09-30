# go-sqlfmt

An SQL formatter written in Go, primarily a **CLI tool** with the library available as **pkg/sqlfmt**.

This project is a port of [snowsql-formatter](https://github.com/Snowflake-Labs/snowsql-formatter) from JavaScript into Go with enhancements, including colorized output support.

**Supported SQL Dialects**: [Standard SQL][], [PostgreSQL][], [MySQL][], [SQLite][], [Oracle PL/SQL][], [IBM DB2][], and [Couchbase N1QL][].

## Quick Start

### Installation

**As a CLI tool**:

```shell
go install github.com/MeKo-Christian/go-sqlfmt@latest
```

**As a Go library**:

```shell
go get -u github.com/MeKo-Christian/go-sqlfmt
```

### Basic Usage

**CLI**:

```bash
# Format a SQL file
sqlfmt format query.sql

# Format with colors
sqlfmt pretty-format query.sql

# Format for specific dialect
sqlfmt format --lang=postgresql query.sql

# Use a configuration file for persistent settings
echo "language: postgresql\nkeyword_case: lowercase" > .sqlfmt.yaml
sqlfmt format query.sql  # Uses settings from .sqlfmt.yaml
```

**Library**:

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

Output:

```sql
SELECT
  *
FROM
  foo
WHERE
  goo = 'taco'
```

## Documentation

### Usage Guides

- **[CLI Usage](docs/cli-usage.md)** - Complete CLI commands and options
- **[Library Usage](docs/library-usage.md)** - Go API documentation and examples
- **[Configuration](docs/configuration.md)** - Configuration options and customization

### SQL Dialects

- **[PostgreSQL](docs/dialects/postgresql.md)** - PostgreSQL-specific features
- **[MySQL](docs/dialects/mysql.md)** - MySQL-specific features
- **[SQLite](docs/dialects/sqlite.md)** - SQLite-specific features
- **[Standard SQL](docs/dialects/standard-sql.md)** - ANSI SQL support
- **[Oracle PL/SQL](docs/dialects/plsql.md)** - Oracle PL/SQL support
- **[IBM DB2](docs/dialects/db2.md)** - IBM DB2 dialect support
- **[Couchbase N1QL](docs/dialects/n1ql.md)** - N1QL (SQL for JSON) support

### Development

- **[Testing](docs/testing.md)** - Testing guide and commands
- **[Contributing](docs/contributing.md)** - Development setup and guidelines

## Key Features

- **Multiple SQL Dialects** - Support for 7 major SQL variants
- **Colored Output** - ANSI color formatting for terminal display
- **Configuration Files** - Project-wide and user-wide settings with `.sqlfmt.yaml` files
- **Flexible Configuration** - Customizable indentation, keywords, and formatting rules
- **Parameter Replacement** - Named and indexed placeholder substitution
- **Comprehensive Testing** - Multi-level testing with golden files and snapshots

## License

[MIT](LICENSE)

[standard sql]: https://en.wikipedia.org/wiki/SQL:2011
[couchbase n1ql]: http://www.couchbase.com/n1ql
[ibm db2]: https://www.ibm.com/analytics/us/en/technology/db2/
[oracle pl/sql]: http://www.oracle.com/technetwork/database/features/plsql/index.html
[postgresql]: https://www.postgresql.org/docs/
[mysql]: https://dev.mysql.com/doc/
[sqlite]: https://www.sqlite.org/docs.html
