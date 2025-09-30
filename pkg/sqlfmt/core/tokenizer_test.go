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

func TestTokenizerCompoundKeywords(t *testing.T) {
	// Test that compound keywords are matched before their component words
	cfg := &TokenizerConfig{
		ReservedWords: []string{
			"DO UPDATE", "DO NOTHING", "ON CONFLICT",
			"ORDER BY", "GROUP BY", "UNION ALL",
			"DO", "ON", "ORDER", "GROUP", "UNION",
		},
		ReservedTopLevelWords:         []string{"SELECT", "UPDATE"},
		ReservedNewlineWords:          []string{"AND", "OR"},
		ReservedTopLevelWordsNoIndent: []string{},
		StringTypes:                   []string{"''"},
		OpenParens:                    []string{"("},
		CloseParens:                   []string{")"},
		LineCommentTypes:              []string{"--"},
		SpecialWordChars:              []string{},
	}

	tokenizer := newTokenizer(cfg)

	tests := []struct {
		name          string
		input         string
		expectedValue string
		expectedType  types.TokenType
		description   string
	}{
		{
			name:          "DO UPDATE compound",
			input:         "DO UPDATE SET name = 'test'",
			expectedValue: "DO UPDATE",
			expectedType:  types.TokenTypeReserved,
			description:   "Should match 'DO UPDATE' as single token, not 'DO' alone",
		},
		{
			name:          "DO NOTHING compound",
			input:         "DO NOTHING;",
			expectedValue: "DO NOTHING",
			expectedType:  types.TokenTypeReserved,
			description:   "Should match 'DO NOTHING' as single token, not 'DO' alone",
		},
		{
			name:          "ON CONFLICT compound",
			input:         "ON CONFLICT (id)",
			expectedValue: "ON CONFLICT",
			expectedType:  types.TokenTypeReserved,
			description:   "Should match 'ON CONFLICT' as single token, not 'ON' alone",
		},
		{
			name:          "ORDER BY compound",
			input:         "ORDER BY name",
			expectedValue: "ORDER BY",
			expectedType:  types.TokenTypeReserved,
			description:   "Should match 'ORDER BY' as single token",
		},
		{
			name:          "GROUP BY compound",
			input:         "GROUP BY id",
			expectedValue: "GROUP BY",
			expectedType:  types.TokenTypeReserved,
			description:   "Should match 'GROUP BY' as single token",
		},
		{
			name:          "UNION ALL compound",
			input:         "UNION ALL SELECT",
			expectedValue: "UNION ALL",
			expectedType:  types.TokenTypeReserved,
			description:   "Should match 'UNION ALL' as single token",
		},
		{
			name:          "DO UPDATE with extra spaces",
			input:         "DO  UPDATE SET",
			expectedValue: "DO  UPDATE",
			expectedType:  types.TokenTypeReserved,
			description:   "Should match 'DO UPDATE' even with multiple spaces",
		},
		{
			name:          "DO UPDATE with tab",
			input:         "DO\tUPDATE SET",
			expectedValue: "DO\tUPDATE",
			expectedType:  types.TokenTypeReserved,
			description:   "Should match 'DO UPDATE' even with tab character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tokenizer.getReservedWordToken(tt.input, types.Token{})
			require.False(t, token.Empty(), "Token should not be empty for: %s", tt.description)
			require.Equal(t, tt.expectedType, token.Type, "Token type mismatch for: %s", tt.description)
			require.Equal(t, tt.expectedValue, token.Value, "Token value mismatch for: %s\nExpected: %q\nGot: %q",
				tt.description, tt.expectedValue, token.Value)
		})
	}
}

