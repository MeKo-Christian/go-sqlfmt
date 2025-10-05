package sqlfmt

import (
	"testing"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDefaultConfig tests the default configuration creation.
func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	require.NotNil(t, config)
	assert.Equal(t, StandardSQL, config.Language)
	assert.Equal(t, DefaultIndent, config.Indent)
	assert.Equal(t, DefaultKeywordCase, config.KeywordCase)
	assert.Equal(t, DefaultLinesBetweenQueries, config.LinesBetweenQueries)
	assert.Equal(t, DefaultMaxLineLength, config.MaxLineLength)
	assert.False(t, config.AlignColumnNames)
	assert.False(t, config.AlignAssignments)
	assert.False(t, config.AlignValues)
	assert.False(t, config.PreserveCommentIndent)
	assert.Equal(t, 1, config.CommentMinSpacing)
	assert.NotNil(t, config.Params)
	assert.NotNil(t, config.ColorConfig)
	assert.NotNil(t, config.TokenizerConfig)
}

// TestConfigWithLang tests language configuration.
func TestConfigWithLang(t *testing.T) {
	tests := []struct {
		name     string
		language Language
	}{
		{"Standard SQL", StandardSQL},
		{"PostgreSQL", PostgreSQL},
		{"MySQL", MySQL},
		{"SQLite", SQLite},
		{"PL/SQL", PLSQL},
		{"DB2", DB2},
		{"N1QL", N1QL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithLang(tt.language)
			assert.Equal(t, tt.language, config.Language)
		})
	}
}

// TestConfigWithIndent tests indent configuration.
func TestConfigWithIndent(t *testing.T) {
	tests := []struct {
		name   string
		indent string
	}{
		{"two spaces", "  "},
		{"four spaces", "    "},
		{"tab", "\t"},
		{"three spaces", "   "},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithIndent(tt.indent)
			assert.Equal(t, tt.indent, config.Indent)
		})
	}
}

// TestConfigWithKeywordCase tests keyword case configuration.
func TestConfigWithKeywordCase(t *testing.T) {
	tests := []struct {
		name        string
		keywordCase KeywordCase
	}{
		{"preserve", KeywordCasePreserve},
		{"uppercase", KeywordCaseUppercase},
		{"lowercase", KeywordCaseLowercase},
		{"dialect", KeywordCaseDialect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithKeywordCase(tt.keywordCase)
			assert.Equal(t, tt.keywordCase, config.KeywordCase)
		})
	}
}

// TestConfigWithUppercase tests backward compatibility uppercase method.
func TestConfigWithUppercase(t *testing.T) {
	config := NewDefaultConfig().WithUppercase()
	assert.Equal(t, KeywordCaseUppercase, config.KeywordCase)
}

// TestConfigWithLinesBetweenQueries tests lines between queries configuration.
func TestConfigWithLinesBetweenQueries(t *testing.T) {
	tests := []struct {
		name  string
		lines int
	}{
		{"zero", 0},
		{"one", 1},
		{"two", 2},
		{"five", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithLinesBetweenQueries(tt.lines)
			assert.Equal(t, tt.lines, config.LinesBetweenQueries)
		})
	}
}

// TestConfigWithAlignColumnNames tests column alignment configuration.
func TestConfigWithAlignColumnNames(t *testing.T) {
	tests := []struct {
		name  string
		align bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithAlignColumnNames(tt.align)
			assert.Equal(t, tt.align, config.AlignColumnNames)
		})
	}
}

// TestConfigWithAlignAssignments tests assignment alignment configuration.
func TestConfigWithAlignAssignments(t *testing.T) {
	tests := []struct {
		name  string
		align bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithAlignAssignments(tt.align)
			assert.Equal(t, tt.align, config.AlignAssignments)
		})
	}
}

// TestConfigWithAlignValues tests value alignment configuration.
func TestConfigWithAlignValues(t *testing.T) {
	tests := []struct {
		name  string
		align bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithAlignValues(tt.align)
			assert.Equal(t, tt.align, config.AlignValues)
		})
	}
}

