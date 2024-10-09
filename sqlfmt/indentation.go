package sqlfmt

import (
	"strings"
)

// Define the indent types as constants
const (
	indentTypeTopLevel   = "top-level"
	indentTypeBlockLevel = "block-level"
)

// Indentation manages indentation levels.
type indentation struct {
	indent      string
	indentTypes []string
}

// newIndentation creates a new indentation instance with a default indent value of two spaces.
func newIndentation(indent string) *indentation {
	return &indentation{
		indent:      indent,
		indentTypes: []string{},
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
	if len(i.indentTypes) > 0 && i.indentTypes[len(i.indentTypes)-1] == indentTypeTopLevel {
		i.indentTypes = i.indentTypes[:len(i.indentTypes)-1] // pop the last element
	}
}

// decreaseBlockLevel decreases indentation by one block-level indent.
// If there are top-level indents within the block-level indent,
// it discards those as well.
func (i *indentation) decreaseBlockLevel() {
	for len(i.indentTypes) > 0 {
		type_ := i.indentTypes[len(i.indentTypes)-1]         // peek the last element
		i.indentTypes = i.indentTypes[:len(i.indentTypes)-1] // pop the last element
		if type_ != indentTypeTopLevel {
			break
		}
	}
}

// resetIndentation resets the indentation levels.
func (i *indentation) resetIndentation() {
	i.indentTypes = []string{}
}
