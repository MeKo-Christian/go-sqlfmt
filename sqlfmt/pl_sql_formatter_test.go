package sqlfmt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPLSQLFormatter_Format(t *testing.T) {
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
			name:  "formats only -- as a line comment",
			query: "SELECT col FROM\n-- This is a comment\nMyTable;\n",
			exp: Dedent(`
              SELECT
                col
              FROM
                -- This is a comment
                MyTable;
            `),
		},
		{
			name:  "recognizes _, $, #, . and @ as part of identifiers",
			query: "SELECT my_col$1#, col.2@ FROM tbl\n",
			exp: Dedent(`
              SELECT
                my_col$1#,
                col.2@
              FROM
                tbl
            `),
		},
		{
			name:  "formats short CREATE TABLE",
			query: "CREATE TABLE items (a INT PRIMARY KEY, b TEXT);",
			exp:   "CREATE TABLE items (a INT PRIMARY KEY, b TEXT);",
		},
		{
			name:  "formats long CREATE TABLE",
			query: "CREATE TABLE items (a INT PRIMARY KEY, b TEXT, c INT NOT NULL, d INT NOT NULL);",
			exp: Dedent(`
              CREATE TABLE items (
                a INT PRIMARY KEY,
                b TEXT,
                c INT NOT NULL,
                d INT NOT NULL
              );
            `),
		},
		{
			name:  "formats INSERT without INTO",
			query: "INSERT Customers (ID, MoneyBalance, Address, City) VALUES (12,-123.4, 'Skagen 2111','Stv');",
			exp: Dedent(`
              INSERT
                Customers (ID, MoneyBalance, Address, City)
              VALUES
                (12, -123.4, 'Skagen 2111', 'Stv');
            `),
		},
		{
			name:  "formats ALTER TABLE ... MODIFY query",
			query: "ALTER TABLE supplier MODIFY supplier_name char(100) NOT NULL;",
			exp: Dedent(`
              ALTER TABLE
                supplier
              MODIFY
                supplier_name char(100) NOT NULL;
            `),
		},
		{
			name:  "formats ALTER TABLE ... ALTER COLUMN query",
			query: "ALTER TABLE supplier ALTER COLUMN supplier_name VARCHAR(100) NOT NULL;",
			exp: Dedent(`
              ALTER TABLE
                supplier
              ALTER COLUMN
                supplier_name VARCHAR(100) NOT NULL;
            `),
		},
		{
			name:  "recognizes ?[0-9]* placeholders",
			query: "SELECT ?1, ?25, ?;",
			exp: Dedent(`
              SELECT
                ?1,
                ?25,
                ?;
            `),
		},
		{
			name:  "replaces ? numbered placeholders with param values",
			query: "SELECT ?1, ?2, ?0;",
			exp: Dedent(`
              SELECT
                second,
                third,
                first;
            `),
			cfg: Config{
				Params: NewListParams([]string{"first", "second", "third"}),
			},
		},
		{
			name:  "replaces ? indexed placeholders with param values",
			query: "SELECT ?, ?, ?;",
			exp: Dedent(`
              SELECT
                first,
                second,
                third;
            `),
			cfg: Config{
				Params: NewListParams([]string{"first", "second", "third"}),
			},
		},
		{
			name:  "formats SELECT query with CROSS JOIN",
			query: "SELECT a, b FROM t CROSS JOIN t2 on t.id = t2.id_t",
			exp: Dedent(`
              SELECT
                a,
                b
              FROM
                t
                CROSS JOIN t2 on t.id = t2.id_t
            `),
		},
		{
			name:  "formats SELECT query with CROSS APPLY",
			query: "SELECT a, b FROM t CROSS APPLY fn(t.id)",
			exp: Dedent(`
              SELECT
                a,
                b
              FROM
                t
                CROSS APPLY fn(t.id)
            `),
		},
		{
			name:  "formats simple SELECT",
			query: "SELECT N, M FROM t",
			exp: Dedent(`
              SELECT
                N,
                M
              FROM
                t
            `),
		},
		{
			name:  "formats simple SELECT with national characters",
			query: "SELECT N'value'",
			exp: Dedent(`
              SELECT
                N'value'
            `),
		},
		{
			name:  "formats SELECT query with OUTER APPLY",
			query: "SELECT a, b FROM t OUTER APPLY fn(t.id)",
			exp: Dedent(`
              SELECT
                a,
                b
              FROM
                t
                OUTER APPLY fn(t.id)
            `),
		},
		{
			name:  "formats CASE ... WHEN with a blank expression",
			query: "CASE WHEN option = 'foo' THEN 1 WHEN option = 'bar' THEN 2 WHEN option = 'baz' THEN 3 ELSE 4 END;",
			exp: Dedent(`
              CASE
                WHEN option = 'foo' THEN 1
                WHEN option = 'bar' THEN 2
                WHEN option = 'baz' THEN 3
                ELSE 4
              END;
            `),
		},
		{
			name:  "formats CASE ... WHEN inside SELECT",
			query: "SELECT foo, bar, CASE baz WHEN 'one' THEN 1 WHEN 'two' THEN 2 ELSE 3 END FROM table",
			exp: Dedent(`
              SELECT
                foo,
                bar,
                CASE
                  baz
                  WHEN 'one' THEN 1
                  WHEN 'two' THEN 2
                  ELSE 3
                END
              FROM
                table
            `),
		},
		{
			name:  "formats CASE ... WHEN with an expression",
			query: "CASE toString(getNumber()) WHEN 'one' THEN 1 WHEN 'two' THEN 2 WHEN 'three' THEN 3 ELSE 4 END;",
			exp: Dedent(`
              CASE
                toString(getNumber())
                WHEN 'one' THEN 1
                WHEN 'two' THEN 2
                WHEN 'three' THEN 3
                ELSE 4
              END;
            `),
		},
		{
			name:  "properly converts to uppercase in case statements",
			query: "case toString(getNumber()) when 'one' then 1 when 'two' then 2 when 'three' then 3 else 4 end;",
			exp: Dedent(`
              CASE
                toString(getNumber())
                WHEN 'one' THEN 1
                WHEN 'two' THEN 2
                WHEN 'three' THEN 3
                ELSE 4
              END;
            `),
			cfg: Config{
				Uppercase: true,
			},
		},
		{
			name: "formats Oracle recursive sub queries",
			query: `
              WITH t1(id, parent_id) AS (
                -- Anchor member.
                SELECT id, parent_id FROM tab1 WHERE parent_id IS NULL
                MINUS
                -- Recursive member.
                SELECT t2.id, t2.parent_id FROM tab1 t2, t1 WHERE t2.parent_id = t1.id
              ) SEARCH BREADTH FIRST BY id SET order1,
              another AS (SELECT * FROM dual)
              SELECT id, parent_id FROM t1 ORDER BY order1;
            `,
			exp: Dedent(`
              WITH t1(id, parent_id) AS (
                -- Anchor member.
                SELECT
                  id,
                  parent_id
                FROM
                  tab1
                WHERE
                  parent_id IS NULL
                MINUS
                -- Recursive member.
                SELECT
                  t2.id,
                  t2.parent_id
                FROM
                  tab1 t2,
                  t1
                WHERE
                  t2.parent_id = t1.id
              ) SEARCH BREADTH FIRST BY id SET order1,
              another AS (
                SELECT
                  *
                FROM
                  dual
              )
              SELECT
                id,
                parent_id
              FROM
                t1
              ORDER BY
                order1;
            `),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if !tt.cfg.Empty() {
				if tt.cfg.Indent == "" {
					tt.cfg.Indent = DefaultIndent
				}
				result = NewPLSQLFormatter(&tt.cfg).Format(tt.query)
			} else {
				result = NewPLSQLFormatter(NewDefaultConfig()).Format(tt.query)
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
