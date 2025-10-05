package core

import (
	"strings"
	"testing"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
	"github.com/stretchr/testify/require"
)

const (
	testUpdateKeyword = "UPDATE"
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
		ReservedTopLevelWords:         []string{"SELECT", testUpdateKeyword},
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
		if tokens[j].Value == testUpdateKeyword {
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
		ReservedTopLevelWords:         []string{"INSERT INTO", "INSERT", testUpdateKeyword},
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

		// Make sure we don't have standalone "DO" before testUpdateKeyword
		if tok.Value == "DO" {
			checkStandaloneDoToken(t, tokens, i)
		}
	}

	require.True(t, foundOnConflict, "Should find 'ON CONFLICT' compound keyword in tokens")
	require.True(t, foundDoUpdate, "Should find 'DO UPDATE' compound keyword in tokens")
}

// ============================================================================
// Comment Tests
// ============================================================================

func TestTokenizerLineComments(t *testing.T) {
	tests := []struct {
		name            string
		lineCommentType []string
		input           string
		expectedValue   string
		expectedType    types.TokenType
	}{
		{
			name:            "double dash comment",
			lineCommentType: []string{"--"},
			input:           "-- This is a comment\nSELECT",
			expectedValue:   "-- This is a comment\n",
			expectedType:    types.TokenTypeLineComment,
		},
		{
			name:            "hash comment",
			lineCommentType: []string{"#"},
			input:           "# This is a comment\nSELECT",
			expectedValue:   "# This is a comment\n",
			expectedType:    types.TokenTypeLineComment,
		},
		{
			name:            "comment at end of string",
			lineCommentType: []string{"--"},
			input:           "-- Comment at end",
			expectedValue:   "-- Comment at end",
			expectedType:    types.TokenTypeLineComment,
		},
		{
			name:            "comment with CRLF",
			lineCommentType: []string{"--"},
			input:           "-- Comment\r\nSELECT",
			expectedValue:   "-- Comment\r\n",
			expectedType:    types.TokenTypeLineComment,
		},
		{
			name:            "comment with CR only",
			lineCommentType: []string{"--"},
			input:           "-- Comment\rSELECT",
			expectedValue:   "-- Comment\r",
			expectedType:    types.TokenTypeLineComment,
		},
		{
			name:            "empty comment",
			lineCommentType: []string{"--"},
			input:           "--\nSELECT",
			expectedValue:   "--\n",
			expectedType:    types.TokenTypeLineComment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TokenizerConfig{
				LineCommentTypes: tt.lineCommentType,
			}
			tokenizer := newTokenizer(cfg)
			token := tokenizer.getLineCommentToken(tt.input)
			require.False(t, token.Empty(), "Line comment token should not be empty")
			require.Equal(t, tt.expectedType, token.Type)
			require.Equal(t, tt.expectedValue, token.Value)
		})
	}
}

func TestTokenizerBlockComments(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValue string
		expectedType  types.TokenType
	}{
		{
			name:          "simple block comment",
			input:         "/* comment */ SELECT",
			expectedValue: "/* comment */",
			expectedType:  types.TokenTypeBlockComment,
		},
		{
			name:          "multi-line block comment",
			input:         "/* line 1\nline 2\nline 3 */ SELECT",
			expectedValue: "/* line 1\nline 2\nline 3 */",
			expectedType:  types.TokenTypeBlockComment,
		},
		{
			name:          "nested content",
			input:         "/* outer /* inner */ still in comment */ SELECT",
			expectedValue: "/* outer /* inner */",
			expectedType:  types.TokenTypeBlockComment,
		},
		{
			name:          "unclosed block comment",
			input:         "/* unclosed comment SELECT",
			expectedValue: "/* unclosed comment SELECT",
			expectedType:  types.TokenTypeBlockComment,
		},
		{
			name:          "empty block comment",
			input:         "/**/ SELECT",
			expectedValue: "/**/",
			expectedType:  types.TokenTypeBlockComment,
		},
		{
			name:          "block comment with special chars",
			input:         "/* @#$%^&*() */ SELECT",
			expectedValue: "/* @#$%^&*() */",
			expectedType:  types.TokenTypeBlockComment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TokenizerConfig{}
			tokenizer := newTokenizer(cfg)
			token := tokenizer.getBlockCommentToken(tt.input)
			require.False(t, token.Empty(), "Block comment token should not be empty")
			require.Equal(t, tt.expectedType, token.Type)
			require.Equal(t, tt.expectedValue, token.Value)
		})
	}
}

// ============================================================================
// String Literal Tests
// ============================================================================

