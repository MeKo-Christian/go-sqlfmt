-- MySQL Analytics: User engagement and content performance analysis
-- Complex queries with window functions, CTEs, and advanced aggregations

-- User engagement analysis with cohort retention
WITH user_cohorts AS (
    SELECT
        u.id as user_id,
        DATE_FORMAT(u.created_at, '%Y-%m-01') as cohort_month,
        u.created_at as first_seen,
        COUNT(DISTINCT DATE(s.created_at)) as active_days,
        COUNT(s.id) as total_sessions,
        SUM(s.duration_minutes) as total_session_minutes,
        MAX(s.created_at) as last_activity
    FROM users u
    LEFT JOIN user_sessions s ON u.id = s.user_id
        AND s.created_at >= u.created_at
    WHERE u.is_active = TRUE
    GROUP BY u.id, DATE_FORMAT(u.created_at, '%Y-%m-01'), u.created_at
),
cohort_metrics AS (
    SELECT
        cohort_month,
        COUNT(DISTINCT user_id) as cohort_size,
        AVG(active_days) as avg_active_days,
        AVG(total_sessions) as avg_sessions_per_user,
        AVG(total_session_minutes) as avg_minutes_per_user,
        AVG(DATEDIFF(CURRENT_DATE, first_seen)) as avg_account_age_days
    FROM user_cohorts
    GROUP BY cohort_month
),
retention_analysis AS (
    SELECT
        c.cohort_month,
        c.cohort_size,
        -- Calculate retention for each month after signup
        COUNT(DISTINCT CASE WHEN DATEDIFF(s.created_at, u.created_at) BETWEEN 0 AND 30 THEN u.id END) / c.cohort_size * 100 as month_1_retention,
        COUNT(DISTINCT CASE WHEN DATEDIFF(s.created_at, u.created_at) BETWEEN 31 AND 60 THEN u.id END) / c.cohort_size * 100 as month_2_retention,
        COUNT(DISTINCT CASE WHEN DATEDIFF(s.created_at, u.created_at) BETWEEN 61 AND 90 THEN u.id END) / c.cohort_size * 100 as month_3_retention,
        COUNT(DISTINCT CASE WHEN DATEDIFF(s.created_at, u.created_at) BETWEEN 91 AND 180 THEN u.id END) / c.cohort_size * 100 as month_6_retention,
        COUNT(DISTINCT CASE WHEN DATEDIFF(s.created_at, u.created_at) BETWEEN 181 AND 365 THEN u.id END) / c.cohort_size * 100 as month_12_retention
    FROM cohort_metrics c
    JOIN users u ON DATE_FORMAT(u.created_at, '%Y-%m-01') = c.cohort_month
    LEFT JOIN user_sessions s ON u.id = s.user_id
    GROUP BY c.cohort_month, c.cohort_size
)
SELECT
    cohort_month,
    cohort_size,
    ROUND(month_1_retention, 1) as month_1_pct,
    ROUND(month_2_retention, 1) as month_2_pct,
    ROUND(month_3_retention, 1) as month_3_pct,
    ROUND(month_6_retention, 1) as month_6_pct,
    ROUND(month_12_retention, 1) as month_12_pct
FROM retention_analysis
ORDER BY cohort_month DESC;

