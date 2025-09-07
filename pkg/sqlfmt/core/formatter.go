package core

import (
	"regexp"
	"strings"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/utils"
)

var (
	limitKeywordRegex              = regexp.MustCompile(`(?i)^LIMIT$`)
	newLineFollowByWhitespaceRegex = regexp.MustCompile(`\n[ \t]*`)
	atLeastOneWhitespaceRegex      = regexp.MustCompile(`\s+`)
)

// trimSpacesEnd removes trailing spaces and tabs from a string.
func trimSpacesEnd(b *strings.Builder) {
	s := b.String()
	s = strings.TrimRight(s, " \t")
	b.Reset()
	b.WriteString(s)
}

// formatter formats SQL queries for better readability.
type formatter struct {
	cfg                  *Config
	indentation          *utils.Indentation
	inlineBlock          *utils.InlineBlock
	params               *utils.Params
	tokenizer            *tokenizer
	tokenOverride        func(tok types.Token, previousReservedWord types.Token) types.Token
	previousReservedWord types.Token
	tokens               []types.Token
	index                int
}

// newFormatter creates a new formatter instance.
func newFormatter(cfg *Config, tokenizer *tokenizer,
	tokenOverride func(tok types.Token, previousReservedWord types.Token) types.Token,
) *formatter {
	if cfg.ColorConfig == nil {
		cfg.ColorConfig = &ColorConfig{}
	}
	return &formatter{
		cfg:                  cfg,
		indentation:          utils.NewIndentation(cfg.Indent),
		inlineBlock:          utils.NewInlineBlock(),
		params:               utils.NewParams(cfg.Params),
		tokenizer:            tokenizer,
		tokenOverride:        tokenOverride,
		previousReservedWord: types.Token{},
		tokens:               []types.Token{},
		index:                0,
	}
}

// format formats whitespace in a SQL string to make it easier to read.
func (f *formatter) format(query string) string {
	f.tokens = f.tokenizer.tokenize(query)
	formattedQuery := f.getFormattedQueryFromTokens()
	return strings.TrimSpace(formattedQuery)
}

// FormatQuery is a public wrapper function for creating a formatter and formatting a query.
func FormatQuery(
	cfg *Config,

	tokenOverride func(tok types.Token,

		previousReservedWord types.Token) types.Token,

	query string,
) string {
	tokenizer := newTokenizer(cfg.TokenizerConfig)
	formatter := newFormatter(cfg, tokenizer, tokenOverride)
	return formatter.format(query)
}

// getFormattedQueryFromTokens processes the types.Tokens to create a formatted query.
func (f *formatter) getFormattedQueryFromTokens() string {
	formattedQuery := &strings.Builder{}

	for i, tok := range f.tokens {
		f.index = i

		if f.tokenOverride != nil {
			tok = f.tokenOverride(tok, f.previousReservedWord)
		}

		f.formatToken(tok, formattedQuery)
	}
	return formattedQuery.String()
}

func (f *formatter) formatToken(tok types.Token, formattedQuery *strings.Builder) {
	formatters := map[types.TokenType]func(types.Token, *strings.Builder){
		types.TokenTypeWhitespace:               func(t types.Token, q *strings.Builder) {},
		types.TokenTypeLineComment:              f.formatLineComment,
		types.TokenTypeBlockComment:             f.formatBlockComment,
		types.TokenTypeReservedTopLevel:         f.formatReservedTopLevelToken,
		types.TokenTypeReservedTopLevelNoIndent: f.formatReservedTopLevelNoIndentToken,
		types.TokenTypeReservedNewline:          f.formatReservedNewlineToken,
		types.TokenTypeReserved:                 f.formatReservedToken,
		types.TokenTypeOpenParen:                f.formatOpeningParentheses,
		types.TokenTypeCloseParen:               f.formatClosingParentheses,
		types.TokenTypeWord:                     f.formatWordOrPlaceholder,
		types.TokenTypePlaceholder:              f.formatWordOrPlaceholder,
		types.TokenTypeString:                   f.formatString,
		types.TokenTypeNumber:                   f.formatNumber,
		types.TokenTypeBoolean:                  f.formatBoolean,
		types.TokenTypeSpecialOperator:          f.formatSpecialOperator,
	}

	if formatter, ok := formatters[tok.Type]; ok {
		formatter(tok, formattedQuery)
	} else {
		f.formatDefaultToken(tok, formattedQuery)
	}
}

func (f *formatter) formatReservedTopLevelToken(tok types.Token, formattedQuery *strings.Builder) {
	f.formatTopLevelReservedWord(tok, formattedQuery)
	f.previousReservedWord = tok
}

func (f *formatter) formatReservedTopLevelNoIndentToken(tok types.Token, formattedQuery *strings.Builder) {
	f.formatTopLevelReservedWordNoIndent(tok, formattedQuery)
	f.previousReservedWord = tok
}

