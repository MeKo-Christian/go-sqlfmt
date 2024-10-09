package sqlfmt

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWordRegex(t *testing.T) {
	re := createWordRegex([]string{})

	tests := []struct {
		input string
		match string
	}{
		{input: "TEXT", match: "TEXT"},
		{input: "TEXT);", match: "TEXT"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			matches := re.FindStringSubmatch(tt.input)
			require.Equal(t, tt.match, matches[0])
		})
	}
}
