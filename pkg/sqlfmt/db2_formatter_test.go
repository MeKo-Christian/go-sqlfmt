package sqlfmt

import (
	"testing"
)

func TestDB2Formatter_FormatBasic(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats FETCH FIRST like LIMIT",
			query: "SELECT col1 FROM tbl ORDER BY col2 DESC FETCH FIRST 20 ROWS ONLY;",
			exp: Dedent(`
              SELECT
                col1
              FROM
                tbl
              ORDER BY
                col2 DESC
              FETCH FIRST
                20 ROWS ONLY;
            `),
		},
		{
			name: "formats only -- as a line comment",
			query: `
              SELECT col FROM
              -- This is a comment
              MyTable;
            `,
			exp: Dedent(`
              SELECT
                col
              FROM
                -- This is a comment
                MyTable;
            `),
		},
	}

	runFormatterTests(t, tests, NewDB2Formatter)
}

func TestDB2Formatter_FormatIdentifiers(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "recognizes @ and # as part of identifiers",
			query: "SELECT col#1, @col2 FROM tbl",
			exp: Dedent(`
              SELECT
                col#1,
                @col2
              FROM
                tbl
            `),
		},
		{
			name:  "recognizes :variables",
			query: "SELECT :variable;",
			exp: Dedent(`
              SELECT
                :variable;
            `),
		},
		{
			name:  "replaces :variables with param values",
			query: "SELECT :variable",
			exp: Dedent(`
              SELECT
                "variable value"
            `),
			cfg: Config{
				Params: NewMapParams(map[string]string{
					"variable": "\"variable value\"",
				}),
			},
		},
	}

	runFormatterTests(t, tests, NewDB2Formatter)
}

func TestDB2Formatter_Format(t *testing.T) {
	TestDB2Formatter_FormatBasic(t)
	TestDB2Formatter_FormatIdentifiers(t)
}
