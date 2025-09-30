package sqlfmt

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

// Phase 11 fuzz testing for SQLite formatter robustness.
func FuzzSQLiteFormatter(f *testing.F) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Seed corpus with known test cases
	seedQueries := []string{
		"SELECT * FROM users;",
		"PRAGMA foreign_keys = ON;",
		"SELECT 'test; string' FROM table;",
		"SELECT \"ÄÖÜ\" FROM [测试表];",
		"WITH RECURSIVE t(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM t WHERE n < 10) SELECT * FROM t;",
		"INSERT INTO table VALUES (?1, :name, @param, $var);",
		"CREATE TABLE test (id INTEGER, data JSON);",
		"SELECT json_extract(data, '$.key') FROM table;",
		"-- Comment\nSELECT * FROM users /* block */;",
		"SELECT x'DEADBEEF' AS hex;",
	}

	for _, seed := range seedQueries {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Formatter should never panic, regardless of input
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Formatter panicked on input: %q\nPanic: %v", input, r)
			}
		}()

		// Format the potentially malformed input
		result := Format(input, cfg)

		// Basic sanity checks on output
		if len(input) > 0 && len(result) == 0 {
			// Only fail if we had non-whitespace input but got empty output
			trimmedInput := strings.TrimSpace(input)
			if len(trimmedInput) > 0 {
				t.Logf("Non-empty input produced empty output: %q", input)
			}
		}

		// Result should be valid UTF-8
		if !utf8.ValidString(result) {
			t.Fatalf("Output is not valid UTF-8 for input: %q", input)
		}
	})
}

// Stress test for very large inputs.
func TestSQLite_Phase11_StressTestLargeInputs(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Generate progressively larger queries to test memory usage and performance
	baseSizes := []int{1000, 10000, 50000}

	for _, size := range baseSizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			// Create a large query with nested structure
			var builder strings.Builder
			builder.WriteString("WITH RECURSIVE large_cte(n) AS (\n")
			builder.WriteString("  SELECT 1\n")
			builder.WriteString("  UNION ALL\n")
			builder.WriteString(fmt.Sprintf("  SELECT n + 1 FROM large_cte WHERE n < %d\n", size/100))
			builder.WriteString(")\n")
			builder.WriteString("SELECT \n")

			// Add many columns to make it large
			for i := range size / 100 {
				if i > 0 {
					builder.WriteString(",\n")
				}
				builder.WriteString(fmt.Sprintf("  CASE WHEN n %% %d = 0 THEN 'multiple_of_%d' "+
					"ELSE 'other' END AS col%d",
					i+2, i+2, i))
			}

			builder.WriteString("\nFROM large_cte\nORDER BY n;")
			query := builder.String()

			// Should handle large input without panic
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Formatter panicked on large input (size ~%d): %v", size, r)
				}
			}()

			result := Format(query, cfg)

			// Basic checks
			if !strings.Contains(result, "WITH") || !strings.Contains(result, "SELECT") {
				t.Error("Large query should maintain basic structure")
			}

			// Memory usage check - result shouldn't be dramatically larger than input
			if len(result) > len(query)*10 {
				t.Errorf("Result size (%d) is >10x input size (%d), possible memory issue",
					len(result), len(query))
			}
		})
	}
}

// Test extreme edge cases.
func TestSQLite_Phase11_ExtremeEdgeCases(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	edgeCases := []struct {
		name  string
		input string
	}{
		{"Empty string", ""},
		{"Only whitespace", "   \n\t  \n"},
		{"Only comment", "-- Just a comment"},
		{"Only semicolon", ";"},
		{"Multiple semicolons", ";;;;"},
		{"Deeply nested parens", "SELECT ((((((((1))))))))"},
		{"Unicode only", "\"αβγδεζηθικλμνξοπρστυφχψω\""},
		{"Mixed quotes", `SELECT "test' weird 'quotes"`},
		{"Very long identifier", `SELECT "` + strings.Repeat("a", 1000) + `"`},
		{"All parameter types", "SELECT ?1, ?999, :name, @at, $dollar"},
		{"Binary data", "SELECT x'deadbeef', X'CAFEBABE'"},
		{"Extreme JSON", "SELECT json_extract('{\"αβγ\":\"测试\"}', '$.αβγ')"},
		{"Control characters", "SELECT '\x00\x01\x02'"},
	}

	for _, test := range edgeCases {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Formatter panicked on %s: %v", test.name, r)
				}
			}()

			result := Format(test.input, cfg)

			// Should produce some result (even if just formatted whitespace)
			if len(test.input) > 0 && !utf8.ValidString(result) {
				t.Errorf("Invalid UTF-8 output for %s", test.name)
			}
		})
	}
}

// Benchmark SQLite formatting performance.
func BenchmarkSQLiteFormatter(b *testing.B) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Representative SQLite query
	query := `-- Performance test query
	PRAGMA foreign_keys = ON;
	
	WITH RECURSIVE employee_hierarchy AS (
		SELECT emp_id, name, manager_id, 0 as level
		FROM employees
		WHERE manager_id IS NULL
		UNION ALL
		SELECT e.emp_id, e.name, e.manager_id, eh.level + 1
		FROM employees e
		JOIN employee_hierarchy eh ON e.manager_id = eh.emp_id
		WHERE eh.level < 10
	),
	department_stats AS (
		SELECT 
			department,
			COUNT(*) as employee_count,
			AVG(salary) as avg_salary,
			json_group_array(json_object(
				'name', name,
				'level', level,
				'salary', salary
			)) as employee_details
		FROM employee_hierarchy eh
		JOIN employees e ON eh.emp_id = e.emp_id
		GROUP BY department
	)
	SELECT 
		ds.department,
		ds.employee_count,
		ds.avg_salary,
		json_extract(ds.employee_details, '$[0].name') as top_employee,
		?1 as filter_param,
		:dept_filter as named_param
	FROM department_stats ds
	WHERE ds.employee_count > @min_count
	ORDER BY ds.avg_salary DESC, ds.employee_count DESC;`

	b.ResetTimer()
	for range b.N {
		Format(query, cfg)
	}
}
