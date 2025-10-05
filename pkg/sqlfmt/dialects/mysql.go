package dialects

import (
	"strings"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/core"
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
)

var (
	// MySQL reuses standard SQL reserved words and adds MySQL-specific ones.
	// This is a basic set - will be expanded later.
	mySQLReservedWords = append(standardSQLReservedWords, []string{
		// Basic MySQL keywords
		"AUTO_INCREMENT", "BINARY", "BLOB", "BOOLEAN", "BTREE", "HASH",
		"ENGINE", "INNODB", "MYISAM", "MEMORY", "CHARACTER SET", "CHARSET",
		"COLLATION", "SIGNED", "UNSIGNED", "ZEROFILL",
		// Basic MySQL functions
		"IFNULL", "ISNULL", "CONCAT", "LENGTH", "SUBSTRING",
		// Storage engines and options
		"STORAGE", "DYNAMIC", "FIXED", "COMPRESSED", "REDUNDANT", "COMPACT",
		// Regex operators
		"NOT REGEXP", "NOT RLIKE",
		// Locking and insert clauses
		"FOR UPDATE", "FOR SHARE", "LOCK IN SHARE MODE",
		"INSERT IGNORE", "REPLACE",
		// MySQL upsert (ON DUPLICATE KEY UPDATE)
		"ON DUPLICATE KEY UPDATE",
		// CTEs and window functions (MySQL 8.0+)
		"WITH", "WITH RECURSIVE", "WINDOW", "OVER", "PARTITION BY",
		"RANGE", "ROWS", "UNBOUNDED", "PRECEDING", "FOLLOWING", "CURRENT ROW",
		// Window function names
		"ROW_NUMBER", "RANK", "DENSE_RANK", "NTILE", "LAG", "LEAD",
		"FIRST_VALUE", "LAST_VALUE", "NTH_VALUE", "PERCENT_RANK", "CUME_DIST",
		// DDL Essentials
		"UNIQUE", "FULLTEXT", "SPATIAL", "USING",
		"ALGORITHM", "INSTANT", "INPLACE", "COPY",
		"LOCK", "NONE", "SHARED", "EXCLUSIVE",
		"GENERATED", "ALWAYS", "VIRTUAL", "STORED",
		"CONSTRAINT", "CHECK", "FOREIGN KEY", "REFERENCES",
		// ALTER TABLE compound keywords
		"ADD COLUMN", "MODIFY COLUMN", "DROP COLUMN",
		// Stored Routines
		"PROCEDURE", "FUNCTION", "BEGIN", "END",
		"RETURNS", "RETURN", "DETERMINISTIC", "NOT DETERMINISTIC",
		"READS SQL DATA", "MODIFIES SQL DATA", "NO SQL", "CONTAINS SQL",
		"SQL SECURITY", "DEFINER", "INVOKER",
		"DECLARE", "HANDLER", "CURSOR", "CONTINUE", "EXIT", "UNDO",
		"IF", "THEN", "ELSE", "ELSEIF", "END IF",
		"WHILE", "DO", "END WHILE",
		"LOOP", "END LOOP", "LEAVE", "ITERATE",
		"REPEAT", "UNTIL", "END REPEAT",
		"CALL", "INOUT", "OUT", "IN",
	}...)

	// MySQL extends standard SQL top-level words with specific clauses.
	// IMPORTANT: Multi-word keywords must come first so they match before single words
	mySQLReservedTopLevelWords = append([]string{
		// Multi-word keywords first (longer matches)
		"INSERT IGNORE", "ON DUPLICATE KEY UPDATE",
		// ALTER TABLE compound keywords
		"ADD COLUMN", "MODIFY COLUMN", "DROP COLUMN",
		// CTEs and window functions (MySQL 8.0+)
		"WITH RECURSIVE",
		// DDL statements
		"CREATE UNIQUE INDEX", "CREATE FULLTEXT INDEX", "CREATE SPATIAL INDEX", "CREATE INDEX",
		"DROP INDEX", "ALTER TABLE",
		// Stored procedures and functions
		"CREATE PROCEDURE", "CREATE FUNCTION", "DROP PROCEDURE", "DROP FUNCTION",
		"ALTER PROCEDURE", "ALTER FUNCTION",
		// Single-word keywords
		"REPLACE", "WITH", "WINDOW",
	}, standardSQLReservedTopLevelWords...)

	// MySQL extends top-level words without indent to include locking clauses.
	mySQLReservedTopLevelWordsNoIndent = append(standardSQLReservedTopLevelWordsNoIndent, []string{
		// Locking clauses - appear at end of SELECT but don't add indentation
		"LOCK IN SHARE MODE", "FOR UPDATE", "FOR SHARE",
		// BEGIN starts at base level (for stored procedures)
		"BEGIN",
	}...)

	// MySQL extends newline words to include stored routine keywords.
	mySQLReservedNewlineWords = append(standardSQLReservedNewlineWords, []string{
		// ALTER TABLE compound keywords
		"ADD COLUMN", "MODIFY COLUMN", "DROP COLUMN",
		// ALTER TABLE options (not LOCK - conflicts with "LOCK IN SHARE MODE")
		"ALGORITHM",
		// Stored procedure control flow keywords - proper indentation structure
		"ELSEIF", "ELSE", "END IF", "END WHILE", "END LOOP", "END REPEAT",
		// DO is already in standard newline words, WHILE/IF are opening parens
		"UNTIL", // UNTIL ends a REPEAT block
		"DECLARE", "RETURN", "CALL", "LEAVE", "ITERATE", "EXIT",
		"OPEN", "CLOSE", "FETCH",
	}...)
)

