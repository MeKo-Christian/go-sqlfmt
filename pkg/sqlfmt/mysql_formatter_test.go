package sqlfmt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMySQLFormatter_Format(t *testing.T) {
	t.Run("formats simple SELECT", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE active = true;"
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
		query := "CREATE TABLE items (a INT PRIMARY KEY, b TEXT);"
		exp := "CREATE TABLE items (a INT PRIMARY KEY, b TEXT);"
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
}