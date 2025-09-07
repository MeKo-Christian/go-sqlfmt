package core

import "github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/utils"

// Formatter interface for SQL formatting.
type Formatter interface {
	Format(query string) string
}

// Language type for SQL dialect identification.
type Language string

const (
	StandardSQL Language = "sql"
	PLSQL       Language = "pl/sql"
	DB2         Language = "db2"
	N1QL        Language = "n1ql"
	PostgreSQL  Language = "postgresql"
)

// Config represents the configuration for formatting.
type Config struct {
	Language            Language
	Indent              string
	Uppercase           bool
	LinesBetweenQueries int
	Params              *utils.ParamsConfig
	ColorConfig         *ColorConfig
	TokenizerConfig     *TokenizerConfig
}

// TokenizerConfig represents tokenizer configuration.
type TokenizerConfig struct {
	ReservedWords                 []string
	ReservedTopLevelWords         []string
	ReservedNewlineWords          []string
	ReservedTopLevelWordsNoIndent []string
	StringTypes                   []string
	OpenParens                    []string
	CloseParens                   []string
	IndexedPlaceholderTypes       []string
	NamedPlaceholderTypes         []string
	LineCommentTypes              []string
	SpecialWordChars              []string
}

// ColorConfig represents color formatting configuration.
type ColorConfig struct {
	ReservedWordFormatOptions []utils.ANSIFormatOption
	StringFormatOptions       []utils.ANSIFormatOption
	NumberFormatOptions       []utils.ANSIFormatOption
	BooleanFormatOptions      []utils.ANSIFormatOption
	CommentFormatOptions      []utils.ANSIFormatOption
	FunctionCallFormatOptions []utils.ANSIFormatOption
}
