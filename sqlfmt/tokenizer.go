package sqlfmt

import (
	"regexp"
	"strings"
)

type tokenizer struct {
	whitespaceRegex               *regexp.Regexp
	numberRegex                   *regexp.Regexp
	operatorRegex                 *regexp.Regexp
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

func newTokenizer(cfg TokenizerConfig) *tokenizer {
	return &tokenizer{
		whitespaceRegex:               regexp.MustCompile(`^(\s+)`),
		numberRegex:                   regexp.MustCompile(`^((-\s*)?[0-9]+(\.[0-9]+)?|0x[0-9a-fA-F]+|0b[01]+)\b`),
		operatorRegex:                 regexp.MustCompile(`^(!=|<>|==|<=|>=|=>|!<|!>|\|\||::|->>|->|~~\*|~~|!~~\*|!~~|~\*|!~\*|!~|.)`),
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
	return regexp.MustCompile(`^(` + pattern + `)\b`)
}

func createWordRegex(specialChars []string) *regexp.Regexp {
	// `^([\pL\pM\pN\pPc\pCf` nope
	// `^([\pL\pM\pN\pPc` nope
	// `^([\pL\pM\pN` yep
	// `^([\pL\pM\pN\pCf` yep
	specialVariableChars := regexp.QuoteMeta(`_@'"[]$:` + "`")
	pattern := `^([\pL\pM\pN\pCf` + specialVariableChars + strings.Join(specialChars, ``) + `]+)`
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
	return regexp.MustCompile(`^(` + strings.Join(patterns, `|`) + `)`)
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
	typesRegex := strings.Join(types, `|`)
	return regexp.MustCompile(`^((?:` + escapeRegExp(typesRegex) + `)(?:` + escapeRegExp(pattern) + `))`)
}

func (t *tokenizer) tokenize(input string) []token {
	var (
		tok  token
		toks []token
	)
	for len(input) > 0 {
		//fmt.Println(
		//	"getNextToken out",
		//	"output tok:", t.getNextToken(input, tok),
		//	"input:", input,
		//)
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

// EscapeRegExp escapes special characters in a string for use in a regular expression
func escapeRegExp(s string) string {
	// Special characters to be escaped in regular expressions
	re := regexp.MustCompile(`[.*+?^${}()|[\]\\]`)
	return re.ReplaceAllString(s, `\$0`)
}

func (t *tokenizer) getEscapedPlaceholderKey(key string, quoteChar string) string {
	// Create a regex to match the quote character
	escapedQuote := escapeRegExp("\\" + quoteChar)
	re := regexp.MustCompile(escapedQuote)

	// Replace the quoteChar with quoteChar
	return re.ReplaceAllString(key, quoteChar)
}

func (t *tokenizer) getNumberToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeNumber, t.numberRegex)
}

// // Punctuation and symbols
//
//	getOperatorToken(input) {
//	  return this.getTokenOnFirstMatch({
//	    input,
//	    type: tokenTypes.OPERATOR,
//	    regex: this.OPERATOR_REGEX
//	  });
//	}
func (t *tokenizer) getOperatorToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeOperator, t.operatorRegex)
}

//	getReservedWordToken(input, previousToken) {
//	  // A reserved word cannot be preceded by a "."
//	  // this makes it so in "my_table.from", "from" is not considered a reserved word
//	  if (previousToken && previousToken.value && previousToken.value === '.') {
//	    return;
//	  }
//	  return (
//	    this.getTopLevelReservedToken(input) ||
//	    this.getNewlineReservedToken(input) ||
//	    this.getTopLevelReservedTokenNoIndent(input) ||
//	    this.getPlainReservedToken(input)
//	  );
//	}
func (t *tokenizer) getReservedWordToken(input string, prevTok token) token {
	// A reserved word cannot be preceded by a "."
	// this makes it so in "my_table.from", "from" is not considered a reserved word
	if !prevTok.empty() && prevTok.value == "." {
		return token{} // TODO: this return value might be wrong
	}

	return firstNonEmptyToken(
		t.getTopLevelReservedToken(input),
		t.getNewlineReservedToken(input),
		t.getTopLevelReservedTokenNoIndent(input),
		t.getPlainReservedToken(input),
	)
}

//	getTopLevelReservedToken(input) {
//	  return this.getTokenOnFirstMatch({
//	    input,
//	    type: tokenTypes.RESERVED_TOP_LEVEL,
//	    regex: this.RESERVED_TOP_LEVEL_REGEX
//	  });
//	}
func (t *tokenizer) getTopLevelReservedToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeReservedTopLevel, t.reservedTopLevelRegex)
}

//	getNewlineReservedToken(input) {
//	  return this.getTokenOnFirstMatch({
//	    input,
//	    type: tokenTypes.RESERVED_NEWLINE,
//	    regex: this.RESERVED_NEWLINE_REGEX
//	  });
//	}
func (t *tokenizer) getNewlineReservedToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeReservedNewline, t.reservedNewlineRegex)
}

//	getTopLevelReservedTokenNoIndent(input) {
//	  return this.getTokenOnFirstMatch({
//	    input,
//	    type: tokenTypes.RESERVED_TOP_LEVEL_NO_INDENT,
//	    regex: this.RESERVED_TOP_LEVEL_NO_INDENT_REGEX
//	  });
//	}
func (t *tokenizer) getTopLevelReservedTokenNoIndent(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeReservedTopLevelNoIndent, t.reservedTopLevelNoIndentRegex)
}

//	getPlainReservedToken(input) {
//	  return this.getTokenOnFirstMatch({
//	    input,
//	    type: tokenTypes.RESERVED,
//	    regex: this.RESERVED_PLAIN_REGEX
//	  });
//	}
func (t *tokenizer) getPlainReservedToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeReserved, t.reservedPlainRegex)
}

//	getWordToken(input) {
//	  return this.getTokenOnFirstMatch({
//	    input,
//	    type: tokenTypes.WORD,
//	    regex: this.WORD_REGEX
//	  });
//	}
func (t *tokenizer) getWordToken(input string) token {
	return t.getTokenOnFirstMatch(input, tokenTypeWord, t.wordRegex)
}

//	getTokenOnFirstMatch({ input, type, regex }) {
//	  const matches = input.match(regex);
//
//	  if (matches) {
//	    return { type, value: matches[1] };
//	  }
//	}
func (t *tokenizer) getTokenOnFirstMatch(input string, typ tokenType, re *regexp.Regexp) token {
	matches := re.FindStringSubmatch(input)
	//fmt.Println("getTokenOnFirstMatch", "matches:", matches, "typ:", typ, "input:", input)

	if len(matches) > 0 {
		return token{typ: typ, value: matches[0]} // TODO: might be matches[1]
	}

	return token{}
}

func firstNonEmptyToken(toks ...token) token {
	for _, tok := range toks {
		if !tok.empty() {
			return tok
		}
	}
	return token{}
}
