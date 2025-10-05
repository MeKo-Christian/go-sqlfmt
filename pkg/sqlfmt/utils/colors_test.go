package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddANSIFormats(t *testing.T) {
	tests := []struct {
		name     string
		options  []ANSIFormatOption
		input    string
		expected string
	}{
		{
			name:     "single color - red",
			options:  []ANSIFormatOption{ColorRed},
			input:    "hello",
			expected: "\033[31mhello\033[0m",
		},
		{
			name:     "single color - blue",
			options:  []ANSIFormatOption{ColorBlue},
			input:    "world",
			expected: "\033[34mworld\033[0m",
		},
		{
			name:     "multiple formats - bold red",
			options:  []ANSIFormatOption{FormatBold, ColorRed},
			input:    "error",
			expected: "\033[31m\033[1merror\033[0m\033[0m",
		},
		{
			name:     "multiple formats - underline green",
			options:  []ANSIFormatOption{FormatUnderline, ColorGreen},
			input:    "success",
			expected: "\033[32m\033[4msuccess\033[0m\033[0m",
		},
		{
			name:     "background color",
			options:  []ANSIFormatOption{BgColorYellow},
			input:    "highlighted",
			expected: "\033[43mhighlighted\033[0m",
		},
		{
			name:     "text and background color",
			options:  []ANSIFormatOption{ColorWhite, BgColorBlue},
			input:    "contrast",
			expected: "\033[44m\033[37mcontrast\033[0m\033[0m",
		},
		{
			name:     "bright color",
			options:  []ANSIFormatOption{ColorBrightCyan},
			input:    "bright",
			expected: "\033[96mbright\033[0m",
		},
		{
			name:     "no formatting option",
			options:  []ANSIFormatOption{NoFormatting},
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "empty string",
			options:  []ANSIFormatOption{ColorRed},
			input:    "",
			expected: "",
		},
		{
			name:     "multiline string",
			options:  []ANSIFormatOption{ColorGreen},
			input:    "line one\nline two\nline three",
			expected: "\033[32mline one\033[0m\n\033[32mline two\033[0m\n\033[32mline three\033[0m",
		},
		{
			name:     "multiline with empty lines",
			options:  []ANSIFormatOption{ColorBlue},
			input:    "first\n\nthird",
			expected: "\033[34mfirst\033[0m\n\n\033[34mthird\033[0m",
		},
		{
			name:     "single character",
			options:  []ANSIFormatOption{ColorYellow},
			input:    "x",
			expected: "\033[33mx\033[0m",
		},
		{
			name:     "empty options list",
			options:  []ANSIFormatOption{},
			input:    "text",
			expected: "text",
		},
		{
			name:     "multiple NoFormatting options",
			options:  []ANSIFormatOption{NoFormatting, NoFormatting},
			input:    "text",
			expected: "text",
		},
		{
			name:     "complex combination",
			options:  []ANSIFormatOption{FormatBold, FormatUnderline, ColorPurple, BgColorGray},
			input:    "styled",
			expected: "\033[100m\033[35m\033[4m\033[1mstyled\033[0m\033[0m\033[0m\033[0m",
		},
		{
			name:     "orange color",
			options:  []ANSIFormatOption{ColorOrange},
			input:    "warning",
			expected: "\033[38;5;208mwarning\033[0m",
		},
		{
			name:     "gray text",
			options:  []ANSIFormatOption{ColorGray},
			input:    "comment",
			expected: "\033[90mcomment\033[0m",
		},
		{
			name:     "dim formatting",
			options:  []ANSIFormatOption{FormatDim},
			input:    "faded",
			expected: "\033[2mfaded\033[0m",
		},
		{
			name:     "multiline only empty lines",
			options:  []ANSIFormatOption{ColorRed},
			input:    "\n\n",
			expected: "\n\n",
		},
		{
			name:     "trailing newline",
			options:  []ANSIFormatOption{ColorCyan},
			input:    "text\n",
			expected: "\033[36mtext\033[0m\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddANSIFormats(tt.options, tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestVisibleLength(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "plain text",
			input:    "hello world",
			expected: 11,
		},
		{
			name:     "text with red color",
			input:    "\033[31mhello\033[0m",
			expected: 5,
		},
		{
			name:     "text with blue color",
			input:    "\033[34mworld\033[0m",
			expected: 5,
		},
		{
			name:     "text with multiple ANSI codes",
			input:    "\033[1m\033[31mbold red\033[0m",
			expected: 8,
		},
		{
			name:     "text with background color",
			input:    "\033[43mhighlighted\033[0m",
			expected: 11,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "only ANSI codes",
			input:    "\033[31m\033[0m",
			expected: 0,
		},
		{
			name:     "single character with color",
			input:    "\033[32mx\033[0m",
			expected: 1,
		},
		{
			name:     "bright color code",
			input:    "\033[96mbright cyan\033[0m",
			expected: 11,
		},
		{
			name:     "256 color code (orange)",
			input:    "\033[38;5;208morange\033[0m",
			expected: 6,
		},
		{
			name:     "multiple colored segments",
			input:    "\033[31mred\033[0m \033[32mgreen\033[0m",
			expected: 9, // "red green"
		},
		{
			name:     "text with special characters",
			input:    "\033[34mhello, world!\033[0m",
			expected: 13,
		},
		{
			name:     "unicode characters with color",
			input:    "\033[35m日本語\033[0m",
			expected: 9, // UTF-8 byte length (3 chars × 3 bytes each)
		},
		{
			name:     "mixed ANSI and plain text",
			input:    "plain \033[31mred\033[0m more plain",
			expected: 20, // "plain red more plain"
		},
		{
			name:     "nested ANSI codes",
			input:    "\033[1m\033[4m\033[31mtext\033[0m",
			expected: 4,
		},
		{
			name:     "underline formatting",
			input:    "\033[4munderlined\033[0m",
			expected: 10,
		},
		{
			name:     "dim formatting",
			input:    "\033[2mfaded text\033[0m",
			expected: 10,
		},
		{
			name:     "reverse formatting",
			input:    "\033[7mreversed\033[0m",
			expected: 8,
		},
		{
			name:     "background color code",
			input:    "\033[44mblue bg\033[0m",
			expected: 7,
		},
		{
			name:     "bright background color",
			input:    "\033[101mbright red bg\033[0m",
			expected: 13,
		},
		{
			name:     "complex combination",
			input:    "\033[1m\033[4m\033[31m\033[44mcomplex\033[0m",
			expected: 7,
		},
		{
			name:     "uppercase ANSI code",
			input:    "\033[31Mred\033[0M",
			expected: 3,
		},
		{
			name:     "text with tabs",
			input:    "\033[32mtab\there\033[0m",
			expected: 8, // "tab\there"
		},
		{
			name:     "numbers and symbols",
			input:    "\033[33m123!@#\033[0m",
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VisibleLength(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestANSIFormatOptionConstants tests that all ANSI format option constants are defined correctly.
func TestANSIFormatOptionConstants(t *testing.T) {
	tests := []struct {
		name     string
		option   ANSIFormatOption
		expected string
	}{
		{"NoFormatting", NoFormatting, ""},
		{"FormatReset", FormatReset, "\033[0m"},
		{"FormatBold", FormatBold, "\033[1m"},
		{"FormatDim", FormatDim, "\033[2m"},
		{"FormatUnderline", FormatUnderline, "\033[4m"},
		{"FormatBlink", FormatBlink, "\033[5m"},
		{"FormatReverse", FormatReverse, "\033[7m"},
		{"FormatHidden", FormatHidden, "\033[8m"},

		// Text Colors
		{"ColorRed", ColorRed, "\033[31m"},
		{"ColorOrange", ColorOrange, "\033[38;5;208m"},
		{"ColorYellow", ColorYellow, "\033[33m"},
		{"ColorGreen", ColorGreen, "\033[32m"},
		{"ColorBlue", ColorBlue, "\033[34m"},
		{"ColorPurple", ColorPurple, "\033[35m"},
		{"ColorCyan", ColorCyan, "\033[36m"},
		{"ColorWhite", ColorWhite, "\033[37m"},
		{"ColorGray", ColorGray, "\033[90m"},

		// Background Colors
		{"BgColorRed", BgColorRed, "\033[41m"},
		{"BgColorOrange", BgColorOrange, "\033[48;5;208m"},
		{"BgColorGreen", BgColorGreen, "\033[42m"},
		{"BgColorYellow", BgColorYellow, "\033[43m"},
		{"BgColorBlue", BgColorBlue, "\033[44m"},
		{"BgColorPurple", BgColorPurple, "\033[45m"},
		{"BgColorCyan", BgColorCyan, "\033[46m"},
		{"BgColorWhite", BgColorWhite, "\033[47m"},
		{"BgColorGray", BgColorGray, "\033[100m"},

		// Bright Colors
		{"ColorBrightRed", ColorBrightRed, "\033[91m"},
		{"ColorBrightGreen", ColorBrightGreen, "\033[92m"},
		{"ColorBrightYellow", ColorBrightYellow, "\033[93m"},
		{"ColorBrightBlue", ColorBrightBlue, "\033[94m"},
		{"ColorBrightPurple", ColorBrightPurple, "\033[95m"},
		{"ColorBrightCyan", ColorBrightCyan, "\033[96m"},
		{"ColorBrightWhite", ColorBrightWhite, "\033[97m"},

		// Bright Background Colors
		{"BgColorBrightRed", BgColorBrightRed, "\033[101m"},
		{"BgColorBrightGreen", BgColorBrightGreen, "\033[102m"},
		{"BgColorBrightYellow", BgColorBrightYellow, "\033[103m"},
		{"BgColorBrightBlue", BgColorBrightBlue, "\033[104m"},
		{"BgColorBrightPurple", BgColorBrightPurple, "\033[105m"},
		{"BgColorBrightCyan", BgColorBrightCyan, "\033[106m"},
		{"BgColorBrightWhite", BgColorBrightWhite, "\033[107m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, string(tt.option))
		})
	}
}

// TestAddANSIFormat_EmptyLines tests that empty lines are not formatted.
func TestAddANSIFormat_EmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		option   ANSIFormatOption
		input    string
		expected string
	}{
		{
			name:     "single empty line",
			option:   ColorRed,
			input:    "",
			expected: "",
		},
		{
			name:     "two empty lines",
			option:   ColorBlue,
			input:    "\n",
			expected: "\n",
		},
		{
			name:     "text with blank line in middle",
			option:   ColorGreen,
			input:    "first\n\nlast",
			expected: "\033[32mfirst\033[0m\n\n\033[32mlast\033[0m",
		},
		{
			name:     "multiple consecutive empty lines",
			option:   ColorYellow,
			input:    "start\n\n\n\nend",
			expected: "\033[33mstart\033[0m\n\n\n\n\033[33mend\033[0m",
		},
		{
			name:     "only spaces are not empty",
			option:   ColorCyan,
			input:    "   ",
			expected: "\033[36m   \033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addANSIFormat(tt.option, tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestVisibleLength_EdgeCases tests edge cases for visible length calculation.
func TestVisibleLength_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "incomplete ANSI sequence at end",
			input:    "text\033[31",
			expected: 4, // Only "text" is visible, incomplete sequence is stripped
		},
		{
			name:     "ANSI escape without terminator",
			input:    "\033[incomplete",
			expected: 9, // "[incomplete" (escape char consumed, rest is visible until terminator found)
		},
		{
			name:     "multiple escapes in a row",
			input:    "\033[31m\033[1m\033[4mtext",
			expected: 4,
		},
		{
			name:     "escape in middle of word",
			input:    "hel\033[31mlo",
			expected: 5,
		},
		{
			name:     "lowercase letter terminator",
			input:    "\033[31mred\033[0m",
			expected: 3,
		},
		{
			name:     "space after escape code",
			input:    "\033[32m hello\033[0m",
			expected: 6, // " hello"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VisibleLength(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