func TestTokenizerStringLiterals(t *testing.T) {
	tests := []struct {
		name          string
		stringTypes   []string
		input         string
		expectedValue string
		expectedType  types.TokenType
	}{
		// Single-quoted strings
		{
			name:          "single-quoted string",
			stringTypes:   []string{"''"},
			input:         "'hello world'",
			expectedValue: "'hello world'",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "single-quoted with escaped quote",
			stringTypes:   []string{"''"},
			input:         "'it\\'s working'",
			expectedValue: "'it\\'s working'",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "empty single-quoted string",
			stringTypes:   []string{"''"},
			input:         "''",
			expectedValue: "''",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "unclosed single-quoted string",
			stringTypes:   []string{"''"},
			input:         "'unclosed",
			expectedValue: "'unclosed",
			expectedType:  types.TokenTypeString,
		},
		// Double-quoted strings
		{
			name:          "double-quoted string",
			stringTypes:   []string{"\"\""},
			input:         "\"hello world\"",
			expectedValue: "\"hello world\"",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "double-quoted with escaped quote",
			stringTypes:   []string{"\"\""},
			input:         "\"say \\\"hello\\\"\"",
			expectedValue: "\"say \\\"hello\\\"\"",
			expectedType:  types.TokenTypeString,
		},
		// Backtick strings
		{
			name:          "backtick string",
			stringTypes:   []string{"``"},
			input:         "`column_name`",
			expectedValue: "`column_name`",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "backtick with spaces",
			stringTypes:   []string{"``"},
			input:         "`my column`",
			expectedValue: "`my column`",
			expectedType:  types.TokenTypeString,
		},
		// Bracket strings
		{
			name:          "bracket string",
			stringTypes:   []string{"[]"},
			input:         "[column name]",
			expectedValue: "[column name]",
			expectedType:  types.TokenTypeString,
		},
		// N-prefixed strings
		{
			name:          "N-prefixed string",
			stringTypes:   []string{"N''"},
			input:         "N'unicode text'",
			expectedValue: "N'unicode text'",
			expectedType:  types.TokenTypeString,
		},
		// Hex strings
		{
			name:          "hex string lowercase",
			stringTypes:   []string{"X''"},
			input:         "x'48656c6c6f'",
			expectedValue: "x'48656c6c6f'",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "hex string uppercase",
			stringTypes:   []string{"X''"},
			input:         "X'48656C6C6F'",
			expectedValue: "X'48656C6C6F'",
			expectedType:  types.TokenTypeString,
		},
		// Binary strings
		{
			name:          "binary string lowercase",
			stringTypes:   []string{"B''"},
			input:         "b'010101'",
			expectedValue: "b'010101'",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "binary string uppercase",
			stringTypes:   []string{"B''"},
			input:         "B'111000'",
			expectedValue: "B'111000'",
			expectedType:  types.TokenTypeString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TokenizerConfig{
				StringTypes: tt.stringTypes,
			}
			tokenizer := newTokenizer(cfg)
			token := tokenizer.getStringToken(tt.input)
			require.False(t, token.Empty(), "String token should not be empty")
			require.Equal(t, tt.expectedType, token.Type)
			require.Equal(t, tt.expectedValue, token.Value)
		})
	}
}

func TestTokenizerDollarQuotedStrings(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValue string
		expectedType  types.TokenType
	}{
		{
			name:          "simple dollar-quoted string",
			input:         "$$hello world$$",
			expectedValue: "$$hello world$$",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "dollar-quoted with tag",
			input:         "$tag$hello world$tag$",
			expectedValue: "$tag$hello world$tag$",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "dollar-quoted multi-line",
			input:         "$$line 1\nline 2\nline 3$$",
			expectedValue: "$$line 1\nline 2\nline 3$$",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "dollar-quoted with nested single quotes",
			input:         "$$it's a 'test'$$",
			expectedValue: "$$it's a 'test'$$",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "dollar-quoted with function",
			input:         "$func$SELECT * FROM users$func$",
			expectedValue: "$func$SELECT * FROM users$func$",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "unclosed dollar-quoted",
			input:         "$$unclosed string",
			expectedValue: "$$unclosed string",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "dollar-quoted with underscore tag",
			input:         "$my_tag$content$my_tag$",
			expectedValue: "$my_tag$content$my_tag$",
			expectedType:  types.TokenTypeString,
		},
		{
			name:          "dollar-quoted with numeric tag",
			input:         "$tag123$content$tag123$",
			expectedValue: "$tag123$content$tag123$",
			expectedType:  types.TokenTypeString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TokenizerConfig{}
			tokenizer := newTokenizer(cfg)
			token := tokenizer.getDollarQuotedToken(tt.input)
			require.False(t, token.Empty(), "Dollar-quoted token should not be empty")
			require.Equal(t, tt.expectedType, token.Type)
			require.Equal(t, tt.expectedValue, token.Value)
		})
	}
}

