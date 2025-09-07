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
		// Phase 4: Regex operators
		"NOT REGEXP", "NOT RLIKE",
		// Phase 5: Core clauses and locking
		"FOR UPDATE", "FOR SHARE", "LOCK IN SHARE MODE",
		"INSERT IGNORE", "REPLACE",
		// Phase 6: MySQL "Upsert"
		"ON DUPLICATE KEY UPDATE",
		// Phase 7: CTEs & Window Functions (8.0+)
		"WITH", "WITH RECURSIVE", "WINDOW", "OVER", "PARTITION BY",
		"RANGE", "ROWS", "UNBOUNDED", "PRECEDING", "FOLLOWING", "CURRENT ROW",
		// Window function names
		"ROW_NUMBER", "RANK", "DENSE_RANK", "NTILE", "LAG", "LEAD",
		"FIRST_VALUE", "LAST_VALUE", "NTH_VALUE", "PERCENT_RANK", "CUME_DIST",
		// Phase 8: DDL Essentials
		"UNIQUE", "FULLTEXT", "SPATIAL", "USING",
		"ALGORITHM", "INSTANT", "INPLACE", "COPY",
		"LOCK", "NONE", "SHARED", "EXCLUSIVE",
		"GENERATED", "ALWAYS", "VIRTUAL", "STORED",
		"CONSTRAINT", "CHECK", "FOREIGN KEY", "REFERENCES",
		// Phase 9: Stored Routines
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
		"DELIMITER",
	}...)

	// MySQL extends standard SQL top-level words with specific clauses.
	mySQLReservedTopLevelWords = append(standardSQLReservedTopLevelWords, []string{
		"INSERT IGNORE", "REPLACE", "ON DUPLICATE KEY UPDATE",
		// Phase 7: CTEs & Window Functions (8.0+)
		"WITH", "WITH RECURSIVE", "WINDOW",
		// Phase 8: DDL Essentials
		"CREATE INDEX", "CREATE UNIQUE INDEX", "CREATE FULLTEXT INDEX", "CREATE SPATIAL INDEX",
		"DROP INDEX", "ALTER TABLE",
		// Phase 9: Stored Routines
		"CREATE PROCEDURE", "CREATE FUNCTION", "DROP PROCEDURE", "DROP FUNCTION",
		"ALTER PROCEDURE", "ALTER FUNCTION",
	}...)

	// MySQL extends top-level words without indent to include locking clauses.
	mySQLReservedTopLevelWordsNoIndent = append(standardSQLReservedTopLevelWordsNoIndent, []string{
		// Phase 5: Locking clauses - appear at end of SELECT but don't add indentation
		"LOCK IN SHARE MODE", "FOR UPDATE", "FOR SHARE",
		// Phase 9: BEGIN starts at base level like PostgreSQL
		"BEGIN",
	}...)

	// MySQL extends newline words to include stored routine keywords
	mySQLReservedNewlineWords = append(standardSQLReservedNewlineWords, []string{
		// Phase 9: Stored routine control flow
		"BEGIN", "END", "END IF", "END WHILE", "END LOOP", "END REPEAT",
		"IF", "ELSEIF", "ELSE", // Remove THEN from newline words - it should follow WHEN/ELSE
		"WHILE", "DO", "LOOP", "REPEAT", "UNTIL",
		"DECLARE", "RETURN", "CALL", "LEAVE", "ITERATE",
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
		// Phase 2: Enhanced string and identifier support
		// Backticks for identifiers, single/double quotes for strings
		StringTypes:             []string{"''", "\"\"", "``"},
		OpenParens:              []string{"(", "CASE", "BEGIN", "IF", "WHILE", "LOOP", "REPEAT"},
		CloseParens:             []string{")", "END", "END IF", "END WHILE", "END LOOP", "END REPEAT"},
		IndexedPlaceholderTypes: []string{"?"},       // MySQL uses ? for parameters
		NamedPlaceholderTypes:   []string{},          // Phase 2: Still no named parameters
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
func (msf *MySQLFormatter) tokenOverride(tok types.Token, previousReservedWord types.Token) types.Token {
	// Phase 9: Handle DELIMITER statements as pass-through
	if tok.Type == types.TokenTypeReserved && tok.Value == "DELIMITER" {
		// Treat DELIMITER as a special pass-through token that shouldn't be formatted
		return types.Token{
			Type:  types.TokenTypeLineComment, // Use comment type to preserve formatting
			Value: tok.Value,
			Key:   tok.Key,
		}
	}
	
	
	// Phase 9: Handle stored routine specific keywords
	if tok.Type == types.TokenTypeReserved {
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
	}
	
	// Phase 6: Handle VALUES as function in ON DUPLICATE KEY UPDATE context
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
	
	// Phase 8: Handle GENERATED ALWAYS AS sequence formatting
	if tok.Type == types.TokenTypeReserved && tok.Value == "AS" {
		// If previous reserved word is "GENERATED ALWAYS", keep AS as a normal reserved word
		if previousReservedWord.Value == "GENERATED ALWAYS" {
			return tok
		}
	}
	
	// Phase 8: Handle table options in DDL statements
	if tok.Type == types.TokenTypeReserved {
		switch tok.Value {
		case "ALGORITHM", "LOCK":
			// These should be treated as regular reserved words in DDL context
			return tok
		case "INSTANT", "INPLACE", "COPY", "NONE", "SHARED", "EXCLUSIVE":
			// These are option values that should be formatted as words when following their keywords
			if previousReservedWord.Value == "ALGORITHM" || previousReservedWord.Value == "LOCK" {
				return types.Token{
					Type:  types.TokenTypeWord,
					Value: strings.ToLower(tok.Value),
					Key:   tok.Key,
				}
			}
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
	}
	
	// Phase 4: Handle MySQL-specific operators
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
	
	return tok
}
