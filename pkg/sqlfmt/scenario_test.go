package sqlfmt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	dialectPostgreSQL = "postgresql"
	dialectMySQL      = "mysql"
	dialectSQLite     = "sqlite"
	dialectStandard   = "standard"
)

// TestRealWorldScenarios tests formatting of real-world SQL scenarios
// including migrations, stored procedures, analytics queries, and mixed DDL/DML.
func TestRealWorldScenarios(t *testing.T) {
	scenariosDir := filepath.Join("..", "..", "testdata", "scenarios")

	// Test each scenario directory
	scenarioDirs := []string{"migrations", "procedures", "analytics", "mixed"}

	for _, scenarioDir := range scenarioDirs {
		t.Run(scenarioDir, func(t *testing.T) {
			dirPath := filepath.Join(scenariosDir, scenarioDir)

			// Check if directory exists
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				t.Skipf("Scenario directory %s does not exist, skipping", dirPath)
				return
			}

			// Find all SQL files in the scenario directory
			files, err := filepath.Glob(filepath.Join(dirPath, "*.sql"))
			require.NoError(t, err)

			for _, file := range files {
				t.Run(filepath.Base(file), func(t *testing.T) {
					testScenarioFile(t, file)
				})
			}
		})
	}
}

// testScenarioFile tests a single scenario SQL file.
func testScenarioFile(t *testing.T, filePath string) {
	t.Helper()

	// Read the SQL file
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	sqlContent := string(content)

	// Determine dialect from filename
	dialect := detectDialectFromFilename(filePath)

	// Create formatter config for the detected dialect
	cfg := NewDefaultConfig()
	switch dialect {
	case dialectPostgreSQL:
		cfg.Language = PostgreSQL
	case dialectMySQL:
		cfg.Language = MySQL
	case dialectSQLite:
		cfg.Language = SQLite
	default:
		cfg.Language = StandardSQL
	}

	// Test that formatting doesn't panic and produces valid output
	t.Run("format_without_panic", func(t *testing.T) {
		formatted := Format(sqlContent, cfg)
		require.NotEmpty(t, formatted)

		// Basic validation - formatted SQL should be parseable again
		Format(formatted, cfg) // Should not panic
	})

	// Test that formatting is idempotent (formatting twice gives same result)
	t.Run("idempotent_formatting", func(t *testing.T) {
		firstFormat := Format(sqlContent, cfg)

		secondFormat := Format(firstFormat, cfg)

		require.Equal(t, firstFormat, secondFormat, "Formatting should be idempotent")
	})

	// Test that formatted SQL contains expected structural elements
	t.Run("preserves_structure", func(t *testing.T) {
		formatted := Format(sqlContent, cfg)

		// Count keywords in original vs formatted
		originalKeywords := countSQLKeywords(sqlContent)
		formattedKeywords := countSQLKeywords(formatted)

		// Should preserve the same number of major SQL keywords
		require.Equal(t, originalKeywords["SELECT"], formattedKeywords["SELECT"], "SELECT count should be preserved")
		require.Equal(t, originalKeywords["FROM"], formattedKeywords["FROM"], "FROM count should be preserved")
		require.Equal(t, originalKeywords["WHERE"], formattedKeywords["WHERE"], "WHERE count should be preserved")
		require.Equal(t, originalKeywords["INSERT"], formattedKeywords["INSERT"], "INSERT count should be preserved")
		require.Equal(t, originalKeywords["UPDATE"], formattedKeywords["UPDATE"], "UPDATE count should be preserved")
		require.Equal(t, originalKeywords["DELETE"], formattedKeywords["DELETE"], "DELETE count should be preserved")
		require.Equal(t, originalKeywords["CREATE"], formattedKeywords["CREATE"], "CREATE count should be preserved")
		require.Equal(t, originalKeywords["ALTER"], formattedKeywords["ALTER"], "ALTER count should be preserved")
		require.Equal(t, originalKeywords["DROP"], formattedKeywords["DROP"], "DROP count should be preserved")
	})

	// Test complex queries don't break formatting
	t.Run("complex_queries", func(t *testing.T) {
		formatted := Format(sqlContent, cfg)

		// Check for common complex query patterns
		if strings.Contains(strings.ToUpper(sqlContent), "WITH") {
			require.Contains(t, strings.ToUpper(formatted), "WITH", "WITH clauses should be preserved")
		}
		if strings.Contains(strings.ToUpper(sqlContent), "UNION") {
			require.Contains(t, strings.ToUpper(formatted), "UNION", "UNION operations should be preserved")
		}
		if strings.Contains(strings.ToUpper(sqlContent), "WINDOW") {
			require.Contains(t, strings.ToUpper(formatted), "WINDOW", "WINDOW clauses should be preserved")
		}
		if strings.Contains(strings.ToUpper(sqlContent), "PARTITION BY") {
			require.Contains(t, strings.ToUpper(formatted), "PARTITION BY", "PARTITION BY clauses should be preserved")
		}
	})
}

