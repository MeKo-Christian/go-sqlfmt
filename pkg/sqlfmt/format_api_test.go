package sqlfmt

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatWithNilConfig tests Format with nil config (should use defaults).
func TestFormatWithNilConfig(t *testing.T) {
	query := "SELECT * FROM users WHERE id = 1"
	result := Format(query)

	assert.Contains(t, result, "SELECT")
	assert.Contains(t, result, "FROM")
	assert.Contains(t, result, "WHERE")
	assert.NotEmpty(t, result)
}

// TestFormatWithEmptyConfig tests Format with an empty config.
func TestFormatWithEmptyConfig(t *testing.T) {
	query := "SELECT * FROM users WHERE id = 1"
	config := &Config{} // Empty config, should use zero values
	result := Format(query, config)

	assert.NotEmpty(t, result)
	// Empty config will have empty indent, but should still format
	assert.Contains(t, result, "SELECT")
}

// TestFormatWithDefaultConfig tests Format with default config.
func TestFormatWithDefaultConfig(t *testing.T) {
	query := "SELECT * FROM users WHERE id = 1"
	config := NewDefaultConfig()
	result := Format(query, config)

	assert.Contains(t, result, "SELECT")
	assert.Contains(t, result, "FROM")
	assert.Contains(t, result, "users")
	// Should use default indent (two spaces)
	assert.Contains(t, result, "  ")
}

// TestFormatWithMultipleConfigs tests that multiple configs causes panic.
func TestFormatWithMultipleConfigs(t *testing.T) {
	query := "SELECT * FROM users"
	config1 := NewDefaultConfig()
	config2 := NewDefaultConfig()

	assert.Panics(t, func() {
		Format(query, config1, config2)
	}, "should panic with multiple configs")
}

// TestFormatWithAllDialects tests Format with all supported dialects.
func TestFormatWithAllDialects(t *testing.T) {
	tests := []struct {
		name     string
		language Language
		query    string
	}{
		{"StandardSQL", StandardSQL, "SELECT * FROM users"},
		{"PostgreSQL", PostgreSQL, "SELECT * FROM users"},
		{"MySQL", MySQL, "SELECT * FROM users"},
		{"SQLite", SQLite, "SELECT * FROM users"},
		{"PLSQL", PLSQL, "SELECT * FROM users"},
		{"DB2", DB2, "SELECT * FROM users"},
		{"N1QL", N1QL, "SELECT * FROM users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithLang(tt.language)
			result := Format(tt.query, config)

			assert.NotEmpty(t, result)
			assert.Contains(t, result, "SELECT")
			assert.Contains(t, result, "users")
		})
	}
}

// TestFormatWithAllKeywordCases tests Format with different keyword cases.
func TestFormatWithAllKeywordCases(t *testing.T) {
	tests := []struct {
		name        string
		keywordCase KeywordCase
		query       string
		expectUpper bool
		expectLower bool
	}{
		{
			name:        "uppercase",
			keywordCase: KeywordCaseUppercase,
			query:       "select * from users where id = 1",
			expectUpper: true,
			expectLower: false,
		},
		{
			name:        "lowercase",
			keywordCase: KeywordCaseLowercase,
			query:       "SELECT * FROM USERS WHERE ID = 1",
			expectUpper: false,
			expectLower: true,
		},
		{
			name:        "preserve uppercase",
			keywordCase: KeywordCasePreserve,
			query:       "SELECT * FROM USERS",
			expectUpper: true,
			expectLower: false,
		},
		{
			name:        "preserve lowercase",
			keywordCase: KeywordCasePreserve,
			query:       "select * from users",
			expectUpper: false,
			expectLower: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithKeywordCase(tt.keywordCase)
			result := Format(tt.query, config)

			if tt.expectUpper {
				assert.Contains(t, result, "SELECT", "should contain uppercase SELECT")
				assert.Contains(t, result, "FROM", "should contain uppercase FROM")
			}
			if tt.expectLower {
				assert.Contains(t, result, "select", "should contain lowercase select")
				assert.Contains(t, result, "from", "should contain lowercase from")
			}
		})
	}
}

