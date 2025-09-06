package core

import (
	"testing"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
	"github.com/stretchr/testify/require"
)

func TestTokenizerDirectRegexAccess(t *testing.T) {
	// Test direct regex functionality that was previously tested in the disabled test
	cfg := &TokenizerConfig{
		ReservedWords:                 []string{"SELECT", "FROM", "WHERE"},
		ReservedTopLevelWords:         []string{"SELECT", "FROM", "WHERE"},
		ReservedNewlineWords:          []string{"AND", "OR"},
		ReservedTopLevelWordsNoIndent: []string{"UNION ALL", "EXCEPT"},
		StringTypes:                   []string{"''", "\"\"", "``"},
		OpenParens:                    []string{"(", "["},
		CloseParens:                   []string{")", "]"},
		IndexedPlaceholderTypes:       []string{"?"},
		NamedPlaceholderTypes:         []string{"@", ":"},
		LineCommentTypes:              []string{"--", "#"},
		SpecialWordChars:              []string{},
	}

	tokenizer := newTokenizer(cfg)

	// Test word regex
	testWordRegex(t, tokenizer)
	// Test boolean regex
	testBooleanRegex(t, tokenizer)
	// Test function call regex
	testFunctionCallRegex(t, tokenizer)
}

func testWordRegex(t *testing.T, tokenizer *tokenizer) {
	t.Helper()
	tests := []struct {
		name  string
		input string
	}{
		{name: "word regex matches TEXT", input: "TEXT"},
		{name: "word regex on punctuation boundary", input: "TEXT);"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wordToken := tokenizer.getWordToken(tt.input)
			require.False(t, wordToken.Empty(), "Expected non-empty word token")
			require.Contains(t, wordToken.Value, "TEXT")
		})
	}
}

func testBooleanRegex(t *testing.T, tokenizer *tokenizer) {
	t.Helper()
	tests := []struct {
		name        string
		input       string
		expectEmpty bool
	}{
		{name: "boolean regex matches true", input: "true", expectEmpty: false},
		{name: "boolean regex matches false", input: "false", expectEmpty: false},
		{name: "boolean regex matches TRUE", input: "TRUE", expectEmpty: false},
		{name: "boolean regex case insensitive", input: "tRUE", expectEmpty: false},
		{name: "boolean regex word boundary", input: "trueDat", expectEmpty: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boolToken := tokenizer.getBooleanToken(tt.input)
			if tt.expectEmpty {
				require.True(t, boolToken.Empty(), "Expected empty boolean token")
			} else {
				require.False(t, boolToken.Empty(), "Expected non-empty boolean token")
			}
		})
	}
}

func testFunctionCallRegex(t *testing.T, tokenizer *tokenizer) {
	t.Helper()
	tests := []struct {
		name  string
		input string
	}{
		{name: "function call no args", input: "call()"},
		{name: "function call with args", input: "CALL_WITH_ARGS(arg1, 3+4, arg2);"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcToken := tokenizer.getTokenOnFirstMatch(tt.input, types.TokenTypeReserved, tokenizer.functionCallRegex)
			require.False(t, funcToken.Empty(), "Expected non-empty function token")
		})
	}
}

func TestTokenizerPlaceholders(t *testing.T) {
	cfg := &TokenizerConfig{
		IndexedPlaceholderTypes: []string{"?"},
		NamedPlaceholderTypes:   []string{"@", ":"},
		StringTypes:             []string{"''", "\"\"", "``", "[]"},
	}

	tokenizer := newTokenizer(cfg)

	tests := []struct {
		name     string
		input    string
		expected string
		key      string
	}{
		{name: "indexed placeholder ?", input: "?", expected: "?", key: ""},
		{name: "indexed placeholder ?1", input: "?1", expected: "?1", key: "1"},
		{name: "indexed placeholder ?22", input: "?22", expected: "?22", key: "22"},
		{name: "named placeholder @var", input: "@variable", expected: "@variable", key: "variable"},
		{name: "string named placeholder", input: "@'var name'", expected: "@'var name'", key: "var name"},
		{name: "bracket placeholder", input: "@[var name]", expected: "@[var name]", key: "var name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tokenizer.getPlaceholderToken(tt.input)
			require.False(t, token.Empty(), "Placeholder token should not be empty")
			require.Equal(t, tt.expected, token.Value)
			if tt.key != "" {
				require.Equal(t, tt.key, token.Key)
			}
		})
	}
}

func TestTokenizerReservedWords(t *testing.T) {
	cfg := &TokenizerConfig{
		ReservedTopLevelWords:         []string{"SELECT", "FROM"},
		ReservedTopLevelWordsNoIndent: []string{"UNION ALL"},
		ReservedNewlineWords:          []string{"AND", "OR"},
		ReservedWords:                 []string{"WHERE", "ORDER BY"},
	}

	tokenizer := newTokenizer(cfg)

	tests := []struct {
		name     string
		input    string
		expected types.TokenType
	}{
		{name: "top level SELECT", input: "SELECT", expected: types.TokenTypeReservedTopLevel},
		{name: "top level FROM", input: "FROM", expected: types.TokenTypeReservedTopLevel},
		{name: "no indent UNION ALL", input: "UNION ALL", expected: types.TokenTypeReservedTopLevelNoIndent},
		{name: "newline AND", input: "AND", expected: types.TokenTypeReservedNewline},
		{name: "newline OR", input: "OR", expected: types.TokenTypeReservedNewline},
		{name: "reserved WHERE", input: "WHERE", expected: types.TokenTypeReserved},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tokenizer.getReservedWordToken(tt.input, types.Token{})
			require.False(t, token.Empty(), "Reserved word token should not be empty")
			require.Equal(t, tt.expected, token.Type)
		})
	}
}

func TestTokenizerDotPrecedence(t *testing.T) {
	// Test that reserved words after dots are not treated as reserved
	cfg := &TokenizerConfig{
		ReservedTopLevelWords: []string{"FROM"},
	}

	tokenizer := newTokenizer(cfg)

	// "FROM" after a dot should not be treated as reserved
	prevToken := types.Token{Value: ".", Type: types.TokenTypeOperator}
	token := tokenizer.getReservedWordToken("FROM", prevToken)
	require.True(t, token.Empty(), "Reserved word after dot should be ignored, got: %+v", token)

	// Test that "FROM" is normally treated as reserved
	normalToken := tokenizer.getReservedWordToken("FROM", types.Token{})
	require.False(t, normalToken.Empty(), "FROM should normally be treated as reserved")
}
