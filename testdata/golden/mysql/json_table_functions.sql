-- MySQL JSON_TABLE functions for converting JSON to relational format
SELECT
    jt.product_id,
    jt.product_name,
    jt.price,
    jt.category,
    jt.tags,
    jt.specifications
FROM products p,
JSON_TABLE(
    p.product_data,
    '$' COLUMNS (
        product_id INT PATH '$.id',
        product_name VARCHAR(100) PATH '$.name',
        price DECIMAL(10,2) PATH '$.pricing.base_price',
        category VARCHAR(50) PATH '$.category',
        tags JSON PATH '$.tags',
        specifications JSON PATH '$.specifications'
    )
) AS jt
WHERE p.active = 1;

-- Nested JSON_TABLE with multiple levels
SELECT
    u.user_id,
    u.username,
    profile.profile_name,
    profile.email,
    addr.city,
    addr.state,
    addr.zip_code
FROM users u,
JSON_TABLE(
    u.user_profile,
    '$' COLUMNS (
        profile_name VARCHAR(100) PATH '$.personal.name',
        email VARCHAR(150) PATH '$.contact.email',
        NESTED PATH '$.addresses[*]' COLUMNS (
            city VARCHAR(50) PATH '$.city',
            state VARCHAR(50) PATH '$.state',
            zip_code VARCHAR(10) PATH '$.zip_code'
        )
    )
) AS profile,
JSON_TABLE(
    u.user_profile,
    '$.addresses[*]' COLUMNS (
        city VARCHAR(50) PATH '$.city',
        state VARCHAR(50) PATH '$.state',
        zip_code VARCHAR(10) PATH '$.zip_code'
    )
) AS addr
WHERE u.active = 1;

-- JSON_TABLE with arrays and complex nesting
SELECT
    order_id,
    customer_name,
    item.item_name,
    item.quantity,
    item.unit_price,
    item.total_price,
    spec.spec_name,
    spec.spec_value
FROM orders o,
JSON_TABLE(
    o.order_details,
    '$' COLUMNS (
        order_id INT PATH '$.order_id',
        customer_name VARCHAR(100) PATH '$.customer.name',
        NESTED PATH '$.items[*]' COLUMNS (
            item_name VARCHAR(100) PATH '$.name',
            quantity INT PATH '$.quantity',
            unit_price DECIMAL(8,2) PATH '$.price',
            total_price DECIMAL(10,2) PATH '$.total',
            NESTED PATH '$.specifications[*]' COLUMNS (
                spec_name VARCHAR(50) PATH '$.name',
                spec_value VARCHAR(100) PATH '$.value'
            )
        )
    )
) AS item,
JSON_TABLE(
    o.order_details,
    '$.items[*].specifications[*]' COLUMNS (
        spec_name VARCHAR(50) PATH '$.name',
        spec_value VARCHAR(100) PATH '$.value'
    )
) AS spec
WHERE o.status = 'completed';

-- JSON_TABLE with conditional columns and error handling
SELECT
    product_id,
    name,
    CASE
        WHEN category IS NOT NULL THEN category
        ELSE 'uncategorized'
    END AS category,
    COALESCE(price, 0) AS price,
    COALESCE(discount, 0) AS discount,
    CASE
        WHEN in_stock = 1 THEN 'available'
        ELSE 'out_of_stock'
    END AS availability
FROM products p,
JSON_TABLE(
    p.product_data,
    '$' COLUMNS (
        product_id INT PATH '$.id',
        name VARCHAR(100) PATH '$.name',
        category VARCHAR(50) PATH '$.category' ERROR ON EMPTY,
        price DECIMAL(10,2) PATH '$.pricing.price' NULL ON EMPTY,
        discount DECIMAL(5,2) PATH '$.pricing.discount' DEFAULT 0 ON EMPTY,
        in_stock BOOLEAN PATH '$.inventory.available' DEFAULT FALSE ON EMPTY
    )
) AS jt
WHERE p.created_at >= DATE_SUB(CURDATE(), INTERVAL 1 YEAR);

-- JSON_TABLE with aggregation and joins
SELECT
    c.category_name,
    COUNT(jt.product_id) AS product_count,
    AVG(jt.price) AS avg_price,
    MIN(jt.price) AS min_price,
    MAX(jt.price) AS max_price,
    SUM(jt.stock_quantity) AS total_stock
