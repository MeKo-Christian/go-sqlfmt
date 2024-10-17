package sqlfmt

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestRegexes(t *testing.T) {
	tz := newTokenizer(NewStandardSQLTokenizerConfig())

	tests := []struct {
		input string
		match string
		re    *regexp.Regexp
	}{
		{input: "TEXT", match: "TEXT", re: tz.wordRegex},
		{input: "TEXT);", match: "TEXT", re: tz.wordRegex},
		{input: "TEXT ", match: "TEXT", re: tz.wordRegex},
		{input: "table\nWHERE", match: "table", re: tz.wordRegex},
		{input: "table\rWHERE", match: "table", re: tz.wordRegex},
		{input: "column::int", match: "column", re: tz.wordRegex},

		{input: "?", match: "?", re: tz.indexedPlaceholderRegex},
		{input: "?0", match: "?0", re: tz.indexedPlaceholderRegex},
		{input: "?1", match: "?1", re: tz.indexedPlaceholderRegex},
		{input: "?22", match: "?22", re: tz.indexedPlaceholderRegex},

		{input: "@variable", match: "@variable", re: tz.identNamedPlaceholderRegex},
		{input: "@variable", match: "@variable", re: tz.identNamedPlaceholderRegex},
		{input: "@variable", match: "@variable", re: tz.identNamedPlaceholderRegex},
		{input: "@a1_2.3$", match: "@a1_2.3$", re: tz.identNamedPlaceholderRegex},

		{input: "@'var name'", match: "@'var name'", re: tz.stringNamedPlaceholderRegex},
		{input: `@"var name"`, match: `@"var name"`, re: tz.stringNamedPlaceholderRegex},
		{input: "@`var name`", match: "@`var name`", re: tz.stringNamedPlaceholderRegex},
		{input: "@`var name`, ", match: "@`var name`", re: tz.stringNamedPlaceholderRegex},
		{input: "@[var name]", match: "@[var name]", re: tz.stringNamedPlaceholderRegex},

		{input: "UNION ALL", match: "UNION ALL", re: tz.reservedTopLevelNoIndentRegex},

		{input: "true", match: "true", re: tz.booleanRegex},
		{input: "false", match: "false", re: tz.booleanRegex},
		{input: "TRUE ", match: "TRUE", re: tz.booleanRegex},
		{input: "true2you", match: "", re: tz.booleanRegex},
		{input: "tRUE\n", match: "tRUE", re: tz.booleanRegex},
		{input: "trueDat\t", match: "", re: tz.booleanRegex},
		{input: "(true)", match: "", re: tz.booleanRegex},
		{input: "true)", match: "true", re: tz.booleanRegex},
		{input: "true;", match: "true", re: tz.booleanRegex},

		{input: "call()", match: "call()", re: tz.functionCallRegex},
		{input: "CALL_WITH_ARGS(arg1, 3+4,\targ2);", match: "CALL_WITH_ARGS(arg1, 3+4,\targ2)", re: tz.functionCallRegex},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			matches := tt.re.FindStringSubmatch(tt.input)
			if tt.match == "" {
				require.Len(t, matches, 0)
			} else {
				require.Truef(t, len(matches) > 0, "expected to find at least one match")
				require.Equal(t, tt.match, matches[0])
			}
		})
	}
}
