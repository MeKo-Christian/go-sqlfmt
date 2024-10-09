package sqlfmt

import (
	"regexp"
	"strings"
)

// trimSpacesEnd removes trailing spaces and tabs from a string.
func trimSpacesEnd(str string) string {
	return strings.TrimRight(str, " \t")
}

// formatter formats SQL queries for better readability.
type formatter struct {
	cfg                  Config
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
func newFormatter(cfg Config, tokenizer *tokenizer, tokenOverride func(tok token, previousReservedWord token) token) *formatter {
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
	formattedQuery := ""

	for i, tok := range f.tokens {
		f.index = i

		if f.tokenOverride != nil {
			tok = f.tokenOverride(tok, f.previousReservedWord)
		}

		switch tok.typ {
		case tokenTypeWhitespace:
			// Ignore whitespace
		case tokenTypeLineComment:
			formattedQuery = f.formatLineComment(tok, formattedQuery)
		case tokenTypeBlockComment:
			formattedQuery = f.formatBlockComment(tok, formattedQuery)
		case tokenTypeReservedTopLevel:
			formattedQuery = f.formatTopLevelReservedWord(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeReservedTopLevelNoIndent:
			formattedQuery = f.formatTopLevelReservedWordNoIndent(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeReservedNewline:
			formattedQuery = f.formatNewlineReservedWord(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeReserved:
			formattedQuery = f.formatWithSpaces(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeOpenParen:
			formattedQuery = f.formatOpeningParentheses(tok, formattedQuery)
		case tokenTypeCloseParen:
			formattedQuery = f.formatClosingParentheses(tok, formattedQuery)
		case tokenTypeWord, tokenTypePlaceholder:
			if f.nextToken().typ == tokenTypePlaceholder {
				formattedQuery += tok.value
			} else if tok.typ == tokenTypePlaceholder {
				formattedQuery = f.formatPlaceholder(tok, formattedQuery)
			} else {
				formattedQuery = f.formatWithSpaces(tok, formattedQuery)
			}
		default:
			switch tok.value {
			case ",":
				formattedQuery = f.formatComma(tok, formattedQuery)
			case ":":
				formattedQuery = f.formatWithSpaceAfter(tok, formattedQuery)
			case ".":
				formattedQuery = f.formatWithoutSpaces(tok, formattedQuery)
			case ";":
				formattedQuery = f.formatQuerySeparator(tok, formattedQuery)
			default:
				formattedQuery = f.formatWithSpaces(tok, formattedQuery)
			}
		}
	}
	return formattedQuery
}

func (f *formatter) formatLineComment(tok token, query string) string {
	return f.addNewline(query + tok.value)
}

func (f *formatter) formatBlockComment(tok token, query string) string {
	return f.addNewline(f.addNewline(query) + f.indentComment(tok.value))
}

func (f *formatter) indentComment(comment string) string {
	return regexp.MustCompile(`\n[ \t]*`).ReplaceAllString(comment, "\n"+f.indentation.getIndent()+" ")
}

func (f *formatter) formatTopLevelReservedWordNoIndent(tok token, query string) string {
	f.indentation.decreaseTopLevel()
	query = f.addNewline(query) + f.equalizeWhitespace(f.formatReservedWord(tok.value))
	return f.addNewline(query)
}

func (f *formatter) formatTopLevelReservedWord(tok token, query string) string {
	f.indentation.decreaseTopLevel()
	query = f.addNewline(query)
	f.indentation.increaseTopLevel()
	query += f.equalizeWhitespace(f.formatReservedWord(tok.value))
	return f.addNewline(query)
}

func (f *formatter) formatNewlineReservedWord(tok token, query string) string {
	return f.addNewline(query) + f.equalizeWhitespace(f.formatReservedWord(tok.value)) + " "
}

// equalizeWhitespace replaces any sequence of whitespace characters with a single space.
func (f *formatter) equalizeWhitespace(s string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
}

// formatOpeningParentheses increases the block indent level and starts a new line.
func (f *formatter) formatOpeningParentheses(tok token, query string) string {
	preserveWhitespaceFor := map[tokenType]struct{}{
		tokenTypeWhitespace:  {},
		tokenTypeOpenParen:   {},
		tokenTypeLineComment: {},
	}

	if _, ok := preserveWhitespaceFor[f.previousToken().typ]; !ok {
		query = trimSpacesEnd(query)
	}
	query += tok.value
	if f.cfg.Uppercase {
		query = strings.ToUpper(query)
	}

	f.inlineBlock.beginIfPossible(f.tokens, f.index)

	if !f.inlineBlock.isActive() {
		f.indentation.increaseBlockLevel()
		query = f.addNewline(query)
	}
	return query
}

// formatClosingParentheses decreases the block indent level.
func (f *formatter) formatClosingParentheses(tok token, query string) string {
	if f.cfg.Uppercase {
		tok.value = strings.ToUpper(tok.value)
	}

	if f.inlineBlock.isActive() {
		f.inlineBlock.end()
		return f.formatWithSpaceAfter(tok, query)
	} else {
		f.indentation.decreaseBlockLevel()
		return f.formatWithSpaces(tok, f.addNewline(query))
	}
}

func (f *formatter) formatPlaceholder(tok token, query string) string {
	return query + f.params.get(tok.key, tok.value) + " " // TODO: args to get might be wrong here
}

// formatComma formats commas, which start a new line.
func (f *formatter) formatComma(tok token, query string) string {
	query = trimSpacesEnd(query) + tok.value + " "

	if f.inlineBlock.isActive() {
		return query
	} else if matched, _ := regexp.MatchString(`^LIMIT$`, f.previousReservedWord.value); matched {
		return query
	} else {
		return f.addNewline(query)
	}
}

func (f *formatter) formatWithSpaceAfter(tok token, query string) string {
	return trimSpacesEnd(query) + tok.value + " "
}

func (f *formatter) formatWithoutSpaces(tok token, query string) string {
	return trimSpacesEnd(query) + tok.value
}

func (f *formatter) formatWithSpaces(tok token, query string) string {
	value := tok.value
	if tok.typ == tokenTypeReserved {
		value = f.formatReservedWord(tok.value)
	}
	return query + value + " "
}

func (f *formatter) formatReservedWord(value string) string {
	if f.cfg.Uppercase {
		return strings.ToUpper(value)
	}
	return value
}

func (f *formatter) formatQuerySeparator(tok token, query string) string {
	f.indentation.resetIndentation()
	return trimSpacesEnd(query) + tok.value + strings.Repeat("\n", f.cfg.LinesBetweenQueries)
}

func (f *formatter) addNewline(query string) string {
	query = trimSpacesEnd(query)
	if !strings.HasSuffix(query, "\n") {
		query += "\n"
	}
	return query + f.indentation.getIndent()
}

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

func (f *formatter) nextToken(offset ...int) token {
	o := 1
	if len(offset) > 0 {
		o = offset[0]
	}
	if f.index+o >= len(f.tokens) {
		return token{} // return an empty token struct
	}
	return f.tokens[f.index+o]
}
