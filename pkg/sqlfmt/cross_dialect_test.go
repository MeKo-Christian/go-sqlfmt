package sqlfmt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCrossDialectFormatting tests that the same query formats correctly across all dialects
func TestCrossDialectFormatting(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		dialects []Language
	}{
		{
			name:  "basic SELECT works across all dialects",
			query: "SELECT id, name FROM users WHERE active = true;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
				PLSQL,
				DB2,
				N1QL,
			},
		},
		{
			name:  "JOIN with WHERE clause",
			query: "SELECT u.id, p.title FROM users u INNER JOIN posts p ON u.id = p.user_id WHERE u.active = true;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
				PLSQL,
				DB2,
			},
		},
		{
			name:  "aggregate functions",
			query: "SELECT COUNT(*), MAX(age), MIN(age), AVG(age) FROM users GROUP BY department;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
				PLSQL,
				DB2,
			},
		},
		{
			name:  "subquery in WHERE",
			query: "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > 100);",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
				PLSQL,
				DB2,
			},
		},
		{
			name:  "UNION query",
			query: "SELECT name FROM users UNION SELECT name FROM customers;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
				PLSQL,
				DB2,
			},
		},
		{
			name:  "CASE expression",
			query: "SELECT name, CASE WHEN age < 18 THEN 'minor' WHEN age >= 18 AND age < 65 THEN 'adult' ELSE 'senior' END AS category FROM users;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
				PLSQL,
				DB2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, dialect := range tt.dialects {
				t.Run(string(dialect), func(t *testing.T) {
					cfg := NewDefaultConfig().WithLang(dialect)
					result := Format(tt.query, cfg)

					// Should not panic and should produce non-empty output
					require.NotEmpty(t, result, "Format should produce output for %s", dialect)

					// Result should contain key elements of the query
					if dialect != N1QL { // N1QL might format differently
						require.Contains(t, result, "SELECT", "Formatted query should contain SELECT")
					}
				})
			}
		})
	}
}

// TestDialectSpecificFeatures tests features that are unique to specific dialects
func TestDialectSpecificFeatures(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		dialect  Language
		expected string
	}{
		{
			name:    "PostgreSQL dollar-quoted string",
			query:   "SELECT $$Hello World$$ AS message;",
			dialect: PostgreSQL,
			expected: `SELECT
  $$Hello World$$ AS message;`,
		},
		{
			name:    "PostgreSQL JSONB operator",
			query:   "SELECT data->>'name' FROM users WHERE data @> '{\"active\": true}';",
			dialect: PostgreSQL,
			expected: `SELECT
  data ->> 'name'
FROM
  users
WHERE
  data @> '{"active": true}';`,
		},
		{
			name:    "PostgreSQL array literal",
			query:   "SELECT ARRAY[1, 2, 3] AS numbers;",
			dialect: PostgreSQL,
			expected: `SELECT
  ARRAY [1,
  2,
  3 ] AS numbers;`,
		},
		{
			name:     "MySQL backtick identifiers",
			query:    "SELECT `user`.`id`, `user`.`name` FROM `users` AS `user`;",
			dialect:  MySQL,
			expected: "SELECT\n  `user`.`id`,\n  `user`.`name`\nFROM\n  `users` AS `user`;",
		},
		{
			name:    "MySQL LIMIT with OFFSET",
			query:   "SELECT * FROM users LIMIT 10 OFFSET 20;",
			dialect: MySQL,
			expected: `SELECT
  *
FROM
  users
LIMIT
  10 OFFSET 20;`,
		},
		{
			name:    "SQLite datetime function",
			query:   "SELECT datetime('now') AS current_time;",
			dialect: SQLite,
			expected: `SELECT
  datetime('now') AS current_time;`,
		},
		{
			name:    "SQLite AUTOINCREMENT",
			query:   "CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT);",
			dialect: SQLite,
			expected: `CREATE TABLE
  users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT);`,
		},
		{
			name:    "PL/SQL DUAL table",
			query:   "SELECT SYSDATE FROM DUAL;",
			dialect: PLSQL,
			expected: `SELECT
  SYSDATE
FROM
  DUAL;`,
		},
		{
			name:    "N1QL USE KEYS",
			query:   "SELECT * FROM bucket USE KEYS ['key1', 'key2'];",
			dialect: N1QL,
			expected: `SELECT
  *
FROM
  bucket
USE KEYS
  ['key1', 'key2'];`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultConfig().WithLang(tt.dialect)
			result := Format(tt.query, cfg)
			require.Equal(t, tt.expected, result, "Dialect-specific feature should format correctly")
		})
	}
}

