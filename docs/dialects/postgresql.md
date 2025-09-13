# PostgreSQL Dialect

PostgreSQL support in go-sqlfmt includes comprehensive formatting for PostgreSQL-specific syntax and features.

## Basic Usage

### Library Usage

Use the PostgreSQL dialect by setting the language to `sqlfmt.PostgreSQL`:

```go
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.PostgreSQL)
fmt.Println(sqlfmt.Format("SELECT 'a'::text AS casted", cfg))
```

### CLI Usage

```bash
# Format PostgreSQL-specific queries
sqlfmt format --lang=postgresql query.sql

# Format with colors for PostgreSQL syntax
sqlfmt pretty-format --lang=postgresql query.sql

# Validate PostgreSQL formatting
sqlfmt validate --lang=postgresql *.sql
```

## Example Queries

### Type Casting

```go
query := "SELECT value::text, price::numeric(10,2) FROM products"
fmt.Println(sqlfmt.Format(query, cfg))
```

### JSON/JSONB Operations

```go
query := "SELECT data->>'name', metadata#>'{address,city}' FROM users"
fmt.Println(sqlfmt.Format(query, cfg))
```

### Common Table Expressions

```go
query := `WITH RECURSIVE category_tree AS (
  SELECT id, name, parent_id FROM categories WHERE parent_id IS NULL
  UNION ALL
  SELECT c.id, c.name, c.parent_id FROM categories c
  JOIN category_tree ct ON c.parent_id = ct.id
) SELECT * FROM category_tree`
fmt.Println(sqlfmt.Format(query, cfg))
```

### Window Functions with FILTER