// TestFormatWithIndentOptions tests Format with different indent settings.
func TestFormatWithIndentOptions(t *testing.T) {
	query := "SELECT * FROM users WHERE id = 1"

	tests := []struct {
		name   string
		indent string
	}{
		{"two spaces", "  "},
		{"four spaces", "    "},
		{"tab", "\t"},
		{"single space", " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithIndent(tt.indent)
			result := Format(query, config)

			assert.Contains(t, result, tt.indent, "should use specified indent")
		})
	}
}

// TestFormatWithAlignmentOptions tests Format with various alignment settings.
func TestFormatWithAlignmentOptions(t *testing.T) {
	tests := []struct {
		name             string
		query            string
		alignColumns     bool
		alignAssignments bool
		alignValues      bool
	}{
		{
			name:         "align column names",
			query:        "SELECT id, name, email FROM users",
			alignColumns: true,
		},
		{
			name:             "align assignments",
			query:            "UPDATE users SET name = 'John', email = 'john@example.com' WHERE id = 1",
			alignAssignments: true,
		},
		{
			name:        "align values",
			query:       "INSERT INTO users (id, name) VALUES (1, 'John'), (2, 'Jane')",
			alignValues: true,
		},
		{
			name:             "all alignment options",
			query:            "UPDATE users SET name = 'John', email = 'john@example.com' WHERE id = 1",
			alignColumns:     true,
			alignAssignments: true,
			alignValues:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().
				WithAlignColumnNames(tt.alignColumns).
				WithAlignAssignments(tt.alignAssignments).
				WithAlignValues(tt.alignValues)

			result := Format(tt.query, config)
			assert.NotEmpty(t, result, "should produce output")
		})
	}
}

// TestFormatWithParams tests Format with parameter replacement.
func TestFormatWithParams(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		params *Params
		expect string
	}{
		{
			name:   "map params",
			query:  "SELECT * FROM users WHERE id = :id AND name = :name",
			params: NewMapParams(map[string]string{"id": "123", "name": "'John'"}),
			expect: "123",
		},
		{
			name:   "list params",
			query:  "SELECT * FROM users WHERE id = ? AND name = ?",
			params: NewListParams([]string{"123", "'John'"}),
			expect: "123",
		},
		{
			name:   "nil params",
			query:  "SELECT * FROM users WHERE id = ?",
			params: nil,
			expect: "?", // Should preserve placeholder
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithParams(tt.params)
			result := Format(tt.query, config)

			assert.Contains(t, result, tt.expect)
		})
	}
}

// TestPrettyFormatWithNilConfig tests PrettyFormat with nil config.
func TestPrettyFormatWithNilConfig(t *testing.T) {
	query := "SELECT * FROM users WHERE id = 1"
	result := PrettyFormat(query)

	// Should have ANSI color codes
	assert.Contains(t, result, "\033[", "should contain ANSI codes")
	assert.Contains(t, result, "SELECT")
}

// TestPrettyFormatWithDefaultColors tests PrettyFormat with default color config.
func TestPrettyFormatWithDefaultColors(t *testing.T) {
	query := "SELECT 'hello', 123, true FROM users"
	result := PrettyFormat(query)

	// Should contain default colors
	assert.Contains(t, result, "\033[", "should contain ANSI codes")
	assert.NotEmpty(t, result)
}

