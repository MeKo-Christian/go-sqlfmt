-- MySQL 8.0+ Window Functions with complex expressions
SELECT
    employee_id,
    department,
    salary,
    hire_date,
    ROW_NUMBER() OVER (
        PARTITION BY department
        ORDER BY salary DESC
    ) AS dept_salary_rank,
    RANK() OVER (
        PARTITION BY department
        ORDER BY salary DESC
    ) AS dept_salary_rank_dense,
    DENSE_RANK() OVER (
        PARTITION BY department
        ORDER BY salary DESC
    ) AS dept_salary_rank_no_gaps,
    PERCENT_RANK() OVER (
        PARTITION BY department
        ORDER BY salary DESC
    ) AS dept_salary_percentile,
    CUME_DIST() OVER (
        PARTITION BY department
        ORDER BY salary DESC
    ) AS dept_salary_cumulative_dist,
    NTILE(4) OVER (
        PARTITION BY department
        ORDER BY salary DESC
    ) AS dept_salary_quartile,
    LAG(salary, 1, 0) OVER (
        PARTITION BY department
        ORDER BY hire_date
    ) AS prev_salary_by_hire,
    LEAD(salary, 1, 0) OVER (
        PARTITION BY department
        ORDER BY hire_date
    ) AS next_salary_by_hire,
    FIRST_VALUE(salary) OVER (
        PARTITION BY department
        ORDER BY hire_date
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
    ) AS first_hired_salary,
    LAST_VALUE(salary) OVER (
        PARTITION BY department
        ORDER BY hire_date
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
    ) AS last_hired_salary,
    NTH_VALUE(salary, 2) OVER (
        PARTITION BY department
        ORDER BY hire_date
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
    ) AS second_hired_salary
FROM employees
WHERE active = 1
ORDER BY department, hire_date;

-- Advanced window functions with named windows and complex frames
SELECT
    product_id,
    product_name,
    category,
    sales_month,
    monthly_sales,
    SUM(monthly_sales) OVER w_product AS cumulative_sales,
    AVG(monthly_sales) OVER (
        PARTITION BY product_id
        ORDER BY sales_month
        ROWS BETWEEN 3 PRECEDING AND CURRENT ROW
    ) AS moving_avg_4months,
    SUM(monthly_sales) OVER w_category_product AS category_product_sales_6months,
    ROW_NUMBER() OVER w_category_monthly AS monthly_category_rank,
    RANK() OVER (
        PARTITION BY category
        ORDER BY SUM(monthly_sales) OVER w_product DESC
    ) AS overall_category_rank,
    LAG(monthly_sales, 12) OVER w_product AS sales_year_ago,
    (monthly_sales - LAG(monthly_sales, 12, 0) OVER w_product) /
        NULLIF(LAG(monthly_sales, 12, 0) OVER w_product, 0) * 100 AS yoy_growth_percent
FROM product_sales
WHERE sales_month >= '2020-01-01'
WINDOW
    w_product AS (PARTITION BY product_id ORDER BY sales_month),
    w_category AS (PARTITION BY category),
    w_category_product AS (PARTITION BY category, product_id ORDER BY sales_month
                          RANGE BETWEEN INTERVAL 6 MONTH PRECEDING AND CURRENT ROW),
    w_category_monthly AS (PARTITION BY category, sales_month ORDER BY monthly_sales DESC)
ORDER BY category, product_id, sales_month;

