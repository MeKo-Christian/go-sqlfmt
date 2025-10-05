package sqlfmt

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

// Fuzz testing for PostgreSQL formatter robustness.
func FuzzPostgreSQLFormatter(f *testing.F) {
	cfg := &Config{Language: PostgreSQL, Indent: "  "}

	// Seed corpus with known PostgreSQL test cases
	seedQueries := []string{
		"SELECT * FROM users;",
		`SELECT "user_id", "full_name" FROM "user_table";`,
		"SELECT $1, $2, $3 FROM table;",
		"SELECT 'test; string' FROM table;",
		`SELECT "ÄÖÜ" FROM "测试表";`,
		"WITH RECURSIVE t(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM t WHERE n < 10) SELECT * FROM t;",
		"SELECT * FROM users WHERE created_at > CURRENT_TIMESTAMP - INTERVAL '30 days';",
		"CREATE TABLE test (id SERIAL, data JSONB);",
		"SELECT data->'key', data->>'key' FROM table;",
		"-- Comment\nSELECT * FROM users /* block */;",
		"SELECT E'\\xDEADBEEF' AS hex;",
		"INSERT INTO users VALUES (DEFAULT, $1, $2) RETURNING id;",
		"SELECT * FROM users WHERE name ILIKE '%john%';",
		"SELECT array_agg(name ORDER BY name) FROM users GROUP BY department;",
		"SELECT * FROM generate_series(1, 10) AS t(n);",
	}

	for _, seed := range seedQueries {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Formatter should never panic, regardless of input
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("PostgreSQL formatter panicked on input: %q\nPanic: %v", input, r)
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

// Stress test for very large PostgreSQL inputs.
func TestPostgreSQL_StressTestLargeInputs(t *testing.T) {
	cfg := &Config{Language: PostgreSQL, Indent: "  "}

	// Generate progressively larger queries to test memory usage and performance
	baseSizes := []int{1000, 10000, 50000}

	for _, size := range baseSizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			// Create a large PostgreSQL query with complex structure
			var builder strings.Builder
			builder.WriteString("WITH RECURSIVE large_cte(n, data) AS (\n")
			builder.WriteString("  SELECT 1, '{\"start\": true}'::jsonb\n")
			builder.WriteString("  UNION ALL\n")
			builder.WriteString(fmt.Sprintf("  SELECT n + 1, data || jsonb_build_object('step', n) "+
				"FROM large_cte WHERE n < %d\n", size/100))
			builder.WriteString(")\n")
			builder.WriteString("SELECT \n")

			// Add many columns with PostgreSQL-specific functions
			for i := range size / 100 {
				if i > 0 {
					builder.WriteString(",\n")
				}
				builder.WriteString(fmt.Sprintf("  CASE WHEN n %% %d = 0 THEN CONCAT('multiple_of_', %d::text) "+
					"WHEN data->>'flag' = 'true' THEN 'jsonb_flag_set' "+
					"ELSE COALESCE(name, 'unknown') END AS col%d",
					i+2, i+2, i))
			}

			builder.WriteString("\nFROM users u\n")
			builder.WriteString("LEFT JOIN user_profiles p ON u.id = p.user_id\n")
			builder.WriteString("LEFT JOIN large_cte l ON l.n <= 10\n")
			builder.WriteString("WHERE u.active = true\n")
			builder.WriteString("AND u.name ILIKE $1\n")
			builder.WriteString("AND p.settings @> '{\"notifications\": true}'::jsonb\n")
			builder.WriteString("AND u.created_at > CURRENT_TIMESTAMP - INTERVAL '30 days'\n")
			builder.WriteString("ORDER BY u.created_at DESC\n")
			builder.WriteString("LIMIT 1000;")

			query := builder.String()

			// Should handle large input without panic
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("PostgreSQL formatter panicked on large input (size ~%d): %v", size, r)
				}
			}()

			result := Format(query, cfg)

			// Basic checks
			if !strings.Contains(result, "WITH") || !strings.Contains(result, "SELECT") {
				t.Error("Large PostgreSQL query should maintain basic structure")
			}

			// Memory usage check - result shouldn't be dramatically larger than input
			if len(result) > len(query)*10 {
				t.Errorf("Result size (%d) is >10x input size (%d), possible memory issue",
					len(result), len(query))
			}
		})
	}
}

