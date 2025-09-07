package dialects

import "github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/core"

var (
	sqliteReservedWords = []string{
		"ABORT", "ACTION", "ADD", "AFTER", "ALL", "ALTER", "ALWAYS", "ANALYZE", "AND", "AS", "ASC",
		"ATTACH", "AUTOINCREMENT", "BEFORE", "BEGIN", "BETWEEN", "BY", "CASCADE", "CASE", "CAST",
		"CHECK", "COLLATE", "COLUMN", "COMMIT", "CONFLICT", "CONSTRAINT", "CREATE", "CROSS",
		"CURRENT", "CURRENT_DATE", "CURRENT_TIME", "CURRENT_TIMESTAMP", "DATABASE", "DEFAULT",
		"DEFERRABLE", "DEFERRED", "DELETE", "DESC", "DETACH", "DISTINCT", "DO", "DROP", "EACH",
		"ELSE", "END", "ESCAPE", "EXCEPT", "EXCLUDE", "EXCLUSIVE", "EXISTS", "EXPLAIN", "FAIL",
		"FILTER", "FIRST", "FOLLOWING", "FOR", "FOREIGN", "FROM", "FULL", "GENERATED", "GLOB",
		"GROUP", "GROUPS", "HAVING", "IF", "IGNORE", "IMMEDIATE", "IN", "INDEX", "INDEXED",
		"INITIALLY", "INNER", "INSERT", "INSTEAD", "INTERSECT", "INTO", "IS", "ISNULL", "JOIN",
		"KEY", "LAST", "LEFT", "LIKE", "LIMIT", "MATCH", "MATERIALIZED", "NATURAL", "NO", "NOT",
		"NOTHING", "NOTNULL", "NULL", "NULLS", "OF", "OFFSET", "ON", "OR", "ORDER", "OTHERS",
		"OUTER", "OVER", "PARTITION", "PLAN", "PRAGMA", "PRECEDING", "PRIMARY", "QUERY", "RAISE",
		"RANGE", "RECURSIVE", "REFERENCES", "REGEXP", "REINDEX", "RELEASE", "RENAME", "REPLACE",
		"RESTRICT", "RETURNING", "RIGHT", "ROLLBACK", "ROW", "ROWS", "SAVEPOINT", "SELECT", "SET",
		"STORED", "STRICT", "TABLE", "TEMP", "TEMPORARY", "THEN", "TIES", "TO", "TRANSACTION",
		"TRIGGER", "UNBOUNDED", "UNION", "UNIQUE", "UPDATE", "USING", "VACUUM", "VALUES", "VIEW",
		"VIRTUAL", "WHEN", "WHERE", "WINDOW", "WITH", "WITHOUT",
	}

	sqliteReservedTopLevelWords = []string{
		"ADD", "AFTER", "ALTER COLUMN", "ALTER TABLE", "DELETE FROM", "EXCEPT", "FETCH FIRST", "FROM", "GROUP BY",
		"HAVING", "INSERT INTO", "INSERT", "LIMIT", "ORDER BY", "SELECT", "SET", "UPDATE", "VALUES", "WHERE",
		"WITH", "PRAGMA",
	}

	sqliteReservedTopLevelWordsNoIndent = []string{
		"INTERSECT ALL", "INTERSECT", "UNION ALL", "UNION",
	}

	sqliteReservedNewlineWords = []string{
		"AND", "CROSS JOIN", "ELSE", "INNER JOIN", "JOIN", "LEFT JOIN", "LEFT OUTER JOIN", "OR",
		"OUTER JOIN", "RIGHT JOIN", "RIGHT OUTER JOIN", "WHEN",
	}
)

type SQLiteFormatter struct {
	cfg *Config
}

func NewSQLiteFormatter(cfg *Config) *SQLiteFormatter {
	cfg.TokenizerConfig = NewSQLiteTokenizerConfig()
	return &SQLiteFormatter{cfg: cfg}
}

func NewSQLiteTokenizerConfig() *TokenizerConfig {
	return &TokenizerConfig{
		ReservedWords:                 sqliteReservedWords,
		ReservedTopLevelWords:         sqliteReservedTopLevelWords,
		ReservedNewlineWords:          sqliteReservedNewlineWords,
		ReservedTopLevelWordsNoIndent: sqliteReservedTopLevelWordsNoIndent,
		StringTypes:                   []string{`""`, "''", "``", "[]", "X''"},  // SQLite identifier and string quoting
		OpenParens:                    []string{"(", "CASE"},
		CloseParens:                   []string{")", "END"},
		IndexedPlaceholderTypes:       []string{"?"},                             // SQLite supports ? and ?NNN
		NamedPlaceholderTypes:         []string{":", "@", "$"},                   // SQLite named parameters
		LineCommentTypes:              []string{"--"},                           // SQLite only supports -- comments, not #
		SpecialWordChars:              []string{},
	}
}

func (sf *SQLiteFormatter) Format(query string) string {
	return core.FormatQuery(
		sf.cfg,
		nil,
		query,
	)
}