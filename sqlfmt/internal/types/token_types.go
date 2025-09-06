package types

type TokenType string

type Token struct {
	Type  TokenType
	Value string
	Key   string
}

func (t Token) Empty() bool {
	return t.Value == "" || t.Type == TokenTypeEmpty
}

const (
	TokenTypeEmpty                    TokenType = ""
	TokenTypeWhitespace               TokenType = "whitespace"
	TokenTypeWord                     TokenType = "word"
	TokenTypeString                   TokenType = "string"
	TokenTypeReserved                 TokenType = "reserved"
	TokenTypeReservedTopLevel         TokenType = "reserved-top-level"
	TokenTypeReservedTopLevelNoIndent TokenType = "reserved-top-level-no-indent"
	TokenTypeReservedNewline          TokenType = "reserved-newline"
	TokenTypeOperator                 TokenType = "operator"
	TokenTypeOpenParen                TokenType = "open-paren"
	TokenTypeCloseParen               TokenType = "close-paren"
	TokenTypeLineComment              TokenType = "line-comment"
	TokenTypeBlockComment             TokenType = "block-comment"
	TokenTypeNumber                   TokenType = "number"
	TokenTypePlaceholder              TokenType = "placeholder"
	TokenTypeBoolean                  TokenType = "boolean"
)