// TestPrettyFormatWithCustomColors tests PrettyFormat with custom color configuration.
func TestPrettyFormatWithCustomColors(t *testing.T) {
	colorConfig := &ColorConfig{
		ReservedWordFormatOptions: []utils.ANSIFormatOption{utils.ColorRed},
		StringFormatOptions:       []utils.ANSIFormatOption{utils.ColorBlue},
		NumberFormatOptions:       []utils.ANSIFormatOption{utils.ColorGreen},
		BooleanFormatOptions:      []utils.ANSIFormatOption{utils.ColorPurple},
		CommentFormatOptions:      []utils.ANSIFormatOption{utils.ColorGray},
		FunctionCallFormatOptions: []utils.ANSIFormatOption{utils.ColorCyan},
	}

	config := NewDefaultConfig().WithColorConfig(colorConfig)
	query := "SELECT 'test', 123, true FROM users -- comment"
	result := PrettyFormat(query, config)

	// Should have color codes
	assert.Contains(t, result, "\033[", "should contain ANSI codes")
	assert.NotEmpty(t, result)
}

// TestPrettyFormatWithEmptyColorConfig tests that empty color config gets default colors.
func TestPrettyFormatWithEmptyColorConfig(t *testing.T) {
	config := NewDefaultConfig().WithColorConfig(&ColorConfig{})
	query := "SELECT * FROM users"
	result := PrettyFormat(query, config)

	// Should still have colors (defaults applied)
	assert.Contains(t, result, "\033[", "should contain ANSI codes from defaults")
}

// TestPrettyFormatMultipleConfigsPanic tests that PrettyFormat panics with multiple configs.
func TestPrettyFormatMultipleConfigsPanic(t *testing.T) {
	query := "SELECT * FROM users"
	config1 := NewDefaultConfig()
	config2 := NewDefaultConfig()

	assert.Panics(t, func() {
		PrettyFormat(query, config1, config2)
	}, "should panic with multiple configs")
}

// TestPrettyPrintOutput tests PrettyPrint actually prints output.
func TestPrettyPrintOutput(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	query := "SELECT * FROM users"
	PrettyPrint(query)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "SELECT")
	assert.Contains(t, output, "\033[", "should contain ANSI codes")
}

// TestPrettyPrintWithConfig tests PrettyPrint with custom config.
func TestPrettyPrintWithConfig(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	query := "SELECT * FROM users"
	config := NewDefaultConfig().WithKeywordCase(KeywordCaseUppercase)
	PrettyPrint(query, config)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "SELECT")
	assert.Contains(t, output, "FROM")
}

// TestPrettyPrintMultipleConfigsPanic tests that PrettyPrint panics with multiple configs.
func TestPrettyPrintMultipleConfigsPanic(t *testing.T) {
	query := "SELECT * FROM users"
	config1 := NewDefaultConfig()
	config2 := NewDefaultConfig()

	assert.Panics(t, func() {
		PrettyPrint(query, config1, config2)
	}, "should panic with multiple configs")
}

// TestFormatComplexConfigCombinations tests Format with complex config combinations.
func TestFormatComplexConfigCombinations(t *testing.T) {
	query := "SELECT id, name FROM users WHERE id = 1 AND status = 'active'"

	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "all features enabled",
			config: NewDefaultConfig().
				WithLang(PostgreSQL).
				WithIndent("    ").
				WithKeywordCase(KeywordCaseUppercase).
				WithLinesBetweenQueries(3).
				WithAlignColumnNames(true).
				WithAlignAssignments(true).
				WithAlignValues(true).
				WithMaxLineLength(80),
		},
		{
			name: "minimal config",
			config: NewDefaultConfig().
				WithLang(StandardSQL).
				WithIndent("  ").
				WithKeywordCase(KeywordCasePreserve),
		},
		{
			name: "custom colors and formatting",
			config: NewDefaultConfig().
				WithLang(MySQL).
				WithKeywordCase(KeywordCaseLowercase).
				WithColorConfig(NewDefaultColorConfig()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Format(query, tt.config)
			assert.NotEmpty(t, result, "should produce formatted output")
			assert.NotEqual(t, query, result, "should format the query")
		})
	}
}