-- Content performance analysis with engagement metrics
WITH content_engagement AS (
    SELECT
        c.id as content_id,
        c.title,
        c.content_type,
        c.published_at,
        c.author_id,
        COUNT(DISTINCT v.user_id) as unique_views,
        COUNT(v.id) as total_views,
        COUNT(DISTINCT l.user_id) as unique_likes,
        COUNT(DISTINCT cm.user_id) as unique_comments,
        COUNT(DISTINCT s.user_id) as unique_shares,
        AVG(v.time_spent_seconds) as avg_time_spent,
        MAX(v.created_at) as last_viewed
    FROM content c
    LEFT JOIN content_views v ON c.id = v.content_id
    LEFT JOIN content_likes l ON c.id = l.content_id
    LEFT JOIN content_comments cm ON c.id = cm.content_id
    LEFT JOIN content_shares s ON c.id = s.content_id
    WHERE c.status = 'published'
      AND c.published_at >= DATE_SUB(CURRENT_DATE, INTERVAL 90 DAY)
    GROUP BY c.id, c.title, c.content_type, c.published_at, c.author_id
),
engagement_scores AS (
    SELECT
        content_id,
        title,
        content_type,
        published_at,
        author_id,
        unique_views,
        total_views,
        unique_likes,
        unique_comments,
        unique_shares,
        avg_time_spent,
        last_viewed,
        -- Calculate engagement rate (likes + comments + shares) / views
        CASE
            WHEN unique_views > 0 THEN
                (unique_likes + unique_comments + unique_shares) / unique_views * 100
            ELSE 0
        END as engagement_rate_pct,
        -- Calculate virality score
        CASE
            WHEN unique_views > 0 THEN
                unique_shares / unique_views * 100
            ELSE 0
        END as virality_score,
        -- Calculate content quality score
        (unique_likes * 1.0 + unique_comments * 2.0 + unique_shares * 3.0 + avg_time_spent / 60 * 0.5) as quality_score
    FROM content_engagement
),
content_ranking AS (
    SELECT
        *,
        ROW_NUMBER() OVER (ORDER BY quality_score DESC) as overall_rank,
        ROW_NUMBER() OVER (PARTITION BY content_type ORDER BY quality_score DESC) as type_rank,
        NTILE(4) OVER (ORDER BY quality_score DESC) as quality_quartile
    FROM engagement_scores
),
performance_summary AS (
    SELECT
        content_type,
        COUNT(*) as total_content,
        ROUND(AVG(unique_views), 0) as avg_views,
        ROUND(AVG(engagement_rate_pct), 2) as avg_engagement_rate,
        ROUND(AVG(virality_score), 2) as avg_virality,
        ROUND(AVG(quality_score), 2) as avg_quality_score,
        COUNT(CASE WHEN quality_quartile = 1 THEN 1 END) as top_quartile_count
    FROM content_ranking
    GROUP BY content_type
)
SELECT
    content_type,
    total_content,
    avg_views,
    CONCAT(avg_engagement_rate, '%') as avg_engagement_rate,
    CONCAT(avg_virality, '%') as avg_virality,
    avg_quality_score,
    ROUND(top_quartile_count / total_content * 100, 1) as top_quartile_percentage
FROM performance_summary
ORDER BY avg_quality_score DESC;

-- Advanced user segmentation with RFM analysis
WITH user_rfm AS (
    SELECT
        u.id as user_id,
        u.username,
        u.email,
        MAX(o.order_date) as last_order_date,
        DATEDIFF(CURRENT_DATE, MAX(o.order_date)) as recency_days,
        COUNT(o.id) as frequency_orders,
        SUM(o.total_amount) as monetary_total,
        AVG(o.total_amount) as monetary_avg,
        MIN(o.order_date) as first_order_date,
        DATEDIFF(MAX(o.order_date), MIN(o.order_date)) as customer_lifespan_days
    FROM users u
    LEFT JOIN orders o ON u.id = o.customer_id AND o.status = 'completed'
    WHERE u.is_active = TRUE
    GROUP BY u.id, u.username, u.email
),
rfm_scores AS (
    SELECT
        *,
        -- Calculate R score (1-5, 5 being most recent)
        CASE
            WHEN recency_days <= 7 THEN 5
            WHEN recency_days <= 30 THEN 4
            WHEN recency_days <= 90 THEN 3
            WHEN recency_days <= 180 THEN 2
            ELSE 1
        END as r_score,
        -- Calculate F score (1-5, 5 being most frequent)
        CASE
            WHEN frequency_orders >= 10 THEN 5
            WHEN frequency_orders >= 5 THEN 4
            WHEN frequency_orders >= 2 THEN 3
            WHEN frequency_orders = 1 THEN 2
            ELSE 1
        END as f_score,
        -- Calculate M score (1-5, 5 being highest monetary value)
        CASE
            WHEN monetary_total >= 1000 THEN 5
            WHEN monetary_total >= 500 THEN 4
            WHEN monetary_total >= 100 THEN 3
            WHEN monetary_total >= 10 THEN 2
            ELSE 1
        END as m_score,
        -- Calculate RFM score combination
        CONCAT(
            CASE WHEN recency_days <= 7 THEN 5 WHEN recency_days <= 30 THEN 4
                 WHEN recency_days <= 90 THEN 3 WHEN recency_days <= 180 THEN 2 ELSE 1 END,
            CASE WHEN frequency_orders >= 10 THEN 5 WHEN frequency_orders >= 5 THEN 4
                 WHEN frequency_orders >= 2 THEN 3 WHEN frequency_orders = 1 THEN 2 ELSE 1 END,
            CASE WHEN monetary_total >= 1000 THEN 5 WHEN monetary_total >= 500 THEN 4
                 WHEN monetary_total >= 100 THEN 3 WHEN monetary_total >= 10 THEN 2 ELSE 1 END
        ) as rfm_score
    FROM user_rfm
),
segmentation AS (
    SELECT
        *,
        CASE
            WHEN r_score >= 4 AND f_score >= 4 AND m_score >= 4 THEN 'Champions'
            WHEN r_score >= 3 AND f_score >= 3 AND m_score >= 3 THEN 'Loyal Customers'
            WHEN r_score >= 3 AND f_score >= 1 AND m_score >= 1 THEN 'Potential Loyalists'
            WHEN r_score >= 2 AND f_score >= 2 AND m_score >= 2 THEN 'At Risk'
            WHEN r_score >= 2 AND f_score >= 1 AND m_score >= 1 THEN 'Need Attention'
            WHEN r_score >= 1 AND f_score >= 1 AND m_score >= 1 THEN 'Lost'
            ELSE 'New Customers'
        END as customer_segment
    FROM rfm_scores
),
segment_summary AS (
    SELECT
        customer_segment,
        COUNT(*) as customer_count,
        ROUND(AVG(recency_days), 0) as avg_recency_days,
        ROUND(AVG(frequency_orders), 1) as avg_frequency,
        ROUND(AVG(monetary_total), 2) as avg_monetary,
        ROUND(SUM(monetary_total), 2) as segment_revenue,
        ROUND(AVG(r_score), 1) as avg_r_score,
        ROUND(AVG(f_score), 1) as avg_f_score,
        ROUND(AVG(m_score), 1) as avg_m_score
    FROM segmentation
    GROUP BY customer_segment
)
SELECT
    customer_segment,
    customer_count,
    ROUND(customer_count / SUM(customer_count) OVER () * 100, 1) as segment_percentage,
    avg_recency_days,
    avg_frequency,
    avg_monetary,
    segment_revenue,
    ROUND(segment_revenue / SUM(segment_revenue) OVER () * 100, 1) as revenue_percentage,
    avg_r_score,
    avg_f_score,
    avg_m_score