// ============================================================================
// Operator Tests
// ============================================================================

func TestTokenizerOperators(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValue string
		expectedType  types.TokenType
	}{
		// Comparison operators
		{name: "not equal !=", input: "!= 5", expectedValue: "!=", expectedType: types.TokenTypeOperator},
		{name: "not equal <>", input: "<> 5", expectedValue: "<>", expectedType: types.TokenTypeOperator},
		{name: "null-safe equal <=>", input: "<=> NULL", expectedValue: "<=>", expectedType: types.TokenTypeOperator},
		{name: "equal ==", input: "== 5", expectedValue: "==", expectedType: types.TokenTypeOperator},
		{name: "less than or equal <=", input: "<= 5", expectedValue: "<=", expectedType: types.TokenTypeOperator},
		{name: "greater than or equal >=", input: ">= 5", expectedValue: ">=", expectedType: types.TokenTypeOperator},
		{name: "fat arrow =>", input: "=> value", expectedValue: "=>", expectedType: types.TokenTypeOperator},
		{name: "not less than !<", input: "!< 5", expectedValue: "!<", expectedType: types.TokenTypeOperator},
		{name: "not greater than !>", input: "!> 5", expectedValue: "!>", expectedType: types.TokenTypeOperator},
		// String operators
		{name: "concatenation ||", input: "|| 'text'", expectedValue: "||", expectedType: types.TokenTypeOperator},
		// Type cast operator
		{name: "type cast ::", input: "::integer", expectedValue: "::", expectedType: types.TokenTypeOperator},
		// JSON operators
		{name: "json arrow ->>", input: "->> 'key'", expectedValue: "->>", expectedType: types.TokenTypeOperator},
		{name: "json arrow ->", input: "-> 'key'", expectedValue: "->", expectedType: types.TokenTypeOperator},
		{name: "json path #>>", input: "#>> '{a,b}'", expectedValue: "#>>", expectedType: types.TokenTypeOperator},
		{name: "json path #>", input: "#> '{a,b}'", expectedValue: "#>", expectedType: types.TokenTypeOperator},
		// Bit shift operators
		{name: "left shift <<", input: "<< 2", expectedValue: "<<", expectedType: types.TokenTypeOperator},
		{name: "right shift >>", input: ">> 2", expectedValue: ">>", expectedType: types.TokenTypeOperator},
		// JSON existence operators
		{name: "json exists ?|", input: "?| array", expectedValue: "?|", expectedType: types.TokenTypeOperator},
		{name: "json exists ?&", input: "?& array", expectedValue: "?&", expectedType: types.TokenTypeOperator},
		{name: "json exists ?", input: "? 'key'", expectedValue: "?", expectedType: types.TokenTypeOperator},
		// JSON containment operators
		{name: "json contains @>", input: "@> '{}'", expectedValue: "@>", expectedType: types.TokenTypeOperator},
		{name: "json contained <@", input: "<@ '{}'", expectedValue: "<@", expectedType: types.TokenTypeOperator},
		// Pattern matching operators
		{name: "case insensitive like ~~*", input: "~~* 'pattern'", expectedValue: "~~*", expectedType: types.TokenTypeOperator},
		{name: "like ~~", input: "~~ 'pattern'", expectedValue: "~~", expectedType: types.TokenTypeOperator},
		{name: "not like case insensitive !~~*", input: "!~~* 'pattern'", expectedValue: "!~~*", expectedType: types.TokenTypeOperator},
		{name: "not like !~~", input: "!~~ 'pattern'", expectedValue: "!~~", expectedType: types.TokenTypeOperator},
		{name: "regex case insensitive ~*", input: "~* 'regex'", expectedValue: "~*", expectedType: types.TokenTypeOperator},
		{name: "not regex case insensitive !~*", input: "!~* 'regex'", expectedValue: "!~*", expectedType: types.TokenTypeOperator},
		{name: "not regex !~", input: "!~ 'regex'", expectedValue: "!~", expectedType: types.TokenTypeOperator},
		// Arithmetic operators
		{name: "plus +", input: "+ 5", expectedValue: "+", expectedType: types.TokenTypeOperator},
		{name: "minus -", input: "- 5", expectedValue: "-", expectedType: types.TokenTypeOperator},
		{name: "multiply *", input: "* 5", expectedValue: "*", expectedType: types.TokenTypeOperator},
		{name: "divide /", input: "/ 5", expectedValue: "/", expectedType: types.TokenTypeOperator},
		{name: "modulo %", input: "% 5", expectedValue: "%", expectedType: types.TokenTypeOperator},
		// Other operators
		{name: "less than <", input: "< 5", expectedValue: "<", expectedType: types.TokenTypeOperator},
		{name: "greater than >", input: "> 5", expectedValue: ">", expectedType: types.TokenTypeOperator},
		{name: "equal =", input: "= 5", expectedValue: "=", expectedType: types.TokenTypeOperator},
		{name: "dot .", input: ". column", expectedValue: ".", expectedType: types.TokenTypeOperator},
		{name: "comma ,", input: ", value", expectedValue: ",", expectedType: types.TokenTypeOperator},
		{name: "semicolon ;", input: "; SELECT", expectedValue: ";", expectedType: types.TokenTypeOperator},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TokenizerConfig{}
			tokenizer := newTokenizer(cfg)
			token := tokenizer.getOperatorToken(tt.input)
			require.False(t, token.Empty(), "Operator token should not be empty")
			require.Equal(t, tt.expectedType, token.Type)
			require.Equal(t, tt.expectedValue, token.Value)
		})
	}
}

