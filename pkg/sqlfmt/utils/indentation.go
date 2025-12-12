package utils

import (
	"strings"
)

type indentType string

const (
	indentTypeNone       indentType = ""
	indentTypeTopLevel   indentType = "top-level"
	indentTypeBlockLevel indentType = "block-level"
	indentTypeProcedural indentType = "procedural-block"
)

type indentSource string

const (
	indentSourceTopLevel    indentSource = "top-level"
	indentSourceBlock       indentSource = "block"
	indentSourceProcedural  indentSource = "procedural-block"
)

// IndentEntry represents a single indentation level with metadata about its origin
type IndentEntry struct {
	Type    indentType
	Source  indentSource
	Keyword string
}

type Indentation struct {
	indent      string
	indentStack []IndentEntry
}

func NewIndentation(indent string) *Indentation {
	return &Indentation{
		indent:      indent,
		indentStack: []IndentEntry{},
	}
}

// GetIndent returns the current indentation string.
func (i *Indentation) GetIndent() string {
	return strings.Repeat(i.indent, len(i.indentStack))
}

// IncreaseTopLevel increases indentation by one top-level indent.
func (i *Indentation) IncreaseTopLevel() {
	i.indentStack = append(i.indentStack, IndentEntry{
		Type:   indentTypeTopLevel,
		Source: indentSourceTopLevel,
	})
}

// IncreaseBlockLevel increases indentation by one block-level indent.
func (i *Indentation) IncreaseBlockLevel() {
	i.indentStack = append(i.indentStack, IndentEntry{
		Type:   indentTypeBlockLevel,
		Source: indentSourceBlock,
	})
}

// DecreaseTopLevel decreases indentation by one top-level indent.
// Does nothing when the previous indent is not top-level.
func (i *Indentation) DecreaseTopLevel() {
	if len(i.indentStack) > 0 && i.lastIndentType() == indentTypeTopLevel {
		i.popIndentEntry()
	}
}

// DecreaseBlockLevel decreases indentation by one block-level indent.
// If there are top-level indents within the block-level indent,
// it discards those as well.
func (i *Indentation) DecreaseBlockLevel() {
	for len(i.indentStack) > 0 {
		if i.popIndentEntry().Type != indentTypeTopLevel {
			break
		}
	}
}

// ResetIndentation resets the indentation levels.
func (i *Indentation) ResetIndentation() {
	i.indentStack = []IndentEntry{}
}

// IncreaseProcedural increases indentation by one procedural block indent.
// The keyword parameter tracks which keyword created this indent (e.g., "BEGIN", "IF", "LOOP").
func (i *Indentation) IncreaseProcedural(keyword string) {
	i.indentStack = append(i.indentStack, IndentEntry{
		Type:    indentTypeProcedural,
		Source:  indentSourceProcedural,
		Keyword: keyword,
	})
}

// DecreaseProcedural decreases indentation by removing the most recent procedural block indent.
// Only removes procedural indents, leaving top-level and block-level indents intact.
func (i *Indentation) DecreaseProcedural() {
	// Search from the end for the first procedural indent and remove it
	for j := len(i.indentStack) - 1; j >= 0; j-- {
		if i.indentStack[j].Type == indentTypeProcedural {
			// Remove this entry by slicing
			i.indentStack = append(i.indentStack[:j], i.indentStack[j+1:]...)
			break
		}
	}
}

// GetProceduralDepth returns the count of procedural block indents currently in the stack.
func (i *Indentation) GetProceduralDepth() int {
	count := 0
	for _, entry := range i.indentStack {
		if entry.Type == indentTypeProcedural {
			count++
		}
	}
	return count
}

// ResetToProceduralBase removes all top-level and block-level indents while preserving
// procedural block indents. This is useful for semicolons within procedural blocks that
// should reset statement-level indentation but maintain the procedural context.
func (i *Indentation) ResetToProceduralBase() {
	// Filter to keep only procedural indents
	newStack := make([]IndentEntry, 0, len(i.indentStack))
	for _, entry := range i.indentStack {
		if entry.Type == indentTypeProcedural {
			newStack = append(newStack, entry)
		}
	}
	i.indentStack = newStack
}

// lastIndentType peeks at the last indentType in the indent stack.
// Returns indentTypeNone if the stack is empty.
func (i *Indentation) lastIndentType() indentType {
	if len(i.indentStack) > 0 {
		return i.indentStack[len(i.indentStack)-1].Type
	}
	return indentTypeNone
}

// popIndentEntry pops the last entry from the indent stack and returns it.
// Returns an empty IndentEntry with indentTypeNone if the stack is empty.
func (i *Indentation) popIndentEntry() IndentEntry {
	if len(i.indentStack) == 0 {
		return IndentEntry{Type: indentTypeNone}
	}

	lastEntry := i.indentStack[len(i.indentStack)-1]
	i.indentStack = i.indentStack[:len(i.indentStack)-1]
	return lastEntry
}
