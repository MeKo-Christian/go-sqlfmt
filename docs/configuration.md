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

## Configuration Files

go-sqlfmt supports configuration files for persistent settings across your project or user environment. Configuration files use YAML format and are automatically discovered and loaded by the CLI.

### Supported File Names

The following configuration file names are recognized (in order of precedence):

- `.sqlfmtrc`
- `.sqlfmt.yaml`
- `.sqlfmt.yml`
- `sqlfmt.yaml`
- `sqlfmt.yml`

### Search Order

go-sqlfmt searches for configuration files in the following order:

1. **Current directory and parent directories** - Searches upward from the current working directory until reaching the git root (if in a git repository)
2. **User home directory** - Falls back to `~/.sqlfmtrc`, `~/.sqlfmt.yaml`, or `~/.sqlfmt.yml`

The first configuration file found is used. This allows for project-specific configurations that override user-wide defaults.

### Configuration Precedence

Settings are applied in the following order (later sources override earlier ones):

1. **Default values** - Built-in defaults
2. **Configuration file** - Settings from discovered config file
3. **CLI flags** - Command-line arguments (highest priority)

### Configuration Options

All configuration options can be specified in YAML format:

```yaml
# SQL dialect to use by default
language: postgresql

# Keyword casing options
# Options: preserve, uppercase, lowercase, dialect
keyword_case: lowercase

# Indentation string (spaces or tabs)
indent: "    "

# Number of lines between separate queries
lines_between_queries: 1
```

#### Available Options

**`language`** (string)

The SQL dialect to use for formatting. Valid values:

- `sql` or `standard` - Standard SQL (ANSI SQL)
- `postgresql` or `postgres` - PostgreSQL dialect
- `mysql` or `mariadb` - MySQL dialect
- `sqlite` - SQLite dialect
- `pl/sql`, `plsql`, or `oracle` - Oracle PL/SQL dialect
- `db2` - IBM DB2 dialect
- `n1ql` - Couchbase N1QL dialect

**`keyword_case`** (string)

How to format SQL keywords. Valid values:

- `preserve` - Keep original casing (default)
- `uppercase` - Convert to UPPERCASE
- `lowercase` - Convert to lowercase
- `dialect` - Use dialect-specific casing

**`indent`** (string)

The indentation string to use. Examples:

- `"  "` - 2 spaces (default)
- `"    "` - 4 spaces
- `"\t"` - Tab character

**`lines_between_queries`** (integer)

Number of blank lines to insert between separate SQL queries. Default: `2`

### Example Configuration Files

#### Project-Specific Configuration

Place a `.sqlfmt.yaml` file in your project root:

```yaml
# .sqlfmt.yaml - PostgreSQL project
language: postgresql
keyword_case: lowercase
indent: "  "
lines_between_queries: 1
```

```yaml
# .sqlfmt.yaml - MySQL project
language: mysql
keyword_case: uppercase
indent: "    "
lines_between_queries: 2
```

#### User-Wide Configuration

Place a `.sqlfmtrc` file in your home directory for personal defaults:

```yaml
# ~/.sqlfmtrc
language: sql
keyword_case: preserve
indent: "  "
lines_between_queries: 2
```

### Using Configuration Files with CLI

When you run CLI commands, the configuration file is automatically loaded:

```bash
# Uses settings from .sqlfmt.yaml if present
sqlfmt format query.sql

# CLI flags override config file settings
sqlfmt format --lang=mysql query.sql

# Still uses indent and keyword_case from config file,
# but overrides language to mysql
```

### Verifying Configuration

To see which configuration is being used, you can check the formatted output or use CLI flags to override specific settings:

```bash
# Format with config file settings
sqlfmt format query.sql

# Override specific settings
sqlfmt format --keyword-case=uppercase query.sql
```

### Multi-Dialect Projects

For projects using multiple SQL dialects, you can:

1. Create a base configuration in your home directory
2. Override with project-specific settings in subdirectories
3. Use CLI flags for file-specific formatting

Example project structure:

```
myproject/
├── .sqlfmt.yaml           # Default: PostgreSQL
├── mysql/
│   ├── .sqlfmt.yaml       # Override: MySQL
│   └── schema.sql
└── postgresql/
    └── migrations.sql     # Uses root config
```

### Shell Aliases (Alternative Approach)

If you prefer shell aliases over configuration files:

```bash
# ~/.bashrc or ~/.zshrc
alias sqlfmt-pg='sqlfmt format --lang=postgresql --indent="  " --keyword-case=lowercase'
alias sqlfmt-mysql='sqlfmt format --lang=mysql --indent="\t"'

# Usage
sqlfmt-pg query.sql
sqlfmt-mysql query.sql
```

This comprehensive configuration system allows you to customize go-sqlfmt for your specific needs and SQL dialects.
