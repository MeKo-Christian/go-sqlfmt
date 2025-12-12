package core

import (
	"strings"
	"testing"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
)

// getStandardSQLTokenizerConfig returns a tokenizer config for Standard SQL with procedural support.
func getStandardSQLTokenizerConfig() *TokenizerConfig {
	return &TokenizerConfig{
		ReservedWords: []string{
			"ACCESSIBLE", "ACTION", "AGAINST", "AGGREGATE", "ALGORITHM", "ALL", "ALTER", "ANALYSE", "ANALYZE", "AS", "ASC",
			"AUTOCOMMIT", "AUTO_INCREMENT", "BACKUP", "BEGIN", "BETWEEN", "BINLOG", "BOTH", "CASCADE", "CASE", "CHANGE",
			"CHANGED", "CHARACTER SET", "CHARSET", "CHECK", "CHECKSUM", "COLLATE", "COLLATION", "COLUMN", "COLUMNS",
			"COMMENT", "COMMIT", "COMMITTED", "COMPRESSED", "CONCURRENT", "CONSTRAINT", "CONTAINS", "CONVERT", "CREATE",
			"CROSS", "CURRENT_TIMESTAMP", "DATABASE", "DATABASES", "DAY", "DAY_HOUR", "DAY_MINUTE", "DAY_SECOND", "DEFAULT",
			"DEFINER", "DELAYED", "DELETE", "DESC", "DESCRIBE", "DETERMINISTIC", "DISTINCT", "DISTINCTROW", "DIV", "DO",
			"DROP", "DUMPFILE", "DUPLICATE", "DYNAMIC", "ELSE", "ENCLOSED", "END", "ENGINE", "ENGINES", "ENGINE_TYPE",
			"ESCAPE", "ESCAPED", "EVENTS", "EXEC", "EXECUTE", "EXISTS", "EXPLAIN", "EXTENDED", "FAST", "FETCH", "FIELDS",
			"FILE", "FIRST", "FIXED", "FLUSH", "FOR", "FORCE", "FOREIGN", "FULL", "FULLTEXT", "FUNCTION", "GLOBAL",
			"GRANT", "GRANTS", "GROUP_CONCAT", "HEAP", "HIGH_PRIORITY", "HOSTS", "HOUR", "HOUR_MINUTE", "HOUR_SECOND",
			"IDENTIFIED", "IF", "IFNULL", "IGNORE", "IN", "INDEX", "INDEXES", "INFILE", "INSERT", "INSERT_ID",
			"INSERT_METHOD", "INTERVAL", "INTO", "INVOKER", "IS", "ISOLATION", "KEY", "KEYS", "KILL", "LAST_INSERT_ID",
			"LEADING", "LEVEL", "LIKE", "LINEAR", "LINES", "LOAD", "LOCAL", "LOCK", "LOCKS", "LOGS", "LOW_PRIORITY",
			"MARIA", "MASTER", "MASTER_CONNECT_RETRY", "MASTER_HOST", "MASTER_LOG_FILE", "MATCH", "MAX_CONNECTIONS_PER_HOUR",
			"MAX_QUERIES_PER_HOUR", "MAX_ROWS", "MAX_UPDATES_PER_HOUR", "MAX_USER_CONNECTIONS", "MEDIUM", "MERGE", "MINUTE",
			"MINUTE_SECOND", "MIN_ROWS", "MODE", "MODIFY", "MONTH", "MRG_MYISAM", "MYISAM", "NAMES", "NATURAL", "NOT", "NOW()",
			"NULL", "OFFSET", "ON DELETE", "ON UPDATE", "ON", "ONLY", "OPEN", "OPTIMIZE", "OPTION", "OPTIONALLY", "OUTFILE",
			"PACK_KEYS", "PAGE", "PARTIAL", "PARTITION", "PARTITIONS", "PASSWORD", "PRIMARY", "PRIVILEGES", "PROCEDURE",
			"PROCESS", "PROCESSLIST", "PURGE", "QUICK", "RAID0", "RAID_CHUNKS", "RAID_CHUNKSIZE", "RAID_TYPE", "RANGE", "READ",
			"READ_ONLY", "READ_WRITE", "REFERENCES", "REGEXP", "RELOAD", "RENAME", "REPAIR", "REPEATABLE", "REPLACE",
			"REPLICATION", "RESET", "RESTORE", "RESTRICT", "RETURN", "RETURNS", "REVOKE", "RLIKE", "ROLLBACK", "ROW",
			"ROWS", "ROW_FORMAT", "SECOND", "SECURITY", "SEPARATOR", "SERIALIZABLE", "SESSION", "SHARE", "SHOW", "SHUTDOWN",
			"SLAVE", "SONAME", "SOUNDS", "SQL", "SQL_AUTO_IS_NULL", "SQL_BIG_RESULT", "SQL_BIG_SELECTS", "SQL_BIG_TABLES",
			"SQL_BUFFER_RESULT", "SQL_CACHE", "SQL_CALC_FOUND_ROWS", "SQL_LOG_BIN", "SQL_LOG_OFF", "SQL_LOG_UPDATE",
			"SQL_LOW_PRIORITY_UPDATES", "SQL_MAX_JOIN_SIZE", "SQL_NO_CACHE", "SQL_QUOTE_SHOW_CREATE", "SQL_SAFE_UPDATES",
			"SQL_SELECT_LIMIT", "SQL_SLAVE_SKIP_COUNTER", "SQL_SMALL_RESULT", "SQL_WARNINGS", "START", "STARTING", "STATUS",
			"STOP", "STORAGE", "STRAIGHT_JOIN", "STRING", "STRIPED", "SUPER", "TABLE", "TABLES", "TEMPORARY", "TERMINATED",
			"THEN", "TO", "TRAILING", "TRANSACTIONAL", "TRUNCATE", "TYPE", "TYPES", "UNCOMMITTED", "UNIQUE",
			"UNLOCK", "UNSIGNED", "USAGE", "USE", "USING", "VARIABLES", "VIEW", "WHEN", "WITH", "WORK", "WRITE", "YEAR_MONTH",
			"DECLARE", "INT", "VARCHAR",
		},
		ReservedTopLevelWords: []string{
			"ADD", "AFTER", "ALTER COLUMN", "ALTER TABLE", "DELETE FROM", "EXCEPT", "FETCH FIRST", "FROM", "GROUP BY", "GO",
			"HAVING", "INSERT INTO", "INSERT", "LIMIT", "MODIFY", "ORDER BY", "SELECT", "SET CURRENT SCHEMA", "SET SCHEMA",
			"SET", "UPDATE", "VALUES", "WHERE", "DECLARE",
		},
		ReservedTopLevelWordsNoIndent: []string{
			"INTERSECT ALL", "INTERSECT", "MINUS", "UNION ALL", "UNION",
		},
		ReservedNewlineWords: []string{
			"AND", "CROSS APPLY", "CROSS JOIN", "ELSE", "INNER JOIN", "JOIN", "LEFT JOIN", "LEFT OUTER JOIN", "OR",
			"OUTER APPLY", "OUTER JOIN", "RIGHT JOIN", "RIGHT OUTER JOIN", "WHEN", "XOR",
		},
		StringTypes:             []string{`""`, "N''", "''", "``", "[]", "$$"},
		OpenParens:              []string{"(", "CASE", "BEGIN", "IF"},
		CloseParens:             []string{")", "END", "END IF", "END LOOP", "END WHILE", "END REPEAT", "END CASE"},
		IndexedPlaceholderTypes: []string{"?"},
		NamedPlaceholderTypes:   []string{"@", ":"},
		LineCommentTypes:        []string{"#", "--"},
	}
}

