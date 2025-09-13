# MySQL Dialect

MySQL support in go-sqlfmt provides comprehensive formatting for MySQL-specific syntax and features, including MySQL 8.0 enhancements.

## Basic Usage

### Library Usage

Use the MySQL dialect by setting the language to `sqlfmt.MySQL`:

```go
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.MySQL)
fmt.Println(sqlfmt.Format("SELECT id, data->>'$.name' FROM users WHERE id = ?", cfg))
```

### CLI Usage

```bash
# Format MySQL-specific queries
sqlfmt format --lang=mysql query.sql

# Format with colors for MySQL syntax
sqlfmt pretty-format --lang=mysql query.sql

# Validate MySQL formatting
sqlfmt validate --lang=mysql *.sql
```

## Supported Features

### Comments

MySQL supports standard SQL comments along with MySQL-specific features:

- `-- comment` - SQL standard line comments
- `# comment` - MySQL-style hash line comments
- `/* block comment */` - Standard block comments
- `/*! version comment */` - MySQL versioned comments (preserved verbatim)

```sql
-- Standard SQL comment
# MySQL-style comment
/* Regular block comment */
/*!50001 CREATE ALGORITHM=UNDEFINED VIEW */
```

### Identifier Quoting

MySQL supports backtick-quoted identifiers for table and column names:

- `` `table-name` `` - Backtick-quoted identifiers (MySQL standard)
- `` `column name with spaces` `` - Handles spaces and special characters

```sql
SELECT `user_id`, `first name`
FROM `user-profiles`
WHERE `created-date` > '2023-01-01';
```

### Placeholders

MySQL parameter binding uses positional placeholders:

- `?` - Positional parameter (1-based indexing)

```sql
SELECT * FROM users
WHERE id = ? AND status = ?;
```

**Note**: MySQL does not support named parameters (`@var`, `:var`) or numbered parameters (`$1`).

### JSON Operations (MySQL 5.7+)

MySQL provides JSON path operators for working with JSON data:

- `->` - JSON path extraction (returns JSON)
- `->>` - JSON value extraction as text

```sql
SELECT
    profile->>'$.name' as name,
    settings->'$.theme' as theme_json,
    data->'$.address'->>'$.city' as city
FROM users
WHERE profile->>'$.active' = 'true';
```

### Special Operators

#### NULL-Safe Equality

- `<=>` - NULL-safe equality comparison

```sql
SELECT * FROM users
WHERE status <=> NULL;  -- Finds NULL values safely
```

#### Regular Expression Matching

- `REGEXP` / `RLIKE` - Regular expression matching
- `NOT REGEXP` - Negated regular expression matching

```sql
SELECT * FROM products
WHERE name REGEXP '^[A-Z].*Pro$'
   OR description NOT REGEXP 'deprecated|legacy';
```

#### Bitwise Operators

Full support for bitwise operations:

- `|`, `&`, `^` - Bitwise OR, AND, XOR
- `~` - Bitwise NOT
- `<<`, `>>` - Bitwise shift left/right

```sql
SELECT permissions & 0x04 as can_read,
       flags | 0x01 as with_flag,
       value << 2 as shifted
FROM user_permissions;
```

### LIMIT Clauses

Both MySQL LIMIT syntax styles are supported:

- `LIMIT n OFFSET m` - Standard SQL syntax
- `LIMIT m, n` - MySQL traditional syntax (offset, count)

```sql
-- Standard syntax
SELECT * FROM users
ORDER BY created_at
LIMIT 10 OFFSET 20;

-- MySQL traditional syntax
SELECT * FROM users
ORDER BY created_at
LIMIT 20, 10;
```

### Row Locking

MySQL provides several row locking options:

- `FOR UPDATE` - Exclusive lock
- `FOR SHARE` - Shared lock
- `LOCK IN SHARE MODE` - Legacy shared lock syntax

```sql
SELECT * FROM accounts
WHERE user_id = ?
FOR UPDATE;

SELECT * FROM products
WHERE category_id = ?
FOR SHARE;
```

### UPSERT Operations

MySQL provides multiple approaches for handling duplicate key conflicts:

#### INSERT ... ON DUPLICATE KEY UPDATE

```sql
INSERT INTO users (id, name, email, updated_at)
VALUES (1, 'John Doe', 'john@example.com', NOW())
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    email = VALUES(email),
    updated_at = NOW();
```

#### REPLACE INTO

```sql
REPLACE INTO cache (key, value, expires_at)
VALUES ('user:1', '{"name":"John"}', '2023-12-31 23:59:59');
```

#### INSERT IGNORE

```sql
INSERT IGNORE INTO tags (name, category)
VALUES ('mysql', 'database'), ('postgresql', 'database');
```

### Advanced Features (MySQL 8.0+)

#### Common Table Expressions (CTEs)

```sql
WITH RECURSIVE employee_hierarchy AS (
    -- Anchor member
    SELECT employee_id, name, manager_id, 1 as level
    FROM employees
    WHERE manager_id IS NULL

    UNION ALL

    -- Recursive member
    SELECT e.employee_id, e.name, e.manager_id, eh.level + 1
    FROM employees e
    INNER JOIN employee_hierarchy eh ON e.manager_id = eh.employee_id
)
SELECT * FROM employee_hierarchy
ORDER BY level, name;
```

