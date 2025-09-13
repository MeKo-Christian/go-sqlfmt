# Oracle PL/SQL Dialect

Oracle PL/SQL dialect provides formatting for Oracle Database SQL and PL/SQL procedural language features.

## Basic Usage

### Library Usage

Use the PL/SQL dialect by setting the language to `sqlfmt.PLSQL`:

```go
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.PLSQL)
fmt.Println(sqlfmt.Format(query, cfg))
```

### CLI Usage

```bash
# Format PL/SQL queries
sqlfmt format --lang=pl/sql query.sql
sqlfmt format --lang=plsql query.sql
sqlfmt format --lang=oracle query.sql
```

## Supported Features

### Oracle-Specific SQL Features

#### DUAL Table

```sql
SELECT SYSDATE FROM DUAL;
SELECT USER FROM DUAL;
```

#### Oracle Functions

- Date functions: SYSDATE, SYSTIMESTAMP, TO_DATE, TO_CHAR
- String functions: SUBSTR, INSTR, LENGTH, RTRIM, LTRIM
- Numeric functions: ROUND, TRUNC, MOD, POWER
- Conversion functions: TO_NUMBER, TO_CHAR, TO_DATE

```sql
SELECT
  TO_CHAR(hire_date, 'YYYY-MM-DD') AS formatted_date,
  SUBSTR(last_name, 1, 10) AS short_name,
  ROUND(salary * 1.1, 2) AS increased_salary
FROM employees
WHERE hire_date > TO_DATE('2020-01-01', 'YYYY-MM-DD');
```

#### CONNECT BY (Hierarchical Queries)

```sql
SELECT
  employee_id,
  last_name,
  manager_id,
  LEVEL,
  SYS_CONNECT_BY_PATH(last_name, '/') AS path
FROM employees
START WITH manager_id IS NULL
CONNECT BY PRIOR employee_id = manager_id
ORDER SIBLINGS BY last_name;
```

#### MERGE Statement

```sql
MERGE INTO target_table t
USING source_table s ON (t.id = s.id)
WHEN MATCHED THEN
  UPDATE SET t.name = s.name, t.updated_at = SYSDATE
WHEN NOT MATCHED THEN
  INSERT (id, name, created_at)
  VALUES (s.id, s.name, SYSDATE);
```

### PL/SQL Procedural Constructs

#### Anonymous Blocks

```sql
DECLARE
  v_count NUMBER;
  v_message VARCHAR2(100);
BEGIN
  SELECT COUNT(*) INTO v_count FROM employees;

  IF v_count > 100 THEN
    v_message := 'Large company';
  ELSE
    v_message := 'Small company';
  END IF;

  DBMS_OUTPUT.PUT_LINE(v_message);
EXCEPTION
  WHEN NO_DATA_FOUND THEN
    DBMS_OUTPUT.PUT_LINE('No data found');
  WHEN OTHERS THEN
    DBMS_OUTPUT.PUT_LINE('Error: ' || SQLCODE || ' ' || SQLERRM);
END;
/
```

#### Stored Procedures

```sql
CREATE OR REPLACE PROCEDURE update_salary(
  p_employee_id IN employees.employee_id%TYPE,
  p_percentage IN NUMBER
) AS
  v_current_salary employees.salary%TYPE;
BEGIN
  SELECT salary INTO v_current_salary
  FROM employees
  WHERE employee_id = p_employee_id;

  UPDATE employees
  SET salary = salary * (1 + p_percentage / 100)
  WHERE employee_id = p_employee_id;

  COMMIT;
EXCEPTION
  WHEN NO_DATA_FOUND THEN
    RAISE_APPLICATION_ERROR(-20001, 'Employee not found');
END update_salary;
/
```

#### Functions

```sql
CREATE OR REPLACE FUNCTION get_employee_bonus(
  p_employee_id IN NUMBER,
  p_year IN NUMBER
) RETURN NUMBER
IS
  v_bonus NUMBER := 0;
  v_performance_rating NUMBER;
BEGIN
  SELECT performance_rating
  INTO v_performance_rating
  FROM performance_reviews
  WHERE employee_id = p_employee_id
    AND review_year = p_year;

  CASE v_performance_rating
    WHEN 5 THEN v_bonus := 5000;
    WHEN 4 THEN v_bonus := 3000;
    WHEN 3 THEN v_bonus := 1000;
    ELSE v_bonus := 0;
  END CASE;

  RETURN v_bonus;
END get_employee_bonus;
/
```

#### Packages

```sql
CREATE OR REPLACE PACKAGE employee_pkg AS
  PROCEDURE hire_employee(
    p_first_name VARCHAR2,
    p_last_name VARCHAR2,
    p_email VARCHAR2
  );

  FUNCTION get_employee_count RETURN NUMBER;
END employee_pkg;
/

CREATE OR REPLACE PACKAGE BODY employee_pkg AS
  PROCEDURE hire_employee(
    p_first_name VARCHAR2,
    p_last_name VARCHAR2,
    p_email VARCHAR2
  ) IS
  BEGIN
    INSERT INTO employees (first_name, last_name, email, hire_date)
    VALUES (p_first_name, p_last_name, p_email, SYSDATE);

    COMMIT;
  END hire_employee;

  FUNCTION get_employee_count RETURN NUMBER IS
    v_count NUMBER;
  BEGIN
    SELECT COUNT(*) INTO v_count FROM employees;
    RETURN v_count;
  END get_employee_count;
END employee_pkg;
/
```

