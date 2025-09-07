-- MySQL Upsert Operations Test File
-- Tests INSERT...ON DUPLICATE KEY UPDATE, REPLACE, and INSERT IGNORE
-- Basic ON DUPLICATE KEY UPDATE
INSERT INTO
  users (id, username, email, status)
VALUES
  (1, 'john_doe', 'john@example.com', 'active')
ON DUPLICATE KEY UPDATE
  email = VALUES(email),
  updated_at = NOW();

-- Multiple value rows with ON DUPLICATE KEY UPDATE
INSERT INTO
  product_inventory (product_id, location_id, quantity, reserved)
VALUES
  (100, 1, 50, 0),
  (101, 1, 75, 5),
  (102, 1, 25, 2)
ON DUPLICATE KEY UPDATE
  quantity = quantity + VALUES(quantity),
  reserved = GREATEST(
    reserved,
    VALUES(reserved)
  ),
  last_updated = CURRENT_TIMESTAMP;

-- Complex ON DUPLICATE KEY UPDATE with expressions
INSERT INTO
  user_statistics (
    user_id,
    login_count,
    last_login,
    total_time_spent,
    favorite_features
  )
VALUES
  (
    ?,
    1,
    NOW(),
    300,
    JSON_ARRAY('dashboard', 'reports')
  )
ON DUPLICATE KEY UPDATE
  login_count = login_count + VALUES(login_count),
  last_login = GREATEST(
    last_login,
    VALUES(last_login)
  ),
  total_time_spent = total_time_spent + VALUES(total_time_spent),
  favorite_features = JSON_MERGE_PATCH(
    favorite_features,
    VALUES(favorite_features)
  ),
  streak_days = CASE
    WHEN DATEDIFF(
      VALUES
(last_login),
        last_login
    ) = 1 THEN streak_days + 1
    WHEN DATEDIFF(
      VALUES
(last_login),
        last_login
    ) > 1 THEN 1
    ELSE streak_days
  END,
  updated_at = NOW();

-- INSERT IGNORE with single row
INSERT IGNORE
  INTO unique_tokens (token_hash, user_id, expires_at, token_type)
VALUES
  (
    SHA2('random_token_123', 256),
    42,
    DATE_ADD(NOW(), INTERVAL 1 HOUR),
    'session'
  );

-- INSERT IGNORE with multiple rows
INSERT IGNORE
  INTO user_preferences (user_id, preference_key, preference_value)
VALUES
  (1, 'theme', 'dark'),
  (1, 'language', 'en'),
  (1, 'timezone', 'UTC'),
  (2, 'theme', 'light'),
  (2, 'language', 'es'),
  (3, 'notifications', 'enabled');

-- REPLACE with single row
REPLACE
  INTO cache_entries (cache_key, cache_value, expires_at, cache_tags)
VALUES
  (
    'user:profile:123',
    '{"name": "John Doe", "email": "john@example.com", "verified": true}',
    DATE_ADD(NOW(), INTERVAL 2 HOUR),
    JSON_ARRAY('user', 'profile', 'active')
  );

-- REPLACE with multiple rows
REPLACE
  INTO daily_summaries (
    summary_date,
    user_id,
    page_views,
    session_duration,
    actions_taken
  )
VALUES
  (
    CURDATE(),
    100,
    45,
    1800,
    JSON_ARRAY('login', 'view_dashboard', 'create_report')
  ),
  (
    CURDATE(),
    101,
    23,
    900,
    JSON_ARRAY('login', 'view_profile')
  ),
  (
    CURDATE(),
    102,
    67,
    3600,
    JSON_ARRAY(
      'login',
      'view_dashboard',
      'edit_settings',
      'logout'
    )
  );

-- Complex REPLACE with JSON and generated values
REPLACE
  INTO user_activity_log (
    user_id,
    activity_date,
    activity_data,
    activity_summary,
    risk_score
  )
VALUES
  (
    ?,
    CURDATE(),
    JSON_OBJECT(
      'login_time',
      NOW(),
      'ip_address',
      '192.168.1.100',
      'user_agent',
      'Mozilla/5.0...',
      'actions',
      JSON_ARRAY('login', 'view_reports', 'export_data')
    ),
    'User performed high-privilege actions',
    CASE
      WHEN JSON_CONTAINS(
        JSON_OBJECT(
          'actions',
          JSON_ARRAY('export_data', 'admin_panel')
        ),
        JSON_EXTRACT(
          VALUES
(activity_data),
            '$.actions[*]'
        )
      ) THEN 'HIGH'
      ELSE 'NORMAL'
    END
  );

-- INSERT ... ON DUPLICATE KEY UPDATE with JSON operations
INSERT INTO
  user_session_tracking (
    session_id,
    user_id,
    session_data,
    page_sequence,
    last_activity
  )
VALUES
  (
    UUID(),
    ?,
    JSON_OBJECT(
      'browser',
      'Chrome/91.0',
      'platform',
      'Windows',
      'started_at',
      NOW(),
      'initial_referrer',
      'https://google.com'
    ),
    JSON_ARRAY('/dashboard'),
    NOW()
  )
ON DUPLICATE KEY UPDATE
  session_data = JSON_MERGE_PATCH(
    session_data,
    JSON_OBJECT(
      'last_seen',
      NOW(),
      'page_count',
      COALESCE(JSON_EXTRACT(session_data, '$.page_count'), 0) + 1
    )
  ),
  page_sequence = JSON_ARRAY_APPEND(page_sequence, '$', '/new-page'),
  last_activity =
VALUES
(last_activity),
  session_duration = TIMESTAMPDIFF(
    SECOND,
    JSON_UNQUOTE(JSON_EXTRACT(session_data, '$.started_at')),
    NOW()
  );

-- Combination: INSERT IGNORE with ON DUPLICATE KEY UPDATE
INSERT IGNORE
  INTO audit_log (
    table_name,
    record_id,
    operation,
    old_values,
    new_values,
    changed_by,
    changed_at
  )
VALUES
  (
    'users',
    123,
    'UPDATE',
    JSON_OBJECT('email', 'old@example.com', 'status', 'inactive'),
    JSON_OBJECT('email', 'new@example.com', 'status', 'active'),
    'system',
    NOW()
  )
ON DUPLICATE KEY UPDATE
  operation = CONCAT(
    operation,
    ',',
    VALUES
(operation)
  ),
  new_values = JSON_MERGE_PATCH(
    new_values,
    VALUES
(new_values)
  ),
  changed_at =
VALUES
(changed_at),
  change_count = COALESCE(change_count, 0) + 1;

-- Upsert with complex WHERE-like logic in UPDATE clause
INSERT INTO
  price_history (
    product_id,
    price,
    effective_date,
    currency,
    price_source
  )
VALUES
  (?, ?, CURDATE(), 'USD', 'api')
ON DUPLICATE KEY UPDATE
  price = CASE
    WHEN
    VALUES
(price) != price
      AND ABS(
        VALUES
(price) - price
      ) / price > 0.1 THEN
    VALUES
(price) -- Only update if price change > 10%
      ELSE price
  END,
  last_updated = CASE
    WHEN
    VALUES
(price) != price THEN NOW()
      ELSE last_updated
  END,
  price_change_count = CASE
    WHEN
    VALUES
(price) != price THEN price_change_count + 1
      ELSE price_change_count
  END,
  price_source =
VALUES
(price_source);