#### Window Functions

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
    SUM(salary) OVER (
        PARTITION BY department
        ORDER BY hire_date
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) as running_dept_total
FROM employees;
```

### DDL Support

#### Index Creation

```sql
-- B-Tree index (default)
CREATE INDEX idx_users_email ON users (email);

-- Unique index
CREATE UNIQUE INDEX idx_users_username ON users (username);

-- Full-text index
CREATE FULLTEXT INDEX idx_posts_content ON posts (title, content);

-- Spatial index
CREATE SPATIAL INDEX idx_locations_coords ON locations (coordinates);

-- Hash index (Memory engine)
CREATE INDEX idx_session_id USING HASH ON sessions (session_id);
```

#### Table Options

MySQL table creation and alteration options are preserved:

```sql
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB
  CHARACTER SET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

-- Online DDL options
ALTER TABLE users
ADD COLUMN email VARCHAR(255) UNIQUE,
ALGORITHM=INSTANT,
LOCK=NONE;
```

#### Generated Columns

```sql
CREATE TABLE products (
    id INT PRIMARY KEY,
    price DECIMAL(10,2),
    tax_rate DECIMAL(3,2),
    price_with_tax DECIMAL(10,2) GENERATED ALWAYS AS (price * (1 + tax_rate)) VIRTUAL,
    price_category VARCHAR(20) GENERATED ALWAYS AS (
        CASE
            WHEN price < 10 THEN 'budget'
            WHEN price < 100 THEN 'standard'
            ELSE 'premium'
        END
    ) STORED
);
```

### Stored Routines (Basic Support)

Basic formatting support for MySQL stored procedures and functions:

#### Procedures

```sql
CREATE PROCEDURE UpdateUserStats(IN user_id INT)
BEGIN
    DECLARE total_orders INT DEFAULT 0;

    SELECT COUNT(*) INTO total_orders
    FROM orders
    WHERE customer_id = user_id;

    UPDATE users
    SET order_count = total_orders,
        last_updated = NOW()
    WHERE id = user_id;
END;
```

#### Functions

```sql
CREATE FUNCTION CalculateDiscount(
    original_price DECIMAL(10,2),
    discount_percent DECIMAL(3,2)
) RETURNS DECIMAL(10,2)
DETERMINISTIC
READS SQL DATA
BEGIN
    RETURN original_price * (1 - discount_percent / 100);
END;
```

### Control Flow Statements

Proper indentation for MySQL control flow:

```sql
CREATE PROCEDURE ProcessOrders()
BEGIN
    DECLARE done INT DEFAULT 0;
    DECLARE order_id INT;
    DECLARE order_cursor CURSOR FOR
        SELECT id FROM orders WHERE status = 'pending';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1;

    OPEN order_cursor;

    order_loop: LOOP
        FETCH order_cursor INTO order_id;

        IF done THEN
            LEAVE order_loop;
        END IF;

        -- Process order logic here
        UPDATE orders SET status = 'processing' WHERE id = order_id;
    END LOOP;

    CLOSE order_cursor;
END;
```

## Current Limitations

### ANSI_QUOTES Mode

`ANSI_QUOTES` mode is not supported - double quotes are always treated as string literals rather than identifier quotes.

### DELIMITER Statements

`DELIMITER` statements are treated as pass-through without formatting to avoid breaking copy/paste workflows:

```sql
DELIMITER //
CREATE PROCEDURE Example()
BEGIN
    -- Procedure body preserved as-is
END //
DELIMITER ;
```

### Versioned Comments

`/*! ... */` versioned comments are preserved verbatim without internal reformatting to maintain MySQL version compatibility.

### Complex Stored Routines

Stored routine formatting is lightweight - complex procedural logic may need manual formatting adjustment.

## Testing

### Run MySQL Tests

```bash
# All MySQL tests
go test ./pkg/sqlfmt -run TestMySQL

# Specific test patterns
go test ./pkg/sqlfmt -run "TestMySQL.*JSON"      # JSON operations
go test ./pkg/sqlfmt -run "TestMySQL.*CTE"       # Common Table Expressions
go test ./pkg/sqlfmt -run "TestMySQL.*Window"    # Window functions
go test ./pkg/sqlfmt -run "TestMySQL.*Upsert"    # UPSERT operations

# Golden file tests
just test-golden
```

### Test Data Locations

- **Input files**: `testdata/input/mysql/*.sql`
- **Expected output**: `testdata/golden/mysql/*.sql`

## Implementation Status

For detailed implementation progress, see `PLAN-MYSQL.md`.

**Current Status**: ✅ **Comprehensive MySQL support implemented**

- [x] Comments (including `#` and versioned comments)
- [x] Backtick identifier quoting
- [x] Positional placeholders (`?`)
- [x] JSON operators (`->`, `->>`)
- [x] Special operators (`<=>`, REGEXP, bitwise)
- [x] LIMIT syntax variations
- [x] Row locking (`FOR UPDATE`, `FOR SHARE`)
- [x] UPSERT operations (`ON DUPLICATE KEY UPDATE`, `REPLACE`, `INSERT IGNORE`)
- [x] CTEs and window functions (MySQL 8.0+)
- [x] DDL features (indexes, table options, generated columns)
- [x] Basic stored routine support
- [x] Control flow statements

The MySQL dialect provides comprehensive support for MySQL 5.7+ features with enhanced support for MySQL 8.0 functionality.
