-- MySQL CTEs and Window Functions Test File
-- Tests WITH clauses, recursive CTEs, and window functions (MySQL 8.0+)
-- Simple CTE
WITH
  active_users AS (
    SELECT
      user_id,
      username,
      email,
      created_at
    FROM
      users
    WHERE
      status = 'active'
      AND last_login >= DATE_SUB(NOW(), INTERVAL 30 DAY)
  )
SELECT
  au.username,
  au.email,
  COUNT(o.id) as order_count
FROM
  active_users au
  LEFT JOIN orders o ON au.user_id = o.customer_id
GROUP BY
  au.user_id,
  au.username,
  au.email
ORDER BY
  order_count DESC;

-- Multiple CTEs
WITH
  high_value_customers AS (
    SELECT
      customer_id,
      SUM(total_amount) as lifetime_value
    FROM
      orders
    WHERE
      order_date >= DATE_SUB(CURDATE(), INTERVAL 1 YEAR)
    GROUP BY
      customer_id
    HAVING
      lifetime_value >= 1000
  ),
  recent_orders AS (
    SELECT
      customer_id,
      order_id,
      total_amount,
      order_date
    FROM
      orders
    WHERE
      order_date >= DATE_SUB(CURDATE(), INTERVAL 90 DAY)
  ),
  customer_segments AS (
    SELECT
      hvc.customer_id,
      hvc.lifetime_value,
      COUNT(ro.order_id) as recent_order_count,
      AVG(ro.total_amount) as avg_recent_order,
      CASE
        WHEN hvc.lifetime_value >= 5000 THEN 'VIP'
        WHEN hvc.lifetime_value >= 2000 THEN 'Premium'
        ELSE 'Standard'
      END as customer_tier
    FROM
      high_value_customers hvc
      LEFT JOIN recent_orders ro ON hvc.customer_id = ro.customer_id
    GROUP BY
      hvc.customer_id,
      hvc.lifetime_value
  )
SELECT
  cs.customer_tier,
  COUNT(*) as customer_count,
  AVG(cs.lifetime_value) as avg_lifetime_value,
  AVG(cs.recent_order_count) as avg_recent_orders
FROM
  customer_segments cs
GROUP BY
  cs.customer_tier
ORDER BY
  avg_lifetime_value DESC;

-- Recursive CTE: Organization hierarchy
WITH RECURSIVE
  org_chart AS (
    -- Base case: CEO and top-level executives
    SELECT
      employee_id,
      name,
      title,
      manager_id,
      salary,
      0 as level,
      CAST(name AS CHAR(1000)) as reporting_chain,
      JSON_ARRAY(employee_id) as management_path
    FROM
      employees
    WHERE
      manager_id IS NULL
    UNION ALL
    -- Recursive case: Direct reports
    SELECT
      e.employee_id,
      e.name,
      e.title,
      e.manager_id,
      e.salary,
      oc.level + 1,
      CONCAT(oc.reporting_chain, ' → ', e.name) as reporting_chain,
      JSON_ARRAY_APPEND(oc.management_path, '$', e.employee_id) as management_path
    FROM
      employees e
      INNER JOIN org_chart oc ON e.manager_id = oc.employee_id
    WHERE
      oc.level < 6 -- Prevent runaway recursion
  )
SELECT
  level,
  COUNT(*) as employee_count,
  AVG(salary) as avg_salary_by_level,
  MIN(salary) as min_salary,
  MAX(salary) as max_salary,
  JSON_ARRAYAGG(name) as employees_at_level
FROM
  org_chart
GROUP BY
  level
ORDER BY
  level;

-- Recursive CTE: Category tree with JSON paths
WITH RECURSIVE
  category_tree AS (
    -- Root categories
    SELECT
      category_id,
      category_name,
      parent_category_id,
      0 as depth,
      category_name as category_path,
      JSON_OBJECT(
        'id',
        category_id,
        'name',
        category_name,
        'children',
        JSON_ARRAY()
      ) as tree_structure
    FROM
      categories
    WHERE
      parent_category_id IS NULL
    UNION ALL
    -- Child categories  
    SELECT
      c.category_id,
      c.category_name,
      c.parent_category_id,
      ct.depth + 1,
      CONCAT(ct.category_path, ' / ', c.category_name),
      JSON_OBJECT(
        'id',
        c.category_id,
        'name',
        c.category_name,
        'parent_path',
        ct.category_path,
        'depth',
        ct.depth + 1
      ) as tree_structure
    FROM
      categories c
      INNER JOIN category_tree ct ON c.parent_category_id = ct.category_id
    WHERE
      ct.depth < 4
  )
SELECT
  category_id,
  category_name,
  depth,
  category_path,
  tree_structure,
  JSON_LENGTH(tree_structure->'$.children') as direct_children_count
FROM
  category_tree
ORDER BY
  depth,
  category_path;

-- Basic window functions
SELECT
  product_id,
  product_name,
  category,
  price,
  -- Ranking functions
  ROW_NUMBER() OVER (
    ORDER BY
      price DESC
  ) as price_rank,
  RANK() OVER (
    ORDER BY
      price DESC
  ) as price_rank_with_ties,
  DENSE_RANK() OVER (
    ORDER BY
      price DESC
  ) as dense_price_rank,
  PERCENT_RANK() OVER (
    ORDER BY
      price
  ) as price_percentile,
  CUME_DIST() OVER (
    ORDER BY
      price
  ) as cumulative_distribution,
  NTILE(4) OVER (
    ORDER BY
      price
  ) as price_quartile,
  -- Partition-based rankings
  ROW_NUMBER() OVER (
    PARTITION BY category
    ORDER BY
      price DESC
  ) as category_rank,
  RANK() OVER (
    PARTITION BY category
    ORDER BY
      price DESC
  ) as category_rank_ties
