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
	// Alignment state
	inSelectClause          bool
	selectColumnLengths     []int
	currentColumnLength     int
	currentSelectIndex      int
	inUpdateSetClause       bool
	updateAssignmentLengths []int
	currentAssignmentLength int
	currentUpdateIndex      int
	inInsertValuesClause    bool
	insertValuesLengths     []int
	currentValuesLength     int
	currentInsertIndex      int
	// Line length tracking
	currentLineLength int
}

// newFormatter creates a new formatter instance.
func newFormatter(cfg *Config, tokenizer *tokenizer,
	tokenOverride func(tok types.Token, previousReservedWord types.Token) types.Token,
) *formatter {
	if cfg.ColorConfig == nil {
		cfg.ColorConfig = &ColorConfig{}
	}
	return &formatter{
		cfg:                     cfg,
		indentation:             utils.NewIndentation(cfg.Indent),
		inlineBlock:             utils.NewInlineBlock(),
		params:                  utils.NewParams(cfg.Params),
		tokenizer:               tokenizer,
		tokenOverride:           tokenOverride,
		previousReservedWord:    types.Token{},
		tokens:                  []types.Token{},
		index:                   0,
		inSelectClause:          false,
		selectColumnLengths:     []int{},
		currentColumnLength:     0,
		currentSelectIndex:      0,
		inUpdateSetClause:       false,
		updateAssignmentLengths: []int{},
		currentAssignmentLength: 0,
		currentUpdateIndex:      0,
		inInsertValuesClause:    false,
		insertValuesLengths:     []int{},
		currentValuesLength:     0,
		currentInsertIndex:      0,
		currentLineLength:       0,
	}
}

// format formats whitespace in a SQL string to make it easier to read.
func (f *formatter) format(query string) string {
	f.tokens = f.tokenizer.tokenize(query)

	// Pre-analyze for alignment if needed
	if f.cfg.AlignColumnNames {
		f.analyzeSelectClauses()
	}
	if f.cfg.AlignAssignments {
		f.analyzeUpdateSetClauses()
	}
	if f.cfg.AlignValues {
		f.analyzeInsertValuesClauses()
	}

	formattedQuery := f.getFormattedQueryFromTokens()
	return strings.TrimSpace(formattedQuery)
}

// analyzeSelectClauses performs a pre-analysis pass to collect alignment information for SELECT clauses.
func (f *formatter) analyzeSelectClauses() {
	f.selectColumnLengths = []int{}

	for i, tok := range f.tokens {
		f.index = i

		if tok.Type == types.TokenTypeReservedTopLevel && strings.ToUpper(tok.Value) == "SELECT" {
			f.analyzeSelectClause()
		}
	}
}

// analyzeUpdateSetClauses performs a pre-analysis pass to collect alignment information for UPDATE SET clauses.
func (f *formatter) analyzeUpdateSetClauses() {
	f.updateAssignmentLengths = []int{}

	for i, tok := range f.tokens {
		f.index = i

		if tok.Type == types.TokenTypeReservedTopLevel && strings.ToUpper(tok.Value) == "UPDATE" {
			f.analyzeUpdateSetClause()
		}
	}
}

// analyzeInsertValuesClauses performs a pre-analysis pass to collect alignment information for INSERT VALUES clauses.
func (f *formatter) analyzeInsertValuesClauses() {
	f.insertValuesLengths = []int{}

	for i, tok := range f.tokens {
		f.index = i

		if tok.Type == types.TokenTypeReservedTopLevel && strings.ToUpper(tok.Value) == "INSERT" {
			f.analyzeInsertValuesClause()
		}
	}
}

