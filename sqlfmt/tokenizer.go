package sqlfmt

import (
	"regexp"
	"strings"
)

type tokenizer struct {
	whitespaceRegex               *regexp.Regexp
	numberRegex                   *regexp.Regexp
	operatorRegex                 *regexp.Regexp
	booleanRegex                  *regexp.Regexp
	functionCallRegex             *regexp.Regexp
	blockCommentRegex             *regexp.Regexp
	lineCommentRegex              *regexp.Regexp
	reservedTopLevelRegex         *regexp.Regexp
	reservedTopLevelNoIndentRegex *regexp.Regexp
	reservedNewlineRegex          *regexp.Regexp
	reservedPlainRegex            *regexp.Regexp
	wordRegex                     *regexp.Regexp
	stringRegex                   *regexp.Regexp
	openParenRegex                *regexp.Regexp
	closeParenRegex               *regexp.Regexp
	indexedPlaceholderRegex       *regexp.Regexp
	identNamedPlaceholderRegex    *regexp.Regexp
	stringNamedPlaceholderRegex   *regexp.Regexp
}

func newTokenizer(cfg *TokenizerConfig) *tokenizer {
	return &tokenizer{
		whitespaceRegex:               regexp.MustCompile(`^(\s+)`),
		numberRegex:                   regexp.MustCompile(`^((-\s*)?[0-9]+(\.[0-9]+)?|0x[0-9a-fA-F]+|0b[01]+)\b`),
		operatorRegex:                 regexp.MustCompile(`^(!=|<>|==|<=|>=|=>|!<|!>|\|\||::|->>|->|~~\*|~~|!~~\*|!~~|~\*|!~\*|!~|.)`),
		booleanRegex:                  regexp.MustCompile(`(?i)^(\b(true|false)\b)`),
		functionCallRegex:             regexp.MustCompile(`(?i)^(\b(\w+)\s*\(([^)]*)\))`),
		blockCommentRegex:             regexp.MustCompile(`^(/\*(?s:.)*?(?:\*/|$))`),
		lineCommentRegex:              createLineCommentRegex(cfg.LineCommentTypes),
		reservedTopLevelRegex:         createReservedWordRegex(cfg.ReservedTopLevelWords),
		reservedTopLevelNoIndentRegex: createReservedWordRegex(cfg.ReservedTopLevelWordsNoIndent),
		reservedNewlineRegex:          createReservedWordRegex(cfg.ReservedNewlineWords),
		reservedPlainRegex:            createReservedWordRegex(cfg.ReservedWords),
		wordRegex:                     createWordRegex(cfg.SpecialWordChars),
		stringRegex:                   createStringRegex(cfg.StringTypes),
		openParenRegex:                createParenRegex(cfg.OpenParens),
		closeParenRegex:               createParenRegex(cfg.CloseParens),
		indexedPlaceholderRegex:       createPlaceholderRegex(cfg.IndexedPlaceholderTypes, `[0-9]*`),
		identNamedPlaceholderRegex:    createPlaceholderRegex(cfg.NamedPlaceholderTypes, `[a-zA-Z0-9._$]+`),
		stringNamedPlaceholderRegex:   createPlaceholderRegex(cfg.NamedPlaceholderTypes, createStringPattern(cfg.StringTypes)),
	}
}

func createLineCommentRegex(lineCommentTypes []string) *regexp.Regexp {
	pattern := `^((?:` + strings.Join(lineCommentTypes, `|`) + `).*?(?:\r\n|\r|\n|$))`
	return regexp.MustCompile(pattern)
}

func createReservedWordRegex(reservedWords []string) *regexp.Regexp {
	pattern := strings.Join(reservedWords, `|`)
	pattern = strings.ReplaceAll(pattern, " ", `\s+`)
	return regexp.MustCompile(`(?i)^(` + pattern + `)\b`)
}

func createWordRegex(specialChars []string) *regexp.Regexp {
	specialVariableChars := regexp.QuoteMeta(`_@'"[]$?` + "`")
	// `\pPc` was removed from the regex because it was matching ")" like in "TEXT);"
	// `\pCf` was removed from the regex because it was matching "\n" and such
	pattern := `^([\pL\pM\pN` + specialVariableChars + strings.Join(specialChars, ``) + `]+)`
	return regexp.MustCompile(pattern)
}

func createStringRegex(stringTypes []string) *regexp.Regexp {
	pattern := `^(` + createStringPattern(stringTypes) + `)`
	return regexp.MustCompile(pattern)
}

func createStringPattern(stringTypes []string) string {
	patterns := map[string]string{
		"``":   "((`[^`]*($|`))+)",
		"[]":   "((\\[[^\\]]*($|\\]))(\\][^\\]]*($|\\]))*)",
		"\"\"": "((\"[^\"\\\\]*(?:\\\\.[^\"\\\\]*)*(\"|$))+)",
		"''":   "(('[^'\\\\]*(?:\\\\.[^'\\\\]*)*('|$))+)",
		"N''":  "((N'[^N'\\\\]*(?:\\\\.[^N'\\\\]*)*('|$))+)",
		"$$":   "((\\$\\$[^\\$]*($|\\$\\$))+)",
	}
	var result []string
	for _, t := range stringTypes {
		result = append(result, patterns[t])
	}
	return strings.Join(result, "|")
}

func createParenRegex(parens []string) *regexp.Regexp {
	patterns := make([]string, len(parens))
	for i, p := range parens {
		patterns[i] = escapeParen(p)
	}
	return regexp.MustCompile(`(?i)^(` + strings.Join(patterns, `|`) + `)`)
}

func escapeParen(paren string) string {
	if len(paren) == 1 {
		return regexp.QuoteMeta(paren)
	} else {
		return `\b` + paren + `\b`
	}
}