// ============================================================================
// Number Tests
// ============================================================================

func TestTokenizerNumbers(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValue string
		expectedType  types.TokenType
	}{
		// Integers
		{name: "positive integer", input: "123", expectedValue: "123", expectedType: types.TokenTypeNumber},
		{name: "zero", input: "0", expectedValue: "0", expectedType: types.TokenTypeNumber},
		{name: "large integer", input: "999999999", expectedValue: "999999999", expectedType: types.TokenTypeNumber},
		// Negative numbers
		{name: "negative integer", input: "-123", expectedValue: "-123", expectedType: types.TokenTypeNumber},
		{name: "negative with space", input: "- 123", expectedValue: "- 123", expectedType: types.TokenTypeNumber},
		// Decimals
		{name: "decimal", input: "123.456", expectedValue: "123.456", expectedType: types.TokenTypeNumber},
		{name: "negative decimal", input: "-123.456", expectedValue: "-123.456", expectedType: types.TokenTypeNumber},
		{name: "zero decimal", input: "0.0", expectedValue: "0.0", expectedType: types.TokenTypeNumber},
		// Hexadecimal
		{name: "hex lowercase", input: "0x1a2b3c", expectedValue: "0x1a2b3c", expectedType: types.TokenTypeNumber},
		{name: "hex uppercase", input: "0x1A2B3C", expectedValue: "0x1A2B3C", expectedType: types.TokenTypeNumber},
		{name: "hex mixed case", input: "0xAbCdEf", expectedValue: "0xAbCdEf", expectedType: types.TokenTypeNumber},
		{name: "hex zero", input: "0x0", expectedValue: "0x0", expectedType: types.TokenTypeNumber},
		// Binary
		{name: "binary", input: "0b101010", expectedValue: "0b101010", expectedType: types.TokenTypeNumber},
		{name: "binary all ones", input: "0b111111", expectedValue: "0b111111", expectedType: types.TokenTypeNumber},
		{name: "binary all zeros", input: "0b000000", expectedValue: "0b000000", expectedType: types.TokenTypeNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TokenizerConfig{}
			tokenizer := newTokenizer(cfg)
			token := tokenizer.getNumberToken(tt.input)
			require.False(t, token.Empty(), "Number token should not be empty")
			require.Equal(t, tt.expectedType, token.Type)
			require.Equal(t, tt.expectedValue, token.Value)
		})
	}
}

// ============================================================================
// Whitespace Tests
// ============================================================================

func TestTokenizerWhitespace(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValue string
		expectedType  types.TokenType
	}{
		{name: "single space", input: " SELECT", expectedValue: " ", expectedType: types.TokenTypeWhitespace},
		{name: "multiple spaces", input: "    SELECT", expectedValue: "    ", expectedType: types.TokenTypeWhitespace},
		{name: "tab", input: "\tSELECT", expectedValue: "\t", expectedType: types.TokenTypeWhitespace},
		{name: "multiple tabs", input: "\t\t\tSELECT", expectedValue: "\t\t\t", expectedType: types.TokenTypeWhitespace},
		{name: "newline", input: "\nSELECT", expectedValue: "\n", expectedType: types.TokenTypeWhitespace},
		{name: "carriage return", input: "\rSELECT", expectedValue: "\r", expectedType: types.TokenTypeWhitespace},
		{name: "CRLF", input: "\r\nSELECT", expectedValue: "\r\n", expectedType: types.TokenTypeWhitespace},
		{name: "mixed whitespace", input: " \t\n\r SELECT", expectedValue: " \t\n\r ", expectedType: types.TokenTypeWhitespace},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TokenizerConfig{}
			tokenizer := newTokenizer(cfg)
			token := tokenizer.getWhitespaceToken(tt.input)
			require.False(t, token.Empty(), "Whitespace token should not be empty")
			require.Equal(t, tt.expectedType, token.Type)
			require.Equal(t, tt.expectedValue, token.Value)
		})
	}
}