// TestConfigWithMaxLineLength tests max line length configuration.
func TestConfigWithMaxLineLength(t *testing.T) {
	tests := []struct {
		name      string
		maxLength int
	}{
		{"unlimited", 0},
		{"80 chars", 80},
		{"120 chars", 120},
		{"240 chars", 240},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithMaxLineLength(tt.maxLength)
			assert.Equal(t, tt.maxLength, config.MaxLineLength)
		})
	}
}

// TestConfigWithPreserveCommentIndent tests comment indent preservation configuration.
func TestConfigWithPreserveCommentIndent(t *testing.T) {
	tests := []struct {
		name     string
		preserve bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig().WithPreserveCommentIndent(tt.preserve)
			assert.Equal(t, tt.preserve, config.PreserveCommentIndent)
		})
	}
}

// TestConfigWithCommentMinSpacing tests comment minimum spacing configuration.
func TestConfigWithCommentMinSpacing(t *testing.T) {
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
			assert.Equal(t, tt.spacing, config.CommentMinSpacing)
		})
	}
}

// TestConfigWithParams tests params configuration.
func TestConfigWithParams(t *testing.T) {
	mapParams := NewMapParams(map[string]string{"key": "value"})
	config := NewDefaultConfig().WithParams(mapParams)
	assert.Equal(t, mapParams, config.Params)

	listParams := NewListParams([]string{"val1", "val2"})
	config = NewDefaultConfig().WithParams(listParams)
	assert.Equal(t, listParams, config.Params)
}

// TestConfigWithColorConfig tests color configuration.
func TestConfigWithColorConfig(t *testing.T) {
	colorConfig := &ColorConfig{
		ReservedWordFormatOptions: []utils.ANSIFormatOption{utils.ColorRed},
	}
	config := NewDefaultConfig().WithColorConfig(colorConfig)
	assert.Equal(t, colorConfig, config.ColorConfig)
}

// TestConfigWithTokenizerConfig tests tokenizer configuration.
func TestConfigWithTokenizerConfig(t *testing.T) {
	tokenizerConfig := &TokenizerConfig{
		ReservedWords: []string{"SELECT", "FROM"},
	}
	config := NewDefaultConfig().WithTokenizerConfig(tokenizerConfig)
	assert.Equal(t, tokenizerConfig, config.TokenizerConfig)
}

// TestConfigChaining tests that all With methods return the config for chaining.
func TestConfigChaining(t *testing.T) {
	config := NewDefaultConfig().
		WithLang(PostgreSQL).
		WithIndent("    ").
		WithKeywordCase(KeywordCaseUppercase).
		WithLinesBetweenQueries(3).
		WithAlignColumnNames(true).
		WithAlignAssignments(true).
		WithAlignValues(true).
		WithMaxLineLength(120).
		WithPreserveCommentIndent(true).
		WithCommentMinSpacing(2)

	assert.Equal(t, PostgreSQL, config.Language)
	assert.Equal(t, "    ", config.Indent)
	assert.Equal(t, KeywordCaseUppercase, config.KeywordCase)
	assert.Equal(t, 3, config.LinesBetweenQueries)
	assert.True(t, config.AlignColumnNames)
	assert.True(t, config.AlignAssignments)
	assert.True(t, config.AlignValues)
	assert.Equal(t, 120, config.MaxLineLength)
	assert.True(t, config.PreserveCommentIndent)
	assert.Equal(t, 2, config.CommentMinSpacing)
}

// TestConfigEmpty tests the Empty method.
func TestConfigEmpty(t *testing.T) {
	// Empty config
	emptyConfig := &Config{}
	assert.True(t, emptyConfig.Empty())

	// Non-empty config
	nonEmptyConfig := NewDefaultConfig()
	assert.False(t, nonEmptyConfig.Empty())

	// Partially filled config
	partialConfig := &Config{Language: PostgreSQL}
	assert.False(t, partialConfig.Empty())
}

// TestNewMapParams tests map params creation.
func TestNewMapParams(t *testing.T) {
	t.Run("with values", func(t *testing.T) {
		params := NewMapParams(map[string]string{"key1": "value1", "key2": "value2"})
		require.NotNil(t, params)
		assert.Equal(t, "value1", params.MapParams["key1"])
		assert.Equal(t, "value2", params.MapParams["key2"])
		assert.Nil(t, params.ListParams)
	})

	t.Run("with nil", func(t *testing.T) {
		params := NewMapParams(nil)
		require.NotNil(t, params)
		assert.NotNil(t, params.MapParams)
		assert.Empty(t, params.MapParams)
	})
}

