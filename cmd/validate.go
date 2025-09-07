package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [files...]",
	Short: "Check if SQL files are properly formatted",
	Long: `Validate that SQL files are properly formatted according to the specified dialect.

This command checks if files would be changed by running format. It's useful for 
CI/CD pipelines to ensure code is properly formatted.

Exit codes:
  0 - All files are properly formatted
  1 - One or more files need formatting
  2 - Error occurred

Examples:
  sqlfmt validate file.sql                    # Validate single file
  sqlfmt validate --lang=postgresql *.sql    # Validate all SQL files
  cat file.sql | sqlfmt validate -            # Validate stdin`,
	Args: cobra.ArbitraryArgs,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	// Reuse format flags but exclude --write and --color as they don't make sense for validation
	validateCmd.Flags().StringVar(&lang, "lang", "sql", "SQL dialect (sql, postgresql, pl/sql, db2, n1ql)")
	validateCmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	validateCmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert keywords to uppercase")
	validateCmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
}

func shouldValidateStdin(args []string) bool {
	return len(args) == 0 || (len(args) == 1 && args[0] == "-")
}

func runValidate(cmd *cobra.Command, args []string) error {
	config := buildConfig()
	var hasErrors bool

	// If no args or args is "-", validate stdin
	if shouldValidateStdin(args) {
		hasErrors = validateStdinAndSetErrors(config)
	} else {
		hasErrors = validateFilesAndSetErrors(args, config)
	}

	if hasErrors {
		os.Exit(1)
	}

	return nil
}

func validateStdinAndSetErrors(config *sqlfmt.Config) bool {
	valid, err := validateStdin(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return true
	}
	return !valid
}

func validateFilesAndSetErrors(filenames []string, config *sqlfmt.Config) bool {
	var hasErrors bool
	for _, filename := range filenames {
		valid, err := validateFile(filename, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to validate %s: %v\n", filename, err)
			hasErrors = true
			continue
		}
		if !valid {
			hasErrors = true
		}
	}
	return hasErrors
}

func validateStdin(config *sqlfmt.Config) (bool, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return false, fmt.Errorf("failed to read stdin: %w", err)
	}

	original := string(input)
	formatted := sqlfmt.Format(original, config)

	if strings.TrimSpace(original) != strings.TrimSpace(formatted) {
		fmt.Println("stdin: needs formatting")
		return false, nil
	}

	fmt.Println("stdin: properly formatted")
	return true, nil
}

func validateFile(filename string, config *sqlfmt.Config) (bool, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	original := string(content)
	formatted := sqlfmt.Format(original, config)

	if strings.TrimSpace(original) != strings.TrimSpace(formatted) {
		fmt.Printf("%s: needs formatting\n", filename)
		return false, nil
	}

	fmt.Printf("%s: properly formatted\n", filename)
	return true, nil
}
