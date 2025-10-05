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
		// Test that it can actually format something to verify it's working
		result := formatter.Format("SELECT 1")
		require.NotEmpty(t, result)
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
            CREATE FUNCTION
              test() RETURNS TEXT AS $$
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
            CREATE FUNCTION
              add_numbers(a int, b int) RETURNS int AS $function$
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
            CREATE OR REPLACE FUNCTION
              test_js(str VARCHAR) RETURNS VARCHAR LANGUAGE javascript AS $js$
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
	formatter := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL))

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
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats multiple numbered placeholders", func(t *testing.T) {
		query := "INSERT INTO users (name, email, age, active) VALUES ($1, $2, $3, $4);"
		exp := Dedent(`
            INSERT INTO
              users (name, email, age, active)
            VALUES
              ($1, $2, $3, $4);
        `)
		runFormattingTest(t, formatter, query, exp)
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
		runFormattingTest(t, formatter, query, exp)
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
		runFormattingTest(t, formatter, query, exp)
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
		runFormattingTest(t, formatter, query, exp)
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

func TestPostgreSQLFormatter_TypeCasts(t *testing.T) {
	t.Run("formats basic type casts", func(t *testing.T) {
		query := "SELECT id::integer, name::text FROM users;"
		exp := Dedent(`
            SELECT
              id::integer,
              name::text
            FROM
              users;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex type casts with functions", func(t *testing.T) {
		query := "SELECT json_data::jsonb->>'key'::text, UPPER(name)::varchar(50) FROM users;"
		exp := Dedent(`
            SELECT
              json_data::jsonb ->> 'key'::text,
              UPPER(name)::varchar(50)
            FROM
              users;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats array type casts", func(t *testing.T) {
		query := "SELECT ARRAY[1,2,3]::integer[], tags::text[] FROM posts;"
		exp := Dedent(`
            SELECT
              ARRAY [1,
              2,
              3 ]::integer[],
              tags::text[]
            FROM
              posts;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats chained operations with casts", func(t *testing.T) {
		query := "SELECT (price * 1.1)::numeric(10,2), LENGTH(description::text)::integer FROM products;"
		exp := Dedent(`
            SELECT
              (price * 1.1)::numeric(10, 2),
              LENGTH(description::text)::integer
            FROM
              products;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats date/time casts", func(t *testing.T) {
		query := `SELECT created_at::date, updated_at::timestamp, '2023-01-01'::date FROM events ` +
			`WHERE created_at::date = '2023-01-01'::date;`
		exp := Dedent(`
            SELECT
              created_at::date,
              updated_at::timestamp,
              '2023-01-01'::date
            FROM
              events
            WHERE
              created_at::date = '2023-01-01'::date;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats nested casts and complex expressions", func(t *testing.T) {
		query := "SELECT ((data->>'count')::integer * 2)::text, COALESCE(value::numeric, 0)::integer FROM analytics;"
		exp := Dedent(`
            SELECT
              ((data ->> 'count')::integer * 2)::text,
              COALESCE(value::numeric, 0)::integer
            FROM
              analytics;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats casts in WHERE clauses", func(t *testing.T) {
		query := `SELECT * FROM users WHERE id::text LIKE '%5%' ` +
			`AND created_at::date BETWEEN '2023-01-01'::date AND '2023-12-31'::date;`
		exp := Dedent(`
            SELECT
              *
            FROM
              users
            WHERE
              id::text LIKE '%5%'
              AND created_at::date BETWEEN '2023-01-01'::date
              AND '2023-12-31'::date;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats casts with custom types", func(t *testing.T) {
		query := "SELECT location::geography, status::user_status_enum, metadata::custom_type FROM users;"
		exp := Dedent(`
            SELECT
              location::geography,
              status::user_status_enum,
              metadata::custom_type
            FROM
              users;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("handles cast operator mixed with other PostgreSQL operators", func(t *testing.T) {
		query := `SELECT data::jsonb->>'key', data::jsonb->'array'->0::text, ` +
			`tags::text[] @> ARRAY['tag1']::text[] FROM posts;`
		exp := Dedent(`
            SELECT
              data::jsonb ->> 'key',
              data::jsonb -> 'array' -> 0::text,
              tags::text[] @> ARRAY ['tag1']::text[]
            FROM
              posts;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_JSONOperators(t *testing.T) {
	t.Run("formats JSON path extraction operators", func(t *testing.T) {
		query := "SELECT data->'name', data->>'email', info->'address'->>'city' FROM users;"
		exp := Dedent(`
            SELECT
              data -> 'name',
              data ->> 'email',
              info -> 'address' ->> 'city'
            FROM
              users;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats JSONB path operators", func(t *testing.T) {
		query := "SELECT data#>'{name,first}', data#>>'{contact,email}' FROM profiles;"
		exp := Dedent(`
            SELECT
              data #> '{name,first}',
              data #>> '{contact,email}'
            FROM
              profiles;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats JSONB containment operators", func(t *testing.T) {
		query := "SELECT * FROM products WHERE tags @> '[\"electronics\"]' AND price <@ '[100, 200]';"
		exp := Dedent(`
            SELECT
              *
            FROM
              products
            WHERE
              tags @> '["electronics"]'
              AND price <@ '[100, 200]';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats JSONB existence operators", func(t *testing.T) {
		query := "SELECT * FROM docs WHERE data ? 'key1' AND data ?| array['key2','key3'] AND data ?& array['req1','req2'];"
		exp := Dedent(`
            SELECT
              *
            FROM
              docs
            WHERE
              data ? 'key1'
              AND data ?| array ['key2',
              'key3' ]
              AND data ?& array ['req1',
              'req2' ];
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex nested JSON operations", func(t *testing.T) {
		query := "SELECT users.id, profile.data->'settings'->>'theme', " +

			"metadata#>'{tags,0}' FROM users " +

			"JOIN profiles profile ON users.id = profile.user_id WHERE profile.data @> " +

			"'{\"active\": true}' AND metadata ? 'priority';"
		exp := Dedent(`
            SELECT
              users.id,
              profile.data -> 'settings' ->> 'theme',
              metadata #> '{tags,0}'
            FROM
              users
              JOIN profiles profile ON users.id = profile.user_id
            WHERE
              profile.data @> '{"active": true}'
              AND metadata ? 'priority';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats JSON operators with type casts", func(t *testing.T) {
		query := "SELECT (data->>'count')::integer, (info#>>'{price,amount}')::numeric(10,2) FROM items;"
		exp := Dedent(`
            SELECT
              (data ->> 'count')::integer,
              (info #>> '{price,amount}')::numeric(10, 2)
            FROM
              items;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats JSON operations in WHERE clauses", func(t *testing.T) {
		query := "UPDATE users SET data = data || '{\"updated\": true}' " +

			"WHERE data->>'status' = 'active' AND settings @> '{\"notifications\": true}';"
		exp := Dedent(`
            UPDATE
              users
            SET
              data = data || '{"updated": true}'
            WHERE
              data ->> 'status' = 'active'
              AND settings @> '{"notifications": true}';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats JSON array operations", func(t *testing.T) {
		query := "SELECT array_data->0, array_data#>'{0,name}', tags[1], items @> '[{\"type\": \"book\"}]' FROM collections;"
		exp := Dedent(`
            SELECT
              array_data -> 0,
              array_data #> '{0,name}',
              tags[1],
              items @> '[{"type": "book"}]'
            FROM
              collections;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats mixed JSON and SQL operations", func(t *testing.T) {
		query := "SELECT CASE WHEN data ? 'premium' THEN data->>'premium_feature' ELSE 'standard' END, " +

			"LENGTH(info#>>'{description}') FROM accounts WHERE created_at > '2023-01-01' AND " +

			"(settings @> '{\"beta\": true}' " +

			"OR data->>'tier' = 'pro');"
		exp := Dedent(`
            SELECT
              CASE
                WHEN data ? 'premium' THEN data ->> 'premium_feature'
                ELSE 'standard'
              END,
              LENGTH(info #>> '{description}')
            FROM
              accounts
            WHERE
              created_at > '2023-01-01'
              AND (
                settings @> '{"beta": true}'
                OR data ->> 'tier' = 'pro'
              );
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_PatternMatching(t *testing.T) {
	t.Run("formats ILIKE queries", func(t *testing.T) {
		query := "SELECT name FROM users WHERE name ILIKE '%john%';"
		exp := Dedent(`
            SELECT
              name
            FROM
              users
            WHERE
              name ILIKE '%john%';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats ILIKE with complex patterns", func(t *testing.T) {
		query := "SELECT * FROM products WHERE name ILIKE '%laptop%' OR description ILIKE '%computer%';"
		exp := Dedent(`
            SELECT
              *
            FROM
              products
            WHERE
              name ILIKE '%laptop%'
              OR description ILIKE '%computer%';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats SIMILAR TO queries", func(t *testing.T) {
		query := "SELECT email FROM users WHERE email SIMILAR TO '%@gmail\\.com';"
		exp := Dedent(`
            SELECT
              email
            FROM
              users
            WHERE
              email SIMILAR TO '%@gmail\.com';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats SIMILAR TO with SQL regex patterns", func(t *testing.T) {
		query := "SELECT phone FROM contacts WHERE phone SIMILAR TO '[0-9]{3}-[0-9]{3}-[0-9]{4}';"
		exp := Dedent(`
            SELECT
              phone
            FROM
              contacts
            WHERE
              phone SIMILAR TO '[0-9]{3}-[0-9]{3}-[0-9]{4}';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats regex operators", func(t *testing.T) {
		query := "SELECT name FROM users WHERE name ~ '^[A-Z]' AND email !~ '@temp\\.';"
		exp := Dedent(`
            SELECT
              name
            FROM
              users
            WHERE
              name ~ '^[A-Z]'
              AND email !~ '@temp\.';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats case-insensitive regex operators", func(t *testing.T) {
		query := "SELECT title FROM articles WHERE title ~* 'postgresql|postgres' AND content !~* 'deprecated';"
		exp := Dedent(`
            SELECT
              title
            FROM
              articles
            WHERE
              title ~* 'postgresql|postgres'
              AND content !~* 'deprecated';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats negated pattern operators", func(t *testing.T) {
		query := "SELECT * FROM logs WHERE message !~ 'ERROR|FATAL' AND level NOT ILIKE 'debug';"
		exp := Dedent(`
            SELECT
              *
            FROM
              logs
            WHERE
              message !~ 'ERROR|FATAL'
              AND level NOT ILIKE 'debug';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats mixed pattern matching with other operators", func(t *testing.T) {
		query := "SELECT u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id " +
			"WHERE u.email ILIKE '%@company.com' AND p.content ~ 'urgent|important' " +
			"AND p.tags @> '[\"announcement\"]';"
		exp := Dedent(`
            SELECT
              u.name,
              p.title
            FROM
              users u
              JOIN posts p ON u.id = p.user_id
            WHERE
              u.email ILIKE '%@company.com'
              AND p.content ~ 'urgent|important'
              AND p.tags @> '["announcement"]';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats pattern matching in CASE statements", func(t *testing.T) {
		query := "SELECT CASE WHEN email ~ '^[a-z]+@[a-z]+\\.[a-z]+$' THEN 'valid' " +
			"WHEN email ILIKE '%@temp%' THEN 'temporary' ELSE 'invalid' END AS email_status FROM users;"
		exp := Dedent(`
            SELECT
              CASE
                WHEN email ~ '^[a-z]+@[a-z]+\.[a-z]+$' THEN 'valid'
                WHEN email ILIKE '%@temp%' THEN 'temporary'
                ELSE 'invalid'
              END AS email_status
            FROM
              users;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex regex patterns with escapes", func(t *testing.T) {
		query := "SELECT * FROM files WHERE filename ~* '\\.(jpg|png|gif)$' AND path !~ '/temp/|/cache/';"
		exp := Dedent(`
            SELECT
              *
            FROM
              files
            WHERE
              filename ~* '\.(jpg|png|gif)$'
              AND path !~ '/temp/|/cache/';
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats pattern matching with parameters", func(t *testing.T) {
		query := "SELECT name FROM products WHERE name ILIKE $1 AND description ~ $2;"
		exp := Dedent(`
            SELECT
              name
            FROM
              products
            WHERE
              name ILIKE $1
              AND description ~ $2;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_CTEs(t *testing.T) {
	TestCTEs(t, NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)), PostgreSQL)
}

func TestPostgreSQLFormatter_RETURNING(t *testing.T) {
	t.Run("formats INSERT with RETURNING", func(t *testing.T) {
		query := "INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com') RETURNING id, created_at;"
		exp := Dedent(`
            INSERT INTO
              users (name, email)
            VALUES
              ('John Doe', 'john@example.com')
            RETURNING
              id,
              created_at;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats UPDATE with RETURNING", func(t *testing.T) {
		query := "UPDATE users SET last_login = NOW() WHERE id = 1 RETURNING id, name, last_login;"
		exp := Dedent(`
            UPDATE
              users
            SET
              last_login = NOW()
            WHERE
              id = 1
            RETURNING
              id,
              name,
              last_login;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats DELETE with RETURNING", func(t *testing.T) {
		query := "DELETE FROM users WHERE active = false RETURNING id, name, deleted_at;"
		exp := Dedent(`
            DELETE FROM
              users
            WHERE
              active = false
            RETURNING
              id,
              name,
              deleted_at;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats RETURNING with expressions", func(t *testing.T) {
		query := "INSERT INTO orders (user_id, amount) VALUES (1, 100.50) " +
			"RETURNING id, amount * 1.1 AS amount_with_tax, CURRENT_TIMESTAMP as created;"
		exp := Dedent(`
            INSERT INTO
              orders (user_id, amount)
            VALUES
              (1, 100.50)
            RETURNING
              id,
              amount * 1.1 AS amount_with_tax,
              CURRENT_TIMESTAMP as created;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats RETURNING with complex UPDATE", func(t *testing.T) {
		query := "UPDATE products SET price = price * 1.1, updated_at = NOW() " +
			"WHERE category = 'electronics' RETURNING id, name, price, updated_at;"
		exp := Dedent(`
            UPDATE
              products
            SET
              price = price * 1.1,
              updated_at = NOW()
            WHERE
              category = 'electronics'
            RETURNING
              id,
              name,
              price,
              updated_at;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_UPSERT(t *testing.T) {
	t.Run("formats INSERT with ON CONFLICT DO NOTHING", func(t *testing.T) {
		query := "INSERT INTO users (id, name, email) VALUES (1, 'John', 'john@example.com') ON CONFLICT (id) DO NOTHING;"
		exp := Dedent(`
            INSERT INTO
              users (id, name, email)
            VALUES
              (1, 'John', 'john@example.com')
              ON CONFLICT (id) DO NOTHING;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats INSERT with ON CONFLICT DO UPDATE", func(t *testing.T) {
		query := "INSERT INTO users (id, name, email, login_count) " +
			"VALUES (1, 'John', 'john@example.com', 1) ON CONFLICT (id) DO UPDATE " +
			"SET login_count = users.login_count + 1, last_login = NOW();"
		exp := Dedent(`
            INSERT INTO
              users (id, name, email, login_count)
            VALUES
              (1, 'John', 'john@example.com', 1)
              ON CONFLICT (id) DO UPDATE
            SET
              login_count = users.login_count + 1,
              last_login = NOW();
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats UPSERT with RETURNING", func(t *testing.T) {
		query := "INSERT INTO products (sku, name, price) VALUES ('ABC123', 'Product A', 29.99) " +
			"ON CONFLICT (sku) DO UPDATE SET price = EXCLUDED.price, updated_at = NOW() " +
			"RETURNING id, sku, name, price;"
		exp := Dedent(`
            INSERT INTO
              products (sku, name, price)
            VALUES
              ('ABC123', 'Product A', 29.99)
              ON CONFLICT (sku) DO UPDATE
            SET
              price = EXCLUDED.price,
              updated_at = NOW()
            RETURNING
              id,
              sku,
              name,
              price;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats ON CONFLICT with WHERE condition", func(t *testing.T) {
		query := "INSERT INTO stats (date, user_id, views) VALUES ('2023-01-01', 1, 5) " +
			"ON CONFLICT (date, user_id) WHERE active = true DO UPDATE " +
			"SET views = stats.views + EXCLUDED.views;"
		exp := Dedent(`
            INSERT INTO
              stats (date, user_id, views)
            VALUES
              ('2023-01-01', 1, 5)
              ON CONFLICT (date, user_id)
            WHERE
              active = true DO UPDATE
            SET
              views = stats.views + EXCLUDED.views;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex UPSERT with multiple columns and expressions", func(t *testing.T) {
		query := Dedent(`
            INSERT INTO inventory (product_id, location_id, quantity, last_updated)
            VALUES (100, 1, 50, NOW())
            ON CONFLICT (product_id, location_id) DO UPDATE SET
                quantity = inventory.quantity + EXCLUDED.quantity,
                last_updated = GREATEST(inventory.last_updated, EXCLUDED.last_updated),
                modified_by = 'system'
            WHERE inventory.last_updated < EXCLUDED.last_updated
            RETURNING product_id, location_id, quantity, last_updated;
        `)
		exp := Dedent(`
            INSERT INTO
              inventory (product_id, location_id, quantity, last_updated)
            VALUES
              (100, 1, 50, NOW())
              ON CONFLICT (product_id, location_id) DO UPDATE
            SET
              quantity = inventory.quantity + EXCLUDED.quantity,
              last_updated = GREATEST(inventory.last_updated, EXCLUDED.last_updated),
              modified_by = 'system'
            WHERE
              inventory.last_updated < EXCLUDED.last_updated
            RETURNING
              product_id,
              location_id,
              quantity,
              last_updated;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_WindowFunctions(t *testing.T) {
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
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
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
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
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
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats aggregate window function with FILTER", func(t *testing.T) {
		query := "SELECT department, COUNT(*) FILTER (WHERE salary > 50000) OVER (PARTITION BY department) " +
			"as high_earners FROM employees;"
		exp := Dedent(`
            SELECT
              department,
              COUNT(*) FILTER (
                WHERE
                  salary > 50000
              ) OVER (PARTITION BY department) as high_earners
            FROM
              employees;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex window function with RANGE frame", func(t *testing.T) {
		query := "SELECT date, amount, SUM(amount) OVER (ORDER BY date RANGE BETWEEN INTERVAL '7 days' " +
			"PRECEDING AND CURRENT ROW) as rolling_week FROM transactions;"
		exp := Dedent(`
            SELECT
              date,
              amount,
              SUM(amount) OVER (
                ORDER BY
                  date RANGE BETWEEN INTERVAL '7 days' PRECEDING
                  AND CURRENT ROW
              ) as rolling_week
            FROM
              transactions;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
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
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
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
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_LateralJoins(t *testing.T) {
	t.Run("formats LATERAL JOIN", func(t *testing.T) {
		query := "SELECT u.name, t.amount FROM users u LEFT JOIN LATERAL " +
			"(SELECT amount FROM transactions WHERE user_id = u.id ORDER BY created_at DESC LIMIT 1) t ON true;"
		exp := Dedent(`
            SELECT
              u.name,
              t.amount
            FROM
              users u
              LEFT JOIN LATERAL (
                SELECT
                  amount
                FROM
                  transactions
                WHERE
                  user_id = u.id
                ORDER BY
                  created_at DESC
                LIMIT
                  1
              ) t ON true;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats CROSS JOIN LATERAL", func(t *testing.T) {
		query := "SELECT u.name, func.result FROM users u CROSS JOIN LATERAL some_function(u.id, u.name) AS func(result);"
		exp := Dedent(`
            SELECT
              u.name,
              func.result
            FROM
              users u
              CROSS JOIN LATERAL some_function(u.id, u.name) AS func(result);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_NullsOrdering(t *testing.T) {
	t.Run("formats ORDER BY with NULLS FIRST", func(t *testing.T) {
		query := "SELECT name, salary FROM employees ORDER BY salary DESC NULLS FIRST;"
		exp := Dedent(`
            SELECT
              name,
              salary
            FROM
              employees
            ORDER BY
              salary DESC NULLS FIRST;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats ORDER BY with NULLS LAST", func(t *testing.T) {
		query := "SELECT name, department, salary FROM employees ORDER BY department ASC NULLS LAST, salary DESC NULLS FIRST;"
		exp := Dedent(`
            SELECT
              name,
              department,
              salary
            FROM
              employees
            ORDER BY
              department ASC NULLS LAST,
              salary DESC NULLS FIRST;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats window function with NULLS ordering", func(t *testing.T) {
		query := "SELECT name, salary, RANK() OVER (ORDER BY salary DESC NULLS LAST) as salary_rank FROM employees;"
		exp := Dedent(`
            SELECT
              name,
              salary,
              RANK() OVER (
                ORDER BY
                  salary DESC NULLS LAST
              ) as salary_rank
            FROM
              employees;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_ArrayAndRangeSupport(t *testing.T) {
	t.Run("formats array subscript operations", func(t *testing.T) {
		query := "SELECT arr[1], arr[2:5], matrix[1][2], data[0] FROM table1;"
		exp := Dedent(`
            SELECT
              arr[1],
              arr[2:5 ],
              matrix[1][2],
              data[0]
            FROM
              table1;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats ARRAY constructor", func(t *testing.T) {
		query := "SELECT ARRAY[1,2,3], ARRAY['a','b','c'] FROM table1;"
		exp := Dedent(`
            SELECT
              ARRAY [1,
              2,
              3 ],
              ARRAY ['a',
              'b',
              'c' ]
            FROM
              table1;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats array concatenation with || operator", func(t *testing.T) {
		query := "SELECT array1 || array2, 'text' || 'concat', arr || ARRAY[4,5] FROM table1;"
		exp := Dedent(`
            SELECT
              array1 || array2,
              'text' || 'concat',
              arr || ARRAY [4,
              5 ]
            FROM
              table1;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats UNNEST function", func(t *testing.T) {
		query := "SELECT unnest(ARRAY[1,2,3]) as value, unnest('{a,b,c}'::text[]) as letter FROM generate_series(1,3);"
		exp := Dedent(`
            SELECT
              unnest(ARRAY [1, 2, 3 ]) as value,
              unnest('{a,b,c}'::text[]) as letter
            FROM
              generate_series(1, 3);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats multi-dimensional arrays", func(t *testing.T) {
		query := "SELECT ARRAY[[1,2],[3,4]], matrix[1][1], data[2][3] FROM table1;"
		exp := Dedent(`
            SELECT
              ARRAY [[1,
              2 ],
              [3,
              4 ]],
              matrix[1][1],
              data[2][3]
            FROM
              table1;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats array operations with type casts", func(t *testing.T) {
		query := "SELECT ARRAY[1,2,3]::integer[], '{1,2,3}'::int[], tags::text[] FROM posts;"
		exp := Dedent(`
            SELECT
              ARRAY [1,
              2,
              3 ]::integer[],
              '{1,2,3}'::int[],
              tags::text[]
            FROM
              posts;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats array aggregation functions", func(t *testing.T) {
		query := "SELECT array_agg(name), array_agg(DISTINCT category ORDER BY category) FROM products GROUP BY brand;"
		exp := Dedent(`
            SELECT
              array_agg(name),
              array_agg(
                DISTINCT category
                ORDER BY
                  category
              )
            FROM
              products
            GROUP BY
              brand;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex array expressions", func(t *testing.T) {
		query := "SELECT array_length(tags, 1), array_position(statuses, 'active'), " +
			"array_remove(items, null) || ARRAY['new'] FROM table1;"
		exp := Dedent(`
            SELECT
              array_length(tags, 1),
              array_position(statuses, 'active'),
              array_remove(items, null) || ARRAY ['new']
            FROM
              table1;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_DOBlocks(t *testing.T) {
	t.Run("formats simple DO block", func(t *testing.T) {
		query := "DO $$ BEGIN RAISE NOTICE 'Hello, World!'; END $$;"
		exp := "DO\n  $$ BEGIN RAISE NOTICE 'Hello, World!'; END $$;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats DO block with LANGUAGE specification", func(t *testing.T) {
		query := "DO $do$ BEGIN PERFORM pg_notify('test', 'message'); END $do$ LANGUAGE plpgsql;"
		exp := "DO\n  $do$ BEGIN PERFORM pg_notify('test', 'message'); END $do$ LANGUAGE plpgsql;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats DO block with DECLARE section", func(t *testing.T) {
		query := Dedent(`
            DO $$
            DECLARE
                user_count integer;
                user_name text;
            BEGIN
                SELECT COUNT(*), 'admin' INTO user_count, user_name FROM users;
                RAISE NOTICE 'Found % users, type: %', user_count, user_name;
            END
            $$ LANGUAGE plpgsql;
        `)
		exp := Dedent(`
            DO
              $$
            DECLARE
                user_count integer;
                user_name text;
            BEGIN
                SELECT COUNT(*), 'admin' INTO user_count, user_name FROM users;
                RAISE NOTICE 'Found % users, type: %', user_count, user_name;
            END
            $$ LANGUAGE plpgsql;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats DO block with complex logic", func(t *testing.T) {
		query := Dedent(`
            DO $block$
            DECLARE
                rec RECORD;
                counter INTEGER := 0;
            BEGIN
                FOR rec IN SELECT * FROM users WHERE active = true LOOP
                    UPDATE users SET last_login = NOW() WHERE id = rec.id;
                    counter := counter + 1;
                END LOOP;
                RAISE NOTICE 'Updated % users', counter;
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE WARNING 'Failed to update users: %', SQLERRM;
            END
            $block$ LANGUAGE plpgsql;
        `)
		exp := Dedent(`
            DO
              $block$
            DECLARE
                rec RECORD;
                counter INTEGER := 0;
            BEGIN
                FOR rec IN SELECT * FROM users WHERE active = true LOOP
                    UPDATE users SET last_login = NOW() WHERE id = rec.id;
                    counter := counter + 1;
                END LOOP;
                RAISE NOTICE 'Updated % users', counter;
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE WARNING 'Failed to update users: %', SQLERRM;
            END
            $block$ LANGUAGE plpgsql;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_Functions(t *testing.T) {
	t.Run("formats basic CREATE FUNCTION", func(t *testing.T) {
		query := "CREATE FUNCTION get_user_count() RETURNS INTEGER AS $$ SELECT COUNT(*) FROM users; $$ LANGUAGE SQL;"
		exp := "CREATE FUNCTION\n  get_user_count() RETURNS INTEGER AS $$ SELECT COUNT(*) FROM users; $$ LANGUAGE SQL;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats CREATE OR REPLACE FUNCTION", func(t *testing.T) {
		query := "CREATE OR REPLACE FUNCTION add_numbers(a INTEGER, b INTEGER) RETURNS INTEGER AS $$ " +
			"SELECT a + b; $$ LANGUAGE SQL IMMUTABLE;"
		exp := "CREATE OR REPLACE FUNCTION\n  add_numbers(a INTEGER, b INTEGER) RETURNS INTEGER AS $$ " +
			"SELECT a + b; $$ LANGUAGE SQL IMMUTABLE;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats function with multiple modifiers", func(t *testing.T) {
		query := `CREATE FUNCTION get_secure_data(user_id INTEGER) RETURNS TABLE(id INTEGER, name TEXT) AS $$
			SELECT id, name FROM users WHERE id = user_id AND active = true; $$ LANGUAGE SQL STABLE SECURITY DEFINER;`
		exp := "CREATE FUNCTION\n  get_secure_data(user_id INTEGER) RETURNS TABLE(id INTEGER, name TEXT) AS $$\n" +
			"\t\t\tSELECT id, name FROM users WHERE id = user_id AND active = true; $$ LANGUAGE SQL STABLE SECURITY DEFINER;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		require.Equal(t, exp, result)
	})

	t.Run("formats function with SETOF return type", func(t *testing.T) {
		query := `CREATE FUNCTION get_all_users() RETURNS SETOF users AS $$
			SELECT * FROM users ORDER BY name; $$ LANGUAGE SQL STABLE;`
		exp := "CREATE FUNCTION\n  get_all_users() RETURNS SETOF users AS $$\n" +
			"\t\t\tSELECT * FROM users ORDER BY name; $$ LANGUAGE SQL STABLE;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		require.Equal(t, exp, result)
	})

	t.Run("formats function with COST and ROWS", func(t *testing.T) {
		query := `CREATE FUNCTION expensive_calculation(n INTEGER) RETURNS INTEGER AS $$
			SELECT factorial(n);
			$$ LANGUAGE SQL IMMUTABLE COST 1000 ROWS 1;`
		exp := "CREATE FUNCTION\n  expensive_calculation(n INTEGER) RETURNS INTEGER AS $$\n" +
			"\t\t\tSELECT factorial(n);\n\t\t\t$$ LANGUAGE SQL IMMUTABLE COST 1000 ROWS 1;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		require.Equal(t, exp, result)
	})

	t.Run("formats VOLATILE function with complex logic", func(t *testing.T) {
		query := Dedent(`
            CREATE OR REPLACE FUNCTION update_user_stats(user_id INTEGER) 
            RETURNS VOID AS $func$
            DECLARE
                current_count INTEGER;
            BEGIN
                SELECT COUNT(*) INTO current_count FROM user_actions WHERE user_id = user_id;
                UPDATE users SET action_count = current_count, updated_at = NOW() WHERE id = user_id;
            END
            $func$ LANGUAGE plpgsql VOLATILE SECURITY DEFINER;
        `)
		exp := Dedent(`
            CREATE OR REPLACE FUNCTION
              update_user_stats(user_id INTEGER) RETURNS VOID AS $func$
            DECLARE
                current_count INTEGER;
            BEGIN
                SELECT COUNT(*) INTO current_count FROM user_actions WHERE user_id = user_id;
                UPDATE users SET action_count = current_count, updated_at = NOW() WHERE id = user_id;
            END
            $func$ LANGUAGE plpgsql VOLATILE SECURITY DEFINER;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats trigger function", func(t *testing.T) {
		query := Dedent(`
            CREATE OR REPLACE FUNCTION update_modified_time() 
            RETURNS TRIGGER AS $$
            BEGIN
                NEW.updated_at = NOW();
                RETURN NEW;
            END
            $$ LANGUAGE plpgsql VOLATILE;
        `)
		exp := Dedent(`
            CREATE OR REPLACE FUNCTION
              update_modified_time() RETURNS TRIGGER AS $$
            BEGIN
                NEW.updated_at = NOW();
                RETURN NEW;
            END
            $$ LANGUAGE plpgsql VOLATILE;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats function with all stability options", func(t *testing.T) {
		query := `CREATE FUNCTION test_immutable() RETURNS INTEGER AS $$
			SELECT 42;
			$$ LANGUAGE SQL IMMUTABLE STRICT LEAKPROOF PARALLEL SAFE
			COST 1;`
		exp := "CREATE FUNCTION\n  test_immutable() RETURNS INTEGER AS $$\n\t\t\tSELECT 42;\n\t\t\t$$ " +
			"LANGUAGE SQL IMMUTABLE STRICT LEAKPROOF PARALLEL SAFE COST 1;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		require.Equal(t, exp, result)
	})

	t.Run("formats function with default parameters", func(t *testing.T) {
		query := `CREATE FUNCTION greet_user(
			name TEXT DEFAULT 'Anonymous',
			greeting TEXT DEFAULT 'Hello'
		) RETURNS TEXT AS $$
			SELECT greeting || ', ' || name || '!';
			$$ LANGUAGE SQL IMMUTABLE;`
		exp := "CREATE FUNCTION\n  greet_user(\n    name TEXT DEFAULT 'Anonymous',\n    greeting TEXT DEFAULT 'Hello'\n  ) " +
			"RETURNS TEXT AS $$\n\t\t\tSELECT greeting || ', ' || name || '!';\n\t\t\t$$ LANGUAGE SQL IMMUTABLE;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_ComplexPLpgSQL(t *testing.T) {
	t.Run("formats function with exception handling", func(t *testing.T) {
		query := Dedent(`
            CREATE OR REPLACE FUNCTION safe_divide(numerator NUMERIC, denominator NUMERIC)
            RETURNS NUMERIC AS $$
            DECLARE
                result NUMERIC;
            BEGIN
                IF denominator = 0 THEN
                    RAISE EXCEPTION 'Division by zero is not allowed';
                END IF;
                result := numerator / denominator;
                RETURN result;
            EXCEPTION
                WHEN division_by_zero THEN
                    RAISE WARNING 'Caught division by zero: %', SQLERRM;
                    RETURN NULL;
                WHEN OTHERS THEN
                    RAISE LOG 'Unexpected error in safe_divide: %', SQLERRM;
                    RETURN NULL;
            END
            $$ LANGUAGE plpgsql STABLE;
        `)
		exp := Dedent(`
            CREATE OR REPLACE FUNCTION
              safe_divide(numerator NUMERIC, denominator NUMERIC) RETURNS NUMERIC AS $$
            DECLARE
                result NUMERIC;
            BEGIN
                IF denominator = 0 THEN
                    RAISE EXCEPTION 'Division by zero is not allowed';
                END IF;
                result := numerator / denominator;
                RETURN result;
            EXCEPTION
                WHEN division_by_zero THEN
                    RAISE WARNING 'Caught division by zero: %', SQLERRM;
                    RETURN NULL;
                WHEN OTHERS THEN
                    RAISE LOG 'Unexpected error in safe_divide: %', SQLERRM;
                    RETURN NULL;
            END
            $$ LANGUAGE plpgsql STABLE;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats function with control flow structures", func(t *testing.T) {
		query := Dedent(`
            CREATE OR REPLACE FUNCTION process_users(min_age INTEGER DEFAULT 18)
            RETURNS TEXT AS $$
            DECLARE
                user_rec RECORD;
                processed_count INTEGER := 0;
                message TEXT;
            BEGIN
                FOR user_rec IN 
                    SELECT id, name, age, status FROM users 
                    WHERE age >= min_age 
                    ORDER BY age DESC
                LOOP
                    IF user_rec.status = 'active' THEN
                        UPDATE users SET last_processed = NOW() WHERE id = user_rec.id;
                        processed_count := processed_count + 1;
                    ELSIF user_rec.status = 'pending' THEN
                        UPDATE users SET status = 'active', last_processed = NOW() WHERE id = user_rec.id;
                        processed_count := processed_count + 1;
                    ELSE
                        CONTINUE;
                    END IF;
                END LOOP;
                
                message := format('Processed %s users with minimum age %s', processed_count, min_age);
                RETURN message;
            END
            $$ LANGUAGE plpgsql VOLATILE;
        `)
		exp := Dedent(`
            CREATE OR REPLACE FUNCTION
              process_users(min_age INTEGER DEFAULT 18) RETURNS TEXT AS $$
            DECLARE
                user_rec RECORD;
                processed_count INTEGER := 0;
                message TEXT;
            BEGIN
                FOR user_rec IN 
                    SELECT id, name, age, status FROM users 
                    WHERE age >= min_age 
                    ORDER BY age DESC
                LOOP
                    IF user_rec.status = 'active' THEN
                        UPDATE users SET last_processed = NOW() WHERE id = user_rec.id;
                        processed_count := processed_count + 1;
                    ELSIF user_rec.status = 'pending' THEN
                        UPDATE users SET status = 'active', last_processed = NOW() WHERE id = user_rec.id;
                        processed_count := processed_count + 1;
                    ELSE
                        CONTINUE;
                    END IF;
                END LOOP;
                
                message := format('Processed %s users with minimum age %s', processed_count, min_age);
                RETURN message;
            END
            $$ LANGUAGE plpgsql VOLATILE;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats advanced DO block with dynamic SQL", func(t *testing.T) {
		query := Dedent(`
            DO $$
            DECLARE
                table_name TEXT := 'users';
                column_name TEXT := 'email';
                sql_query TEXT;
                result_count INTEGER;
            BEGIN
                sql_query := format('SELECT COUNT(*) FROM %I WHERE %I IS NOT NULL', table_name, column_name);
                EXECUTE sql_query INTO result_count;
                RAISE NOTICE 'Table % has % non-null values in column %', table_name, result_count, column_name;
                
                IF result_count > 1000 THEN
                    sql_query := format('CREATE INDEX CONCURRENTLY IF NOT EXISTS ' +
                        'idx_%s_%s ON %I', table_name, column_name, table_name, column_name);
                    EXECUTE sql_query;
                    RAISE NOTICE 'Created index for % rows', result_count;
                END IF;
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE WARNING 'Error in dynamic SQL execution: %', SQLERRM;
            END
            $$;
        `)
		exp := Dedent(`
            DO
              $$
            DECLARE
                table_name TEXT := 'users';
                column_name TEXT := 'email';
                sql_query TEXT;
                result_count INTEGER;
            BEGIN
                sql_query := format('SELECT COUNT(*) FROM %I WHERE %I IS NOT NULL', table_name, column_name);
                EXECUTE sql_query INTO result_count;
                RAISE NOTICE 'Table % has % non-null values in column %', table_name, result_count, column_name;
                
                IF result_count > 1000 THEN
                    sql_query := format('CREATE INDEX CONCURRENTLY IF NOT EXISTS ' +
                        'idx_%s_%s ON %I (%I)', table_name, column_name, table_name, column_name);
                    EXECUTE sql_query;
                    RAISE NOTICE 'Created index for % rows', result_count;
                END IF;
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE WARNING 'Error in dynamic SQL execution: %', SQLERRM;
            END
            $$;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_DDL_IndexOperations(t *testing.T) {
	t.Run("formats basic CREATE INDEX", func(t *testing.T) {
		query := "CREATE INDEX idx_users_email ON users (email);"
		exp := Dedent(`
            CREATE INDEX
              idx_users_email ON users (email);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats CREATE UNIQUE INDEX", func(t *testing.T) {
		query := "CREATE UNIQUE INDEX idx_users_username ON users (username);"
		exp := Dedent(`
            CREATE UNIQUE INDEX
              idx_users_username ON users (username);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats CREATE INDEX with CONCURRENTLY", func(t *testing.T) {
		query := "CREATE INDEX CONCURRENTLY idx_orders_created_at ON orders (created_at);"
		exp := Dedent(`
            CREATE INDEX
              CONCURRENTLY idx_orders_created_at ON orders (created_at);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats CREATE INDEX with IF NOT EXISTS", func(t *testing.T) {
		query := "CREATE INDEX IF NOT EXISTS idx_products_category ON products (category);"
		exp := Dedent(`
            CREATE INDEX
              IF NOT EXISTS idx_products_category ON products (category);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats CREATE INDEX with GIN method", func(t *testing.T) {
		query := "CREATE INDEX idx_docs_content ON documents USING GIN (content);"
		exp := Dedent(`
            CREATE INDEX
              idx_docs_content ON documents USING GIN (content);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats CREATE INDEX with GIST method", func(t *testing.T) {
		query := "CREATE INDEX idx_locations_point ON locations USING GIST (location);"
		exp := Dedent(`
            CREATE INDEX
              idx_locations_point ON locations USING GIST (location);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats CREATE INDEX with INCLUDE clause", func(t *testing.T) {
		query := "CREATE INDEX idx_users_active_include ON users (active) INCLUDE (name, email);"
		exp := Dedent(`
            CREATE INDEX
              idx_users_active_include ON users (active) INCLUDE (name, email);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats CREATE INDEX with WHERE clause (partial index)", func(t *testing.T) {
		query := "CREATE INDEX idx_users_active ON users (id) WHERE active = true;"
		exp := Dedent(`
            CREATE INDEX
              idx_users_active ON users (id)
            WHERE
              active = true;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats CREATE INDEX with multiple columns", func(t *testing.T) {
		query := "CREATE INDEX idx_orders_user_date ON orders (user_id, created_at DESC, status);"
		exp := Dedent(`
            CREATE INDEX
              idx_orders_user_date ON orders (user_id, created_at DESC, status);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex CREATE INDEX with all options", func(t *testing.T) {
		query := "CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_complex ON products " +
			"USING BTREE (category, price DESC) INCLUDE (name) WHERE active = true;"
		exp := Dedent(`
            CREATE UNIQUE INDEX
              CONCURRENTLY IF NOT EXISTS idx_complex ON products USING BTREE (category, price DESC) INCLUDE (name)
            WHERE
              active = true;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestPostgreSQLFormatter_DDL_AlterAndDrop(t *testing.T) {
	t.Run("formats DROP INDEX", func(t *testing.T) {
		query := "DROP INDEX idx_users_email;"
		exp := Dedent(`
            DROP INDEX
              idx_users_email;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats DROP INDEX with IF EXISTS", func(t *testing.T) {
		query := "DROP INDEX IF EXISTS idx_products_category;"
		exp := Dedent(`
            DROP INDEX
              IF EXISTS idx_products_category;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats DROP INDEX with CONCURRENTLY", func(t *testing.T) {
		query := "DROP INDEX CONCURRENTLY idx_orders_status;"
		exp := Dedent(`
            DROP INDEX
              CONCURRENTLY idx_orders_status;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats REINDEX INDEX", func(t *testing.T) {
		query := "REINDEX INDEX idx_users_email;"
		exp := Dedent(`
            REINDEX
              INDEX idx_users_email;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats REINDEX TABLE", func(t *testing.T) {
		query := "REINDEX TABLE users;"
		exp := Dedent(`
            REINDEX
              TABLE users;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats ALTER TABLE ADD COLUMN", func(t *testing.T) {
		query := "ALTER TABLE users ADD COLUMN phone VARCHAR(20);"
		exp := Dedent(`
            ALTER TABLE
              users
            ADD
              COLUMN phone VARCHAR(20);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats ALTER TABLE DROP COLUMN", func(t *testing.T) {
		query := "ALTER TABLE users DROP COLUMN phone;"
		exp := Dedent(`
            ALTER TABLE
              users DROP COLUMN phone;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats ALTER TABLE ADD CONSTRAINT", func(t *testing.T) {
		query := "ALTER TABLE orders ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id);"
		exp := Dedent(`
            ALTER TABLE
              orders
            ADD
              CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id);
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats ALTER TABLE RENAME COLUMN", func(t *testing.T) {
		query := "ALTER TABLE users RENAME COLUMN name TO full_name;"
		exp := Dedent(`
            ALTER TABLE
              users RENAME COLUMN name TO full_name;
        `)
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats VACUUM", func(t *testing.T) {
		query := "VACUUM ANALYZE users;"
		exp := "VACUUM ANALYZE users;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
		require.Equal(t, exp, result)
	})

	t.Run("formats CLUSTER", func(t *testing.T) {
		query := "CLUSTER users USING idx_users_email;"
		exp := "CLUSTER users USING idx_users_email;"
		result := NewPostgreSQLFormatter(NewDefaultConfig().WithLang(PostgreSQL)).Format(query)
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
