package sqlfmt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	basicSelectQuery = "SELECT id, name FROM users WHERE active = true;"
	basicCreateTable = "CREATE TABLE items (a INT PRIMARY KEY, b TEXT);"
)

func TestMySQLFormatter_Format(t *testing.T) {
	t.Run("formats simple SELECT", func(t *testing.T) {
		query := basicSelectQuery
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            WHERE
              active = true;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats short CREATE TABLE", func(t *testing.T) {
		query := basicCreateTable
		exp := basicCreateTable
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		require.Equal(t, exp, result)
	})

	t.Run("formats long CREATE TABLE", func(t *testing.T) {
		query := "CREATE TABLE items (a INT PRIMARY KEY, b TEXT, c INT NOT NULL, d INT NOT NULL);"
		exp := Dedent(`
            CREATE TABLE items (
              a INT PRIMARY KEY,
              b TEXT,
              c INT NOT NULL,
              d INT NOT NULL
            );
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats SELECT with backtick identifiers", func(t *testing.T) {
		query := "SELECT `user_id`, `full_name` FROM `user_table` WHERE `active` = 1;"
		exp := Dedent("" +
			"SELECT\n" +
			"  `user_id`,\n" +
			"  `full_name`\n" +
			"FROM\n" +
			"  `user_table`\n" +
			"WHERE\n" +
			"  `active` = 1;")
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats SELECT with MySQL comment styles", func(t *testing.T) {
		query := "SELECT id -- line comment\nFROM users # hash comment\nWHERE active = 1;"
		exp := Dedent(`
            SELECT
              id -- line comment
            FROM
              users # hash comment
            WHERE
              active = 1;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats SELECT with placeholders", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE active = ? AND created_at > ?;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            WHERE
              active = ?
              AND created_at > ?;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats INSERT with AUTO_INCREMENT", func(t *testing.T) {
		query := "INSERT INTO users (name, email) VALUES ('John', 'john@example.com');"
		exp := Dedent(`
            INSERT INTO
              users (name, email)
            VALUES
              ('John', 'john@example.com');
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats basic MySQL functions", func(t *testing.T) {
		query := "SELECT CONCAT(first_name, ' ', last_name) AS full_name, LENGTH(email) FROM users;"
		exp := Dedent(`
            SELECT
              CONCAT(first_name, ' ', last_name) AS full_name,
              LENGTH(email)
            FROM
              users;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	// Phase 2 specific tests
	t.Run("handles MySQL versioned comments", func(t *testing.T) {
		query := "SELECT /*! SQL_CALC_FOUND_ROWS */ id, name FROM users WHERE active = 1;"
		exp := Dedent(`
            SELECT
              /*! SQL_CALC_FOUND_ROWS */
              id,
              name
            FROM
              users
            WHERE
              active = 1;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles mixed comment types", func(t *testing.T) {
		query := "SELECT id /* block comment */, name -- line comment\n" +
			"FROM users # hash comment\n" +
			"WHERE /*! STRAIGHT_JOIN */ active = 1;"
		exp := Dedent(`
            SELECT
              id
              /* block comment */
            ,
              name -- line comment
            FROM
              users # hash comment
            WHERE
              /*! STRAIGHT_JOIN */
              active = 1;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles hex and bit literals", func(t *testing.T) {
		query := "SELECT 0xFF as hex_value, 0b1010 as bin_value, id FROM test_table WHERE flags = 0x10;"
		exp := Dedent(`
            SELECT
              0xFF as hex_value,
              0b1010 as bin_value,
              id
            FROM
              test_table
            WHERE
              flags = 0x10;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles boolean TRUE/FALSE literals", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE active = TRUE AND deleted = FALSE;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            WHERE
              active = TRUE
              AND deleted = FALSE;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles double-quoted strings", func(t *testing.T) {
		query := `SELECT 'single quoted', "double quoted", ` + "`backtick quoted`" + ` FROM test;`
		exp := Dedent(`
            SELECT
              'single quoted',
              "double quoted",
              ` + "`backtick quoted`" + `
            FROM
              test;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("comprehensive Phase 2 integration test", func(t *testing.T) {
		query := `SELECT /*! SQL_CALC_FOUND_ROWS */ ` + "`user_id`" + `, "full_name", 0xFF as flags
				  FROM ` + "`user_table`" + ` -- main user table  
				  WHERE active = TRUE # active users only
				  AND flags & 0b1010 > 0 /* bitwise check */
				  AND created_at > ?;`
		exp := Dedent(`
            SELECT
              /*! SQL_CALC_FOUND_ROWS */
              ` + "`user_id`" + `,
              "full_name",
              0xFF as flags
            FROM
              ` + "`user_table`" + ` -- main user table
            WHERE
              active = TRUE # active users only
              AND flags & 0b1010 > 0
              /* bitwise check */
              AND created_at > ?;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	// Phase 4: Operators & Special Tokens Tests

	t.Run("formats JSON extraction operators", func(t *testing.T) {
		query := "SELECT doc->'$.user', data->>'$.name' FROM json_table WHERE settings->'$.enabled' = 'true';"
		exp := Dedent(`
            SELECT
              doc->'$.user',
              data->>'$.name'
            FROM
              json_table
            WHERE
              settings->'$.enabled' = 'true';
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats JSON extraction chains", func(t *testing.T) {
		query := "SELECT doc->'$.a'->>'$.b' AS nested_value FROM json_data;"
		exp := Dedent(`
            SELECT
              doc->'$.a'->>'$.b' AS nested_value
            FROM
              json_data;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats NULL-safe equality operator", func(t *testing.T) {
		query := "SELECT id FROM users WHERE status <=> NULL OR active <=> deleted;"
		exp := Dedent(`
            SELECT
              id
            FROM
              users
            WHERE
              status <=> NULL
              OR active <=> deleted;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats NULL-safe equality in joins", func(t *testing.T) {
		query := "SELECT u.id, p.name FROM users u LEFT JOIN profiles p ON u.profile_id <=> p.id;"
		exp := Dedent(`
            SELECT
              u.id,
              p.name
            FROM
              users u
              LEFT JOIN profiles p ON u.profile_id <=> p.id;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats REGEXP and RLIKE operators", func(t *testing.T) {
		query := "SELECT name FROM users WHERE email REGEXP '^[a-z]+@' AND username RLIKE '[0-9]+$';"
		exp := Dedent(`
            SELECT
              name
            FROM
              users
            WHERE
              email REGEXP '^[a-z]+@'
              AND username RLIKE '[0-9]+$';
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats NOT REGEXP as single logical unit", func(t *testing.T) {
		query := "SELECT col FROM table WHERE col NOT REGEXP '^foo|bar$' AND val NOT RLIKE 'pattern';"
		exp := Dedent(`
            SELECT
              col
            FROM
              table
            WHERE
              col NOT REGEXP '^foo|bar$'
              AND val NOT RLIKE 'pattern';
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats bitwise operators", func(t *testing.T) {
		query := "SELECT flags & 15 AS masked, flags | 32 AS set_bit, flags ^ 7 AS xor_result FROM config;"
		exp := Dedent(`
            SELECT
              flags & 15 AS masked,
              flags | 32 AS set_bit,
              flags ^ 7 AS xor_result
            FROM
              config;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats bitwise shift operators", func(t *testing.T) {
		query := "SELECT value << 2 AS left_shift, value >> 1 AS right_shift, ~flags AS inverted FROM data;"
		exp := Dedent(`
            SELECT
              value << 2 AS left_shift,
              value >> 1 AS right_shift,
              ~ flags AS inverted
            FROM
              data;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("verifies logical OR behavior (not concatenation)", func(t *testing.T) {
		query := "SELECT name FROM users WHERE active = 1 OR deleted = 0 AND (status = 'pending' || status = 'active');"
		exp := Dedent(`
            SELECT
              name
            FROM
              users
            WHERE
              active = 1
              OR deleted = 0
              AND (status = 'pending' || status = 'active');
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("comprehensive Phase 4 integration test", func(t *testing.T) {
		query := `SELECT u.id, 
                  profile->'$.name' AS profile_name,
                  settings->>'$.theme' AS theme,
                  flags & 0xFF AS permissions
                  FROM users u
                  LEFT JOIN user_data d ON u.id <=> d.user_id
                  WHERE u.email REGEXP '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
                  AND u.status NOT REGEXP '^(banned|suspended)$'
                  AND (u.flags | 4) > 0;`
		exp := Dedent(`
            SELECT
              u.id,
              profile->'$.name' AS profile_name,
              settings->>'$.theme' AS theme,
              flags & 0xFF AS permissions
            FROM
              users u
              LEFT JOIN user_data d ON u.id <=> d.user_id
            WHERE
              u.email REGEXP '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
              AND u.status NOT REGEXP '^(banned|suspended)$'
              AND (u.flags | 4) > 0;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

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

	t.Run("formats INSERT IGNORE statement", func(t *testing.T) {
		query := "INSERT IGNORE INTO users (name, email) VALUES ('John', 'john@example.com');"
		exp := Dedent(`
            INSERT IGNORE
              INTO users (name, email)
            VALUES
              ('John', 'john@example.com');
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats REPLACE statement", func(t *testing.T) {
		query := "REPLACE INTO users (id, name, email) VALUES (1, 'John', 'john@example.com');"
		exp := Dedent(`
            REPLACE
              INTO users (id, name, email)
            VALUES
              (1, 'John', 'john@example.com');
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
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
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("comprehensive Phase 5 integration test", func(t *testing.T) {
		query := `-- Phase 5 comprehensive test featuring all new clauses
                  SELECT u.id, u.name, p.settings->'$.theme' AS theme
                  FROM users u
                  LEFT JOIN profiles p ON u.id = p.user_id
                  WHERE u.active = TRUE
                  AND u.email REGEXP '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
                  ORDER BY u.created_at DESC
                  LIMIT 20, 10
                  FOR UPDATE;
                  
                  INSERT IGNORE INTO user_sessions (user_id, token, created_at) 
                  VALUES (?, ?, NOW());
                  
                  REPLACE INTO user_preferences (user_id, preference_key, preference_value)
                  VALUES (1, 'theme', 'dark'), (1, 'language', 'en');`
		exp := Dedent(`
            -- Phase 5 comprehensive test featuring all new clauses
            SELECT
              u.id,
              u.name,
              p.settings->'$.theme' AS theme
            FROM
              users u
              LEFT JOIN profiles p ON u.id = p.user_id
            WHERE
              u.active = TRUE
              AND u.email REGEXP '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
            ORDER BY
              u.created_at DESC
            LIMIT
              20, 10
            FOR UPDATE
            ;

            INSERT IGNORE
              INTO user_sessions (user_id, token, created_at)
            VALUES
              (?, ?, NOW());

            REPLACE
              INTO user_preferences (user_id, preference_key, preference_value)
            VALUES
              (1, 'theme', 'dark'),
              (1, 'language', 'en');
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	// Phase 6: ON DUPLICATE KEY UPDATE Tests

	t.Run("formats basic INSERT with ON DUPLICATE KEY UPDATE", func(t *testing.T) {
		query := "INSERT INTO users (id, name, email) VALUES (1, 'John', 'john@example.com') ON DUPLICATE KEY UPDATE name = VALUES(name);"
		exp := Dedent(`
            INSERT INTO
              users (id, name, email)
            VALUES
              (1, 'John', 'john@example.com')
            ON DUPLICATE KEY UPDATE
              name = VALUES(name);
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats ON DUPLICATE KEY UPDATE with multiple assignments", func(t *testing.T) {
		query := "INSERT INTO products (id, name, price, stock) VALUES (1, 'Product A', 19.99, 100) ON DUPLICATE KEY UPDATE name = VALUES(name), price = VALUES(price), stock = stock + VALUES(stock), updated_at = NOW();"
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
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats INSERT IGNORE with ON DUPLICATE KEY UPDATE", func(t *testing.T) {
		query := "INSERT IGNORE INTO users (username, email, status) VALUES ('john123', 'john@example.com', 'active') ON DUPLICATE KEY UPDATE email = VALUES(email), last_login = NOW();"
		exp := Dedent(`
            INSERT IGNORE
              INTO users (username, email, status)
            VALUES
              ('john123', 'john@example.com', 'active')
            ON DUPLICATE KEY UPDATE
              email = VALUES(email),
              last_login = NOW();
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex ON DUPLICATE KEY UPDATE with expressions", func(t *testing.T) {
		query := "INSERT INTO inventory (product_id, location_id, quantity) VALUES (100, 1, 50) ON DUPLICATE KEY UPDATE quantity = quantity + VALUES(quantity), last_updated = GREATEST(last_updated, NOW()), modifier = CONCAT('system_', UNIX_TIMESTAMP());"
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
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats multiple VALUES with ON DUPLICATE KEY UPDATE", func(t *testing.T) {
		query := "INSERT INTO user_prefs (user_id, pref_key, pref_value) VALUES (1, 'theme', 'dark'), (2, 'lang', 'en'), (3, 'notify', 'true') ON DUPLICATE KEY UPDATE pref_value = VALUES(pref_value), updated_at = NOW();"
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
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
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
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

// Phase 7: CTEs & Window Functions Tests.
func TestMySQLFormatter_CTEs(t *testing.T) {
	t.Run("formats simple WITH query", func(t *testing.T) {
		query := "WITH user_count AS (SELECT COUNT(*) as total FROM users) SELECT total FROM user_count;"
		exp := Dedent(`
            WITH
              user_count AS (
                SELECT
                  COUNT(*) as total
                FROM
                  users
              )
            SELECT
              total
            FROM
              user_count;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats WITH RECURSIVE query", func(t *testing.T) {
		query := Dedent(`
            WITH RECURSIVE employee_tree AS (
                SELECT id, name, manager_id, 1 as level FROM employees WHERE manager_id IS NULL
                UNION ALL
                SELECT e.id, e.name, e.manager_id, et.level + 1 
                FROM employees e JOIN employee_tree et ON e.manager_id = et.id
            ) SELECT * FROM employee_tree ORDER BY level, name;
        `)
		exp := Dedent(`
            WITH RECURSIVE
              employee_tree AS (
                SELECT
                  id,
                  name,
                  manager_id,
                  1 as level
                FROM
                  employees
                WHERE
                  manager_id IS NULL
                UNION ALL
                SELECT
                  e.id,
                  e.name,
                  e.manager_id,
                  et.level + 1
                FROM
                  employees e
                  JOIN employee_tree et ON e.manager_id = et.id
              )
            SELECT
              *
            FROM
              employee_tree
            ORDER BY
              level,
              name;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats multiple CTEs", func(t *testing.T) {
		query := Dedent(`
            WITH active_users AS (SELECT * FROM users WHERE active = true),
                 recent_orders AS (SELECT * FROM orders WHERE created_at > '2023-01-01')
            SELECT u.name, o.total FROM active_users u JOIN recent_orders o ON u.id = o.user_id;
        `)
		exp := Dedent(`
            WITH
              active_users AS (
                SELECT
                  *
                FROM
                  users
                WHERE
                  active = true
              ),
              recent_orders AS (
                SELECT
                  *
                FROM
                  orders
                WHERE
                  created_at > '2023-01-01'
              )
            SELECT
              u.name,
              o.total
            FROM
              active_users u
              JOIN recent_orders o ON u.id = o.user_id;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex recursive CTE with unions", func(t *testing.T) {
		query := Dedent(`
            WITH RECURSIVE category_hierarchy AS (
                SELECT id, name, parent_id, 0 as depth, JSON_ARRAY(id) as path
                FROM categories WHERE parent_id IS NULL
                UNION
                SELECT c.id, c.name, c.parent_id, ch.depth + 1, JSON_ARRAY_APPEND(ch.path, '$', c.id)
                FROM categories c JOIN category_hierarchy ch ON c.parent_id = ch.id
                WHERE NOT JSON_CONTAINS(ch.path, CAST(c.id AS JSON))
            ) SELECT id, name, depth, path FROM category_hierarchy;
        `)
		exp := Dedent(`
            WITH RECURSIVE
              category_hierarchy AS (
                SELECT
                  id,
                  name,
                  parent_id,
                  0 as depth,
                  JSON_ARRAY(id) as path
                FROM
                  categories
                WHERE
                  parent_id IS NULL
                UNION
                SELECT
                  c.id,
                  c.name,
                  c.parent_id,
                  ch.depth + 1,
                  JSON_ARRAY_APPEND(ch.path, '$', c.id)
                FROM
                  categories c
                  JOIN category_hierarchy ch ON c.parent_id = ch.id
                WHERE
                  NOT JSON_CONTAINS(ch.path, CAST(c.id AS JSON))
              )
            SELECT
              id,
              name,
              depth,
              path
            FROM
              category_hierarchy;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestMySQLFormatter_WindowFunctions(t *testing.T) {
	t.Run("formats basic window function with OVER", func(t *testing.T) {
		query := "SELECT name, salary, AVG(salary) OVER () as avg_salary FROM employees;"
		exp := Dedent(`
            SELECT
              name,
              salary,
              AVG(salary) OVER () as avg_salary
            FROM
              employees;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats window function with PARTITION BY", func(t *testing.T) {
		query := "SELECT department, name, salary, AVG(salary) OVER (PARTITION BY department) as dept_avg FROM employees;"
		exp := Dedent(`
            SELECT
              department,
              name,
              salary,
              AVG(salary) OVER (PARTITION BY department) as dept_avg
            FROM
              employees;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats window function with ORDER BY and frame specification", func(t *testing.T) {
		query := "SELECT name, salary, SUM(salary) OVER (ORDER BY salary ROWS BETWEEN UNBOUNDED PRECEDING " +
			"AND CURRENT ROW) as running_total FROM employees;"
		exp := Dedent(`
            SELECT
              name,
              salary,
              SUM(salary) OVER (
                ORDER BY
                  salary ROWS BETWEEN UNBOUNDED PRECEDING
                  AND CURRENT ROW
              ) as running_total
            FROM
              employees;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex window function with RANGE frame", func(t *testing.T) {
		query := "SELECT date, amount, SUM(amount) OVER (ORDER BY date RANGE BETWEEN INTERVAL 7 DAY " +
			"PRECEDING AND CURRENT ROW) as rolling_week FROM transactions;"
		exp := Dedent(`
            SELECT
              date,
              amount,
              SUM(amount) OVER (
                ORDER BY
                  date RANGE BETWEEN INTERVAL 7 DAY PRECEDING
                  AND CURRENT ROW
              ) as rolling_week
            FROM
              transactions;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats named WINDOW definition", func(t *testing.T) {
		query := "SELECT name, salary, AVG(salary) OVER w, SUM(salary) OVER w FROM employees " +
			"WINDOW w AS (PARTITION BY department ORDER BY salary);"
		exp := Dedent(`
            SELECT
              name,
              salary,
              AVG(salary) OVER w,
              SUM(salary) OVER w
            FROM
              employees
            WINDOW
              w AS (
                PARTITION BY department
                ORDER BY
                  salary
              );
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats multiple window functions", func(t *testing.T) {
		query := "SELECT name, salary, ROW_NUMBER() OVER (ORDER BY salary DESC) as rank, " +
			"LAG(salary, 1) OVER (ORDER BY salary DESC) as prev_salary FROM employees;"
		exp := Dedent(`
            SELECT
              name,
              salary,
              ROW_NUMBER() OVER (
                ORDER BY
                  salary DESC
              ) as rank,
              LAG(salary, 1) OVER (
                ORDER BY
                  salary DESC
              ) as prev_salary
            FROM
              employees;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats MySQL-specific window functions", func(t *testing.T) {
		query := "SELECT name, salary, RANK() OVER (ORDER BY salary DESC) as salary_rank, " +
			"DENSE_RANK() OVER (ORDER BY salary DESC) as dense_rank, " +
			"NTILE(4) OVER (ORDER BY salary DESC) as quartile FROM employees;"
		exp := Dedent(`
            SELECT
              name,
              salary,
              RANK() OVER (
                ORDER BY
                  salary DESC
              ) as salary_rank,
              DENSE_RANK() OVER (
                ORDER BY
                  salary DESC
              ) as dense_rank,
              NTILE(4) OVER (
                ORDER BY
                  salary DESC
              ) as quartile
            FROM
              employees;
        `)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

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

	// Phase 8: DDL Essentials Tests
	t.Run("Phase 8: DDL Essentials", func(t *testing.T) {
		t.Run("formats CREATE INDEX with USING BTREE", func(t *testing.T) {
			query := "CREATE INDEX idx_user_email ON users (email) USING BTREE;"
			exp := Dedent(`
				CREATE INDEX
				  idx_user_email ON users (email) USING BTREE;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats CREATE UNIQUE INDEX", func(t *testing.T) {
			query := "CREATE UNIQUE INDEX uk_user_username ON users (username);"
			exp := Dedent(`
				CREATE UNIQUE INDEX
				  uk_user_username ON users (username);
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats CREATE FULLTEXT INDEX", func(t *testing.T) {
			query := "CREATE FULLTEXT INDEX ft_post_content ON posts (title, content);"
			exp := Dedent(`
				CREATE FULLTEXT INDEX
				  ft_post_content ON posts (title, content);
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats CREATE SPATIAL INDEX with USING HASH", func(t *testing.T) {
			query := "CREATE SPATIAL INDEX sp_location ON venues (coordinates) USING HASH;"
			exp := Dedent(`
				CREATE SPATIAL INDEX
				  sp_location ON venues (coordinates) USING HASH;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats multi-column index", func(t *testing.T) {
			query := "CREATE INDEX idx_user_status_created ON users (status, created_at, updated_at) USING BTREE;"
			exp := Dedent(`
				CREATE INDEX
				  idx_user_status_created ON users (status, created_at, updated_at) USING BTREE;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats ALTER TABLE with ALGORITHM and LOCK", func(t *testing.T) {
			query := "ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active', ALGORITHM=INSTANT, LOCK=NONE;"
			exp := Dedent(`
				ALTER TABLE
				  users
				ADD
				  COLUMN status VARCHAR(20) DEFAULT 'active',
				  ALGORITHM = instant,
				  LOCK = none;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats complex ALTER TABLE with multiple options", func(t *testing.T) {
			query := "ALTER TABLE products MODIFY COLUMN price DECIMAL(10,2) NOT NULL, ADD CONSTRAINT chk_price CHECK (price > 0), ALGORITHM=INPLACE, LOCK=SHARED;"
			exp := Dedent(`
				ALTER TABLE
				  products
				MODIFY
				  COLUMN price DECIMAL(10, 2) NOT NULL,
				ADD
				  CONSTRAINT chk_price CHECK (price > 0),
				  ALGORITHM = inplace,
				  LOCK = shared;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats GENERATED ALWAYS AS VIRTUAL", func(t *testing.T) {
			query := "CREATE TABLE orders (id INT PRIMARY KEY, subtotal DECIMAL(10,2), tax_rate DECIMAL(3,4), tax_amount DECIMAL(10,2) GENERATED ALWAYS AS (subtotal * tax_rate) VIRTUAL);"
			exp := Dedent(`
				CREATE TABLE orders (
				  id INT PRIMARY KEY,
				  subtotal DECIMAL(10, 2),
				  tax_rate DECIMAL(3, 4),
				  tax_amount DECIMAL(10, 2) GENERATED ALWAYS AS (subtotal * tax_rate) VIRTUAL
				);
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats GENERATED ALWAYS AS STORED", func(t *testing.T) {
			query := "CREATE TABLE orders (id INT PRIMARY KEY, subtotal DECIMAL(10,2), tax_rate DECIMAL(3,4), total_amount DECIMAL(10,2) GENERATED ALWAYS AS (subtotal + (subtotal * tax_rate)) STORED);"
			exp := Dedent(`
				CREATE TABLE orders (
				  id INT PRIMARY KEY,
				  subtotal DECIMAL(10, 2),
				  tax_rate DECIMAL(3, 4),
				  total_amount DECIMAL(10, 2) GENERATED ALWAYS AS (subtotal + (subtotal * tax_rate)) STORED
				);
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats complex generated column expressions", func(t *testing.T) {
			query := "CREATE TABLE products (id INT, name VARCHAR(100), price DECIMAL(10,2), discount DECIMAL(5,2), final_price DECIMAL(10,2) GENERATED ALWAYS AS (CASE WHEN discount > 0 THEN price - (price * discount / 100) ELSE price END) STORED);"
			exp := Dedent(`
				CREATE TABLE products (
				  id INT,
				  name VARCHAR(100),
				  price DECIMAL(10, 2),
				  discount DECIMAL(5, 2),
				  final_price DECIMAL(10, 2) GENERATED ALWAYS AS (
				    CASE
				      WHEN discount > 0 THEN price - (price * discount / 100)
				      ELSE price
				    END
				  ) STORED
				);
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("comprehensive Phase 8 DDL integration test", func(t *testing.T) {
			query := "CREATE TABLE users (id INT PRIMARY KEY AUTO_INCREMENT, username VARCHAR(50) NOT NULL, email VARCHAR(100) NOT NULL, status VARCHAR(20) DEFAULT 'active', created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, full_name VARCHAR(150) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL); CREATE UNIQUE INDEX uk_username ON users (username) USING BTREE; CREATE INDEX idx_status_created ON users (status, created_at); ALTER TABLE users ADD COLUMN first_name VARCHAR(50), ADD COLUMN last_name VARCHAR(50), ALGORITHM=INSTANT, LOCK=NONE;"
			exp := Dedent(`
				CREATE TABLE users (
				  id INT PRIMARY KEY AUTO_INCREMENT,
				  username VARCHAR(50) NOT NULL,
				  email VARCHAR(100) NOT NULL,
				  status VARCHAR(20) DEFAULT 'active',
				  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				  full_name VARCHAR(150) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL
				);

				CREATE UNIQUE INDEX
				  uk_username ON users (username) USING BTREE;

				CREATE INDEX
				  idx_status_created ON users (status, created_at);

				ALTER TABLE
				  users
				ADD
				  COLUMN first_name VARCHAR(50),
				ADD
				  COLUMN last_name VARCHAR(50),
				  ALGORITHM = instant,
				  LOCK = none;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})
	})

	t.Run("Phase 9: Stored Routines", func(t *testing.T) {
		t.Run("formats basic stored procedure", func(t *testing.T) {
			query := "CREATE PROCEDURE GetUserCount() BEGIN SELECT COUNT(*) FROM users; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  GetUserCount() BEGIN
				    SELECT
				      COUNT(*)
				    FROM
				      users;
				
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats stored procedure with parameters", func(t *testing.T) {
			query := "CREATE PROCEDURE UpdateUserStatus(IN user_id INT, IN new_status VARCHAR(20)) BEGIN UPDATE users SET status = new_status WHERE id = user_id; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  UpdateUserStatus(IN user_id INT, IN new_status VARCHAR(20)) BEGIN
				    UPDATE
				      users
				    SET
				      status = new_status
				    WHERE
				      id = user_id;
				
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats stored function with RETURNS clause", func(t *testing.T) {
			query := "CREATE FUNCTION CalculateDiscount(price DECIMAL(10,2), discount_percent INT) RETURNS DECIMAL(10,2) DETERMINISTIC BEGIN RETURN price - (price * discount_percent / 100); END;"
			exp := Dedent(`
				CREATE FUNCTION
				  CalculateDiscount(price DECIMAL(10, 2), discount_percent INT) RETURNS DECIMAL(10, 2) DETERMINISTIC BEGIN
				    RETURN price - (price * discount_percent / 100);
				
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats procedure with IF/ELSE statements", func(t *testing.T) {
			query := "CREATE PROCEDURE CheckStock(IN product_id INT, OUT stock_status VARCHAR(20)) BEGIN DECLARE stock_count INT; SELECT quantity INTO stock_count FROM inventory WHERE id = product_id; IF stock_count > 100 THEN SET stock_status = 'High'; ELSEIF stock_count > 20 THEN SET stock_status = 'Medium'; ELSE SET stock_status = 'Low'; END IF; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  CheckStock(IN product_id INT, OUT stock_status VARCHAR(20))
				BEGIN
				  DECLARE stock_count INT;
				  
				  SELECT
				    quantity INTO stock_count
				  FROM
				    inventory
				  WHERE
				    id = product_id;
				  
				  IF stock_count > 100 THEN
				    SET
				      stock_status = 'High';
				  ELSEIF stock_count > 20 THEN
				    SET
				      stock_status = 'Medium';
				  ELSE
				    SET
				      stock_status = 'Low';
				  END IF;
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats procedure with WHILE loop", func(t *testing.T) {
			query := "CREATE PROCEDURE GenerateNumbers(IN max_num INT) BEGIN DECLARE counter INT DEFAULT 0; CREATE TEMPORARY TABLE numbers (num INT); WHILE counter < max_num DO INSERT INTO numbers VALUES (counter); SET counter = counter + 1; END WHILE; SELECT * FROM numbers; DROP TEMPORARY TABLE numbers; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  GenerateNumbers(IN max_num INT)
				BEGIN
				  DECLARE counter INT DEFAULT 0;
				  
				  CREATE TEMPORARY TABLE numbers (num INT);
				  
				  WHILE counter < max_num DO
				    INSERT INTO
				      numbers
				    VALUES
				      (counter);
				    
				    SET
				      counter = counter + 1;
				  END WHILE;
				  
				  SELECT
				    *
				  FROM
				    numbers;
				  
				  DROP TEMPORARY TABLE numbers;
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats procedure with LOOP and LEAVE", func(t *testing.T) {
			query := "CREATE PROCEDURE ProcessBatch(IN batch_size INT) BEGIN DECLARE done INT DEFAULT 0; DECLARE counter INT DEFAULT 0; process_loop: LOOP IF counter >= batch_size THEN LEAVE process_loop; END IF; CALL ProcessSingleItem(counter); SET counter = counter + 1; END LOOP; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  ProcessBatch(IN batch_size INT)
				BEGIN
				  DECLARE done INT DEFAULT 0;
				  
				  DECLARE counter INT DEFAULT 0;
				  
				  process_loop: LOOP
				    IF counter >= batch_size THEN
				      LEAVE process_loop;
				    END IF;
				    
				    CALL ProcessSingleItem(counter);
				    
				    SET
				      counter = counter + 1;
				  END LOOP;
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats procedure with REPEAT UNTIL", func(t *testing.T) {
			query := "CREATE PROCEDURE WaitForCondition() BEGIN DECLARE attempts INT DEFAULT 0; REPEAT SET attempts = attempts + 1; CALL CheckCondition(@result); UNTIL @result = 1 OR attempts > 10 END REPEAT; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  WaitForCondition()
				BEGIN
				  DECLARE attempts INT DEFAULT 0;
				  
				  REPEAT
				    SET
				      attempts = attempts + 1;
				    
				    CALL CheckCondition(@result);
				  UNTIL @result = 1
				  OR attempts > 10 END REPEAT;
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats function with characteristics", func(t *testing.T) {
			query := "CREATE FUNCTION GetUserName(user_id INT) RETURNS VARCHAR(100) READS SQL DATA SQL SECURITY DEFINER BEGIN DECLARE user_name VARCHAR(100); SELECT name INTO user_name FROM users WHERE id = user_id; RETURN user_name; END;"
			exp := Dedent(`
				CREATE FUNCTION
				  GetUserName(user_id INT) RETURNS VARCHAR(100) READS SQL DATA SQL SECURITY definer
				BEGIN
				  DECLARE user_name VARCHAR(100);
				  
				  SELECT
				    name INTO user_name
				  FROM
				    users
				  WHERE
				    id = user_id;
				  
				  RETURN user_name;
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats procedure with cursor", func(t *testing.T) {
			query := "CREATE PROCEDURE ProcessAllUsers() BEGIN DECLARE done INT DEFAULT FALSE; DECLARE user_id INT; DECLARE user_cursor CURSOR FOR SELECT id FROM users; DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE; OPEN user_cursor; read_loop: LOOP FETCH user_cursor INTO user_id; IF done THEN LEAVE read_loop; END IF; CALL ProcessUser(user_id); END LOOP; CLOSE user_cursor; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  ProcessAllUsers()
				BEGIN
				  DECLARE done INT DEFAULT FALSE;
				  
				  DECLARE user_id INT;
				  
				  DECLARE user_cursor CURSOR FOR
				  SELECT
				    id
				  FROM
				    users;
				  
				  DECLARE CONTINUE HANDLER FOR NOT FOUND
				  SET
				    done = TRUE;
				  
				  OPEN user_cursor;
				  
				  read_loop: LOOP
				    FETCH user_cursor INTO user_id;
				    
				    IF done THEN
				      LEAVE read_loop;
				    END IF;
				    
				    CALL ProcessUser(user_id);
				  END LOOP;
				  
				  CLOSE user_cursor;
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("handles DELIMITER statements as pass-through", func(t *testing.T) {
			// Note: In real usage, DELIMITER would change how statements are terminated
			// For formatting purposes, we treat it as a pass-through
			query := "DELIMITER $$ CREATE PROCEDURE TestProc() BEGIN SELECT 1; END$$ DELIMITER ;"
			// The formatter should preserve DELIMITER but still format the procedure
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			// Check that the result contains formatted procedure
			require.Contains(t, result, "CREATE PROCEDURE")
			require.Contains(t, result, "BEGIN")
			require.Contains(t, result, "END")
		})

		t.Run("formats nested BEGIN/END blocks", func(t *testing.T) {
			query := "CREATE PROCEDURE NestedBlocks() BEGIN DECLARE x INT; BEGIN DECLARE y INT; SET y = 10; BEGIN DECLARE z INT; SET z = y * 2; END; END; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  NestedBlocks()
				BEGIN
				  DECLARE x INT;
				  
				  BEGIN
				    DECLARE y INT;
				    
				    SET
				      y = 10;
				    
				    BEGIN
				      DECLARE z INT;
				      
				      SET
				        z = y * 2;
				    END;
				  END;
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats complex stored procedure", func(t *testing.T) {
			query := "CREATE PROCEDURE ComplexProc(IN category VARCHAR(50), OUT total DECIMAL(10,2)) BEGIN DECLARE done INT DEFAULT 0; DECLARE prod_price DECIMAL(10,2); DECLARE cur CURSOR FOR SELECT price FROM products WHERE category_name = category; DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1; SET total = 0; OPEN cur; read_loop: LOOP FETCH cur INTO prod_price; IF done THEN LEAVE read_loop; END IF; SET total = total + prod_price; END LOOP; CLOSE cur; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  ComplexProc(IN category VARCHAR(50), OUT total DECIMAL(10, 2))
				BEGIN
				  DECLARE done INT DEFAULT 0;
				  
				  DECLARE prod_price DECIMAL(10, 2);
				  
				  DECLARE cur CURSOR FOR
				  SELECT
				    price
				  FROM
				    products
				  WHERE
				    category_name = category;
				  
				  DECLARE CONTINUE HANDLER FOR NOT FOUND
				  SET
				    done = 1;
				  
				  SET
				    total = 0;
				  
				  OPEN cur;
				  
				  read_loop: LOOP
				    FETCH cur INTO prod_price;
				    
				    IF done THEN
				      LEAVE read_loop;
				    END IF;
				    
				    SET
				      total = total + prod_price;
				  END LOOP;
				  
				  CLOSE cur;
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})
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
              users
              /*! USE INDEX (idx_name) */
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
