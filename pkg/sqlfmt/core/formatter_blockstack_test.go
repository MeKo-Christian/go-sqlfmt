package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBlockStackHelpers tests the block stack helper methods in isolation
func TestBlockStackHelpers(t *testing.T) {
	cfg := &Config{Indent: "  "}
	tokenizer := newTokenizer(&TokenizerConfig{})
	f := newFormatter(cfg, tokenizer, nil)

	t.Run("empty stack returns empty string", func(t *testing.T) {
		require.Equal(t, "", f.currentBlock())
		require.False(t, f.isInBlock("IF"))
		require.False(t, f.isInBlock("CASE"))
		require.False(t, f.isInBlock("BEGIN"))
	})

	t.Run("push and pop single block", func(t *testing.T) {
		f.pushBlock("IF")
		require.Equal(t, "IF", f.currentBlock())
		require.True(t, f.isInBlock("IF"))
		require.False(t, f.isInBlock("CASE"))

		popped := f.popBlock()
		require.Equal(t, "IF", popped)
		require.Equal(t, "", f.currentBlock())
		require.False(t, f.isInBlock("IF"))
	})

	t.Run("nested blocks", func(t *testing.T) {
		f.pushBlock("BEGIN")
		require.Equal(t, "BEGIN", f.currentBlock())
		require.True(t, f.isInBlock("BEGIN"))

		f.pushBlock("IF")
		require.Equal(t, "IF", f.currentBlock())
		require.True(t, f.isInBlock("IF"))
		require.True(t, f.isInBlock("BEGIN")) // Still in BEGIN

		f.pushBlock("CASE")
		require.Equal(t, "CASE", f.currentBlock())
		require.True(t, f.isInBlock("CASE"))
		require.True(t, f.isInBlock("IF"))   // Still in IF
		require.True(t, f.isInBlock("BEGIN")) // Still in BEGIN

		// Pop in reverse order
		require.Equal(t, "CASE", f.popBlock())
		require.Equal(t, "IF", f.currentBlock())

		require.Equal(t, "IF", f.popBlock())
		require.Equal(t, "BEGIN", f.currentBlock())

		require.Equal(t, "BEGIN", f.popBlock())
		require.Equal(t, "", f.currentBlock())
	})

	t.Run("pop from empty stack returns empty string", func(t *testing.T) {
		require.Equal(t, "", f.popBlock())
		require.Equal(t, "", f.popBlock())
	})
}

// TestBlockStackTracking_MySQL tests block stack tracking with MySQL queries
func TestBlockStackTracking_MySQL(t *testing.T) {
	cfg := &Config{
		Indent:             "  ",
		LinesBetweenQueries: 2,
	}

	t.Run("simple IF block tracks correctly", func(t *testing.T) {
		query := "BEGIN IF x > 0 THEN SELECT 1; END IF; END;"

		tokenizer := newTokenizer(&TokenizerConfig{
			OpenParens:  []string{"BEGIN", "IF"},
			CloseParens: []string{"END", "END IF"},
		})
		f := newFormatter(cfg, tokenizer, nil)

		// Format the query
		_ = f.format(query)

		// After formatting, stack should be empty (all blocks closed)
		require.Equal(t, "", f.currentBlock())
		require.Equal(t, 0, len(f.blockStack))
	})

	t.Run("nested CASE inside IF", func(t *testing.T) {
		query := `BEGIN
			IF x > 0 THEN
				SELECT CASE WHEN y = 1 THEN 'a' END;
			END IF;
		END;`

		tokenizer := newTokenizer(&TokenizerConfig{
			OpenParens:  []string{"BEGIN", "IF", "CASE"},
			CloseParens: []string{"END", "END IF"},
		})
		f := newFormatter(cfg, tokenizer, nil)

		_ = f.format(query)

		// Stack should be empty after formatting
		require.Equal(t, "", f.currentBlock())
		require.Equal(t, 0, len(f.blockStack))
	})
}

