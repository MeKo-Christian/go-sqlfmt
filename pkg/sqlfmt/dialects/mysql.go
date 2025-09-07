package dialects

import (
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/core"
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
)

var (
	// MySQL reuses standard SQL reserved words and adds MySQL-specific ones.
	// This is a basic set for Phase 1 - will be expanded in later phases.
	mySQLReservedWords = append(standardSQLReservedWords, []string{
		// Basic MySQL keywords for Phase 1
		"AUTO_INCREMENT", "BINARY", "BLOB", "BOOLEAN", "BTREE", "HASH",
		"ENGINE", "INNODB", "MYISAM", "MEMORY", "CHARACTER SET", "CHARSET",
		"COLLATION", "SIGNED", "UNSIGNED", "ZEROFILL",
		// Basic MySQL functions
		"IFNULL", "ISNULL", "CONCAT", "LENGTH", "SUBSTRING",
		// Storage engines and options
		"STORAGE", "DYNAMIC", "FIXED", "COMPRESSED", "REDUNDANT", "COMPACT",
	}...)

	// MySQL uses the same top-level words as standard SQL for now.
	mySQLReservedTopLevelWords = standardSQLReservedTopLevelWords

	// MySQL uses the same top-level words without indent as standard SQL.
	mySQLReservedTopLevelWordsNoIndent = standardSQLReservedTopLevelWordsNoIndent

	// MySQL uses the same newline words as standard SQL for now.
	mySQLReservedNewlineWords = standardSQLReservedNewlineWords
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
		OpenParens:              []string{"(", "CASE"},
		CloseParens:             []string{")", "END"},
		IndexedPlaceholderTypes: []string{"?"},        // MySQL uses ? for parameters
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
// Phase 1: Basic implementation, will be expanded in later phases.
func (msf *MySQLFormatter) tokenOverride(tok types.Token, previousReservedWord types.Token) types.Token {
	// Phase 1: No special token handling yet - this will be expanded
	// in later phases for JSON operators, NULL-safe equality, etc.
	return tok
}