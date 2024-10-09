package sqlfmt

type tokenType string

const (
	tokenTypeEmpty                    tokenType = ""
	tokenTypeWhitespace               tokenType = "whitespace"
	tokenTypeWord                     tokenType = "word"
	tokenTypeString                   tokenType = "string"
	tokenTypeReserved                 tokenType = "reserved"
	tokenTypeReservedTopLevel         tokenType = "reserved-top-level"
	tokenTypeReservedTopLevelNoIndent tokenType = "reserved-top-level-no-indent"
	tokenTypeReservedNewline          tokenType = "reserved-newline"
	tokenTypeOperator                 tokenType = "operator"
	tokenTypeOpenParen                tokenType = "open-paren"
	tokenTypeCloseParen               tokenType = "close-paren"
	tokenTypeLineComment              tokenType = "line-comment"
	tokenTypeBlockComment             tokenType = "block-comment"
	tokenTypeNumber                   tokenType = "number"
	tokenTypePlaceholder              tokenType = "placeholder"
)

type token struct {
	typ   tokenType
	value string
	key   string
}

func (t token) empty() bool {
	return t.value == "" || t.typ == tokenTypeEmpty
}