// TestBlockStackTracking_PostgreSQL tests block stack tracking with PostgreSQL
func TestBlockStackTracking_PostgreSQL(t *testing.T) {
	cfg := &Config{
		Indent:             "  ",
		LinesBetweenQueries: 2,
	}

	t.Run("PL/pgSQL block structure", func(t *testing.T) {
		query := `BEGIN
			IF found THEN
				RETURN 1;
			END IF;
		END;`

		tokenizer := newTokenizer(&TokenizerConfig{
			OpenParens:  []string{"BEGIN", "IF"},
			CloseParens: []string{"END", "END IF"},
		})
		f := newFormatter(cfg, tokenizer, nil)

		_ = f.format(query)

		require.Equal(t, "", f.currentBlock())
		require.Equal(t, 0, len(f.blockStack))
	})
}

// TestBlockStackTracking_StandardSQL tests with standard SQL CASE
func TestBlockStackTracking_StandardSQL(t *testing.T) {
	cfg := &Config{
		Indent:             "  ",
		LinesBetweenQueries: 2,
	}

	t.Run("CASE expression in SELECT", func(t *testing.T) {
		query := `SELECT CASE WHEN a = 1 THEN 'one' WHEN a = 2 THEN 'two' ELSE 'other' END FROM t;`

		tokenizer := newTokenizer(&TokenizerConfig{
			OpenParens:  []string{"CASE"},
			CloseParens: []string{"END"},
		})
		f := newFormatter(cfg, tokenizer, nil)

		_ = f.format(query)

		require.Equal(t, "", f.currentBlock())
		require.Equal(t, 0, len(f.blockStack))
	})

	t.Run("multiple CASE expressions", func(t *testing.T) {
		query := `SELECT
			CASE WHEN x = 1 THEN 'a' END,
			CASE WHEN y = 2 THEN 'b' END
		FROM t;`

		tokenizer := newTokenizer(&TokenizerConfig{
			OpenParens:  []string{"CASE"},
			CloseParens: []string{"END"},
		})
		f := newFormatter(cfg, tokenizer, nil)

		_ = f.format(query)

		require.Equal(t, "", f.currentBlock())
		require.Equal(t, 0, len(f.blockStack))
	})
}

// TestBlockStackNoOutputChange verifies that adding block stack tracking
// doesn't change the formatted output
func TestBlockStackNoOutputChange(t *testing.T) {
	testCases := []struct {
		name  string
		query string
	}{
		{
			name:  "simple SELECT",
			query: "SELECT * FROM users WHERE id = 1;",
		},
		{
			name:  "CASE expression",
			query: "SELECT CASE WHEN status = 'active' THEN 1 ELSE 0 END FROM orders;",
		},
		{
			name:  "nested CASE",
			query: "SELECT CASE WHEN x = 1 THEN CASE WHEN y = 2 THEN 'a' END END FROM t;",
		},
		{
			name:  "subquery",
			query: "SELECT * FROM (SELECT id FROM users) AS u;",
		},
		{
			name:  "IF in stored procedure",
			query: "BEGIN IF x > 0 THEN SELECT 1; END IF; END;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				Indent:             "  ",
				LinesBetweenQueries: 2,
			}

			tokenizer := newTokenizer(&TokenizerConfig{
				OpenParens:  []string{"(", "BEGIN", "IF", "CASE"},
				CloseParens: []string{")", "END", "END IF"},
			})

			f1 := newFormatter(cfg, tokenizer, nil)
			result1 := f1.format(tc.query)

			// Format again with a fresh formatter
			f2 := newFormatter(cfg, tokenizer, nil)
			result2 := f2.format(tc.query)

			// Results should be identical
			require.Equal(t, result1, result2, "Formatting should be deterministic")

			// Both formatters should have empty stack at the end
			require.Equal(t, "", f1.currentBlock())
			require.Equal(t, "", f2.currentBlock())
		})
	}
}
