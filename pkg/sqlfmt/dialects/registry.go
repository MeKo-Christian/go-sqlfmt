package dialects

import (
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/core"
)

// Formatter interface is re-exported from core package.
type Formatter = core.Formatter

// Re-export types from core.
type (
	Config          = core.Config
	Language        = core.Language
	TokenizerConfig = core.TokenizerConfig
	ColorConfig     = core.ColorConfig
)

// Re-export constants from core.
const (
	StandardSQL = core.StandardSQL
	PLSQL       = core.PLSQL
	DB2         = core.DB2
	N1QL        = core.N1QL
	PostgreSQL  = core.PostgreSQL
	MySQL       = core.MySQL
)

// CreateFormatterForLanguage creates a formatter based on the language configuration.
func CreateFormatterForLanguage(c *Config) Formatter {
	switch c.Language {
	case DB2:
		return NewDB2Formatter(c)
	case N1QL:
		return NewN1QLFormatter(c)
	case PLSQL:
		return NewPLSQLFormatter(c)
	case PostgreSQL:
		return NewPostgreSQLFormatter(c)
	case MySQL:
		return NewMySQLFormatter(c)
	default:
		return NewStandardSQLFormatter(c)
	}
}
