package utils

import (
	"strings"
)

type indentType string

const (
	indentTypeNone       indentType = ""
	indentTypeTopLevel   indentType = "top-level"
	indentTypeBlockLevel indentType = "block-level"
)

type Indentation struct {
	indent      string
	indentTypes []indentType
}

func NewIndentation(indent string) *Indentation {
	return &Indentation{
		indent:      indent,
		indentTypes: []indentType{},
	}
}

// getIndent returns the current indentation string.
func (i *Indentation) GetIndent() string {
	return strings.Repeat(i.indent, len(i.indentTypes))
}

// increaseTopLevel increases indentation by one top-level indent.
func (i *Indentation) IncreaseTopLevel() {
	i.indentTypes = append(i.indentTypes, indentTypeTopLevel)
}

// increaseBlockLevel increases indentation by one block-level indent.
func (i *Indentation) IncreaseBlockLevel() {
	i.indentTypes = append(i.indentTypes, indentTypeBlockLevel)
}

// decreaseTopLevel decreases indentation by one top-level indent.
// Does nothing when the previous indent is not top-level.
func (i *Indentation) DecreaseTopLevel() {
	if len(i.indentTypes) > 0 && i.lastIndentType() == indentTypeTopLevel {
		i.popIndentType()
	}
}

// decreaseBlockLevel decreases indentation by one block-level indent.
// If there are top-level indents within the block-level indent,
// it discards those as well.
func (i *Indentation) DecreaseBlockLevel() {
	for len(i.indentTypes) > 0 {
		if i.popIndentType() != indentTypeTopLevel {
			break
		}
	}
}

// resetIndentation resets the indentation levels.
func (i *Indentation) ResetIndentation() {
	i.indentTypes = []indentType{}
}

// lastIndentType peeks at the last indentType in the list of indentTypes.
// Returns indentTypeNone if the list is empty.
func (i *Indentation) lastIndentType() indentType {
	if len(i.indentTypes) > 0 {
		return i.indentTypes[len(i.indentTypes)-1]
	}
	return indentTypeNone
}

// popIndentType pops the last element from the list of indentTypes
// and returns it. Returns indentTypeNone if the list is empty.
func (i *Indentation) popIndentType() indentType {
	if len(i.indentTypes) == 0 {
		return indentTypeNone
	}

	lastIndent := i.indentTypes[len(i.indentTypes)-1]
	i.indentTypes = i.indentTypes[:len(i.indentTypes)-1] // pop the last element
	return lastIndent
}
