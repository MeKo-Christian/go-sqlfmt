# IBM DB2 Dialect

IBM DB2 dialect provides formatting for DB2-specific SQL syntax and features.

## Basic Usage

### Library Usage

Use the DB2 dialect by setting the language to `sqlfmt.DB2`:

```go
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.DB2)
fmt.Println(sqlfmt.Format(query, cfg))
```

### CLI Usage

```bash
# Format DB2 queries
sqlfmt format --lang=db2 query.sql
```

## Supported Features

### DB2-Specific SQL Features

#### DB2 System Tables and Views

```sql
SELECT *
FROM SYSCAT.TABLES
WHERE TABSCHEMA = 'MYSCHEMA';

SELECT *
FROM SYSIBM.SYSDUMMY1;
```

#### DB2 Built-in Functions

**Date and Time Functions:**

```sql
SELECT
  CURRENT_DATE,
  CURRENT_TIME,
  CURRENT_TIMESTAMP,
  DAYS(CURRENT_DATE) - DAYS(hire_date) AS days_employed
FROM employees;
```

**String Functions:**

```sql
SELECT
  SUBSTR(last_name, 1, 5) AS short_name,
  POSSTR(email, '@') AS at_position,
  LEFT(first_name, 1) || '.' AS initial,
  RTRIM(LTRIM(full_name)) AS trimmed_name
FROM employees;
```

**Numeric Functions:**

```sql
SELECT
  TRUNCATE(salary, 0) AS whole_salary,
  MOD(employee_id, 10) AS id_remainder,
  SIGN(performance_rating - 3) AS rating_indicator
FROM employees;
```

#### WITH Clause (Common Table Expressions)

```sql
WITH regional_sales AS (
  SELECT region, SUM(sales_amount) AS total_sales
  FROM sales
  GROUP BY region
),
top_regions AS (
  SELECT region, total_sales
  FROM regional_sales
  WHERE total_sales > (
    SELECT AVG(total_sales) * 1.2
    FROM regional_sales
  )
)
SELECT r.region, r.total_sales
FROM top_regions r
ORDER BY r.total_sales DESC;
```

#### MERGE Statement

```sql
MERGE INTO target_employees T
USING source_employees S ON T.emp_id = S.emp_id
WHEN MATCHED THEN
  UPDATE SET
    T.salary = S.salary,
    T.last_updated = CURRENT_TIMESTAMP
WHEN NOT MATCHED THEN
  INSERT (emp_id, name, salary, hire_date)
  VALUES (S.emp_id, S.name, S.salary, CURRENT_DATE);
```

#### Window Functions

```sql
SELECT
  employee_id,
  department,
  salary,
  ROW_NUMBER() OVER (
    PARTITION BY department
    ORDER BY salary DESC
  ) AS dept_rank,
  SUM(salary) OVER (
    PARTITION BY department
  ) AS dept_total,
  LAG(salary, 1) OVER (
    ORDER BY hire_date
  ) AS prev_salary
FROM employees;
```

### DB2 Data Types

#### Common DB2 Data Types

- `SMALLINT`, `INTEGER`, `BIGINT`
- `DECIMAL(p,s)`, `NUMERIC(p,s)`
- `REAL`, `DOUBLE`
- `CHARACTER(n)`, `VARCHAR(n)`
- `CHAR(n) FOR BIT DATA`
- `CLOB`, `BLOB`
- `DATE`, `TIME`, `TIMESTAMP`
- `XML`

#### Example Usage

```sql
CREATE TABLE products (
  product_id INTEGER NOT NULL PRIMARY KEY,
  product_name VARCHAR(100) NOT NULL,
  description CLOB,
  price DECIMAL(10,2),
  image_data BLOB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  metadata XML
);
```

### DB2 Expressions and Operators

#### CASE Expressions

```sql
SELECT
  employee_id,
  salary,
  CASE
    WHEN salary < 30000 THEN 'Low'
    WHEN salary < 60000 THEN 'Medium'
    WHEN salary < 90000 THEN 'High'
    ELSE 'Executive'
  END AS salary_band
FROM employees;
```

#### DB2 Specific Operators

