package sqlfmt

import (
	"strings"
	"testing"
)

// BenchmarkFormatSmall benchmarks formatting of small queries (< 100 characters).
func BenchmarkFormatSmall(b *testing.B) {
	query := "SELECT id, name FROM users WHERE active = true ORDER BY name"
	b.ResetTimer()
	for range b.N {
		Format(query)
	}
}

// BenchmarkFormatMedium benchmarks formatting of medium queries (100-1000 characters).
func BenchmarkFormatMedium(b *testing.B) {
	query := `SELECT
	u.id,
	u.username,
	u.email,
	p.first_name,
	p.last_name,
	p.created_at,
	CASE
		WHEN u.status = 'active' THEN 'Active User'
		WHEN u.status = 'inactive' THEN 'Inactive User'
		ELSE 'Unknown Status'
	END as status_description
FROM users u
LEFT JOIN profiles p ON u.id = p.user_id
LEFT JOIN user_permissions up ON u.id = up.user_id
WHERE u.active = true
AND u.email_verified = true
AND p.created_at > '2023-01-01'
AND up.permission_level >= 5
ORDER BY u.username ASC, p.created_at DESC
LIMIT 100`
	b.ResetTimer()
	for range b.N {
		Format(query)
	}
}

// BenchmarkFormatLarge benchmarks formatting of large queries (1000-10000 characters).
func BenchmarkFormatLarge(b *testing.B) {
	query := `-- Complex PostgreSQL query with multiple CTEs and window functions
WITH RECURSIVE employee_hierarchy AS (
	SELECT
		emp_id,
		name,
		manager_id,
		department,
		salary,
		hire_date,
		0 as level,
		ARRAY[emp_id] as path
	FROM employees
	WHERE manager_id IS NULL
	UNION ALL
	SELECT
		e.emp_id,
		e.name,
		e.manager_id,
		e.department,
		e.salary,
		e.hire_date,
		eh.level + 1,
		eh.path || e.emp_id
	FROM employees e
	JOIN employee_hierarchy eh ON e.manager_id = eh.emp_id
	WHERE eh.level < 10
),
department_stats AS (
	SELECT
		department,
		COUNT(*) as total_employees,
		AVG(salary) as avg_salary,
		MIN(salary) as min_salary,
		MAX(salary) as max_salary,
		SUM(salary) as total_salary,
		STDDEV(salary) as salary_stddev,
		PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY salary) as median_salary,
		STRING_AGG(name, ', ' ORDER BY salary DESC) as top_earners
	FROM employee_hierarchy
	GROUP BY department
),
ranked_employees AS (
	SELECT
		eh.*,
		ds.avg_salary,
		ds.median_salary,
		RANK() OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as dept_salary_rank,
		DENSE_RANK() OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as dept_dense_rank,
		ROW_NUMBER() OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as dept_row_num,
		LAG(eh.salary) OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as next_lower_salary,
		LEAD(eh.salary) OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as next_higher_salary,
		COUNT(*) OVER (PARTITION BY eh.department) as dept_size,
		SUM(eh.salary) OVER (PARTITION BY eh.department) as dept_total_salary
	FROM employee_hierarchy eh
	JOIN department_stats ds ON eh.department = ds.department
)
SELECT
	re.name,
	re.department,
	re.level,
	re.salary,
	re.avg_salary,
	re.median_salary,
	re.dept_salary_rank,
	re.dept_dense_rank,
	re.dept_row_num,
	re.next_lower_salary,
	re.next_higher_salary,
	re.dept_size,
	re.dept_total_salary,
	CASE
		WHEN re.salary > re.avg_salary THEN 'Above Average'
		WHEN re.salary = re.avg_salary THEN 'Average'
		ELSE 'Below Average'
	END as salary_category,
	EXTRACT(YEAR FROM AGE(CURRENT_DATE, re.hire_date)) as years_of_service,
	re.path as hierarchy_path
FROM ranked_employees re
WHERE re.salary > re.median_salary
ORDER BY
	re.department,
	re.salary DESC,
	re.name`
	b.ResetTimer()
	for range b.N {
		Format(query)
	}
}

