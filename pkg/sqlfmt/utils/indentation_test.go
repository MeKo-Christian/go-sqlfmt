package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewIndentation tests the NewIndentation constructor.
func TestNewIndentation(t *testing.T) {
	tests := []struct {
		name   string
		indent string
	}{
		{
			name:   "two spaces",
			indent: "  ",
		},
		{
			name:   "four spaces",
			indent: "    ",
		},
		{
			name:   "single tab",
			indent: "\t",
		},
		{
			name:   "single space",
			indent: " ",
		},
		{
			name:   "eight spaces",
			indent: "        ",
		},
		{
			name:   "empty string",
			indent: "",
		},
		{
			name:   "custom string",
			indent: "-->",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ind := NewIndentation(tt.indent)
			require.NotNil(t, ind)
			require.Empty(t, ind.GetIndent(), "New indentation should start at zero level")
		})
	}
}

// TestGetIndent tests the GetIndent method.
func TestGetIndent(t *testing.T) {
	tests := []struct {
		name     string
		indent   string
		setup    func(*Indentation)
		expected string
	}{
		{
			name:   "no indentation",
			indent: "  ",
			setup: func(i *Indentation) {
				// No setup needed
			},
			expected: "",
		},
		{
			name:   "single top-level indent",
			indent: "  ",
			setup: func(i *Indentation) {
				i.IncreaseTopLevel()
			},
			expected: "  ",
		},
		{
			name:   "double top-level indent",
			indent: "  ",
			setup: func(i *Indentation) {
				i.IncreaseTopLevel()
				i.IncreaseTopLevel()
			},
			expected: "    ",
		},
		{
			name:   "single block-level indent",
			indent: "  ",
			setup: func(i *Indentation) {
				i.IncreaseBlockLevel()
			},
			expected: "  ",
		},
		{
			name:   "mixed top and block level",
			indent: "  ",
			setup: func(i *Indentation) {
				i.IncreaseBlockLevel()
				i.IncreaseTopLevel()
			},
			expected: "    ",
		},
		{
			name:   "tab indentation",
			indent: "\t",
			setup: func(i *Indentation) {
				i.IncreaseTopLevel()
				i.IncreaseTopLevel()
			},
			expected: "\t\t",
		},
		{
			name:   "four space indentation",
			indent: "    ",
			setup: func(i *Indentation) {
				i.IncreaseTopLevel()
			},
			expected: "    ",
		},
		{
			name:   "deeply nested",
			indent: "  ",
			setup: func(i *Indentation) {
				i.IncreaseBlockLevel()
				i.IncreaseTopLevel()
				i.IncreaseTopLevel()
				i.IncreaseBlockLevel()
				i.IncreaseTopLevel()
			},
			expected: "          ", // 5 levels * 2 spaces
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ind := NewIndentation(tt.indent)
			tt.setup(ind)
			result := ind.GetIndent()
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestIncreaseTopLevel tests the IncreaseTopLevel method.
func TestIncreaseTopLevel(t *testing.T) {
	t.Run("single increase", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())
	})

	t.Run("multiple increases", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "      ", ind.GetIndent())
	})

	t.Run("increase after block level", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "    ", ind.GetIndent())
	})
}

// TestIncreaseBlockLevel tests the IncreaseBlockLevel method.
func TestIncreaseBlockLevel(t *testing.T) {
	t.Run("single increase", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		require.Equal(t, "  ", ind.GetIndent())
	})

	t.Run("multiple increases", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseBlockLevel()
		require.Equal(t, "      ", ind.GetIndent())
	})

	t.Run("increase after top level", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		ind.IncreaseBlockLevel()
		require.Equal(t, "    ", ind.GetIndent())
	})
}

// TestDecreaseTopLevel tests the DecreaseTopLevel method.
func TestDecreaseTopLevel(t *testing.T) {
	t.Run("decrease single top-level indent", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())

		ind.DecreaseTopLevel()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("decrease multiple top-level indents", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "      ", ind.GetIndent())

		ind.DecreaseTopLevel()
		require.Equal(t, "    ", ind.GetIndent())

		ind.DecreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())

		ind.DecreaseTopLevel()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("decrease when last indent is block-level does nothing", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		require.Equal(t, "  ", ind.GetIndent())

		ind.DecreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent(), "Should not decrease block-level indent")
	})

	t.Run("decrease when empty does nothing", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.DecreaseTopLevel()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("decrease top-level within block-level", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "    ", ind.GetIndent())

		ind.DecreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())
	})
}

