package core

import (
	"regexp"
	"sort"
	"strings"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
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
	regex := `^(!=|<>|<=>|==|<=|>=|=>|!<|!>|\|\||::|->>|->|#>>|#>|<<|>>|` +
		`\?\||\?&|\?|@>|<@|~~\*|~~|!~~\*|!~~|~\*|!~\*|!~|.)`
	return &tokenizer{
		whitespaceRegex:               regexp.MustCompile(`^(\s+)`),
		numberRegex:                   regexp.MustCompile(`^((-\s*)?[0-9]+(\.[0-9]+)?|0x[0-9a-fA-F]+|0b[01]+)\b`),
		operatorRegex:                 regexp.MustCompile(regex),
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
		stringNamedPlaceholderRegex: createPlaceholderRegex(
			cfg.NamedPlaceholderTypes, createStringPattern(cfg.StringTypes)),
	}
}

func createLineCommentRegex(lineCommentTypes []string) *regexp.Regexp {
	pattern := `^((?:` + strings.Join(lineCommentTypes, `|`) + `).*?(?:\r\n|\r|\n|$))`
	return regexp.MustCompile(pattern)
}

func createReservedWordRegex(reservedWords []string) *regexp.Regexp {
	// Sort reserved words by length in descending order. This is crucial for the tokenizer
	// to prioritize longer matches, like "DO UPDATE" over "DO".
	sort.Slice(reservedWords, func(i, j int) bool {
		return len(reservedWords[i]) > len(reservedWords[j])
	})
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
		"X''":  "(((?i)[Xx]'[0-9a-fA-F]*($|'))+)", // Hex blob literals
		"B''":  "(((?i)[Bb]'[01]*($|'))+)",        // Binary literals
		"$$":   "((\\$\\$[^\\$]*($|\\$\\$))+)",
	}
	result := make([]string, 0, len(stringTypes))
	for _, t := range stringTypes {
		result = append(result, patterns[t])
	}
	return strings.Join(result, "|")
}

func createParenRegex(parens []string) *regexp.Regexp {
	// Sort by length descending to prioritize longer matches (e.g., "END IF" before "END")
	sorted := make([]string, len(parens))
	copy(sorted, parens)
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i]) > len(sorted[j])
	})

	patterns := make([]string, len(sorted))
	for i, p := range sorted {
		patterns[i] = escapeParen(p)
	}
	return regexp.MustCompile(`(?i)^(` + strings.Join(patterns, `|`) + `)`)
}

func escapeParen(paren string) string {
	if len(paren) == 1 {
		return regexp.QuoteMeta(paren)
	} else {
		// For multi-word keywords, escape spaces and use word boundaries
		// This ensures "END IF" is treated as a unit and not matched by just "END"
		escaped := regexp.QuoteMeta(paren)
		return `\b` + escaped + `\b`
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

func (t *tokenizer) tokenize(input string) []types.Token {
	var (
		tok  types.Token
		toks []types.Token
	)
	for len(input) > 0 {
		tok = t.getNextToken(input, tok)
		input = input[len(tok.Value):]
		toks = append(toks, tok)
	}
	return toks
}

func (t *tokenizer) getNextToken(input string, prevTok types.Token) types.Token {
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

func (t *tokenizer) getWhitespaceToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeWhitespace, t.whitespaceRegex)
}

func (t *tokenizer) getCommentToken(input string) types.Token {
	tok := t.getLineCommentToken(input)
	if !tok.Empty() {
		return tok
	}
	return t.getBlockCommentToken(input)
}

func (t *tokenizer) getLineCommentToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeLineComment, t.lineCommentRegex)
}

func (t *tokenizer) getBlockCommentToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeBlockComment, t.blockCommentRegex)
}

func (t *tokenizer) getStringToken(input string) types.Token {
	// Check for dollar-quoted strings first as they require special handling
	if dollarQuotedToken := t.getDollarQuotedToken(input); !dollarQuotedToken.Empty() {
		return dollarQuotedToken
	}
	return t.getTokenOnFirstMatch(input, types.TokenTypeString, t.stringRegex)
}

