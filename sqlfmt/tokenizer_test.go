package sqlfmt

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWordRegex(t *testing.T) {
	re := newTokenizer(TokenizerConfig{}).wordRegex

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

func TestIndexedPlaceholderRegex(t *testing.T) {
	re := newTokenizer(TokenizerConfig{
		IndexedPlaceholderTypes: []string{"?"},
	}).indexedPlaceholderRegex

	tests := []struct {
		input string
		match string
	}{
		{input: "?", match: "?"},
		{input: "?0", match: "?0"},
		{input: "?1", match: "?1"},
		{input: "?22", match: "?22"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			matches := re.FindStringSubmatch(tt.input)
			if tt.match == "" {
				require.Len(t, matches, 0)
			} else {
				require.Truef(t, len(matches) > 0, "expected to find at least one match")
				require.Equal(t, tt.match, matches[0])
			}
		})
	}
}

func TestIdentNamedPlaceholderRegex(t *testing.T) {
	re := newTokenizer(TokenizerConfig{
		NamedPlaceholderTypes: []string{"@", ":"},
	}).identNamedPlaceholderRegex

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