// BenchmarkFormatVeryLarge benchmarks formatting of very large queries (> 10000 characters).
func BenchmarkFormatVeryLarge(b *testing.B) {
	// Build a very large query by repeating complex patterns
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`-- Very large complex query with multiple complex subqueries
SELECT
	main_query.user_id,
	main_query.total_orders,
	main_query.total_spent,
	main_query.avg_order_value,
	main_query.first_order_date,
	main_query.last_order_date,
	main_query.customer_lifetime_value,
	main_query.customer_segment,
	user_details.full_name,
	user_details.email,
	user_details.registration_date,
	user_details.last_login,
	user_details.account_status,
	order_summary.most_frequent_category,
	order_summary.favorite_payment_method,
	order_summary.preferred_shipping_method,
	behavior_metrics.total_page_views,
	behavior_metrics.total_sessions,
	behavior_metrics.avg_session_duration,
	behavior_metrics.bounce_rate,
	behavior_metrics.conversion_rate,
	geographic_info.country,
	geographic_info.city,
	geographic_info.timezone,
	loyalty_info.loyalty_tier,
	loyalty_info.points_balance,
	loyalty_info.points_earned_this_month,
	loyalty_info.next_tier_threshold
FROM (
	-- Main customer summary
	SELECT
		u.id as user_id,
		COUNT(o.id) as total_orders,
		COALESCE(SUM(o.total_amount), 0) as total_spent,
		COALESCE(AVG(o.total_amount), 0) as avg_order_value,
		MIN(o.created_at) as first_order_date,
		MAX(o.created_at) as last_order_date,
		CASE
			WHEN SUM(o.total_amount) > 10000 THEN 'VIP'
			WHEN SUM(o.total_amount) > 1000 THEN 'Gold'
			WHEN SUM(o.total_amount) > 100 THEN 'Silver'
			ELSE 'Bronze'
		END as customer_segment,
		COALESCE(SUM(o.total_amount), 0) * 1.2 as customer_lifetime_value
	FROM users u
	LEFT JOIN orders o ON u.id = o.user_id AND o.status = 'completed'
	GROUP BY u.id
) main_query
LEFT JOIN (
	-- User details
	SELECT
		id,
		CONCAT(first_name, ' ', last_name) as full_name,
		email,
		created_at as registration_date,
		last_login_at as last_login,
		status as account_status
	FROM users
) user_details ON main_query.user_id = user_details.id
LEFT JOIN (
	-- Order summary with complex aggregations
	SELECT
		user_id,
		(
			SELECT category
			FROM order_items oi
			JOIN products p ON oi.product_id = p.id
			WHERE oi.order_id IN (SELECT id FROM orders WHERE user_id = o.user_id)
			GROUP BY category
			ORDER BY COUNT(*) DESC
			LIMIT 1
		) as most_frequent_category,
		(
			SELECT payment_method
			FROM orders
			WHERE user_id = o.user_id
			GROUP BY payment_method
			ORDER BY COUNT(*) DESC
			LIMIT 1
		) as favorite_payment_method,
		(
			SELECT shipping_method
			FROM orders
			WHERE user_id = o.user_id
			GROUP BY shipping_method
			ORDER BY COUNT(*) DESC
			LIMIT 1
		) as preferred_shipping_method
	FROM orders o
	GROUP BY user_id
) order_summary ON main_query.user_id = order_summary.user_id
LEFT JOIN (
	-- User behavior metrics
	SELECT
		user_id,
		COUNT(*) as total_page_views,
		COUNT(DISTINCT session_id) as total_sessions,
		AVG(session_duration) as avg_session_duration,
		SUM(CASE WHEN page_views = 1 THEN 1 ELSE 0 END) * 1.0 / COUNT(*) as bounce_rate,
		SUM(CASE WHEN converted = true THEN 1 ELSE 0 END) * 1.0 / COUNT(*) as conversion_rate
	FROM user_sessions
	GROUP BY user_id
) behavior_metrics ON main_query.user_id = behavior_metrics.user_id
LEFT JOIN (
	-- Geographic information
	SELECT
		user_id,
		country,
		city,
		timezone
	FROM user_addresses ua
	WHERE ua.is_primary = true
) geographic_info ON main_query.user_id = geographic_info.user_id
LEFT JOIN (
	-- Loyalty program information
	SELECT
		user_id,
		tier as loyalty_tier,
		points_balance,
		points_earned_this_month,
		CASE
			WHEN tier = 'Bronze' THEN 1000
			WHEN tier = 'Silver' THEN 5000
			WHEN tier = 'Gold' THEN 15000
			WHEN tier = 'Platinum' THEN 50000
			ELSE 100000
		END as next_tier_threshold
	FROM loyalty_accounts
) loyalty_info ON main_query.user_id = loyalty_info.user_id
WHERE main_query.total_orders > 0
ORDER BY main_query.total_spent DESC, main_query.customer_lifetime_value DESC
LIMIT 1000`)

	query := queryBuilder.String()
	b.ResetTimer()
	for range b.N {
		Format(query)
	}
}