// analyzeSelectClause analyzes a single SELECT clause to determine column alignment lengths.
func (f *formatter) analyzeSelectClause() {
	// Find the end of the SELECT clause (FROM, WHERE, etc.)
	endIndex := f.findSelectClauseEnd()
	if endIndex == -1 {
		return
	}

	// Collect column lengths by simulating formatting
	columnLengths := []int{}
	currentLength := 0

	for i := f.index + 1; i < endIndex; i++ {
		tok := f.tokens[i]

		if tok.Type == types.TokenTypeReservedTopLevel && f.isSelectClauseTerminator(tok.Value) {
			break
		}

		if tok.Value == "," {
			if currentLength > 0 {
				columnLengths = append(columnLengths, currentLength)
				currentLength = 0
			}
		} else if tok.Type != types.TokenTypeWhitespace && tok.Type != types.TokenTypeLineComment && tok.Type != types.TokenTypeBlockComment {
			// Approximate rendered length
			if tok.Type == types.TokenTypeReserved {
				currentLength += len(f.formatReservedWord(tok.Value)) + 1 // +1 for space
			} else {
				currentLength += len(tok.Value) + 1 // +1 for space
			}
		}
	}

	// Add the last column if we ended without a comma
	if currentLength > 0 {
		columnLengths = append(columnLengths, currentLength)
	}

	// Store the maximum length for alignment
	if len(columnLengths) > 0 {
		maxLength := 0
		for _, length := range columnLengths {
			if length > maxLength {
				maxLength = length
			}
		}
		f.selectColumnLengths = append(f.selectColumnLengths, maxLength)
	}
}

// analyzeUpdateSetClause analyzes a single UPDATE SET clause to determine assignment alignment lengths.
func (f *formatter) analyzeUpdateSetClause() {
	// Find the SET keyword
	setIndex := -1
	for i := f.index + 1; i < len(f.tokens); i++ {
		tok := f.tokens[i]
		if tok.Type == types.TokenTypeReservedTopLevel && strings.ToUpper(tok.Value) == "SET" {
			setIndex = i
			break
		}
	}
	if setIndex == -1 {
		return
	}

	// Find the end of the SET clause (WHERE, etc.)
	endIndex := f.findUpdateSetClauseEnd(setIndex)
	if endIndex == -1 {
		endIndex = len(f.tokens)
	}

	// Collect assignment lengths by simulating formatting
	assignmentLengths := []int{}
	currentLength := 0

	for i := setIndex + 1; i < endIndex; i++ {
		tok := f.tokens[i]

		if tok.Type == types.TokenTypeReservedTopLevel && f.isUpdateSetClauseTerminator(tok.Value) {
			break
		}

		switch tok.Value {
		case "=":
			// Store the length up to the equals sign
			if currentLength > 0 {
				assignmentLengths = append(assignmentLengths, currentLength)
				currentLength = 0
			}
		case ",":
			// Skip commas
			continue
		default:
			if tok.Type != types.TokenTypeWhitespace && tok.Type != types.TokenTypeLineComment && tok.Type != types.TokenTypeBlockComment {
				// Approximate rendered length
				if tok.Type == types.TokenTypeReserved {
					currentLength += len(f.formatReservedWord(tok.Value)) + 1 // +1 for space
				} else {
					currentLength += len(tok.Value) + 1 // +1 for space
				}
			}
		}
	}

	// Store the maximum length for alignment
	if len(assignmentLengths) > 0 {
		maxLength := 0
		for _, length := range assignmentLengths {
			if length > maxLength {
				maxLength = length
			}
		}
		f.updateAssignmentLengths = append(f.updateAssignmentLengths, maxLength)
	}
}

// analyzeInsertValuesClause analyzes a single INSERT VALUES clause to determine values alignment lengths.
func (f *formatter) analyzeInsertValuesClause() {
	// Find the VALUES keyword
	valuesIndex := -1
	for i := f.index + 1; i < len(f.tokens); i++ {
		tok := f.tokens[i]
		if tok.Type == types.TokenTypeReservedTopLevel && strings.ToUpper(tok.Value) == "VALUES" {
			valuesIndex = i
			break
		}
	}
	if valuesIndex == -1 {
		return
	}

	// For INSERT VALUES, we want to keep all values in each tuple on the same line
	// So we just need to detect that VALUES alignment is enabled for this INSERT
	f.insertValuesLengths = append(f.insertValuesLengths, 1) // Just mark that alignment is enabled
}