// TestFormatWithLinesBetweenQueries tests Format with different line spacing.
func TestFormatWithLinesBetweenQueries(t *testing.T) {
	query := "SELECT * FROM users; SELECT * FROM orders;"

	tests := []struct {
		name         string
		linesBetween int
		expectLines  string
	}{
		{"one line", 1, ";\n"},
		{"two lines", 2, ";\n\n"},
		{"three lines", 3, ";\n\n\n"},
		{"zero lines", 0, ";"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithLinesBetweenQueries(tt.linesBetween)
			result := Format(query, config)

			if tt.linesBetween > 0 {
				assert.Contains(t, result, tt.expectLines)
			}
		})
	}
}

// TestGetFormatterWithColor tests getFormatter color handling.
func TestGetFormatterWithColor(t *testing.T) {
	query := "SELECT * FROM users"

	t.Run("forceWithColor creates default colors", func(t *testing.T) {
		// When forceWithColor is true and no color config, should add defaults
		formatter := getFormatter(true)
		result := formatter.Format(query)
		assert.Contains(t, result, "\033[", "should have color codes")
	})

	t.Run("forceWithColor respects existing colors", func(t *testing.T) {
		config := NewDefaultConfig().WithColorConfig(&ColorConfig{
			ReservedWordFormatOptions: []utils.ANSIFormatOption{utils.ColorRed},
		})
		formatter := getFormatter(true, config)
		result := formatter.Format(query)
		assert.Contains(t, result, "\033[", "should have color codes")
	})

	t.Run("no force color with empty config", func(t *testing.T) {
		config := NewDefaultConfig()
		formatter := getFormatter(false, config)
		result := formatter.Format(query)
		// Should not force colors when forceWithColor is false
		assert.NotEmpty(t, result)
	})
}

// TestConvertParams tests param conversion for different dialects.
func TestConvertParams(t *testing.T) {
	t.Run("nil params", func(t *testing.T) {
		result := convertParams(nil, StandardSQL)
		assert.Nil(t, result)
	})

	t.Run("map params", func(t *testing.T) {
		params := NewMapParams(map[string]string{"key": "value"})
		result := convertParams(params, StandardSQL)
		require.NotNil(t, result)
		assert.Equal(t, "value", result.MapParams["key"])
	})

	t.Run("list params", func(t *testing.T) {
		params := NewListParams([]string{"val1", "val2"})
		result := convertParams(params, StandardSQL)
		require.NotNil(t, result)
		assert.Equal(t, []string{"val1", "val2"}, result.ListParams)
	})

	t.Run("SQLite uses 1-based indexing", func(t *testing.T) {
		params := NewListParams([]string{"val"})
		result := convertParams(params, SQLite)
		require.NotNil(t, result)
		assert.True(t, result.UseSQLiteIndexing)
	})

	t.Run("other dialects use 0-based indexing", func(t *testing.T) {
		params := NewListParams([]string{"val"})
		result := convertParams(params, PostgreSQL)
		require.NotNil(t, result)
		assert.False(t, result.UseSQLiteIndexing)
	})
}

// TestConvertColorConfig tests color config conversion.
func TestConvertColorConfig(t *testing.T) {
	t.Run("nil color config", func(t *testing.T) {
		result := convertColorConfig(nil)
		assert.Nil(t, result)
	})

	t.Run("empty color config", func(t *testing.T) {
		cc := &ColorConfig{}
		result := convertColorConfig(cc)
		require.NotNil(t, result)
		assert.Empty(t, result.ReservedWordFormatOptions)
	})

	t.Run("full color config", func(t *testing.T) {
		cc := NewDefaultColorConfig()
		result := convertColorConfig(cc)
		require.NotNil(t, result)
		assert.NotEmpty(t, result.ReservedWordFormatOptions)
		assert.NotEmpty(t, result.StringFormatOptions)
		assert.NotEmpty(t, result.NumberFormatOptions)
	})
}

