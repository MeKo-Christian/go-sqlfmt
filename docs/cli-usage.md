# CLI Usage

The CLI provides powerful formatting capabilities for SQL files and stdin with multiple specialized commands.

## Available Commands

- `sqlfmt format [files...]` - Format SQL files or stdin with configurable options
- `sqlfmt pretty-format [files...]` - Format SQL with ANSI color formatting
- `sqlfmt pretty-print [files...]` - Format and print SQL with colors (stdout only)
- `sqlfmt validate [files...]` - Check if SQL files are properly formatted
- `sqlfmt dialects` - List all supported SQL dialects
- `sqlfmt --help` - Show help and available commands
- `sqlfmt --version` - Show version information

## Basic Examples

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
sqlfmt format --lang=mysql query.sql
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

## Advanced CLI Usage

```bash
# Format multiple files
sqlfmt format *.sql

# Format with colors and write to file
sqlfmt pretty-format --write query.sql

# Validate formatting (useful for CI)
sqlfmt validate query.sql
sqlfmt validate --lang=postgresql *.sql
sqlfmt validate --lang=mysql *.sql

# List supported SQL dialects
sqlfmt dialects

# Get help for specific commands
sqlfmt format --help
sqlfmt pretty-format --help
sqlfmt validate --help
```

## CLI Options

| Flag              | Description                                                     | Default           | Available In          |
| ----------------- | --------------------------------------------------------------- | ----------------- | --------------------- |
| `--lang`          | SQL dialect (sql, postgresql, mysql, pl/sql, db2, n1ql, sqlite) | `sql`             | All commands          |
| `--indent`        | Indentation string                                              | `"  "` (2 spaces) | All commands          |
| `--write`         | Write result to file instead of stdout                          | `false`           | format, pretty-format |
| `--color`         | Enable ANSI color formatting                                    | `false`           | format only           |
| `--uppercase`     | Convert keywords to uppercase                                   | `false`           | All commands          |
| `--lines-between` | Lines between queries                                           | `2`               | All commands          |

**Note**: The `pretty-format` and `pretty-print` commands automatically enable color formatting. Use `pretty-print` when you only want stdout output, and `pretty-format` when you need the `--write` option.

## Configuration Files

sqlfmt supports configuration files to set default options without specifying them on every command. This is especially useful for maintaining consistent formatting across a project or for personal preferences.

### Creating a Configuration File

Configuration files use YAML format and can be named:

- `.sqlfmtrc`
- `.sqlfmt.yaml`
- `.sqlfmt.yml`
- `sqlfmt.yaml`
- `sqlfmt.yml`

### Project-Specific Configuration

Place a `.sqlfmt.yaml` file in your project root to set project-wide defaults:

```yaml
# .sqlfmt.yaml
language: postgresql
keyword_case: lowercase
indent: "  "
lines_between_queries: 1
```

Now all commands in this directory will use these settings:

```bash
# Uses postgresql dialect from config
sqlfmt format query.sql

# Still uses config settings, but overrides language
sqlfmt format --lang=mysql query.sql
```

### User-Wide Configuration

Place a `.sqlfmtrc` in your home directory for personal defaults:

```yaml
# ~/.sqlfmtrc
language: sql
keyword_case: preserve
indent: "    "
lines_between_queries: 2
```

### Configuration File Search

sqlfmt searches for configuration files in this order:

1. Current directory (and parent directories up to git root)
2. User home directory

The first config file found is used. CLI flags always override config file settings.

### Example Workflows

**PostgreSQL project with team standards:**

```bash
# Create project config
cat > .sqlfmt.yaml << EOF
language: postgresql
keyword_case: lowercase
indent: "  "
lines_between_queries: 1
EOF

# Everyone on the team gets consistent formatting
sqlfmt format migrations/*.sql --write
```

**Personal preferences for all SQL files:**

```bash
# Create user config
cat > ~/.sqlfmtrc << EOF
language: sql
keyword_case: uppercase
indent: "\t"
EOF

# Apply to any SQL file
sqlfmt format ~/queries/report.sql
```

**Override config for specific dialects:**

```bash
# Project config uses PostgreSQL, but override for MySQL file
sqlfmt format --lang=mysql legacy_mysql_schema.sql
```

See the [Configuration Guide](configuration.md#configuration-files) for complete details on configuration options.

## Dialect-Specific Usage

### PostgreSQL

```bash
# Format PostgreSQL queries with dialect-specific features
sqlfmt format --lang=postgresql migrations/*.sql

# Validate PostgreSQL formatting
sqlfmt validate --lang=postgresql --uppercase functions.sql
```

### MySQL

```bash
# Format MySQL queries
sqlfmt format --lang=mysql schema.sql

# Format with colors for MySQL syntax
sqlfmt pretty-format --lang=mysql stored_procedures.sql
```

### SQLite

```bash
# Format SQLite queries
sqlfmt format --lang=sqlite database_setup.sql

# Validate SQLite formatting
sqlfmt validate --lang=sqlite *.sql
```

## Integration Examples

### Git Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit
sqlfmt validate --lang=postgresql migrations/*.sql
if [ $? -ne 0 ]; then
    echo "SQL formatting validation failed"
    exit 1
fi
```

### CI/CD Pipeline

```yaml
# GitHub Actions example
- name: Validate SQL formatting
  run: |
    go install github.com/MeKo-Christian/go-sqlfmt@latest
    sqlfmt validate --lang=postgresql sql/*.sql
    sqlfmt validate --lang=mysql migrations/*.sql
```

### Makefile Integration

```makefile
.PHONY: format-sql validate-sql

format-sql:
	sqlfmt format --write --lang=postgresql sql/*.sql
	sqlfmt format --write --lang=mysql migrations/*.sql

validate-sql:
	sqlfmt validate --lang=postgresql sql/*.sql
	sqlfmt validate --lang=mysql migrations/*.sql
```

## Tips and Best Practices

1. **Use dialect-specific formatting**: Always specify `--lang` for the best formatting results
2. **Validate in CI**: Add `sqlfmt validate` to your CI pipeline to catch formatting issues
3. **Consistent indentation**: Use the same `--indent` setting across your project
4. **Color output**: Use `pretty-format` or `pretty-print` for terminal viewing
5. **Batch processing**: Format multiple files at once with glob patterns like `*.sql`

## Troubleshooting

### Common Issues

**Files not formatting**: Check if the file has valid SQL syntax. The formatter requires parseable SQL.

**Unexpected output**: Verify you're using the correct `--lang` dialect for your SQL variant.

**Performance with large files**: For very large SQL files, consider splitting them or using the library API directly.

**Unicode characters**: The formatter supports UTF-8 encoding. Ensure your files are properly encoded.
