package sqlfmt

import (
	"strings"
)

type indentType string

const (
	indentTypeNone       indentType = ""
	indentTypeTopLevel   indentType = "top-level"
	indentTypeBlockLevel indentType = "block-level"
)

type indentation struct {
	indent             string
	indentTypes        []indentType
	rainbowIndentation bool
}

func newIndentation(indent string) *indentation {
	return &indentation{
		indent:      indent,
		indentTypes: []indentType{},
	}
}

// getIndent returns the current indentation string.
func (i *indentation) getIndent() string {
	return strings.Repeat(i.indent, len(i.indentTypes))
}

// increaseTopLevel increases indentation by one top-level indent.
func (i *indentation) increaseTopLevel() {
	i.indentTypes = append(i.indentTypes, indentTypeTopLevel)
}

// increaseBlockLevel increases indentation by one block-level indent.
func (i *indentation) increaseBlockLevel() {
	i.indentTypes = append(i.indentTypes, indentTypeBlockLevel)
}

// decreaseTopLevel decreases indentation by one top-level indent.
// Does nothing when the previous indent is not top-level.
func (i *indentation) decreaseTopLevel() {
	if len(i.indentTypes) > 0 && i.lastIndentType() == indentTypeTopLevel {
		i.popIndentType()
	}
}

// decreaseBlockLevel decreases indentation by one block-level indent.
// If there are top-level indents within the block-level indent,
// it discards those as well.
func (i *indentation) decreaseBlockLevel() {
	for len(i.indentTypes) > 0 {
		if i.popIndentType() != indentTypeTopLevel {
			break
		}
	}
}

// resetIndentation resets the indentation levels.
func (i *indentation) resetIndentation() {
	i.indentTypes = []indentType{}
}

// lastIndentType peeks at the last indentType in the list of indentTypes.
// Returns indentTypeNone if the list is empty.
func (i *indentation) lastIndentType() indentType {
	if len(i.indentTypes) > 0 {
		return i.indentTypes[len(i.indentTypes)-1]
	}
	return indentTypeNone
}

// popIndentType pops the last element from the list of indentTypes
// and returns it. Returns indentTypeNone if the list is empty.
func (i *indentation) popIndentType() indentType {
	if len(i.indentTypes) == 0 {
		return indentTypeNone
	}

	lastIndent := i.indentTypes[len(i.indentTypes)-1]
	i.indentTypes = i.indentTypes[:len(i.indentTypes)-1] // pop the last element
	return lastIndent
}
