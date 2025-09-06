package sqlfmt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const testQuery = "SELECT id, name FROM users WHERE active = true;"

func TestPostgreSQLFormatter_Format(t *testing.T) {
	t.Run("formats simple SELECT", func(t *testing.T) {
		query := testQuery
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

func TestPostgreSQLFormatter_DollarQuotes(t *testing.T) {
	t.Run("formats basic dollar-quoted strings", func(t *testing.T) {
		query := "SELECT $$hello world$$ AS message;"
		exp := Dedent(`
            SELECT
              $$hello world$$ AS message;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats tagged dollar-quoted strings", func(t *testing.T) {
		query := "SELECT $tag$hello world$tag$ AS message;"
		exp := Dedent(`
            SELECT
              $tag$hello world$tag$ AS message;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats multi-line dollar-quoted strings", func(t *testing.T) {
		query := Dedent(`
            CREATE FUNCTION test() RETURNS TEXT AS $$
            BEGIN
                RETURN 'Hello, World!';
            END;
            $$ LANGUAGE plpgsql;
        `)
		exp := Dedent(`
            CREATE FUNCTION test() RETURNS TEXT AS $$
            BEGIN
                RETURN 'Hello, World!';
            END;
            $$ LANGUAGE plpgsql;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats tagged dollar-quoted function bodies", func(t *testing.T) {
		query := Dedent(`
            CREATE FUNCTION add_numbers(a int, b int) RETURNS int AS $function$
            BEGIN
                RETURN a + b;
            END;
            $function$ LANGUAGE plpgsql;
        `)
		exp := Dedent(`
            CREATE FUNCTION add_numbers(a int, b int) RETURNS int AS $function$
            BEGIN
                RETURN a + b;
            END;
            $function$ LANGUAGE plpgsql;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles nested quotes inside dollar-quotes", func(t *testing.T) {
		query := `SELECT $body$He said "Hello" and 'Goodbye'$body$ AS message;`
		exp := Dedent(`
            SELECT
              $body$He said "Hello" and 'Goodbye'$body$ AS message;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles dollar signs inside dollar-quotes", func(t *testing.T) {
		query := `SELECT $$Price: $25.99$$ AS message;`
		exp := Dedent(`
            SELECT
              $$Price: $25.99$$ AS message;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles complex JavaScript function body", func(t *testing.T) {
		query := Dedent(`
            CREATE OR REPLACE FUNCTION test_js(str VARCHAR) RETURNS VARCHAR LANGUAGE javascript AS $js$
            return (str.length <= 1 
                ? str : str.substring(0,1) + '_' + test_js(str.substring(1)));
            $js$;
        `)
		exp := Dedent(`
            CREATE
            OR REPLACE FUNCTION test_js(str VARCHAR) RETURNS VARCHAR LANGUAGE javascript AS $js$
            return (str.length <= 1 
                ? str : str.substring(0,1) + '_' + test_js(str.substring(1)));
            $js$;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles multiple dollar-quoted strings", func(t *testing.T) {
		query := `SELECT $tag1$first$tag1$, $tag2$second$tag2$ AS messages;`
		exp := Dedent(`
            SELECT
              $tag1$first$tag1$,
              $tag2$second$tag2$ AS messages;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles empty dollar-quoted strings", func(t *testing.T) {
		query := `SELECT $$$$ AS empty_string;`
		exp := Dedent(`
            SELECT
              $$$$ AS empty_string;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles complex tag names", func(t *testing.T) {
		query := `SELECT $my_tag_123$content$my_tag_123$ AS tagged;`
		exp := Dedent(`
            SELECT
              $my_tag_123$content$my_tag_123$ AS tagged;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_NumberedPlaceholders(t *testing.T) {
	t.Run("formats basic numbered placeholders", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE id = $1 AND active = $2;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            WHERE
              id = $1
              AND active = $2;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats multiple numbered placeholders", func(t *testing.T) {
		query := "INSERT INTO users (name, email, age, active) VALUES ($1, $2, $3, $4);"
		exp := Dedent(`
            INSERT INTO
              users (name, email, age, active)
            VALUES
              ($1, $2, $3, $4);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats high-numbered placeholders", func(t *testing.T) {
		query := "SELECT * FROM users WHERE col1 = $10 AND col2 = $100;"
		exp := Dedent(`
            SELECT
              *
            FROM
              users
            WHERE
              col1 = $10
              AND col2 = $100;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles placeholders in complex expressions", func(t *testing.T) {
		query := "SELECT CASE WHEN age > $1 THEN 'adult' ELSE 'minor' END FROM users WHERE status IN ($2, $3);"
		exp := Dedent(`
            SELECT
              CASE
                WHEN age > $1 THEN 'adult'
                ELSE 'minor'
              END
            FROM
              users
            WHERE
              status IN ($2, $3);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("distinguishes placeholders from dollar-quotes", func(t *testing.T) {
		query := "SELECT $1, $$text with $25.99 price$$ FROM products WHERE price > $2;"
		exp := Dedent(`
            SELECT
              $1,
              $$text with $25.99 price$$
            FROM
              products
            WHERE
              price > $2;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_ParameterSubstitution(t *testing.T) {
	t.Run("substitutes numbered placeholders with 0-based params", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE id = $1 AND active = $2;"
		exp := Dedent(`
            SELECT
              id,
              name
            FROM
              users
            WHERE
              id = 'param1'
              AND active = 'param2';
        `)
		cfg := NewDefaultConfig().WithLang(PostgreSQL)
		cfg.Params = NewListParams([]string{"param0", "'param1'", "'param2'"})
		result := NewPostgreSQLFormatter(cfg).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("substitutes out-of-order placeholders", func(t *testing.T) {
		query := "SELECT $2, $1, $0;"
		exp := Dedent(`
            SELECT
              'third',
              'second',
              'first';
        `)
		cfg := NewDefaultConfig().WithLang(PostgreSQL)
		cfg.Params = NewListParams([]string{"'first'", "'second'", "'third'"})
		result := NewPostgreSQLFormatter(cfg).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles repeated placeholders", func(t *testing.T) {
		query := "SELECT $0, $1, $0, $1;"
		exp := Dedent(`
            SELECT
              'value1',
              'value2',
              'value1',
              'value2';
        `)
		cfg := NewDefaultConfig().WithLang(PostgreSQL)
		cfg.Params = NewListParams([]string{"'value1'", "'value2'"})
		result := NewPostgreSQLFormatter(cfg).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("leaves unreplaced placeholders when params missing", func(t *testing.T) {
		query := "SELECT $0, $1, $2;"
		exp := Dedent(`
            SELECT
              'first',
              $1,
              $2;
        `)
		cfg := NewDefaultConfig().WithLang(PostgreSQL)
		cfg.Params = NewListParams([]string{"'first'"})
		result := NewPostgreSQLFormatter(cfg).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("ignores dollar-quotes during substitution", func(t *testing.T) {
		query := "SELECT $0, $$function body with $2$$ AS func;"
		exp := Dedent(`
            SELECT
              'param1',
              $$function body with $2$$ AS func;
        `)
		cfg := NewDefaultConfig().WithLang(PostgreSQL)
		cfg.Params = NewListParams([]string{"'param1'", "'param2'"})
		result := NewPostgreSQLFormatter(cfg).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles placeholders with sufficient params", func(t *testing.T) {
		query := "SELECT $0, $1, $2;"
		exp := Dedent(`
            SELECT
              'first',
              'second',
              'third';
        `)
		cfg := NewDefaultConfig().WithLang(PostgreSQL)
		cfg.Params = NewListParams([]string{"'first'", "'second'", "'third'"})
		result := NewPostgreSQLFormatter(cfg).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func ExamplePostgreSQLFormatter_Format() {
	cfg := NewDefaultConfig().WithLang(PostgreSQL)
	formatter := NewPostgreSQLFormatter(cfg)

	query := testQuery
	result := formatter.Format(query)
	fmt.Println(result)
	// Output:
	// SELECT
	//   id,
	//   name
	// FROM
	//   users
	// WHERE
	//   active = true;
}