// TestNewListParams tests list params creation.
func TestNewListParams(t *testing.T) {
	t.Run("with values", func(t *testing.T) {
		params := NewListParams([]string{"val1", "val2", "val3"})
		require.NotNil(t, params)
		assert.Equal(t, []string{"val1", "val2", "val3"}, params.ListParams)
		assert.Nil(t, params.MapParams)
	})

	t.Run("with nil", func(t *testing.T) {
		params := NewListParams(nil)
		require.NotNil(t, params)
		assert.NotNil(t, params.ListParams)
		assert.Empty(t, params.ListParams)
	})
}

// TestNewDefaultColorConfig tests default color configuration.
func TestNewDefaultColorConfig(t *testing.T) {
	colorConfig := NewDefaultColorConfig()

	require.NotNil(t, colorConfig)
	assert.NotEmpty(t, colorConfig.ReservedWordFormatOptions)
	assert.NotEmpty(t, colorConfig.StringFormatOptions)
	assert.NotEmpty(t, colorConfig.NumberFormatOptions)
	assert.NotEmpty(t, colorConfig.BooleanFormatOptions)
	assert.NotEmpty(t, colorConfig.CommentFormatOptions)
	assert.NotEmpty(t, colorConfig.FunctionCallFormatOptions)

	// Check specific defaults
	assert.Contains(t, colorConfig.ReservedWordFormatOptions, utils.ColorCyan)
	assert.Contains(t, colorConfig.ReservedWordFormatOptions, utils.FormatBold)
	assert.Contains(t, colorConfig.StringFormatOptions, utils.ColorGreen)
	assert.Contains(t, colorConfig.NumberFormatOptions, utils.ColorBrightBlue)
	assert.Contains(t, colorConfig.BooleanFormatOptions, utils.ColorPurple)
	assert.Contains(t, colorConfig.BooleanFormatOptions, utils.FormatBold)
	assert.Contains(t, colorConfig.CommentFormatOptions, utils.ColorGray)
	assert.Contains(t, colorConfig.FunctionCallFormatOptions, utils.ColorBrightCyan)
}

// TestColorConfigEmpty tests the ColorConfig Empty method.
func TestColorConfigEmpty(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		emptyConfig := &ColorConfig{}
		assert.True(t, emptyConfig.Empty())
	})

	t.Run("not empty", func(t *testing.T) {
		nonEmptyConfig := NewDefaultColorConfig()
		assert.False(t, nonEmptyConfig.Empty())
	})

	t.Run("partially filled", func(t *testing.T) {
		partialConfig := &ColorConfig{
			ReservedWordFormatOptions: []utils.ANSIFormatOption{utils.ColorRed},
		}
		assert.False(t, partialConfig.Empty())
	})
}

// TestLanguageConstants tests that all language constants are defined correctly.
func TestLanguageConstants(t *testing.T) {
	languages := []Language{
		StandardSQL,
		PostgreSQL,
		MySQL,
		SQLite,
		PLSQL,
		DB2,
		N1QL,
	}

	// Ensure no duplicates
	seen := make(map[Language]bool)
	for _, lang := range languages {
		assert.False(t, seen[lang], "duplicate language: %s", lang)
		seen[lang] = true
	}

	// Ensure all are non-empty
	for _, lang := range languages {
		assert.NotEmpty(t, string(lang), "language constant should not be empty")
	}
}

// TestKeywordCaseConstants tests that all keyword case constants are defined correctly.
func TestKeywordCaseConstants(t *testing.T) {
	cases := []KeywordCase{
		KeywordCasePreserve,
		KeywordCaseUppercase,
		KeywordCaseLowercase,
		KeywordCaseDialect,
	}

	// Ensure no duplicates
	seen := make(map[KeywordCase]bool)
	for _, kc := range cases {
		assert.False(t, seen[kc], "duplicate keyword case: %s", kc)
		seen[kc] = true
	}

	// Ensure all are non-empty
	for _, kc := range cases {
		assert.NotEmpty(t, string(kc), "keyword case constant should not be empty")
	}
}

