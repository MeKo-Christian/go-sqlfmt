# SQLite Dialect

SQLite support in go-sqlfmt provides comprehensive formatting for SQLite-specific syntax and features.

## Basic Usage

### Library Usage

Use the SQLite dialect by setting the language to `sqlfmt.SQLite`:

```go
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.SQLite)
fmt.Println(sqlfmt.Format("SELECT data->>'name' FROM users WHERE id = ?", cfg))
```

### CLI Usage

```bash
# Format SQLite-specific queries
sqlfmt format --lang=sqlite query.sql

# Format with colors for SQLite syntax
sqlfmt pretty-format --lang=sqlite query.sql

# Validate SQLite formatting
sqlfmt validate --lang=sqlite *.sql
```

## Supported Features

### Comments

Standard SQL comments are supported:

- `-- comment` - SQL standard line comments
- `/* block comment */` - Standard block comments

```sql
-- This is a line comment
/* This is a
   block comment */
SELECT * FROM users;
```

### Identifier Quoting

SQLite supports multiple identifier quoting styles for compatibility:

- `"double-quoted"` identifiers (standard SQL)
- `` `backtick-quoted` `` identifiers (MySQL-style)
- `[bracket-quoted]` identifiers (SQL Server-style)

```sql
SELECT "user id", `first-name`, [last name]
FROM "user-data"
WHERE "created_date" > '2023-01-01';
```

### Placeholders

All SQLite parameter binding styles are supported with 1-based indexing:

- `?` - Anonymous positional parameter
- `?NNN` - Numbered parameter (e.g., `?1`, `?2`)
- `:name` - Named parameter with colon prefix
- `@name` - Named parameter with at-sign prefix
- `$name` - Named parameter with dollar prefix

```sql
-- Anonymous positional
SELECT * FROM users WHERE id = ? AND status = ?;

-- Numbered parameters
SELECT * FROM users WHERE id = ?1 AND department = ?2 AND status = ?1;

-- Named parameters (various styles)
SELECT * FROM users WHERE name = :username AND email = @email AND role = $role;
```

### JSON Operations (SQLite 3.38+)

SQLite provides JSON path operators:

- `->` - JSON path extraction
- `->>` - JSON value extraction as text

```sql
SELECT
    profile->>'name' as name,
    settings->'theme' as theme_json,
    data->'address'->>'city' as city
FROM users
WHERE profile->>'active' = 'true';
```

### LIMIT Clauses

Both SQLite LIMIT syntax styles are supported:

- `LIMIT n OFFSET m` (standard)
- `LIMIT m, n` (MySQL-compatible)

```sql
-- Standard syntax
SELECT * FROM users
ORDER BY created_at
LIMIT 10 OFFSET 20;

-- MySQL-compatible syntax
SELECT * FROM users
ORDER BY created_at
LIMIT 20, 10;
```

### UPSERT Operations (SQLite 3.24+)

SQLite provides comprehensive UPSERT support:

#### INSERT ... ON CONFLICT

```sql
INSERT INTO users (id, name, email, updated_at)
VALUES (1, 'John Doe', 'john@example.com', datetime('now'))
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    email = excluded.email,
    updated_at = datetime('now');
```

#### INSERT ... ON CONFLICT ... DO NOTHING

```sql
INSERT INTO tags (name, category)
VALUES ('sqlite', 'database')
ON CONFLICT(name) DO NOTHING;
```

#### INSERT OR REPLACE

```sql
INSERT OR REPLACE INTO cache (key, value, expires_at)
VALUES ('user:1', '{"name":"John"}', datetime('now', '+1 day'));
```

### Advanced Features

#### Common Table Expressions (CTEs)

```sql
WITH RECURSIVE fibonacci(n, fib_n, next_fib_n) AS (
    -- Base case
    SELECT 1, 0, 1

    UNION ALL

    -- Recursive case
    SELECT n + 1, next_fib_n, fib_n + next_fib_n
    FROM fibonacci
    WHERE n < 10
)
SELECT n, fib_n FROM fibonacci;
```

#### Window Functions (SQLite 3.25+)

Full support for window functions with proper formatting:

```sql
SELECT
    employee_id,
    name,
    salary,
    department,
    ROW_NUMBER() OVER (
        PARTITION BY department
        ORDER BY salary DESC
    ) as dept_rank,
    AVG(salary) OVER (
        PARTITION BY department
    ) as dept_avg_salary,
    SUM(salary) OVER (
        ORDER BY hire_date
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) as running_total
FROM employees;
```

#### Generated Columns

```sql
CREATE TABLE products (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    price REAL NOT NULL,
    tax_rate REAL NOT NULL DEFAULT 0.08,
    price_with_tax REAL GENERATED ALWAYS AS (price * (1 + tax_rate)) VIRTUAL,
    price_category TEXT GENERATED ALWAYS AS (
        CASE
            WHEN price < 10 THEN 'budget'
            WHEN price < 100 THEN 'standard'
            ELSE 'premium'
        END
    ) STORED
);
```

