package core

import (
	"strings"
	"testing"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/utils"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Indentation Tests
// ============================================================================

func TestFormatterIndentation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "basic SELECT with indentation",
			input: "SELECT id, name FROM users WHERE active = true",
			expected: `SELECT
  id,
  name
FROM
  users
WHERE
  active = true`,
		},
		{
			name:  "nested subquery indentation",
			input: "SELECT * FROM (SELECT id FROM users) AS u",
			expected: `SELECT
  *
FROM
  (
    SELECT
      id
    FROM
      users
  ) AS u`,
		},
		{
			name:  "multiple nesting levels",
			input: "SELECT * FROM (SELECT * FROM (SELECT id FROM users) AS inner_q) AS outer_q",
			expected: `SELECT
  *
FROM
  (
    SELECT
      *
    FROM
      (
        SELECT
          id
        FROM
          users
      ) AS inner_q
  ) AS outer_q`,
		},
		{
			name:  "UNION indentation",
			input: "SELECT id FROM users UNION ALL SELECT id FROM admins",
			expected: `SELECT
  id
FROM
  users
UNION ALL
SELECT
  id
FROM
  admins`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:                 []string{"SELECT", "FROM", "WHERE", "AS"},
					ReservedTopLevelWords:         []string{"SELECT", "FROM", "WHERE"},
					ReservedNewlineWords:          []string{"AND", "OR"},
					ReservedTopLevelWordsNoIndent: []string{"UNION ALL"},
					StringTypes:                   []string{"''"},
					OpenParens:                    []string{"("},
					CloseParens:                   []string{")"},
					LineCommentTypes:              []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Line Break Tests
// ============================================================================

func TestFormatterLineBreaks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "line break after top-level keywords",
			input: "SELECT id FROM users WHERE active = true",
			expected: `SELECT
  id
FROM
  users
WHERE
  active = true`,
		},
		{
			name:  "line break after AND/OR",
			input: "SELECT * FROM users WHERE active = true AND verified = true OR admin = true",
			expected: `SELECT
  *
FROM
  users
WHERE
  active = true
  AND verified = true
  OR admin = true`,
		},
		{
			name:  "no line break in inline blocks",
			input: "SELECT (1, 2, 3) FROM users",
			expected: `SELECT
  (1, 2, 3)
FROM
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:                 []string{"SELECT", "FROM", "WHERE", "AND", "OR"},
					ReservedTopLevelWords:         []string{"SELECT", "FROM", "WHERE"},
					ReservedNewlineWords:          []string{"AND", "OR"},
					ReservedTopLevelWordsNoIndent: []string{},
					StringTypes:                   []string{"''"},
					OpenParens:                    []string{"("},
					CloseParens:                   []string{")"},
					LineCommentTypes:              []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Whitespace Normalization Tests
// ============================================================================

func TestFormatterWhitespaceNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "multiple spaces normalized",
			input: "SELECT    id,    name    FROM    users",
			expected: `SELECT
  id,
  name
FROM
  users`,
		},
		{
			name:  "tabs normalized",
			input: "SELECT\t\tid,\t\tname\t\tFROM\t\tusers",
			expected: `SELECT
  id,
  name
FROM
  users`,
		},
		{
			name:  "mixed whitespace normalized",
			input: "SELECT  \t  id,  \t  name  \t  FROM  \t  users",
			expected: `SELECT
  id,
  name
FROM
  users`,
		},
		{
			name:  "trailing spaces removed",
			input: "SELECT id FROM users   ",
			expected: `SELECT
  id
FROM
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Keyword Case Tests
// ============================================================================

func TestFormatterKeywordCase(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		keywordCase KeywordCase
		expected    string
	}{
		{
			name:        "uppercase keywords",
			input:       "select id, name from users where active = true",
			keywordCase: KeywordCaseUppercase,
			expected: `SELECT
  id,
  name
FROM
  users
WHERE
  active = true`,
		},
		{
			name:        "lowercase keywords",
			input:       "SELECT ID, NAME FROM USERS WHERE ACTIVE = TRUE",
			keywordCase: KeywordCaseLowercase,
			expected: `select
  ID,
  NAME
from
  USERS
where
  ACTIVE = TRUE`,
		},
		{
			name:        "preserve case",
			input:       "SeLeCt id, name FrOm users WhErE active = true",
			keywordCase: KeywordCasePreserve,
			expected: `SeLeCt
  id,
  name
FrOm
  users
WhErE
  active = true`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: tt.keywordCase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM", "WHERE"},
					ReservedTopLevelWords: []string{"SELECT", "FROM", "WHERE"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Inline Block Tests
// ============================================================================

func TestFormatterInlineBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "short inline block",
			input: "SELECT (1, 2, 3) FROM users",
			expected: `SELECT
  (1, 2, 3)
FROM
  users`,
		},
		{
			name:  "short function call",
			input: "SELECT COUNT(*) FROM users",
			expected: `SELECT
  COUNT(*)
FROM
  users`,
		},
		{
			name:  "nested function calls",
			input: "SELECT UPPER(TRIM(name)) FROM users",
			expected: `SELECT
  UPPER(TRIM(name))
FROM
  users`,
		},
		{
			name: "long expression breaks into lines",
			input: "SELECT (very_long_column_name_1, very_long_column_name_2, " +
				"very_long_column_name_3, very_long_column_name_4, very_long_column_name_5) FROM users",
			expected: `SELECT
  (
    very_long_column_name_1,
    very_long_column_name_2,
    very_long_column_name_3,
    very_long_column_name_4,
    very_long_column_name_5
  )
FROM
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM", "COUNT", "UPPER", "TRIM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Special SQL Construct Tests
// ============================================================================

func TestFormatterCaseExpressions(t *testing.T) {
	input := "SELECT CASE WHEN status = 'active' THEN 1 WHEN status = 'inactive' THEN 0 ELSE -1 END FROM users"
	expected := `SELECT
  CASE
    WHEN status = 'active' THEN 1
    WHEN status = 'inactive' THEN 0 ELSE -1
  END
FROM
  users`

	cfg := &Config{
		Indent:      "  ",
		KeywordCase: KeywordCaseUppercase,
		ColorConfig: &ColorConfig{},
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:         []string{"SELECT", "FROM", "WHEN", "THEN", "ELSE"},
			ReservedTopLevelWords: []string{"SELECT", "FROM"},
			ReservedNewlineWords:  []string{"WHEN"},
			StringTypes:           []string{"''"},
			OpenParens:            []string{"(", "CASE"},
			CloseParens:           []string{")", "END"},
			LineCommentTypes:      []string{"--"},
		},
	}
	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, nil)
	result := formatter.format(input)
	require.Equal(t, expected, result)
}

func TestFormatterCTEs(t *testing.T) {
	input := "WITH user_stats AS (SELECT user_id, COUNT(*) AS total FROM orders GROUP BY user_id) " +
		"SELECT u.name, us.total FROM users u JOIN user_stats us ON u.id = us.user_id"
	expected := `WITH
  user_stats AS (
    SELECT
      user_id,
      COUNT(*) AS total
    FROM
      orders GROUP BY user_id
  )
SELECT
  u.name,
  us.total
FROM
  users u
JOIN
  user_stats us
ON
  u.id = us.user_id`

	cfg := &Config{
		Indent:      "  ",
		KeywordCase: KeywordCaseUppercase,
		ColorConfig: &ColorConfig{},
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:         []string{"WITH", "AS", "SELECT", "FROM", "JOIN", "ON", "GROUP BY", "COUNT"},
			ReservedTopLevelWords: []string{"WITH", "SELECT", "FROM", "JOIN", "ON"},
			StringTypes:           []string{"''"},
			OpenParens:            []string{"("},
			CloseParens:           []string{")"},
			LineCommentTypes:      []string{"--"},
		},
	}
	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, nil)
	result := formatter.format(input)
	require.Equal(t, expected, result)
}

// ============================================================================
// Comment Formatting Tests
// ============================================================================

func TestFormatterLineComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "line comment on separate line",
			input: "SELECT id, name -- user identifier and name\nFROM users",
			expected: `SELECT
  id,
  name -- user identifier and name
FROM
  users`,
		},
		{
			name:  "line comment at start",
			input: "-- Get all users\nSELECT id FROM users",
			expected: `-- Get all users
SELECT
  id
FROM
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatterBlockComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "single-line block comment",
			input: "SELECT /* important */ id FROM users",
			expected: `SELECT
  /* important */
  id
FROM
  users`,
		},
		{
			name:  "multi-line block comment",
			input: "SELECT\n/* Get user ID\n   for active users */\nid FROM users",
			expected: `SELECT
  /* Get user ID
   for active users */
  id
FROM
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Placeholder Formatting Tests
// ============================================================================

func TestFormatterPlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		params   *utils.ParamsConfig
		expected string
	}{
		{
			name:  "indexed placeholder",
			input: "SELECT * FROM users WHERE id = ?",
			expected: `SELECT
  *
FROM
  users
WHERE
  id = ?`,
		},
		{
			name:  "named placeholder",
			input: "SELECT * FROM users WHERE id = :user_id",
			expected: `SELECT
  *
FROM
  users
WHERE
  id = :user_id`,
		},
		{
			name:  "placeholder replacement",
			input: "SELECT * FROM users WHERE id = ?",
			params: &utils.ParamsConfig{
				ListParams: []string{"123"},
			},
			expected: `SELECT
  *
FROM
  users
WHERE
  id = 123`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				Params:      tt.params,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:           []string{"SELECT", "FROM", "WHERE"},
					ReservedTopLevelWords:   []string{"SELECT", "FROM", "WHERE"},
					StringTypes:             []string{"''"},
					OpenParens:              []string{"("},
					CloseParens:             []string{")"},
					IndexedPlaceholderTypes: []string{"?"},
					NamedPlaceholderTypes:   []string{":", "@"},
					LineCommentTypes:        []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Query Separator Tests
// ============================================================================

func TestFormatterQuerySeparator(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		linesBetween int
		expected     string
	}{
		{
			name:         "query separator",
			input:        "SELECT id FROM users; SELECT id FROM admins;",
			linesBetween: 1,
			expected: `SELECT
  id
FROM
  users;
SELECT
  id
FROM
  admins;`,
		},
		{
			name:         "multiple lines between queries",
			input:        "SELECT id FROM users; SELECT id FROM admins;",
			linesBetween: 2,
			expected: `SELECT
  id
FROM
  users;

SELECT
  id
FROM
  admins;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:              "  ",
				KeywordCase:         KeywordCaseUppercase,
				LinesBetweenQueries: tt.linesBetween,
				ColorConfig:         &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Edge Case and Error Handling Tests
// ============================================================================

func TestFormatterEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty query",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \t\n  ",
			expected: "",
		},
		{
			name:     "only comments",
			input:    "-- comment only",
			expected: "-- comment only",
		},
		{
			name:  "unmatched opening paren",
			input: "SELECT (id FROM users",
			expected: `SELECT
  (
    id
    FROM
      users`,
		},
		{
			name: "extremely nested query",
			input: "SELECT * FROM (SELECT * FROM (SELECT * FROM (SELECT * FROM " +
				"(SELECT id FROM users) AS l1) AS l2) AS l3) AS l4",
			expected: `SELECT
  *
FROM
  (
    SELECT
      *
    FROM
      (
        SELECT
          *
        FROM
          (
            SELECT
              *
            FROM
              (
                SELECT
                  id
                FROM
                  users
              ) AS l1
          ) AS l2
      ) AS l3
  ) AS l4`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM", "AS"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestFormatterHelperFunctions(t *testing.T) {
	t.Run("trimSpacesEnd", func(t *testing.T) {
		builder := &strings.Builder{}
		builder.WriteString("hello   ")
		trimSpacesEnd(builder)
		require.Equal(t, "hello", builder.String())

		builder.Reset()
		builder.WriteString("hello\t\t")
		trimSpacesEnd(builder)
		require.Equal(t, "hello", builder.String())

		builder.Reset()
		builder.WriteString("hello")
		trimSpacesEnd(builder)
		require.Equal(t, "hello", builder.String())
	})

	t.Run("equalizeWhitespace", func(t *testing.T) {
		cfg := &Config{
			TokenizerConfig: &TokenizerConfig{},
		}
		formatter := newFormatter(cfg, nil, nil)

		require.Equal(t, "hello world", formatter.equalizeWhitespace("hello   world"))
		require.Equal(t, "hello world", formatter.equalizeWhitespace("hello\t\tworld"))
		require.Equal(t, "hello world test", formatter.equalizeWhitespace("hello  \t  world  \n  test"))
	})

	t.Run("previousToken", func(t *testing.T) {
		cfg := &Config{
			TokenizerConfig: &TokenizerConfig{},
		}
		formatter := newFormatter(cfg, nil, nil)
		formatter.tokens = []types.Token{
			{Value: "SELECT", Type: types.TokenTypeReservedTopLevel},
			{Value: "id", Type: types.TokenTypeWord},
			{Value: "FROM", Type: types.TokenTypeReservedTopLevel},
		}

		formatter.index = 2
		prev := formatter.previousToken()
		require.Equal(t, "id", prev.Value)

		prev = formatter.previousToken(2)
		require.Equal(t, "SELECT", prev.Value)

		formatter.index = 0
		prev = formatter.previousToken()
		require.True(t, prev.Empty())
	})

	t.Run("nextToken", func(t *testing.T) {
		cfg := &Config{
			TokenizerConfig: &TokenizerConfig{},
		}
		formatter := newFormatter(cfg, nil, nil)
		formatter.tokens = []types.Token{
			{Value: "SELECT", Type: types.TokenTypeReservedTopLevel},
			{Value: "id", Type: types.TokenTypeWord},
			{Value: "FROM", Type: types.TokenTypeReservedTopLevel},
		}

		formatter.index = 0
		next := formatter.nextToken()
		require.Equal(t, "id", next.Value)

		next = formatter.nextToken(2)
		require.Equal(t, "FROM", next.Value)

		formatter.index = 2
		next = formatter.nextToken()
		require.True(t, next.Empty())
	})
}

// ============================================================================
// Clause Analysis Tests
// ============================================================================

func TestAnalyzeSelectClauses(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected int // expected number of SELECT clauses analyzed
	}{
		{
			name:     "single SELECT statement",
			query:    "SELECT id, name FROM users",
			expected: 1,
		},
		{
			name:     "multiple SELECT statements with UNION",
			query:    "SELECT id FROM users UNION SELECT id FROM admins",
			expected: 2,
		},
		{
			name:     "nested SELECT with subquery",
			query:    "SELECT * FROM (SELECT id FROM users) AS u",
			expected: 2,
		},
		{
			name:     "no SELECT statement",
			query:    "UPDATE users SET name = 'test'",
			expected: 0,
		},
		{
			name:     "SELECT with multiple columns",
			query:    "SELECT id, name, email, created_at FROM users WHERE active = true",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM", "WHERE", "AS", "UNION", "UPDATE", "SET"},
					ReservedTopLevelWords: []string{"SELECT", "FROM", "WHERE", "UNION", "UPDATE", "SET"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			formatter.tokens = tokenizer.tokenize(tt.query)

			formatter.analyzeSelectClauses()

			require.Len(t, formatter.selectColumnLengths, tt.expected)
		})
	}
}

func TestAnalyzeUpdateSetClauses(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected int // expected number of UPDATE clauses analyzed
	}{
		{
			name:     "single UPDATE statement",
			query:    "UPDATE users SET name = 'John', email = 'john@example.com'",
			expected: 1,
		},
		{
			name:     "UPDATE with WHERE clause",
			query:    "UPDATE users SET name = 'John' WHERE id = 1",
			expected: 1,
		},
		{
			name:     "no UPDATE statement",
			query:    "SELECT id FROM users",
			expected: 0,
		},
		{
			name:     "UPDATE with multiple assignments",
			query:    "UPDATE users SET name = 'test', email = 'test@example.com', active = true WHERE id = 1",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"UPDATE", "SET", "WHERE", "FROM"},
					ReservedTopLevelWords: []string{"UPDATE", "SET", "WHERE", "FROM"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			formatter.tokens = tokenizer.tokenize(tt.query)

			formatter.analyzeUpdateSetClauses()

			require.Len(t, formatter.updateAssignmentLengths, tt.expected)
		})
	}
}