func createPlaceholderRegex(types []string, pattern string) *regexp.Regexp {
	if len(types) == 0 {
		return nil
	}
	esc := make([]string, 0, len(types))
	for _, t := range types {
		esc = append(esc, regexp.QuoteMeta(t))
	}
	typesRegex := strings.Join(esc, `|`)
	return regexp.MustCompile(`^((?:` + typesRegex + `)(?:` + pattern + `))`)
}

func (t *tokenizer) tokenize(input string) []token {
	var (
		tok  token
		toks []token
	)
	for len(input) > 0 {
		tok = t.getNextToken(input, tok)
		input = input[len(tok.value):]
		toks = append(toks, tok)
	}
	return toks
}

func (t *tokenizer) getNextToken(input string, prevTok token) token {
	return firstNonEmptyToken(
		t.getWhitespaceToken(input),
		t.getCommentToken(input),
		t.getStringToken(input),
		t.getOpenParenToken(input),
		t.getCloseParenToken(input),
		t.getPlaceholderToken(input),
		t.getNumberToken(input),
		t.getReservedWordToken(input, prevTok),
		t.getBooleanToken(input),
		t.getWordToken(input),
		t.getOperatorToken(input),
	)
}

func (t *tokenizer) getWhitespaceToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeWhitespace, t.whitespaceRegex)
}

func (t *tokenizer) getCommentToken(input string) token {
	tok := t.getLineCommentToken(input)
	if !tok.empty() {
		return tok
	}
	return t.getBlockCommentToken(input)
}

func (t *tokenizer) getLineCommentToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeLineComment, t.lineCommentRegex)
}

func (t *tokenizer) getBlockCommentToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeBlockComment, t.blockCommentRegex)
}

func (t *tokenizer) getStringToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeString, t.stringRegex)
}

func (t *tokenizer) getOpenParenToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeOpenParen, t.openParenRegex)
}

func (t *tokenizer) getCloseParenToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeCloseParen, t.closeParenRegex)
}

func (t *tokenizer) getPlaceholderToken(input string) token {
	return firstNonEmptyToken(
		t.getIdentNamedPlaceholderToken(input),
		t.getStringNamedPlaceholderToken(input),
		t.getIndexedPlaceholderToken(input),
	)
}

func (t *tokenizer) getIdentNamedPlaceholderToken(input string) token {
	tok := t.getTokenOnFirstMatch(input, tokenTypePlaceholder, t.identNamedPlaceholderRegex)
	if tok.value != "" {
		tok.key = tok.value[1:] // Remove the first character
	}
	return tok
}

func (t *tokenizer) getStringNamedPlaceholderToken(input string) token {
	tok := t.getTokenOnFirstMatch(input, tokenTypePlaceholder, t.stringNamedPlaceholderRegex)
	if tok.value != "" {
		l := len(tok.value)
		tok.key = t.getEscapedPlaceholderKey(tok.value[2:l-1], tok.value[l-1:])
	}
	return tok
}

func (t *tokenizer) getIndexedPlaceholderToken(input string) token {
	tok := t.getTokenOnFirstMatch(input, tokenTypePlaceholder, t.indexedPlaceholderRegex)
	if tok.value != "" {
		// Remove the first character so ?2 becomes 2
		tok.key = tok.value[1:]
	}
	return tok
}

func (t *tokenizer) getEscapedPlaceholderKey(key string, quoteChar string) string {
	// Create a regex to match the quote character
	escapedQuote := regexp.QuoteMeta("\\" + quoteChar)
	re := regexp.MustCompile(escapedQuote)

	// Replace the quoteChar with quoteChar
	return re.ReplaceAllString(key, quoteChar)
}

func (t *tokenizer) getNumberToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeNumber, t.numberRegex)
}

func (t *tokenizer) getOperatorToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeOperator, t.operatorRegex)
}

func (t *tokenizer) getReservedWordToken(input string, prevTok token) token {
	// A reserved word cannot be preceded by a "."
	// this makes it so in "my_table.from", "from" is not considered a reserved word
	if !prevTok.empty() && prevTok.value == "." {
		return token{}
	}

	return firstNonEmptyToken(
		t.getTopLevelReservedToken(input),
		t.getNewlineReservedToken(input),
		t.getTopLevelReservedTokenNoIndent(input),
		t.getPlainReservedToken(input),
	)
}

func (t *tokenizer) getTopLevelReservedToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeReservedTopLevel, t.reservedTopLevelRegex)
}

func (t *tokenizer) getNewlineReservedToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeReservedNewline, t.reservedNewlineRegex)
}

func (t *tokenizer) getTopLevelReservedTokenNoIndent(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeReservedTopLevelNoIndent, t.reservedTopLevelNoIndentRegex)
}

func (t *tokenizer) getPlainReservedToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeReserved, t.reservedPlainRegex)
}

func (t *tokenizer) getBooleanToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeBoolean, t.booleanRegex)
}

func (t *tokenizer) getWordToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeWord, t.wordRegex)
}

// getTokenOnFirstMatch uses the regex re to search for string submatches in input.
// If one or more submatches are found, the first one is returned in a new token with
// the token type typ as the tokenType.
func (t *tokenizer) getTokenOnFirstMatch(input string, typ tokenType, re *regexp.Regexp) token {
	if re == nil {
		return token{}
	}

	matches := re.FindStringSubmatch(input)

	if len(matches) > 0 {
		return token{typ: typ, value: matches[0]}
	}

	return token{}
}

// firstNonEmptyToken returns the first token in the list of given tokens, toks,
// that is not empty.
func firstNonEmptyToken(toks ...token) token {
	for _, tok := range toks {
		if !tok.empty() {
			return tok
		}
	}
	return token{}
}
