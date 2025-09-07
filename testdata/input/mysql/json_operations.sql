-- MySQL JSON Operations Test File
-- Tests JSON extraction operators and related functionality

-- Basic JSON extraction with -> (returns JSON)
SELECT 
    user_id,
    profile->'$.name' as profile_name,
    profile->'$.contact'->'$.email' as nested_email,
    settings->'$.preferences'->'$.theme' as theme_setting
FROM user_data
WHERE profile->'$.status' = '"active"';

-- JSON unquoting with ->> (returns scalar)
SELECT 
    user_id,
    profile->>'$.name' as name_text,
    profile->>'$.contact.email' as email_address,
    settings->>'$.preferences.language' as user_language
FROM user_data
WHERE profile->>'$.verified' = 'true';

-- Chained JSON operations
SELECT 
    event_id,
    data->'$.user'->'$.profile'->>'$.name' as user_name,
    data->'$.transaction'->'$.items'[0]->>'$.name' as first_item,
    data->'$.metadata'->'$.tracking'->'$.source'->>'$.campaign' as campaign_source
FROM events
WHERE data->'$.type'->>'$.category' = 'purchase';

-- JSON array access
SELECT 
    order_id,
    items->'$[0]'->>'$.name' as first_item_name,
    items->'$[1]'->>'$.price' as second_item_price,
    JSON_LENGTH(items->'$') as item_count,
    items->'$[*]'->>'$.category' as all_categories
FROM order_data
WHERE JSON_LENGTH(items) > 1;

-- JSON path expressions with filters
SELECT 
    product_id,
    features->'$[*]' as all_features,
    features->'$[0].name' as first_feature_name,
    reviews->'$[*].rating' as all_ratings,
    reviews->'$[?(@.rating >= 4)]' as good_reviews
FROM product_catalog
WHERE JSON_EXTRACT(features, '$[0].enabled') = true;

-- JSON functions with extraction operators
SELECT 
    customer_id,
    preferences->'$.notifications' as notification_settings,
    JSON_TYPE(preferences->'$.theme') as theme_type,
    JSON_VALID(preferences->>'$.custom_css') as css_valid,
    CASE 
        WHEN preferences->'$.plan'->>'$.type' = 'premium' 
        THEN preferences->'$.plan'->'$.features'
        ELSE JSON_ARRAY('basic')
    END as available_features
FROM customer_preferences;

-- Complex WHERE clauses with JSON operators
SELECT *
FROM analytics_events
WHERE event_data->'$.user'->>'$.role' IN ('admin', 'moderator')
  AND event_data->'$.timestamp' > JSON_QUOTE(DATE_SUB(NOW(), INTERVAL 1 DAY))
  AND event_data->'$.properties'->'$.utm_source' IS NOT NULL
  AND JSON_CONTAINS(event_data->'$.tags', '"important"');

-- JSON operations in JOINs
SELECT 
    u.user_id,
    u.username,
    p.profile->'$.contact'->>'$.email' as email,
    s.settings->'$.theme'->>'$.mode' as theme_mode
FROM users u
LEFT JOIN user_profiles p ON u.user_id = p.user_id 
    AND p.profile->'$.status'->>'$.active' = 'true'
INNER JOIN user_settings s ON u.user_id = s.user_id
    AND s.settings->'$.version' >= 2.0;

-- JSON aggregation with operators
SELECT 
    category,
    COUNT(*) as product_count,
    JSON_ARRAYAGG(name) as product_names,
    JSON_OBJECTAGG(
        metadata->>'$.sku', 
        metadata->'$.price'->>'$.amount'
    ) as sku_prices,
    AVG(CAST(metadata->'$.rating'->>'$.average' AS DECIMAL(3,2))) as avg_rating
FROM products
WHERE metadata->'$.featured'->>'$.status' = 'active'
GROUP BY category
HAVING COUNT(*) >= 3;

-- JSON modification with extraction (UPDATE context)
UPDATE user_preferences 
SET settings = JSON_SET(
    settings,
    '$.theme', JSON_OBJECT('mode', settings->>'$.theme', 'updated', NOW()),
    '$.notifications.email', CASE 
        WHEN settings->'$.plan'->>'$.type' = 'premium' 
        THEN true 
        ELSE settings->'$.notifications'->>'$.email' 
    END
)
WHERE settings->'$.version' < 3
  AND settings->'$.migrated'->>'$.status' != 'complete';

-- JSON operations with window functions
SELECT 
    user_id,
    session_data->'$.timestamp'->>'$.start' as session_start,
    session_data->'$.actions' as session_actions,
    JSON_LENGTH(session_data->'$.actions') as action_count,
    LAG(JSON_LENGTH(session_data->'$.actions'), 1) OVER (
        PARTITION BY user_id 
        ORDER BY session_data->'$.timestamp'->>'$.start'
    ) as prev_action_count,
    SUM(JSON_LENGTH(session_data->'$.actions')) OVER (
        PARTITION BY user_id 
        ORDER BY session_data->'$.timestamp'->>'$.start'
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) as cumulative_actions
FROM user_sessions
WHERE session_data->'$.timestamp'->>'$.start' >= CURDATE()
ORDER BY user_id, session_data->'$.timestamp'->>'$.start';