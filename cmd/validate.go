package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	showDiff     bool
)

// ValidationResult represents the result of validating a single file.
type ValidationResult struct {
	File  string `json:"file"`
	Valid bool   `json:"valid"`
	Diff  string `json:"diff,omitempty"`
	Error string `json:"error,omitempty"`
}

// ValidationSummary represents the overall validation summary.
type ValidationSummary struct {
	TotalFiles   int                `json:"total_files"`
	ValidFiles   int                `json:"valid_files"`
	InvalidFiles int                `json:"invalid_files"`
	ErrorFiles   int                `json:"error_files"`
	Results      []ValidationResult `json:"results"`
}

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
  sqlfmt validate --output=json *.sql        # JSON output mode
  sqlfmt validate --diff file.sql            # Show what would change
  cat file.sql | sqlfmt validate -            # Validate stdin`,
	Args: cobra.ArbitraryArgs,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	// Reuse format flags but exclude --write and --color as they don't make sense for validation
	validateCmd.Flags().StringVar(&lang, "lang", "sql", "Dialect")
	validateCmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	validateCmd.Flags().BoolVar(
		&uppercase,
		"uppercase",
		false,
		"Uppercase keywords (deprecated)",
	)
	validateCmd.Flags().StringVar(
		&keywordCase,
		"keyword-case",
		"preserve",
		"Keyword case",
	)
	validateCmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
	validateCmd.Flags().StringVar(
		&outputFormat,
		"output",
		"text",
		"Output format (text or json)",
	)
	validateCmd.Flags().BoolVar(
		&showDiff,
		"diff",
		false,
		"Show differences for files that need formatting",
	)
}

func shouldValidateStdin(args []string) bool {
	return len(args) == 0 || (len(args) == 1 && args[0] == "-")
}

func runValidate(cmd *cobra.Command, args []string) error {
	config := buildConfig(cmd)
	summary := &ValidationSummary{
		Results: make([]ValidationResult, 0),
	}

	// If no args or args is "-", validate stdin
	if shouldValidateStdin(args) {
		result := validateStdinWithResult(config)
		summary.Results = append(summary.Results, result)
	} else {
		for _, filename := range args {
			result := validateFileWithResult(filename, config)
			summary.Results = append(summary.Results, result)
		}
	}

	// Calculate summary statistics
	for _, result := range summary.Results {
		summary.TotalFiles++
		switch {
		case result.Error != "":
			summary.ErrorFiles++
		case result.Valid:
			summary.ValidFiles++
		default:
			summary.InvalidFiles++
		}
	}

	// Output results based on format
	if outputFormat == "json" {
		outputJSON(summary)
	} else {
		outputText(summary)
	}

	// Exit with appropriate code
	if summary.InvalidFiles > 0 || summary.ErrorFiles > 0 {
		os.Exit(1)
	}

	return nil
}

func validateStdinWithResult(config *sqlfmt.Config) ValidationResult {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ValidationResult{
			File:  "stdin",
			Valid: false,
			Error: fmt.Sprintf("failed to read stdin: %v", err),
		}
	}

	original := string(input)
	formatted := sqlfmt.Format(original, config)

	result := ValidationResult{
		File:  "stdin",
		Valid: strings.TrimSpace(original) == strings.TrimSpace(formatted),
	}

	if !result.Valid && showDiff {
		result.Diff = generateDiff(original, formatted)
	}

	return result
}

func validateFileWithResult(filename string, config *sqlfmt.Config) ValidationResult {
	content, err := os.ReadFile(filename)
	if err != nil {
		return ValidationResult{
			File:  filename,
			Valid: false,
			Error: fmt.Sprintf("failed to read file: %v", err),
		}
	}

	original := string(content)
	formatted := sqlfmt.Format(original, config)

	result := ValidationResult{
		File:  filename,
		Valid: strings.TrimSpace(original) == strings.TrimSpace(formatted),
	}

	if !result.Valid && showDiff {
		result.Diff = generateDiff(original, formatted)
	}

	return result
}

func generateDiff(original, formatted string) string {
	// Simple diff showing before/after
	var diff strings.Builder
	diff.WriteString("--- Before\n")
	diff.WriteString("+++ After\n")

	// Split into lines for better diff
	origLines := strings.Split(original, "\n")
	formLines := strings.Split(formatted, "\n")

	maxLines := len(origLines)
	if len(formLines) > maxLines {
		maxLines = len(formLines)
	}

	for i := 0; i < maxLines; i++ {
		var origLine, formLine string
		if i < len(origLines) {
			origLine = origLines[i]
		}
		if i < len(formLines) {
			formLine = formLines[i]
		}

		if origLine != formLine {
			if origLine != "" {
				diff.WriteString(fmt.Sprintf("- %s\n", origLine))
			}
			if formLine != "" {
				diff.WriteString(fmt.Sprintf("+ %s\n", formLine))
			}
		}
	}

	return diff.String()
}

func outputJSON(summary *ValidationSummary) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(summary); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
}

func outputText(summary *ValidationSummary) {
	// Print individual results
	for _, result := range summary.Results {
		switch {
		case result.Error != "":
			fmt.Printf("%s: ERROR - %s\n", result.File, result.Error)
		case result.Valid:
			fmt.Printf("%s: properly formatted\n", result.File)
		default:
			fmt.Printf("%s: needs formatting\n", result.File)
			if result.Diff != "" {
				fmt.Println(result.Diff)
			}
		}
	}

	// Print summary if multiple files
	if summary.TotalFiles > 1 {
		fmt.Println()
		fmt.Println("Summary:")
		fmt.Printf("  Files checked: %d\n", summary.TotalFiles)
		fmt.Printf("  Files valid:   %d\n", summary.ValidFiles)
		fmt.Printf("  Files invalid: %d\n", summary.InvalidFiles)
		if summary.ErrorFiles > 0 {
			fmt.Printf("  Files errors:  %d\n", summary.ErrorFiles)
		}
	}
}
