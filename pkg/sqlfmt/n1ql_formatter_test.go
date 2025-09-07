package sqlfmt

import (
	"testing"
)

func TestN1QLFormatter_FormatBasic(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats SELECT query with element selection expression",
			query: "SELECT order_lines[0].productId FROM orders;",
			exp: Dedent(`
              SELECT
                order_lines[0].productId
              FROM
                orders;
            `),
		},
		{
			name:  "formats SELECT query with primary key querying",
			query: "SELECT fname, email FROM tutorial USE KEYS ['dave', 'ian'];",
			exp: Dedent(`
              SELECT
                fname,
                email
              FROM
                tutorial
              USE KEYS
                ['dave', 'ian'];
            `),
		},
		{
			name:  "formats INSERT with {} object literal",
			query: "INSERT INTO heroes (KEY, VALUE) VALUES ('123', {'id':1,'type':'Tarzan'});",
			exp: Dedent(`
              INSERT INTO
                heroes (KEY, VALUE)
              VALUES
                ('123', {'id': 1, 'type': 'Tarzan'});
            `),
		},
	}

	runFormatterTests(t, tests, NewN1QLFormatter)
}

func TestN1QLFormatter_FormatComplex(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name: "formats INSERT with large object and array literals",
			query: `
              INSERT INTO heroes (KEY, VALUE) VALUES ('123', {'id': 1, 'type': 'Tarzan',
              'array': [123456789, 123456789, 123456789, 123456789, 123456789], 'hello': 'world'});
            `,
			exp: Dedent(`
              INSERT INTO
                heroes (KEY, VALUE)
              VALUES
                (
                  '123',
                  {
                    'id': 1,
                    'type': 'Tarzan',
                    'array': [
                      123456789,
                      123456789,
                      123456789,
                      123456789,
                      123456789
                    ],
                    'hello': 'world'
                  }
                );
            `),
		},
		{
			name:  "formats SELECT query with UNNEST top level reserved word",
			query: "SELECT * FROM tutorial UNNEST tutorial.children c;",
			exp: Dedent(`
              SELECT
                *
              FROM
                tutorial
              UNNEST
                tutorial.children c;
            `),
		},
		{
			name: "formats SELECT query with NEST and USE KEYS",
			query: `
              SELECT * FROM usr
              USE KEYS 'Elinor_33313792' NEST orders_with_users orders
              ON KEYS ARRAY s.order_id FOR s IN usr.shipped_order_history END;
            `,
			exp: Dedent(`
              SELECT
                *
              FROM
                usr
              USE KEYS
                'Elinor_33313792'
              NEST
                orders_with_users orders ON KEYS ARRAY s.order_id FOR s IN usr.shipped_order_history END;
            `),
		},
	}

	runFormatterTests(t, tests, NewN1QLFormatter)
}

func TestN1QLFormatter_FormatOperations(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats explained DELETE query with USE KEYS and RETURNING",
			query: "EXPLAIN DELETE FROM tutorial t USE KEYS 'baldwin' RETURNING t",
			exp: Dedent(`
              EXPLAIN DELETE FROM
                tutorial t
              USE KEYS
                'baldwin' RETURNING t
            `),
		},
		{
			name:  "formats UPDATE query with USE KEYS and RETURNING",
			query: "UPDATE tutorial USE KEYS 'baldwin' SET type = 'actor' RETURNING tutorial.type",
			exp: Dedent(`
              UPDATE
                tutorial
              USE KEYS
                'baldwin'
              SET
                type = 'actor' RETURNING tutorial.type
            `),
		},
	}

	runFormatterTests(t, tests, NewN1QLFormatter)
}

func TestN1QLFormatter_FormatVariables(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "recognizes $variables",
			query: "SELECT $variable, $'var name', $\"var name\", $`var name`;",
			exp: Dedent(`
              SELECT
                $variable,
                $'var name',
                $"var name",
                ` + "$`var name`;" + `
            `),
		},
		{
			name:  "replaces $variables with param values",
			query: "SELECT $variable, $'var name', $\"var name\", $`var name`;",
			exp: Dedent(`
              SELECT
                "variable value",
                'var value',
                'var value',
                'var value';
            `),
			cfg: Config{
				Params: NewMapParams(map[string]string{
					"variable": `"variable value"`,
					"var name": "'var value'",
				}),
			},
		},
		{
			name:  "replaces $ numbered placeholders with param values",
			query: "SELECT $1, $2, $0;",
			exp: Dedent(`
              SELECT
                second,
                third,
                first;
            `),
			cfg: Config{
				Params: NewListParams([]string{
					"first",
					"second",
					"third",
				}),
			},
		},
	}

	runFormatterTests(t, tests, NewN1QLFormatter)
}

func TestN1QLFormatter_Format(t *testing.T) {
	TestN1QLFormatter_FormatBasic(t)
	TestN1QLFormatter_FormatComplex(t)
	TestN1QLFormatter_FormatOperations(t)
	TestN1QLFormatter_FormatVariables(t)
}
