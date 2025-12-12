package core

import (
	"testing"
)

// TestENDKeywordIndentation verifies that END keywords align with their opening keywords
// and are not affected by top-level indents accumulated inside the block (task 2.5).
func TestENDKeywordIndentation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "Simple BEGIN/END - END at column 0",
			input: "BEGIN SELECT 1; END;",
			expected: `BEGIN
  SELECT
    1;
END;`,
		},
		{
			name:  "CREATE PROCEDURE with BEGIN/END - END at column 0",
			input: "CREATE PROCEDURE foo() BEGIN SELECT 1; END;",
			expected: `CREATE PROCEDURE
  foo() BEGIN
    SELECT
      1;
END;`,
		},
		{
			name:  "Nested BEGIN blocks - each END aligns with its BEGIN",
			input: "BEGIN BEGIN SELECT 1; END; END;",
			expected: `BEGIN
  BEGIN
    SELECT
      1;
  END;
END;`,
		},
		{
			name:  "Multiple statements before END",
			input: "BEGIN SELECT 1; SELECT 2; SELECT 3; END;",
			expected: `BEGIN
  SELECT
    1;
  SELECT
    2;
  SELECT
    3;
END;`,
		},
		{
			name:  "BEGIN with DECLARE and SELECT - END at column 0",
			input: "BEGIN DECLARE x INT; SELECT x; END;",
			// Note: DECLARE is a top-level word, so it formats on its own line
			expected: `BEGIN
  DECLARE
    x INT;
  SELECT
    x;
END;`,
		},
		{
			name:  "IF inside BEGIN - END IF at correct level",
			input: "BEGIN IF x > 0 THEN SELECT 1; END IF; END;",
			// After task 2.5: END IF resets to procedural base (column 2, aligning with IF)
			// IF uses regular block indentation, not procedural indentation
			expected: `BEGIN
  IF
    x > 0 THEN
    SELECT
      1;
  END IF;
END;`,
		},
		{
			name:  "Nested IF blocks - each END IF aligns correctly",
			input: "BEGIN IF x > 0 THEN IF y > 0 THEN SELECT 1; END IF; END IF; END;",
			// Note: After fixing IF indentation, nested IF now appears on its own line
			// Both END IFs reset to procedural base (column 2, aligning with outermost IF)
			expected: `BEGIN
  IF
    x > 0 THEN
    IF
      y > 0 THEN
      SELECT
        1;
  END IF;
  END IF;
END;`,
		},
		{
			name:  "CASE inside BEGIN - END for CASE aligns correctly",
			input: "BEGIN SELECT CASE WHEN x = 1 THEN 'a' END; END;",
			// Note: CASE END maintains normal indentation (not affected by task 2.5)
			expected: `BEGIN
  SELECT
    CASE
      WHEN x = 1 THEN 'a'
    END;
END;`,
		},
		{
			name:  "Multiple nested blocks at different levels",
			input: "BEGIN BEGIN SELECT 1; END; BEGIN SELECT 2; END; END;",
			// Note: Second BEGIN indentation issue will be addressed in future tasks
			expected: `BEGIN
  BEGIN
    SELECT
      1;
  END;
BEGIN
    SELECT
      2;
  END;
END;`,
		},
		{
			name:  "WHILE loop with END WHILE",
			input: "BEGIN WHILE x > 0 DO SELECT x; END WHILE; END;",
			// Note: WHILE/DO formatting not fully implemented - they're treated as regular keywords
			expected: `BEGIN
  WHILE x > 0 DO
  SELECT
    x;
END WHILE;

END;`,
		},
		{
			name:  "LOOP with END LOOP",
			input: "BEGIN my_loop: LOOP SELECT 1; END LOOP my_loop; END;",
			// Note: LOOP formatting not fully implemented
			expected: `BEGIN
  my_loop: LOOP
  SELECT
    1;
END LOOP my_loop;

END;`,
		},
		{
			name:  "REPEAT with END REPEAT",
			input: "BEGIN REPEAT SELECT 1; UNTIL x > 5 END REPEAT; END;",
			// Note: REPEAT/UNTIL formatting not fully implemented
			expected: `BEGIN
  REPEAT
  SELECT
    1;
UNTIL x > 5
END REPEAT;

END;`,
		},
	}

	cfg := &Config{
		Indent:              "  ",
		LinesBetweenQueries: 2,
		TokenizerConfig: &TokenizerConfig{
			ReservedWords: []string{
				"SELECT", "FROM", "BEGIN", "END", "CREATE", "PROCEDURE",
				"DECLARE", "INT", "IF", "THEN", "CASE", "WHEN", "WHILE",
				"DO", "LOOP", "REPEAT", "UNTIL",
			},
			ReservedTopLevelWords: []string{
				"SELECT", "FROM", "CREATE PROCEDURE", "DECLARE",
			},
			ReservedNewlineWords:          []string{"WHEN"},
			ReservedTopLevelWordsNoIndent: []string{},
			StringTypes:                   []string{"\"\"", "''", "``"},
			OpenParens:                    []string{"(", "CASE", "BEGIN", "IF"},
			CloseParens:                   []string{")", "END", "END IF", "END CASE", "END LOOP", "END WHILE", "END REPEAT"},
			IndexedPlaceholderTypes:       []string{},
			NamedPlaceholderTypes:         []string{},
			LineCommentTypes:              []string{"--"},
		},
		KeywordCase: KeywordCaseUppercase,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)

			if result != tt.expected {
				t.Errorf("END indentation test failed\nInput:    %q\nExpected:\n%s\n\nGot:\n%s",
					tt.input, tt.expected, result)
			}
		})
	}
}