// BenchmarkFormatDeeplyNested benchmarks formatting of deeply nested queries (10+ subqueries).
func BenchmarkFormatDeeplyNested(b *testing.B) {
	query := `SELECT
	level_1.id,
	level_1.name,
	level_1.total_value,
	level_1.avg_sub_value,
	level_1.max_nested_value
FROM (
	-- Level 1: Main aggregation
	SELECT
		parent.id,
		parent.name,
		SUM(level_2.total_value) as total_value,
		AVG(level_2.total_value) as avg_sub_value,
		MAX(level_2.max_nested_value) as max_nested_value
	FROM parent_table parent
	LEFT JOIN (
		-- Level 2: First level subquery
		SELECT
			child.parent_id,
			SUM(level_3.total_value) as total_value,
			MAX(level_3.max_nested_value) as max_nested_value
		FROM child_table child
		LEFT JOIN (
			-- Level 3: Second level subquery
			SELECT
				grandchild.child_id,
				SUM(level_4.total_value) as total_value,
				MAX(level_4.max_nested_value) as max_nested_value
			FROM grandchild_table grandchild
			LEFT JOIN (
				-- Level 4: Third level subquery
				SELECT
					greatgrandchild.grandchild_id,
					SUM(level_5.total_value) as total_value,
					MAX(level_5.max_nested_value) as max_nested_value
				FROM greatgrandchild_table greatgrandchild
				LEFT JOIN (
					-- Level 5: Fourth level subquery
					SELECT
						ggcchild.greatgrandchild_id,
						SUM(level_6.total_value) as total_value,
						MAX(level_6.max_nested_value) as max_nested_value
					FROM ggcchild_table ggcchild
					LEFT JOIN (
						-- Level 6: Fifth level subquery
						SELECT
							gggcchild.ggcchild_id,
							SUM(level_7.total_value) as total_value,
							MAX(level_7.max_nested_value) as max_nested_value
						FROM gggcchild_table gggcchild
						LEFT JOIN (
							-- Level 7: Sixth level subquery
							SELECT
								ggggcchild.gggcchild_id,
								SUM(level_8.total_value) as total_value,
								MAX(level_8.max_nested_value) as max_nested_value
							FROM ggggcchild_table ggggcchild
							LEFT JOIN (
								-- Level 8: Seventh level subquery
								SELECT
									gggggcchild.ggggcchild_id,
									SUM(level_9.total_value) as total_value,
									MAX(level_9.max_nested_value) as max_nested_value
								FROM gggggcchild_table gggggcchild
								LEFT JOIN (
									-- Level 9: Eighth level subquery
									SELECT
										ggggggcchild.gggggcchild_id,
										SUM(level_10.total_value) as total_value,
										MAX(level_10.max_nested_value) as max_nested_value
									FROM ggggggcchild_table ggggggcchild
									LEFT JOIN (
										-- Level 10: Ninth level subquery
										SELECT
											gggggggcchild.ggggggcchild_id,
											SUM(level_11.total_value) as total_value,
											MAX(level_11.max_nested_value) as max_nested_value
										FROM gggggggcchild_table gggggggcchild
										LEFT JOIN (
											-- Level 11: Tenth level subquery (deep nesting)
											SELECT
												ggggggggcchild.gggggggcchild_id,
												SUM(level_12.total_value) as total_value,
												MAX(level_12.max_nested_value) as max_nested_value
											FROM ggggggggcchild_table ggggggggcchild
											LEFT JOIN (
												-- Level 12: Eleventh level subquery (very deep)
												SELECT
													gggggggggcchild.ggggggggcchild_id,
													COUNT(*) as total_value,
													42 as max_nested_value
												FROM gggggggggcchild_table gggggggggcchild
												WHERE active = true
											) level_12 ON ggggggggcchild.id = level_12.ggggggggcchild_id
											WHERE ggggggggcchild.active = true
										) level_11 ON ggggggcchild.id = level_11.gggggggcchild_id
										WHERE ggggggcchild.active = true
									) level_10 ON gggggcchild.id = level_10.ggggggcchild_id
									WHERE gggggcchild.active = true
								) level_9 ON ggggcchild.id = level_9.gggggcchild_id
								WHERE ggggcchild.active = true
							) level_8 ON gggcchild.id = level_8.ggggcchild_id
							WHERE gggcchild.active = true
						) level_7 ON ggchild.id = level_7.gggcchild_id
						WHERE ggchild.active = true
					) level_6 ON gchild.id = level_6.ggchild_id
					WHERE gchild.active = true
				) level_5 ON grandchild.id = level_5.gchild_id
				WHERE grandchild.active = true
			) level_4 ON child.id = level_4.grandchild_id
			WHERE child.active = true
		) level_3 ON parent.id = level_3.child_id
		WHERE parent.active = true
	) level_2 ON parent.id = level_2.parent_id
	WHERE parent.active = true
	GROUP BY parent.id, parent.name
) level_1
WHERE level_1.total_value > 0
ORDER BY level_1.total_value DESC, level_1.max_nested_value DESC`
	b.ResetTimer()
	for range b.N {
		Format(query)
	}
}

