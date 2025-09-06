package sqlfmt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostgreSQLFormatter_Format(t *testing.T) {
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
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats short CREATE TABLE", func(t *testing.T) {
		query := "CREATE TABLE items (a INT PRIMARY KEY, b TEXT);"
		exp := "CREATE TABLE items (a INT PRIMARY KEY, b TEXT);"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
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
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_DML(t *testing.T) {
	t.Run("formats INSERT", func(t *testing.T) {
		query := "INSERT INTO customers (id, balance, address, city) VALUES (12, -123.4, 'Skagen 2111', 'Stv');"
		exp := Dedent(`
            INSERT INTO
              customers (id, balance, address, city)
            VALUES
              (12, -123.4, 'Skagen 2111', 'Stv');
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats UPDATE", func(t *testing.T) {
		query := "UPDATE users SET name = 'John Doe', active = false WHERE id = 1;"
		exp := Dedent(`
            UPDATE
              users
            SET
              name = 'John Doe',
              active = false
            WHERE
              id = 1;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_Initialization(t *testing.T) {
	t.Run("creates formatter with default config", func(t *testing.T) {
		cfg := NewDefaultConfig().WithLang(PostgreSQL)
		formatter := NewPostgreSQLFormatter(cfg)
		require.NotNil(t, formatter)
		require.NotNil(t, formatter.cfg)
		require.Equal(t, PostgreSQL, formatter.cfg.Language)
	})

	t.Run("creates tokenizer config", func(t *testing.T) {
		config := NewPostgreSQLTokenizerConfig()
		require.NotNil(t, config)
		require.NotEmpty(t, config.ReservedWords)
		require.NotEmpty(t, config.ReservedTopLevelWords)
		require.Contains(t, config.LineCommentTypes, "--")
		require.Contains(t, config.StringTypes, "$$")
	})
}

func TestPostgreSQLFormatter_WithColorConfig(t *testing.T) {
	t.Run("formats with color config", func(t *testing.T) {
		cfg := &Config{
			Language:    PostgreSQL,
			Indent:      DefaultIndent,
			ColorConfig: NewDefaultColorConfig(),
		}

		formatter := NewPostgreSQLFormatter(cfg)
		result := formatter.Format("SELECT id FROM users;")

		require.NotEmpty(t, result)
		require.Contains(t, result, "SELECT")
		require.Contains(t, result, "FROM")
	})
}

func ExamplePostgreSQLFormatter_Format() {
	cfg := NewDefaultConfig().WithLang(PostgreSQL)
	formatter := NewPostgreSQLFormatter(cfg)

	query := "SELECT id FROM users WHERE active = true;"
	result := formatter.Format(query)
	fmt.Println(result)
	// Output:
	// SELECT
	//   id
	// FROM
	//   users
	// WHERE
	//   active = true;
}
