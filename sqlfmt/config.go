package sqlfmt

import "reflect"

type Language string

const (
	StandardSQL Language = "sql"

	DefaultIndent              = "  " // two spaces
	DefaultLinesBetweenQueries = 2
)

type Config struct {
	Language            Language
	Indent              string
	Uppercase           bool
	LinesBetweenQueries int
	Params              Params
}

func NewDefaultConfig() Config {
	return Config{
		Language:            StandardSQL,
		Indent:              DefaultIndent,
		LinesBetweenQueries: DefaultLinesBetweenQueries,
	}
}

func (c Config) Empty() bool {
	return reflect.DeepEqual(c, Config{})
}

type Params struct {
	MapParams  map[string]string
	ListParams []string
}

func NewMapParams(params map[string]string) Params {
	return Params{
		MapParams: params,
	}
}

func NewListParams(params []string) Params {
	return Params{
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

func NewTokenizerConfig(
	reservedWords []string,
	reservedTopLevelWords []string,
	reservedNewlineWords []string,
	reservedTopLevelWordsNoIndent []string,
	stringTypes []string,
	openParens []string,
	closeParens []string,
	indexedPlaceholderTypes []string,
	namedPlaceholderTypes []string,
	lineCommentTypes []string,
	specialWordChars []string,
) TokenizerConfig {
	return TokenizerConfig{
		ReservedWords:                 reservedWords,
		ReservedTopLevelWords:         reservedTopLevelWords,
		ReservedNewlineWords:          reservedNewlineWords,
		ReservedTopLevelWordsNoIndent: reservedTopLevelWordsNoIndent,
		StringTypes:                   stringTypes,
		OpenParens:                    openParens,
		CloseParens:                   closeParens,
		IndexedPlaceholderTypes:       indexedPlaceholderTypes,
		NamedPlaceholderTypes:         namedPlaceholderTypes,
		LineCommentTypes:              lineCommentTypes,
		SpecialWordChars:              specialWordChars,
	}
}
