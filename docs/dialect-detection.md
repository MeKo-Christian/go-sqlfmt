# SQL Dialect Auto-Detection

go-sqlfmt can automatically detect the SQL dialect from file extensions and SQL content, eliminating the need to manually specify the dialect for most use cases.

## Overview

The auto-detection feature works in two stages:

1. **File Extension Detection** - Checks the file extension for dialect-specific patterns
2. **Content-Based Detection** - Analyzes the SQL content for dialect-specific syntax patterns

## Usage

Use the `--auto-detect` flag with the `format` command:

```bash
# Auto-detect from file extension
sqlfmt format --auto-detect query.mysql

# Auto-detect from content (when extension is generic)
sqlfmt format --auto-detect query.sql

# Works with stdin too
cat query.sql | sqlfmt format --auto-detect -
```

When auto-detection succeeds, the formatted output will indicate which dialect was detected:

```bash
$ sqlfmt format --auto-detect --write query.mysql
Formatted query.mysql (detected as mysql)
```

## Supported File Extensions

The following file extensions are recognized:

| Extension | Dialect | Examples |
|-----------|---------|----------|
| `.psql`, `.pgsql` | PostgreSQL | `query.psql`, `migration.pgsql` |
| `.mysql` | MySQL | `schema.mysql` |
| `.sqlite` | SQLite | `database.sqlite` |
| `.plsql` | PL/SQL | `procedure.plsql` |
| `.ora.sql` | PL/SQL | `oracle.ora.sql` |

### Compound Extensions

Files with compound extensions are also supported:

| Extension | Dialect | Examples |
|-----------|---------|----------|
| `.mysql.sql` | MySQL | `query.mysql.sql` |
| `.psql.sql` | PostgreSQL | `migration.psql.sql` |
| `.sqlite.sql` | SQLite | `schema.sqlite.sql` |
| `.db.sql` | SQLite | `database.db.sql` |
| `.my.sql` | MySQL | `legacy.my.sql` |

## Content-Based Detection Patterns

When file extension detection fails or for generic `.sql` files, go-sqlfmt analyzes the SQL content for dialect-specific patterns.

### PostgreSQL Patterns

- Type casting: `column::integer`, `value::text`
- Dollar quoting: `$$content$$`, `$tag$content$tag$`
- Positional parameters: `$1`, `$2`, `$3`
- RETURNING clause: `INSERT ... RETURNING id`
- JSON operators: `->`, `->>`, `#>`, `#>>`, `@>`, `<@`
- PostgreSQL types: `::jsonb`, `::tsvector`, `::int4range`
- Functions: `generate_series()`, `unnest()`, `array_agg()`

### MySQL Patterns

- Backtick identifiers: `` `column_name` ``, `` `table`.`column` ``
- INSERT variants: `INSERT IGNORE`, `REPLACE INTO`
- UPDATE syntax: `ON DUPLICATE KEY UPDATE`
- MySQL functions: `GROUP_CONCAT()`, `FOUND_ROWS()`
- Storage engine: `ENGINE=InnoDB`, `ENGINE=MyISAM`
- Character set: `CHARSET=utf8`, `COLLATE=utf8_general_ci`
- Index hints: `FORCE INDEX`, `USE INDEX`, `IGNORE INDEX`

### SQLite Patterns

- Pragmas: `PRAGMA foreign_keys = ON`
- Table options: `WITHOUT ROWID`
- Autoincrement: `INTEGER PRIMARY KEY AUTOINCREMENT`
- Attach/detach: `ATTACH DATABASE`, `DETACH DATABASE`
- SQLite functions: `VACUUM`, `ANALYZE`, `REINDEX`
- EXPLAIN: `EXPLAIN QUERY PLAN`

### PL/SQL Patterns

- Block structure: `BEGIN ... END;`
- Exception handling: `EXCEPTION WHEN ... THEN`
- Procedure/function: `CREATE PROCEDURE`, `CREATE FUNCTION`
- Package: `CREATE PACKAGE`, `PACKAGE BODY`
- Cursors: `OPEN ... FOR`, `FETCH ... INTO`
- PL/SQL types: `PLS_INTEGER`, `BOOLEAN`, `REF CURSOR`
- Dynamic SQL: `EXECUTE IMMEDIATE`

## Detection Priority

Detection follows this priority order:

1. **File Extension** (highest priority)
2. **PostgreSQL** content patterns
3. **PL/SQL** content patterns
4. **MySQL** content patterns
5. **SQLite** content patterns (lowest priority)

This ordering ensures that more distinctive patterns are checked first, reducing false positives.

## Limitations

### False Positives

- Some SQL constructs may match patterns from multiple dialects
- Generic SQL keywords might trigger detection when they shouldn't
- Complex queries mixing multiple dialects may not be detected correctly

### False Negatives

- Very simple queries lacking distinctive dialect features
- Custom SQL that doesn't use standard dialect patterns
- Non-standard SQL extensions not covered by the detection patterns

### Recommendations

- Use explicit `--lang` flag for complex or mixed-dialect projects
- For projects with multiple dialects, consider:
  - Using dialect-specific file extensions
  - Configuration files with per-file overrides
  - Explicit `--lang` flags in scripts

## Configuration

Auto-detection can be combined with configuration files. The detected dialect will override the default language in your `.sqlfmt.yaml`:

```yaml
# .sqlfmt.yaml
indent: "  "
keyword_case: "lowercase"
```

```bash
# Will use detected dialect with lowercase keywords and 2-space indent
sqlfmt format --auto-detect query.mysql
```

## Examples

### File Extension Detection

```bash
# PostgreSQL file
$ sqlfmt format --auto-detect migration.psql
Formatted migration.psql (detected as postgresql)

# MySQL file
$ sqlfmt format --auto-detect schema.mysql.sql
Formatted schema.mysql.sql (detected as mysql)
```

### Content-Based Detection

```bash
# Generic .sql file with PostgreSQL syntax
$ cat query.sql
SELECT id::integer, data->>'name' FROM users WHERE id = $1;

$ sqlfmt format --auto-detect query.sql
SELECT
  id::integer,
  data->>'name'
FROM
  users
WHERE
  id = $1;
Formatted query.sql (detected as postgresql)
```

### Fallback Behavior

When auto-detection fails, go-sqlfmt falls back to Standard SQL:

```bash
# Very simple query without distinctive features
$ echo "SELECT * FROM users;" | sqlfmt format --auto-detect -
SELECT
  *
FROM
  users;
# No detection message = fell back to standard SQL
```