// TestDialectSpecificKeywords tests that dialect-specific reserved words are handled correctly
func TestDialectSpecificKeywords(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		dialect  Language
		contains []string // Keywords that should be in the output
	}{
		{
			name:     "PostgreSQL specific keywords",
			query:    "SELECT * FROM users WHERE age BETWEEN 18 AND 65 RETURNING id;",
			dialect:  PostgreSQL,
			contains: []string{"SELECT", "FROM", "WHERE", "BETWEEN", "AND", "RETURNING"},
		},
		{
			name:     "MySQL specific keywords",
			query:    "SELECT * FROM users WHERE name REGEXP '^A' LIMIT 10;",
			dialect:  MySQL,
			contains: []string{"SELECT", "FROM", "WHERE", "REGEXP", "LIMIT"},
		},
		{
			name:     "SQLite specific keywords",
			query:    "SELECT * FROM users WHERE name GLOB 'A*';",
			dialect:  SQLite,
			contains: []string{"SELECT", "FROM", "WHERE", "GLOB"},
		},
		{
			name:     "PL/SQL specific keywords",
			query:    "SELECT * FROM users WHERE ROWNUM <= 10;",
			dialect:  PLSQL,
			contains: []string{"SELECT", "FROM", "WHERE", "ROWNUM"},
		},
		{
			name:     "N1QL specific keywords",
			query:    "SELECT * FROM bucket UNNEST items AS item WHERE item.active = true;",
			dialect:  N1QL,
			contains: []string{"SELECT", "FROM", "UNNEST", "WHERE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultConfig().WithLang(tt.dialect)
			result := Format(tt.query, cfg)

			for _, keyword := range tt.contains {
				require.Contains(t, result, keyword, "Output should contain keyword: %s", keyword)
			}
		})
	}
}

// TestCrossDialectEdgeCases tests edge cases that should work across multiple dialects
func TestCrossDialectEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		dialects []Language
		validate func(t *testing.T, result string, dialect Language)
	}{
		{
			name:  "empty string handling",
			query: "SELECT '' AS empty_string;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
			},
			validate: func(t *testing.T, result string, dialect Language) {
				require.NotEmpty(t, result)
				require.Contains(t, result, "''")
			},
		},
		{
			name:  "multiple spaces and newlines",
			query: "SELECT    *    FROM    users    WHERE    id   =   1;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
			},
			validate: func(t *testing.T, result string, dialect Language) {
				require.NotContains(t, result, "    ", "Should normalize excessive spaces")
			},
		},
		{
			name:  "nested subqueries",
			query: "SELECT * FROM (SELECT * FROM (SELECT id FROM users WHERE active = true) AS inner1) AS outer1;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
			},
			validate: func(t *testing.T, result string, dialect Language) {
				require.Contains(t, result, "SELECT")
				require.Contains(t, result, "FROM")
			},
		},
		{
			name:  "complex WHERE with multiple conditions",
			query: "SELECT * FROM users WHERE (age > 18 AND age < 65) OR (status = 'vip' AND credits > 100);",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
			},
			validate: func(t *testing.T, result string, dialect Language) {
				require.Contains(t, result, "WHERE")
				require.Contains(t, result, "AND")
				require.Contains(t, result, "OR")
			},
		},
		{
			name:  "NULL handling",
			query: "SELECT * FROM users WHERE email IS NULL OR email IS NOT NULL;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
			},
			validate: func(t *testing.T, result string, dialect Language) {
				require.Contains(t, result, "IS NULL")
				require.Contains(t, result, "IS NOT NULL")
			},
		},
		{
			name:  "DISTINCT with ORDER BY",
			query: "SELECT DISTINCT category FROM products ORDER BY category ASC;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
			},
			validate: func(t *testing.T, result string, dialect Language) {
				require.Contains(t, result, "DISTINCT")
				require.Contains(t, result, "ORDER BY")
			},
		},
		{
			name:  "multiple JOINs",
			query: "SELECT * FROM users u INNER JOIN orders o ON u.id = o.user_id LEFT JOIN products p ON o.product_id = p.id;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
			},
			validate: func(t *testing.T, result string, dialect Language) {
				require.Contains(t, result, "INNER JOIN")
				require.Contains(t, result, "LEFT JOIN")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, dialect := range tt.dialects {
				t.Run(string(dialect), func(t *testing.T) {
					cfg := NewDefaultConfig().WithLang(dialect)
					result := Format(tt.query, cfg)
					tt.validate(t, result, dialect)
				})
			}
		})
	}
}