// getDollarQuotedToken scans for PostgreSQL-style dollar-quoted strings.
func (t *tokenizer) getDollarQuotedToken(input string) types.Token {
	return scanDollarQuotedString(input)
}

func (t *tokenizer) getOpenParenToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeOpenParen, t.openParenRegex)
}

func (t *tokenizer) getCloseParenToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeCloseParen, t.closeParenRegex)
}

func (t *tokenizer) getPlaceholderToken(input string) types.Token {
	return firstNonEmptyToken(
		t.getIdentNamedPlaceholderToken(input),
		t.getStringNamedPlaceholderToken(input),
		t.getIndexedPlaceholderToken(input),
	)
}

func (t *tokenizer) getIdentNamedPlaceholderToken(input string) types.Token {
	// Don't match @ if it's part of @> or <@ JSON operators
	if len(input) >= 2 && input[0] == '@' && input[1] == '>' {
		return types.Token{}
	}
	if len(input) >= 2 && input[0] == '<' && input[1] == '@' {
		return types.Token{}
	}

	tok := t.getTokenOnFirstMatch(input, types.TokenTypePlaceholder, t.identNamedPlaceholderRegex)
	if tok.Value != "" {
		tok.Key = tok.Value[1:] // Remove the first character
	}
	return tok
}

func (t *tokenizer) getStringNamedPlaceholderToken(input string) types.Token {
	if t.shouldSkipStringNamedPlaceholder(input) {
		return types.Token{}
	}

	tok := t.getTokenOnFirstMatch(input, types.TokenTypePlaceholder, t.stringNamedPlaceholderRegex)
	if tok.Value != "" {
		l := len(tok.Value)
		tok.Key = t.getEscapedPlaceholderKey(tok.Value[2:l-1], tok.Value[l-1:])
	}
	return tok
}

func (t *tokenizer) shouldSkipStringNamedPlaceholder(input string) bool {
	if len(input) < 2 {
		return false
	}
	if input[0] == '@' && (input[1] == '>' || input[1] == '<' && len(input) > 2 && input[2] == '@') {
		return true
	}
	if input[0] == '?' && (input[1] == '|' || input[1] == '&') {
		return true
	}
	return false
}

func (t *tokenizer) getIndexedPlaceholderToken(input string) types.Token {
	// Don't match ? if it's part of ?|, ?& JSON existence operators
	if len(input) >= 2 && input[0] == '?' && (input[1] == '|' || input[1] == '&') {
		return types.Token{}
	}

	tok := t.getTokenOnFirstMatch(input, types.TokenTypePlaceholder, t.indexedPlaceholderRegex)
	if tok.Value != "" {
		// Remove the first character so ?2 becomes 2
		tok.Key = tok.Value[1:]
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

func (t *tokenizer) getNumberToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeNumber, t.numberRegex)
}

func (t *tokenizer) getOperatorToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeOperator, t.operatorRegex)
}

func (t *tokenizer) getReservedWordToken(input string, prevTok types.Token) types.Token {
	// A reserved word cannot be preceded by a "."
	// this makes it so in "my_table.from", "from" is not considered a reserved word
	if !prevTok.Empty() && prevTok.Value == "." {
		return types.Token{}
	}

	return firstNonEmptyToken(
		t.getTopLevelReservedToken(input),
		t.getNewlineReservedToken(input),
		t.getTopLevelReservedTokenNoIndent(input),
		t.getPlainReservedToken(input),
	)
}

func (t *tokenizer) getTopLevelReservedToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeReservedTopLevel, t.reservedTopLevelRegex)
}

func (t *tokenizer) getNewlineReservedToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeReservedNewline, t.reservedNewlineRegex)
}