// TestConvertTokenizerConfig tests tokenizer config conversion.
func TestConvertTokenizerConfig(t *testing.T) {
	t.Run("nil tokenizer config", func(t *testing.T) {
		result := convertTokenizerConfig(nil)
		assert.Nil(t, result)
	})

	t.Run("empty tokenizer config", func(t *testing.T) {
		tc := &TokenizerConfig{}
		result := convertTokenizerConfig(tc)
		require.NotNil(t, result)
		assert.Empty(t, result.ReservedWords)
	})

	t.Run("full tokenizer config", func(t *testing.T) {
		tc := &TokenizerConfig{
			ReservedWords:         []string{"SELECT", "FROM"},
			ReservedTopLevelWords: []string{"SELECT"},
			StringTypes:           []string{"'"},
		}
		result := convertTokenizerConfig(tc)
		require.NotNil(t, result)
		assert.Equal(t, []string{"SELECT", "FROM"}, result.ReservedWords)
		assert.Equal(t, []string{"SELECT"}, result.ReservedTopLevelWords)
		assert.Equal(t, []string{"'"}, result.StringTypes)
	})
}

// TestFormatEmptyQuery tests formatting an empty query.
func TestFormatEmptyQuery(t *testing.T) {
	result := Format("")
	assert.Empty(t, result, "empty query should return empty string")
}

// TestFormatWhitespaceOnly tests formatting a whitespace-only query.
func TestFormatWhitespaceOnly(t *testing.T) {
	tests := []string{
		"   ",
		"\t\t",
		"\n\n",
		"  \t\n  ",
	}

	for _, query := range tests {
		t.Run("whitespace", func(t *testing.T) {
			result := Format(query)
			// Should handle whitespace-only input gracefully
			assert.True(t, len(strings.TrimSpace(result)) == 0 || result == "")
		})
	}
}

// TestFormatVeryLongQuery tests formatting a very long query.
func TestFormatVeryLongQuery(t *testing.T) {
	// Build a query with many columns
	var columns []string
	for i := 0; i < 100; i++ {
		columns = append(columns, "column"+string(rune('0'+i%10)))
	}
	query := "SELECT " + strings.Join(columns, ", ") + " FROM users"

	result := Format(query)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "SELECT")
	assert.Contains(t, result, "FROM")
}

// TestFormatWithCommentMinSpacing tests comment spacing configuration.
func TestFormatWithCommentMinSpacing(t *testing.T) {
	query := "SELECT * FROM users -- comment"

	tests := []struct {
		name    string
		spacing int
	}{
		{"one space", 1},
		{"two spaces", 2},
		{"four spaces", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithCommentMinSpacing(tt.spacing)
			result := Format(query, config)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "comment")
		})
	}
}

// TestFormatWithPreserveCommentIndent tests comment indent preservation.
func TestFormatWithPreserveCommentIndent(t *testing.T) {
	query := `SELECT * FROM users
    -- indented comment
WHERE id = 1`

	t.Run("preserve enabled", func(t *testing.T) {
		config := NewDefaultConfig().WithPreserveCommentIndent(true)
		result := Format(query, config)
		assert.Contains(t, result, "comment")
	})

	t.Run("preserve disabled", func(t *testing.T) {
		config := NewDefaultConfig().WithPreserveCommentIndent(false)
		result := Format(query, config)
		assert.Contains(t, result, "comment")
	})
}

// TestNewTokenizerConfigs tests all tokenizer config creators.
func TestNewTokenizerConfigs(t *testing.T) {
	t.Run("NewStandardSQLTokenizerConfig", func(t *testing.T) {
		config := NewStandardSQLTokenizerConfig()
		require.NotNil(t, config)
		assert.NotEmpty(t, config.ReservedWords)
	})

	t.Run("NewPostgreSQLTokenizerConfig", func(t *testing.T) {
		config := NewPostgreSQLTokenizerConfig()
		require.NotNil(t, config)
		assert.NotEmpty(t, config.ReservedWords)
	})

	t.Run("NewMySQLTokenizerConfig", func(t *testing.T) {
		config := NewMySQLTokenizerConfig()
		require.NotNil(t, config)
		assert.NotEmpty(t, config.ReservedWords)
	})

	t.Run("NewSQLiteTokenizerConfig", func(t *testing.T) {
		config := NewSQLiteTokenizerConfig()
		require.NotNil(t, config)
		assert.NotEmpty(t, config.ReservedWords)
	})
}