type MySQLFormatter struct {
	cfg *Config
}

func NewMySQLFormatter(cfg *Config) *MySQLFormatter {
	cfg.TokenizerConfig = NewMySQLTokenizerConfig()
	return &MySQLFormatter{cfg: cfg}
}

func NewMySQLTokenizerConfig() *TokenizerConfig {
	return &TokenizerConfig{
		ReservedWords:                 mySQLReservedWords,
		ReservedTopLevelWords:         mySQLReservedTopLevelWords,
		ReservedNewlineWords:          mySQLReservedNewlineWords,
		ReservedTopLevelWordsNoIndent: mySQLReservedTopLevelWordsNoIndent,
		// String and identifier support
		// Backticks for identifiers, single/double quotes for strings
		// Hex/bit literal forms X'ABCD' and B'1010'
		StringTypes:             []string{"''", "\"\"", "``", "X''", "B''"},
		OpenParens:              []string{"(", "CASE", "BEGIN", "WHILE", "LOOP", "REPEAT"},
		CloseParens:             []string{")", "END", "END IF", "END WHILE", "END LOOP", "END REPEAT"},
		IndexedPlaceholderTypes: []string{"?"},       // MySQL uses ? for parameters
		NamedPlaceholderTypes:   []string{},          // MySQL doesn't support named parameters
		LineCommentTypes:        []string{"--", "#"}, // MySQL supports both -- and #
		SpecialWordChars:        []string{},          // Default special characters
	}
}

func (msf *MySQLFormatter) Format(query string) string {
	return core.FormatQuery(
		msf.cfg,
		msf.tokenOverride,
		query,
	)
}

// tokenOverride handles MySQL-specific token formatting.
func (msf *MySQLFormatter) tokenOverride(
	tok types.Token,
	previousReservedWord types.Token,
) types.Token {
	// Handle stored procedure specific keywords
	if tok.Type == types.TokenTypeReserved {
		if result := msf.handleStoredRoutineKeywords(tok, previousReservedWord); !result.Empty() {
			return result
		}
	}

	// Handle THEN keyword - should be inline after IF/ELSEIF/WHEN
	// Don't create newline for THEN after these keywords
	if tok.Type == types.TokenTypeReserved && strings.ToUpper(tok.Value) == "THEN" {
		if !previousReservedWord.Empty() {
			prevVal := strings.ToUpper(previousReservedWord.Value)
			// After IF, ELSEIF, or WHEN, THEN should stay inline
			if prevVal == "IF" || prevVal == "ELSEIF" || strings.HasSuffix(prevVal, "WHEN") {
				return types.Token{
					Type:  types.TokenTypeWord,
					Value: strings.ToLower(tok.Value),
					Key:   tok.Key,
				}
			}
		}
	}

	// Handle VALUES as function in ON DUPLICATE KEY UPDATE context
	if result := msf.handleValuesInUpsert(tok, previousReservedWord); !result.Empty() {
		return result
	}

	// Handle GENERATED ALWAYS AS sequence formatting
	if result := msf.handleGeneratedAlwaysAs(tok, previousReservedWord); !result.Empty() {
		return result
	}

	// Handle table options in DDL statements (ALGORITHM, LOCK, etc.)
	if tok.Type == types.TokenTypeReserved {
		if result := msf.handleTableOptions(tok, previousReservedWord); !result.Empty() {
			return result
		}
	}

	// Handle MySQL-specific operators (->>, ->, etc.)
	if result := msf.handleMySQLOperators(tok); !result.Empty() {
		return result
	}

	return tok
}