// findSelectClauseEnd finds the end of the current SELECT clause.
func (f *formatter) findSelectClauseEnd() int {
	for i := f.index + 1; i < len(f.tokens); i++ {
		tok := f.tokens[i]
		if tok.Type == types.TokenTypeReservedTopLevel && f.isSelectClauseTerminator(tok.Value) {
			return i
		}
	}
	return -1
}

// findUpdateSetClauseEnd finds the end of the current UPDATE SET clause.
func (f *formatter) findUpdateSetClauseEnd(setIndex int) int {
	for i := setIndex + 1; i < len(f.tokens); i++ {
		tok := f.tokens[i]
		if tok.Type == types.TokenTypeReservedTopLevel && f.isUpdateSetClauseTerminator(tok.Value) {
			return i
		}
	}
	return -1
}

// isSelectClauseTerminator checks if a token value terminates a SELECT clause.
func (f *formatter) isSelectClauseTerminator(value string) bool {
	terminators := []string{"FROM", "WHERE", "GROUP BY", "ORDER BY", "HAVING", "LIMIT", "UNION", "INTERSECT", "EXCEPT"}
	value = strings.ToUpper(value)
	for _, term := range terminators {
		if value == term {
			return true
		}
	}
	return false
}

// isUpdateSetClauseTerminator checks if a token value terminates an UPDATE SET clause.
func (f *formatter) isUpdateSetClauseTerminator(value string) bool {
	terminators := []string{"WHERE", "FROM", "RETURNING"}
	value = strings.ToUpper(value)
	for _, term := range terminators {
		if value == term {
			return true
		}
	}
	return false
}

// isInsertValuesClauseTerminator checks if a token value terminates an INSERT VALUES clause.
func (f *formatter) isInsertValuesClauseTerminator(value string) bool {
	terminators := []string{"WHERE", "FROM", "RETURNING", "ON"}
	value = strings.ToUpper(value)
	for _, term := range terminators {
		if value == term {
			return true
		}
	}
	return false
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

	// Track SELECT clause state for alignment
	if strings.ToUpper(tok.Value) == "SELECT" {
		f.inSelectClause = true
		f.currentColumnLength = 0
	} else if f.inSelectClause && f.isSelectClauseTerminator(tok.Value) {
		f.inSelectClause = false
		f.currentSelectIndex++
	}

	// Track UPDATE SET clause state for alignment
	switch strings.ToUpper(tok.Value) {
	case "UPDATE":
		f.inUpdateSetClause = false // Reset, will be set when we encounter SET
		f.currentUpdateIndex++
	default:
		if tok.Type == types.TokenTypeReservedTopLevel && strings.ToUpper(tok.Value) == "SET" && f.cfg.AlignAssignments {
			f.inUpdateSetClause = true
			f.currentAssignmentLength = 0
		} else if f.inUpdateSetClause && f.isUpdateSetClauseTerminator(tok.Value) {
			f.inUpdateSetClause = false
		}
	}

	// Track INSERT VALUES clause state for alignment
	switch strings.ToUpper(tok.Value) {
	case "INSERT":
		f.inInsertValuesClause = false // Reset, will be set when we encounter VALUES
		f.currentInsertIndex++
	default:
		if tok.Type == types.TokenTypeReservedTopLevel && strings.ToUpper(tok.Value) == "VALUES" && f.cfg.AlignValues {
			f.inInsertValuesClause = true
			f.currentValuesLength = 0
		} else if f.inInsertValuesClause && f.isInsertValuesClauseTerminator(tok.Value) {
			f.inInsertValuesClause = false
		}
	}

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

	// Check if we're at the start of a line (current line only has indentation)
	atStartOfLine := f.currentLineLength == len(f.indentation.GetIndent())

	if atStartOfLine {
		// Already on a new line, just add the comment without extra spacing
		value = utils.AddANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
		query.WriteString(value)
		f.updateLineLength(tok.Value)
		f.addNewline(query)
		return
	}

	// Check if comment should be inline or on a new line
	isInline := f.shouldCommentBeInline(tok.Value)

	if isInline {
		// Place comment inline with appropriate spacing
		// First, trim any trailing spaces to have clean control over spacing
		trimSpacesEnd(query)
		spacing := f.calculateCommentSpacing()
		query.WriteString(strings.Repeat(" ", spacing))
		// Update line length: remove previous trailing spaces, add new spacing
		f.currentLineLength = len(strings.TrimRight(query.String()[strings.LastIndex(query.String(), "\n")+1:], " \t")) + spacing
	} else {
		// Place comment on new line (no extra spacing needed)
		f.addNewline(query)
	}

	// Add the comment with color formatting
	value = utils.AddANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
	query.WriteString(value)
	f.updateLineLength(tok.Value) // Track original value length, not colored
	f.addNewline(query)
}

