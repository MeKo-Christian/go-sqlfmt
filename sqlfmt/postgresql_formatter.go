package sqlfmt

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
		IndexedPlaceholderTypes:       []string{"?"},
		NamedPlaceholderTypes:         []string{"@", ":"},
		LineCommentTypes:              []string{"--"},
	}
}

func (psf *PostgreSQLFormatter) Format(query string) string {
	return newFormatter(
		psf.cfg,
		newTokenizer(NewPostgreSQLTokenizerConfig()),
		nil,
	).format(query)
}
