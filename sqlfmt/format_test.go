package sqlfmt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatBasic(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "uses given indent config for indention",
			query: "SELECT count(*),Column1 FROM Table1;",
			exp: Dedent(`
				SELECT
					count(*),
					Column1
				FROM
					Table1;
			`),
		},
		{
			name:  "formats simple SET SCHEMA queries",
			query: "SET SCHEMA schema1; SET CURRENT SCHEMA schema2;",
			exp: Dedent(`
				SET SCHEMA
					schema1;

				SET CURRENT SCHEMA
					schema2;
			`),
		},
		{
			name:  "formats simple SELECT query",
			query: "SELECT count(*),Column1 FROM Table1;",
			exp: Dedent(`
				SELECT
					count(*),
					Column1
				FROM
					Table1;
			`),
		},
		{
			name:  "formats complex SELECT",
			query: "SELECT DISTINCT name, ROUND(age/7) field1, 18 + 20 AS field2, 'some string' FROM foo;",
			exp: Dedent(`
				SELECT
					DISTINCT name,
					ROUND(age / 7) field1,
					18 + 20 AS field2,
					'some string'
				FROM
					foo;
			`),
		},
		{
			name:  "formats SELECT with complex WHERE",
			query: "SELECT * FROM foo WHERE Column1 = 'testing' AND ( (Column2 = Column3 OR Column4 >= NOW()) );",
			exp: Dedent(`
				SELECT
					*
				FROM
					foo
				WHERE
					Column1 = 'testing'
					AND (
						(
							Column2 = Column3
							OR Column4 >= NOW()
						)
					);
			`),
		},
		{
			name: "formats SELECT with top level reserved words",
			query: `SELECT * FROM foo WHERE name = 'John' GROUP BY some_column
HAVING column > 10 ORDER BY other_column LIMIT 5;`,
			exp: Dedent(`
				SELECT
					*
				FROM
					foo
				WHERE
					name = 'John'
				GROUP BY
					some_column
				HAVING
					column > 10
				ORDER BY
					other_column
				LIMIT
					5;
			`),
		},
		{
			name:  "formats LIMIT with two comma-separated values on single line",
			query: "LIMIT 5, 10;",
			exp: Dedent(`
				LIMIT
					5, 10;
			`),
		},
		{
			name:  "formats LIMIT of single value followed by another SELECT using commas",
			query: "LIMIT 5; SELECT foo, bar;",
			exp: Dedent(`
				LIMIT
					5;

				SELECT
					foo,
					bar;
			`),
		},
		{
			name:  "formats LIMIT of single value and OFFSET",
			query: "LIMIT 5 OFFSET 8;",
			exp: Dedent(`
				LIMIT
					5 OFFSET 8;
			`),
		},
		{
			name:  "recognizes LIMIT in lowercase",
			query: "limit 5, 10;",
			exp: Dedent(`
				limit
					5, 10;
			`),
		},
		{
			name:  "preserves case of keywords",
			query: "select distinct * frOM foo left join bar WHERe a > 1 and b = 3",
			exp: Dedent(`
				select
				  distinct *
				frOM
				  foo
				  left join bar
				WHERe
				  a > 1
				  and b = 3
			`),
		},
		{
			name:  "formats SELECT query with SELECT query inside it",
			query: "SELECT *, SUM(*) AS sum FROM (SELECT * FROM Posts LIMIT 30) WHERE a > b",
			exp: Dedent(`
                SELECT
                  *,
                  SUM(*) AS sum
                FROM
                  (
                    SELECT
                      *
                    FROM
                      Posts
                    LIMIT
                      30
                  )
                WHERE
                  a > b
            `),
		},
		{
			name: "formats SELECT query with INNER JOIN",
			query: `SELECT customer_id.from, COUNT(order_id) AS total FROM customers
                INNER JOIN orders ON customers.customer_id = orders.customer_id;`,
			exp: Dedent(`
                SELECT
                  customer_id.from,
                  COUNT(order_id) AS total
                FROM
                  customers
                  INNER JOIN orders ON customers.customer_id = orders.customer_id;
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
				result = Format(tt.query, &tt.cfg)
			} else {
				result = Format(tt.query)
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

func TestFormatComments(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name: "formats SELECT query with different comments",
			query: Dedent(`
                SELECT
                /*
                 * This is a block comment
                 */
                * FROM
                -- This is another comment
                MyTable # One final comment
                WHERE 1 = 2;
            `),
			exp: Dedent(`
                SELECT
                  /*
                   * This is a block comment
                   */
                  *
                FROM
                  -- This is another comment
                  MyTable # One final comment
                WHERE
                  1 = 2;
            `),
		},
		{
			name: "maintains block comment indentation",
			query: Dedent(`
                SELECT
                  /*
                   * This is a block comment
                   */
                  *
                FROM
                  MyTable
                WHERE
                  1 = 2;
            `),
			exp: Dedent(`
                SELECT
                  /*
                   * This is a block comment
                   */
                  *
                FROM
                  MyTable
                WHERE
                  1 = 2;
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
				result = Format(tt.query, &tt.cfg)
			} else {
				result = Format(tt.query)
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

func TestFormatInsertUpdateDelete(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name: "formats simple INSERT query",
			query: `INSERT INTO Customers (ID, MoneyBalance, Address, City)
                VALUES (12, -123.4, 'Skagen 2111', 'Stv');`,
			exp: Dedent(`
                INSERT INTO
                  Customers (ID, MoneyBalance, Address, City)
                VALUES
                  (12, -123.4, 'Skagen 2111', 'Stv');
            `),
		},
		{
			name:  "keeps short parenthesized list with nested parenthesis on single line",
			query: "SELECT (a + b * (c - NOW()));",
			exp: Dedent(`
                SELECT
                  (a + b * (c - NOW()));
            `),
		},
		{
			name: "breaks long parenthesized lists to multiple lines",
			query: Dedent(`
                INSERT INTO some_table (id_product, id_shop, id_currency, id_country, id_registration) (
                  SELECT IF(dq.id_discounter_shopping = 2, dq.value, dq.value / 100),
                  IF (dq.id_discounter_shopping = 2, 'amount', 'percentage') FROM foo);
            `),
			exp: Dedent(`
                INSERT INTO
                  some_table (
                    id_product,
                    id_shop,
                    id_currency,
                    id_country,
                    id_registration
                  ) (
                    SELECT
                      IF(
                        dq.id_discounter_shopping = 2,
                        dq.value,
                        dq.value / 100
                      ),
                      IF (
                        dq.id_discounter_shopping = 2,
                        'amount',
                        'percentage'
                      )
                    FROM
                      foo
                  );
            `),
		},
		{
			name:  "formats simple UPDATE query",
			query: `UPDATE Customers SET ContactName='Alfred Schmidt', City='Hamburg' WHERE CustomerName='Alfreds Futterkiste';`,
			exp: Dedent(`
                UPDATE
                  Customers
                SET
                  ContactName = 'Alfred Schmidt',
                  City = 'Hamburg'
                WHERE
                  CustomerName = 'Alfreds Futterkiste';
            `),
		},
		{
			name:  "formats simple DELETE query",
			query: `DELETE FROM Customers WHERE CustomerName='Alfred' AND Phone=5002132;`,
			exp: Dedent(`
                DELETE FROM
                  Customers
                WHERE
                  CustomerName = 'Alfred'
                  AND Phone = 5002132;
            `),
		},
		{
			name:  "formats simple DROP query",
			query: "DROP TABLE IF EXISTS admin_role;",
			exp:   "DROP TABLE IF EXISTS admin_role;",
		},
		{
			name:  "formats incomplete query",
			query: "SELECT count(",
			exp: Dedent(`
                SELECT
                  count(
            `),
		},
		{
			name: "formats query that ends with open comment",
			query: Dedent(`
                SELECT count(*)
                /*Comment
            `),
			exp: Dedent(`
                SELECT
                  count(*)
                  /*Comment
            `),
		},
		{
			name:  "formats UPDATE query with AS part",
			query: `UPDATE customers SET total_orders = order_summary.total FROM (SELECT * FROM bank) AS order_summary`,
			exp: Dedent(`
                UPDATE
                  customers
                SET
                  total_orders = order_summary.total
                FROM
                  (
                    SELECT
                      *
                    FROM
                      bank
                  ) AS order_summary
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
				result = Format(tt.query, &tt.cfg)
			} else {
				result = Format(tt.query)
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

func TestFormatOperatorsJoins(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats top-level and newline multi-word reserved words with inconsistent spacing",
			query: "SELECT * FROM foo LEFT \t OUTER  \n JOIN bar ORDER \n BY blah",
			exp: Dedent(`
                SELECT
                  *
                FROM
                  foo
                  LEFT OUTER JOIN bar
                ORDER BY
                  blah
            `),
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatOperatorsParentheses(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats long double parenthesized queries to multiple lines",
			query: "((foo = '0123456789-0123456789-0123456789-0123456789'))",
			exp: Dedent(`
                (
                  (
                    foo = '0123456789-0123456789-0123456789-0123456789'
                  )
                )
            `),
		},
		{
			name:  "formats short double parenthesized queries to one line",
			query: "((foo = 'bar'))",
			exp:   "((foo = 'bar'))",
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatOperatorsBasic(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats single-char operators",
			query: "SELECT * FROM foo WHERE bar = 'a' AND baz = 'b';",
			exp: Dedent(`
                SELECT
                  *
                FROM
                  foo
                WHERE
                  bar = 'a'
                  AND baz = 'b';
            `),
		},
		{
			name:  "formats single-char operators with tabs",
			query: "SELECT * FROM foo WHERE bar\t = 'a' AND baz = 'b';",
			exp: Dedent(`
                SELECT
                  *
                FROM
                  foo
                WHERE
                  bar = 'a'
                  AND baz = 'b';
            `),
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatOperatorsFunctions(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name: "formats simple CASE query",
			query: Dedent(`
                SELECT CASE
                  WHEN a = 1 THEN 1
                  WHEN a = 2 THEN 2
                  ELSE 3
                END AS result
                FROM foo;
            `),
			exp: Dedent(`
                SELECT
                  CASE
                    WHEN a = 1 THEN 1
                    WHEN a = 2 THEN 2
                    ELSE 3
                  END AS result
                FROM
                  foo;
            `),
		},
		{
			name: "formats simple IF query",
			query: Dedent(`
                SELECT IF(a > 1, 'greater', 'lesser') AS result
                FROM foo;
            `),
			exp: Dedent(`
                SELECT
                  IF(a > 1, 'greater', 'lesser') AS result
                FROM
                  foo;
            `),
		},
		{
			name:  "formats simple EXISTS query",
			query: "SELECT EXISTS(SELECT 1 FROM foo);",
			exp: Dedent(`
                SELECT
                  EXISTS(
                    SELECT
                      1
                    FROM
                      foo
                  );
            `),
		},
		{
			name:  "formats simple COALESCE query",
			query: "SELECT COALESCE(a, b, c) AS result FROM foo;",
			exp: Dedent(`
                SELECT
                  COALESCE(a, b, c) AS result
                FROM
                  foo;
            `),
		},
		{
			name:  "formats simple NULLIF query",
			query: "SELECT NULLIF(a, b) AS result FROM foo;",
			exp: Dedent(`
                SELECT
                  NULLIF(a, b) AS result
                FROM
                  foo;
            `),
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatOperatorsClauses(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats simple GROUP BY query",
			query: "SELECT COUNT(*) FROM foo GROUP BY bar;",
			exp: Dedent(`
				SELECT
				  COUNT(*)
				FROM
				  foo
				GROUP BY
				  bar;
			`),
		},
		{
			name:  "formats simple ORDER BY query",
			query: "SELECT * FROM foo ORDER BY bar DESC;",
			exp: Dedent(`
				SELECT
				  *
				FROM
				  foo
				ORDER BY
				  bar DESC;
			`),
		},
		{
			name:  "formats simple HAVING query",
			query: "SELECT bar, COUNT(*) FROM foo GROUP BY bar HAVING COUNT(*) > 1;",
			exp: Dedent(`
				SELECT
				  bar,
				  COUNT(*)
				FROM
				  foo
				GROUP BY
				  bar
				HAVING
				  COUNT(*) > 1;
			`),
		},
		{
			name:  "formats simple LIMIT query",
			query: "SELECT * FROM foo LIMIT 10;",
			exp: Dedent(`
				SELECT
				  *
				FROM
				  foo
				LIMIT
				  10;
			`),
		},
		{
			name:  "formats simple OFFSET query",
			query: "SELECT * FROM foo LIMIT 10 OFFSET 5;",
			exp: Dedent(`
				SELECT
				  *
				FROM
				  foo
				LIMIT
				  10 OFFSET 5;
			`),
		},
		{
			name:  "formats simple UNION query",
			query: "SELECT * FROM foo UNION SELECT * FROM bar;",
			exp: Dedent(`
				SELECT
				  *
				FROM
				  foo
				UNION
				SELECT
				  *
				FROM
				  bar;
			`),
		},
		{
			name:  "formats simple INTERSECT query",
			query: "SELECT * FROM foo INTERSECT SELECT * FROM bar;",
			exp: Dedent(`
				SELECT
				  *
				FROM
				  foo
				INTERSECT
				SELECT
				  *
				FROM
				  bar;
			`),
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatOperatorsSymbols(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats short double parenthesized queries to one line",
			query: "((foo = 'bar'))",
			exp:   "((foo = 'bar'))",
		},
		{
			name:  "formats single-char operators",
			query: "foo = bar",
			exp:   "foo = bar",
		},
		{
			name:  "formats single-char operators (<)",
			query: "foo < bar",
			exp:   "foo < bar",
		},
		{
			name:  "formats single-char operators (>)",
			query: "foo > bar",
			exp:   "foo > bar",
		},
		{
			name:  "formats single-char operators (+)",
			query: "foo + bar",
			exp:   "foo + bar",
		},
		{
			name:  "formats single-char operators (-)",
			query: "foo - bar",
			exp:   "foo - bar",
		},
		{
			name:  "formats single-char operators (*)",
			query: "foo * bar",
			exp:   "foo * bar",
		},
		{
			name:  "formats single-char operators (/) ",
			query: "foo / bar",
			exp:   "foo / bar",
		},
		{
			name:  "formats single-char operators (%) ",
			query: "foo % bar",
			exp:   "foo % bar",
		},
		{
			name:  "formats multi-char operators (!=)",
			query: "foo != bar",
			exp:   "foo != bar",
		},
		{
			name:  "formats multi-char operators (<>)",
			query: "foo <> bar",
			exp:   "foo <> bar",
		},
		{
			name:  "formats multi-char operators (==)", // N1QL
			query: "foo == bar",
			exp:   "foo == bar",
		},
		{
			name:  "formats multi-char operators (||)", // Oracle, Postgre, N1QL string concat
			query: "foo || bar",
			exp:   "foo || bar",
		},
		{
			name:  "formats multi-char operators (<=)",
			query: "foo <= bar",
			exp:   "foo <= bar",
		},
		{
			name:  "formats multi-char operators (>=)",
			query: "foo >= bar",
			exp:   "foo >= bar",
		},
		{
			name:  "formats multi-char operators (!<)",
			query: "foo !< bar",
			exp:   "foo !< bar",
		},
		{
			name:  "formats multi-char operators (!>)",
			query: "foo !> bar",
			exp:   "foo !> bar",
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatOperatorsLogical(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats logical operators (ALL)",
			query: "foo ALL bar",
			exp:   "foo ALL bar",
		},
		{
			name:  "formats logical operators (= ANY)",
			query: "foo = ANY (1, 2, 3)",
			exp:   "foo = ANY (1, 2, 3)",
		},
		{
			name:  "formats logical operators (EXISTS)",
			query: "EXISTS bar",
			exp:   "EXISTS bar",
		},
		{
			name:  "formats logical operators (IN)",
			query: "foo IN (1, 2, 3)",
			exp:   "foo IN (1, 2, 3)",
		},
		{
			name:  "formats logical operators (LIKE)",
			query: "foo LIKE 'hello%'",
			exp:   "foo LIKE 'hello%'",
		},
		{
			name:  "formats logical operators (IS NULL)",
			query: "foo IS NULL",
			exp:   "foo IS NULL",
		},
		{
			name:  "formats logical operators (UNIQUE)",
			query: "UNIQUE foo",
			exp:   "UNIQUE foo",
		},
		{
			name:  "formats AND/OR operators (BETWEEN)",
			query: "foo BETWEEN bar AND baz",
			exp:   "foo BETWEEN bar\nAND baz",
		},
		{
			name:  "formats AND/OR operators (AND)",
			query: "foo AND bar",
			exp:   "foo\nAND bar",
		},
		{
			name:  "formats AND/OR operators (OR)",
			query: "foo OR bar",
			exp:   "foo\nOR bar",
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatOperators(t *testing.T) {
	TestFormatOperatorsJoins(t)
	TestFormatOperatorsParentheses(t)
	TestFormatOperatorsBasic(t)
	TestFormatOperatorsFunctions(t)
	TestFormatOperatorsClauses(t)
	TestFormatOperatorsSymbols(t)
	TestFormatOperatorsLogical(t)
}

func TestFormatStrings(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "recognizes strings (double quotes)",
			query: "\"foo JOIN bar\"",
			exp:   "\"foo JOIN bar\"",
		},
		{
			name:  "recognizes strings (single quotes)",
			query: "'foo JOIN bar'",
			exp:   "'foo JOIN bar'",
		},
		{
			name:  "recognizes strings (backticks)",
			query: "`foo JOIN bar`",
			exp:   "`foo JOIN bar`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if !tt.cfg.Empty() {
				if tt.cfg.Indent == "" {
					tt.cfg.Indent = DefaultIndent
				}
				result = Format(tt.query, &tt.cfg)
			} else {
				result = Format(tt.query)
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

func TestFormatSpecialStatements(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "keeps separation between multiple statements (semicolon)",
			query: "foo;bar;",
			exp:   "foo;\n\nbar;",
		},
		{
			name:  "keeps separation between multiple statements (newline)",
			query: "foo\n;bar;",
			exp:   "foo;\n\nbar;",
		},
		{
			name:  "keeps separation between multiple statements (multiple newlines)",
			query: "foo\n\n\n;bar;\n\n",
			exp:   "foo;\n\nbar;",
		},
		{
			name: "keeps separation between multiple statements (SELECT)",
			query: `
				SELECT count(*),Column1 FROM Table1;
				SELECT count(*),Column1 FROM Table2;
			`,
			exp: Dedent(`
				SELECT
					count(*),
					Column1
				FROM
					Table1;

				SELECT
					count(*),
					Column1
				FROM
					Table2;
			`),
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatSpecialUnicode(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats unicode correctly",
			query: "SELECT test, тест FROM table;",
			exp: Dedent(`
				SELECT
					test,
					тест
				FROM
					table;
			`),
		},
		{
			name:  "converts keywords to uppercase when option passed in",
			query: "select distinct * frOM foo left join bar WHERe cola > 1 and colb = 3",
			exp: Dedent(`
				SELECT
					DISTINCT *
				FROM
					foo
					LEFT JOIN bar
				WHERE
					cola > 1
					AND colb = 3
			`),
			cfg: Config{Uppercase: true},
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatSpecialConfig(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "line breaks between queries change with config",
			query: "SELECT * FROM foo; SELECT * FROM bar;",
			exp: Dedent(`
				SELECT
					*
				FROM
					foo;

				SELECT
					*
				FROM
					bar;
			`),
			cfg: Config{LinesBetweenQueries: 2},
		},
		{
			name: "correctly indents create statement after select",
			query: `
				SELECT * FROM test;
				CREATE TABLE TEST(id NUMBER NOT NULL, col1 VARCHAR2(20), col2 VARCHAR2(20));
			`,
			exp: Dedent(`
				SELECT
					*
				FROM
					test;

				CREATE TABLE TEST(
					id NUMBER NOT NULL,
					col1 VARCHAR2(20),
					col2 VARCHAR2(20)
				);
			`),
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatSpecialAdvanced(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name: "formats $$ correctly",
			query: Dedent(`
				CREATE
				OR REPLACE FUNCTION RECURSION_TEST (STR VARCHAR) RETURNS VARCHAR LANGUAGE JAVASCRIPT AS $$
				return (STR.length <= 1
                    ? STR : STR.substring(0,1) + '_' + RECURSION_TEST(STR.substring(1)));
				$$;
			`),
			exp: Dedent(`
				CREATE
				OR REPLACE FUNCTION RECURSION_TEST (STR VARCHAR) RETURNS VARCHAR LANGUAGE JAVASCRIPT AS $$
				return (STR.length <= 1
                    ? STR : STR.substring(0,1) + '_' + RECURSION_TEST(STR.substring(1)));
				$$;
			`),
		},
		{
			name:  "formats => correctly",
			query: `select seq4(), uniform(1, 10, random(12)) from table(generator(rowcount => 11000)) v`,
			exp: Dedent(`
				select
					seq4(),
					uniform(1, 10, random(12))
				from
					table(generator(rowcount => 11000)) v
			`),
		},
		{
			name:  "formats UNION ALL on one line",
			query: `SELECT * FROM expr_0 UNION ALL SELECT * FROM expr_1`,
			exp: Dedent(`
                SELECT
                  *
                FROM
                  expr_0
                UNION ALL
                SELECT
                  *
                FROM
                  expr_1
			`),
		},
		{
			name: "formats complex WITH statements correctly",
			query: Dedent(`
				WITH expr_0 AS (
					SELECT 'foo' AS db, 'foo' AS another
					FROM foo
				),
				expr_1 AS (
					WITH expr_2 AS (
						SELECT 'bar' AS db, 'bar' AS another
						FROM bar
					),
					expr_3 AS (
						WITH expr_4 AS (
							SELECT 'goob' AS db, 'goob' AS another,
							FROM goob
						)
						SELECT expr_4.db AS db, expr_4.another AS another, expr_4.p_timeline AS p_timeline,
							OBJECT_INSERT( expr_4.row_object, 'vroom', row_object:r / 2, TRUE) AS row_object
						FROM expr_4
					),
					expr_6 AS (
						SELECT *
						FROM expr_2
						LIMIT
							10
						UNION ALL
						SELECT *
						FROM expr_3
					)
					SELECT *
					FROM expr_6
					WHERE EQUAL_NULL(row_object:p, row_object:h)
				)
				SELECT *
				FROM expr_0
				UNION ALL
				SELECT *
				FROM expr_1
			`),
			exp: Dedent(`
				WITH expr_0 AS (
					SELECT
						'foo' AS db,
						'foo' AS another
					FROM
						foo
				),
				expr_1 AS (
					WITH expr_2 AS (
						SELECT
							'bar' AS db,
							'bar' AS another
						FROM
							bar
					),
					expr_3 AS (
						WITH expr_4 AS (
							SELECT
								'goob' AS db,
								'goob' AS another,
							FROM
								goob
						)
						SELECT
							expr_4.db AS db,
							expr_4.another AS another,
							expr_4.p_timeline AS p_timeline,
							OBJECT_INSERT(
								expr_4.row_object,
								'vroom',
								row_object:r / 2,
								TRUE
							) AS row_object
						FROM
							expr_4
					),
					expr_6 AS (
						SELECT
							*
						FROM
							expr_2
						LIMIT
							10
						UNION ALL
						SELECT
							*
						FROM
							expr_3
					)
					SELECT
						*
					FROM
						expr_6
					WHERE
						EQUAL_NULL(row_object:p, row_object:h)
				)
				SELECT
					*
				FROM
					expr_0
				UNION ALL
				SELECT
					*
				FROM
					expr_1
			`),
		},
	}

	runFormatterTests(t, tests, func(cfg *Config) Formatter {
		return NewStandardSQLFormatter(cfg)
	})
}

func TestFormatSpecial(t *testing.T) {
	TestFormatSpecialStatements(t)
	TestFormatSpecialUnicode(t)
	TestFormatSpecialConfig(t)
	TestFormatSpecialAdvanced(t)
}

func TestFormat(t *testing.T) {
	// This function now calls the individual test functions
	TestFormatBasic(t)
	TestFormatComments(t)
	TestFormatInsertUpdateDelete(t)
	TestFormatOperators(t)
	TestFormatStrings(t)
	TestFormatSpecial(t)
}

func TestPrettyPrint(t *testing.T) {
	q := Dedent(`
        WITH expr_0 AS (
            SELECT
                'foo' AS db, -- inline comment
                'foo' AS another,
                COUNT_ME(*) AS count,
                true AS bool,
                TRUE as tuk,
                false AS bool2,
                8 AS number,
                6.8 AS number2,
                "hi"::int,
                obj:subfield,
            FROM
                /*
                 * block comment
                 */
                foo
        ),
        SELECT
            *,
            3 + (4-5) AS that,
        FROM
            expr_0
    `)

	PrettyPrint(q)
}

func TestPrettyFormat(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
	}{
		{
			name:  "colors reserved words",
			query: `SELECT * FROM foo`,
			exp: Dedent(fmt.Sprintf(`
			    %s%sSELECT%s%s
			      *
			    %s%sFROM%s%s
			      foo
			`, FormatBold, ColorCyan, FormatReset, FormatReset,
				FormatBold, ColorCyan, FormatReset, FormatReset)),
		},
		{
			name:  "colors strings",
			query: `SELECT 'foo'`,
			exp: Dedent(fmt.Sprintf(`
			    %s%sSELECT%s%s
			      %s'foo'%s
			`, FormatBold, ColorCyan, FormatReset, FormatReset,
				ColorGreen, FormatReset)),
		},
		{
			name:  "colors numbers",
			query: `SELECT 9.7`,
			exp: Dedent(fmt.Sprintf(`
			    %s%sSELECT%s%s
			      %s9.7%s
			`, FormatBold, ColorCyan, FormatReset, FormatReset,
				ColorBrightBlue, FormatReset)),
		},
		{
			name:  "colors booleans",
			query: `SELECT true`,
			exp: Dedent(fmt.Sprintf(`
			    %s%sSELECT%s%s
			      %s%strue%s%s
			`, FormatBold, ColorCyan, FormatReset, FormatReset,
				FormatBold, ColorPurple, FormatReset, FormatReset)),
		},
		{
			name:  "colors inline comments",
			query: `SELECT foo -- this is a comment`,
			exp: Dedent(fmt.Sprintf(`
			    %s%sSELECT%s%s
			      foo %s-- this is a comment%s
			`, FormatBold, ColorCyan, FormatReset, FormatReset,
				ColorGray, FormatReset)),
		},
		{
			name: "colors block comments",
			query: `
				SELECT 
				  /*
				   * block comment
				   */
				  foo`,
			exp: Dedent(fmt.Sprintf(`
                %s%sSELECT%s%s
                  %s/*%s
                %s   * block comment%s
                %s   */%s
                  foo
            `, FormatBold, ColorCyan, FormatReset, FormatReset,
				ColorGray, FormatReset,
				ColorGray, FormatReset,
				ColorGray, FormatReset)),
		},
		{
			name:  "colors functions",
			query: `SELECT COUNT_ME(*)`,
			exp: Dedent(fmt.Sprintf(`
			    %s%sSELECT%s%s
			      %sCOUNT_ME%s(*)
			`, FormatBold, ColorCyan, FormatReset, FormatReset,
				ColorBrightCyan, FormatReset)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := strings.TrimRight(tt.exp, "\n\t ")
			exp = strings.TrimLeft(exp, "\n")
			exp = strings.ReplaceAll(exp, "\t", DefaultIndent)

			p := PrettyFormat(tt.query)
			if p != exp {
				fmt.Println("=== QUERY ===")
				fmt.Println(tt.query)
				fmt.Println()

				fmt.Println("=== EXP ===")
				fmt.Println(exp)
				fmt.Println()

				fmt.Println("=== RESULT ===")
				fmt.Println(p)
				fmt.Println()
			}
			require.Equal(t, exp, p)
		})
	}
}

func BenchmarkFormat(b *testing.B) {
	for range b.N {
		Format(`SELECT foo AS a, boo AS b FROM table WHERE foo = bar LIMIT 10`)
	}
}

func BenchmarkPrettyFormat(b *testing.B) {
	for range b.N {
		PrettyFormat(`SELECT foo AS a, boo AS b FROM table WHERE foo = bar LIMIT 10`)
	}
}

// runFormatterTests runs the common test logic for all formatter test files.
func runFormatterTests(t *testing.T, tests []struct {
	name  string
	query string
	exp   string
	cfg   Config
}, formatterFactory func(*Config) Formatter) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if !tt.cfg.Empty() {
				if tt.cfg.Indent == "" {
					tt.cfg.Indent = DefaultIndent
				}
				result = formatterFactory(&tt.cfg).Format(tt.query)
			} else {
				result = formatterFactory(NewDefaultConfig()).Format(tt.query)
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
