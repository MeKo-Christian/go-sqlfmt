-- MySQL-Specific Operators Test File
-- Tests JSON operators, NULL-safe equality, REGEXP, and bitwise operations
-- JSON operators: -> and ->>
SELECT
  user_id,
  profile->'$.name' as profile_name_json, -- Returns JSON
  profile->>'$.name' as profile_name_text, -- Returns text
  settings->'$.preferences'->'$.theme' as theme_json,
  settings->'$.preferences'->>'$.theme' as theme_text,
  metadata->'$.tags' [0] as first_tag_json,
  metadata->'$.tags' [0]->>'$' as first_tag_text
FROM
  users
WHERE
  profile->'$.status'->>'$' = 'active';

-- Chained JSON operators
SELECT
  event_id,
  data->'$.transaction'->'$.items' [0]->'$.product'->>'$.name' as product_name,
  data->'$.user'->'$.profile'->'$.contact'->>'$.email' as user_email,
  payload->'$.metadata'->'$.tracking'->'$.utm'->>'$.campaign' as campaign_id
FROM
  events
WHERE
  data->'$.type'->>'$.category' IN ('purchase', 'subscription');

-- JSON operators in complex expressions
SELECT
  order_id,
  items->'$[*].price' as all_prices,
  CAST(items->>'$[0].price' AS DECIMAL(10, 2)) + CAST(items->>'$[1].price' AS DECIMAL(10, 2)) as first_two_total,
  CASE
    WHEN JSON_LENGTH(items->'$[*]') > 5 THEN 'bulk'
    WHEN JSON_LENGTH(items->'$[*]') > 1 THEN 'multi'
    ELSE 'single'
  END as order_type
FROM
  orders
WHERE
  JSON_EXTRACT(items, '$[0].category') = '"electronics"';

-- NULL-safe equality operator <=>
SELECT
  u.user_id,
  u.username,
  p.profile_data,
  CASE
    WHEN u.status <=> p.status THEN 'MATCH'
    WHEN u.status <=> NULL THEN 'USER_NULL'
    WHEN p.status <=> NULL THEN 'PROFILE_NULL'
    ELSE 'DIFFERENT'
  END as status_comparison
FROM
  users u
  LEFT JOIN user_profiles p ON u.user_id <=> p.user_id;

-- NULL-safe equality in WHERE clauses
SELECT
  customer_id,
  order_total,
  discount_amount
FROM
  orders
WHERE
  discount_amount <=> NULL -- Only NULL discounts
  OR payment_method <=> 'credit_card' -- NULL-safe string comparison
  OR shipping_address_id <=> billing_address_id;

-- Handle NULL addresses
-- NULL-safe equality with computed expressions
SELECT
  a.account_id,
  a.primary_email,
  b.backup_email,
  a.primary_email <=> b.backup_email as emails_match_null_safe,
  COALESCE(a.last_login, '1900-01-01') <=> COALESCE(b.last_backup, '1900-01-01') as dates_match_with_default
FROM
  accounts a
  LEFT JOIN account_backups b ON a.account_id = b.account_id;

-- REGEXP and RLIKE operators
SELECT
  username,
  email,
  phone
FROM
  users
WHERE
  email REGEXP '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
  AND username REGEXP '^[a-zA-Z0-9_]{3,20}$'
  AND phone RLIKE '^(\+1-?)?[0-9]{3}-?[0-9]{3}-?[0-9]{4}$';

-- NOT REGEXP as single logical unit
SELECT
  product_code,
  description
FROM
  products
WHERE
  product_code NOT REGEXP '^(DISCONTINUED|LEGACY)_'
  AND description NOT RLIKE '(deprecated|obsolete|end.of.life)'
  AND category REGEXP '^(electronics|software|hardware)$';

-- REGEXP with complex patterns
SELECT
  log_entry,
  message,
  CASE
    WHEN message REGEXP 'ERROR.*database.*connection' THEN 'DB_ERROR'
    WHEN message REGEXP 'WARNING.*memory.*usage.*[0-9]+%' THEN 'MEMORY_WARNING'
    WHEN message NOT REGEXP '^(INFO|DEBUG)' THEN 'IMPORTANT'
    ELSE 'ROUTINE'
  END as log_category
FROM
  application_logs
WHERE
  created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
  AND message REGEXP '.*(error|warning|critical|fatal).*'
  AND message NOT REGEXP 'test.*environment';

-- Bitwise operators: &, |, ^, ~, <<, >>
SELECT
  user_id,
  permissions,
  permissions & 1 as can_read, -- Bitwise AND
  permissions & 2 as can_write, -- Check write permission  
  permissions & 4 as can_delete, -- Check delete permission
  permissions | 8 as with_admin, -- Bitwise OR - add admin
  permissions ^ 16 as toggle_super, -- Bitwise XOR - toggle super user
  ~ permissions as inverted_permissions, -- Bitwise NOT
  permissions << 1 as left_shifted, -- Left shift  
  permissions >> 1 as right_shifted -- Right shift
