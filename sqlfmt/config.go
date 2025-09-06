package sqlfmt

import (
	"reflect"

	"github.com/maxrichie5/go-sqlfmt/sqlfmt/internal/utils"
)

type Language string

const (
	StandardSQL Language = "sql"
	PLSQL       Language = "pl/sql"
	DB2         Language = "db2"
	N1QL        Language = "n1ql"
	PostgreSQL  Language = "postgresql"

	DefaultIndent              = "  " // two spaces
	DefaultLinesBetweenQueries = 2
)

type Config struct {
	Language            Language
	Indent              string
	Uppercase           bool
	LinesBetweenQueries int
	Params              *Params
	ColorConfig         *ColorConfig
	TokenizerConfig     *TokenizerConfig
}

func NewDefaultConfig() *Config {
	return &Config{
		Language:            StandardSQL,
		Indent:              DefaultIndent,
		LinesBetweenQueries: DefaultLinesBetweenQueries,
		Params:              NewMapParams(nil),
		ColorConfig:         &ColorConfig{},
		TokenizerConfig:     &TokenizerConfig{},
	}
}

func (c *Config) WithLang(lang Language) *Config {
	c.Language = lang
	return c
}

func (c *Config) WithIndent(indent string) *Config {
	c.Indent = indent
	return c
}

func (c *Config) WithUppercase() *Config {
	c.Uppercase = true
	return c
}

func (c *Config) WithLinesBetweenQueries(linesBetweenQueries int) *Config {
	c.LinesBetweenQueries = linesBetweenQueries
	return c
}

func (c *Config) WithParams(params *Params) *Config {
	c.Params = params
	return c
}

func (c *Config) WithColorConfig(config *ColorConfig) *Config {
	c.ColorConfig = config
	return c
}

func (c *Config) WithTokenizerConfig(config *TokenizerConfig) *Config {
	c.TokenizerConfig = config
	return c
}

func (c *Config) Empty() bool {
	return reflect.DeepEqual(*c, Config{})
}

type Params struct {
	MapParams  map[string]string
	ListParams []string
}

func NewMapParams(params map[string]string) *Params {
	if params == nil {
		params = map[string]string{}
	}
	return &Params{
		MapParams: params,
	}
}

func NewListParams(params []string) *Params {
	if params == nil {
		params = []string{}
	}
	return &Params{
		ListParams: params,
	}
}

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

type ColorConfig struct {
	ReservedWordFormatOptions []utils.ANSIFormatOption
	StringFormatOptions       []utils.ANSIFormatOption
	NumberFormatOptions       []utils.ANSIFormatOption
	BooleanFormatOptions      []utils.ANSIFormatOption
	CommentFormatOptions      []utils.ANSIFormatOption
	FunctionCallFormatOptions []utils.ANSIFormatOption
}

func NewDefaultColorConfig() *ColorConfig {
	return &ColorConfig{
		ReservedWordFormatOptions: []utils.ANSIFormatOption{utils.ColorCyan, utils.FormatBold},
		StringFormatOptions:       []utils.ANSIFormatOption{utils.ColorGreen},
		NumberFormatOptions:       []utils.ANSIFormatOption{utils.ColorBrightBlue},
		BooleanFormatOptions:      []utils.ANSIFormatOption{utils.ColorPurple, utils.FormatBold},
		CommentFormatOptions:      []utils.ANSIFormatOption{utils.ColorGray},
		FunctionCallFormatOptions: []utils.ANSIFormatOption{utils.ColorBrightCyan},
	}
}

func (c *ColorConfig) Empty() bool {
	return len(c.ReservedWordFormatOptions) == 0 &&
		len(c.StringFormatOptions) == 0 &&
		len(c.NumberFormatOptions) == 0 &&
		len(c.BooleanFormatOptions) == 0 &&
		len(c.CommentFormatOptions) == 0 &&
		len(c.FunctionCallFormatOptions) == 0
}