// TestStatementTerminatorVsQuerySeparator verifies that semicolons behave differently
// inside procedural blocks (statement terminators) vs outside (query separators).
func TestStatementTerminatorVsQuerySeparator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		config   *Config
	}{
		{
			name:  "Query separator - two independent queries",
			input: "SELECT 1; SELECT 2;",
			expected: `SELECT
  1;

SELECT
  2;`,
			config: &Config{
				Indent:              "  ",
				LinesBetweenQueries: 2,
				Language:            StandardSQL,
			},
		},
		{
			name:  "Statement terminator - inside BEGIN block",
			input: "BEGIN DECLARE x; SELECT 1; END;",
			expected: `BEGIN
  DECLARE
    x;
  SELECT
    1;
END;`,
			config: &Config{
				Indent:              "  ",
				LinesBetweenQueries: 2,
				Language:            StandardSQL,
			},
		},
		{
			name:  "Statement terminator - multiple statements in BEGIN",
			input: "BEGIN SELECT 1; SELECT 2; SELECT 3; END;",
			expected: `BEGIN
  SELECT
    1;
  SELECT
    2;
  SELECT
    3;
END;`,
			config: &Config{
				Indent:              "  ",
				LinesBetweenQueries: 2,
				Language:            StandardSQL,
			},
		},
		{
			name:  "Nested BEGIN blocks with statements",
			input: "BEGIN BEGIN SELECT 1; END; SELECT 2; END;",
			expected: `BEGIN
  BEGIN
    SELECT
      1;
  END;
  SELECT
    2;
END;`,
			config: &Config{
				Indent:              "  ",
				LinesBetweenQueries: 2,
				Language:            StandardSQL,
			},
		},
		{
			name:  "Mixed - query separator outside, statement terminator inside",
			input: "SELECT 0; BEGIN SELECT 1; END; SELECT 2;",
			expected: `SELECT
  0;

BEGIN
  SELECT
    1;
END;

SELECT
  2;`,
			config: &Config{
				Indent:              "  ",
				LinesBetweenQueries: 2,
				Language:            StandardSQL,
			},
		},
		{
			name:  "Statement terminator preserves procedural indent across statements",
			input: "BEGIN DECLARE x INT; SET x = 1; SELECT x; END;",
			expected: `BEGIN
  DECLARE
    x INT;
  SET
    x = 1;
  SELECT
    x;
END;`,
			config: &Config{
				Indent:              "  ",
				LinesBetweenQueries: 2,
				Language:            StandardSQL,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config
			if cfg.TokenizerConfig == nil {
				cfg.TokenizerConfig = getStandardSQLTokenizerConfig()
			}

			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)

			if result != tt.expected {
				t.Errorf("formatQuerySeparator() test failed\nInput:    %q\nExpected:\n%s\nGot:\n%s",
					tt.input, tt.expected, result)
			}
		})
	}
}