FROM
  products
WHERE
  status = 'active'
ORDER BY
  category,
  price DESC;

-- Window functions with frames
SELECT
  sale_date,
  daily_revenue,
  -- Running totals and averages
  SUM(daily_revenue) OVER (
    ORDER BY
      sale_date ROWS BETWEEN UNBOUNDED PRECEDING
      AND CURRENT ROW
  ) as running_total,
  AVG(daily_revenue) OVER (
    ORDER BY
      sale_date ROWS BETWEEN 6 PRECEDING
      AND CURRENT ROW
  ) as seven_day_moving_avg,
  -- Range-based window (last 30 days)
  SUM(daily_revenue) OVER (
    ORDER BY
      sale_date RANGE BETWEEN INTERVAL 29 DAY PRECEDING
      AND CURRENT ROW
  ) as thirty_day_total,
  -- Lead and lag
  LAG(daily_revenue, 1) OVER (
    ORDER BY
      sale_date
  ) as prev_day_revenue,
  LAG(daily_revenue, 7) OVER (
    ORDER BY
      sale_date
  ) as same_day_last_week,
  LEAD(daily_revenue, 1) OVER (
    ORDER BY
      sale_date
  ) as next_day_revenue,
  -- First and last values in window
  FIRST_VALUE(daily_revenue) OVER (
    ORDER BY
      sale_date ROWS BETWEEN 6 PRECEDING
      AND CURRENT ROW
  ) as first_in_week,
  LAST_VALUE(daily_revenue) OVER (
    ORDER BY
      sale_date ROWS BETWEEN CURRENT ROW
      AND 6 FOLLOWING
  ) as last_in_future_week
FROM
  daily_sales
WHERE
  sale_date >= DATE_SUB(CURDATE(), INTERVAL 90 DAY)
ORDER BY
  sale_date;

-- Named window definitions
SELECT
  employee_id,
  name,
  department,
  salary,
  hire_date,
  -- Using named windows
  RANK() OVER dept_salary_window as dept_salary_rank,
  DENSE_RANK() OVER company_salary_window as company_salary_rank,
  ROW_NUMBER() OVER dept_tenure_window as dept_tenure_rank,
  -- Aggregate functions with windows
  AVG(salary) OVER dept_salary_window as dept_avg_salary,
  COUNT(*) OVER dept_salary_window as dept_employee_count,
  -- Complex expressions with windows
  salary - AVG(salary) OVER dept_salary_window as salary_vs_dept_avg,
  (salary - MIN(salary) OVER company_salary_window) / (
    MAX(salary) OVER company_salary_window - MIN(salary) OVER company_salary_window
  ) as salary_normalized
FROM
  employees
WHERE
  status = 'active'
WINDOW
  dept_salary_window AS (
    PARTITION BY department
    ORDER BY
      salary DESC
  ),
  company_salary_window AS (
    ORDER BY
      salary DESC ROWS BETWEEN UNBOUNDED PRECEDING
      AND UNBOUNDED FOLLOWING
  ),
  dept_tenure_window AS (
    PARTITION BY department
    ORDER BY
      hire_date
  )
ORDER BY
  department,
  salary DESC;

-- Complex CTE with window functions and JSON
WITH RECURSIVE
  sales_hierarchy AS (
    -- Territory managers
    SELECT
      employee_id,
      name,
      territory_id,
      manager_id,
      0 as level,
      JSON_ARRAY(territory_id) as territory_chain
    FROM
      sales_team
    WHERE
      manager_id IS NULL
    UNION ALL
    -- Sales representatives
    SELECT
      st.employee_id,
      st.name,
      st.territory_id,
      st.manager_id,
      sh.level + 1,
      JSON_ARRAY_APPEND(sh.territory_chain, '$', st.territory_id) as territory_chain
    FROM
      sales_team st
      INNER JOIN sales_hierarchy sh ON st.manager_id = sh.employee_id
    WHERE
      sh.level < 3
  ),
  sales_performance AS (
    SELECT
      sh.employee_id,
      sh.name,
      sh.level,
      sh.territory_chain,
      COALESCE(SUM(s.amount), 0) as total_sales,
      COUNT(s.sale_id) as sale_count,
      -- Window functions within CTE
      SUM(COALESCE(s.amount, 0)) OVER (
        PARTITION BY JSON_EXTRACT(sh.territory_chain, '$[0]')
        ORDER BY
          sh.level,
          COALESCE(SUM(s.amount), 0) DESC
      ) as territory_running_total,
      RANK() OVER (
        PARTITION BY sh.level
        ORDER BY
          COALESCE(SUM(s.amount), 0) DESC
      ) as performance_rank_by_level
    FROM
      sales_hierarchy sh
      LEFT JOIN sales s ON sh.employee_id = s.salesperson_id
      AND s.sale_date >= DATE_SUB(CURDATE(), INTERVAL 1 YEAR)
    GROUP BY
      sh.employee_id,
      sh.name,
      sh.level,
      sh.territory_chain
  )
SELECT
  sp.name,
  sp.level,
  sp.total_sales,
  sp.performance_rank_by_level,
  JSON_PRETTY(sp.territory_chain) as territory_structure,
  -- Final window functions on CTE results
  PERCENT_RANK() OVER (
    PARTITION BY sp.level
    ORDER BY
      sp.total_sales
  ) as performance_percentile,
  sp.total_sales - AVG(sp.total_sales) OVER (PARTITION BY sp.level) as vs_level_average
FROM
  sales_performance sp
WHERE
  sp.total_sales > 0
ORDER BY
  sp.level,
  sp.performance_rank_by_level;