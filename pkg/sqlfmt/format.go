package sqlfmt

import (
	"fmt"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/core"
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/dialects"
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/utils"
)

type Formatter = dialects.Formatter

// Format formats the SQL query according to an optional config.
func Format(query string, cfg ...*Config) string {
	return getFormatter(false, cfg...).Format(query)
}

// PrettyFormat formats the SQL query the same as Format but with coloring added.
func PrettyFormat(query string, cfg ...*Config) string {
	return getFormatter(true, cfg...).Format(query)
}

// PrettyPrint calls PrettyFormat and prints the formatted query.
func PrettyPrint(query string, cfg ...*Config) {
	fmt.Println(PrettyFormat(query, cfg...))
}

func getFormatter(forceWithColor bool, cfg ...*Config) Formatter {
	c := NewDefaultConfig()

	if len(cfg) > 1 {
		panic("cannot have more than one config")
	}

	if len(cfg) == 1 {
		c = cfg[0]
	}

	if forceWithColor && (c.ColorConfig == nil || c.ColorConfig.Empty()) {
		c.ColorConfig = NewDefaultColorConfig()
	}

	return createFormatterForLanguage(c)
}

func createFormatterForLanguage(c *Config) Formatter {
	// Convert public Config to internal core.Config
	coreCfg := &core.Config{
		Language:            core.Language(c.Language),
		Indent:              c.Indent,
		Uppercase:           c.Uppercase,
		LinesBetweenQueries: c.LinesBetweenQueries,
		Params:              convertParams(c.Params, c.Language),
		ColorConfig:         convertColorConfig(c.ColorConfig),
		TokenizerConfig:     convertTokenizerConfig(c.TokenizerConfig),
	}

	return dialects.CreateFormatterForLanguage(coreCfg)
}

func convertParams(p *Params, language Language) *utils.ParamsConfig {
	if p == nil {
		return nil
	}
	return &utils.ParamsConfig{
		MapParams:         p.MapParams,
		ListParams:        p.ListParams,
		UseSQLiteIndexing: language == SQLite, // Enable 1-based indexing for SQLite
	}
}

func convertColorConfig(cc *ColorConfig) *core.ColorConfig {
	if cc == nil {
		return nil
	}
	return &core.ColorConfig{
		ReservedWordFormatOptions: cc.ReservedWordFormatOptions,
		StringFormatOptions:       cc.StringFormatOptions,
		NumberFormatOptions:       cc.NumberFormatOptions,
		BooleanFormatOptions:      cc.BooleanFormatOptions,
		CommentFormatOptions:      cc.CommentFormatOptions,
		FunctionCallFormatOptions: cc.FunctionCallFormatOptions,
	}
}

func convertTokenizerConfig(tc *TokenizerConfig) *core.TokenizerConfig {
	if tc == nil {
		return nil
	}
	return &core.TokenizerConfig{
		ReservedWords:                 tc.ReservedWords,
		ReservedTopLevelWords:         tc.ReservedTopLevelWords,
		ReservedNewlineWords:          tc.ReservedNewlineWords,
		ReservedTopLevelWordsNoIndent: tc.ReservedTopLevelWordsNoIndent,
		StringTypes:                   tc.StringTypes,
		OpenParens:                    tc.OpenParens,
		CloseParens:                   tc.CloseParens,
		IndexedPlaceholderTypes:       tc.IndexedPlaceholderTypes,
		NamedPlaceholderTypes:         tc.NamedPlaceholderTypes,
		LineCommentTypes:              tc.LineCommentTypes,
		SpecialWordChars:              tc.SpecialWordChars,
	}
}

// Dedent removes any common leading whitespace from every line in a block of text.
// This function is provided for test compatibility.
func Dedent(text string) string {
	return utils.Dedent(text)
}

// Color constants - provided for test compatibility.
const (
	FormatReset = utils.FormatReset
	FormatBold  = utils.FormatBold

	ColorRed        = utils.ColorRed
	ColorGreen      = utils.ColorGreen
	ColorBlue       = utils.ColorBlue
	ColorCyan       = utils.ColorCyan
	ColorPurple     = utils.ColorPurple
	ColorGray       = utils.ColorGray
	ColorBrightBlue = utils.ColorBrightBlue
	ColorBrightCyan = utils.ColorBrightCyan
)

func convertToInternalConfig(c *Config) *core.Config {
	if c == nil {
		c = NewDefaultConfig()
	}
	return &core.Config{
		Language:            core.Language(c.Language),
		Indent:              c.Indent,
		Uppercase:           c.Uppercase,
		LinesBetweenQueries: c.LinesBetweenQueries,
		Params:              convertParams(c.Params, c.Language),
		ColorConfig:         convertColorConfig(c.ColorConfig),
		TokenizerConfig:     convertTokenizerConfig(c.TokenizerConfig),
	}
}

// NewStandardSQLFormatter creates a new standard SQL formatter.
// This function is provided for test compatibility.
func NewStandardSQLFormatter(cfg *Config) Formatter {
	coreCfg := convertToInternalConfig(cfg)
	return dialects.NewStandardSQLFormatter(coreCfg)
}

// NewDB2Formatter creates a new DB2 SQL formatter.
// This function is provided for test compatibility.
func NewDB2Formatter(cfg *Config) Formatter {
	coreCfg := convertToInternalConfig(cfg)
	return dialects.NewDB2Formatter(coreCfg)
}