// BenchmarkFormatMySQL benchmarks MySQL-specific formatting.
func BenchmarkFormatMySQL(b *testing.B) {
	cfg := &Config{Language: MySQL}
	query := `SELECT
	u.id,
	u.username,
	u.email,
	p.first_name,
	p.last_name,
	CASE
		WHEN u.status = 'active' THEN 'Active User'
		WHEN u.status = 'inactive' THEN 'Inactive User'
		ELSE 'Unknown Status'
	END as status_description
FROM users u
LEFT JOIN profiles p ON u.id = p.user_id
WHERE u.active = ?
AND u.email_verified = ?
AND p.created_at > ?
ORDER BY u.username ASC, p.created_at DESC
LIMIT 100`
	b.ResetTimer()
	for range b.N {
		Format(query, cfg)
	}
}

// BenchmarkFormatPostgreSQL benchmarks PostgreSQL-specific formatting.
func BenchmarkFormatPostgreSQL(b *testing.B) {
	cfg := &Config{Language: PostgreSQL}
	query := `SELECT
	u.id,
	u.username,
	u.email,
	p.first_name,
	p.last_name,
	CASE
		WHEN u.status = 'active' THEN 'Active User'
		WHEN u.status = 'inactive' THEN 'Inactive User'
		ELSE 'Unknown Status'
	END as status_description
FROM users u
LEFT JOIN profiles p ON u.id = p.user_id
WHERE u.active = $1
AND u.email_verified = $2
AND p.created_at > $3
ORDER BY u.username ASC, p.created_at DESC
LIMIT 100`
	b.ResetTimer()
	for range b.N {
		Format(query, cfg)
	}
}

// BenchmarkFormatSQLite benchmarks SQLite-specific formatting.
func BenchmarkFormatSQLite(b *testing.B) {
	cfg := &Config{Language: SQLite}
	query := `SELECT
	u.id,
	u.username,
	u.email,
	p.first_name,
	p.last_name,
	CASE
		WHEN u.status = 'active' THEN 'Active User'
		WHEN u.status = 'inactive' THEN 'Inactive User'
		ELSE 'Unknown Status'
	END as status_description
FROM users u
LEFT JOIN profiles p ON u.id = p.user_id
WHERE u.active = ?
AND u.email_verified = ?
AND p.created_at > ?
ORDER BY u.username ASC, p.created_at DESC
LIMIT 100`
	b.ResetTimer()
	for range b.N {
		Format(query, cfg)
	}
}
