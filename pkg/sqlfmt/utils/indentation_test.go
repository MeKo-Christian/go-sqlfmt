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

	t.Run("popIndentEntry correctness", func(t *testing.T) {
		ind := NewIndentation("  ")

		// Pop from empty
		entry := ind.popIndentEntry()
		require.Equal(t, indentTypeNone, entry.Type)

		// Add some indents
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()

		// Pop top-level
		entry = ind.popIndentEntry()
		require.Equal(t, indentTypeTopLevel, entry.Type)
		require.Equal(t, "    ", ind.GetIndent())

		// Pop another top-level
		entry = ind.popIndentEntry()
		require.Equal(t, indentTypeTopLevel, entry.Type)
		require.Equal(t, "  ", ind.GetIndent())

		// Pop block-level
		entry = ind.popIndentEntry()
		require.Equal(t, indentTypeBlockLevel, entry.Type)
		require.Empty(t, ind.GetIndent())

		// Pop from empty again
		entry = ind.popIndentEntry()
		require.Equal(t, indentTypeNone, entry.Type)
	})
}

// TestIncreaseProcedural tests the IncreaseProcedural method.
func TestIncreaseProcedural(t *testing.T) {
	t.Run("single procedural indent", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		require.Equal(t, "  ", ind.GetIndent())
		require.Equal(t, 1, ind.GetProceduralDepth())
	})

	t.Run("multiple procedural indents", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseProcedural("IF")
		ind.IncreaseProcedural("LOOP")
		require.Equal(t, "      ", ind.GetIndent())
		require.Equal(t, 3, ind.GetProceduralDepth())
	})

	t.Run("procedural indent tracks keyword", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		require.Equal(t, 1, len(ind.indentStack))
		require.Equal(t, "BEGIN", ind.indentStack[0].Keyword)
		require.Equal(t, indentTypeProcedural, ind.indentStack[0].Type)
		require.Equal(t, indentSourceProcedural, ind.indentStack[0].Source)
	})

	t.Run("procedural mixed with other indents", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseTopLevel()
		ind.IncreaseProcedural("IF")
		require.Equal(t, "      ", ind.GetIndent())
		require.Equal(t, 2, ind.GetProceduralDepth())
	})
}

// TestDecreaseProcedural tests the DecreaseProcedural method.
func TestDecreaseProcedural(t *testing.T) {
	t.Run("decrease single procedural indent", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		require.Equal(t, "  ", ind.GetIndent())

		ind.DecreaseProcedural()
		require.Empty(t, ind.GetIndent())
		require.Equal(t, 0, ind.GetProceduralDepth())
	})

	t.Run("decrease procedural removes only procedural", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "      ", ind.GetIndent())
		require.Equal(t, 1, ind.GetProceduralDepth())

		// Decrease procedural should only remove the BEGIN, leaving top-levels
		ind.DecreaseProcedural()
		require.Equal(t, "    ", ind.GetIndent())
		require.Equal(t, 0, ind.GetProceduralDepth())
	})

	t.Run("decrease procedural removes most recent procedural", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseTopLevel()
		ind.IncreaseProcedural("IF")
		ind.IncreaseTopLevel()
		require.Equal(t, "        ", ind.GetIndent()) // 4 levels
		require.Equal(t, 2, ind.GetProceduralDepth())

		// Should remove the IF (most recent procedural from end)
		ind.DecreaseProcedural()
		require.Equal(t, "      ", ind.GetIndent()) // 3 levels: BEGIN, top, top
		require.Equal(t, 1, ind.GetProceduralDepth())
	})

	t.Run("decrease procedural when empty does nothing", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.DecreaseProcedural()
		require.Empty(t, ind.GetIndent())
		require.Equal(t, 0, ind.GetProceduralDepth())
	})

	t.Run("decrease procedural when no procedural indents", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		ind.IncreaseBlockLevel()
		require.Equal(t, "    ", ind.GetIndent())

		ind.DecreaseProcedural()
		require.Equal(t, "    ", ind.GetIndent(), "Should not affect non-procedural indents")
	})
}