// TestDefaultConstants tests default constant values.
func TestDefaultConstants(t *testing.T) {
	assert.Equal(t, "  ", DefaultIndent)
	assert.Equal(t, 2, DefaultLinesBetweenQueries)
	assert.Equal(t, 0, DefaultMaxLineLength)
	assert.Equal(t, KeywordCasePreserve, DefaultKeywordCase)
}

// TestConfigIndentAppliedToFormatting tests that indent configuration is applied to formatting.
func TestConfigIndentAppliedToFormatting(t *testing.T) {
	query := "SELECT * FROM users WHERE id = 1"

	// Test with two-space indent (default)
	result1 := Format(query, NewDefaultConfig().WithIndent("  "))
	assert.Contains(t, result1, "  ")

	// Test with four-space indent
	result2 := Format(query, NewDefaultConfig().WithIndent("    "))
	assert.Contains(t, result2, "    ")

	// Test with tab indent
	result3 := Format(query, NewDefaultConfig().WithIndent("\t"))
	assert.Contains(t, result3, "\t")
}

// TestConfigKeywordCaseAppliedToFormatting tests that keyword case is applied to formatting.
func TestConfigKeywordCaseAppliedToFormatting(t *testing.T) {
	query := "select * from users where id = 1"

	// Uppercase
	result1 := Format(query, NewDefaultConfig().WithKeywordCase(KeywordCaseUppercase))
	assert.Contains(t, result1, "SELECT")
	assert.Contains(t, result1, "FROM")
	assert.Contains(t, result1, "WHERE")

	// Lowercase
	result2 := Format(query, NewDefaultConfig().WithKeywordCase(KeywordCaseLowercase))
	assert.Contains(t, result2, "select")
	assert.Contains(t, result2, "from")
	assert.Contains(t, result2, "where")

	// Preserve (input is lowercase)
	result3 := Format(query, NewDefaultConfig().WithKeywordCase(KeywordCasePreserve))
	assert.Contains(t, result3, "select")
	assert.Contains(t, result3, "from")
	assert.Contains(t, result3, "where")
}

// TestConfigLanguageAppliedToFormatting tests that language selection affects formatting.
func TestConfigLanguageAppliedToFormatting(t *testing.T) {
	// PostgreSQL-specific: dollar-quoted strings
	pgQuery := "SELECT $$hello$$"
	pgResult := Format(pgQuery, NewDefaultConfig().WithLang(PostgreSQL))
	assert.Contains(t, pgResult, "$$hello$$")

	// Standard SQL should handle it differently (may treat as error or separate tokens)
	stdResult := Format(pgQuery, NewDefaultConfig().WithLang(StandardSQL))
	// Just ensure it doesn't crash - behavior may vary
	assert.NotEmpty(t, stdResult)
}

// TestConfigLinesBetweenQueriesAppliedToFormatting tests lines between queries configuration.
func TestConfigLinesBetweenQueriesAppliedToFormatting(t *testing.T) {
	query := "SELECT * FROM users; SELECT * FROM orders;"

	// Two lines between (default)
	result1 := Format(query, NewDefaultConfig().WithLinesBetweenQueries(2))
	assert.Contains(t, result1, ";\n\n")

	// Three lines between
	result2 := Format(query, NewDefaultConfig().WithLinesBetweenQueries(3))
	assert.Contains(t, result2, ";\n\n\n")

	// One line between
	result3 := Format(query, NewDefaultConfig().WithLinesBetweenQueries(1))
	assert.Contains(t, result3, ";\n")
}

// TestTokenizerConfigCustomReservedWords tests custom reserved words in tokenizer config.
func TestTokenizerConfigCustomReservedWords(t *testing.T) {
	config := NewDefaultConfig()
	config.TokenizerConfig = &TokenizerConfig{
		ReservedWords: []string{"CUSTOM_KEYWORD"},
	}

	// This test ensures the config can be set - actual tokenizer behavior
	// is tested in tokenizer tests
	assert.NotNil(t, config.TokenizerConfig)
	assert.Contains(t, config.TokenizerConfig.ReservedWords, "CUSTOM_KEYWORD")
}
