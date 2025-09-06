# go-sqlfmt

An SQL formatter written in Go.

This project is https://github.com/Snowflake-Labs/snowsql-formatter ported from javascript into Go with some enhancements, like being able to colorize the output.

There is support for [Standard SQL][], [Couchbase N1QL][], [IBM DB2][], [Oracle PL/SQL][], and [PostgreSQL][] dialects.

## Install

Get the latest version from NPM:

```shell
go get -u github.com/maxrichie5/go-sqlfmt
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/maxrichie5/go-sqlfmt/sqlfmt"
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

## Contributing

Create a branch and open a pull request!

## Next Steps

- Add a `snowsql` dialect
- Add support for SnowSQL specific keywords and constructs

## License

[MIT](https://github.com/maxrichie5/go-sqlfmt/blob/master/LICENSE)

[standard sql]: https://en.wikipedia.org/wiki/SQL:2011
[couchbase n1ql]: http://www.couchbase.com/n1ql
[ibm db2]: https://www.ibm.com/analytics/us/en/technology/db2/
[oracle pl/sql]: http://www.oracle.com/technetwork/database/features/plsql/index.html
[postgresql]: https://www.postgresql.org/docs/
