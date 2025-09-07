package sqlfmt

import (
	"testing"
)

func TestStandardSQLFormatter_FormatDDL(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
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
	}

	runFormatterTests(t, tests, NewStandardSQLFormatter)
}

func TestStandardSQLFormatter_FormatVariables(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "recognizes [] strings",
			query: "[foo JOIN bar]",
			exp:   "[foo JOIN bar]",
		},
		{
			name:  "recognizes @variables",
			query: "SELECT @variable, @a1_2.3$, @'var name', @\"var name\", @`var name`, @[var name];",
			exp: Dedent(`
                SELECT
                  @variable,
                  @a1_2.3$,
                  @'var name',
                  @"var name",
                  @` + "`var name`," + `
                  @[var name];
            `),
		},
		{
			name:  "replaces @variables with param values",
			query: "SELECT @variable, @a1_2.3$, @'var name', @\"var name\", @`var name`, @[var name], @'var\\name';",
			exp: Dedent(`
                SELECT
                  "variable value",
                  'weird value',
                  'var value',
                  'var value',
                  'var value',
                  'var value',
                  'var\ value';
            `),
			cfg: Config{
				Params: NewMapParams(map[string]string{
					"variable":  "\"variable value\"",
					"a1_2.3$":   "'weird value'",
					"var name":  "'var value'",
					"var\\name": "'var\\ value'",
				}),
			},
		},
		{
			name:  "recognizes :variables",
			query: "SELECT :variable, :a1_2.3$, :'var name', :\"var name\", :`var name`, :[var name];",
			exp: Dedent(`
				SELECT
					:variable,
					:a1_2.3$,
					:'var name',
					:"var name",
					:` + "`var name`," + `
					:[var name];
			`),
		},
		{
			name: "replaces :variables with param values",
			query: "SELECT :variable, :a1_2.3$, :'var name', :\"var name\", :`var name`, " +
				":[var name], :'escaped \\'var\\'', :\"^*& weird \\\" var   \";",
			exp: Dedent(`
				SELECT
					"variable value",
					'weird value',
					'var value',
					'var value',
					'var value',
					'var value',
					'weirder value',
					'super weird value';
			`),
			cfg: Config{
				Params: NewMapParams(map[string]string{
					"variable":            "\"variable value\"",
					"a1_2.3$":             "'weird value'",
					"var name":            "'var value'",
					"escaped 'var'":       "'weirder value'",
					"^*& weird \" var   ": "'super weird value'",
				}),
			},
		},
	}

	runFormatterTests(t, tests, NewStandardSQLFormatter)
}

func TestStandardSQLFormatter_FormatPlaceholders(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
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
			name:  "recognizes Snowflake JSON references",
			query: "SELECT foo:bar, foo:bar:baz",
			exp: Dedent(`
				SELECT
					foo:bar,
					foo:bar:baz
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
				Params: NewMapParams(map[string]string{
					"0": "first",
					"1": "second",
					"2": "third",
				}),
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
				Params: NewListParams([]string{
					"first",
					"second",
					"third",
				}),
			},
		},
	}

	runFormatterTests(t, tests, NewStandardSQLFormatter)
}

func TestStandardSQLFormatter_FormatJoins(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats query with GO batch separator",
			query: "SELECT 1 GO SELECT 2",
			exp: Dedent(`
              SELECT
                1
              GO
              SELECT
                2
            `),
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
			name:  "formats simple SELECT with national characters (MSSQL)",
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
			name:  "formats FETCH FIRST like LIMIT",
			query: "SELECT * FETCH FIRST 2 ROWS ONLY;",
			exp: Dedent(`
              SELECT
                *
              FETCH FIRST
                2 ROWS ONLY;
            `),
		},
	}

	runFormatterTests(t, tests, NewStandardSQLFormatter)
}

func TestStandardSQLFormatter_FormatCase(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
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
			name:  "recognizes lowercase CASE ... END",
			query: "case when option = 'foo' then 1 else 2 end;",
			exp: Dedent(`
              case
                when option = 'foo' then 1
                else 2
              end;
            `),
		},
		{
			name:  "ignores words CASE and END inside other strings",
			query: "SELECT CASEDATE, ENDDATE FROM table1;",
			exp: Dedent(`
              SELECT
                CASEDATE,
                ENDDATE
              FROM
                table1;
            `),
		},
	}

	runFormatterTests(t, tests, NewStandardSQLFormatter)
}

func TestStandardSQLFormatter_FormatComments(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats tricky line comments",
			query: "SELECT a#comment, here\nFROM b--comment",
			exp: Dedent(`
              SELECT
                a #comment, here
              FROM
                b --comment
            `),
		},
		{
			name:  "formats line comments followed by semicolon",
			query: "SELECT a FROM b\n--comment\n;",
			exp: Dedent(`
              SELECT
                a
              FROM
                b --comment
              ;
            `),
		},
		{
			name:  "formats line comments followed by comma",
			query: "SELECT a --comment\n, b",
			exp: Dedent(`
              SELECT
                a --comment
              ,
                b
            `),
		},
		{
			name:  "formats line comments followed by close-paren",
			query: "SELECT ( a --comment\n )",
			exp: Dedent(`
              SELECT
                (
                  a --comment
                )
            `),
		},
		{
			name:  "formats line comments followed by open-paren",
			query: "SELECT a --comment\n()",
			exp: Dedent(`
              SELECT
                a --comment
                ()
            `),
		},
		{
			name:  "formats lonely semicolon",
			query: ";",
			exp:   ";",
		},
	}

	runFormatterTests(t, tests, NewStandardSQLFormatter)
}

func TestStandardSQLFormatter_Format(t *testing.T) {
	TestStandardSQLFormatter_FormatDDL(t)
	TestStandardSQLFormatter_FormatVariables(t)
	TestStandardSQLFormatter_FormatPlaceholders(t)
	TestStandardSQLFormatter_FormatJoins(t)
	TestStandardSQLFormatter_FormatCase(t)
	TestStandardSQLFormatter_FormatComments(t)
}