FROM categories c
LEFT JOIN products p ON c.id = p.category_id,
JSON_TABLE(
    p.product_data,
    '$' COLUMNS (
        product_id INT PATH '$.id',
        price DECIMAL(10,2) PATH '$.pricing.price',
        stock_quantity INT PATH '$.inventory.quantity'
    )
) AS jt
WHERE p.active = 1
GROUP BY c.id, c.category_name
HAVING COUNT(jt.product_id) > 0
ORDER BY product_count DESC;

-- JSON_TABLE with window functions and analytics
SELECT
    department,
    employee_name,
    salary,
    ROW_NUMBER() OVER (PARTITION BY department ORDER BY salary DESC) AS dept_rank,
    AVG(salary) OVER (PARTITION BY department) AS dept_avg_salary,
    salary - AVG(salary) OVER (PARTITION BY department) AS salary_vs_avg,
    PERCENT_RANK() OVER (PARTITION BY department ORDER BY salary DESC) AS salary_percentile
FROM employees e,
JSON_TABLE(
    e.employee_data,
    '$' COLUMNS (
        employee_name VARCHAR(100) PATH '$.personal.name',
        department VARCHAR(50) PATH '$.work.department',
        salary DECIMAL(10,2) PATH '$.compensation.salary'
    )
) AS jt
WHERE e.active = 1
ORDER BY department, salary DESC;

-- Complex JSON_TABLE with multiple array expansions
SELECT
    company_id,
    company_name,
    dept.dept_name,
    dept.budget,
    emp.emp_name,
    emp.position,
    emp.salary,
    skill.skill_name,
    skill.proficiency
FROM companies c,
JSON_TABLE(
    c.company_structure,
    '$' COLUMNS (
        company_id INT PATH '$.id',
        company_name VARCHAR(100) PATH '$.name',
        NESTED PATH '$.departments[*]' COLUMNS (
            dept_name VARCHAR(50) PATH '$.name',
            budget DECIMAL(12,2) PATH '$.budget',
            NESTED PATH '$.employees[*]' COLUMNS (
                emp_name VARCHAR(100) PATH '$.name',
                position VARCHAR(50) PATH '$.position',
                salary DECIMAL(10,2) PATH '$.salary',
                NESTED PATH '$.skills[*]' COLUMNS (
                    skill_name VARCHAR(50) PATH '$.name',
                    proficiency VARCHAR(20) PATH '$.level'
                )
            )
        )
    )
) AS dept,
JSON_TABLE(
    c.company_structure,
    '$.departments[*].employees[*]' COLUMNS (
        emp_name VARCHAR(100) PATH '$.name',
        position VARCHAR(50) PATH '$.position',
        salary DECIMAL(10,2) PATH '$.salary'
    )
) AS emp,
JSON_TABLE(
    c.company_structure,
    '$.departments[*].employees[*].skills[*]' COLUMNS (
        skill_name VARCHAR(50) PATH '$.name',
        proficiency VARCHAR(20) PATH '$.level'
    )
) AS skill
WHERE c.active = 1;

-- JSON_TABLE with filtering and path expressions
SELECT
    order_id,
    customer_id,
    product_name,
    quantity,
    unit_price,
    total_amount,
    discount_applied,
    final_amount
FROM orders o,
JSON_TABLE(
    o.order_data,
    '$.line_items[*]' COLUMNS (
        product_name VARCHAR(100) PATH '$.product.name',
        quantity INT PATH '$.quantity',
        unit_price DECIMAL(8,2) PATH '$.pricing.unit_price',
        total_amount DECIMAL(10,2) PATH '$.pricing.total',
        discount_applied DECIMAL(8,2) PATH '$.discounts.applied' DEFAULT 0 ON EMPTY,
        final_amount DECIMAL(10,2) AS (total_amount - discount_applied)
    )
) AS jt
WHERE o.order_date >= DATE_SUB(CURDATE(), INTERVAL 6 MONTH)
    AND jt.final_amount > 100
ORDER BY o.order_date DESC, jt.final_amount DESC;</content>
<parameter name="filePath">/mnt/projekte/Code/go-sqlfmt/testdata/golden/mysql/json_table_functions.sql