// NewPostgreSQLFormatter creates a new PostgreSQL formatter.
// This function is provided for test compatibility.
func NewPostgreSQLFormatter(cfg *Config) Formatter {
	coreCfg := convertToInternalConfig(cfg)
	return dialects.NewPostgreSQLFormatter(coreCfg)
}

// NewPLSQLFormatter creates a new PL/SQL formatter.
// This function is provided for test compatibility.
func NewPLSQLFormatter(cfg *Config) Formatter {
	coreCfg := convertToInternalConfig(cfg)
	return dialects.NewPLSQLFormatter(coreCfg)
}

// NewN1QLFormatter creates a new N1QL formatter.
// This function is provided for test compatibility.
func NewN1QLFormatter(cfg *Config) Formatter {
	coreCfg := convertToInternalConfig(cfg)
	return dialects.NewN1QLFormatter(coreCfg)
}

// NewMySQLFormatter creates a new MySQL formatter.
// This function is provided for test compatibility.
func NewMySQLFormatter(cfg *Config) Formatter {
	coreCfg := convertToInternalConfig(cfg)
	return dialects.NewMySQLFormatter(coreCfg)
}

// NewSQLiteFormatter creates a new SQLite formatter.
// This function is provided for test compatibility.
func NewSQLiteFormatter(cfg *Config) Formatter {
	coreCfg := convertToInternalConfig(cfg)
	return dialects.NewSQLiteFormatter(coreCfg)
}

// Tokenizer configuration functions - provided for test compatibility.
func NewStandardSQLTokenizerConfig() *TokenizerConfig {
	internal := dialects.NewStandardSQLTokenizerConfig()
	return &TokenizerConfig{
		ReservedWords:                 internal.ReservedWords,
		ReservedTopLevelWords:         internal.ReservedTopLevelWords,
		ReservedNewlineWords:          internal.ReservedNewlineWords,
		ReservedTopLevelWordsNoIndent: internal.ReservedTopLevelWordsNoIndent,
		StringTypes:                   internal.StringTypes,
		OpenParens:                    internal.OpenParens,
		CloseParens:                   internal.CloseParens,
		IndexedPlaceholderTypes:       internal.IndexedPlaceholderTypes,
		NamedPlaceholderTypes:         internal.NamedPlaceholderTypes,
		LineCommentTypes:              internal.LineCommentTypes,
		SpecialWordChars:              internal.SpecialWordChars,
	}
}

func NewPostgreSQLTokenizerConfig() *TokenizerConfig {
	internal := dialects.NewPostgreSQLTokenizerConfig()
	return &TokenizerConfig{
		ReservedWords:                 internal.ReservedWords,
		ReservedTopLevelWords:         internal.ReservedTopLevelWords,
		ReservedNewlineWords:          internal.ReservedNewlineWords,
		ReservedTopLevelWordsNoIndent: internal.ReservedTopLevelWordsNoIndent,
		StringTypes:                   internal.StringTypes,
		OpenParens:                    internal.OpenParens,
		CloseParens:                   internal.CloseParens,
		IndexedPlaceholderTypes:       internal.IndexedPlaceholderTypes,
		NamedPlaceholderTypes:         internal.NamedPlaceholderTypes,
		LineCommentTypes:              internal.LineCommentTypes,
		SpecialWordChars:              internal.SpecialWordChars,
	}
}

func NewMySQLTokenizerConfig() *TokenizerConfig {
	internal := dialects.NewMySQLTokenizerConfig()
	return &TokenizerConfig{
		ReservedWords:                 internal.ReservedWords,
		ReservedTopLevelWords:         internal.ReservedTopLevelWords,
		ReservedNewlineWords:          internal.ReservedNewlineWords,
		ReservedTopLevelWordsNoIndent: internal.ReservedTopLevelWordsNoIndent,
		StringTypes:                   internal.StringTypes,
		OpenParens:                    internal.OpenParens,
		CloseParens:                   internal.CloseParens,
		IndexedPlaceholderTypes:       internal.IndexedPlaceholderTypes,
		NamedPlaceholderTypes:         internal.NamedPlaceholderTypes,
		LineCommentTypes:              internal.LineCommentTypes,
		SpecialWordChars:              internal.SpecialWordChars,
	}
}

func NewSQLiteTokenizerConfig() *TokenizerConfig {
	internal := dialects.NewSQLiteTokenizerConfig()
	return &TokenizerConfig{
		ReservedWords:                 internal.ReservedWords,
		ReservedTopLevelWords:         internal.ReservedTopLevelWords,
		ReservedNewlineWords:          internal.ReservedNewlineWords,
		ReservedTopLevelWordsNoIndent: internal.ReservedTopLevelWordsNoIndent,
		StringTypes:                   internal.StringTypes,
		OpenParens:                    internal.OpenParens,
		CloseParens:                   internal.CloseParens,
		IndexedPlaceholderTypes:       internal.IndexedPlaceholderTypes,
		NamedPlaceholderTypes:         internal.NamedPlaceholderTypes,
		LineCommentTypes:              internal.LineCommentTypes,
		SpecialWordChars:              internal.SpecialWordChars,
	}
}