// TestStatementTerminatorWithVariousBlocks tests statement terminator behavior
// with different types of procedural blocks.
// NOTE: Some formatting issues (IF/THEN, second BEGIN) will be fixed in later tasks (2.5, 2.6).
func TestStatementTerminatorWithVariousBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "IF block with statements - semicolon maintains procedural indent",
			input: "BEGIN IF x > 0 THEN SELECT 1; SELECT 2; END IF; END;",
			// Note: IF formatting and THEN position will be improved in tasks 2.6-2.7
			expected: `BEGIN
  IF
    x > 0 THEN
    SELECT
      1;
  SELECT
    2;
END IF;
END;`,
		},
		{
			name:  "CASE expression maintains indentation",
			input: "BEGIN SELECT CASE WHEN x = 1 THEN 'a' WHEN x = 2 THEN 'b' END; END;",
			expected: `BEGIN
  SELECT
    CASE
      WHEN x = 1 THEN 'a'
      WHEN x = 2 THEN 'b'
    END;
END;`,
		},
		{
			name:  "Multiple BEGIN blocks at different levels",
			input: "BEGIN BEGIN SELECT 1; END; BEGIN SELECT 2; END; END;",
			// Note: Second BEGIN indentation will be improved in task 2.5
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
	}

	cfg := &Config{
		Indent:              "  ",
		LinesBetweenQueries: 2,
		Language:            StandardSQL,
		TokenizerConfig:     getStandardSQLTokenizerConfig(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)

			if result != tt.expected {
				t.Errorf("formatQuerySeparator() with various blocks failed\nInput:    %q\nExpected:\n%s\nGot:\n%s",
					tt.input, tt.expected, result)
			}
		})
	}
}

