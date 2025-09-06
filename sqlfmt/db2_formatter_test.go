package sqlfmt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDB2Formatter_Format(t *testing.T) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if !tt.cfg.Empty() {
				if tt.cfg.Indent == "" {
					tt.cfg.Indent = DefaultIndent
				}
				result = NewDB2Formatter(&tt.cfg).Format(tt.query)
			} else {
				result = NewDB2Formatter(NewDefaultConfig()).Format(tt.query)
			}

			exp := strings.TrimRight(tt.exp, "\n\t ")
			exp = strings.TrimLeft(exp, "\n")
			exp = strings.ReplaceAll(exp, "\t", DefaultIndent)

			if result != exp {
				fmt.Println("=== QUERY ===")
				fmt.Println(tt.query)
				fmt.Println()

				fmt.Println("=== EXP ===")
				fmt.Println(exp)
				fmt.Println()

				fmt.Println("=== RESULT ===")
				fmt.Println(result)
				fmt.Println()
			}
			require.Equal(t, exp, result)
		})
	}
}
