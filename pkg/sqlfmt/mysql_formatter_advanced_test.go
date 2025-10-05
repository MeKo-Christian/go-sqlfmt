package sqlfmt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMySQLFormatter_LimitOffset(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("formats LIMIT n OFFSET m syntax", func(t *testing.T) {
		query := "SELECT id, name FROM users ORDER BY created_at DESC LIMIT 10 OFFSET 20;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            ORDER BY
              created_at DESC
            LIMIT
              10 OFFSET 20;
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats LIMIT m, n syntax (MySQL-specific)", func(t *testing.T) {
		query := "SELECT id, name FROM users ORDER BY created_at DESC LIMIT 20, 10;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            ORDER BY
              created_at DESC
            LIMIT
              20, 10;
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_LockingClauses(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("formats FOR UPDATE locking clause", func(t *testing.T) {
		query := "SELECT id, balance FROM accounts WHERE user_id = ? FOR UPDATE;"
		exp := Dedent(`
            SELECT
              id,
              balance
            FROM
              accounts
            WHERE
              user_id = ?
            FOR UPDATE
            ;
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats FOR SHARE locking clause", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE active = 1 FOR SHARE;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            WHERE
              active = 1
            FOR SHARE
            ;
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats legacy LOCK IN SHARE MODE", func(t *testing.T) {
		query := "SELECT id, status FROM orders WHERE customer_id = ? LOCK IN SHARE MODE;"
		exp := Dedent(`
            SELECT
              id,
              status
            FROM
              orders
            WHERE
              customer_id = ?
            LOCK IN SHARE MODE
            ;
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_InsertReplace(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("formats INSERT IGNORE statement", func(t *testing.T) {
		query := "INSERT IGNORE INTO users (name, email) VALUES ('John', 'john@example.com');"
		exp := Dedent(`
            INSERT IGNORE
              INTO users (name, email)
            VALUES
              ('John', 'john@example.com');
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats REPLACE statement", func(t *testing.T) {
		query := "REPLACE INTO users (id, name, email) VALUES (1, 'John', 'john@example.com');"
		exp := Dedent(`
            REPLACE
              INTO users (id, name, email)
            VALUES
              (1, 'John', 'john@example.com');
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats REPLACE with multiple values", func(t *testing.T) {
		query := "REPLACE INTO products (id, name, price) VALUES (1, 'Product A', 19.99), (2, 'Product B', 29.99);"
		exp := Dedent(`
            REPLACE
              INTO products (id, name, price)
            VALUES
              (1, 'Product A', 19.99),
              (2, 'Product B', 29.99);
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_OnDuplicateKeyUpdate(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("formats basic INSERT with ON DUPLICATE KEY UPDATE", func(t *testing.T) {
		query := "INSERT INTO users (id, name, email) VALUES (1, 'John', 'john@example.com') " +
			"ON DUPLICATE KEY UPDATE name = VALUES(name);"
		exp := Dedent(`
            INSERT INTO
              users (id, name, email)
            VALUES
              (1, 'John', 'john@example.com')
            ON DUPLICATE KEY UPDATE
              name = VALUES(name);
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats ON DUPLICATE KEY UPDATE with multiple assignments", func(t *testing.T) {
		query := "INSERT INTO products (id, name, price, stock) VALUES (1, 'Product A', 19.99, 100) " +
			"ON DUPLICATE KEY UPDATE name = VALUES(name), price = VALUES(price), stock = stock + VALUES(stock), " +
			"updated_at = NOW();"
		exp := Dedent(`
            INSERT INTO
              products (id, name, price, stock)
            VALUES
              (1, 'Product A', 19.99, 100)
            ON DUPLICATE KEY UPDATE
              name = VALUES(name),
              price = VALUES(price),
              stock = stock + VALUES(stock),
              updated_at = NOW();
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats INSERT IGNORE with ON DUPLICATE KEY UPDATE", func(t *testing.T) {
		query := "INSERT IGNORE INTO users (username, email, status) VALUES ('john123', 'john@example.com', 'active') " +
			"ON DUPLICATE KEY UPDATE email = VALUES(email), last_login = NOW();"
		exp := Dedent(`
            INSERT IGNORE
              INTO users (username, email, status)
            VALUES
              ('john123', 'john@example.com', 'active')
            ON DUPLICATE KEY UPDATE
              email = VALUES(email),
              last_login = NOW();
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats complex ON DUPLICATE KEY UPDATE with expressions", func(t *testing.T) {
		query := "INSERT INTO inventory (product_id, location_id, quantity) VALUES (100, 1, 50) " +
			"ON DUPLICATE KEY UPDATE quantity = quantity + VALUES(quantity), last_updated = GREATEST(last_updated, NOW()), " +
			"modifier = CONCAT('system_', UNIX_TIMESTAMP());"
		exp := Dedent(`
            INSERT INTO
              inventory (product_id, location_id, quantity)
            VALUES
              (100, 1, 50)
            ON DUPLICATE KEY UPDATE
              quantity = quantity + VALUES(quantity),
              last_updated = GREATEST(last_updated, NOW()),
              modifier = CONCAT('system_', UNIX_TIMESTAMP());
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats multiple VALUES with ON DUPLICATE KEY UPDATE", func(t *testing.T) {
		query := "INSERT INTO user_prefs (user_id, pref_key, pref_value) VALUES (1, 'theme', 'dark'), " +
			"(2, 'lang', 'en'), (3, 'notify', 'true') " +
			"ON DUPLICATE KEY UPDATE pref_value = VALUES(pref_value), " +
			"updated_at = NOW();"
		exp := Dedent(`
            INSERT INTO
              user_prefs (user_id, pref_key, pref_value)
            VALUES
              (1, 'theme', 'dark'),
              (2, 'lang', 'en'),
              (3, 'notify', 'true')
            ON DUPLICATE KEY UPDATE
              pref_value = VALUES(pref_value),
              updated_at = NOW();
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("comprehensive Phase 6 integration test", func(t *testing.T) {
		query := `-- Phase 6 comprehensive test: MySQL upsert with all features
				  INSERT INTO user_analytics (user_id, session_data, visit_count, last_seen) 
				  VALUES (?, '{"page": "home", "source": "direct"}', 1, NOW()) 
				  ON DUPLICATE KEY UPDATE 
				    session_data = VALUES(session_data),
				    visit_count = visit_count + VALUES(visit_count),
				    last_seen = GREATEST(last_seen, VALUES(last_seen)),
				    updated_by = 'system';`
		exp := Dedent(`
            -- Phase 6 comprehensive test: MySQL upsert with all features
            INSERT INTO
              user_analytics (user_id, session_data, visit_count, last_seen)
            VALUES
              (
                ?,
                '{"page": "home", "source": "direct"}',
                1,
                NOW()
              )
            ON DUPLICATE KEY UPDATE
              session_data = VALUES(session_data),
              visit_count = visit_count + VALUES(visit_count),
              last_seen = GREATEST(
                last_seen,
                VALUES(last_seen)
              ),
              updated_by = 'system';
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_Format_Advanced(t *testing.T) {
	// Phase 5: Core Clauses Tests

	t.Run("formats LIMIT n OFFSET m syntax", func(t *testing.T) {
		query := "SELECT id, name FROM users ORDER BY created_at DESC LIMIT 10 OFFSET 20;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            ORDER BY
              created_at DESC
            LIMIT
              10 OFFSET 20;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats LIMIT m, n syntax (MySQL-specific)", func(t *testing.T) {
		query := "SELECT id, name FROM users ORDER BY created_at DESC LIMIT 20, 10;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            ORDER BY
              created_at DESC
            LIMIT
              20, 10;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats FOR UPDATE locking clause", func(t *testing.T) {
		query := "SELECT id, balance FROM accounts WHERE user_id = ? FOR UPDATE;"
		exp := Dedent(`
            SELECT
              id,
              balance
            FROM
              accounts
            WHERE
              user_id = ?
            FOR UPDATE
            ;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats FOR SHARE locking clause", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE active = 1 FOR SHARE;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            WHERE
              active = 1
            FOR SHARE
            ;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats legacy LOCK IN SHARE MODE", func(t *testing.T) {
		query := "SELECT id, status FROM orders WHERE customer_id = ? LOCK IN SHARE MODE;"
		exp := Dedent(`
            SELECT
              id,
              status
            FROM
              orders
            WHERE
              customer_id = ?
            LOCK IN SHARE MODE
            ;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	// Phase 6: MySQL Upsert Tests

	t.Run("formats INSERT with ON DUPLICATE KEY UPDATE", func(t *testing.T) {
		query := "INSERT INTO user_analytics (user_id, session_data, page_views, last_visit) VALUES (?, '{\"page\": \"dashboard\"}', 1, NOW()) ON DUPLICATE KEY UPDATE session_data = JSON_MERGE_PATCH(session_data, VALUES(session_data)), page_views = page_views + VALUES(page_views), last_visit = GREATEST(last_visit, VALUES(last_visit));"
		exp := Dedent(`
            INSERT INTO
              user_analytics (user_id, session_data, page_views, last_visit)
            VALUES
              (?, '{"page": "dashboard"}', 1, NOW())
            ON DUPLICATE KEY UPDATE
              session_data = JSON_MERGE_PATCH(
                session_data,
                VALUES(session_data)
              ),
              page_views = page_views + VALUES(page_views),
              last_visit = GREATEST(
                last_visit,
                VALUES(last_visit)
              );
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats REPLACE statement", func(t *testing.T) {
		query := "REPLACE INTO cache_entries (cache_key, cache_value, expires_at) VALUES ('user:123:profile', '{\"name\": \"John\"}', DATE_ADD(NOW(), INTERVAL 1 HOUR));"
		exp := Dedent(`
            REPLACE
              INTO cache_entries (cache_key, cache_value, expires_at)
            VALUES
              (
                'user:123:profile',
                '{"name": "John"}',
                DATE_ADD(NOW(), INTERVAL 1 HOUR)
              );
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	// Phase 7: Window Functions Tests

	t.Run("formats window functions with JSON operations", func(t *testing.T) {
		query := "SELECT user_id, event_data->'$.timestamp' as timestamp, " +
			"LAG(event_data->'$.value', 1) OVER (PARTITION BY user_id ORDER BY created_at) as prev_value " +
			"FROM events;"
		exp := Dedent(`
            SELECT
              user_id,
              event_data->'$.timestamp' as timestamp,
              LAG(event_data->'$.value', 1) OVER (
                PARTITION BY user_id
                ORDER BY
                  created_at
              ) as prev_value
            FROM
              events;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("comprehensive Phase 7 integration test", func(t *testing.T) {
		query := `WITH RECURSIVE department_hierarchy AS (
            SELECT id, name, parent_id, 0 as level, CAST(id AS CHAR(200)) as path
            FROM departments WHERE parent_id IS NULL
            UNION ALL
            SELECT d.id, d.name, d.parent_id, dh.level + 1, CONCAT(dh.path, '->', d.id) as path
            FROM departments d JOIN department_hierarchy dh ON d.parent_id = dh.id
        ), employee_stats AS (
            SELECT e.*, dh.level as dept_level, dh.path as dept_path,
                   ROW_NUMBER() OVER (PARTITION BY e.department_id ORDER BY e.salary DESC) as dept_rank,
                   AVG(e.salary) OVER (PARTITION BY e.department_id) as dept_avg_salary,
                   LAG(e.hire_date, 1) OVER (PARTITION BY e.department_id ORDER BY e.hire_date) as prev_hire
            FROM employees e JOIN department_hierarchy dh ON e.department_id = dh.id
        )
        SELECT name, department_id, salary, dept_level, dept_rank, 
               ROUND(dept_avg_salary, 2) as avg_sal, prev_hire
        FROM employee_stats 
        WHERE dept_rank <= 3
        ORDER BY dept_level, department_id, dept_rank;`
		exp := Dedent(`
            WITH RECURSIVE
              department_hierarchy AS (
                SELECT
                  id,
                  name,
                  parent_id,
                  0 as level,
                  CAST(id AS CHAR(200)) as path
                FROM
                  departments
                WHERE
                  parent_id IS NULL
                UNION ALL
                SELECT
                  d.id,
                  d.name,
                  d.parent_id,
                  dh.level + 1,
                  CONCAT(dh.path, '->', d.id) as path
                FROM
                  departments d
                  JOIN department_hierarchy dh ON d.parent_id = dh.id
              ),
              employee_stats AS (
                SELECT
                  e.*,
                  dh.level as dept_level,
                  dh.path as dept_path,
                  ROW_NUMBER() OVER (
                    PARTITION BY e.department_id
                    ORDER BY
                      e.salary DESC
                  ) as dept_rank,
                  AVG(e.salary) OVER (PARTITION BY e.department_id) as dept_avg_salary,
                  LAG(e.hire_date, 1) OVER (
                    PARTITION BY e.department_id
                    ORDER BY
                      e.hire_date
                  ) as prev_hire
                FROM
                  employees e
                  JOIN department_hierarchy dh ON e.department_id = dh.id
              )
            SELECT
              name,
              department_id,
              salary,
              dept_level,
              dept_rank,
              ROUND(dept_avg_salary, 2) as avg_sal,
              prev_hire
            FROM
              employee_stats
            WHERE
              dept_rank <= 3
            ORDER BY
              dept_level,
              department_id,
              dept_rank;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	// Phase 12: Final Polish & Edge Cases Tests

	t.Run("formats backtick identifiers with unicode and emoji", func(t *testing.T) {
		query := "SELECT `用户ID`, `имя`, `🚀rocket_field`, `café_name` FROM `表格📊` WHERE `数量` > 100;"
		exp := Dedent(`
            SELECT
              ` + "`用户ID`" + `,
              ` + "`имя`" + `,
              ` + "`🚀rocket_field`" + `,
              ` + "`café_name`" + `
            FROM
              ` + "`表格📊`" + `
            WHERE
              ` + "`数量`" + ` > 100;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats versioned comments preserved verbatim", func(t *testing.T) {
		query := "SELECT /*! STRAIGHT_JOIN */ id, name FROM users /*! USE INDEX (idx_name) */ WHERE /*! SQL_BUFFER_RESULT */ active = 1;"
		exp := Dedent(`
            SELECT
              /*! STRAIGHT_JOIN */
              id,
              name
            FROM
              users /*! USE INDEX (idx_name) */
            WHERE
              /*! SQL_BUFFER_RESULT */
              active = 1;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("keeps NOT REGEXP as single logical unit", func(t *testing.T) {
		query := "SELECT name FROM users WHERE email NOT REGEXP '^admin' AND username NOT RLIKE '^test';"
		exp := Dedent(`
            SELECT
              name
            FROM
              users
            WHERE
              email NOT REGEXP '^admin'
              AND username NOT RLIKE '^test';
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("keeps CONCAT as function (not || operator)", func(t *testing.T) {
		query := "SELECT CONCAT(first_name, ' ', last_name) as full_name, id FROM users WHERE status || status = 'active';"
		exp := Dedent(`
            SELECT
              CONCAT(first_name, ' ', last_name) as full_name,
              id
            FROM
              users
            WHERE
              status || status = 'active';
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles hex/bit literal forms X'ABCD' and B'1010'", func(t *testing.T) {
		query := "SELECT X'41424344' as hex_str, B'1010' as bin_val, 0xFF as hex_num FROM test WHERE flags = X'FF';"
		exp := Dedent(`
            SELECT
              X'41424344' as hex_str,
              B'1010' as bin_val,
              0xFF as hex_num
            FROM
              test
            WHERE
              flags = X'FF';
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestMySQLFormatter_WindowFunctions(t *testing.T) {
	// This function is now empty after breaking down into smaller functions
	// All tests have been moved to:
	// - TestMySQLFormatter_WindowFunctions_Basic
	// - TestMySQLFormatter_WindowFunctions_Advanced
}