// TestDecreaseBlockLevel tests the DecreaseBlockLevel method.
func TestDecreaseBlockLevel(t *testing.T) {
	t.Run("decrease single block-level indent", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		require.Equal(t, "  ", ind.GetIndent())

		ind.DecreaseBlockLevel()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("decrease block-level removes all nested top-levels", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "      ", ind.GetIndent())

		// DecreaseBlockLevel should remove all top-levels AND the block-level
		ind.DecreaseBlockLevel()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("decrease block-level stops at block-level", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "        ", ind.GetIndent()) // 4 levels

		// Should remove the top-level and the second block-level
		ind.DecreaseBlockLevel()
		require.Equal(t, "    ", ind.GetIndent()) // 2 levels remain
	})

	t.Run("decrease when empty does nothing", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.DecreaseBlockLevel()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("complex nesting scenario", func(t *testing.T) {
		ind := NewIndentation("  ")
		// Build: block -> top -> top -> block -> top
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "          ", ind.GetIndent()) // 5 levels

		// Decrease block should remove: top, block
		ind.DecreaseBlockLevel()
		require.Equal(t, "      ", ind.GetIndent()) // 3 levels: block, top, top

		// Decrease block again should remove: top, top, block
		ind.DecreaseBlockLevel()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("multiple blocks in sequence", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseBlockLevel()
		require.Equal(t, "      ", ind.GetIndent())

		ind.DecreaseBlockLevel()
		require.Equal(t, "    ", ind.GetIndent())

		ind.DecreaseBlockLevel()
		require.Equal(t, "  ", ind.GetIndent())

		ind.DecreaseBlockLevel()
		require.Empty(t, ind.GetIndent())
	})
}

// TestResetIndentation tests the ResetIndentation method.
func TestResetIndentation(t *testing.T) {
	t.Run("reset from no indentation", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.ResetIndentation()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("reset from single level", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())

		ind.ResetIndentation()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("reset from multiple levels", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "        ", ind.GetIndent())

		ind.ResetIndentation()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("can increase after reset", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()

		ind.ResetIndentation()
		require.Empty(t, ind.GetIndent())

		ind.IncreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())
	})

	t.Run("multiple resets", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		ind.ResetIndentation()
		ind.ResetIndentation()
		ind.ResetIndentation()
		require.Empty(t, ind.GetIndent())
	})
}

// TestIndentationComplexScenarios tests complex real-world indentation patterns.
func TestIndentationComplexScenarios(t *testing.T) {
	t.Run("SQL subquery pattern", func(t *testing.T) {
		ind := NewIndentation("  ")

		// SELECT (top-level)
		ind.IncreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())

		// Subquery in FROM clause (block-level)
		ind.IncreaseBlockLevel()
		require.Equal(t, "    ", ind.GetIndent())

		// SELECT in subquery (top-level within block)
		ind.IncreaseTopLevel()
		require.Equal(t, "      ", ind.GetIndent())

		// End of subquery SELECT
		ind.DecreaseTopLevel()
		require.Equal(t, "    ", ind.GetIndent())

		// End of subquery block
		ind.DecreaseBlockLevel()
		require.Equal(t, "  ", ind.GetIndent())

		// End of main SELECT
		ind.DecreaseTopLevel()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("nested CTEs", func(t *testing.T) {
		ind := NewIndentation("  ")

		// WITH clause
		ind.IncreaseBlockLevel()
		require.Equal(t, "  ", ind.GetIndent())

		// First CTE SELECT
		ind.IncreaseTopLevel()
		require.Equal(t, "    ", ind.GetIndent())

		// Nested subquery in CTE
		ind.IncreaseBlockLevel()
		require.Equal(t, "      ", ind.GetIndent())

		ind.IncreaseTopLevel()
		require.Equal(t, "        ", ind.GetIndent())

		// Close nested subquery
		ind.DecreaseBlockLevel()
		require.Equal(t, "    ", ind.GetIndent())

		// Close first CTE
		ind.DecreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())

		// Second CTE SELECT
		ind.IncreaseTopLevel()
		require.Equal(t, "    ", ind.GetIndent())

		ind.DecreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())

		// Close WITH block
		ind.DecreaseBlockLevel()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("CASE statement within SELECT", func(t *testing.T) {
		ind := NewIndentation("  ")

		// Main SELECT
		ind.IncreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())

		// CASE block
		ind.IncreaseBlockLevel()
		require.Equal(t, "    ", ind.GetIndent())

		// WHEN clauses (top-level within CASE)
		ind.IncreaseTopLevel()
		require.Equal(t, "      ", ind.GetIndent())

		ind.DecreaseTopLevel()
		require.Equal(t, "    ", ind.GetIndent())

		// End CASE
		ind.DecreaseBlockLevel()
		require.Equal(t, "  ", ind.GetIndent())

		// End SELECT
		ind.DecreaseTopLevel()
		require.Empty(t, ind.GetIndent())
	})
}

