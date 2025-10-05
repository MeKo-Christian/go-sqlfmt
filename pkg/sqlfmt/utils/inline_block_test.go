package utils

import (
	"testing"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
	"github.com/stretchr/testify/require"
)

// TestNewInlineBlock tests the NewInlineBlock constructor.
func TestNewInlineBlock(t *testing.T) {
	ib := NewInlineBlock()
	require.NotNil(t, ib)
	require.False(t, ib.IsActive(), "New inline block should not be active")
}

// TestIsActive tests the IsActive method.
func TestIsActive(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*InlineBlock)
		expected bool
	}{
		{
			name:     "initially inactive",
			setup:    func(ib *InlineBlock) {},
			expected: false,
		},
		{
			name: "active after beginning",
			setup: func(ib *InlineBlock) {
				// Simulate beginning an inline block
				ib.level = 1
			},
			expected: true,
		},
		{
			name: "active with nested levels",
			setup: func(ib *InlineBlock) {
				ib.level = 3
			},
			expected: true,
		},
		{
			name: "inactive after ending",
			setup: func(ib *InlineBlock) {
				ib.level = 1
				ib.End()
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ib := NewInlineBlock()
			tt.setup(ib)
			result := ib.IsActive()
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestEnd tests the End method.
func TestEnd(t *testing.T) {
	t.Run("end decreases level", func(t *testing.T) {
		ib := NewInlineBlock()
		ib.level = 3

		ib.End()
		require.Equal(t, 2, ib.level)
		require.True(t, ib.IsActive())

		ib.End()
		require.Equal(t, 1, ib.level)
		require.True(t, ib.IsActive())

		ib.End()
		require.Equal(t, 0, ib.level)
		require.False(t, ib.IsActive())
	})

	t.Run("end can go negative", func(t *testing.T) {
		ib := NewInlineBlock()
		ib.End()
		require.Equal(t, -1, ib.level)
	})
}

// TestBeginIfPossible tests the BeginIfPossible method.
func TestBeginIfPossible(t *testing.T) {
	tests := []struct {
		name           string
		tokens         []types.Token
		index          int
		initialLevel   int
		expectedLevel  int
		expectedActive bool
	}{
		{
			name: "simple inline block - short expression",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeOperator, Value: "+"},
				{Type: types.TokenTypeNumber, Value: "1"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  1,
			expectedActive: true,
		},
		{
			name: "inline block - exact max length (50)",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "123456789012345678901234567890123456789012345678"}, // 48 chars + ( + ) = 50
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  1,
			expectedActive: true,
		},
		{
			name: "too long - exceeds max length",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "this_is_a_very_long_identifier_that_exceeds_fifty_characters_total"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  0,
			expectedActive: false,
		},
		{
			name: "contains top-level reserved word",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeReservedTopLevel, Value: "SELECT"},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  0,
			expectedActive: false,
		},
		{
			name: "contains newline reserved word",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeReservedNewline, Value: "AND"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  0,
			expectedActive: false,
		},
		{
			name: "contains line comment",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeLineComment, Value: "-- comment"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  0,
			expectedActive: false,
		},
		{
			name: "contains block comment",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeBlockComment, Value: "/* comment */"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  0,
			expectedActive: false,
		},
		{
			name: "contains semicolon",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeOperator, Value: ";"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  0,
			expectedActive: false,
		},
		{
			name: "nested parentheses - valid",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "a"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
				{Type: types.TokenTypeOperator, Value: "+"},
				{Type: types.TokenTypeWord, Value: "b"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  1,
			expectedActive: true,
		},
		{
			name: "already inside inline block - increases level",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   1,
			expectedLevel:  2,
			expectedActive: true,
		},
		{
			name: "already inside inline block - deeply nested",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   5,
			expectedLevel:  6,
			expectedActive: true,
		},
		{
			name: "no closing paren - returns false",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeWord, Value: "y"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  0,
			expectedActive: false,
		},
		{
			name: "empty parentheses",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  1,
			expectedActive: true,
		},
		{
			name: "function call with multiple args",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "a"},
				{Type: types.TokenTypeOperator, Value: ","},
				{Type: types.TokenTypeWord, Value: "b"},
				{Type: types.TokenTypeOperator, Value: ","},
				{Type: types.TokenTypeWord, Value: "c"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          0,
			initialLevel:   0,
			expectedLevel:  1,
			expectedActive: true,
		},
		{
			name: "start from non-zero index",
			tokens: []types.Token{
				{Type: types.TokenTypeWord, Value: "prefix"},
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:          1,
			initialLevel:   0,
			expectedLevel:  1,
			expectedActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ib := NewInlineBlock()
			ib.level = tt.initialLevel

			ib.BeginIfPossible(tt.tokens, tt.index)

			require.Equal(t, tt.expectedLevel, ib.level)
			require.Equal(t, tt.expectedActive, ib.IsActive())
		})
	}
}

// TestIsForbiddenToken tests the isForbiddenToken method.
func TestIsForbiddenToken(t *testing.T) {
	tests := []struct {
		name     string
		token    types.Token
		expected bool
	}{
		{
			name:     "top-level reserved word is forbidden",
			token:    types.Token{Type: types.TokenTypeReservedTopLevel, Value: "SELECT"},
			expected: true,
		},
		{
			name:     "newline reserved word is forbidden",
			token:    types.Token{Type: types.TokenTypeReservedNewline, Value: "AND"},
			expected: true,
		},
		{
			name:     "line comment is forbidden",
			token:    types.Token{Type: types.TokenTypeLineComment, Value: "-- comment"},
			expected: true,
		},
		{
			name:     "block comment is forbidden",
			token:    types.Token{Type: types.TokenTypeBlockComment, Value: "/* comment */"},
			expected: true,
		},
		{
			name:     "semicolon is forbidden",
			token:    types.Token{Type: types.TokenTypeOperator, Value: ";"},
			expected: true,
		},
		{
			name:     "regular word is allowed",
			token:    types.Token{Type: types.TokenTypeWord, Value: "column_name"},
			expected: false,
		},
		{
			name:     "number is allowed",
			token:    types.Token{Type: types.TokenTypeNumber, Value: "123"},
			expected: false,
		},
		{
			name:     "string is allowed",
			token:    types.Token{Type: types.TokenTypeString, Value: "'hello'"},
			expected: false,
		},
		{
			name:     "operator is allowed (except semicolon)",
			token:    types.Token{Type: types.TokenTypeOperator, Value: "+"},
			expected: false,
		},
		{
			name:     "comma is allowed",
			token:    types.Token{Type: types.TokenTypeOperator, Value: ","},
			expected: false,
		},
		{
			name:     "parentheses are allowed",
			token:    types.Token{Type: types.TokenTypeOpenParen, Value: "("},
			expected: false,
		},
		{
			name:     "regular reserved word is allowed",
			token:    types.Token{Type: types.TokenTypeReserved, Value: "NULL"},
			expected: false,
		},
		{
			name:     "placeholder is allowed",
			token:    types.Token{Type: types.TokenTypePlaceholder, Value: "$1"},
			expected: false,
		},
		{
			name:     "boolean is allowed",
			token:    types.Token{Type: types.TokenTypeBoolean, Value: "true"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ib := NewInlineBlock()
			result := ib.isForbiddenToken(tt.token)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInlineBlock tests the isInlineBlock method directly.
func TestIsInlineBlock(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []types.Token
		index    int
		expected bool
	}{
		{
			name: "simple valid inline block",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:    0,
			expected: true,
		},
		{
			name: "exceeds max length",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "very_long_identifier_that_definitely_exceeds_the_max_length_of_fifty_characters"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:    0,
			expected: false,
		},
		{
			name: "has forbidden token",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeReservedTopLevel, Value: "SELECT"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:    0,
			expected: false,
		},
		{
			name: "no closing paren",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
			},
			index:    0,
			expected: false,
		},
		{
			name: "nested valid parens",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "a"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:    0,
			expected: true,
		},
		{
			name: "deeply nested parens - still valid if short",
			tokens: []types.Token{
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:    0,
			expected: true,
		},
		{
			name:     "empty token slice",
			tokens:   []types.Token{},
			index:    0,
			expected: false,
		},
		{
			name: "start from middle index",
			tokens: []types.Token{
				{Type: types.TokenTypeWord, Value: "prefix"},
				{Type: types.TokenTypeOpenParen, Value: "("},
				{Type: types.TokenTypeWord, Value: "x"},
				{Type: types.TokenTypeCloseParen, Value: ")"},
			},
			index:    1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ib := NewInlineBlock()
			result := ib.isInlineBlock(tt.tokens, tt.index)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestInlineBlockComplexScenarios tests complex real-world scenarios.
func TestInlineBlockComplexScenarios(t *testing.T) {
	t.Run("SQL function call - inline", func(t *testing.T) {
		tokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeWord, Value: "user_id"},
			{Type: types.TokenTypeOperator, Value: ","},
			{Type: types.TokenTypeString, Value: "'active'"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
		}

		ib := NewInlineBlock()
		ib.BeginIfPossible(tokens, 0)
		require.True(t, ib.IsActive())
		require.Equal(t, 1, ib.level)
	})

	t.Run("SQL subquery - not inline due to SELECT", func(t *testing.T) {
		tokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeReservedTopLevel, Value: "SELECT"},
			{Type: types.TokenTypeWord, Value: "id"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
		}

		ib := NewInlineBlock()
		ib.BeginIfPossible(tokens, 0)
		require.False(t, ib.IsActive())
	})

	t.Run("CASE expression arguments - inline", func(t *testing.T) {
		tokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeWord, Value: "status"},
			{Type: types.TokenTypeOperator, Value: "="},
			{Type: types.TokenTypeString, Value: "'active'"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
		}

		ib := NewInlineBlock()
		ib.BeginIfPossible(tokens, 0)
		require.True(t, ib.IsActive())
	})

	t.Run("nested inline blocks", func(t *testing.T) {
		// Outer block
		outerTokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeWord, Value: "a"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
		}

		ib := NewInlineBlock()
		ib.BeginIfPossible(outerTokens, 0)
		require.True(t, ib.IsActive())
		require.Equal(t, 1, ib.level)

		// Inner block - should just increment
		innerTokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeWord, Value: "b"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
		}

		ib.BeginIfPossible(innerTokens, 0)
		require.True(t, ib.IsActive())
		require.Equal(t, 2, ib.level)

		// End inner
		ib.End()
		require.Equal(t, 1, ib.level)

		// End outer
		ib.End()
		require.Equal(t, 0, ib.level)
		require.False(t, ib.IsActive())
	})

	t.Run("arithmetic expression", func(t *testing.T) {
		tokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeWord, Value: "price"},
			{Type: types.TokenTypeOperator, Value: "*"},
			{Type: types.TokenTypeNumber, Value: "1.1"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
		}

		ib := NewInlineBlock()
		ib.BeginIfPossible(tokens, 0)
		require.True(t, ib.IsActive())
	})
}

// TestInlineBlockEdgeCases tests edge cases and boundary conditions.
func TestInlineBlockEdgeCases(t *testing.T) {
	t.Run("exactly at max length boundary", func(t *testing.T) {
		// Create a token list that's exactly 50 chars
		longValue := ""
		for range 48 {
			longValue += "x"
		}

		tokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},  // 1 char
			{Type: types.TokenTypeWord, Value: longValue}, // 48 chars
			{Type: types.TokenTypeCloseParen, Value: ")"}, // 1 char
		}

		ib := NewInlineBlock()
		ib.BeginIfPossible(tokens, 0)
		require.True(t, ib.IsActive(), "Should be inline at exactly 50 chars")
	})

	t.Run("one char over max length", func(t *testing.T) {
		// Create a token list that's 51 chars
		longValue := ""
		for range 49 {
			longValue += "x"
		}

		tokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},  // 1 char
			{Type: types.TokenTypeWord, Value: longValue}, // 49 chars
			{Type: types.TokenTypeCloseParen, Value: ")"}, // 1 char
		}

		ib := NewInlineBlock()
		ib.BeginIfPossible(tokens, 0)
		require.False(t, ib.IsActive(), "Should not be inline at 51 chars")
	})

	t.Run("multiple end calls", func(t *testing.T) {
		ib := NewInlineBlock()
		ib.level = 1

		ib.End()
		require.Equal(t, 0, ib.level)

		ib.End()
		require.Equal(t, -1, ib.level)

		ib.End()
		require.Equal(t, -2, ib.level)
	})

	t.Run("semicolon with different token types", func(t *testing.T) {
		// Semicolon as operator
		tokens1 := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeOperator, Value: ";"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
		}

		ib1 := NewInlineBlock()
		ib1.BeginIfPossible(tokens1, 0)
		require.False(t, ib1.IsActive(), "Semicolon should prevent inline")

		// Semicolon as word (edge case)
		tokens2 := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeWord, Value: ";"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
		}

		ib2 := NewInlineBlock()
		ib2.BeginIfPossible(tokens2, 0)
		require.False(t, ib2.IsActive(), "Semicolon value should prevent inline")
	})

	t.Run("unbalanced nested parens", func(t *testing.T) {
		tokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeWord, Value: "x"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
			// Missing closing paren for outer level
		}

		ib := NewInlineBlock()
		result := ib.isInlineBlock(tokens, 0)
		require.False(t, result, "Unbalanced parens should not be inline")
	})
}