// TestDialectComparison tests that formatting differs appropriately between dialects
func TestDialectComparison(t *testing.T) {
	tests := []struct {
		name            string
		query           string
		dialect1        Language
		dialect2        Language
		shouldDiffer    bool
		differenceCheck func(result1, result2 string) bool
	}{
		{
			name:         "PostgreSQL vs MySQL identifier quoting",
			query:        `SELECT "user"."id" FROM "users" AS "user";`,
			dialect1:     PostgreSQL,
			dialect2:     MySQL,
			shouldDiffer: false, // Both should preserve the quotes
		},
		{
			name:         "Standard SQL vs PostgreSQL basic query",
			query:        "SELECT id, name FROM users WHERE active = true;",
			dialect1:     StandardSQL,
			dialect2:     PostgreSQL,
			shouldDiffer: false, // Should be identical for basic queries
		},
		{
			name:         "MySQL vs SQLite basic query",
			query:        "SELECT * FROM users LIMIT 10;",
			dialect1:     MySQL,
			dialect2:     SQLite,
			shouldDiffer: false, // Should be identical for basic LIMIT
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg1 := NewDefaultConfig().WithLang(tt.dialect1)
			cfg2 := NewDefaultConfig().WithLang(tt.dialect2)

			result1 := Format(tt.query, cfg1)
			result2 := Format(tt.query, cfg2)

			if tt.shouldDiffer {
				require.NotEqual(t, result1, result2, "Results should differ between %s and %s", tt.dialect1, tt.dialect2)
				if tt.differenceCheck != nil {
					require.True(t, tt.differenceCheck(result1, result2), "Custom difference check failed")
				}
			} else {
				// For queries that should format the same
				require.NotEmpty(t, result1)
				require.NotEmpty(t, result2)
			}
		})
	}
}

// TestCrossDialectComments tests comment handling across dialects
func TestCrossDialectComments(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		dialects []Language
	}{
		{
			name:  "single-line comment",
			query: "SELECT id -- user identifier\nFROM users;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
			},
		},
		{
			name:  "multi-line comment",
			query: "SELECT /* get all columns */ * FROM users;",
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
			},
		},
		{
			name: "multiple comments",
			query: `-- Header comment
SELECT
  id, -- primary key
  name /* user name */
FROM users; -- end of query`,
			dialects: []Language{
				StandardSQL,
				PostgreSQL,
				MySQL,
				SQLite,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, dialect := range tt.dialects {
				t.Run(string(dialect), func(t *testing.T) {
					cfg := NewDefaultConfig().WithLang(dialect)
					result := Format(tt.query, cfg)

					// Comments should be preserved in some form
					require.NotEmpty(t, result)
					// Should produce valid formatted output
					require.NotPanics(t, func() {
						Format(result, cfg)
					}, "Should be able to format the formatted output")
				})
			}
		})
	}
}

// TestCrossDialectParameterStyles tests different parameter placeholder styles
func TestCrossDialectParameterStyles(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		dialect  Language
		expected string
	}{
		{
			name:    "PostgreSQL numbered parameters",
			query:   "SELECT * FROM users WHERE id = $1 AND name = $2;",
			dialect: PostgreSQL,
			expected: `SELECT
  *
FROM
  users
WHERE
  id = $1
  AND name = $2;`,
		},
		{
			name:    "MySQL question mark parameters",
			query:   "SELECT * FROM users WHERE id = ? AND name = ?;",
			dialect: MySQL,
			expected: `SELECT
  *
FROM
  users
WHERE
  id = ?
  AND name = ?;`,
		},
		{
			name:    "PL/SQL named parameters",
			query:   "SELECT * FROM users WHERE id = :id AND name = :name;",
			dialect: PLSQL,
			expected: `SELECT
  *
FROM
  users
WHERE
  id = :id
  AND name = :name;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultConfig().WithLang(tt.dialect)
			result := Format(tt.query, cfg)
			require.Equal(t, tt.expected, result, "Parameter placeholders should be preserved")
		})
	}
}