func (f *formatter) formatReservedNewlineToken(tok types.Token, formattedQuery *strings.Builder) {
	f.formatNewlineReservedWord(tok, formattedQuery)
	f.previousReservedWord = tok
}

func (f *formatter) formatReservedToken(tok types.Token, formattedQuery *strings.Builder) {
	f.formatWithSpaces(tok, formattedQuery)
	f.previousReservedWord = tok
}

func (f *formatter) formatWordOrPlaceholder(tok types.Token, formattedQuery *strings.Builder) {
	switch {
	case f.nextToken().Type == types.TokenTypePlaceholder:
		formattedQuery.WriteString(tok.Value)
	case tok.Type == types.TokenTypePlaceholder:
		f.formatPlaceholder(tok, formattedQuery)
	default:
		f.formatWithSpaces(tok, formattedQuery)
	}
}

func (f *formatter) formatDefaultToken(tok types.Token, formattedQuery *strings.Builder) {
	switch tok.Value {
	case ",":
		f.formatComma(tok, formattedQuery)
	case ":":
		f.formatWithSpaceAfter(tok, formattedQuery)
	case ".":
		f.formatWithoutSpaceAfter(tok, formattedQuery)
	case ";":
		f.formatQuerySeparator(tok, formattedQuery)
	default:
		f.formatWithSpaces(tok, formattedQuery)
	}
}

func (f *formatter) formatLineComment(tok types.Token, query *strings.Builder) {
	value := tok.Value
	value = utils.AddANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
	query.WriteString(value)
	f.addNewline(query)
}

func (f *formatter) formatBlockComment(tok types.Token, query *strings.Builder) {
	value := tok.Value
	value = f.indentComment(value)
	value = utils.AddANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
	f.addNewline(query)
	query.WriteString(value)
	f.addNewline(query)
}

func (f *formatter) indentComment(comment string) string {
	return newLineFollowByWhitespaceRegex.ReplaceAllString(comment, "\n"+f.indentation.GetIndent()+" ")
}

func (f *formatter) formatTopLevelReservedWordNoIndent(tok types.Token, query *strings.Builder) {
	f.indentation.DecreaseTopLevel()
	f.addNewline(query)
	query.WriteString(f.equalizeWhitespace(f.formatReservedWord(tok.Value)))
	f.addNewline(query)
}

func (f *formatter) formatTopLevelReservedWord(tok types.Token, query *strings.Builder) {
	f.indentation.DecreaseTopLevel()
	f.addNewline(query)

	f.indentation.IncreaseTopLevel()
	query.WriteString(f.equalizeWhitespace(f.formatReservedWord(tok.Value)))

	f.addNewline(query)
}

func (f *formatter) formatNewlineReservedWord(tok types.Token, query *strings.Builder) {
	f.addNewline(query)
	query.WriteString(f.equalizeWhitespace(f.formatReservedWord(tok.Value)))
	query.WriteString(" ")
}

// equalizeWhitespace replaces any sequence of whitespace characters with a single space.
func (f *formatter) equalizeWhitespace(s string) string {
	return atLeastOneWhitespaceRegex.ReplaceAllString(s, " ")
}

func (f *formatter) formatOpeningParentheses(tok types.Token, query *strings.Builder) {
	preserveWhitespaceFor := map[types.TokenType]struct{}{
		types.TokenTypeWhitespace:  {},
		types.TokenTypeOpenParen:   {},
		types.TokenTypeLineComment: {},
	}

	if _, ok := preserveWhitespaceFor[f.previousToken().Type]; !ok {
		trimSpacesEnd(query)
	}

	value := tok.Value
	if f.cfg.Uppercase {
		value = strings.ToUpper(value)
	}
	query.WriteString(value)

	f.inlineBlock.BeginIfPossible(f.tokens, f.index)
	if !f.inlineBlock.IsActive() {
		f.indentation.IncreaseBlockLevel()
		f.addNewline(query)
	}
}

// formatClosingParentheses ends an inline block if one is active, or decreases the
// block level, then adds the closing paren.
func (f *formatter) formatClosingParentheses(tok types.Token, query *strings.Builder) {
	if f.cfg.Uppercase {
		tok.Value = strings.ToUpper(tok.Value)
	}

	if f.inlineBlock.IsActive() {
		f.inlineBlock.End()
		f.formatWithSpaceAfter(tok, query)
	} else {
		f.indentation.DecreaseBlockLevel()
		f.addNewline(query)
		f.formatWithSpaces(tok, query)
	}
}

// formatPlaceholder formats a placeholder by replacing it with a param value
// from the cfg params and adds a space after.
func (f *formatter) formatPlaceholder(tok types.Token, query *strings.Builder) {
	query.WriteString(f.params.Get(tok.Key, tok.Value))
	query.WriteString(" ")
}