FROM segment_summary
ORDER BY segment_revenue DESC;

-- Complex funnel analysis with conversion rates
WITH funnel_steps AS (
    SELECT
        '1. Registration' as step_name, 1 as step_order, COUNT(*) as users
        FROM users WHERE created_at >= DATE_SUB(CURRENT_DATE, INTERVAL 90 DAY)
    UNION ALL
    SELECT
        '2. Email Verified' as step_name, 2 as step_order, COUNT(*) as users
        FROM users WHERE email_verified = TRUE AND created_at >= DATE_SUB(CURRENT_DATE, INTERVAL 90 DAY)
    UNION ALL
    SELECT
        '3. First Login' as step_name, 3 as step_order, COUNT(DISTINCT u.id) as users
        FROM users u
        JOIN user_sessions s ON u.id = s.user_id
        WHERE u.created_at >= DATE_SUB(CURRENT_DATE, INTERVAL 90 DAY)
    UNION ALL
    SELECT
        '4. Profile Complete' as step_name, 4 as step_order, COUNT(*) as users
        FROM users WHERE profile_completed = TRUE AND created_at >= DATE_SUB(CURRENT_DATE, INTERVAL 90 DAY)
    UNION ALL
    SELECT
        '5. First Purchase' as step_name, 5 as step_order, COUNT(DISTINCT u.id) as users
        FROM users u
        JOIN orders o ON u.id = o.customer_id
        WHERE o.status = 'completed' AND u.created_at >= DATE_SUB(CURRENT_DATE, INTERVAL 90 DAY)
    UNION ALL
    SELECT
        '6. Repeat Purchase' as step_name, 6 as step_order, COUNT(DISTINCT u.id) as users
        FROM users u
        JOIN orders o ON u.id = o.customer_id
        WHERE o.status = 'completed' AND u.created_at >= DATE_SUB(CURRENT_DATE, INTERVAL 90 DAY)
        GROUP BY u.id HAVING COUNT(o.id) >= 2
),
funnel_analysis AS (
    SELECT
        step_name,
        step_order,
        users,
        LAG(users) OVER (ORDER BY step_order) as previous_step_users,
        ROUND(
            users / LAG(users) OVER (ORDER BY step_order) * 100,
            2
        ) as conversion_rate_from_previous,
        ROUND(
            users / FIRST_VALUE(users) OVER (ORDER BY step_order) * 100,
            2
        ) as overall_conversion_rate
    FROM funnel_steps
)
SELECT
    step_name,
    users,
    CASE
        WHEN previous_step_users IS NOT NULL THEN
            CONCAT(users, ' (', conversion_rate_from_previous, '% from previous)')
        ELSE CONCAT(users, ' (starting point)')
    END as users_with_conversion,
    CONCAT(overall_conversion_rate, '%') as overall_conversion
FROM funnel_analysis
ORDER BY step_order;