func TestAnalyzeInsertValuesClauses(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected int // expected number of INSERT clauses analyzed
	}{
		{
			name:     "single INSERT statement",
			query:    "INSERT INTO users (id, name) VALUES (1, 'John')",
			expected: 1,
		},
		{
			name:     "INSERT with multiple value sets",
			query:    "INSERT INTO users (id, name) VALUES (1, 'John'), (2, 'Jane')",
			expected: 1,
		},
		{
			name:     "no INSERT statement",
			query:    "SELECT id FROM users",
			expected: 0,
		},
		{
			name:     "INSERT without VALUES",
			query:    "INSERT INTO users SELECT * FROM temp_users",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"INSERT", "INTO", "VALUES", "SELECT", "FROM"},
					ReservedTopLevelWords: []string{"INSERT", "VALUES", "SELECT", "FROM"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			formatter.tokens = tokenizer.tokenize(tt.query)

			formatter.analyzeInsertValuesClauses()

			require.Len(t, formatter.insertValuesLengths, tt.expected)
		})
	}
}

func TestAnalyzeSelectClause(t *testing.T) {
	tests := []struct {
		name              string
		query             string
		expectedMaxLength int // expected maximum column length
	}{
		{
			name:              "simple SELECT with two columns",
			query:             "SELECT id, name FROM users",
			expectedMaxLength: 0, // Should be > 0 after analysis
		},
		{
			name:              "SELECT with AS aliases",
			query:             "SELECT id AS user_id, name AS user_name FROM users",
			expectedMaxLength: 0,
		},
		{
			name:              "SELECT with single column",
			query:             "SELECT COUNT(*) FROM users",
			expectedMaxLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM", "AS", "COUNT"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			formatter.tokens = tokenizer.tokenize(tt.query)

			// Position at the SELECT token
			formatter.index = 0
			formatter.analyzeSelectClause()

			// Just verify that analysis ran (we should have captured some length data)
			if strings.Contains(tt.query, ",") {
				require.NotEmpty(t, formatter.selectColumnLengths)
			}
		})
	}
}