// TestNewDialectFormatters tests all dialect formatter creators.
func TestNewDialectFormatters(t *testing.T) {
	config := NewDefaultConfig()
	query := "SELECT * FROM users"

	tests := []struct {
		name      string
		formatter Formatter
	}{
		{"StandardSQL", NewStandardSQLFormatter(config)},
		{"DB2", NewDB2Formatter(config)},
		{"PostgreSQL", NewPostgreSQLFormatter(config)},
		{"PLSQL", NewPLSQLFormatter(config)},
		{"N1QL", NewN1QLFormatter(config)},
		{"MySQL", NewMySQLFormatter(config)},
		{"SQLite", NewSQLiteFormatter(config)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.formatter)
			result := tt.formatter.Format(query)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "SELECT")
		})
	}
}

// TestConvertToInternalConfig tests the internal config conversion.
func TestConvertToInternalConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		result := convertToInternalConfig(nil)
		require.NotNil(t, result)
		assert.NotEmpty(t, result.Indent)
	})

	t.Run("full config conversion", func(t *testing.T) {
		config := NewDefaultConfig().
			WithLang(PostgreSQL).
			WithIndent("    ").
			WithKeywordCase(KeywordCaseUppercase).
			WithLinesBetweenQueries(3).
			WithParams(NewMapParams(map[string]string{"key": "value"})).
			WithColorConfig(NewDefaultColorConfig()).
			WithTokenizerConfig(&TokenizerConfig{ReservedWords: []string{"SELECT"}}).
			WithAlignColumnNames(true).
			WithAlignAssignments(true).
			WithAlignValues(true).
			WithMaxLineLength(80).
			WithPreserveCommentIndent(true).
			WithCommentMinSpacing(2)

		result := convertToInternalConfig(config)
		require.NotNil(t, result)
		assert.NotEmpty(t, result.Indent)
		assert.True(t, result.AlignColumnNames)
		assert.True(t, result.AlignAssignments)
		assert.True(t, result.AlignValues)
		assert.True(t, result.PreserveCommentIndent)
		assert.Equal(t, 2, result.CommentMinSpacing)
		assert.Equal(t, 80, result.MaxLineLength)
	})
}

// TestDedentFunction tests the Dedent utility function.
func TestDedentFunction(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "removes common indent",
			input: `
				SELECT *
				FROM users
				WHERE id = 1`,
			expect: "SELECT *\nFROM users\nWHERE id = 1",
		},
		{
			name:   "empty string",
			input:  "",
			expect: "",
		},
		{
			name:   "no indent",
			input:  "SELECT * FROM users",
			expect: "SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Dedent(tt.input)
			// Normalize whitespace for comparison
			result = strings.TrimSpace(result)
			expect := strings.TrimSpace(tt.expect)
			assert.Equal(t, expect, result)
		})
	}
}

// TestColorConstants tests that color constants are exported correctly.
func TestColorConstants(t *testing.T) {
	// Just verify the constants exist and have expected values
	assert.NotEmpty(t, FormatReset)
	assert.NotEmpty(t, FormatBold)
	assert.NotEmpty(t, ColorRed)
	assert.NotEmpty(t, ColorGreen)
	assert.NotEmpty(t, ColorBlue)
	assert.NotEmpty(t, ColorCyan)
	assert.NotEmpty(t, ColorPurple)
	assert.NotEmpty(t, ColorGray)
	assert.NotEmpty(t, ColorBrightBlue)
	assert.NotEmpty(t, ColorBrightCyan)
}