func (t *tokenizer) getTopLevelReservedTokenNoIndent(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeReservedTopLevelNoIndent, t.reservedTopLevelNoIndentRegex)
}

func (t *tokenizer) getPlainReservedToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeReserved, t.reservedPlainRegex)
}

func (t *tokenizer) getBooleanToken(input string) types.Token {
	return t.getTokenOnFirstMatch(input, types.TokenTypeBoolean, t.booleanRegex)
}

func (t *tokenizer) getWordToken(input string) types.Token {
	if t.shouldSkipWord(input) {
		return types.Token{}
	}

	tok := t.getTokenOnFirstMatch(input, types.TokenTypeWord, t.wordRegex)

	// Additional check: if we matched a single @ or ?, and it's followed by operator chars, skip
	if t.shouldSkipMatchedWord(tok, input) {
		return types.Token{}
	}

	return tok
}

func (t *tokenizer) shouldSkipWord(input string) bool {
	if len(input) < 2 {
		return false
	}
	if input[0] == '@' && input[1] == '>' {
		return true
	}
	if input[0] == '?' && (input[1] == '|' || input[1] == '&') {
		return true
	}
	return false
}

func (t *tokenizer) shouldSkipMatchedWord(tok types.Token, input string) bool {
	if len(input) < 2 {
		return false
	}
	if tok.Value == "@" && input[1] == '>' {
		return true
	}
	if tok.Value == "?" && (input[1] == '|' || input[1] == '&') {
		return true
	}
	return false
}

// getTokenOnFirstMatch uses the regex re to search for string submatches in input.
// If one or more submatches are found, the first one is returned in a new types.Token with
// the types.Token type typ as the types.TokenType.
func (t *tokenizer) getTokenOnFirstMatch(input string, typ types.TokenType, re *regexp.Regexp) types.Token {
	if re == nil {
		return types.Token{}
	}

	matches := re.FindStringSubmatch(input)

	if len(matches) > 0 {
		return types.Token{Type: typ, Value: matches[0]}
	}

	return types.Token{}
}

// firstNonEmptyToken returns the first types.Token in the list of given types.Tokens, toks,
// that is not empty.
func firstNonEmptyToken(toks ...types.Token) types.Token {
	for _, tok := range toks {
		if !tok.Empty() {
			return tok
		}
	}
	return types.Token{}
}

// scanDollarQuotedString scans for PostgreSQL-style dollar-quoted strings.
// Supports both simple $$ delimited strings and tagged $tag$ delimited strings.
func scanDollarQuotedString(input string) types.Token {
	if len(input) == 0 || input[0] != '$' {
		return types.Token{}
	}

	openingTag := findDollarQuoteTag(input)
	if openingTag == "" {
		return types.Token{}
	}

	return findClosingDollarQuote(input, openingTag)
}

// findDollarQuoteTag extracts the opening dollar-quote tag from input.
func findDollarQuoteTag(input string) string {
	for i := 1; i < len(input); i++ {
		if input[i] == '$' {
			return input[:i+1]
		}
		if !isValidTagChar(input[i]) {
			return ""
		}
	}
	return ""
}

// findClosingDollarQuote searches for the matching closing tag.
func findClosingDollarQuote(input, openingTag string) types.Token {
	searchStart := len(openingTag)
	for i := searchStart; i <= len(input)-len(openingTag); i++ {
		if hasMatchingTag(input, i, openingTag) {
			return types.Token{
				Type:  types.TokenTypeString,
				Value: input[:i+len(openingTag)],
			}
		}
	}

	// No matching closing tag found, return entire input as incomplete string
	return types.Token{
		Type:  types.TokenTypeString,
		Value: input,
	}
}

// isValidTagChar checks if character is valid in dollar-quote tag.
func isValidTagChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}

// hasMatchingTag checks if the input has the matching closing tag at position i.
func hasMatchingTag(input string, i int, tag string) bool {
	return i+len(tag) <= len(input) && input[i:i+len(tag)] == tag
}