### Control Structures

#### IF Statements

```sql
IF condition1 THEN
  statement1;
ELSIF condition2 THEN
  statement2;
ELSE
  statement3;
END IF;
```

#### CASE Statements

```sql
CASE variable
  WHEN value1 THEN statement1;
  WHEN value2 THEN statement2;
  ELSE default_statement;
END CASE;
```

#### Loops

```sql
-- Basic LOOP
LOOP
  statement;
  EXIT WHEN condition;
END LOOP;

-- FOR LOOP
FOR i IN 1..10 LOOP
  statement;
END LOOP;

-- WHILE LOOP
WHILE condition LOOP
  statement;
END LOOP;
```

### Exception Handling

```sql
BEGIN
  -- Risky operations
  statement;
EXCEPTION
  WHEN NO_DATA_FOUND THEN
    handle_no_data;
  WHEN TOO_MANY_ROWS THEN
    handle_too_many_rows;
  WHEN OTHERS THEN
    handle_general_error;
END;
```

### Cursors

#### Explicit Cursors

```sql
DECLARE
  CURSOR emp_cursor IS
    SELECT employee_id, first_name, last_name
    FROM employees
    WHERE department_id = 10;

  emp_rec emp_cursor%ROWTYPE;
BEGIN
  OPEN emp_cursor;

  LOOP
    FETCH emp_cursor INTO emp_rec;
    EXIT WHEN emp_cursor%NOTFOUND;

    DBMS_OUTPUT.PUT_LINE(emp_rec.first_name || ' ' || emp_rec.last_name);
  END LOOP;

  CLOSE emp_cursor;
END;
/
```

#### Cursor FOR Loops

```sql
BEGIN
  FOR emp_rec IN (
    SELECT employee_id, first_name, last_name
    FROM employees
    WHERE department_id = 10
  ) LOOP
    DBMS_OUTPUT.PUT_LINE(emp_rec.first_name || ' ' || emp_rec.last_name);
  END LOOP;
END;
/
```

## Oracle-Specific Features

### Sequence Operations

```sql
INSERT INTO employees (employee_id, first_name, last_name)
VALUES (employee_seq.NEXTVAL, 'John', 'Doe');

SELECT employee_seq.CURRVAL FROM DUAL;
```

### Analytic Functions

```sql
SELECT
  employee_id,
  salary,
  ROW_NUMBER() OVER (ORDER BY salary DESC) AS salary_rank,
  RANK() OVER (PARTITION BY department_id ORDER BY salary DESC) AS dept_rank,
  LAG(salary, 1, 0) OVER (ORDER BY hire_date) AS prev_salary
FROM employees;
```

### WITH Clause (Subquery Factoring)

```sql
WITH
dept_totals AS (
  SELECT department_id, SUM(salary) AS total_salary
  FROM employees
  GROUP BY department_id
),
avg_dept_salary AS (
  SELECT AVG(total_salary) AS avg_total
  FROM dept_totals
)
SELECT d.department_name, dt.total_salary
FROM departments d
JOIN dept_totals dt ON d.department_id = dt.department_id
CROSS JOIN avg_dept_salary ads
WHERE dt.total_salary > ads.avg_total;
```

## Current Limitations

### Comment Handling

- Block comments are supported but nested comments may not format perfectly
- Single-line comments (`--`) are fully supported

### Complex PL/SQL Constructs

- Very complex nested procedures may require manual formatting adjustment
- Dynamic SQL formatting is preserved as-is

### Oracle-Specific Syntax

- Some advanced Oracle features may not have specialized formatting rules
- Proprietary Oracle extensions are treated as standard identifiers

## Testing

### Run PL/SQL Tests

```bash
# All PL/SQL tests
go test ./pkg/sqlfmt -run TestPLSQL

# Golden file tests
just test-golden
```

### Test Data Locations

- **Input files**: `testdata/input/plsql/*.sql`
- **Expected output**: `testdata/golden/plsql/*.sql`

## Implementation Status

**Current Status**: ✅ **Basic PL/SQL support implemented**

- [x] Oracle SQL extensions
- [x] PL/SQL procedural constructs
- [x] Control structures (IF, CASE, loops)
- [x] Exception handling
- [x] Cursor operations
- [x] Package specifications and bodies
- [x] Functions and procedures
- [x] Oracle-specific functions
- [x] Analytic functions

The PL/SQL dialect provides solid support for Oracle Database SQL and PL/SQL procedural language features, with formatting that maintains readability for complex stored procedures and packages.