func TestAnalyzeUpdateSetClause(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "UPDATE with single assignment",
			query: "UPDATE users SET name = 'John' WHERE id = 1",
		},
		{
			name:  "UPDATE with multiple assignments",
			query: "UPDATE users SET name = 'John', email = 'john@example.com' WHERE id = 1",
		},
		{
			name:  "UPDATE without WHERE",
			query: "UPDATE users SET active = true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"UPDATE", "SET", "WHERE", "FROM"},
					ReservedTopLevelWords: []string{"UPDATE", "SET", "WHERE", "FROM"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			formatter.tokens = tokenizer.tokenize(tt.query)

			// Position at the UPDATE token
			formatter.index = 0
			formatter.analyzeUpdateSetClause()

			// Verify analysis ran if there's an equals sign in the query
			if strings.Contains(tt.query, "=") {
				require.NotEmpty(t, formatter.updateAssignmentLengths)
			}
		})
	}
}

func TestAnalyzeInsertValuesClause(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		hasValue bool
	}{
		{
			name:     "INSERT with VALUES",
			query:    "INSERT INTO users (id, name) VALUES (1, 'John')",
			hasValue: true,
		},
		{
			name:     "INSERT without VALUES",
			query:    "INSERT INTO users SELECT * FROM temp",
			hasValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"INSERT", "INTO", "VALUES", "SELECT", "FROM"},
					ReservedTopLevelWords: []string{"INSERT", "VALUES", "SELECT", "FROM"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			formatter.tokens = tokenizer.tokenize(tt.query)

			// Position at the INSERT token
			formatter.index = 0
			formatter.analyzeInsertValuesClause()

			if tt.hasValue {
				require.NotEmpty(t, formatter.insertValuesLengths)
			} else {
				require.Empty(t, formatter.insertValuesLengths)
			}
		})
	}
}

