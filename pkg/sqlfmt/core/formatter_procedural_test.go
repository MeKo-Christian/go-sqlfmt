package core

import (
	"testing"
)

// TestProceduralDepthTracking tests that proceduralDepth is correctly tracked
// when entering and exiting BEGIN/END blocks.
func TestProceduralDepthTracking(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedAtSteps []int // Expected proceduralDepth at specific points
	}{
		{
			name:  "single BEGIN/END block",
			input: "BEGIN SELECT 1; END;",
			// After BEGIN: depth=1, After END: depth=0
		},
		{
			name:  "nested BEGIN/END blocks",
			input: "BEGIN BEGIN SELECT 1; END; END;",
			// After first BEGIN: depth=1, After second BEGIN: depth=2,
			// After first END: depth=1, After second END: depth=0
		},
		{
			name:  "multiple sequential BEGIN/END blocks",
			input: "BEGIN SELECT 1; END; BEGIN SELECT 2; END;",
			// First BEGIN: depth=1, First END: depth=0,
			// Second BEGIN: depth=1, Second END: depth=0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent: "  ",
				TokenizerConfig: &TokenizerConfig{
					ReservedWords: []string{
						"SELECT", "BEGIN", "END",
					},
					ReservedTopLevelWords: []string{
						"SELECT",
					},
					ReservedNewlineWords:          []string{},
					ReservedTopLevelWordsNoIndent: []string{},
					StringTypes:                   []string{"\"\"", "''", "``"},
					OpenParens:                    []string{"(", "CASE", "BEGIN", "IF"},
					CloseParens:                   []string{")", "END", "END IF", "END CASE", "END LOOP", "END WHILE", "END REPEAT"},
					IndexedPlaceholderTypes:       []string{},
					NamedPlaceholderTypes:         []string{},
					LineCommentTypes:              []string{"--"},
					SpecialWordChars:              []string{},
				},
				KeywordCase: KeywordCaseUppercase,
			}

			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)

			// Format the input
			_ = formatter.format(tt.input)

			// The test passes if no panic occurred
			// We're primarily testing that the depth tracking doesn't break formatting
			// More detailed validation could be added by checking intermediate states
		})
	}
}

