# Configuration Guide

This guide covers all configuration options available in go-sqlfmt for customizing SQL formatting behavior.

## Quick Start

### Basic Configuration

```go
// Use default configuration
result := sqlfmt.Format(query)

// Create custom configuration
cfg := sqlfmt.NewDefaultConfig().
    WithLang(sqlfmt.PostgreSQL).
    WithIndent("    ").           // 4 spaces
    WithUppercase(true).          // Uppercase keywords
    WithLinesBetweenQueries(3)    // 3 lines between queries

result := sqlfmt.Format(query, cfg)
```

### CLI Configuration

```bash
# Basic options
sqlfmt format --lang=postgresql --indent="    " --uppercase query.sql

# All options
sqlfmt format \
  --lang=mysql \
  --indent="\t" \
  --uppercase \
  --lines-between=3 \
  --write \
  query.sql
```

## Configuration Options

### Language (SQL Dialect)

**Library**:

```go
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.PostgreSQL)
```

**CLI**:

```bash
sqlfmt format --lang=postgresql query.sql
```

**Available Languages**:

- `sqlfmt.StandardSQL` / `--lang=sql` - Standard SQL (ANSI SQL)
- `sqlfmt.PostgreSQL` / `--lang=postgresql` - PostgreSQL dialect
- `sqlfmt.MySQL` / `--lang=mysql` - MySQL dialect
- `sqlfmt.SQLite` / `--lang=sqlite` - SQLite dialect
- `sqlfmt.PLSQL` / `--lang=pl/sql` - Oracle PL/SQL dialect
- `sqlfmt.DB2` / `--lang=db2` - IBM DB2 dialect
- `sqlfmt.N1QL` / `--lang=n1ql` - Couchbase N1QL dialect

**Default**: `StandardSQL` / `sql`

### Indentation

**Library**:

```go
cfg := sqlfmt.NewDefaultConfig().WithIndent("    ")  // 4 spaces
cfg := sqlfmt.NewDefaultConfig().WithIndent("\t")    // Tabs
cfg := sqlfmt.NewDefaultConfig().WithIndent("  ")    // 2 spaces (default)
```

**CLI**:

```bash
sqlfmt format --indent="    " query.sql   # 4 spaces
sqlfmt format --indent="\t" query.sql     # Tabs
```

**Default**: `"  "` (2 spaces)

### Uppercase Keywords

**Library**:

```go
cfg := sqlfmt.NewDefaultConfig().WithUppercase(true)
```

**CLI**:

```bash
sqlfmt format --uppercase query.sql
```

**Effect**:

```sql
-- With uppercase=false (default)
select * from users where active = true;

-- With uppercase=true
SELECT * FROM users WHERE active = TRUE;
```

**Default**: `false`

### Lines Between Queries

**Library**:

```go
cfg := sqlfmt.NewDefaultConfig().WithLinesBetweenQueries(3)
```

**CLI**:

```bash
sqlfmt format --lines-between=3 query.sql
```

**Effect**:

```sql
SELECT * FROM users;


SELECT * FROM posts;  -- 3 empty lines above
```

**Default**: `2`

### Parameter Replacement

Configure parameter substitution for placeholders in SQL queries.

#### Named Parameters

**Library**:

```go
params := sqlfmt.NewMapParams(map[string]string{
    "username": "'john_doe'",
    "status":   "'active'",
    "limit":    "10",
})

cfg := sqlfmt.NewDefaultConfig().WithParams(params)
```

**Input**:

```sql
SELECT * FROM users WHERE username = @username AND status = @status LIMIT @limit;
```

**Output**:

```sql
SELECT
  *
FROM
  users
WHERE
  username = 'john_doe'
  AND status = 'active'
LIMIT
  10;
```

#### Indexed Parameters

**Library**:

```go
params := sqlfmt.NewListParams([]string{
    "'john_doe'",
    "'active'",
    "10",
})

cfg := sqlfmt.NewDefaultConfig().WithParams(params)
```

**Input**:

```sql
SELECT * FROM users WHERE username = ? AND status = ? LIMIT ?;
```

**Output**: Same as named parameters example above.

#### PostgreSQL Numbered Parameters

**Library**:

```go
params := sqlfmt.NewListParams([]string{
    "'john_doe'",
    "'active'",
    "10",
})

cfg := sqlfmt.NewDefaultConfig().
    WithLang(sqlfmt.PostgreSQL).
    WithParams(params)
```

**Input**:

```sql
SELECT * FROM users WHERE username = $1 AND status = $2 LIMIT $3;
```

**Output**: Same formatting with substituted values.

### Color Configuration

Configure ANSI color formatting for terminal output.

#### Using Pretty Functions

**Library**:

```go
// Use default colors
formatted := sqlfmt.PrettyFormat(query)
sqlfmt.PrettyPrint(query)

// With configuration
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.PostgreSQL)
sqlfmt.PrettyPrint(query, cfg)
```

**CLI**:

```bash
sqlfmt pretty-format query.sql
sqlfmt pretty-print query.sql
sqlfmt format --color query.sql
```

#### Custom Color Configuration

**Library**:

```go
// Create custom color configuration
clr := sqlfmt.NewDefaultColorConfig()

// Customize reserved word colors
clr.ReservedWordFormatOptions = []sqlfmt.ANSIFormatOption{
    sqlfmt.ColorBrightGreen,
    sqlfmt.FormatBold,
}

// Customize string colors
clr.StringFormatOptions = []sqlfmt.ANSIFormatOption{
    sqlfmt.ColorYellow,
}

// Customize number colors
clr.NumberFormatOptions = []sqlfmt.ANSIFormatOption{
    sqlfmt.ColorCyan,
}

// Use custom colors
cfg := sqlfmt.NewDefaultConfig().WithColorConfig(clr)
result := sqlfmt.Format(query, cfg)
```

#### Available Colors

**Basic Colors**:

- `sqlfmt.ColorBlack`
- `sqlfmt.ColorRed`
- `sqlfmt.ColorGreen`
- `sqlfmt.ColorYellow`
- `sqlfmt.ColorBlue`
- `sqlfmt.ColorMagenta`
- `sqlfmt.ColorCyan`
- `sqlfmt.ColorWhite`

**Bright Colors**:

- `sqlfmt.ColorBrightRed`
- `sqlfmt.ColorBrightGreen`
- `sqlfmt.ColorBrightYellow`
- `sqlfmt.ColorBrightBlue`
- `sqlfmt.ColorBrightMagenta`
- `sqlfmt.ColorBrightCyan`
- `sqlfmt.ColorBrightWhite`

**Formatting Options**:

- `sqlfmt.FormatBold`
- `sqlfmt.FormatDim`
- `sqlfmt.FormatItalic`
- `sqlfmt.FormatUnderline`
- `sqlfmt.FormatReverse`

### Tokenizer Configuration

Advanced configuration for customizing SQL tokenization behavior.

#### Basic Tokenizer Customization

**Library**:

```go
// Get standard SQL tokenizer config
tokCfg := sqlfmt.NewStandardSQLTokenizerConfig()

// Add custom reserved words
tokCfg.ReservedWords = append(tokCfg.ReservedWords, "CUSTOM_KEYWORD")

// Add custom top-level words (cause new lines)
tokCfg.ReservedTopLevelWords = append(tokCfg.ReservedTopLevelWords, "CUSTOM_CLAUSE")

// Use custom tokenizer config
cfg := sqlfmt.NewDefaultConfig().WithTokenizerConfig(tokCfg)
```

#### Dialect-Specific Tokenizer Configs

```go
// PostgreSQL tokenizer
pgCfg := sqlfmt.NewPostgreSQLTokenizerConfig()

// MySQL tokenizer
mysqlCfg := sqlfmt.NewMySQLTokenizerConfig()

// Use specific dialect tokenizer
cfg := sqlfmt.NewDefaultConfig().WithTokenizerConfig(pgCfg)
```

#### Tokenizer Configuration Fields

**TokenizerConfig struct**:

```go
type TokenizerConfig struct {
    ReservedWords                 []string // All reserved words
    ReservedTopLevelWords         []string // Words that start new sections
    ReservedNewlineWords          []string // Words that trigger new lines
    ReservedTopLevelWordsNoIndent []string // Top-level words without indentation
    StringTypes                   []string // String literal patterns
    OpenParens                    []string // Opening parentheses
    CloseParens                   []string // Closing parentheses
    IndexedPlaceholderTypes       []string // Indexed placeholder prefixes
    NamedPlaceholderTypes         []string // Named placeholder prefixes
    LineCommentTypes              []string // Line comment prefixes
}
```

**Example Customization**:

```go
tokCfg := &sqlfmt.TokenizerConfig{
    ReservedWords: []string{
        "SELECT", "FROM", "WHERE", "CUSTOM_FUNCTION",
    },
    ReservedTopLevelWords: []string{
        "SELECT", "FROM", "WHERE", "CUSTOM_CLAUSE",
    },
    StringTypes: []string{
        "''",    // Single quotes
        "\"\"",  // Double quotes
        "$$",    // Dollar quotes (PostgreSQL)
    },
    IndexedPlaceholderTypes: []string{"$"},      // $1, $2, etc.
    NamedPlaceholderTypes:   []string{"@", ":"}, // @name, :name
    LineCommentTypes:        []string{"--", "#"}, // -- comment, # comment
}

cfg := sqlfmt.NewDefaultConfig().WithTokenizerConfig(tokCfg)
```

## Configuration Patterns

### Fluent Configuration

Chain configuration methods for readable setup:

```go
cfg := sqlfmt.NewDefaultConfig().
    WithLang(sqlfmt.PostgreSQL).
    WithIndent("    ").
    WithUppercase(true).
    WithLinesBetweenQueries(1).
    WithParams(sqlfmt.NewMapParams(map[string]string{
        "table": "users",
        "limit": "50",
    })).
    WithColorConfig(sqlfmt.NewDefaultColorConfig())

result := sqlfmt.Format(query, cfg)
```

### Configuration Reuse

Save and reuse configurations for consistency:

```go
// Create shared configuration
postgresConfig := sqlfmt.NewDefaultConfig().
    WithLang(sqlfmt.PostgreSQL).
    WithIndent("  ").
    WithUppercase(false)

// Use for multiple queries
query1Result := sqlfmt.Format(query1, postgresConfig)
query2Result := sqlfmt.Format(query2, postgresConfig)
```

### Environment-Specific Configuration

```go
func getConfig(env string) *sqlfmt.Config {
    cfg := sqlfmt.NewDefaultConfig()

    switch env {
    case "development":
        return cfg.WithLang(sqlfmt.PostgreSQL).WithIndent("  ")
    case "production":
        return cfg.WithLang(sqlfmt.PostgreSQL).WithIndent("\t").WithUppercase(true)
    default:
        return cfg
    }
}

result := sqlfmt.Format(query, getConfig("development"))
```

## Dialect-Specific Configuration

### PostgreSQL Configuration

```go
cfg := sqlfmt.NewDefaultConfig().
    WithLang(sqlfmt.PostgreSQL).
    WithParams(sqlfmt.NewListParams([]string{"1", "'active'"}))

// Supports PostgreSQL features:
// - Dollar-quoted strings ($$...$$)
// - JSON operators (->, ->>)
// - Type casting (::)
// - Numbered placeholders ($1, $2)
```

### MySQL Configuration

```go
cfg := sqlfmt.NewDefaultConfig().
    WithLang(sqlfmt.MySQL).
    WithParams(sqlfmt.NewListParams([]string{"1", "'active'"}))

// Supports MySQL features:
// - Backtick identifiers (`table`)
// - JSON operators (->, ->>)
// - MySQL-specific functions
// - Positional placeholders (?)
```

### SQLite Configuration

```go
cfg := sqlfmt.NewDefaultConfig().
    WithLang(sqlfmt.SQLite).
    WithParams(sqlfmt.NewMapParams(map[string]string{
        "id": "1",
        "status": "'active'",
    }))

// Supports SQLite features:
// - Multiple identifier quoting styles
// - All parameter styles (?, :name, @name, $name)
// - JSON operators (SQLite 3.38+)
// - UPSERT operations
```

## Advanced Configuration

### Custom Formatter Implementation

For advanced use cases, you can implement custom formatting behavior:

```go
type CustomConfig struct {
    *sqlfmt.Config
    CustomOption string
}

func (c *CustomConfig) Format(query string) string {
    // Apply custom preprocessing
    processedQuery := customPreprocess(query, c.CustomOption)

    // Use standard formatting
    return sqlfmt.Format(processedQuery, c.Config)
}

func customPreprocess(query, option string) string {
    // Custom logic here
    return query
}
```

### Configuration Validation

```go
func validateConfig(cfg *sqlfmt.Config) error {
    if cfg == nil {
        return fmt.Errorf("configuration cannot be nil")
    }

    // Validate indentation
    if len(cfg.Indent) > 8 {
        return fmt.Errorf("indentation too large: %d characters", len(cfg.Indent))
    }

    // Validate lines between queries
    if cfg.LinesBetweenQueries < 0 || cfg.LinesBetweenQueries > 10 {
        return fmt.Errorf("invalid lines between queries: %d", cfg.LinesBetweenQueries)
    }

    return nil
}

// Use validation
cfg := sqlfmt.NewDefaultConfig().WithIndent("        ") // 8 spaces
if err := validateConfig(cfg); err != nil {
    log.Fatal(err)
}
```

## Configuration Best Practices

1. **Use appropriate dialects**: Choose the correct SQL dialect for your database
2. **Consistent indentation**: Stick to one indentation style across your project
3. **Reasonable line spacing**: Keep lines between queries reasonable (1-3)
4. **Validate parameters**: Ensure parameter values are properly quoted
5. **Reuse configurations**: Create shared configurations for consistency
6. **Test with real queries**: Validate configuration with your actual SQL queries

## CLI Configuration File

While go-sqlfmt doesn't currently support configuration files, you can create shell aliases or scripts:

```bash
# ~/.bashrc or ~/.zshrc
alias sqlfmt-pg='sqlfmt format --lang=postgresql --indent="  " --uppercase'
alias sqlfmt-mysql='sqlfmt format --lang=mysql --indent="\t"'

# Usage
sqlfmt-pg query.sql
sqlfmt-mysql query.sql
```

Or create a wrapper script:

```bash
#!/bin/bash
# sqlfmt-wrapper.sh

LANG=${SQLFMT_LANG:-sql}
INDENT=${SQLFMT_INDENT:-"  "}
UPPERCASE=${SQLFMT_UPPERCASE:-false}

sqlfmt format \
  --lang="$LANG" \
  --indent="$INDENT" \
  $([ "$UPPERCASE" = "true" ] && echo "--uppercase") \
  "$@"
```

```bash
# Environment-based configuration
export SQLFMT_LANG=postgresql
export SQLFMT_INDENT="    "
export SQLFMT_UPPERCASE=true

./sqlfmt-wrapper.sh query.sql
```

This comprehensive configuration system allows you to customize go-sqlfmt for your specific needs and SQL dialects.