func (f *formatter) formatBlockComment(tok types.Token, query *strings.Builder) {
	value := tok.Value

	// Check if this is a single-line block comment (no newlines inside)
	isSingleLine := !strings.Contains(tok.Value, "\n")

	if isSingleLine {
		// Check if we're at the start of a line (current line only has indentation)
		atStartOfLine := f.currentLineLength == len(f.indentation.GetIndent())

		if atStartOfLine {
			// Already on a new line, just add the comment without extra spacing
			value = utils.AddANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
			query.WriteString(value)
			f.updateLineLength(tok.Value)
			f.addNewline(query)
			return
		}

		// Treat single-line block comments like line comments
		if f.shouldCommentBeInline(tok.Value) {
			// Place comment inline with appropriate spacing
			// First, trim any trailing spaces to have clean control over spacing
			trimSpacesEnd(query)
			spacing := f.calculateCommentSpacing()
			query.WriteString(strings.Repeat(" ", spacing))
			// Update line length: remove previous trailing spaces, add new spacing
			f.currentLineLength = len(strings.TrimRight(query.String()[strings.LastIndex(query.String(), "\n")+1:], " \t")) + spacing
		} else {
			// Place comment on new line
			f.addNewline(query)
		}

		// Add the comment with color formatting
		value = utils.AddANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
		query.WriteString(value)
		f.updateLineLength(tok.Value) // Track original value length, not colored
		f.addNewline(query)
	} else {
		// Multi-line block comment: keep current behavior (separate lines)
		value = f.indentComment(value)
		value = utils.AddANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
		f.addNewline(query)
		query.WriteString(value)
		f.updateLineLength(tok.Value) // Track original value length, not colored
		f.addNewline(query)
	}
}

func (f *formatter) indentComment(comment string) string {
	return newLineFollowByWhitespaceRegex.ReplaceAllString(comment, "\n"+f.indentation.GetIndent()+" ")
}

// shouldCommentBeInline determines if a comment should be placed inline or on a new line.
func (f *formatter) shouldCommentBeInline(comment string) bool {
	// Calculate spacing needed
	spacing := f.calculateCommentSpacing()

	// Check if comment fits on current line
	return f.commentFitsOnLine(comment, spacing)
}

// calculateCommentSpacing returns the number of spaces to add before an inline comment.
func (f *formatter) calculateCommentSpacing() int {
	if f.cfg.CommentMinSpacing > 0 {
		return f.cfg.CommentMinSpacing
	}
	return 1 // default
}

// commentFitsOnLine checks if adding a comment with spacing fits within MaxLineLength.
func (f *formatter) commentFitsOnLine(comment string, spacing int) bool {
	// If no max line length is set, always allow inline
	if f.cfg.MaxLineLength <= 0 {
		return true
	}

	// Calculate total length: current line + spacing + comment
	commentLength := utils.VisibleLength(comment)
	totalLength := f.currentLineLength + spacing + commentLength

	return totalLength <= f.cfg.MaxLineLength
}

