package sqlfmt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMySQLFormatter_Format_Basic(t *testing.T) {
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
              id /* block comment */
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
              AND flags & 0b1010 > 0 /* bitwise check */
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
}
