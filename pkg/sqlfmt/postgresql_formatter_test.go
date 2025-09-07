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
              ARRAY[1,
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
              tags::text[] @> ARRAY['tag1']::text[]
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
              AND data ?| array['key2',
              'key3' ]
              AND data ?& array['req1',
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