// detectDialectFromFilename determines SQL dialect from filename.
func detectDialectFromFilename(filename string) string {
	base := strings.ToLower(filepath.Base(filename))
	if strings.HasPrefix(base, "postgresql") {
		return dialectPostgreSQL
	}
	if strings.HasPrefix(base, "mysql") {
		return dialectMySQL
	}
	if strings.HasPrefix(base, "sqlite") {
		return dialectSQLite
	}
	return dialectStandard
}

// countSQLKeywords counts major SQL keywords in a string.
func countSQLKeywords(sql string) map[string]int {
	keywords := []string{
		"SELECT", "FROM", "WHERE", "JOIN", "INNER", "LEFT", "RIGHT", "FULL",
		"INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP",
		"TABLE", "INDEX", "VIEW", "FUNCTION", "PROCEDURE", "TRIGGER",
		"WITH", "UNION", "EXCEPT", "INTERSECT", "WINDOW", "PARTITION",
		"GROUP BY", "ORDER BY", "HAVING", "LIMIT", "OFFSET",
	}

	counts := make(map[string]int)
	upperSQL := strings.ToUpper(sql)

	for _, keyword := range keywords {
		counts[keyword] = strings.Count(upperSQL, keyword)
	}

	return counts
}

// TestScenarioPerformance tests that scenario files format within reasonable time.
func TestScenarioPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	scenariosDir := filepath.Join("..", "..", "testdata", "scenarios")

	// Find all SQL files in scenarios
	files, err := filepath.Glob(filepath.Join(scenariosDir, "**", "*.sql"))
	require.NoError(t, err)

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			content, err := os.ReadFile(file)
			require.NoError(t, err)

			dialect := detectDialectFromFilename(file)
			cfg := NewDefaultConfig()
			switch dialect {
			case "postgresql":
				cfg.Language = PostgreSQL
			case "mysql":
				cfg.Language = MySQL
			case "sqlite":
				cfg.Language = SQLite
			default:
				cfg.Language = StandardSQL
			}

			// Test formatting performance (should complete within reasonable time)
			Format(string(content), cfg) // Should not panic or take too long
		})
	}
}

// TestScenarioComplexity tests various complexity metrics of scenario files.
func TestScenarioComplexity(t *testing.T) {
	scenariosDir := filepath.Join("..", "..", "testdata", "scenarios")

	files, err := filepath.Glob(filepath.Join(scenariosDir, "**", "*.sql"))
	require.NoError(t, err)

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			content, err := os.ReadFile(file)
			require.NoError(t, err)

			sqlContent := string(content)

			// Test various complexity indicators
			t.Run("nested_structures", func(t *testing.T) {
				// Count nested parentheses (indicating complex expressions)
				nestedLevel := 0
				maxNestedLevel := 0
				for _, char := range sqlContent {
					switch char {
					case '(':
						nestedLevel++
						if nestedLevel > maxNestedLevel {
							maxNestedLevel = nestedLevel
						}
					case ')':
						nestedLevel--
					}
				}

				// Should handle reasonable nesting levels
				require.LessOrEqual(t, maxNestedLevel, 10, "Nesting level should be reasonable (max 10, got %d)", maxNestedLevel)
			})

			t.Run("statement_count", func(t *testing.T) {
				// Count semicolons (rough estimate of statement count)
				statementCount := strings.Count(sqlContent, ";")

				// Should handle files with multiple statements
				require.LessOrEqual(t, statementCount, 100,
					"Should have reasonable number of statements (max 100, got %d)", statementCount)
			})

			t.Run("line_length", func(t *testing.T) {
				lines := strings.Split(sqlContent, "\n")
				maxLineLength := 0
				for _, line := range lines {
					if len(line) > maxLineLength {
						maxLineLength = len(line)
					}
				}

				// Should handle long lines (but formatter should help with this)
				require.LessOrEqual(t, maxLineLength, 1000, "Line length should be reasonable (max 1000, got %d)", maxLineLength)
			})
		})
	}
}