// TestInlineBlockStateMachine tests state transitions.
func TestInlineBlockStateMachine(t *testing.T) {
	t.Run("state transitions", func(t *testing.T) {
		ib := NewInlineBlock()

		// State 0: Inactive
		require.False(t, ib.IsActive())
		require.Equal(t, 0, ib.level)

		// Transition to State 1: Active (level 1)
		validTokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeWord, Value: "x"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
		}
		ib.BeginIfPossible(validTokens, 0)
		require.True(t, ib.IsActive())
		require.Equal(t, 1, ib.level)

		// Transition to State 2: Active (level 2)
		ib.BeginIfPossible(validTokens, 0)
		require.True(t, ib.IsActive())
		require.Equal(t, 2, ib.level)

		// Back to State 1
		ib.End()
		require.True(t, ib.IsActive())
		require.Equal(t, 1, ib.level)

		// Back to State 0
		ib.End()
		require.False(t, ib.IsActive())
		require.Equal(t, 0, ib.level)
	})

	t.Run("invalid block doesn't change state", func(t *testing.T) {
		ib := NewInlineBlock()

		invalidTokens := []types.Token{
			{Type: types.TokenTypeOpenParen, Value: "("},
			{Type: types.TokenTypeReservedTopLevel, Value: "SELECT"},
			{Type: types.TokenTypeCloseParen, Value: ")"},
		}

		ib.BeginIfPossible(invalidTokens, 0)
		require.False(t, ib.IsActive())
		require.Equal(t, 0, ib.level)
	})
}
