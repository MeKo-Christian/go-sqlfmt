package sqlfmt

import (
	"regexp"
	"strings"
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
	indentation          *indentation
	inlineBlock          *inlineBlock
	params               *params
	tokenizer            *tokenizer                                        // Assume tokenizer is defined in your code
	tokenOverride        func(tok token, previousReservedWord token) token // Assume token is defined in your code
	previousReservedWord token
	tokens               []token
	index                int
}

// newFormatter creates a new formatter instance.
func newFormatter(cfg *Config, tokenizer *tokenizer, tokenOverride func(tok token, previousReservedWord token) token) *formatter {
	if cfg.ColorConfig == nil {
		cfg.ColorConfig = &ColorConfig{}
	}
	return &formatter{
		cfg:                  cfg,
		indentation:          newIndentation(cfg.Indent),
		inlineBlock:          newInlineBlock(),
		params:               newParams(cfg.Params),
		tokenizer:            tokenizer,
		tokenOverride:        tokenOverride,
		previousReservedWord: token{},
		tokens:               []token{},
		index:                0,
	}
}

// format formats whitespace in a SQL string to make it easier to read.
func (f *formatter) format(query string) string {
	f.tokens = f.tokenizer.tokenize(query)
	formattedQuery := f.getFormattedQueryFromTokens()
	return strings.TrimSpace(formattedQuery)
}

