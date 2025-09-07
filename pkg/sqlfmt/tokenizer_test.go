package sqlfmt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTokenizerRegexes tests the tokenizer's regex matching through the public API.
// This tests the tokenizer indirectly by verifying that inputs are tokenized correctly.
func TestTokenizerRegexes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected formatted output that proves correct tokenization
	}{
		// Word regex tests
		{name: "simple word", input: "TEXT", expected: "TEXT"},
		{name: "word with punctuation", input: "TEXT);", expected: "TEXT\n);"},
		{name: "word with space", input: "TEXT ", expected: "TEXT"},
		{name: "word with newline", input: "table\nWHERE", expected: "table\nWHERE"},
		{name: "word with carriage return", input: "table\rWHERE", expected: "table\nWHERE"},
		{name: "word with PostgreSQL cast", input: "column::int", expected: "column :: int"},

		// Indexed placeholder tests
		{name: "question mark placeholder", input: "SELECT * WHERE id = ?", expected: "SELECT\n  *\nWHERE\n  id = ?"},
		{name: "numbered placeholder", input: "SELECT * WHERE id = ?1", expected: "SELECT\n  *\nWHERE\n  id = ?1"},
		{name: "multi digit placeholder", input: "SELECT * WHERE id = ?22", expected: "SELECT\n  *\nWHERE\n  id = ?22"},

		// Named placeholder tests
		{name: "at variable", input: "SELECT * WHERE name = @variable", expected: "SELECT\n  *\nWHERE\n  name = @variable"},
		{
			name:     "complex variable name",
			input:    "SELECT * WHERE name = @a1_2.3$",
			expected: "SELECT\n  *\nWHERE\n  name = @a1_2.3$",
		},

		// String named placeholder tests
		{
			name:     "single quoted placeholder",
			input:    "SELECT * WHERE name = @'var name'",
			expected: "SELECT\n  *\nWHERE\n  name = @'var name'",
		},
		{name: "double quoted placeholder",
			input: "SELECT * WHERE name = @\"var name\"",
			expected: "SELECT\n" +
				"  *\n" +
				"WHERE\n" +
				"  name = @\"var name\""},
		{name: "backtick quoted placeholder",
			input: "SELECT * WHERE name = @`var name`",
			expected: "SELECT\n" +
				"  *\n" +
				"WHERE\n" +
				"  name = @`var name`"},
		{name: "bracket quoted placeholder",
			input: "SELECT * WHERE name = @[var name]",
			expected: "SELECT\n" +
				"  *\n" +
				"WHERE\n" +
				"  name = @[var name]"},

		// Reserved top level no-indent tests
		{name: "union all", input: "SELECT 1 UNION ALL SELECT 2", expected: "SELECT\n  1\nUNION ALL\nSELECT\n  2"},

		// Boolean tests
		{name: "true boolean", input: "SELECT true", expected: "SELECT\n  true"},
		{name: "false boolean", input: "SELECT false", expected: "SELECT\n  false"},
		{name: "TRUE uppercase", input: "SELECT TRUE", expected: "SELECT\n  TRUE"},
		{name: "mixed case true", input: "SELECT tRUE", expected: "SELECT\n  tRUE"},

		// Function call tests
		{name: "simple function call", input: "SELECT call()", expected: "SELECT\n  call()"},
		{name: "function with args",
			input: "SELECT CALL_WITH_ARGS(" +
				"arg1, 3+4, arg2)",
			expected: "SELECT\n" +
				"  CALL_WITH_ARGS(arg1, 3 + 4, arg2)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the public Format function to test tokenization indirectly
			actual := Format(tt.input)
			// The fact that Format works without errors proves the tokenizer is working
			require.NotEmpty(t, actual, "Format should return non-empty result")
			// For specific cases where we expect exact output, verify it
			if tt.expected != "" {
				require.Equal(t, tt.expected, actual)
			}
		})
	}
}

// TestTokenizerWithDifferentDialects tests that tokenizer configurations work for different SQL dialects.
func TestTokenizerWithDifferentDialects(t *testing.T) {
	tests := []struct {
		name     string
		language Language
		input    string
	}{
		{name: "Standard SQL", language: StandardSQL, input: "SELECT * FROM users WHERE name = ?"},
		{name: "PostgreSQL", language: PostgreSQL, input: "SELECT * FROM users WHERE name = $1"},
		{name: "PL/SQL", language: PLSQL, input: "SELECT * FROM users WHERE name = :name"},
		{name: "DB2", language: DB2, input: "SELECT * FROM users WHERE name = ?"},
		{name: "N1QL", language: N1QL, input: "SELECT * FROM users WHERE name = $name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultConfig().WithLang(tt.language)
			actual := Format(tt.input, cfg)
			// Verify that formatting works without errors for each dialect
			require.NotEmpty(t, actual, "Format should return non-empty result for %s", tt.language)
			// Verify that the input is actually transformed (not just returned as-is)
			require.NotEqual(t, tt.input, actual, "Format should transform input for %s", tt.language)
		})
	}
}