// TestGetProceduralDepth tests the GetProceduralDepth method.
func TestGetProceduralDepth(t *testing.T) {
	t.Run("zero when empty", func(t *testing.T) {
		ind := NewIndentation("  ")
		require.Equal(t, 0, ind.GetProceduralDepth())
	})

	t.Run("zero when only non-procedural indents", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, 0, ind.GetProceduralDepth())
	})

	t.Run("counts only procedural indents", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseTopLevel()
		ind.IncreaseProcedural("IF")
		ind.IncreaseBlockLevel()
		ind.IncreaseProcedural("LOOP")
		ind.IncreaseTopLevel()

		require.Equal(t, 3, ind.GetProceduralDepth())
	})

	t.Run("updates correctly after decrease", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseProcedural("IF")
		ind.IncreaseProcedural("LOOP")
		require.Equal(t, 3, ind.GetProceduralDepth())

		ind.DecreaseProcedural()
		require.Equal(t, 2, ind.GetProceduralDepth())

		ind.DecreaseProcedural()
		require.Equal(t, 1, ind.GetProceduralDepth())

		ind.DecreaseProcedural()
		require.Equal(t, 0, ind.GetProceduralDepth())
	})
}

// TestResetToProceduralBase tests the ResetToProceduralBase method.
func TestResetToProceduralBase(t *testing.T) {
	t.Run("reset from empty does nothing", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.ResetToProceduralBase()
		require.Empty(t, ind.GetIndent())
		require.Equal(t, 0, ind.GetProceduralDepth())
	})

	t.Run("reset removes only top-level and block-level indents", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseTopLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "        ", ind.GetIndent()) // 4 levels

		ind.ResetToProceduralBase()
		require.Equal(t, "  ", ind.GetIndent()) // Only BEGIN remains
		require.Equal(t, 1, ind.GetProceduralDepth())
	})

	t.Run("reset preserves all procedural indents", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseTopLevel()
		ind.IncreaseProcedural("IF")
		ind.IncreaseTopLevel()
		ind.IncreaseProcedural("LOOP")
		ind.IncreaseTopLevel()
		require.Equal(t, "            ", ind.GetIndent()) // 6 levels
		require.Equal(t, 3, ind.GetProceduralDepth())

		ind.ResetToProceduralBase()
		require.Equal(t, "      ", ind.GetIndent()) // 3 procedural indents remain
		require.Equal(t, 3, ind.GetProceduralDepth())
	})

	t.Run("reset when only procedural indents preserves all", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseProcedural("IF")
		require.Equal(t, "    ", ind.GetIndent())

		ind.ResetToProceduralBase()
		require.Equal(t, "    ", ind.GetIndent())
		require.Equal(t, 2, ind.GetProceduralDepth())
	})

	t.Run("reset when only non-procedural indents clears all", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseTopLevel()
		require.Equal(t, "      ", ind.GetIndent())

		ind.ResetToProceduralBase()
		require.Empty(t, ind.GetIndent())
		require.Equal(t, 0, ind.GetProceduralDepth())
	})

	t.Run("complex mixed scenario", func(t *testing.T) {
		ind := NewIndentation("  ")
		// Build: BEGIN, top, block, IF, top, top, LOOP
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseTopLevel()
		ind.IncreaseBlockLevel()
		ind.IncreaseProcedural("IF")
		ind.IncreaseTopLevel()
		ind.IncreaseTopLevel()
		ind.IncreaseProcedural("LOOP")
		require.Equal(t, "              ", ind.GetIndent()) // 7 levels
		require.Equal(t, 3, ind.GetProceduralDepth())

		ind.ResetToProceduralBase()
		require.Equal(t, "      ", ind.GetIndent()) // Only BEGIN, IF, LOOP remain
		require.Equal(t, 3, ind.GetProceduralDepth())

		// Verify we can still add indents after reset
		ind.IncreaseTopLevel()
		require.Equal(t, "        ", ind.GetIndent())
	})
}

