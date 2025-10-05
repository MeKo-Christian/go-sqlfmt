package sqlfmt

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

// Fuzz testing for MySQL formatter robustness.
func FuzzMySQLFormatter(f *testing.F) {
	cfg := &Config{Language: MySQL, Indent: "  "}

	// Seed corpus with known MySQL test cases
	seedQueries := []string{
		"SELECT * FROM users;",
		"SELECT `user_id`, `full_name` FROM `user_table`;",
		"INSERT INTO table VALUES (?, ?, ?);",
		"SELECT 'test; string' FROM table;",
		"SELECT \"ÄÖÜ\" FROM `测试表`;",
		"WITH RECURSIVE t(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM t WHERE n < 10) SELECT * FROM t;",
		"SELECT /*! SQL_CALC_FOUND_ROWS */ * FROM users LIMIT 10;",
		"CREATE TABLE test (id INT AUTO_INCREMENT, data JSON);",
		"SELECT JSON_EXTRACT(data, '$.key') FROM table;",
		"-- Comment\nSELECT * FROM users /* block */;",
		"SELECT 0xDEADBEEF AS hex;",
		"INSERT INTO users SET name = ?, email = ? ON DUPLICATE KEY UPDATE updated_at = NOW();",
		"SELECT * FROM users WHERE MATCH(title, content) AGAINST('search term' IN NATURAL LANGUAGE MODE);",
		"SELECT * FROM users FORCE INDEX(idx_name) WHERE name LIKE 'John%';",
		"SELECT GROUP_CONCAT(name SEPARATOR ', ') FROM users GROUP BY department;",
	}

	for _, seed := range seedQueries {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Formatter should never panic, regardless of input
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("MySQL formatter panicked on input: %q\nPanic: %v", input, r)
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

// Stress test for very large MySQL inputs.
func TestMySQL_StressTestLargeInputs(t *testing.T) {
	cfg := &Config{Language: MySQL, Indent: "  "}

	// Generate progressively larger queries to test memory usage and performance
	baseSizes := []int{1000, 10000, 50000}

	for _, size := range baseSizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			// Create a large MySQL query with complex structure
			var builder strings.Builder
			builder.WriteString("SELECT /*! SQL_CALC_FOUND_ROWS */\n")

			// Add many columns with MySQL-specific functions
			for i := range size / 100 {
				if i > 0 {
					builder.WriteString(",\n")
				}
				builder.WriteString(fmt.Sprintf("  CASE WHEN id %% %d = 0 THEN CONCAT('multiple_of_', %d) "+
					"WHEN JSON_EXTRACT(settings, '$.flag') = 'true' THEN 'json_flag_set' "+
					"ELSE COALESCE(name, 'unknown') END AS col%d",
					i+2, i+2, i))
			}

			builder.WriteString("\nFROM users u\n")
			builder.WriteString("LEFT JOIN user_profiles p ON u.id = p.user_id\n")
			builder.WriteString("LEFT JOIN user_settings s ON u.id = s.user_id\n")
			builder.WriteString("WHERE u.active = 1\n")
			builder.WriteString("AND MATCH(u.name, u.email) AGAINST(? IN NATURAL LANGUAGE MODE)\n")
			builder.WriteString("AND JSON_CONTAINS(s.settings, '{\"notifications\": true}')\n")
			builder.WriteString("ORDER BY u.created_at DESC\n")
			builder.WriteString("LIMIT 1000;")

			query := builder.String()

			// Should handle large input without panic
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("MySQL formatter panicked on large input (size ~%d): %v", size, r)
				}
			}()

			result := Format(query, cfg)

			// Basic checks
			if !strings.Contains(result, "SELECT") || !strings.Contains(result, "FROM") {
				t.Error("Large MySQL query should maintain basic structure")
			}

			// Memory usage check - result shouldn't be dramatically larger than input
			if len(result) > len(query)*10 {
				t.Errorf("Result size (%d) is >10x input size (%d), possible memory issue",
					len(result), len(query))
			}
		})
	}
}

// Test extreme edge cases for MySQL.
func TestMySQL_ExtremeEdgeCases(t *testing.T) {
	cfg := &Config{Language: MySQL, Indent: "  "}

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
		{"Unicode only", "`αβγδεζηθικλμνξοπρστυφχψω`"},
		{"Mixed quotes and backticks", "SELECT `test' \"weird\" 'quotes`"},
		{"Very long identifier", "SELECT `" + strings.Repeat("a", 1000) + "`"},
		{"All parameter types", "SELECT ?, ?, :name, @var, $dollar"},
		{"Binary data", "SELECT 0xDEADBEEF, X'CAFEBABE'"},
		{"Extreme JSON", "SELECT JSON_EXTRACT('[{\"αβγ\":\"测试\"}]', '$[0].αβγ')"},
		{"Control characters", "SELECT '\\x00\\x01\\x02'"},
		{"MySQL hints", "SELECT /*+ INDEX(users idx_name) */ * FROM users"},
		{"Complex ON DUPLICATE KEY", "INSERT INTO t VALUES (1) ON DUPLICATE KEY UPDATE c=c+1, d=VALUES(c)+1"},
		{"Fulltext search", "SELECT MATCH(title,content) AGAINST ('search' IN BOOLEAN MODE)"},
		{"Window functions", "SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) FROM users"},
	}

	for _, test := range edgeCases {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("MySQL formatter panicked on %s: %v", test.name, r)
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

// Benchmark MySQL formatting performance.
func BenchmarkMySQLFormatter(b *testing.B) {
	cfg := &Config{Language: MySQL, Indent: "  "}

	// Representative MySQL query with various MySQL-specific features
	query := `-- MySQL Performance test query
	SELECT /*! SQL_CALC_FOUND_ROWS */
		u.id,
		u.username,
		u.email,
		u.created_at,
		p.first_name,
		p.last_name,
		p.settings,
		CASE
			WHEN u.status = 'active' THEN 'Active User'
			WHEN u.status = 'inactive' THEN 'Inactive User'
			ELSE 'Unknown Status'
		END as status_description,
		JSON_EXTRACT(p.settings, '$.theme') as theme,
		JSON_CONTAINS(p.settings, '"notifications"', '$.preferences') as has_notifications,
		MATCH(u.username, u.email) AGAINST('john' IN NATURAL LANGUAGE MODE) as relevance_score
	FROM users u
	LEFT JOIN user_profiles p ON u.id = p.user_id
	WHERE u.active = ?
		AND u.email_verified = ?
		AND p.created_at > ?
		AND JSON_EXTRACT(p.settings, '$.account_type') IN ('premium', 'vip')
	ORDER BY u.username ASC, p.created_at DESC
	LIMIT 100;`

	b.ResetTimer()
	for range b.N {
		Format(query, cfg)
	}
}
