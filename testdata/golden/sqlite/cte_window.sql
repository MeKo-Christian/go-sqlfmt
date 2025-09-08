-- SQLite CTEs and Window Functions Test File
-- Tests WITH clauses, recursive CTEs, and window functions (SQLite 3.25+/3.28+)
-- Simple CTE
WITH
  active_users AS (
    SELECT
      user_id,
      username,
      email
    FROM
      users
    WHERE
      status = 'active'
      AND last_login >= datetime('now', '-30 days')
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

-- Multiple CTEs with SQLite-specific features  
WITH
  high_value_customers AS (
    SELECT
      customer_id,
      SUM(total_amount) as lifetime_value
    FROM
      orders
    WHERE
      order_date >= date('now', '-1 year')
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
      order_date >= date('now', '-90 days')
  )
SELECT
  hvc.customer_id,
  hvc.lifetime_value,
  COUNT(ro.order_id) as recent_order_count,
  AVG(ro.total_amount) as avg_recent_order
FROM
  high_value_customers hvc
  LEFT JOIN recent_orders ro ON hvc.customer_id = ro.customer_id
GROUP BY
  hvc.customer_id,
  hvc.lifetime_value
ORDER BY
  hvc.lifetime_value DESC;

-- Recursive CTE for hierarchical data
WITH
  RECURSIVE employee_hierarchy(id, name, manager_id, level, path) AS (
    SELECT
      id,
      name,
      manager_id,
      0 as level,
      CAST(name AS TEXT) as path
    FROM
      employees
    WHERE
      manager_id IS NULL
    UNION ALL
    SELECT
      e.id,
      e.name,
      e.manager_id,
      eh.level + 1,
      eh.path || ' -> ' || e.name
    FROM
      employees e
      JOIN employee_hierarchy eh ON e.manager_id = eh.id
    WHERE
      eh.level < 5
  )
SELECT
  id,
  name,
  level,
  path
FROM
  employee_hierarchy
ORDER BY
  level,
  name;

-- CTE with placeholders (all SQLite types)
WITH
  filtered_data AS (
    SELECT
      user_id,
      score,
      category,
      data -> 'profile' ->> 'name' as profile_name
    FROM
      user_scores
    WHERE
      active = ?1
      AND created_at > :min_date
      AND department = @dept
      AND status = $status
      AND score >= ?
    ORDER BY
      score DESC
  )
SELECT
  category,
  COUNT(*) as user_count,
  AVG(score) as avg_score,
  MAX(score) as max_score
FROM
  filtered_data
GROUP BY
  category
ORDER BY
  avg_score DESC;

-- Basic Window Functions
SELECT
  employee_id,
  salary,
  department,
  ROW_NUMBER() OVER (
    ORDER BY
      salary DESC
  ) as overall_rank,
  RANK() OVER (
    PARTITION BY department
    ORDER BY
      salary DESC
  ) as dept_rank,
  DENSE_RANK() OVER (
    PARTITION BY department
    ORDER BY
      salary DESC
  ) as dept_dense_rank
FROM
  employees
ORDER BY
  department,
  salary DESC;

-- Window Functions with LAG/LEAD
SELECT
  id,
  date,
  value,
  category,
  LAG(value, 1) OVER (
    PARTITION BY category
    ORDER BY
      date
  ) as prev_value,
  LEAD(value, 1) OVER (
    PARTITION BY category
    ORDER BY
      date
  ) as next_value,
  LAG(value, 2, 0) OVER (
    PARTITION BY category
    ORDER BY
      date
  ) as prev_2_value
FROM
  measurements
ORDER BY
  category,
  date;

-- Window Functions with Frames (ROWS)
SELECT
  date,
  sales_amount,
  SUM(sales_amount) OVER (
    ORDER BY
      date ROWS BETWEEN UNBOUNDED PRECEDING
      AND CURRENT ROW
  ) as cumulative_sales,
  AVG(sales_amount) OVER (
    ORDER BY
      date ROWS BETWEEN 2 PRECEDING
      AND 2 FOLLOWING
  ) as moving_avg_5day
FROM
  daily_sales
ORDER BY
  date;

-- Window Functions with Frames (RANGE and GROUPS)
SELECT
  product_id,
  sale_date,
  amount,
  SUM(amount) OVER (
    PARTITION BY product_id
    ORDER BY
      sale_date RANGE BETWEEN INTERVAL '7 days' PRECEDING
      AND CURRENT ROW
  ) as weekly_total,
  COUNT(*) OVER (
    PARTITION BY product_id
    ORDER BY
      amount GROUPS BETWEEN 1 PRECEDING
      AND 1 FOLLOWING
  ) as group_count
FROM
  product_sales
ORDER BY
  product_id,
  sale_date;

-- Window Functions with Placeholders
SELECT
  user_id,
  score,
  test_date,
  RANK() OVER (
    PARTITION BY test_date
    ORDER BY
      score DESC
  ) as daily_rank
FROM
  test_scores
WHERE
  test_date >= ?1
  AND score > :min_score
  AND user_id IN (@user_list)
  AND active = $active_flag
ORDER BY
  test_date,
  score DESC
LIMIT
  ?;

-- CTE with Window Functions Combined
WITH
  monthly_sales AS (
    SELECT
      DATE(order_date, 'start of month') as month,
      customer_id,
      SUM(amount) as monthly_total
    FROM
      orders
    WHERE
      order_date >= ?1
    GROUP BY
      DATE(order_date, 'start of month'),
      customer_id
  )
SELECT
  month,
  customer_id,
  monthly_total,
  LAG(monthly_total, 1) OVER (
    PARTITION BY customer_id
    ORDER BY
      month
  ) as prev_month_total,
  monthly_total - LAG(monthly_total, 1) OVER (
    PARTITION BY customer_id
    ORDER BY
      month
  ) as month_change,
  RANK() OVER (
    PARTITION BY month
    ORDER BY
      monthly_total DESC
  ) as monthly_rank
FROM
  monthly_sales
ORDER BY
  month,
  monthly_total DESC;

-- Complex Recursive CTE with Window Functions and All SQLite Features
WITH
  RECURSIVE category_hierarchy AS (
    SELECT
      id,
      name,
      parent_id,
      0 as level,
      CAST(id AS TEXT) as path,
      name as root_category
    FROM
      categories
    WHERE
      parent_id IS NULL
    UNION ALL
    SELECT
      c.id,
      c.name,
      c.parent_id,
      ch.level + 1,
      ch.path || '/' || CAST(c.id AS TEXT),
      ch.root_category
    FROM
      categories c
      JOIN category_hierarchy ch ON c.parent_id = ch.id
    WHERE
      ch.level < :max_depth
  ),
  product_stats AS (
    SELECT
      category_id,
      COUNT(*) as product_count,
      AVG(price) as avg_price,
      SUM(
        CASE
          WHEN data -> 'featured' = 'true' THEN 1
          ELSE 0
        END
      ) as featured_count,
      'Stats for ' || category_id as description
    FROM
      products
    WHERE
      active = 1
      AND price > ?1
      AND data -> 'metadata' ->> 'supplier' IS NOT NULL
    GROUP BY
      category_id
  )
SELECT
  ch.level,
  ch.name,
  ch.path,
  ch.root_category,
  ps.product_count,
  ps.avg_price,
  ps.featured_count,
  ROW_NUMBER() OVER (
    PARTITION BY ch.level
    ORDER BY
      ps.product_count DESC
  ) as level_rank,
  RANK() OVER (
    ORDER BY
      ps.avg_price DESC
  ) as price_rank,
  LAG(ps.product_count, 1) OVER (
    PARTITION BY ch.root_category
    ORDER BY
      ch.level,
      ps.product_count
  ) as prev_count,
  SUM(ps.product_count) OVER (
    PARTITION BY ch.root_category
    ORDER BY
      ch.level ROWS BETWEEN UNBOUNDED PRECEDING
      AND CURRENT ROW
  ) as cumulative_products
FROM
  category_hierarchy ch
  LEFT JOIN product_stats ps ON ch.id = ps.category_id
WHERE
  ch.level <= @max_display_level
  AND (
    ps.product_count IS NULL
    OR ps.product_count > 0
  )
ORDER BY
  ch.root_category,
  ch.level,
  ps.product_count DESC NULLS LAST;

-- SQLite-specific CTE with UPSERT and Window Functions
WITH
  conflict_resolution AS (
    INSERT INTO
      user_activity (user_id, activity_date, activity_count)
    VALUES
      (?1, date('now'), 1)
    ON CONFLICT
      (user_id, activity_date)
    DO UPDATE
    SET
      activity_count = activity_count + 1,
      updated_at = datetime('now') RETURNING user_id,
      activity_date,
      activity_count
  )
SELECT
  user_id,
  activity_date,
  activity_count,
  ROW_NUMBER() OVER (
    ORDER BY
      activity_count DESC
  ) as activity_rank
FROM
  conflict_resolution;

-- Window Functions with Complex Expressions and JSON
SELECT
  user_id,
  data -> 'profile' ->> 'name' as name,
  data -> 'stats' -> 'score' as score,
  data -> 'preferences' ->> 'theme' as theme,
  RANK() OVER (
    PARTITION BY data -> 'preferences' ->> 'theme'
    ORDER BY
      CAST(data -> 'stats' -> 'score' AS INTEGER) DESC
  ) as theme_rank,
  LAG(CAST(data -> 'stats' -> 'score' AS INTEGER), 1, 0) OVER (
    PARTITION BY user_id
    ORDER BY
      created_at
  ) as prev_score,
  'User: ' || data -> 'profile' ->> 'name' || ' (Score: ' || data -> 'stats' -> 'score' || ')' as display_name
FROM
  user_profiles
WHERE
  data -> 'active' = 'true'
  AND CAST(data -> 'stats' -> 'score' AS INTEGER) > :min_score
ORDER BY
  theme,
  CAST(data -> 'stats' -> 'score' AS INTEGER) DESC;