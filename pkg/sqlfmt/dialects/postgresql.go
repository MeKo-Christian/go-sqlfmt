package dialects

import (
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/core"
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
)

var (
	// PostgreSQL reuses standard SQL reserved words and adds PostgreSQL-specific ones.
	postgreSQLReservedWords = append(standardSQLReservedWords, []string{
		"ILIKE", "SIMILAR TO", "ON CONFLICT", "DO UPDATE", "DO NOTHING",
		// Window function keywords
		"WINDOW", "OVER", "PARTITION BY", "FILTER", "RANGE", "ROWS", "GROUPS",
		"UNBOUNDED", "PRECEDING", "FOLLOWING", "CURRENT ROW",
		"EXCLUDE", "TIES", "NO OTHERS",
		// NULLS ordering
		"NULLS FIRST", "NULLS LAST",
		// LATERAL joins
		"LATERAL",
		// Array functions
		"ARRAY", "UNNEST",
		// Procedural keywords
		"LANGUAGE", "RETURNS", "AS", "DECLARE", "BEGIN",
		// Function modifiers
		"IMMUTABLE", "STABLE", "VOLATILE", "STRICT", "CALLED ON NULL INPUT",
		"SECURITY DEFINER", "SECURITY INVOKER", "LEAKPROOF", "NOT LEAKPROOF",
		// Return types
		"SETOF", "TABLE", "TRIGGER", "VOID",
		// Function options
		"COST", "ROWS", "SUPPORT", "PARALLEL SAFE", "PARALLEL UNSAFE", "PARALLEL RESTRICTED",
		// DDL and Index keywords
		"CONCURRENTLY", "IF NOT EXISTS", "IF EXISTS",
		// Index methods
		"BTREE", "HASH", "GIN", "GIST", "SPGIST", "BRIN",
		// Index options
		"INCLUDE", "TABLESPACE", "WITH", "FILLFACTOR", "FASTUPDATE",
		// Additional DDL keywords
		"REINDEX", "CLUSTER", "VACUUM", "ANALYZE",
	}...)

	// PostgreSQL adds CTE and RETURNING support to top-level words.
	postgreSQLReservedTopLevelWords = append(standardSQLReservedTopLevelWords, []string{
		"WITH", "WITH RECURSIVE", "RETURNING", "WINDOW",
		// Procedural blocks and functions
		"DO", "CREATE FUNCTION", "CREATE OR REPLACE FUNCTION",
		// DDL operations
		"CREATE INDEX", "CREATE UNIQUE INDEX", "DROP INDEX", "REINDEX",
	}...)

	postgreSQLReservedTopLevelWordsNoIndent = standardSQLReservedTopLevelWordsNoIndent

	// Add LATERAL join support to newline words.
	postgreSQLReservedNewlineWords = append(standardSQLReservedNewlineWords, []string{
		"LATERAL JOIN", "LEFT LATERAL JOIN", "RIGHT LATERAL JOIN", "CROSS JOIN LATERAL",
	}...)
)

type PostgreSQLFormatter struct {
	cfg *Config
}

func NewPostgreSQLFormatter(cfg *Config) *PostgreSQLFormatter {
	cfg.TokenizerConfig = NewPostgreSQLTokenizerConfig()
	return &PostgreSQLFormatter{cfg: cfg}
}

func NewPostgreSQLTokenizerConfig() *TokenizerConfig {
	return &TokenizerConfig{
		ReservedWords:                 postgreSQLReservedWords,
		ReservedTopLevelWords:         postgreSQLReservedTopLevelWords,
		ReservedNewlineWords:          postgreSQLReservedNewlineWords,
		ReservedTopLevelWordsNoIndent: postgreSQLReservedTopLevelWordsNoIndent,
		StringTypes:                   []string{`""`, "N''", "''", "``", "$$"},
		OpenParens:                    []string{"(", "CASE"},
		CloseParens:                   []string{")", "END"},
		IndexedPlaceholderTypes:       []string{"$"},
		NamedPlaceholderTypes:         []string{"@", ":"},
		LineCommentTypes:              []string{"--"},
	}
}

func (psf *PostgreSQLFormatter) Format(query string) string {
	return core.FormatQuery(
		psf.cfg,
		psf.tokenOverride,
		query,
	)
}

// tokenOverride handles PostgreSQL-specific token formatting.
func (psf *PostgreSQLFormatter) tokenOverride(tok types.Token, previousReservedWord types.Token) types.Token {
	// Handle type cast operator :: - format without spaces (PostgreSQL convention)
	if tok.Type == types.TokenTypeOperator && tok.Value == "::" {
		// Create a new token with a modified type to handle special formatting
		return types.Token{
			Type:  types.TokenTypeSpecialOperator,
			Value: tok.Value,
			Key:   tok.Key,
		}
	}

	return tok
}