// Test extreme edge cases for PostgreSQL.
func TestPostgreSQL_ExtremeEdgeCases(t *testing.T) {
	cfg := &Config{Language: PostgreSQL, Indent: "  "}

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
		{"Unicode only", `"αβγδεζηθικλμνξοπρστυφχψω"`},
		{"Mixed quotes", `SELECT "test' 'weird' "quotes"`},
		{"Very long identifier", `SELECT "` + strings.Repeat("a", 1000) + `"`},
		{"All parameter types", "SELECT $1, $2, :name, @var, ?placeholder"},
		{"Binary data", "SELECT E'\\xDEADBEEF', '\\xCAFEBABE'"},
		{"Extreme JSONB", "SELECT '[{\"αβγ\":\"测试\"}]'::jsonb->0->'αβγ'"},
		{"Control characters", "SELECT E'\\x00\\x01\\x02'"},
		{"CTEs and window functions", "WITH t AS (SELECT 1) SELECT ROW_NUMBER() OVER () FROM t"},
		{"Complex RETURNING", "INSERT INTO t VALUES (1) RETURNING id, created_at"},
		{"Array operations", "SELECT ARRAY[1,2,3][1:2], array_agg(name) FROM users"},
		{"Interval operations", "SELECT CURRENT_TIMESTAMP + INTERVAL '1 day 2 hours'"},
		{"Dollar quoting", `SELECT $$test 'string' $$`},
		{"Schema qualified", `SELECT * FROM "schema"."table" WHERE "column" = $1`},
	}

	for _, test := range edgeCases {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("PostgreSQL formatter panicked on %s: %v", test.name, r)
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

// Benchmark PostgreSQL formatting performance.
func BenchmarkPostgreSQLFormatter(b *testing.B) {
	cfg := &Config{Language: PostgreSQL, Indent: "  "}

	// Representative PostgreSQL query with various PostgreSQL-specific features
	query := `-- PostgreSQL Performance test query
	WITH RECURSIVE employee_hierarchy AS (
		SELECT
			emp_id,
			name,
			manager_id,
			department,
			salary,
			hire_date,
			0 as level,
			ARRAY[emp_id] as path
		FROM employees
		WHERE manager_id IS NULL
		UNION ALL
		SELECT
			e.emp_id,
			e.name,
			e.manager_id,
			e.department,
			e.salary,
			e.hire_date,
			eh.level + 1,
			eh.path || e.emp_id
		FROM employees e
		JOIN employee_hierarchy eh ON e.manager_id = eh.emp_id
		WHERE eh.level < 10
	),
	department_stats AS (
		SELECT
			department,
			COUNT(*) as total_employees,
			AVG(salary) as avg_salary,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY salary) as median_salary,
			STRING_AGG(name, ', ' ORDER BY salary DESC) as top_earners,
			jsonb_agg(jsonb_build_object(
				'name', name,
				'level', level,
				'salary', salary
			)) as employee_details
		FROM employee_hierarchy
		GROUP BY department
	)
	SELECT
		eh.name,
		eh.department,
		eh.level,
		eh.salary,
		ds.avg_salary,
		ds.median_salary,
		RANK() OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as dept_salary_rank,
		DENSE_RANK() OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as dept_dense_rank,
		ROW_NUMBER() OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as dept_row_num,
		LAG(eh.salary) OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as next_lower_salary,
		LEAD(eh.salary) OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as next_higher_salary,
		COUNT(*) OVER (PARTITION BY eh.department) as dept_size,
		SUM(eh.salary) OVER (PARTITION BY eh.department) as dept_total_salary,
		CASE
			WHEN eh.salary > ds.avg_salary THEN 'Above Average'
			WHEN eh.salary = ds.avg_salary THEN 'Average'
			ELSE 'Below Average'
		END as salary_category,
		EXTRACT(YEAR FROM AGE(CURRENT_DATE, eh.hire_date)) as years_of_service,
		eh.path as hierarchy_path
	FROM employee_hierarchy eh
	JOIN department_stats ds ON eh.department = ds.department
	WHERE eh.salary > ds.median_salary
	ORDER BY eh.department, eh.salary DESC, eh.name
	LIMIT 100;`

	b.ResetTimer()
	for range b.N {
		Format(query, cfg)
	}
}
