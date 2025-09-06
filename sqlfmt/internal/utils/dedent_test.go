package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDedent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "simple indented text",
			input: `  line one
  line two
  line three`,
			expected: `line one
line two
line three`,
		},
		{
			name: "mixed indentation",
			input: `    first line
      second line
    third line`,
			expected: `first line
  second line
third line`,
		},
		{
			name: "no indentation",
			input: `line one
line two`,
			expected: `line one
line two`,
		},
		{
			name: "empty lines preserved",
			input: `  line one

  line three`,
			expected: `line one

line three`,
		},
		{
			name:     "single line",
			input:    `    single line`,
			expected: `single line`,
		},
		{
			name:     "empty string",
			input:    ``,
			expected: ``,
		},
		{
			name:     "tabs and spaces mixed",
			input:    "\t\tline one\n\t\tline two",
			expected: "line one\nline two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Dedent(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