-- Window functions with GROUPS frame mode (MySQL 8.0+)
SELECT
    sensor_id,
    reading_time,
    temperature,
    AVG(temperature) OVER (
        PARTITION BY sensor_id
        ORDER BY reading_time
        GROUPS BETWEEN 5 PRECEDING AND CURRENT ROW
    ) AS avg_temp_6readings,
    SUM(temperature) OVER (
        PARTITION BY sensor_id
        ORDER BY reading_time
        GROUPS BETWEEN UNBOUNDED PRECEDING AND 2 FOLLOWING
    ) AS sum_temp_unbounded_to_2forward,
    COUNT(*) OVER (
        PARTITION BY sensor_id
        ORDER BY reading_time
        GROUPS BETWEEN 10 PRECEDING AND 10 FOLLOWING
    ) AS count_readings_21window,
    MIN(temperature) OVER (
        PARTITION BY sensor_id
        ORDER BY reading_time
        GROUPS BETWEEN CURRENT ROW AND UNBOUNDED FOLLOWING
    ) AS min_future_temp,
    MAX(temperature) OVER (
        PARTITION BY sensor_id
        ORDER BY reading_time
        GROUPS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS max_past_temp
FROM sensor_readings
WHERE reading_time >= DATE_SUB(NOW(), INTERVAL 24 HOUR)
ORDER BY sensor_id, reading_time;

-- Window functions with EXCLUDE clauses (MySQL 8.0+)
SELECT
    department,
    employee_id,
    salary,
    SUM(salary) OVER (
        PARTITION BY department
        ORDER BY salary DESC
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
        EXCLUDE CURRENT ROW
    ) AS total_salary_excl_current,
    AVG(salary) OVER (
        PARTITION BY department
        ORDER BY salary DESC
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
        EXCLUDE GROUP
    ) AS avg_salary_excl_ties,
    COUNT(*) OVER (
        PARTITION BY department
        ORDER BY salary DESC
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
        EXCLUDE TIES
    ) AS count_excl_ties,
    RANK() OVER (
        PARTITION BY department
        ORDER BY salary DESC
    ) AS salary_rank,
    DENSE_RANK() OVER (
        PARTITION BY department
        ORDER BY salary DESC
    ) AS salary_dense_rank
FROM employees
WHERE active = 1
ORDER BY department, salary DESC;

-- Window functions with complex expressions and aggregation
SELECT
    customer_id,
    order_date,
    order_total,
    ROW_NUMBER() OVER w_customer AS order_number,
    SUM(order_total) OVER w_customer AS customer_lifetime_value,
    AVG(order_total) OVER (
        PARTITION BY customer_id
        ORDER BY order_date
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS avg_order_value_to_date,
    MIN(order_total) OVER w_customer AS smallest_order,
    MAX(order_total) OVER w_customer AS largest_order,
    COUNT(*) OVER w_customer AS total_orders,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY order_total) OVER w_customer AS median_order_value,
    STDDEV(order_total) OVER w_customer AS order_value_stddev,
    VARIANCE(order_total) OVER w_customer AS order_value_variance,
    LAG(order_total) OVER w_customer AS previous_order_value,
    LEAD(order_total) OVER w_customer AS next_order_value,
    order_total - LAG(order_total, 1, 0) OVER w_customer AS order_value_change,
    CASE
        WHEN order_total > AVG(order_total) OVER w_customer THEN 'above_average'
        WHEN order_total < AVG(order_total) OVER w_customer THEN 'below_average'
        ELSE 'average'
    END AS order_size_category
FROM customer_orders
WHERE order_date >= '2023-01-01'
WINDOW w_customer AS (PARTITION BY customer_id ORDER BY order_date)
ORDER BY customer_id, order_date;

-- Window functions with JSON data and complex analytics
SELECT
    user_id,
    session_start,
    session_data->>'$.duration' AS session_duration,
    session_data->>'$.page_views' AS page_views,
    session_data->>'$.events_count' AS events_count,
    ROW_NUMBER() OVER (
        PARTITION BY user_id
        ORDER BY session_start DESC
    ) AS session_number,
    SUM(CAST(session_data->>'$.duration' AS DECIMAL(10,2))) OVER w_user AS total_duration,
    AVG(CAST(session_data->>'$.page_views' AS UNSIGNED)) OVER w_user AS avg_page_views,
    MAX(CAST(session_data->>'$.events_count' AS UNSIGNED)) OVER w_user AS max_events,
    JSON_LENGTH(session_data->>'$.events') AS events_array_length,
    LAG(CAST(session_data->>'$.duration' AS DECIMAL(10,2)), 1) OVER w_user AS prev_session_duration,
    LEAD(CAST(session_data->>'$.page_views' AS UNSIGNED), 1) OVER w_user AS next_session_page_views,
    CUME_DIST() OVER (
        PARTITION BY DATE(session_start)
        ORDER BY CAST(session_data->>'$.duration' AS DECIMAL(10,2)) DESC
    ) AS daily_duration_percentile
FROM user_sessions
WHERE session_start >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
WINDOW w_user AS (PARTITION BY user_id ORDER BY session_start)
ORDER BY user_id, session_start DESC;

-- Window functions with recursive CTE and complex hierarchies
WITH RECURSIVE employee_hierarchy AS (
    SELECT
        id,
        name,
        manager_id,
        department_id,
        salary,
        0 AS level,
        CAST(id AS CHAR(200)) AS path
    FROM employees
    WHERE manager_id IS NULL

    UNION ALL

    SELECT
        e.id,
        e.name,
        e.manager_id,
        e.department_id,
        e.salary,
        eh.level + 1,
        CONCAT(eh.path, ',', e.id)
    FROM employees e
    JOIN employee_hierarchy eh ON e.manager_id = eh.id
),
hierarchy_stats AS (
    SELECT
        *,
        ROW_NUMBER() OVER (PARTITION BY department_id ORDER BY level, salary DESC) AS dept_position,
        COUNT(*) OVER (PARTITION BY manager_id) AS reports_count,
        AVG(salary) OVER (PARTITION BY department_id) AS dept_avg_salary,
        SUM(salary) OVER (PARTITION BY manager_id ORDER BY salary DESC
                         ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS cumulative_team_salary
    FROM employee_hierarchy
)
SELECT
    name,
    level,
    salary,
    dept_position,
    reports_count,
    dept_avg_salary,
    cumulative_team_salary,
    salary - dept_avg_salary AS salary_vs_dept_avg,
    CASE
        WHEN salary > dept_avg_salary THEN 'above_dept_avg'
        WHEN salary < dept_avg_salary THEN 'below_dept_avg'
        ELSE 'at_dept_avg'
    END AS salary_category
FROM hierarchy_stats
ORDER BY department_id, level, salary DESC;</content>
<parameter name="filePath">/mnt/projekte/Code/go-sqlfmt/testdata/golden/mysql/window_functions_advanced.sql