FROM
  user_permissions
WHERE
  (permissions & 7) > 0 -- Has any of read/write/delete
  AND (permissions | 32) <> permissions;

-- Doesn't have flag 32
-- Complex bitwise operations
SELECT
  product_id,
  feature_flags,
  feature_flags & 0xFF as basic_features, -- Mask lower 8 bits
  (feature_flags >> 8) & 0xFF as advanced_features, -- Extract middle 8 bits
  feature_flags | 0x80000000 as with_enabled_flag, -- Set high bit
  feature_flags ^ (feature_flags & - feature_flags) as clear_lowest_bit, -- Clear lowest set bit
  CASE
    WHEN (feature_flags & (feature_flags - 1)) = 0
    AND feature_flags > 0 THEN 'POWER_OF_TWO'
    ELSE 'MULTIPLE_FLAGS'
  END as flag_type
FROM
  product_features
WHERE
  feature_flags > 0;

-- Bitwise operations with JSON
SELECT
  config_id,
  settings->'$.flags' as json_flags,
  CAST(settings->>'$.numeric_flags' AS UNSIGNED) as numeric_flags,
  CAST(settings->>'$.numeric_flags' AS UNSIGNED) & 15 as low_nibble,
  CAST(settings->>'$.numeric_flags' AS UNSIGNED) | CAST(settings->>'$.additional_flags' AS UNSIGNED) as combined_flags
FROM
  system_configurations
WHERE
  CAST(settings->>'$.numeric_flags' AS UNSIGNED) & CAST(settings->>'$.required_flags' AS UNSIGNED) = CAST(settings->>'$.required_flags' AS UNSIGNED);

-- Mixed operators in complex queries
SELECT
  e.event_id,
  e.user_id,
  e.event_data->'$.action'->>'$' as action,
  u.permissions, -- JSON + NULL-safe equality
  CASE
    WHEN e.event_data->'$.user_id'->>'$' <=> CAST(u.user_id AS CHAR) THEN 'VERIFIED'
    ELSE 'UNVERIFIED'
  END as user_verification, -- Bitwise + REGEXP
  CASE
    WHEN u.permissions & 1 > 0
    AND e.event_data->>'$.action' REGEXP '^read_' THEN 'ALLOWED'
    WHEN u.permissions & 2 > 0
    AND e.event_data->>'$.action' REGEXP '^(create|update)_' THEN 'ALLOWED'
    WHEN u.permissions & 4 > 0
    AND e.event_data->>'$.action' REGEXP '^delete_' THEN 'ALLOWED'
    ELSE 'DENIED'
  END as permission_check, -- JSON + Bitwise
  (
    CAST(e.event_data->>'$.flags' AS UNSIGNED) | u.permissions
  ) as effective_permissions
FROM
  events e
  JOIN users u ON e.user_id = u.user_id
WHERE
  e.event_data->>'$.action' NOT REGEXP '^(test|debug)_'
  AND e.user_id <=> u.user_id -- NULL-safe join condition
  AND (
    u.permissions & CAST(
      e.event_data->>'$.required_permission' AS UNSIGNED
    )
  ) > 0
  AND e.created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY);

-- Operators in aggregation and window functions
SELECT
  DATE(created_at) as event_date,
  COUNT(*) as total_events, -- Aggregate with JSON operators
  COUNT(
    CASE
      WHEN data->>'$.status' = 'success' THEN 1
    END
  ) as success_count,
  AVG(CAST(data->>'$.duration_ms' AS UNSIGNED)) as avg_duration, -- Aggregate with bitwise operations
  BIT_OR(CAST(data->>'$.flags' AS UNSIGNED)) as all_flags_combined,
  BIT_AND(CAST(data->>'$.flags' AS UNSIGNED)) as common_flags_only, -- Window functions with operators
  LAG(COUNT(*), 1) OVER (
    ORDER BY
      DATE(created_at)
  ) as prev_day_count,
  COUNT(*) - LAG(COUNT(*), 1) OVER (
    ORDER BY
      DATE(created_at)
  ) as day_change, -- Complex window expression
  SUM(
    CASE
      WHEN data->>'$.priority' REGEXP '^(high|urgent)$' THEN 1
      ELSE 0
    END
  ) OVER (
    ORDER BY
      DATE(created_at) ROWS BETWEEN 6 PRECEDING
      AND CURRENT ROW
  ) as high_priority_7day
FROM
  application_events
WHERE
  created_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
  AND data->>'$.environment' NOT REGEXP '(test|staging)'
  AND (
    CAST(data->>'$.user_id' AS UNSIGNED) <=> NULL
    OR CAST(data->>'$.user_id' AS UNSIGNED) > 0
  )
GROUP BY
  DATE(created_at)
ORDER BY
  event_date;