// ============================================================================
// Parentheses Tests
// ============================================================================

func TestTokenizerParentheses(t *testing.T) {
	tests := []struct {
		name          string
		openParens    []string
		closeParens   []string
		input         string
		expectedValue string
		expectedType  types.TokenType
		testOpen      bool
	}{
		{name: "open paren", openParens: []string{"("}, input: "(SELECT", expectedValue: "(", expectedType: types.TokenTypeOpenParen, testOpen: true},
		{name: "close paren", closeParens: []string{")"}, input: ") SELECT", expectedValue: ")", expectedType: types.TokenTypeCloseParen, testOpen: false},
		{name: "open bracket", openParens: []string{"["}, input: "[column]", expectedValue: "[", expectedType: types.TokenTypeOpenParen, testOpen: true},
		{name: "close bracket", closeParens: []string{"]"}, input: "] FROM", expectedValue: "]", expectedType: types.TokenTypeCloseParen, testOpen: false},
		{name: "open CASE", openParens: []string{"CASE"}, input: "CASE WHEN", expectedValue: "CASE", expectedType: types.TokenTypeOpenParen, testOpen: true},
		{name: "close END", closeParens: []string{"END"}, input: "END FROM", expectedValue: "END", expectedType: types.TokenTypeCloseParen, testOpen: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TokenizerConfig{
				OpenParens:  tt.openParens,
				CloseParens: tt.closeParens,
			}
			tokenizer := newTokenizer(cfg)

			var token types.Token
			if tt.testOpen {
				token = tokenizer.getOpenParenToken(tt.input)
			} else {
				token = tokenizer.getCloseParenToken(tt.input)
			}

			require.False(t, token.Empty(), "Paren token should not be empty")
			require.Equal(t, tt.expectedType, token.Type)
			require.Equal(t, tt.expectedValue, token.Value)
		})
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestTokenizerEdgeCases(t *testing.T) {
	cfg := &TokenizerConfig{
		ReservedWords:         []string{"SELECT", "FROM", "WHERE"},
		ReservedTopLevelWords: []string{"SELECT", "FROM"},
		StringTypes:           []string{"''", "\"\""},
		OpenParens:            []string{"("},
		CloseParens:           []string{")"},
		LineCommentTypes:      []string{"--"},
	}
	tokenizer := newTokenizer(cfg)

	t.Run("empty input", func(t *testing.T) {
		tokens := tokenizer.tokenize("")
		require.Empty(t, tokens, "Empty input should produce no tokens")
	})

	t.Run("very long identifier", func(t *testing.T) {
		longIdentifier := strings.Repeat("a", 1500)
		token := tokenizer.getWordToken(longIdentifier)
		require.False(t, token.Empty())
		require.Equal(t, longIdentifier, token.Value)
	})

	t.Run("very long string", func(t *testing.T) {
		longString := "'" + strings.Repeat("x", 2000) + "'"
		token := tokenizer.getStringToken(longString)
		require.False(t, token.Empty())
		require.Equal(t, longString, token.Value)
	})

	t.Run("unicode characters in identifiers", func(t *testing.T) {
		unicodeIdentifiers := []string{
			"tableナマエ",
			"用户表",
			"таблица",
			"πίνακας",
			"café",
		}
		for _, identifier := range unicodeIdentifiers {
			token := tokenizer.getWordToken(identifier)
			require.False(t, token.Empty(), "Should tokenize unicode identifier: %s", identifier)
			require.Contains(t, token.Value, identifier[:len(identifier)/2])
		}
	})

	t.Run("unicode in strings", func(t *testing.T) {
		unicodeStrings := []string{
			"'Hello 世界'",
			"'Привет мир'",
			"'مرحبا العالم'",
			"'שלום עולם'",
			"'🚀🌟💻'",
		}
		for _, str := range unicodeStrings {
			token := tokenizer.getStringToken(str)
			require.False(t, token.Empty(), "Should tokenize unicode string: %s", str)
			require.Equal(t, str, token.Value)
		}
	})

	t.Run("malformed input - unclosed string", func(t *testing.T) {
		input := "'unclosed string SELECT FROM"
		token := tokenizer.getStringToken(input)
		require.False(t, token.Empty())
		require.Equal(t, input, token.Value, "Should return entire unclosed string")
	})

	t.Run("malformed input - unclosed block comment", func(t *testing.T) {
		input := "/* unclosed comment SELECT FROM"
		token := tokenizer.getBlockCommentToken(input)
		require.False(t, token.Empty())
		require.Equal(t, input, token.Value, "Should return entire unclosed comment")
	})

	t.Run("mixed special characters", func(t *testing.T) {
		input := "@#$%^&*()"
		// Should tokenize as individual operators
		tokens := tokenizer.tokenize(input)
		require.NotEmpty(t, tokens)
	})

	t.Run("consecutive operators", func(t *testing.T) {
		input := "+-*/"
		tokens := tokenizer.tokenize(input)
		require.Len(t, tokens, 4)
		for _, tok := range tokens {
			require.Equal(t, types.TokenTypeOperator, tok.Type)
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		input := "   \t\n\r   "
		tokens := tokenizer.tokenize(input)
		require.Len(t, tokens, 1)
		require.Equal(t, types.TokenTypeWhitespace, tokens[0].Type)
	})
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestTokenizerComplexQueries(t *testing.T) {
	cfg := &TokenizerConfig{
		ReservedWords: []string{
			"SELECT", "FROM", "WHERE", "AND", "OR", "ORDER BY", "GROUP BY",
			"INNER JOIN", "LEFT JOIN", "ON", "AS", "IN", "LIMIT",
		},
		ReservedTopLevelWords:         []string{"SELECT", "FROM"},
		ReservedNewlineWords:          []string{"AND", "OR"},
		ReservedTopLevelWordsNoIndent: []string{},
		StringTypes:                   []string{"''", "\"\"", "``"},
		OpenParens:                    []string{"(", "CASE"},
		CloseParens:                   []string{")", "END"},
		IndexedPlaceholderTypes:       []string{"?", "$"},
		NamedPlaceholderTypes:         []string{"@", ":"},
		LineCommentTypes:              []string{"--", "#"},
	}
	tokenizer := newTokenizer(cfg)

	t.Run("complex SELECT with JOINs", func(t *testing.T) {
		query := `SELECT u.id, u.name, o.total
		          FROM users u
		          INNER JOIN orders o ON u.id = o.user_id
		          WHERE u.active = true AND o.status = 'completed'
		          ORDER BY o.created_at DESC
		          LIMIT 10`
		tokens := tokenizer.tokenize(query)
		require.NotEmpty(t, tokens)

		// Check for expected reserved words
		foundSelect := false
		foundFrom := false
		foundInnerJoin := false
		foundWhere := false
		foundOrderBy := false

		for _, tok := range tokens {
			switch tok.Value {
			case "SELECT":
				foundSelect = true
			case "FROM":
				foundFrom = true
			case "INNER JOIN":
				foundInnerJoin = true
			case "WHERE":
				foundWhere = true
			case "ORDER BY":
				foundOrderBy = true
			}
		}

		require.True(t, foundSelect, "Should find SELECT")
		require.True(t, foundFrom, "Should find FROM")
		require.True(t, foundInnerJoin, "Should find INNER JOIN")
		require.True(t, foundWhere, "Should find WHERE")
		require.True(t, foundOrderBy, "Should find ORDER BY")
	})

	t.Run("query with comments and strings", func(t *testing.T) {
		query := `-- Get active users
		          SELECT id, name, 'test value' AS status
		          FROM users
		          WHERE email = "test@example.com" -- Filter by email`
		tokens := tokenizer.tokenize(query)
		require.NotEmpty(t, tokens)

		// Count token types
		lineComments := 0
		strings := 0
		reservedWords := 0

		for _, tok := range tokens {
			switch tok.Type {
			case types.TokenTypeLineComment:
				lineComments++
			case types.TokenTypeString:
				strings++
			case types.TokenTypeReserved, types.TokenTypeReservedTopLevel:
				reservedWords++
			}
		}

		require.Equal(t, 2, lineComments, "Should find 2 line comments")
		require.Equal(t, 2, strings, "Should find 2 strings")
		require.Positive(t, reservedWords, "Should find reserved words")
	})

	t.Run("query with placeholders", func(t *testing.T) {
		query := "SELECT * FROM users WHERE id = $1 AND email = :email AND status = @status"
		tokens := tokenizer.tokenize(query)
		require.NotEmpty(t, tokens)

		placeholders := 0
		for _, tok := range tokens {
			if tok.Type == types.TokenTypePlaceholder {
				placeholders++
			}
		}

		require.Equal(t, 3, placeholders, "Should find 3 placeholders")
	})

	t.Run("query with CASE expression", func(t *testing.T) {
		query := `SELECT CASE WHEN status = 'active' THEN 1 ELSE 0 END FROM users`
		tokens := tokenizer.tokenize(query)
		require.NotEmpty(t, tokens)

		foundCase := false
		foundEnd := false

		for _, tok := range tokens {
			if tok.Value == "CASE" && tok.Type == types.TokenTypeOpenParen {
				foundCase = true
			}
			if tok.Value == "END" && tok.Type == types.TokenTypeCloseParen {
				foundEnd = true
			}
		}

		require.True(t, foundCase, "Should find CASE as open paren")
		require.True(t, foundEnd, "Should find END as close paren")
	})

	t.Run("query with subquery", func(t *testing.T) {
		query := `SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > 100)`
		tokens := tokenizer.tokenize(query)
		require.NotEmpty(t, tokens)

		openParens := 0
		closeParens := 0

		for _, tok := range tokens {
			if tok.Type == types.TokenTypeOpenParen {
				openParens++
			}
			if tok.Type == types.TokenTypeCloseParen {
				closeParens++
			}
		}

		require.Equal(t, openParens, closeParens, "Open and close parens should match")
		require.Positive(t, openParens, "Should have parentheses")
	})
}

func TestTokenizerDialectSpecificFeatures(t *testing.T) {
	t.Run("PostgreSQL dollar-quoted functions", func(t *testing.T) {
		cfg := &TokenizerConfig{
			ReservedWords:                 []string{"CREATE", "FUNCTION", "RETURNS", "AS", "LANGUAGE"},
			ReservedTopLevelWords:         []string{"CREATE"},
			ReservedNewlineWords:          []string{},
			ReservedTopLevelWordsNoIndent: []string{},
			StringTypes:                   []string{"''"},
			OpenParens:                    []string{"("},
			CloseParens:                   []string{")"},
			LineCommentTypes:              []string{"--"},
		}
		tokenizer := newTokenizer(cfg)

		query := `CREATE FUNCTION test_func() RETURNS integer AS $$
		          BEGIN
		              RETURN 42;
		          END;
		          $$ LANGUAGE plpgsql`

		tokens := tokenizer.tokenize(query)
		require.NotEmpty(t, tokens)

		foundDollarQuoted := false
		for _, tok := range tokens {
			if tok.Type == types.TokenTypeString && strings.HasPrefix(tok.Value, "$$") {
				foundDollarQuoted = true
				require.Contains(t, tok.Value, "BEGIN")
				require.Contains(t, tok.Value, "RETURN 42")
			}
		}

		require.True(t, foundDollarQuoted, "Should find dollar-quoted string")
	})

	t.Run("MySQL backtick identifiers", func(t *testing.T) {
		cfg := &TokenizerConfig{
			ReservedWords:                 []string{"SELECT", "FROM"},
			ReservedTopLevelWords:         []string{"SELECT", "FROM"},
			ReservedNewlineWords:          []string{},
			ReservedTopLevelWordsNoIndent: []string{},
			StringTypes:                   []string{"``"},
			OpenParens:                    []string{"("},
			CloseParens:                   []string{")"},
			LineCommentTypes:              []string{"--"},
		}
		tokenizer := newTokenizer(cfg)

		query := "SELECT `user_id`, `user_name` FROM `user_table`"
		tokens := tokenizer.tokenize(query)
		require.NotEmpty(t, tokens)

		backtickStrings := 0
		for _, tok := range tokens {
			if tok.Type == types.TokenTypeString && strings.HasPrefix(tok.Value, "`") {
				backtickStrings++
			}
		}

		require.Equal(t, 3, backtickStrings, "Should find 3 backtick-quoted identifiers")
	})

	t.Run("PostgreSQL JSON operators", func(t *testing.T) {
		cfg := &TokenizerConfig{
			ReservedWords:                 []string{"SELECT", "FROM", "WHERE"},
			ReservedTopLevelWords:         []string{"SELECT", "FROM"},
			ReservedNewlineWords:          []string{},
			ReservedTopLevelWordsNoIndent: []string{},
			StringTypes:                   []string{"''"},
			OpenParens:                    []string{"("},
			CloseParens:                   []string{")"},
			LineCommentTypes:              []string{"--"},
		}
		tokenizer := newTokenizer(cfg)

		// Test individual operators first
		arrowOp := tokenizer.getOperatorToken("-> 'key'")
		require.Equal(t, "->", arrowOp.Value, "Should tokenize -> operator")

		doubleArrowOp := tokenizer.getOperatorToken("->> 'key'")
		require.Equal(t, "->>", doubleArrowOp.Value, "Should tokenize ->> operator")

		containsOp := tokenizer.getOperatorToken("@> '{}'")
		require.Equal(t, "@>", containsOp.Value, "Should tokenize @> operator")

		// Now test in full query - just verify operators are present
		query := "SELECT id FROM tbl WHERE col @> val"
		tokens := tokenizer.tokenize(query)
		require.NotEmpty(t, tokens)

		foundContainsOp := false
		for _, tok := range tokens {
			if tok.Type == types.TokenTypeOperator && tok.Value == "@>" {
				foundContainsOp = true
				break
			}
		}
		require.True(t, foundContainsOp, "Should find @> operator in query")
	})
}

func TestTokenizerGetNextToken(t *testing.T) {
	cfg := &TokenizerConfig{
		ReservedWords:           []string{"SELECT", "FROM"},
		ReservedTopLevelWords:   []string{"SELECT", "FROM"},
		ReservedNewlineWords:    []string{"AND", "OR"},
		StringTypes:             []string{"''"},
		OpenParens:              []string{"("},
		CloseParens:             []string{")"},
		IndexedPlaceholderTypes: []string{"?"},
		LineCommentTypes:        []string{"--"},
	}
	tokenizer := newTokenizer(cfg)

	tests := []struct {
		name         string
		input        string
		expectedType types.TokenType
		prevToken    types.Token
	}{
		{
			name:         "whitespace first",
			input:        " SELECT",
			expectedType: types.TokenTypeWhitespace,
			prevToken:    types.Token{},
		},
		{
			name:         "comment first",
			input:        "-- comment\nSELECT",
			expectedType: types.TokenTypeLineComment,
			prevToken:    types.Token{},
		},
		{
			name:         "string first",
			input:        "'string' SELECT",
			expectedType: types.TokenTypeString,
			prevToken:    types.Token{},
		},
		{
			name:         "open paren first",
			input:        "(SELECT",
			expectedType: types.TokenTypeOpenParen,
			prevToken:    types.Token{},
		},
		{
			name:         "close paren first",
			input:        ") FROM",
			expectedType: types.TokenTypeCloseParen,
			prevToken:    types.Token{},
		},
		{
			name:         "placeholder first",
			input:        "? FROM",
			expectedType: types.TokenTypePlaceholder,
			prevToken:    types.Token{},
		},
		{
			name:         "number first",
			input:        "123 FROM",
			expectedType: types.TokenTypeNumber,
			prevToken:    types.Token{},
		},
		{
			name:         "reserved word first",
			input:        "SELECT *",
			expectedType: types.TokenTypeReservedTopLevel,
			prevToken:    types.Token{},
		},
		{
			name:         "boolean first",
			input:        "true AND",
			expectedType: types.TokenTypeBoolean,
			prevToken:    types.Token{},
		},
		{
			name:         "word first",
			input:        "column_name FROM",
			expectedType: types.TokenTypeWord,
			prevToken:    types.Token{},
		},
		{
			name:         "operator first",
			input:        "+ 5",
			expectedType: types.TokenTypeOperator,
			prevToken:    types.Token{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tokenizer.getNextToken(tt.input, tt.prevToken)
			require.False(t, token.Empty(), "Should get a token")
			require.Equal(t, tt.expectedType, token.Type, "Token type should match")
		})
	}
}

func TestTokenizerFullTokenization(t *testing.T) {
	cfg := &TokenizerConfig{
		ReservedWords: []string{
			"SELECT", "FROM", "WHERE", "AND", "OR", "ORDER BY",
			"INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE",
		},
		ReservedTopLevelWords:         []string{"SELECT", "FROM", "INSERT", "UPDATE", "DELETE"},
		ReservedNewlineWords:          []string{"AND", "OR"},
		ReservedTopLevelWordsNoIndent: []string{},
		StringTypes:                   []string{"''"},
		OpenParens:                    []string{"("},
		CloseParens:                   []string{")"},
		IndexedPlaceholderTypes:       []string{"?"},
		LineCommentTypes:              []string{"--"},
	}
	tokenizer := newTokenizer(cfg)

	query := `SELECT id, name, email FROM users WHERE active = true ORDER BY created_at`
	tokens := tokenizer.tokenize(query)

	require.NotEmpty(t, tokens, "Should produce tokens")

	// Verify no empty tokens
	for i, tok := range tokens {
		require.False(t, tok.Empty(), "Token at index %d should not be empty", i)
		require.NotEmpty(t, tok.Value, "Token value at index %d should not be empty", i)
	}

	// Verify token sequence makes sense
	hasSelect := false
	hasFrom := false
	hasWhere := false
	hasOrderBy := false

	for _, tok := range tokens {
		switch tok.Value {
		case "SELECT":
			hasSelect = true
		case "FROM":
			hasFrom = true
		case "WHERE":
			hasWhere = true
		case "ORDER BY":
			hasOrderBy = true
		}
	}

	require.True(t, hasSelect)
	require.True(t, hasFrom)
	require.True(t, hasWhere)
	require.True(t, hasOrderBy)
}
