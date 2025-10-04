package cmd

import (
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [files...]",
	Short: "Check if SQL files are properly formatted (alias for validate)",
	Long: `Check that SQL files are properly formatted according to the specified dialect.

This command is an alias for 'validate'. It checks if files would be changed by
running format. It's useful for CI/CD pipelines to ensure code is properly formatted.

Exit codes:
  0 - All files are properly formatted
  1 - One or more files need formatting
  2 - Error occurred

Examples:
  sqlfmt check file.sql                    # Check single file
  sqlfmt check --lang=postgresql *.sql    # Check all SQL files
  sqlfmt check --output=json *.sql        # JSON output mode
  sqlfmt check --diff file.sql            # Show what would change
  cat file.sql | sqlfmt check -            # Check stdin`,
	Args: cobra.ArbitraryArgs,
	RunE: runValidate, // Use the same function as validate
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Add the same flags as validate command
	checkCmd.Flags().StringVar(&lang, "lang", "sql", "Dialect")
	checkCmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	checkCmd.Flags().BoolVar(
		&uppercase,
		"uppercase",
		false,
		"Uppercase keywords (deprecated)",
	)
	checkCmd.Flags().StringVar(
		&keywordCase,
		"keyword-case",
		"preserve",
		"Keyword case",
	)
	checkCmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
	checkCmd.Flags().StringVar(
		&outputFormat,
		"output",
		"text",
		"Output format (text or json)",
	)
	checkCmd.Flags().BoolVar(
		&showDiff,
		"diff",
		false,
		"Show differences for files that need formatting",
	)
}