#### Table Constraints

SQLite-specific table options:

- `WITHOUT ROWID` tables
- `STRICT` tables (SQLite 3.37+)

```sql
CREATE TABLE coordinates (
    x REAL,
    y REAL,
    PRIMARY KEY (x, y)
) WITHOUT ROWID;

CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    age INTEGER
) STRICT;
```

#### Triggers

```sql
CREATE TRIGGER update_modified_time
    AFTER UPDATE ON users
    FOR EACH ROW
BEGIN
    UPDATE users
    SET modified_at = datetime('now')
    WHERE id = NEW.id;
END;
```

#### Views

```sql
CREATE VIEW active_users AS
SELECT u.id, u.name, u.email, p.settings
FROM users u
JOIN profiles p ON u.id = p.user_id
WHERE u.active = 1;
```

### Pattern Matching

SQLite supports multiple pattern matching approaches:

- `LIKE` - SQL-standard pattern matching
- `GLOB` - Unix shell-style pattern matching

```sql
-- SQL LIKE patterns
SELECT * FROM files
WHERE name LIKE '%.txt'
   OR path LIKE '%/temp/%';

-- Unix GLOB patterns
SELECT * FROM files
WHERE name GLOB '*.{jpg,png,gif}'
   OR path GLOB '/home/*/Documents/*';
```

### NULL Handling

Enhanced NULL handling (SQLite 3.39+):

- `IS NULL`, `IS NOT NULL` (standard)
- `IS DISTINCT FROM`, `IS NOT DISTINCT FROM`

```sql
SELECT * FROM users
WHERE status IS DISTINCT FROM 'active'
   OR last_login IS NOT DISTINCT FROM NULL;
```

### String Operations

#### Concatenation

String concatenation using the `||` operator:

```sql
SELECT first_name || ' ' || last_name AS full_name,
       'User: ' || name || ' (' || email || ')' AS display_name
FROM users;
```

#### String Functions

SQLite provides comprehensive string functions:

```sql
SELECT
    UPPER(name) as name_upper,
    LENGTH(email) as email_length,
    SUBSTR(phone, 1, 3) as area_code,
    REPLACE(address, 'Street', 'St') as short_address
FROM contacts;
```

### Binary Data

#### Blob Literals

Binary data literals using `X'hexstring'` format:

```sql
INSERT INTO images (name, data)
VALUES ('icon.png', X'89504E470D0A1A0A0000000D49484452');

SELECT name, LENGTH(data) as size_bytes
FROM images
WHERE data LIKE X'89504E47%';  -- PNG signature
```

### PRAGMA Statements

`PRAGMA` statements receive minimal formatting to preserve functionality:

```sql
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -64000;  -- 64MB cache

-- Query pragmas
PRAGMA table_info('users');
PRAGMA index_list('users');
```

## Current Limitations

### REGEXP Operator

`REGEXP` operator support depends on user-defined functions and is treated as a standard operator. SQLite doesn't have built-in regex support.

### PRAGMA Values

`PRAGMA` values are preserved as-is without validation or reformatting to maintain SQLite compatibility.

### Unicode Identifiers

Unicode identifiers are preserved without case coercion inside quotes.

### Version Requirements

Some features require specific SQLite versions:

- UPSERT operations: SQLite 3.24+
- Window functions: SQLite 3.25+
- JSON operators: SQLite 3.38+
- `IS [NOT] DISTINCT FROM`: SQLite 3.39+
- `STRICT` tables: SQLite 3.37+

## Testing

### Run SQLite Tests

```bash
# All SQLite tests
go test ./pkg/sqlfmt -run TestSQLite

# Specific test patterns
go test ./pkg/sqlfmt -run "TestSQLite.*JSON"      # JSON operations
go test ./pkg/sqlfmt -run "TestSQLite.*UPSERT"    # UPSERT operations
go test ./pkg/sqlfmt -run "TestSQLite.*Window"    # Window functions
go test ./pkg/sqlfmt -run "TestSQLite.*CTE"       # Common Table Expressions

# Alternative test command
just test-sqlite

# Golden file tests
just test-golden
```

### Test Data Locations

- **Input files**: `testdata/input/sqlite/*.sql`
- **Expected output**: `testdata/golden/sqlite/*.sql`

## Implementation Status

**Current Status**: ✅ **Comprehensive SQLite support implemented**

- [x] Multiple identifier quoting styles
- [x] All parameter binding styles
- [x] JSON operators (SQLite 3.38+)
- [x] LIMIT syntax variations
- [x] UPSERT operations (SQLite 3.24+)
- [x] CTEs and window functions
- [x] Generated columns and table constraints
- [x] Triggers and views
- [x] Pattern matching (LIKE, GLOB)
- [x] NULL handling enhancements
- [x] String operations and blob literals
- [x] PRAGMA statement handling

The SQLite dialect provides comprehensive support for SQLite 3.24+ features with enhanced support for newer SQLite functionality.
