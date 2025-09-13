# Standard SQL Dialect

Standard SQL (ANSI SQL) support provides formatting for core SQL language features according to SQL standard specifications.

## Basic Usage

### Library Usage

Standard SQL is the default dialect:

```go
// Using default configuration (Standard SQL)
fmt.Println(sqlfmt.Format(query))

// Explicitly setting Standard SQL
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.StandardSQL)
fmt.Println(sqlfmt.Format(query, cfg))
```

### CLI Usage

```bash
# Format using Standard SQL (default)
sqlfmt format query.sql

# Explicitly specify Standard SQL dialect
sqlfmt format --lang=sql query.sql
sqlfmt format --lang=standard query.sql
```

## Supported Features

### Core SQL Statements

- `SELECT` - Query data with proper indentation
- `INSERT` - Insert statements with value formatting
- `UPDATE` - Update statements with SET clause formatting
- `DELETE` - Delete statements with WHERE clause formatting
- `CREATE TABLE` - Table creation with column definitions
- `ALTER TABLE` - Table modification statements
- `DROP TABLE` - Table deletion statements

### Query Clauses

- `FROM` - Table references with proper alignment
- `WHERE` - Conditional clauses with logical operators
- `GROUP BY` - Grouping clauses
- `HAVING` - Group filtering clauses
- `ORDER BY` - Sorting clauses with ASC/DESC
- `LIMIT` - Result limiting (where supported)

### Joins

- `INNER JOIN` - Inner join operations
- `LEFT JOIN` / `LEFT OUTER JOIN` - Left outer joins
- `RIGHT JOIN` / `RIGHT OUTER JOIN` - Right outer joins
- `FULL JOIN` / `FULL OUTER JOIN` - Full outer joins
- `CROSS JOIN` - Cross product joins

### Subqueries and Expressions

- Subqueries in SELECT, FROM, WHERE clauses
- `CASE` expressions with proper indentation
- Aggregate functions (COUNT, SUM, AVG, MIN, MAX)
- Window functions (basic OVER clauses)
- Common Table Expressions (CTEs) with `WITH`

### Data Types

Standard SQL data types are recognized and formatted:

- Numeric types: INTEGER, DECIMAL, NUMERIC, REAL, DOUBLE
- Character types: CHAR, VARCHAR, TEXT
- Date/time types: DATE, TIME, TIMESTAMP
- Boolean type: BOOLEAN

### Operators

- Arithmetic operators: +, -, \*, /, %
- Comparison operators: =, <>, !=, <, <=, >, >=
- Logical operators: AND, OR, NOT
- Pattern matching: LIKE, NOT LIKE
- NULL testing: IS NULL, IS NOT NULL
- Set operations: IN, NOT IN, EXISTS, NOT EXISTS

### Comments

- Line comments: `-- comment text`
- Block comments: `/* comment text */`

### String Literals

- Single-quoted strings: `'text'`
- Escaped quotes: `'can''t'`
- National character strings: `N'text'`

### Placeholders

Standard placeholder support:

- Named placeholders: `@param`, `:param`
- Anonymous placeholders: `?`

## Example Formatting

### Basic Query

Input:

```sql
select u.id,u.name,p.title from users u join posts p on u.id=p.user_id where u.active=1 order by p.created_at desc
```

Output:

```sql
SELECT
  u.id,
  u.name,
  p.title
FROM
  users u
  JOIN posts p ON u.id = p.user_id
WHERE
  u.active = 1
ORDER BY
  p.created_at DESC
```

### Complex Query with CTE

Input:

```sql
with active_users as (select id,name from users where active=1),recent_posts as (select user_id,title,created_at from posts where created_at>='2023-01-01') select u.name,count(p.title) as post_count from active_users u left join recent_posts p on u.id=p.user_id group by u.id,u.name having count(p.title)>5 order by post_count desc
```

Output:

```sql
WITH active_users AS (
  SELECT
    id,
    name
  FROM
    users
  WHERE
    active = 1
),
recent_posts AS (
  SELECT
    user_id,
    title,
    created_at
  FROM
    posts
  WHERE
    created_at >= '2023-01-01'
)
SELECT
  u.name,
  COUNT(p.title) AS post_count
FROM
  active_users u
  LEFT JOIN recent_posts p ON u.id = p.user_id
GROUP BY
  u.id,
  u.name
HAVING
  COUNT(p.title) > 5
ORDER BY
  post_count DESC
```

### CASE Expression

```sql
SELECT
  name,
  CASE
    WHEN age < 18 THEN 'Minor'
    WHEN age < 65 THEN 'Adult'
    ELSE 'Senior'
  END AS age_group
FROM
  users
```

## Configuration Options

### Indentation

```go
cfg := sqlfmt.NewDefaultConfig().WithIndent("    ") // 4 spaces
cfg := sqlfmt.NewDefaultConfig().WithIndent("\t")   // tabs
```

### Uppercase Keywords

```go
cfg := sqlfmt.NewDefaultConfig().WithUppercase(true)
```

Output:

```sql
SELECT
  *
FROM
  users
WHERE
  active = 1
```

### Lines Between Queries

```go
cfg := sqlfmt.NewDefaultConfig().WithLinesBetweenQueries(3)
```

## Testing

### Run Standard SQL Tests

```bash
# All Standard SQL tests
go test ./pkg/sqlfmt -run TestFormat

# Specific test patterns
go test ./pkg/sqlfmt -run TestStandard
go test ./pkg/sqlfmt -run TestSQL

# Golden file tests
just test-golden
```

## Implementation Status

**Current Status**: ✅ **Comprehensive Standard SQL support**

- [x] All core SQL statements
- [x] JOIN operations
- [x] Subqueries and CTEs
- [x] CASE expressions
- [x] Window functions (basic)
- [x] Aggregate functions
- [x] Standard operators and functions
- [x] Comments and string literals
- [x] Parameter placeholders
- [x] Configurable formatting options

Standard SQL dialect provides the foundation for all other SQL dialects and implements comprehensive ANSI SQL formatting capabilities.