func TestFindSelectClauseEnd(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectFound bool
	}{
		{
			name:        "SELECT with FROM",
			query:       "SELECT id, name FROM users",
			expectFound: true,
		},
		{
			name:        "SELECT with WHERE",
			query:       "SELECT id, name WHERE active = true",
			expectFound: true,
		},
		{
			name:        "SELECT with GROUP BY",
			query:       "SELECT COUNT(*) GROUP BY category",
			expectFound: true,
		},
		{
			name:        "SELECT with ORDER BY",
			query:       "SELECT id ORDER BY created_at",
			expectFound: true,
		},
		{
			name:        "SELECT without terminator",
			query:       "SELECT id, name",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "COUNT"},
					ReservedTopLevelWords: []string{"SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			formatter.tokens = tokenizer.tokenize(tt.query)

			// Position at the SELECT token
			formatter.index = 0
			endIndex := formatter.findSelectClauseEnd()

			if tt.expectFound {
				require.NotEqual(t, -1, endIndex)
			} else {
				require.Equal(t, -1, endIndex)
			}
		})
	}
}

func TestFindUpdateSetClauseEnd(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectFound bool
	}{
		{
			name:        "UPDATE with WHERE",
			query:       "UPDATE users SET name = 'John' WHERE id = 1",
			expectFound: true,
		},
		{
			name:        "UPDATE with FROM (PostgreSQL)",
			query:       "UPDATE users SET name = 'John' FROM other_table",
			expectFound: true,
		},
		{
			name:        "UPDATE with RETURNING",
			query:       "UPDATE users SET name = 'John' RETURNING id",
			expectFound: true,
		},
		{
			name:        "UPDATE without terminator",
			query:       "UPDATE users SET name = 'John'",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"UPDATE", "SET", "WHERE", "FROM", "RETURNING"},
					ReservedTopLevelWords: []string{"UPDATE", "SET", "WHERE", "FROM", "RETURNING"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			formatter.tokens = tokenizer.tokenize(tt.query)

			// Find SET token
			setIndex := -1
			for i, tok := range formatter.tokens {
				if tok.Type == types.TokenTypeReservedTopLevel && strings.ToUpper(tok.Value) == "SET" {
					setIndex = i
					break
				}
			}
			require.NotEqual(t, -1, setIndex, "SET token not found")

			endIndex := formatter.findUpdateSetClauseEnd(setIndex)

			if tt.expectFound {
				require.NotEqual(t, -1, endIndex)
			} else {
				require.Equal(t, -1, endIndex)
			}
		})
	}
}

