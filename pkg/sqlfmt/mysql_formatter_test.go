package sqlfmt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	basicSelectQuery    = "SELECT id, name FROM users WHERE active = true;"
	basicCreateTable    = "CREATE TABLE items (a INT PRIMARY KEY, b TEXT);"
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
		query := `SELECT 'single quoted', "double quoted", `+"`backtick quoted`"+` FROM test;`
		exp := Dedent(`
            SELECT
              'single quoted',
              "double quoted",
              `+"`backtick quoted`"+`
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
              `+"`user_id`"+`,
              "full_name",
              0xFF as flags
            FROM
              `+"`user_table`"+` -- main user table  
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
}