```sql
-- Concatenation operator
SELECT first_name CONCAT ' ' CONCAT last_name AS full_name
FROM employees;

-- Pattern matching with LIKE
SELECT *
FROM products
WHERE product_name LIKE '%computer%'
   OR description LIKE 'Laptop%';
```

### Subqueries and Set Operations

#### Correlated Subqueries

```sql
SELECT e1.employee_id, e1.salary
FROM employees e1
WHERE e1.salary > (
  SELECT AVG(e2.salary)
  FROM employees e2
  WHERE e2.department = e1.department
);
```

#### Set Operations

```sql
SELECT employee_id FROM current_employees
UNION
SELECT employee_id FROM former_employees
INTERSECT
SELECT employee_id FROM pension_eligible
EXCEPT
SELECT employee_id FROM terminated_employees;
```

### DB2 Joins

```sql
SELECT
  e.employee_id,
  e.name,
  d.department_name,
  p.project_name
FROM employees e
INNER JOIN departments d
  ON e.department_id = d.department_id
LEFT OUTER JOIN project_assignments pa
  ON e.employee_id = pa.employee_id
LEFT OUTER JOIN projects p
  ON pa.project_id = p.project_id
WHERE e.status = 'ACTIVE';
```

### DB2 Built-in Global Variables

```sql
SELECT
  USER,
  CURRENT_SCHEMA,
  CURRENT_SERVER,
  CURRENT_PATH
FROM SYSIBM.SYSDUMMY1;
```

### DB2 Identity Columns

```sql
CREATE TABLE orders (
  order_id INTEGER GENERATED ALWAYS AS IDENTITY (
    START WITH 1
    INCREMENT BY 1
    NO CACHE
  ),
  customer_id INTEGER NOT NULL,
  order_date DATE DEFAULT CURRENT_DATE,
  total_amount DECIMAL(10,2)
);

-- Using identity values
INSERT INTO orders (customer_id, total_amount)
VALUES (12345, 99.99);

SELECT order_id, IDENTITY_VAL_LOCAL() AS last_id
FROM orders;
```

### DB2 Sequences

```sql
CREATE SEQUENCE employee_id_seq
  START WITH 1000
  INCREMENT BY 1
  NO CACHE
  NO CYCLE;

INSERT INTO employees (employee_id, name)
VALUES (NEXT VALUE FOR employee_id_seq, 'John Doe');
```

## DB2-Specific Syntax

### FETCH FIRST Clause

```sql
SELECT employee_id, salary
FROM employees
ORDER BY salary DESC
FETCH FIRST 10 ROWS ONLY;
```

### OFFSET Clause (DB2 11+)

```sql
SELECT employee_id, name
FROM employees
ORDER BY hire_date
OFFSET 20 ROWS
FETCH NEXT 10 ROWS ONLY;
```

### DB2 Table Expressions

```sql
SELECT *
FROM TABLE(VALUES
  (1, 'Alice', 50000),
  (2, 'Bob', 60000),
  (3, 'Charlie', 55000)
) AS temp_emp(id, name, salary);
```

## Current Limitations

### Complex DB2 Features

- Some advanced DB2-specific constructs may not have specialized formatting
- DB2 stored procedures formatting is basic
- Complex XML operations may need manual formatting

### DB2 Versions

- Formatting rules are generally compatible across DB2 versions
- Some newer features may not have specific formatting enhancements

### z/OS Specifics

- z/OS DB2 specific features are treated as standard SQL where possible
- Some mainframe-specific syntax may not have enhanced formatting

## Testing

### Run DB2 Tests

```bash
# All DB2 tests
go test ./pkg/sqlfmt -run TestDB2

# Golden file tests
just test-golden
```

### Test Data Locations

- **Input files**: `testdata/input/db2/*.sql`
- **Expected output**: `testdata/golden/db2/*.sql`

## Implementation Status

**Current Status**: ✅ **Basic DB2 support implemented**

- [x] DB2 SQL syntax and keywords
- [x] DB2 built-in functions
- [x] Window functions
- [x] Common Table Expressions (WITH clause)
- [x] MERGE statements
- [x] DB2 data types
- [x] Set operations
- [x] Identity columns and sequences
- [x] FETCH FIRST and OFFSET clauses

The DB2 dialect provides solid support for IBM DB2 SQL features with formatting that maintains readability for complex queries and database operations.
