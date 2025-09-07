package utils

import "github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"

// Define the maximum length for inline blocks.
const inlineMaxLength = 50

// InlineBlock is a bookkeeper for inline blocks.
type InlineBlock struct {
	level int
}

// NewInlineBlock creates a new InlineBlock instance.
func NewInlineBlock() *InlineBlock {
	return &InlineBlock{
		level: 0,
	}
}

// beginIfPossible begins an inline block when lookahead through upcoming types.Tokens determines
// that the block would be smaller than inlineMaxLength.
func (ib *InlineBlock) BeginIfPossible(toks []types.Token, index int) {
	switch {
	case ib.level == 0 && ib.isInlineBlock(toks, index):
		ib.level = 1
	case ib.level > 0:
		ib.level++
	default:
		ib.level = 0
	}
}

// end finishes the current inline block. There might be several nested ones.
func (ib *InlineBlock) End() {
	ib.level--
}

// isActive returns true when inside an inline block.
func (ib *InlineBlock) IsActive() bool {
	return ib.level > 0
}

// isInlineBlock checks if this should be an inline parentheses block.
func (ib *InlineBlock) isInlineBlock(toks []types.Token, index int) bool {
	length := 0
	level := 0

	for i := index; i < len(toks); i++ {
		t := toks[i]
		length += len(t.Value)

		// Overran max length
		if length > inlineMaxLength {
			return false
		}

		switch t.Type {
		case types.TokenTypeOpenParen:
			level++
		case types.TokenTypeCloseParen:
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

// isForbiddenToken checks if the types.Token is forbidden inside an inline parentheses block.
func (ib *InlineBlock) isForbiddenToken(t types.Token) bool {
	return t.Type == types.TokenTypeReservedTopLevel ||
		t.Type == types.TokenTypeReservedNewline ||
		t.Type == types.TokenTypeLineComment || // original just had comment, not line comment
		t.Type == types.TokenTypeBlockComment ||
		t.Value == ";"
}