```go
query := `SELECT
  product_id,
  SUM(quantity) FILTER (WHERE status = 'shipped') OVER (PARTITION BY product_id) AS shipped_qty,
  COUNT(*) OVER (ORDER BY created_at ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS running_count
FROM orders`
fmt.Println(sqlfmt.Format(query, cfg))
```

## Supported Features

### String Literals

- **Dollar-quoted strings**: `$$...$$`, `$tag$...$tag$`
- **Standard quoted strings** with proper escaping

Examples:

```sql
SELECT $$Hello, World!$$;
SELECT $function$
CREATE OR REPLACE FUNCTION...
$function$;
SELECT $tag$Some content with 'quotes'$tag$;
```

### Operators

- **Type cast operator**: `::` (formatted without spaces)
- **JSON path operators**: `->`, `->>` (extract JSON)
- **JSONB operators**: `#>`, `#>>` (extract path)
- **Containment operators**: `@>`, `<@`
- **Existence operators**: `?`, `?|`, `?&`
- **Regex operators**: `~`, `!~`, `~*`, `!~*`
- **Array concatenation**: `||`

Examples:

```sql
SELECT 'text'::varchar;                    -- Type casting
SELECT data->>'name' FROM users;           -- JSON extraction
SELECT metadata#>'{user,profile}';         -- JSONB path
SELECT tags @> ARRAY['postgresql'];        -- Containment
SELECT data ? 'key';                       -- Key existence
SELECT name ~ '^[A-Z]';                    -- Regex matching
SELECT arr1 || arr2;                       -- Array concatenation
```

### Placeholders

- **Named**: `@foo`, `:foo`
- **Numbered**: `$1`, `$2`, `$3` (1-based indexing)
- **Anonymous**: `?`

Examples:

```sql
SELECT * FROM users WHERE id = $1 AND status = $2;
SELECT * FROM users WHERE name = @username;
SELECT * FROM users WHERE active = ?;
```

### Advanced SQL Features

#### Common Table Expressions

- `WITH` - Common table expressions
- `WITH RECURSIVE` - Recursive CTEs

```sql
WITH sales_data AS (
    SELECT region, SUM(amount) as total
    FROM sales
    GROUP BY region
)
SELECT * FROM sales_data WHERE total > 1000;
```

#### Window Functions

- `OVER` - Window specifications
- `PARTITION BY` - Partitioning
- `FILTER (WHERE ...)` - Filtered aggregates
- Frame specifications: `ROWS`, `RANGE`, `GROUPS`

```sql
SELECT
    product_id,
    SUM(quantity) FILTER (WHERE status = 'completed') OVER (
        PARTITION BY category
        ORDER BY created_at
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) as running_total
FROM orders;
```

#### LATERAL Joins

- `LEFT LATERAL JOIN`
- `CROSS JOIN LATERAL`

```sql
SELECT u.name, recent.order_date
FROM users u
LEFT LATERAL JOIN (
    SELECT order_date
    FROM orders o
    WHERE o.user_id = u.id
    ORDER BY order_date DESC
    LIMIT 1
) recent ON true;
```

#### Array Operations

- Array subscripts: `[1]`, `[1:3]`
- `ARRAY` constructor
- Array functions: `unnest()`, `array_agg()`

```sql
SELECT
    tags[1] as first_tag,
    tags[2:4] as middle_tags,
    ARRAY[1, 2, 3] as numbers
FROM posts;
```

#### RETURNING Clauses

- `INSERT ... RETURNING`
- `UPDATE ... RETURNING`
- `DELETE ... RETURNING`

```sql
INSERT INTO users (name, email)
VALUES ('John', 'john@example.com')
RETURNING id, created_at;
```

#### Pattern Matching

- `ILIKE` - Case-insensitive LIKE
- `SIMILAR TO` - SQL regex patterns

```sql
SELECT * FROM users
WHERE name ILIKE '%john%'
   OR email SIMILAR TO '%@(gmail|yahoo)\.com';
```

#### Ordering Enhancements

- `NULLS FIRST`
- `NULLS LAST`

```sql
SELECT * FROM products
ORDER BY price DESC NULLS LAST,
         name ASC NULLS FIRST;
```

### PL/pgSQL and Functions

#### DO Blocks

```sql
DO $$
BEGIN
    PERFORM some_procedure();
    IF found THEN
        RAISE NOTICE 'Procedure executed successfully';
    END IF;
END
$$ LANGUAGE plpgsql;
```

#### Function Definitions

- `CREATE [OR REPLACE] FUNCTION`
- `RETURNS` specifications
- Function modifiers: `IMMUTABLE`, `STABLE`, `VOLATILE`
- Security options: `SECURITY DEFINER`, `SECURITY INVOKER`

```sql
CREATE OR REPLACE FUNCTION calculate_total(
    base_amount NUMERIC,
    tax_rate NUMERIC DEFAULT 0.08
) RETURNS NUMERIC
LANGUAGE SQL
IMMUTABLE
SECURITY DEFINER
AS $$
    SELECT base_amount + (base_amount * tax_rate);
$$;
```

### DDL Features

#### Concurrent Operations

- `CREATE INDEX CONCURRENTLY`
- `DROP INDEX CONCURRENTLY`

```sql
CREATE INDEX CONCURRENTLY idx_users_email
ON users (email)
WHERE active = true;
```

#### Index Methods

- `USING BTREE`, `USING GIN`, `USING GIST`
- `USING HASH`, `USING SP-GIST`, `USING BRIN`

```sql
CREATE INDEX idx_users_data_gin
ON users USING GIN (profile_data);
```

#### Covering Indexes

- `INCLUDE (columns)`

```sql
CREATE INDEX idx_orders_status_include
ON orders (status)
INCLUDE (created_at, total_amount);
```

#### Conditional DDL

- `IF NOT EXISTS`
- `IF EXISTS`

```sql
CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    table_name TEXT NOT NULL
);
```

## Current Limitations

### UPSERT Formatting

**Known Issue**: `INSERT ... ON CONFLICT` statements have formatting limitations due to tokenizer architecture constraints.

**Current behavior**:

```sql
INSERT INTO users (id, name, email)
VALUES
  (1, 'John', 'john@example.com') ON CONFLICT (id)
DO
  NOTHING;
```

**Expected behavior**:

```sql
INSERT INTO users (id, name, email)
VALUES (1, 'John', 'john@example.com') ON CONFLICT (id) DO NOTHING;
```

### Complex PL/pgSQL Indentation

Some complex procedural constructs may need manual indentation adjustment.

## Testing

### Run PostgreSQL Tests

```shell
# All PostgreSQL tests
go test ./pkg/sqlfmt -run TestPostgreSQL

# Specific PostgreSQL formatter tests
go test ./pkg/sqlfmt -run TestPostgreSQLFormatter

# Golden file tests
just test-golden

# Feature-specific test patterns
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Dollar"    # Dollar-quoted strings
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Cast"      # Type casting
go test ./pkg/sqlfmt -run "TestPostgreSQL.*JSON"      # JSON operations
go test ./pkg/sqlfmt -run "TestPostgreSQL.*CTE"       # Common Table Expressions
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Window"    # Window functions
go test ./pkg/sqlfmt -run "TestPostgreSQL.*Array"     # Array operations
```

### Test Data Locations

- **Input files**: `testdata/input/postgresql/*.sql`
- **Expected output**: `testdata/golden/postgresql/*.sql`
- **Snapshots**: `__snapshots__/` directories

### Update Tests

```shell
# Update snapshots if output changes
just update-snapshots

# Or manually
UPDATE_SNAPS=true go test ./pkg/sqlfmt -run TestSnapshot
```

## Implementation Status

For detailed implementation progress and roadmap, see `PLAN-POSTGRESQL.md`.

**Current Status**: ✅ **Most PostgreSQL features implemented successfully**

- [x] Dollar-quoted strings
- [x] Type cast operators
- [x] JSON/JSONB operators
- [x] Numbered placeholders ($1, $2, ...)
- [x] Pattern matching operators
- [x] CTEs and window functions
- [x] LATERAL joins
- [x] Array operations
- [x] RETURNING clauses
- [x] Basic PL/pgSQL support
- [x] DDL enhancements
- [ ] UPSERT formatting (known limitation)

The PostgreSQL dialect provides comprehensive support for most PostgreSQL-specific syntax and formatting conventions.