func (f *formatter) formatTopLevelReservedWordNoIndent(tok types.Token, query *strings.Builder) {
	f.indentation.DecreaseTopLevel()
	f.addNewline(query)
	value := f.equalizeWhitespace(f.formatReservedWord(tok.Value))
	query.WriteString(value)
	f.updateLineLength(value)
	f.addNewline(query)
}

func (f *formatter) formatTopLevelReservedWord(tok types.Token, query *strings.Builder) {
	f.indentation.DecreaseTopLevel()
	f.addNewline(query)

	f.indentation.IncreaseTopLevel()
	value := f.equalizeWhitespace(f.formatReservedWord(tok.Value))
	query.WriteString(value)
	f.updateLineLength(value)

	f.addNewline(query)
}

func (f *formatter) formatNewlineReservedWord(tok types.Token, query *strings.Builder) {
	// Check if we need to break line due to max length
	// Even if this is a newline reserved word, we still want to honor that behavior
	f.addNewline(query)
	value := f.equalizeWhitespace(f.formatReservedWord(tok.Value))
	query.WriteString(value)
	query.WriteString(" ")
	f.updateLineLength(value + " ")
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
	// For parentheses, only apply casing if they are SQL keywords (unlikely, but preserve old logic)
	if f.cfg.KeywordCase == KeywordCaseUppercase {
		value = strings.ToUpper(value)
	}
	query.WriteString(value)
	f.updateLineLength(value)

	f.inlineBlock.BeginIfPossible(f.tokens, f.index)
	// For INSERT VALUES alignment, treat as inline even if not detected as such
	if !f.inlineBlock.IsActive() && f.cfg.AlignValues && f.inInsertValuesClause {
		// Skip indentation for VALUES parentheses when alignment is enabled
		return
	}
	if !f.inlineBlock.IsActive() {
		f.indentation.IncreaseBlockLevel()
		f.addNewline(query)
	}
}