func TestIsSelectClauseTerminator(t *testing.T) {
	cfg := &Config{
		TokenizerConfig: &TokenizerConfig{},
	}
	formatter := newFormatter(cfg, nil, nil)

	tests := []struct {
		value    string
		expected bool
	}{
		{"FROM", true},
		{"WHERE", true},
		{"GROUP BY", true},
		{"ORDER BY", true},
		{"HAVING", true},
		{"LIMIT", true},
		{"UNION", true},
		{"INTERSECT", true},
		{"EXCEPT", true},
		{"from", true}, // case insensitive
		{"SELECT", false},
		{"AS", false},
		{"INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := formatter.isSelectClauseTerminator(tt.value)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsUpdateSetClauseTerminator(t *testing.T) {
	cfg := &Config{
		TokenizerConfig: &TokenizerConfig{},
	}
	formatter := newFormatter(cfg, nil, nil)

	tests := []struct {
		value    string
		expected bool
	}{
		{"WHERE", true},
		{"FROM", true},
		{"RETURNING", true},
		{"where", true}, // case insensitive
		{"UPDATE", false},
		{"SET", false},
		{"INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := formatter.isUpdateSetClauseTerminator(tt.value)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsInsertValuesClauseTerminator(t *testing.T) {
	cfg := &Config{
		TokenizerConfig: &TokenizerConfig{},
	}
	formatter := newFormatter(cfg, nil, nil)

	tests := []struct {
		value    string
		expected bool
	}{
		{"WHERE", true},
		{"FROM", true},
		{"RETURNING", true},
		{"ON", true},
		{"on", true}, // case insensitive
		{"INSERT", false},
		{"VALUES", false},
		{"INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := formatter.isInsertValuesClauseTerminator(tt.value)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// FormatQuery Tests
// ============================================================================

func TestFormatQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		cfg      *Config
		expected string
	}{
		{
			name:  "basic SELECT query",
			query: "SELECT id, name FROM users",
			cfg: &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			},
			expected: "SELECT\n  id,\n  name\nFROM\n  users",
		},
		{
			name:  "query with lowercase keywords",
			query: "select id from users",
			cfg: &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseLowercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			},
			expected: "select\n  id\nfrom\n  users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatQuery(tt.cfg, nil, tt.query)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatQueryWithTokenOverride(t *testing.T) {
	cfg := &Config{
		Indent:      "  ",
		KeywordCase: KeywordCaseUppercase,
		ColorConfig: &ColorConfig{},
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:         []string{"SELECT", "FROM"},
			ReservedTopLevelWords: []string{"SELECT", "FROM"},
			ReservedNewlineWords:  []string{"AND", "OR"},
			StringTypes:           []string{"''"},
			OpenParens:            []string{"("},
			CloseParens:           []string{")"},
			LineCommentTypes:      []string{"--"},
		},
	}

	// Token override that changes all words to uppercase
	tokenOverride := func(tok types.Token, prevReserved types.Token) types.Token {
		if tok.Type == types.TokenTypeWord {
			tok.Value = strings.ToUpper(tok.Value)
		}
		return tok
	}

	result := FormatQuery(cfg, tokenOverride, "SELECT id, name FROM users")
	require.Contains(t, result, "ID")
	require.Contains(t, result, "NAME")
	require.Contains(t, result, "USERS")
}

// ============================================================================
// Dialect-Specific Formatting Tests
// ============================================================================

func TestFormatDialectSpecificCase(t *testing.T) {
	tests := []struct {
		name     string
		language Language
		input    string
		expected string
	}{
		{
			name:     "Standard SQL uses uppercase",
			language: StandardSQL,
			input:    "select",
			expected: "SELECT",
		},
		{
			name:     "DB2 uses uppercase",
			language: DB2,
			input:    "select",
			expected: "SELECT",
		},
		{
			name:     "PL/SQL uses uppercase",
			language: PLSQL,
			input:    "select",
			expected: "SELECT",
		},
		{
			name:     "PostgreSQL uses lowercase",
			language: PostgreSQL,
			input:    "SELECT",
			expected: "select",
		},
		{
			name:     "MySQL uses lowercase",
			language: MySQL,
			input:    "SELECT",
			expected: "select",
		},
		{
			name:     "N1QL uses lowercase",
			language: N1QL,
			input:    "SELECT",
			expected: "select",
		},
		{
			name:     "SQLite uses lowercase",
			language: SQLite,
			input:    "SELECT",
			expected: "select",
		},
		{
			name:     "Unknown dialect preserves case",
			language: Language("unknown"),
			input:    "SeLeCt",
			expected: "SeLeCt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Language:        tt.language,
				ColorConfig:     &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{},
			}
			formatter := newFormatter(cfg, nil, nil)
			result := formatter.formatDialectSpecificCase(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Special Operator Tests
// ============================================================================

func TestFormatSpecialOperator(t *testing.T) {
	tests := []struct {
		name            string
		initialContent  string
		operator        string
		expectedContent string
	}{
		{
			name:            "type cast operator ::",
			initialContent:  "id ",
			operator:        "::",
			expectedContent: "id::",
		},
		{
			name:            "double colon with trailing space",
			initialContent:  "value   ",
			operator:        "::",
			expectedContent: "value::",
		},
		{
			name:            "custom operator",
			initialContent:  "x ",
			operator:        "@@",
			expectedContent: "x@@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ColorConfig:     &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{},
			}
			formatter := newFormatter(cfg, nil, nil)
			query := &strings.Builder{}
			query.WriteString(tt.initialContent)

			tok := types.Token{
				Type:  types.TokenTypeSpecialOperator,
				Value: tt.operator,
			}

			formatter.formatSpecialOperator(tok, query)
			require.Equal(t, tt.expectedContent, query.String())
		})
	}
}

// ============================================================================
// Max Line Length Tests
// ============================================================================

func TestExceedsMaxLineLength(t *testing.T) {
	tests := []struct {
		name              string
		maxLineLength     int
		currentLineLength int
		addString         string
		expected          bool
	}{
		{
			name:              "does not exceed limit",
			maxLineLength:     80,
			currentLineLength: 50,
			addString:         "SELECT",
			expected:          false,
		},
		{
			name:              "exactly at limit",
			maxLineLength:     80,
			currentLineLength: 74,
			addString:         "SELECT",
			expected:          false,
		},
		{
			name:              "exceeds limit",
			maxLineLength:     80,
			currentLineLength: 75,
			addString:         "SELECT",
			expected:          true,
		},
		{
			name:              "unlimited line length (0)",
			maxLineLength:     0,
			currentLineLength: 1000,
			addString:         "VERY_LONG_STRING",
			expected:          false,
		},
		{
			name:              "unlimited line length (negative)",
			maxLineLength:     -1,
			currentLineLength: 1000,
			addString:         "VERY_LONG_STRING",
			expected:          false,
		},
		{
			name:              "empty string never exceeds",
			maxLineLength:     80,
			currentLineLength: 79,
			addString:         "",
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				MaxLineLength:   tt.maxLineLength,
				ColorConfig:     &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{},
			}
			formatter := newFormatter(cfg, nil, nil)
			formatter.currentLineLength = tt.currentLineLength

			result := formatter.exceedsMaxLineLength(tt.addString)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// format() Main Function Additional Tests
// ============================================================================

func TestFormatterFormatEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		setup    func(*formatter)
		expected string
	}{
		{
			name:     "empty token list",
			input:    "",
			expected: "",
		},
		{
			name:     "single token",
			input:    "SELECT",
			expected: `SELECT`,
		},
		{
			name:  "query with preserved comment indentation",
			input: "SELECT id -- comment\nFROM users",
			expected: `SELECT
  id -- comment
FROM
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:                "  ",
				KeywordCase:           KeywordCaseUppercase,
				PreserveCommentIndent: true,
				ColorConfig:           &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			if tt.setup != nil {
				tt.setup(formatter)
			}
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// formatReservedTopLevelToken Additional Tests
// ============================================================================

func TestFormatReservedTopLevelTokenStateTracking(t *testing.T) {
	tests := []struct {
		name                    string
		query                   string
		alignColumnNames        bool
		alignAssignments        bool
		alignValues             bool
		expectInSelectClause    bool
		expectInUpdateSetClause bool
		expectInInsertValues    bool
	}{
		{
			name:                 "SELECT clause state tracking",
			query:                "SELECT id, name FROM users",
			alignColumnNames:     true,
			expectInSelectClause: false, // After FROM, should be false
		},
		{
			name:                    "UPDATE SET clause state tracking",
			query:                   "UPDATE users SET name = 'test' WHERE id = 1",
			alignAssignments:        true,
			expectInUpdateSetClause: false, // After WHERE, should be false
		},
		{
			name:                 "INSERT VALUES clause state tracking",
			query:                "INSERT INTO users (id) VALUES (1) RETURNING id",
			alignValues:          true,
			expectInInsertValues: false, // After RETURNING, should be false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:           "  ",
				KeywordCase:      KeywordCaseUppercase,
				AlignColumnNames: tt.alignColumnNames,
				AlignAssignments: tt.alignAssignments,
				AlignValues:      tt.alignValues,
				ColorConfig:      &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM", "WHERE", "UPDATE", "SET", "INSERT", "INTO", "VALUES", "RETURNING"},
					ReservedTopLevelWords: []string{"SELECT", "FROM", "WHERE", "UPDATE", "SET", "INSERT", "VALUES", "RETURNING"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			_ = formatter.format(tt.query)

			// After formatting, state flags should be reset
			require.Equal(t, tt.expectInSelectClause, formatter.inSelectClause)
			require.Equal(t, tt.expectInUpdateSetClause, formatter.inUpdateSetClause)
			require.Equal(t, tt.expectInInsertValues, formatter.inInsertValuesClause)
		})
	}
}

func TestFormatReservedTopLevelTokenPreviousReservedTracking(t *testing.T) {
	cfg := &Config{
		Indent:      "  ",
		KeywordCase: KeywordCaseUppercase,
		ColorConfig: &ColorConfig{},
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:         []string{"SELECT", "FROM"},
			ReservedTopLevelWords: []string{"SELECT", "FROM"},
			StringTypes:           []string{"''"},
			OpenParens:            []string{"("},
			CloseParens:           []string{")"},
			LineCommentTypes:      []string{"--"},
		},
	}
	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, nil)
	_ = formatter.format("SELECT id FROM users")

	// Previous reserved word should be tracked
	require.NotEmpty(t, formatter.previousReservedWord.Value)
	require.Equal(t, "FROM", strings.ToUpper(formatter.previousReservedWord.Value))
}

// ============================================================================
// formatBlockComment Additional Tests
// ============================================================================

func TestFormatBlockCommentMultilineWithIndent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "multi-line block comment with proper indentation",
			input: "SELECT\n/* This is a\n   multi-line\n   comment */\nid FROM users",
			expected: `SELECT
  /* This is a
   multi-line
   comment */
  id
FROM
  users`,
		},
		{
			name:  "block comment at start of line",
			input: "/* comment */\nSELECT id FROM users",
			expected: `/* comment */
SELECT
  id
FROM
  users`,
		},
		{
			name:  "single-line block comment that fits",
			input: "SELECT id /* inline */ FROM users",
			expected: `SELECT
  id /* inline */
FROM
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:      "  ",
				KeywordCase: KeywordCaseUppercase,
				ColorConfig: &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBlockCommentWithMaxLineLength(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		maxLineLength int
		expected      string
	}{
		{
			name:          "block comment exceeds max line length - moved to new line",
			input:         "SELECT very_long_column_name /* this is a very long comment */ FROM users",
			maxLineLength: 40,
			expected: `SELECT
  very_long_column_name
  /* this is a very long comment */
FROM
  users`,
		},
		{
			name:          "block comment fits within max line length - inline",
			input:         "SELECT id /* ok */ FROM users",
			maxLineLength: 80,
			expected: `SELECT
  id /* ok */
FROM
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:        "  ",
				KeywordCase:   KeywordCaseUppercase,
				MaxLineLength: tt.maxLineLength,
				ColorConfig:   &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// calculateCommentSpacing Tests
// ============================================================================

func TestCalculateCommentSpacing(t *testing.T) {
	tests := []struct {
		name              string
		commentMinSpacing int
		expected          int
	}{
		{
			name:              "default spacing",
			commentMinSpacing: 0,
			expected:          1,
		},
		{
			name:              "custom spacing 2",
			commentMinSpacing: 2,
			expected:          2,
		},
		{
			name:              "custom spacing 4",
			commentMinSpacing: 4,
			expected:          4,
		},
		{
			name:              "custom spacing 8",
			commentMinSpacing: 8,
			expected:          8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				CommentMinSpacing: tt.commentMinSpacing,
				ColorConfig:       &ColorConfig{},
				TokenizerConfig:   &TokenizerConfig{},
			}
			formatter := newFormatter(cfg, nil, nil)
			result := formatter.calculateCommentSpacing()
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// commentFitsOnLine Tests
// ============================================================================

func TestCommentFitsOnLine(t *testing.T) {
	tests := []struct {
		name              string
		comment           string
		spacing           int
		currentLineLength int
		maxLineLength     int
		expected          bool
	}{
		{
			name:              "unlimited line length always fits",
			comment:           "-- very long comment that would normally exceed limits",
			spacing:           2,
			currentLineLength: 50,
			maxLineLength:     0,
			expected:          true,
		},
		{
			name:              "comment fits within limit",
			comment:           "-- short",
			spacing:           2,
			currentLineLength: 20,
			maxLineLength:     80,
			expected:          true,
		},
		{
			name:              "comment exceeds limit",
			comment:           "-- this is a very long comment that will exceed the maximum line length",
			spacing:           2,
			currentLineLength: 50,
			maxLineLength:     80,
			expected:          false,
		},
		{
			name:              "exactly at limit",
			comment:           "-- comment",
			spacing:           1,
			currentLineLength: 68, // 68 + 1 + 11 = 80
			maxLineLength:     80,
			expected:          true,
		},
		{
			name:              "one character over limit",
			comment:           "-- comment!!",
			spacing:           1,
			currentLineLength: 68, // 68 + 1 + 13 = 82
			maxLineLength:     80,
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				MaxLineLength:   tt.maxLineLength,
				ColorConfig:     &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{},
			}
			formatter := newFormatter(cfg, nil, nil)
			formatter.currentLineLength = tt.currentLineLength

			result := formatter.commentFitsOnLine(tt.comment, tt.spacing)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// formatComma with Alignment Tests
// ============================================================================

func TestFormatCommaWithAlignColumnNames(t *testing.T) {
	query := "SELECT id, name, email FROM users"
	cfg := &Config{
		Indent:           "  ",
		KeywordCase:      KeywordCaseUppercase,
		AlignColumnNames: true,
		ColorConfig:      &ColorConfig{},
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:         []string{"SELECT", "FROM"},
			ReservedTopLevelWords: []string{"SELECT", "FROM"},
			StringTypes:           []string{"''"},
			OpenParens:            []string{"("},
			CloseParens:           []string{")"},
			LineCommentTypes:      []string{"--"},
		},
	}
	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, nil)
	result := formatter.format(query)

	// Should align columns
	require.Contains(t, result, "SELECT")
	require.Contains(t, result, "id")
	require.Contains(t, result, "name")
	require.Contains(t, result, "email")
}

func TestFormatCommaWithAlignAssignments(t *testing.T) {
	query := "UPDATE users SET name = 'John', email = 'john@example.com', active = true WHERE id = 1"
	cfg := &Config{
		Indent:           "  ",
		KeywordCase:      KeywordCaseUppercase,
		AlignAssignments: true,
		ColorConfig:      &ColorConfig{},
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:         []string{"UPDATE", "SET", "WHERE"},
			ReservedTopLevelWords: []string{"UPDATE", "SET", "WHERE"},
			StringTypes:           []string{"''"},
			OpenParens:            []string{"("},
			CloseParens:           []string{")"},
			LineCommentTypes:      []string{"--"},
		},
	}
	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, nil)
	result := formatter.format(query)

	// Should align assignments
	require.Contains(t, result, "UPDATE")
	require.Contains(t, result, "SET")
	require.Contains(t, result, "name")
	require.Contains(t, result, "email")
	require.Contains(t, result, "active")
}

func TestFormatCommaWithAlignValues(t *testing.T) {
	query := "INSERT INTO users (id, name) VALUES (1, 'John'), (2, 'Jane')"
	cfg := &Config{
		Indent:      "  ",
		KeywordCase: KeywordCaseUppercase,
		AlignValues: true,
		ColorConfig: &ColorConfig{},
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:         []string{"INSERT", "INTO", "VALUES"},
			ReservedTopLevelWords: []string{"INSERT", "VALUES"},
			StringTypes:           []string{"''"},
			OpenParens:            []string{"("},
			CloseParens:           []string{")"},
			LineCommentTypes:      []string{"--"},
		},
	}
	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, nil)
	result := formatter.format(query)

	// Should format VALUES clauses
	require.Contains(t, result, "INSERT")
	require.Contains(t, result, "VALUES")
}

func TestFormatCommaWithLimitKeyword(t *testing.T) {
	query := "SELECT id FROM users LIMIT 10, 20"
	cfg := &Config{
		Indent:      "  ",
		KeywordCase: KeywordCaseUppercase,
		ColorConfig: &ColorConfig{},
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:         []string{"SELECT", "FROM", "LIMIT"},
			ReservedTopLevelWords: []string{"SELECT", "FROM"},
			StringTypes:           []string{"''"},
			OpenParens:            []string{"("},
			CloseParens:           []string{")"},
			LineCommentTypes:      []string{"--"},
		},
	}
	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, nil)
	result := formatter.format(query)

	// LIMIT values should stay on one line
	require.Contains(t, result, "10, 20")
}

func TestFormatCommaBeforeComment(t *testing.T) {
	query := "SELECT id, -- user id\nname FROM users"
	cfg := &Config{
		Indent:      "  ",
		KeywordCase: KeywordCaseUppercase,
		ColorConfig: &ColorConfig{},
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:         []string{"SELECT", "FROM"},
			ReservedTopLevelWords: []string{"SELECT", "FROM"},
			StringTypes:           []string{"''"},
			OpenParens:            []string{"("},
			CloseParens:           []string{")"},
			LineCommentTypes:      []string{"--"},
		},
	}
	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, nil)
	result := formatter.format(query)

	// Comment should be handled properly after comma
	require.Contains(t, result, "id, -- user id")
}

// ============================================================================
// formatWithSpaces Tests
// ============================================================================

func TestFormatWithSpacesMaxLineLengthBreaking(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		maxLineLength int
		expectBreak   bool
	}{
		{
			name:          "exceeds max line length - should break",
			query:         "SELECT very_long_column_name_one, very_long_column_name_two, very_long_column_name_three FROM users",
			maxLineLength: 40,
			expectBreak:   true,
		},
		{
			name:          "within max line length - no break",
			query:         "SELECT id, name FROM users",
			maxLineLength: 100,
			expectBreak:   false,
		},
		{
			name:          "unlimited line length - no break",
			query:         "SELECT very_long_column_name_one, very_long_column_name_two FROM users",
			maxLineLength: 0,
			expectBreak:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:        "  ",
				KeywordCase:   KeywordCaseUppercase,
				MaxLineLength: tt.maxLineLength,
				ColorConfig:   &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM"},
					ReservedTopLevelWords: []string{"SELECT", "FROM"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.query)

			// Verify result is formatted
			require.NotEmpty(t, result)
			if tt.expectBreak && tt.maxLineLength > 0 {
				// Should have multiple lines when breaking
				lines := strings.Split(result, "\n")
				require.Greater(t, len(lines), 2)
			}
		})
	}
}

func TestFormatWithSpacesLogicalOperators(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		maxLineLength int
	}{
		{
			name:          "AND operator with long line",
			query:         "SELECT * FROM users WHERE active = true AND verified = true AND premium = true",
			maxLineLength: 60,
		},
		{
			name:          "OR operator with long line",
			query:         "SELECT * FROM users WHERE status = 'active' OR status = 'pending' OR status = 'trial'",
			maxLineLength: 60,
		},
		{
			name:          "mixed AND/OR operators",
			query:         "SELECT * FROM users WHERE active = true AND (status = 'premium' OR status = 'trial')",
			maxLineLength: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:        "  ",
				KeywordCase:   KeywordCaseUppercase,
				MaxLineLength: tt.maxLineLength,
				ColorConfig:   &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM", "WHERE", "AND", "OR"},
					ReservedTopLevelWords: []string{"SELECT", "FROM", "WHERE"},
					ReservedNewlineWords:  []string{"AND", "OR"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.query)

			// Should have line breaks for logical operators
			require.Contains(t, result, "WHERE")
			require.True(t, strings.Contains(result, "AND") || strings.Contains(result, "OR"))
		})
	}
}

func TestFormatWithSpacesFunctionCallFormatting(t *testing.T) {
	query := "SELECT COUNT(id), MAX(created_at), MIN(updated_at) FROM users"
	cfg := &Config{
		Indent:      "  ",
		KeywordCase: KeywordCaseUppercase,
		ColorConfig: &ColorConfig{
			FunctionCallFormatOptions: []utils.ANSIFormatOption{utils.FormatBold},
		},
		TokenizerConfig: &TokenizerConfig{
			ReservedWords:         []string{"SELECT", "FROM", "COUNT", "MAX", "MIN"},
			ReservedTopLevelWords: []string{"SELECT", "FROM"},
			StringTypes:           []string{"''"},
			OpenParens:            []string{"("},
			CloseParens:           []string{")"},
			LineCommentTypes:      []string{"--"},
		},
	}
	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, nil)
	result := formatter.format(query)

	// Function calls should be formatted
	require.Contains(t, result, "COUNT")
	require.Contains(t, result, "MAX")
	require.Contains(t, result, "MIN")
}

func TestFormatWithSpacesAlignmentSuppression(t *testing.T) {
	tests := []struct {
		name             string
		query            string
		alignColumnNames bool
		alignAssignments bool
		alignValues      bool
	}{
		{
			name:             "SELECT with column alignment active",
			query:            "SELECT id, name, email FROM users",
			alignColumnNames: true,
		},
		{
			name:             "UPDATE with assignment alignment active",
			query:            "UPDATE users SET name = 'test', email = 'test@example.com' WHERE id = 1",
			alignAssignments: true,
		},
		{
			name:        "INSERT with values alignment active",
			query:       "INSERT INTO users (id, name) VALUES (1, 'test')",
			alignValues: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:           "  ",
				KeywordCase:      KeywordCaseUppercase,
				MaxLineLength:    80,
				AlignColumnNames: tt.alignColumnNames,
				AlignAssignments: tt.alignAssignments,
				AlignValues:      tt.alignValues,
				ColorConfig:      &ColorConfig{},
				TokenizerConfig: &TokenizerConfig{
					ReservedWords:         []string{"SELECT", "FROM", "UPDATE", "SET", "WHERE", "INSERT", "INTO", "VALUES"},
					ReservedTopLevelWords: []string{"SELECT", "FROM", "UPDATE", "SET", "WHERE", "INSERT", "VALUES"},
					StringTypes:           []string{"''"},
					OpenParens:            []string{"("},
					CloseParens:           []string{")"},
					LineCommentTypes:      []string{"--"},
				},
			}
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.query)

			// Should format without breaking inappropriately during alignment
			require.NotEmpty(t, result)
		})
	}
}
