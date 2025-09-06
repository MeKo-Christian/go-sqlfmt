package utils

import (
	"strings"
	"unicode"
)

// Dedent removes any common leading whitespace from every line in a block of text.
func Dedent(text string) string {
	// Split the text into lines
	lines := strings.Split(text, "\n")

	// Find the minimum indentation level across all lines (ignore blank lines)
	minIndent := -1
	for _, line := range lines {
		trimmedLine := strings.TrimLeftFunc(line, unicode.IsSpace)
		if len(trimmedLine) == 0 {
			// Ignore blank lines
			continue
		}

		// Count leading spaces on non-blank lines
		leadingSpaces := len(line) - len(trimmedLine)

		if minIndent == -1 || leadingSpaces < minIndent {
			minIndent = leadingSpaces
		}
	}

	// If no indentation was found, return the original text
	if minIndent <= 0 {
		return text
	}

	// Remove the common leading whitespace from each line
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}

	// Join the lines back together and return the result
	return strings.Join(lines, "\n")
}