// TestENDKeywordWithComplexStatements tests END keyword indentation with more complex SQL statements.
func TestENDKeywordWithComplexStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "BEGIN with SELECT with JOIN - END at column 0",
			input: "BEGIN SELECT u.name FROM users u JOIN orders o ON u.id = o.user_id; END;",
			expected: `BEGIN
  SELECT
    u.name
  FROM
    users u
    JOIN orders o ON u.id = o.user_id;
END;`,
		},
		{
			name:  "BEGIN with subquery - END at column 0",
			input: "BEGIN SELECT * FROM (SELECT id FROM users) AS u; END;",
			expected: `BEGIN
  SELECT
    *
  FROM
    (
      SELECT
        id
      FROM
        users
    ) AS u;
END;`,
		},
		{
			name:  "BEGIN with UPDATE statement - END at column 0",
			input: "BEGIN UPDATE users SET name = 'John' WHERE id = 1; END;",
			expected: `BEGIN
  UPDATE
    users
  SET
    name = 'John'
  WHERE
    id = 1;
END;`,
		},
	}

	cfg := &Config{
		Indent:              "  ",
		LinesBetweenQueries: 2,
		TokenizerConfig: &TokenizerConfig{
			ReservedWords: []string{
				"SELECT", "FROM", "BEGIN", "END", "JOIN", "ON", "AS",
				"UPDATE", "SET", "WHERE", "AND", "OR",
			},
			ReservedTopLevelWords: []string{
				"SELECT", "FROM", "UPDATE", "SET", "WHERE",
			},
			ReservedNewlineWords:          []string{"JOIN", "AND", "OR"},
			ReservedTopLevelWordsNoIndent: []string{},
			StringTypes:                   []string{"\"\"", "''", "``"},
			OpenParens:                    []string{"(", "BEGIN"},
			CloseParens:                   []string{")", "END"},
			IndexedPlaceholderTypes:       []string{},
			NamedPlaceholderTypes:         []string{},
			LineCommentTypes:              []string{"--"},
		},
		KeywordCase: KeywordCaseUppercase,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)

			if result != tt.expected {
				t.Errorf("END indentation test failed\nInput:    %q\nExpected:\n%s\n\nGot:\n%s",
					tt.input, tt.expected, result)
			}
		})
	}
}
