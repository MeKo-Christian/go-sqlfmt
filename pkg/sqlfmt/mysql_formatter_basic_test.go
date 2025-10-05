package sqlfmt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func runFormattingTest(t *testing.T, formatter Formatter, query, expected string) {
	t.Helper()
	result := formatter.Format(query)
	exp := strings.TrimSpace(strings.ReplaceAll(expected, "\t", DefaultIndent))
	require.Equal(t, exp, result)
}

func TestMySQLFormatter_SelectStatements(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

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
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats SELECT with backtick identifiers", func(t *testing.T) {
		query := "SELECT `user_id`, `full_name` FROM `user_table` WHERE `active` = 1;"
		exp := Dedent(`
            SELECT
              ` + "`user_id`" + `,
              ` + "`full_name`" + `
            FROM
              ` + "`user_table`" + `
            WHERE
              ` + "`active`" + ` = 1;
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats SELECT with MySQL comment styles", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE active = 1 # This is a comment\nORDER BY id;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            WHERE
              active = 1 # This is a comment
            ORDER BY
              id;
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats SELECT with placeholders", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE id = ? AND name = ?;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            WHERE
              id = ?
              AND name = ?;
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_InsertStatements(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("formats INSERT with AUTO_INCREMENT", func(t *testing.T) {
		query := "INSERT INTO users (name, email) VALUES ('John', 'john@example.com');"
		exp := Dedent(`
            INSERT INTO
              users (name, email)
            VALUES
              ('John', 'john@example.com');
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_Functions(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("formats basic MySQL functions", func(t *testing.T) {
		query := "SELECT NOW(), CURDATE(), CURTIME(), UNIX_TIMESTAMP(), UUID();"
		exp := Dedent(`
            SELECT
              NOW(),
              CURDATE(),
              CURTIME(),
              UNIX_TIMESTAMP(),
              UUID();
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_Comments(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("handles MySQL versioned comments", func(t *testing.T) {
		query := "SELECT id FROM users WHERE id = 1 /*!40100 AND name = 'test' */;"
		exp := Dedent(`
            SELECT
              id
            FROM
              users
            WHERE
              id = 1 /*!40100 AND name = 'test' */;
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("handles mixed comment types", func(t *testing.T) {
		query := "SELECT id, /* block comment */ name # line comment\nFROM users;"
		exp := Dedent(`
            SELECT
              id,
              /* block comment */
              name # line comment
            FROM
              users;
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_Literals(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("handles hex and bit literals", func(t *testing.T) {
		query := "SELECT 0x1234, 0b1010, X'1234', b'1010';"
		exp := Dedent(`
            SELECT
              0x1234,
              0b1010,
              X'1234',
              b'1010';
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("handles boolean TRUE/FALSE literals", func(t *testing.T) {
		query := "SELECT TRUE, FALSE, true, false FROM dual;"
		exp := Dedent(`
            SELECT
              TRUE,
              FALSE,
              true,
              false
            FROM
              dual;
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("handles double-quoted strings", func(t *testing.T) {
		query := `SELECT "hello", 'world' FROM users;`
		exp := Dedent(`
            SELECT
              "hello",
              'world'
            FROM
              users;
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_Operators(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("formats NULL-safe equality operator", func(t *testing.T) {
		query := "SELECT * FROM users WHERE name <=> 'John';"
		exp := Dedent(`
            SELECT
              *
            FROM
              users
            WHERE
              name <=> 'John';
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats NULL-safe equality in joins", func(t *testing.T) {
		query := "SELECT u.name FROM users u JOIN profiles p ON u.id <=> p.user_id;"
		exp := Dedent(`
            SELECT
              u.name
            FROM
              users u
            JOIN
              profiles p
            ON
              u.id <=> p.user_id;
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats REGEXP and RLIKE operators", func(t *testing.T) {
		query := "SELECT * FROM users WHERE name REGEXP '^J' AND email RLIKE '@example\\.com$';"
		exp := Dedent(`
            SELECT
              *
            FROM
              users
            WHERE
              name REGEXP '^J'
              AND email RLIKE '@example\.com$';
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats NOT REGEXP as single logical unit", func(t *testing.T) {
		query := "SELECT * FROM users WHERE name NOT REGEXP '^J';"
		exp := Dedent(`
            SELECT
              *
            FROM
              users
            WHERE
              name NOT REGEXP '^J';
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats bitwise operators", func(t *testing.T) {
		query := "SELECT a & b, a | b, a ^ b, ~a, a << 2, a >> 1 FROM numbers;"
		exp := Dedent(`
            SELECT
              a & b,
              a | b,
              a ^ b,
              ~a,
              a << 2,
              a >> 1
            FROM
              numbers;
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_JSON(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("formats JSON extraction operators", func(t *testing.T) {
		query := "SELECT data->'$.name', data->>'$.name' FROM users;"
		exp := Dedent(`
            SELECT
              data->'$.name',
              data->>'$.name'
            FROM
              users;
        `)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats JSON extraction chains", func(t *testing.T) {
		query := "SELECT data->'$.user'->'$.name', data->'$.items[0]'->>'$.price' FROM orders;"
		exp := Dedent(`
            SELECT
              data->'$.user'->'$.name',
              data->'$.items[0]'->>'$.price'
            FROM
              orders;
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_Integration(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("comprehensive Phase 2 integration test", func(t *testing.T) {
		query := "SELECT u.id, u.name, u.email, p.bio, p.avatar_url FROM users u " +
			"LEFT JOIN profiles p ON u.id = p.user_id WHERE u.active = true AND u.created_at > '2023-01-01' " +
			"ORDER BY u.created_at DESC LIMIT 10;"
		exp := Dedent(`
            SELECT
              u.id,
              u.name,
              u.email,
              p.bio,
              p.avatar_url
            FROM
              users u
            LEFT JOIN
              profiles p
            ON
              u.id = p.user_id
            WHERE
              u.active = true
              AND u.created_at > '2023-01-01'
            ORDER BY
              u.created_at DESC
            LIMIT
              10;
        `)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_Format_Basic(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("formats short CREATE TABLE", func(t *testing.T) {
		query := basicCreateTable
		exp := basicCreateTable
		runFormattingTest(t, formatter, query, exp)
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
		runFormattingTest(t, formatter, query, exp)
	})
}