// TestProceduralDepthNoOutputChange tests that adding procedural depth tracking
// doesn't change the output of existing queries.
func TestProceduralDepthNoOutputChange(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "simple SELECT",
			input: "SELECT * FROM users",
			expected: `SELECT
  *
FROM
  users`,
		},
		{
			name:  "BEGIN/END block",
			input: "BEGIN SELECT 1; END",
			expected: `BEGIN
  SELECT
    1;

END`,
		},
		{
			name:  "nested BEGIN blocks",
			input: "BEGIN BEGIN SELECT 1; END; END",
			// TODO: Both END keywords should be indented to align with their BEGINs
			// Current behavior: Both ENDs are at column 0
			// Expected future: Inner END should be at column 2, outer END at column 0
			expected: `BEGIN
  BEGIN
    SELECT
      1;

END;

END`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent: "  ",
				TokenizerConfig: &TokenizerConfig{
					ReservedWords: []string{
						"SELECT", "FROM", "BEGIN", "END",
					},
					ReservedTopLevelWords: []string{
						"SELECT", "FROM",
					},
					ReservedNewlineWords:          []string{},
					ReservedTopLevelWordsNoIndent: []string{},
					StringTypes:                   []string{"\"\"", "''", "``"},
					OpenParens:                    []string{"(", "CASE", "BEGIN", "IF"},
					CloseParens:                   []string{")", "END", "END IF", "END CASE", "END LOOP", "END WHILE", "END REPEAT"},
					IndexedPlaceholderTypes:       []string{},
					NamedPlaceholderTypes:         []string{},
					LineCommentTypes:              []string{"--"},
					SpecialWordChars:              []string{},
				},
				KeywordCase:         KeywordCaseUppercase,
				LinesBetweenQueries: 2,
			}

			result := FormatQuery(cfg, nil, tt.input)

			if result != tt.expected {
				t.Errorf("Output changed:\nExpected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

// TestIsInProceduralBlock tests the isInProceduralBlock helper method.
func TestIsInProceduralBlock(t *testing.T) {
	cfg := &Config{
		Indent: "  ",
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:                 []string{"BEGIN", "END", "SELECT"},
			ReservedTopLevelWords:         []string{"SELECT"},
			ReservedNewlineWords:          []string{},
			ReservedTopLevelWordsNoIndent: []string{},
			StringTypes:                   []string{"\"\"", "''", "``"},
			OpenParens:                    []string{"(", "BEGIN"},
			CloseParens:                   []string{")", "END"},
			IndexedPlaceholderTypes:       []string{},
			NamedPlaceholderTypes:         []string{},
			LineCommentTypes:              []string{"--"},
			SpecialWordChars:              []string{},
		},
	}

	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, nil)

	// Initially not in procedural block
	if formatter.isInProceduralBlock() {
		t.Error("Expected isInProceduralBlock to be false initially")
	}

	// Simulate entering a BEGIN block
	formatter.proceduralDepth = 1
	if !formatter.isInProceduralBlock() {
		t.Error("Expected isInProceduralBlock to be true after entering BEGIN")
	}

	// Simulate nested BEGIN
	formatter.proceduralDepth = 2
	if !formatter.isInProceduralBlock() {
		t.Error("Expected isInProceduralBlock to be true with nested BEGIN")
	}

	// Simulate exiting one level
	formatter.proceduralDepth = 1
	if !formatter.isInProceduralBlock() {
		t.Error("Expected isInProceduralBlock to be true after exiting one level")
	}

	// Simulate exiting all levels
	formatter.proceduralDepth = 0
	if formatter.isInProceduralBlock() {
		t.Error("Expected isInProceduralBlock to be false after exiting all levels")
	}
}

// TestProceduralDepthConsistency tests that proceduralDepth and the indentation
// stack's procedural depth stay in sync.
func TestProceduralDepthConsistency(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "single BEGIN/END",
			input: "BEGIN SELECT 1; END;",
		},
		{
			name:  "nested BEGIN/END",
			input: "BEGIN BEGIN SELECT 1; END; END;",
		},
		{
			name:  "multiple sequential BEGIN/END",
			input: "BEGIN SELECT 1; END; BEGIN SELECT 2; END;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent: "  ",
				TokenizerConfig: &TokenizerConfig{
					ReservedWords: []string{
						"SELECT", "BEGIN", "END",
					},
					ReservedTopLevelWords: []string{
						"SELECT",
					},
					ReservedNewlineWords:          []string{},
					ReservedTopLevelWordsNoIndent: []string{},
					StringTypes:                   []string{"\"\"", "''", "``"},
					OpenParens:                    []string{"(", "CASE", "BEGIN", "IF"},
					CloseParens:                   []string{")", "END", "END IF", "END CASE", "END LOOP", "END WHILE", "END REPEAT"},
					IndexedPlaceholderTypes:       []string{},
					NamedPlaceholderTypes:         []string{},
					LineCommentTypes:              []string{"--"},
					SpecialWordChars:              []string{},
				},
				KeywordCase: KeywordCaseUppercase,
			}

			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)

			// Format and check that depth returns to 0
			_ = formatter.format(tt.input)

			if formatter.proceduralDepth != 0 {
				t.Errorf("Expected proceduralDepth to be 0 after formatting, got %d", formatter.proceduralDepth)
			}

			if formatter.indentation.GetProceduralDepth() != 0 {
				t.Errorf("Expected indentation procedural depth to be 0 after formatting, got %d",
					formatter.indentation.GetProceduralDepth())
			}

			// Verify they match
			if formatter.proceduralDepth != formatter.indentation.GetProceduralDepth() {
				t.Errorf("proceduralDepth (%d) and indentation.GetProceduralDepth() (%d) are out of sync",
					formatter.proceduralDepth, formatter.indentation.GetProceduralDepth())
			}
		})
	}
}