// formatClosingParentheses ends an inline block if one is active, or decreases the
// block level, then adds the closing paren.
func (f *formatter) formatClosingParentheses(tok types.Token, query *strings.Builder) {
	// For parentheses, only apply casing if they are SQL keywords (unlikely, but preserve old logic)
	value := tok.Value
	if f.cfg.KeywordCase == KeywordCaseUppercase {
		value = strings.ToUpper(value)
		tok.Value = value
	}

	if f.inlineBlock.IsActive() {
		f.inlineBlock.End()
		f.formatWithSpaceAfter(tok, query)
	} else if f.cfg.AlignValues && f.inInsertValuesClause {
		// For INSERT VALUES alignment, treat as inline block
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
	value := f.params.Get(tok.Key, tok.Value)
	query.WriteString(value)
	query.WriteString(" ")
	f.updateLineLength(value + " ")
}

// formatComma adds the comma to the query and adds a space. If an inline block
// is not active, it will add a new line too.
func (f *formatter) formatComma(tok types.Token, query *strings.Builder) {
	trimSpacesEnd(query)

	// Handle alignment for SELECT clauses
	if f.cfg.AlignColumnNames && f.inSelectClause && f.currentSelectIndex < len(f.selectColumnLengths) {
		maxLength := f.selectColumnLengths[f.currentSelectIndex]
		padding := maxLength - f.currentColumnLength
		if padding > 0 {
			query.WriteString(strings.Repeat(" ", padding))
		}
	}

	// Handle alignment for UPDATE SET clauses
	if f.cfg.AlignAssignments && f.inUpdateSetClause && f.currentUpdateIndex-1 < len(f.updateAssignmentLengths) {
		maxLength := f.updateAssignmentLengths[f.currentUpdateIndex-1]
		padding := maxLength - f.currentAssignmentLength
		if padding > 0 {
			query.WriteString(strings.Repeat(" ", padding))
		}
	}

	// Handle alignment for INSERT VALUES clauses
	if f.cfg.AlignValues && f.inInsertValuesClause && f.currentInsertIndex-1 < len(f.insertValuesLengths) {
		// For INSERT VALUES alignment, keep all values in a tuple on the same line
		// All commas within VALUES should not add newlines
		query.WriteString(tok.Value)
		query.WriteString(" ")
		return
	}

	query.WriteString(tok.Value)
	query.WriteString(" ")
	f.updateLineLength(tok.Value + " ")

	// For alignment, keep assignments on the same line
	if f.cfg.AlignAssignments && f.inUpdateSetClause {
		return
	}

	// For alignment, keep columns on the same line
	if f.cfg.AlignColumnNames && f.inSelectClause {
		return
	}

	// For alignment, keep values on the same line
	if f.cfg.AlignValues && f.inInsertValuesClause {
		return
	}

	// Check if we should break the line based on max line length
	shouldBreakForLength := f.cfg.MaxLineLength > 0 && f.currentLineLength > f.cfg.MaxLineLength

	if f.inlineBlock.IsActive() || (f.cfg.AlignValues && f.inInsertValuesClause) {
		return
	}
	if limitKeywordRegex.MatchString(f.previousReservedWord.Value) && !shouldBreakForLength {
		// avoids creating new lines after LIMIT keyword so that two limit items appear on one line for nicer formatting
		// unless line is too long
		return
	} else {
		// Check if next non-whitespace token is a comment - if so, let the comment formatter decide inline vs newline
		nextTok := f.nextToken()
		// Skip whitespace tokens to find the actual next token
		offset := 1
		for nextTok.Type == types.TokenTypeWhitespace && f.index+offset < len(f.tokens) {
			offset++
			nextTok = f.nextToken(offset)
		}
		isNextComment := nextTok.Type == types.TokenTypeLineComment || nextTok.Type == types.TokenTypeBlockComment

		if !isNextComment {
			f.addNewline(query)
		}
	}

	// Reset column length tracking for next column
	if f.inSelectClause {
		f.currentColumnLength = 0
	}

	// Reset assignment length tracking for next assignment
	if f.inUpdateSetClause {
		f.currentAssignmentLength = 0
	}

	// Reset values length tracking for next value
	if f.inInsertValuesClause {
		f.currentValuesLength = 0
	}
}

// formatWithSpaceAfter returns the query with spaces trimmed off the end,
// the types.Token value, and a space (" ") at the end ("query value ").
func (f *formatter) formatWithSpaceAfter(tok types.Token, query *strings.Builder) {
	trimSpacesEnd(query)
	query.WriteString(tok.Value)
	query.WriteString(" ")
	f.updateLineLength(tok.Value + " ")
}

// formatWithoutSpaceAfter returns the query with spaces trimmed off the end and
// the types.Token value ("query value").
func (f *formatter) formatWithoutSpaceAfter(tok types.Token, query *strings.Builder) {
	trimSpacesEnd(query)
	query.WriteString(tok.Value)
	f.updateLineLength(tok.Value)
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

	// Check if we need to break line before adding this token
	// Break at logical operators (AND, OR) or when line would be too long
	// But don't break if we're in an inline block or alignment is active
	shouldBreak := false
	if f.cfg.MaxLineLength > 0 && !f.inlineBlock.IsActive() {
		tokUpper := strings.ToUpper(tok.Value)
		isLogicalOp := tokUpper == "AND" || tokUpper == "OR"
		wouldExceed := f.exceedsMaxLineLength(value + " ")

		// Break if: we'd exceed the limit, OR this is a logical operator and we're close to the limit
		if wouldExceed || (isLogicalOp && f.currentLineLength > f.cfg.MaxLineLength*3/4) {
			// Don't break during alignment
			if !f.inSelectClause && !f.inUpdateSetClause && !f.inInsertValuesClause {
				shouldBreak = true
			}
		}
	}

	if shouldBreak {
		f.addNewline(query)
	}

	query.WriteString(value)
	query.WriteString(" ")

	// Track current line length
	f.updateLineLength(value + " ")

	// Track column length for SELECT alignment
	if f.inSelectClause {
		f.currentColumnLength += len(value) + 1 // +1 for the space
	}

	// Track assignment length for UPDATE alignment (up to equals sign)
	if f.inUpdateSetClause && tok.Value != "=" {
		nextTok := f.nextToken()
		if nextTok.Value != "=" {
			f.currentAssignmentLength += len(value) + 1 // +1 for the space
		}
	}
}

// formatReservedWord makes sure the reserved word is formatted according to the Config.
func (f *formatter) formatReservedWord(value string) string {
	switch f.cfg.KeywordCase {
	case KeywordCaseUppercase:
		value = strings.ToUpper(value)
	case KeywordCaseLowercase:
		value = strings.ToLower(value)
	case KeywordCaseDialect:
		value = f.formatDialectSpecificCase(value)
	case KeywordCasePreserve:
		// Keep original case
		fallthrough
	default:
		// Keep original case
	}
	value = utils.AddANSIFormats(f.cfg.ColorConfig.ReservedWordFormatOptions, value)
	return value
}

// formatDialectSpecificCase formats the reserved word according to dialect conventions.
func (f *formatter) formatDialectSpecificCase(value string) string {
	switch f.cfg.Language {
	case StandardSQL, DB2, PLSQL:
		// Standard SQL, DB2, and Oracle traditionally use uppercase
		return strings.ToUpper(value)
	case PostgreSQL, MySQL, N1QL, SQLite:
		// PostgreSQL, MySQL, N1QL, and SQLite commonly use lowercase
		return strings.ToLower(value)
	default:
		// Default to preserving original case for unknown dialects
		return value
	}
}

func (f *formatter) formatQuerySeparator(tok types.Token, query *strings.Builder) {
	f.indentation.ResetIndentation()
	trimSpacesEnd(query)
	query.WriteString(tok.Value)
	f.updateLineLength(tok.Value)
	query.WriteString(strings.Repeat("\n", f.cfg.LinesBetweenQueries))
	f.currentLineLength = 0 // Reset after query separator
}

func (f *formatter) formatString(tok types.Token, query *strings.Builder) {
	value := tok.Value
	value = utils.AddANSIFormats(f.cfg.ColorConfig.StringFormatOptions, value)
	query.WriteString(value)
	query.WriteString(" ")
	f.updateLineLength(tok.Value + " ")
}

func (f *formatter) formatNumber(tok types.Token, query *strings.Builder) {
	value := tok.Value
	value = utils.AddANSIFormats(f.cfg.ColorConfig.NumberFormatOptions, value)
	query.WriteString(value)
	query.WriteString(" ")
	f.updateLineLength(tok.Value + " ")
}

func (f *formatter) formatBoolean(tok types.Token, query *strings.Builder) {
	value := tok.Value
	value = utils.AddANSIFormats(f.cfg.ColorConfig.BooleanFormatOptions, value)
	query.WriteString(value)
	query.WriteString(" ")
	f.updateLineLength(tok.Value + " ")
}

func (f *formatter) formatSpecialOperator(tok types.Token, query *strings.Builder) {
	// Special operators like :: (type cast) should be formatted without spaces
	trimSpacesEnd(query)
	query.WriteString(tok.Value)
	f.updateLineLength(tok.Value)
}

// addNewline trims spaces from the end of query, adds a new line character if
// one does not already exist at the end, and adds the indentation to the new
// line.
func (f *formatter) addNewline(query *strings.Builder) {
	trimSpacesEnd(query)
	if !strings.HasSuffix(query.String(), "\n") {
		query.WriteString("\n")
	}
	indent := f.indentation.GetIndent()
	query.WriteString(indent)
	// Reset line length to the indentation length
	f.currentLineLength = len(indent)
}

// updateLineLength updates the current line length by adding the visible length of the string.
func (f *formatter) updateLineLength(s string) {
	f.currentLineLength += utils.VisibleLength(s)
}

// exceedsMaxLineLength checks if adding the given string would exceed the max line length.
func (f *formatter) exceedsMaxLineLength(s string) bool {
	if f.cfg.MaxLineLength <= 0 {
		return false // unlimited
	}
	return f.currentLineLength+utils.VisibleLength(s) > f.cfg.MaxLineLength
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