func TestTokenizerCompoundKeywordPriority(t *testing.T) {
	// Test that longer compound keywords are matched before shorter ones
	cfg := &TokenizerConfig{
		ReservedWords: []string{
			"DO UPDATE", "DO NOTHING", "DO",
			"LEFT OUTER JOIN", "LEFT JOIN", "LEFT",
			"ORDER BY", "ORDER",
		},
		ReservedTopLevelWords:         []string{},
		ReservedNewlineWords:          []string{},
		ReservedTopLevelWordsNoIndent: []string{},
		StringTypes:                   []string{"''"},
		OpenParens:                    []string{"("},
		CloseParens:                   []string{")"},
		LineCommentTypes:              []string{"--"},
		SpecialWordChars:              []string{},
	}

	tokenizer := newTokenizer(cfg)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Longest match: LEFT OUTER JOIN",
			input:    "LEFT OUTER JOIN",
			expected: "LEFT OUTER JOIN",
		},
		{
			name:     "Medium match: LEFT JOIN",
			input:    "LEFT JOIN",
			expected: "LEFT JOIN",
		},
		{
			name:     "Short match: LEFT alone",
			input:    "LEFT LATERAL",
			expected: "LEFT",
		},
		{
			name:     "DO UPDATE before DO",
			input:    "DO UPDATE",
			expected: "DO UPDATE",
		},
		{
			name:     "DO NOTHING before DO",
			input:    "DO NOTHING",
			expected: "DO NOTHING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tokenizer.getReservedWordToken(tt.input, types.Token{})
			require.False(t, token.Empty(), "Expected token for: %s", tt.input)
			require.Equal(t, tt.expected, token.Value, "Expected longest match. Input: %q", tt.input)
		})
	}
}

func checkStandaloneDoToken(t *testing.T, tokens []types.Token, doIndex int) {
	t.Helper()
	// Check if next non-whitespace token is UPDATE
	for j := doIndex + 1; j < len(tokens); j++ {
		if tokens[j].Type == types.TokenTypeWhitespace {
			continue
		}
		if tokens[j].Value == "UPDATE" {
			t.Errorf("Found standalone 'DO' token at index %d followed by 'UPDATE' at %d. "+
				"Should be 'DO UPDATE' compound keyword.", doIndex, j)
		}
		break
	}
}

func TestTokenizerFullUpsertQuery(t *testing.T) {
	// Test full UPSERT query tokenization
	cfg := &TokenizerConfig{
		ReservedWords: []string{
			"DO UPDATE", "DO NOTHING", "ON CONFLICT", "DO",
			"SET", "WHERE", "VALUES",
		},
		ReservedTopLevelWords:         []string{"INSERT INTO", "INSERT", "UPDATE"},
		ReservedNewlineWords:          []string{"AND", "OR"},
		ReservedTopLevelWordsNoIndent: []string{},
		StringTypes:                   []string{"''"},
		OpenParens:                    []string{"("},
		CloseParens:                   []string{")"},
		IndexedPlaceholderTypes:       []string{"?"},
		LineCommentTypes:              []string{"--"},
		SpecialWordChars:              []string{},
	}

	tokenizer := newTokenizer(cfg)

	query := "INSERT INTO users (id, name) VALUES (1, 'John') ON CONFLICT (id) DO UPDATE SET name = 'Jane'"
	tokens := tokenizer.tokenize(query)

	// Find the relevant tokens
	var foundOnConflict, foundDoUpdate bool
	for i, tok := range tokens {
		if tok.Type != types.TokenTypeReserved {
			continue
		}

		if tok.Value == "ON CONFLICT" {
			foundOnConflict = true
			t.Logf("Found ON CONFLICT at index %d", i)
		}
		if tok.Value == "DO UPDATE" {
			foundDoUpdate = true
			t.Logf("Found DO UPDATE at index %d", i)
		}

		// Make sure we don't have standalone "DO" before "UPDATE"
		if tok.Value == "DO" {
			checkStandaloneDoToken(t, tokens, i)
		}
	}

	require.True(t, foundOnConflict, "Should find 'ON CONFLICT' compound keyword in tokens")
	require.True(t, foundDoUpdate, "Should find 'DO UPDATE' compound keyword in tokens")
}