// formatComma adds the comma to the query and adds a space. If an inline block
// is not active, it will add a new line too.
func (f *formatter) formatComma(tok types.Token, query *strings.Builder) {
	trimSpacesEnd(query)
	query.WriteString(tok.Value)
	query.WriteString(" ")

	if f.inlineBlock.IsActive() {
		return
	}
	if limitKeywordRegex.MatchString(f.previousReservedWord.Value) {
		// avoids creating new lines after LIMIT keyword so that two limit items appear on one line for nicer formatting
		return
	} else {
		f.addNewline(query)
	}
}

// formatWithSpaceAfter returns the query with spaces trimmed off the end,
// the types.Token value, and a space (" ") at the end ("query value ").
func (f *formatter) formatWithSpaceAfter(tok types.Token, query *strings.Builder) {
	trimSpacesEnd(query)
	query.WriteString(tok.Value)
	query.WriteString(" ")
}

// formatWithoutSpaceAfter returns the query with spaces trimmed off the end and
// the types.Token value ("query value").
func (f *formatter) formatWithoutSpaceAfter(tok types.Token, query *strings.Builder) {
	trimSpacesEnd(query)
	query.WriteString(tok.Value)
}

// formatWithSpaces returns the query with the value and a space added, plus
// a few more special formatting items.
func (f *formatter) formatWithSpaces(tok types.Token, query *strings.Builder) {
	value := tok.Value
	if tok.Type == types.TokenTypeReserved {
		value = f.formatReservedWord(tok.Value)
	}

	next := f.nextToken()
	if tok.Type == types.TokenTypeWord && !next.Empty() && next.Value == "(" {
		value = utils.AddANSIFormats(f.cfg.ColorConfig.FunctionCallFormatOptions, value)
	}

	query.WriteString(value)
	query.WriteString(" ")
}

// formatReservedWord makes sure the reserved word is formatted according to the Config.
func (f *formatter) formatReservedWord(value string) string {
	if f.cfg.Uppercase {
		value = strings.ToUpper(value)
	}
	value = utils.AddANSIFormats(f.cfg.ColorConfig.ReservedWordFormatOptions, value)
	return value
}

func (f *formatter) formatQuerySeparator(tok types.Token, query *strings.Builder) {
	f.indentation.ResetIndentation()
	trimSpacesEnd(query)
	query.WriteString(tok.Value)
	query.WriteString(strings.Repeat("\n", f.cfg.LinesBetweenQueries))
}

func (f *formatter) formatString(tok types.Token, query *strings.Builder) {
	value := tok.Value
	value = utils.AddANSIFormats(f.cfg.ColorConfig.StringFormatOptions, value)
	query.WriteString(value)
	query.WriteString(" ")
}

func (f *formatter) formatNumber(tok types.Token, query *strings.Builder) {
	value := tok.Value
	value = utils.AddANSIFormats(f.cfg.ColorConfig.NumberFormatOptions, value)
	query.WriteString(value)
	query.WriteString(" ")
}

func (f *formatter) formatBoolean(tok types.Token, query *strings.Builder) {
	value := tok.Value
	value = utils.AddANSIFormats(f.cfg.ColorConfig.BooleanFormatOptions, value)
	query.WriteString(value)
	query.WriteString(" ")
}

func (f *formatter) formatSpecialOperator(tok types.Token, query *strings.Builder) {
	// Special operators like :: (type cast) should be formatted without spaces
	trimSpacesEnd(query)
	query.WriteString(tok.Value)
}

// addNewline trims spaces from the end of query, adds a new line character if
// one does not already exist at the end, and adds the indentation to the new
// line.
func (f *formatter) addNewline(query *strings.Builder) {
	trimSpacesEnd(query)
	if !strings.HasSuffix(query.String(), "\n") {
		query.WriteString("\n")
	}
	query.WriteString(f.indentation.GetIndent())
}

// previousToken peeks at the previous types.Token in the formatters list of types.Tokens with
// the given offset. If no offset is provided, a default of 1 is used.
func (f *formatter) previousToken(offset ...int) types.Token {
	o := 1
	if len(offset) > 0 {
		o = offset[0]
	}
	if f.index-o < 0 {
		return types.Token{} // return an empty types.Token struct
	}
	return f.tokens[f.index-o]
}

// nextToken peeks at the next types.Token in the formatters list of types.Tokens with
// the given offset. If no offset is provided, a default of 1 is used. If
// there is no next types.Token, it returns an empty types.Token.
func (f *formatter) nextToken(offset ...int) types.Token {
	o := 1
	if len(offset) > 0 {
		o = offset[0]
	}
	if f.index+o >= len(f.tokens) {
		return types.Token{}
	}
	return f.tokens[f.index+o]
}