// TestIndentationWithDifferentIndentStrings tests various indent string configurations.
func TestIndentationWithDifferentIndentStrings(t *testing.T) {
	tests := []struct {
		name         string
		indentString string
		levels       int
		expected     string
	}{
		{
			name:         "two spaces - 3 levels",
			indentString: "  ",
			levels:       3,
			expected:     "      ",
		},
		{
			name:         "four spaces - 2 levels",
			indentString: "    ",
			levels:       2,
			expected:     "        ",
		},
		{
			name:         "tab - 4 levels",
			indentString: "\t",
			levels:       4,
			expected:     "\t\t\t\t",
		},
		{
			name:         "single space - 5 levels",
			indentString: " ",
			levels:       5,
			expected:     "     ",
		},
		{
			name:         "custom arrow - 2 levels",
			indentString: "->",
			levels:       2,
			expected:     "->->",
		},
		{
			name:         "empty indent - 3 levels",
			indentString: "",
			levels:       3,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ind := NewIndentation(tt.indentString)
			for range tt.levels {
				ind.IncreaseTopLevel()
			}
			require.Equal(t, tt.expected, ind.GetIndent())
		})
	}
}

// TestIndentationEdgeCases tests edge cases and boundary conditions.
func TestIndentationEdgeCases(t *testing.T) {
	t.Run("many decrease operations on empty", func(t *testing.T) {
		ind := NewIndentation("  ")
		for range 10 {
			ind.DecreaseTopLevel()
			ind.DecreaseBlockLevel()
		}
		require.Empty(t, ind.GetIndent())
	})

	t.Run("alternate increase and decrease", func(t *testing.T) {
		ind := NewIndentation("  ")

		ind.IncreaseTopLevel()
		require.Equal(t, "  ", ind.GetIndent())

		ind.DecreaseTopLevel()
		require.Empty(t, ind.GetIndent())

		ind.IncreaseBlockLevel()
		require.Equal(t, "  ", ind.GetIndent())

		ind.DecreaseBlockLevel()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("very deep nesting", func(t *testing.T) {
		ind := NewIndentation("  ")
		depth := 20

		for i := range depth {
			if i%2 == 0 {
				ind.IncreaseBlockLevel()
			} else {
				ind.IncreaseTopLevel()
			}
		}

		expectedIndent := ""
		for range depth {
			expectedIndent += "  "
		}
		require.Equal(t, expectedIndent, ind.GetIndent())
	})

	t.Run("decrease top-level when only block-levels exist", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseBlockLevel()
		require.Equal(t, "      ", ind.GetIndent())

		// Try to decrease top-level - should do nothing
		ind.DecreaseTopLevel()
		ind.DecreaseTopLevel()
		require.Equal(t, "      ", ind.GetIndent())
	})

	t.Run("reset then continue operations", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()

		ind.ResetIndentation()

		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "    ", ind.GetIndent())

		ind.DecreaseBlockLevel()
		require.Empty(t, ind.GetIndent())
	})
}

// TestIndentationInternalState tests internal state consistency.
func TestIndentationInternalState(t *testing.T) {
	t.Run("lastIndentType correctness", func(t *testing.T) {
		ind := NewIndentation("  ")

		// Empty state
		require.Equal(t, indentTypeNone, ind.lastIndentType())

		// Top-level
		ind.IncreaseTopLevel()
		require.Equal(t, indentTypeTopLevel, ind.lastIndentType())

		// Block-level
		ind.IncreaseBlockLevel()
		require.Equal(t, indentTypeBlockLevel, ind.lastIndentType())

		// Another top-level
		ind.IncreaseTopLevel()
		require.Equal(t, indentTypeTopLevel, ind.lastIndentType())
	})

	t.Run("popIndentType correctness", func(t *testing.T) {
		ind := NewIndentation("  ")

		// Pop from empty
		require.Equal(t, indentTypeNone, ind.popIndentType())

		// Add some indents
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()

		// Pop top-level
		require.Equal(t, indentTypeTopLevel, ind.popIndentType())
		require.Equal(t, "    ", ind.GetIndent())

		// Pop another top-level
		require.Equal(t, indentTypeTopLevel, ind.popIndentType())
		require.Equal(t, "  ", ind.GetIndent())

		// Pop block-level
		require.Equal(t, indentTypeBlockLevel, ind.popIndentType())
		require.Empty(t, ind.GetIndent())

		// Pop from empty again
		require.Equal(t, indentTypeNone, ind.popIndentType())
	})
}
