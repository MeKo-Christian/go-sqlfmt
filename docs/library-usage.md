# Library Usage

The go-sqlfmt library provides a simple Go API for formatting SQL queries programmatically.

## Installation

```shell
go get -u github.com/MeKo-Christian/go-sqlfmt
```

## Basic Usage

### Simple Formatting

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

### Using Configuration

You can use the `Config` to specify formatting options:

```go
package main

import (
    "fmt"
    "github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt"
)

func main() {
    query := `SELECT * FROM users WHERE id = $1`

    // Configure for PostgreSQL with custom settings
    cfg := sqlfmt.NewDefaultConfig().
        WithLang(sqlfmt.PostgreSQL).
        WithIndent("    "). // 4 spaces
        WithUppercase(true)

    fmt.Println(sqlfmt.Format(query, cfg))
}
```

## Configuration Options

### Language Dialects

The library supports multiple SQL dialects:

```go
// Available language constants
sqlfmt.StandardSQL  // Standard SQL (ANSI SQL)
sqlfmt.PostgreSQL   // PostgreSQL
sqlfmt.MySQL        // MySQL
sqlfmt.SQLite       // SQLite
sqlfmt.PLSQL        // Oracle PL/SQL
sqlfmt.DB2          // IBM DB2
sqlfmt.N1QL         // Couchbase N1QL
```

### Basic Configuration

```go
cfg := sqlfmt.NewDefaultConfig().
    WithLang(sqlfmt.PostgreSQL).
    WithIndent("\t").               // Use tabs
    WithUppercase(true).            // Uppercase keywords
    WithLinesBetweenQueries(3)      // 3 lines between queries
```

### Complete Configuration Options

- **Language** - SQL dialect for formatting rules
- **Indentation** - String used for indentation (default: 2 spaces)
- **Lines between queries** - Number of lines between separate queries (default: 2)
- **Uppercase keywords** - Convert keywords to uppercase (default: false)
- **Parameters** - Parameter replacement configuration
- **Color config** - ANSI color formatting configuration
- **Tokenizer config** - Custom tokenization rules

## Colored Output

### Pretty Formatting Functions

```go
// Format with default colors
formatted := sqlfmt.PrettyFormat(query)
fmt.Println(formatted)

// Format and print with colors
sqlfmt.PrettyPrint(query)

// With configuration
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.PostgreSQL)
sqlfmt.PrettyPrint(query, cfg)
```

### Custom Color Configuration

```go
// Create custom color configuration
clr := sqlfmt.NewDefaultColorConfig()
clr.ReservedWordFormatOptions = []sqlfmt.ANSIFormatOption{
    sqlfmt.ColorBrightGreen, sqlfmt.FormatUnderline,
}
clr.StringFormatOptions = []sqlfmt.ANSIFormatOption{
    sqlfmt.ColorYellow,
}

// Use custom colors
cfg := sqlfmt.NewDefaultConfig().WithColorConfig(clr)
fmt.Println(sqlfmt.Format(query, cfg))
```

### Available Color Options

```go
// Colors
sqlfmt.ColorBlack
sqlfmt.ColorRed
sqlfmt.ColorGreen
sqlfmt.ColorYellow
sqlfmt.ColorBlue
sqlfmt.ColorMagenta
sqlfmt.ColorCyan
sqlfmt.ColorWhite
sqlfmt.ColorBrightRed
sqlfmt.ColorBrightGreen
// ... etc

// Formatting
sqlfmt.FormatBold
sqlfmt.FormatDim
sqlfmt.FormatItalic
sqlfmt.FormatUnderline
sqlfmt.FormatReverse
```

## Parameter Replacement

### Named Parameters

```go
query := "SELECT * FROM users WHERE name = @name AND status = @status"

params := sqlfmt.NewMapParams(map[string]string{
    "name":   "'John Doe'",
    "status": "'active'",
})

cfg := sqlfmt.NewDefaultConfig().WithParams(params)
result := sqlfmt.Format(query, cfg)
```

### Indexed Parameters

```go
query := "SELECT * FROM users WHERE id = ? AND status = ?"

params := sqlfmt.NewListParams([]string{"1", "'active'"})

cfg := sqlfmt.NewDefaultConfig().WithParams(params)
result := sqlfmt.Format(query, cfg)
```

### PostgreSQL Numbered Parameters

```go
query := "SELECT * FROM users WHERE id = $1 AND status = $2"

params := sqlfmt.NewListParams([]string{"1", "'active'"})

cfg := sqlfmt.NewDefaultConfig().
    WithLang(sqlfmt.PostgreSQL).
    WithParams(params)

result := sqlfmt.Format(query, cfg)
```

Both parameter replacement approaches result in:

```sql
SELECT
  *
FROM
  users
WHERE
  id = 1
  AND status = 'active'
```

## Advanced Usage

### Custom Tokenizer Configuration

If you need custom tokenization behavior, you can modify the tokenizer configuration:

```go
// Get standard SQL tokenizer config as base
tokCfg := sqlfmt.NewStandardSQLTokenizerConfig()

// Add custom reserved words
tokCfg.ReservedTopLevelWords = append(tokCfg.ReservedTopLevelWords, "CUSTOM_KEYWORD")

// Add custom operators
tokCfg.ReservedWords = append(tokCfg.ReservedWords, "CUSTOM_OPERATOR")

// Use custom configuration
cfg := sqlfmt.NewDefaultConfig().WithTokenizerConfig(tokCfg)
result := sqlfmt.Format(query, cfg)
```

### Dialect-Specific Examples

#### PostgreSQL

```go
query := `
    WITH RECURSIVE series(x) AS (
        SELECT 1
        UNION ALL
        SELECT x + 1 FROM series WHERE x < 10
    )
    SELECT data->>'name' as name,
           value::numeric as amount
    FROM users u
    JOIN series s ON u.id = s.x
    WHERE u.created_at >= $1
`

cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.PostgreSQL)
fmt.Println(sqlfmt.Format(query, cfg))
```

#### MySQL

```go
query := `
    SELECT u.name,
           u.profile->>'$.email' as email,
           COUNT(*) as order_count
    FROM users u
    JOIN orders o ON u.id = o.user_id
    WHERE u.created_at BETWEEN ? AND ?
    GROUP BY u.id
    HAVING order_count > 5
`

cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.MySQL)
fmt.Println(sqlfmt.Format(query, cfg))
```

#### SQLite

```go
query := `
    INSERT INTO users (name, email)
    VALUES (?, ?)
    ON CONFLICT(email) DO UPDATE SET
        name = excluded.name,
        updated_at = datetime('now')
    RETURNING id, name
`

cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.SQLite)
fmt.Println(sqlfmt.Format(query, cfg))
```

## Error Handling

The library is designed to be forgiving and will attempt to format even malformed SQL:

```go
query := "SELECT * FROM users WHERE" // incomplete query

// This will still attempt formatting
result := sqlfmt.Format(query)
fmt.Println(result)
```

However, for best results, ensure your SQL is syntactically correct.

## Performance Considerations

- **Caching**: For repeated formatting of similar queries, consider caching configurations
- **Large queries**: The formatter handles large queries efficiently, but very large files may benefit from streaming approaches
- **Memory usage**: The formatter loads the entire query into memory, so consider memory constraints for very large SQL files

## Integration Examples

### HTTP API Integration

```go
func formatSQLHandler(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading body", http.StatusBadRequest)
        return
    }

    dialect := r.URL.Query().Get("dialect")
    lang := sqlfmt.StandardSQL

    switch dialect {
    case "postgresql":
        lang = sqlfmt.PostgreSQL
    case "mysql":
        lang = sqlfmt.MySQL
    // ... other dialects
    }

    cfg := sqlfmt.NewDefaultConfig().WithLang(lang)
    formatted := sqlfmt.Format(string(body), cfg)

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(formatted))
}
```

### File Processing

```go
func formatSQLFile(filename string, outputFile string, dialect sqlfmt.Language) error {
    content, err := os.ReadFile(filename)
    if err != nil {
        return err
    }

    cfg := sqlfmt.NewDefaultConfig().WithLang(dialect)
    formatted := sqlfmt.Format(string(content), cfg)

    return os.WriteFile(outputFile, []byte(formatted), 0644)
}
```

### Batch Processing

```go
func formatSQLFiles(pattern string, dialect sqlfmt.Language) error {
    files, err := filepath.Glob(pattern)
    if err != nil {
        return err
    }

    cfg := sqlfmt.NewDefaultConfig().WithLang(dialect)

    for _, file := range files {
        content, err := os.ReadFile(file)
        if err != nil {
            log.Printf("Error reading %s: %v", file, err)
            continue
        }

        formatted := sqlfmt.Format(string(content), cfg)

        if err := os.WriteFile(file, []byte(formatted), 0644); err != nil {
            log.Printf("Error writing %s: %v", file, err)
        }
    }

    return nil
}
```

## API Reference

### Core Functions

- `Format(query string, cfg ...*Config) string` - Format SQL query
- `PrettyFormat(query string, cfg ...*Config) string` - Format with colors
- `PrettyPrint(query string, cfg ...*Config)` - Format with colors and print

### Configuration Functions

- `NewDefaultConfig() *Config` - Create default configuration
- `NewDefaultColorConfig() *ColorConfig` - Create default color configuration
- `NewMapParams(map[string]string) *Params` - Create named parameter replacements
- `NewListParams([]string) *Params` - Create indexed parameter replacements

### Language Constants

- `StandardSQL`, `PostgreSQL`, `MySQL`, `SQLite`, `PLSQL`, `DB2`, `N1QL`

For complete API documentation, see the [GoDoc](https://pkg.go.dev/github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt).
