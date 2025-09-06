package dialects

import (
	"github.com/maxrichie5/go-sqlfmt/sqlfmt/internal/core"
	"github.com/maxrichie5/go-sqlfmt/sqlfmt/internal/types"
)

var (
	// PostgreSQL reuses standard SQL reserved words and adds PostgreSQL-specific ones.
	postgreSQLReservedWords = append(standardSQLReservedWords, []string{}...)

	postgreSQLReservedTopLevelWords         = standardSQLReservedTopLevelWords
	postgreSQLReservedTopLevelWordsNoIndent = standardSQLReservedTopLevelWordsNoIndent
	postgreSQLReservedNewlineWords          = standardSQLReservedNewlineWords
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
		StringTypes:                   []string{`""`, "N''", "''", "``", "[]", "$$"},
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