// TestProceduralIndentationComplexScenarios tests real-world stored procedure patterns.
func TestProceduralIndentationComplexScenarios(t *testing.T) {
	t.Run("stored procedure with IF statement", func(t *testing.T) {
		ind := NewIndentation("  ")

		// CREATE PROCEDURE
		ind.IncreaseProcedural("BEGIN")
		require.Equal(t, "  ", ind.GetIndent())

		// DECLARE statement
		ind.IncreaseTopLevel()
		require.Equal(t, "    ", ind.GetIndent())

		// Semicolon resets to procedural base
		ind.ResetToProceduralBase()
		require.Equal(t, "  ", ind.GetIndent())

		// IF statement
		ind.IncreaseProcedural("IF")
		require.Equal(t, "    ", ind.GetIndent())

		// Statement inside IF
		ind.IncreaseTopLevel()
		require.Equal(t, "      ", ind.GetIndent())

		// Semicolon resets to procedural base (both BEGIN and IF)
		ind.ResetToProceduralBase()
		require.Equal(t, "    ", ind.GetIndent())

		// END IF
		ind.DecreaseProcedural()
		require.Equal(t, "  ", ind.GetIndent())
		require.Equal(t, 1, ind.GetProceduralDepth()) // Only BEGIN remains

		// END (procedure)
		ind.DecreaseProcedural()
		require.Empty(t, ind.GetIndent())
		require.Equal(t, 0, ind.GetProceduralDepth())
	})

	t.Run("nested procedural blocks", func(t *testing.T) {
		ind := NewIndentation("  ")

		// BEGIN (outer)
		ind.IncreaseProcedural("BEGIN")
		require.Equal(t, "  ", ind.GetIndent())

		// BEGIN (inner)
		ind.IncreaseProcedural("BEGIN")
		require.Equal(t, "    ", ind.GetIndent())

		// Statement
		ind.IncreaseTopLevel()
		require.Equal(t, "      ", ind.GetIndent())

		// Semicolon
		ind.ResetToProceduralBase()
		require.Equal(t, "    ", ind.GetIndent()) // Two BEGINs remain

		// END (inner)
		ind.DecreaseProcedural()
		require.Equal(t, "  ", ind.GetIndent())

		// END (outer)
		ind.DecreaseProcedural()
		require.Empty(t, ind.GetIndent())
	})

	t.Run("IF inside LOOP inside BEGIN", func(t *testing.T) {
		ind := NewIndentation("  ")

		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseProcedural("LOOP")
		ind.IncreaseProcedural("IF")
		require.Equal(t, "      ", ind.GetIndent())
		require.Equal(t, 3, ind.GetProceduralDepth())

		// Statement inside IF
		ind.IncreaseTopLevel()
		require.Equal(t, "        ", ind.GetIndent())

		// Semicolon
		ind.ResetToProceduralBase()
		require.Equal(t, "      ", ind.GetIndent()) // BEGIN, LOOP, IF

		// Close IF
		ind.DecreaseProcedural()
		require.Equal(t, "    ", ind.GetIndent()) // BEGIN, LOOP

		// Close LOOP
		ind.DecreaseProcedural()
		require.Equal(t, "  ", ind.GetIndent()) // BEGIN

		// Close BEGIN
		ind.DecreaseProcedural()
		require.Empty(t, ind.GetIndent())
	})
}

// TestIndentEntryMetadata tests that IndentEntry properly tracks metadata.
func TestIndentEntryMetadata(t *testing.T) {
	t.Run("top-level entry has correct metadata", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseTopLevel()

		require.Equal(t, 1, len(ind.indentStack))
		entry := ind.indentStack[0]
		require.Equal(t, indentTypeTopLevel, entry.Type)
		require.Equal(t, indentSourceTopLevel, entry.Source)
		require.Empty(t, entry.Keyword)
	})

	t.Run("block-level entry has correct metadata", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseBlockLevel()

		require.Equal(t, 1, len(ind.indentStack))
		entry := ind.indentStack[0]
		require.Equal(t, indentTypeBlockLevel, entry.Type)
		require.Equal(t, indentSourceBlock, entry.Source)
		require.Empty(t, entry.Keyword)
	})

	t.Run("procedural entry has correct metadata", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("WHILE")

		require.Equal(t, 1, len(ind.indentStack))
		entry := ind.indentStack[0]
		require.Equal(t, indentTypeProcedural, entry.Type)
		require.Equal(t, indentSourceProcedural, entry.Source)
		require.Equal(t, "WHILE", entry.Keyword)
	})

	t.Run("mixed entries preserve metadata", func(t *testing.T) {
		ind := NewIndentation("  ")
		ind.IncreaseProcedural("BEGIN")
		ind.IncreaseTopLevel()
		ind.IncreaseProcedural("IF")
		ind.IncreaseBlockLevel()

		require.Equal(t, 4, len(ind.indentStack))

		require.Equal(t, indentTypeProcedural, ind.indentStack[0].Type)
		require.Equal(t, "BEGIN", ind.indentStack[0].Keyword)

		require.Equal(t, indentTypeTopLevel, ind.indentStack[1].Type)
		require.Empty(t, ind.indentStack[1].Keyword)

		require.Equal(t, indentTypeProcedural, ind.indentStack[2].Type)
		require.Equal(t, "IF", ind.indentStack[2].Keyword)

		require.Equal(t, indentTypeBlockLevel, ind.indentStack[3].Type)
		require.Empty(t, ind.indentStack[3].Keyword)
	})
}