// getFormattedQueryFromTokens processes the tokens to create a formatted query.
func (f *formatter) getFormattedQueryFromTokens() string {
	formattedQuery := &strings.Builder{}

	for i, tok := range f.tokens {
		f.index = i

		if f.tokenOverride != nil {
			tok = f.tokenOverride(tok, f.previousReservedWord)
		}

		switch tok.typ {
		case tokenTypeWhitespace:
			// Ignore whitespace
		case tokenTypeLineComment:
			f.formatLineComment(tok, formattedQuery)
		case tokenTypeBlockComment:
			f.formatBlockComment(tok, formattedQuery)
		case tokenTypeReservedTopLevel:
			f.formatTopLevelReservedWord(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeReservedTopLevelNoIndent:
			f.formatTopLevelReservedWordNoIndent(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeReservedNewline:
			f.formatNewlineReservedWord(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeReserved:
			f.formatWithSpaces(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeOpenParen:
			f.formatOpeningParentheses(tok, formattedQuery)
		case tokenTypeCloseParen:
			f.formatClosingParentheses(tok, formattedQuery)
		case tokenTypeWord, tokenTypePlaceholder:
			if f.nextToken().typ == tokenTypePlaceholder {
				formattedQuery.WriteString(tok.value)
			} else if tok.typ == tokenTypePlaceholder {
				f.formatPlaceholder(tok, formattedQuery)
			} else {
				f.formatWithSpaces(tok, formattedQuery)
			}
		case tokenTypeString:
			f.formatString(tok, formattedQuery)
		case tokenTypeNumber:
			f.formatNumber(tok, formattedQuery)
		case tokenTypeBoolean:
			f.formatBoolean(tok, formattedQuery)
		default:
			switch tok.value {
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
	}
	return formattedQuery.String()
}

func (f *formatter) formatLineComment(tok token, query *strings.Builder) {
	value := tok.value
	value = addANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
	query.WriteString(value)
	f.addNewline(query)
}

func (f *formatter) formatBlockComment(tok token, query *strings.Builder) {
	value := tok.value
	value = f.indentComment(value)
	value = addANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
	f.addNewline(query)
	query.WriteString(value)
	f.addNewline(query)
}

func (f *formatter) indentComment(comment string) string {
	return newLineFollowByWhitespaceRegex.ReplaceAllString(comment, "\n"+f.indentation.getIndent()+" ")
}

func (f *formatter) formatTopLevelReservedWordNoIndent(tok token, query *strings.Builder) {
	f.indentation.decreaseTopLevel()
	f.addNewline(query)
	query.WriteString(f.equalizeWhitespace(f.formatReservedWord(tok.value)))
	f.addNewline(query)
}

func (f *formatter) formatTopLevelReservedWord(tok token, query *strings.Builder) {
	f.indentation.decreaseTopLevel()
	f.addNewline(query)

	f.indentation.increaseTopLevel()
	query.WriteString(f.equalizeWhitespace(f.formatReservedWord(tok.value)))

	f.addNewline(query)
}

func (f *formatter) formatNewlineReservedWord(tok token, query *strings.Builder) {
	f.addNewline(query)
	query.WriteString(f.equalizeWhitespace(f.formatReservedWord(tok.value)))
	query.WriteString(" ")
}

// equalizeWhitespace replaces any sequence of whitespace characters with a single space.
func (f *formatter) equalizeWhitespace(s string) string {
	return atLeastOneWhitespaceRegex.ReplaceAllString(s, " ")
}

func (f *formatter) formatOpeningParentheses(tok token, query *strings.Builder) {
	preserveWhitespaceFor := map[tokenType]struct{}{
		tokenTypeWhitespace:  {},
		tokenTypeOpenParen:   {},
		tokenTypeLineComment: {},
	}

	if _, ok := preserveWhitespaceFor[f.previousToken().typ]; !ok {
		trimSpacesEnd(query)
	}

	value := tok.value
	if f.cfg.Uppercase {
		value = strings.ToUpper(value)
	}
	query.WriteString(value)

	f.inlineBlock.beginIfPossible(f.tokens, f.index)
	if !f.inlineBlock.isActive() {
		f.indentation.increaseBlockLevel()
		f.addNewline(query)
	}
}

// formatClosingParentheses ends an inline block if one is active, or decreases the
// block level, then adds the closing paren.
func (f *formatter) formatClosingParentheses(tok token, query *strings.Builder) {
	if f.cfg.Uppercase {
		tok.value = strings.ToUpper(tok.value)
	}

	if f.inlineBlock.isActive() {
		f.inlineBlock.end()
		f.formatWithSpaceAfter(tok, query)
	} else {
		f.indentation.decreaseBlockLevel()
		f.addNewline(query)
		f.formatWithSpaces(tok, query)
	}
}

// formatPlaceholder formats a placeholder by replacing it with a param value
// from the cfg params and adds a space after.
func (f *formatter) formatPlaceholder(tok token, query *strings.Builder) {
	query.WriteString(f.params.get(tok.key, tok.value))
	query.WriteString(" ")
}

// formatComma adds the comma to the query and adds a space. If an inline block
// is not active, it will add a new line too.
func (f *formatter) formatComma(tok token, query *strings.Builder) {
	trimSpacesEnd(query)
	query.WriteString(tok.value)
	query.WriteString(" ")

	if f.inlineBlock.isActive() {
		return
	}
	if limitKeywordRegex.MatchString(f.previousReservedWord.value) {
		// avoids creating new lines after LIMIT keyword so that two limit items appear on one line for nicer formatting
		return
	} else {
		f.addNewline(query)
	}
}

// formatWithSpaceAfter returns the query with spaces trimmed off the end,
// the token value, and a space (" ") at the end ("query value ")
func (f *formatter) formatWithSpaceAfter(tok token, query *strings.Builder) {
	trimSpacesEnd(query)
	query.WriteString(tok.value)
	query.WriteString(" ")
}

// formatWithoutSpaceAfter returns the query with spaces trimmed off the end and
// the token value ("query value")
func (f *formatter) formatWithoutSpaceAfter(tok token, query *strings.Builder) {
	trimSpacesEnd(query)
	query.WriteString(tok.value)
}

// formatWithSpaces returns the query with the value and a space added, plus
// a few more special formatting items.
func (f *formatter) formatWithSpaces(tok token, query *strings.Builder) {
	value := tok.value
	if tok.typ == tokenTypeReserved {
		value = f.formatReservedWord(tok.value)
	}

	next := f.nextToken()
	if tok.typ == tokenTypeWord && !next.empty() && next.value == "(" {
		value = addANSIFormats(f.cfg.ColorConfig.FunctionCallFormatOptions, value)
	}

	query.WriteString(value)
	query.WriteString(" ")
}

// formatReservedWord makes sure the reserved word is formatted according to the Config.
func (f *formatter) formatReservedWord(value string) string {
	if f.cfg.Uppercase {
		value = strings.ToUpper(value)
	}
	value = addANSIFormats(f.cfg.ColorConfig.ReservedWordFormatOptions, value)
	return value
}

func (f *formatter) formatQuerySeparator(tok token, query *strings.Builder) {
	f.indentation.resetIndentation()
	trimSpacesEnd(query)
	query.WriteString(tok.value)
	query.WriteString(strings.Repeat("\n", f.cfg.LinesBetweenQueries))
}

func (f *formatter) formatString(tok token, query *strings.Builder) {
	value := tok.value
	value = addANSIFormats(f.cfg.ColorConfig.StringFormatOptions, value)
	query.WriteString(value)
	query.WriteString(" ")
}

func (f *formatter) formatNumber(tok token, query *strings.Builder) {
	value := tok.value
	value = addANSIFormats(f.cfg.ColorConfig.NumberFormatOptions, value)
	query.WriteString(value)
	query.WriteString(" ")
}

func (f *formatter) formatBoolean(tok token, query *strings.Builder) {
	value := tok.value
	value = addANSIFormats(f.cfg.ColorConfig.BooleanFormatOptions, value)
	query.WriteString(value)
	query.WriteString(" ")
}

// addNewline trims spaces from the end of query, adds a new line character if
// one does not already exist at the end, and adds the indentation to the new
// line.
func (f *formatter) addNewline(query *strings.Builder) {
	trimSpacesEnd(query)
	if !strings.HasSuffix(query.String(), "\n") {
		query.WriteString("\n")
	}
	query.WriteString(f.indentation.getIndent())
}

// previousToken peeks at the previous token in the formatters list of tokens with
// the given offset. If no offset is provided, a default of 1 is used.
func (f *formatter) previousToken(offset ...int) token {
	o := 1
	if len(offset) > 0 {
		o = offset[0]
	}
	if f.index-o < 0 {
		return token{} // return an empty token struct
	}
	return f.tokens[f.index-o]
}

// nextToken peeks at the next token in the formatters list of tokens with
// the given offset. If no offset is provided, a default of 1 is used. If
// there is no next token, it returns an empty token.
func (f *formatter) nextToken(offset ...int) token {
	o := 1
	if len(offset) > 0 {
		o = offset[0]
	}
	if f.index+o >= len(f.tokens) {
		return token{}
	}
	return f.tokens[f.index+o]
}
