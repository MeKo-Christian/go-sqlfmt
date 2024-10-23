package sqlfmt

import (
	"fmt"
	"strings"
)

type ANSIFormatOption string

const (
	NoFormatting ANSIFormatOption = ""

	FormatReset     ANSIFormatOption = "\033[0m"
	FormatBold      ANSIFormatOption = "\033[1m"
	FormatDim       ANSIFormatOption = "\033[2m"
	FormatUnderline ANSIFormatOption = "\033[4m"
	FormatBlink     ANSIFormatOption = "\033[5m"
	FormatReverse   ANSIFormatOption = "\033[7m"
	FormatHidden    ANSIFormatOption = "\033[8m"

	// Text Colors
	ColorRed    ANSIFormatOption = "\033[31m"
	ColorOrange ANSIFormatOption = "\033[38;5;208m"
	ColorYellow ANSIFormatOption = "\033[33m"
	ColorGreen  ANSIFormatOption = "\033[32m"
	ColorBlue   ANSIFormatOption = "\033[34m"
	ColorPurple ANSIFormatOption = "\033[35m"
	ColorCyan   ANSIFormatOption = "\033[36m"
	ColorWhite  ANSIFormatOption = "\033[37m"
	ColorGray   ANSIFormatOption = "\033[90m" // aka bright black

	// Background Colors
	BgColorRed    ANSIFormatOption = "\033[41m"
	BgColorOrange ANSIFormatOption = "\033[48;5;208m"
	BgColorGreen  ANSIFormatOption = "\033[42m"
	BgColorYellow ANSIFormatOption = "\033[43m"
	BgColorBlue   ANSIFormatOption = "\033[44m"
	BgColorPurple ANSIFormatOption = "\033[45m"
	BgColorCyan   ANSIFormatOption = "\033[46m"
	BgColorWhite  ANSIFormatOption = "\033[47m"
	BgColorGray   ANSIFormatOption = "\033[100m"

	// High Intensity (Bright) Colors
	ColorBrightRed    ANSIFormatOption = "\033[91m"
	ColorBrightGreen  ANSIFormatOption = "\033[92m"
	ColorBrightYellow ANSIFormatOption = "\033[93m"
	ColorBrightBlue   ANSIFormatOption = "\033[94m"
	ColorBrightPurple ANSIFormatOption = "\033[95m"
	ColorBrightCyan   ANSIFormatOption = "\033[96m"
	ColorBrightWhite  ANSIFormatOption = "\033[97m"

	// High Intensity (Bright) Background Colors
	BgColorBrightRed    ANSIFormatOption = "\033[101m"
	BgColorBrightGreen  ANSIFormatOption = "\033[102m"
	BgColorBrightYellow ANSIFormatOption = "\033[103m"
	BgColorBrightBlue   ANSIFormatOption = "\033[104m"
	BgColorBrightPurple ANSIFormatOption = "\033[105m"
	BgColorBrightCyan   ANSIFormatOption = "\033[106m"
	BgColorBrightWhite  ANSIFormatOption = "\033[107m"
)

func addANSIFormats(options []ANSIFormatOption, s string) string {
	for _, o := range options {
		s = addANSIFormat(o, s)
	}
	return s
}

// addANSIFormat adds the formatting option to the beginning of each line
// in the given string, and the reset option to the end of each line.
func addANSIFormat(option ANSIFormatOption, s string) string {
	if option == NoFormatting {
		return s
	}

	spl := strings.Split(s, "\n")
	for i := range spl {
		if spl[i] == "" {
			continue
		}
		spl[i] = fmt.Sprintf("%v%s%v", option, spl[i], FormatReset)
	}
	return strings.Join(spl, "\n")
}