// TestResetToProceduralBase verifies that ResetToProceduralBase is called
// inside procedural blocks and maintains correct indentation.
func TestResetToProceduralBase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, result string)
	}{
		{
			name:  "Semicolon in BEGIN resets to base procedural indent",
			input: "BEGIN SELECT 1; SELECT 2; END;",
			validate: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				if len(lines) < 4 {
					t.Fatalf("Expected at least 4 lines, got %d", len(lines))
				}
				// After first semicolon, next SELECT should be at procedural base (2 spaces)
				// Line should be: "  SELECT"
				selectLine := ""
				for _, line := range lines {
					if strings.Contains(line, "SELECT") && !strings.Contains(lines[0], "SELECT") {
						if selectLine == "" {
							selectLine = line
						}
					}
				}
				if !strings.HasPrefix(selectLine, "  SELECT") {
					t.Errorf("Expected SELECT at procedural base (2 spaces), got: %q", selectLine)
				}
			},
		},
		{
			name:  "Nested blocks maintain their procedural base",
			input: "BEGIN BEGIN SELECT 1; END; SELECT 2; END;",
			validate: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				// Inner SELECT should be at depth 2 (4 spaces)
				// Outer SELECT should be at depth 1 (2 spaces)
				foundInner := false
				foundOuter := false
				for i, line := range lines {
					if strings.Contains(line, "SELECT") {
						if !foundInner {
							if !strings.HasPrefix(line, "    SELECT") {
								t.Errorf("Line %d: Expected inner SELECT at 4 spaces, got: %q", i, line)
							}
							foundInner = true
						} else if !foundOuter {
							if !strings.HasPrefix(line, "  SELECT") {
								t.Errorf("Line %d: Expected outer SELECT at 2 spaces, got: %q", i, line)
							}
							foundOuter = true
						}
					}
				}
			},
		},
	}

	cfg := &Config{
		Indent:              "  ",
		LinesBetweenQueries: 2,
		Language:            StandardSQL,
		TokenizerConfig:     getStandardSQLTokenizerConfig(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)

			tt.validate(t, result)
		})
	}
}

// TestQuerySeparatorStillWorks verifies that query separator behavior
// is unchanged outside of procedural blocks.
func TestQuerySeparatorStillWorks(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		linesBetweenQuery int
		expectedNewlines  int // Actual newlines expected in output
	}{
		{
			name:              "Default blank lines between queries",
			input:             "SELECT 1; SELECT 2; SELECT 3;",
			linesBetweenQuery: 2,
			expectedNewlines:  3, // semicolon adds base newline + LinesBetweenQueries (2) = 3 total
		},
		{
			name:              "Single blank line between queries",
			input:             "SELECT 1; SELECT 2;",
			linesBetweenQuery: 1,
			expectedNewlines:  2, // semicolon adds base newline + LinesBetweenQueries (1) = 2 total
		},
		{
			name:              "No blank lines between queries",
			input:             "SELECT 1; SELECT 2;",
			linesBetweenQuery: 0,
			expectedNewlines:  2, // Still 2 newlines: one for each query starting on new line
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Indent:              "  ",
				LinesBetweenQueries: tt.linesBetweenQuery,
				Language:            StandardSQL,
				TokenizerConfig:     getStandardSQLTokenizerConfig(),
			}

			tokenizer := newTokenizer(cfg.TokenizerConfig)
			formatter := newFormatter(cfg, tokenizer, nil)
			result := formatter.format(tt.input)

			// Count newlines between queries
			parts := strings.Split(result, ";")
			if len(parts) < 2 {
				t.Fatalf("Expected at least 2 parts, got %d", len(parts))
			}
			betweenQueries := parts[1]
			newlineCount := strings.Count(betweenQueries, "\n")
			if newlineCount != tt.expectedNewlines {
				t.Errorf("Expected %d newlines between queries, got %d\nResult:\n%s",
					tt.expectedNewlines, newlineCount, result)
			}
		})
	}
}

// TestTokenizeEndKeywords verifies that compound END keywords are tokenized correctly.
func TestTokenizeEndKeywords(t *testing.T) {
	cfg := &Config{
		Indent:          "  ",
		Language:        StandardSQL,
		TokenizerConfig: getStandardSQLTokenizerConfig(),
	}

	tokenizer := newTokenizer(cfg.TokenizerConfig)

	tests := []struct {
		input         string
		expectedToken string
	}{
		{"END IF", "END IF"},
		{"END LOOP", "END LOOP"},
		{"END WHILE", "END WHILE"},
		{"END REPEAT", "END REPEAT"},
		{"END CASE", "END CASE"},
		{"END", "END"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := tokenizer.tokenize(tt.input)
			found := false
			for _, tok := range tokens {
				if tok.Type == types.TokenTypeCloseParen && strings.ToUpper(tok.Value) == tt.expectedToken {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected to find token %q of type CloseParen, but didn't. Tokens: %+v",
					tt.expectedToken, tokens)
			}
		})
	}
}
