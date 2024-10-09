package sqlfmt

// Define the maximum length for inline blocks
const inlineMaxLength = 50

// InlineBlock is a bookkeeper for inline blocks.
type inlineBlock struct {
	level int
}

// newInlineBlock creates a new inlineBlock instance.
func newInlineBlock() *inlineBlock {
	return &inlineBlock{
		level: 0,
	}
}

// beginIfPossible begins an inline block when lookahead through upcoming tokens determines
// that the block would be smaller than inlineMaxLength.
func (ib *inlineBlock) beginIfPossible(toks []token, index int) {
	if ib.level == 0 && ib.isInlineBlock(toks, index) {
		ib.level = 1
	} else if ib.level > 0 {
		ib.level++
	} else {
		ib.level = 0
	}
}

// end finishes the current inline block. There might be several nested ones.
func (ib *inlineBlock) end() {
	ib.level--
}

// isActive returns true when inside an inline block.
func (ib *inlineBlock) isActive() bool {
	return ib.level > 0
}

// isInlineBlock checks if this should be an inline parentheses block.
func (ib *inlineBlock) isInlineBlock(toks []token, index int) bool {
	length := 0
	level := 0

	for i := index; i < len(toks); i++ {
		t := toks[i]
		length += len(t.value)

		// Overran max length
		if length > inlineMaxLength {
			return false
		}

		if t.typ == tokenTypeOpenParen {
			level++
		} else if t.typ == tokenTypeCloseParen {
			level--
			if level == 0 {
				return true
			}
		}

		if ib.isForbiddenToken(t) {
			return false
		}
	}
	return false
}

// isForbiddenToken checks if the token is forbidden inside an inline parentheses block.
func (ib *inlineBlock) isForbiddenToken(t token) bool {
	return t.typ == tokenTypeReservedTopLevel ||
		t.typ == tokenTypeReservedNewline ||
		t.typ == tokenTypeLineComment || // original just had comment, not line comment
		t.typ == tokenTypeBlockComment ||
		t.value == ";"
}