func (msf *MySQLFormatter) handleStoredRoutineKeywords(tok types.Token, previousReservedWord types.Token) types.Token {
	switch tok.Value {
	case "DETERMINISTIC", "NOT DETERMINISTIC":
		// Keep these as reserved words in routine context
		return tok
	case "READS SQL DATA", "MODIFIES SQL DATA", "NO SQL", "CONTAINS SQL":
		// Keep as reserved for routine characteristics
		return tok
	case "SQL SECURITY":
		// Keep as reserved for security specification
		return tok
	case "DEFINER", "INVOKER":
		// These follow SQL SECURITY
		if previousReservedWord.Value == "SQL SECURITY" {
			return types.Token{
				Type:  types.TokenTypeWord,
				Value: strings.ToLower(tok.Value),
				Key:   tok.Key,
			}
		}
		return tok
	}
	return types.Token{}
}

func (msf *MySQLFormatter) handleValuesInUpsert(tok types.Token, previousReservedWord types.Token) types.Token {
	// Handle VALUES as function in ON DUPLICATE KEY UPDATE context
	if tok.Type == types.TokenTypeReservedTopLevel && tok.Value == "VALUES" {
		// If the previous reserved word is "ON DUPLICATE KEY UPDATE", treat VALUES as a word/function
		if previousReservedWord.Value == "ON DUPLICATE KEY UPDATE" {
			return types.Token{
				Type:  types.TokenTypeWord,
				Value: tok.Value,
				Key:   tok.Key,
			}
		}
	}
	return types.Token{}
}

func (msf *MySQLFormatter) handleGeneratedAlwaysAs(tok types.Token, previousReservedWord types.Token) types.Token {
	// Handle GENERATED ALWAYS AS sequence formatting
	if tok.Type == types.TokenTypeReserved && tok.Value == "AS" {
		// If previous reserved word is "GENERATED ALWAYS", keep AS as a normal reserved word
		if previousReservedWord.Value == "GENERATED ALWAYS" {
			return tok
		}
	}
	return types.Token{}
}

func (msf *MySQLFormatter) handleTableOptions(tok types.Token, previousReservedWord types.Token) types.Token {
	switch tok.Value {
	case "ALGORITHM", "LOCK":
		// These should be treated as regular reserved words in DDL context
		return tok
	case "INSTANT", "INPLACE", "COPY", "NONE", "SHARED", "EXCLUSIVE":
		// These are option values that should remain as reserved words (uppercase)
		return tok
	case "VIRTUAL", "STORED":
		// These are generated column storage options
		if previousReservedWord.Value == "GENERATED ALWAYS AS" {
			return types.Token{
				Type:  types.TokenTypeWord,
				Value: strings.ToLower(tok.Value),
				Key:   tok.Key,
			}
		}
		return tok
	}
	return types.Token{}
}

func (msf *MySQLFormatter) handleMySQLOperators(tok types.Token) types.Token {
	// Handle MySQL-specific operators (<=>, ->, ->>)
	if tok.Type == types.TokenTypeOperator {
		switch tok.Value {
		case "<=>":
			// NULL-safe equality: format as normal comparison operator with spaces
			return tok
		case "->", "->>":
			// JSON operators: format without spaces (MySQL convention)
			return types.Token{
				Type:  types.TokenTypeSpecialOperator,
				Value: tok.Value,
				Key:   tok.Key,
			}
		case "<<", ">>":
			// Bitwise shift operators: format with normal spacing
			return tok
		}
	}
	return types.Token{}
}
