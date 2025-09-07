package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt"
	"github.com/spf13/cobra"
)

var prettyFormatCmd = &cobra.Command{
	Use:   "pretty-format [files...]",
	Short: "Format SQL files or stdin with color formatting",
	Long: `Format SQL files or standard input with ANSI color formatting.
This is equivalent to running 'format --color'.

Examples:
  sqlfmt pretty-format file.sql                    # Format file to stdout with colors
  sqlfmt pretty-format --write file.sql           # Format file in place with colors
  cat file.sql | sqlfmt pretty-format -            # Format stdin with colors
  sqlfmt pretty-format --lang=postgresql file.sql # Format with PostgreSQL dialect and colors`,
	Args: cobra.ArbitraryArgs,
	RunE: runPrettyFormat,
}

var prettyPrintCmd = &cobra.Command{
	Use:   "pretty-print [files...]",
	Short: "Format and print SQL files or stdin with color formatting",
	Long: `Format and print SQL files or standard input with ANSI color formatting.
This command always prints to stdout and cannot write to files.

Examples:
  sqlfmt pretty-print file.sql                    # Format and print file with colors
  cat file.sql | sqlfmt pretty-print -            # Format and print stdin with colors
  sqlfmt pretty-print --lang=postgresql file.sql # Format and print with PostgreSQL dialect and colors`,
	Args: cobra.ArbitraryArgs,
	RunE: runPrettyPrint,
}

func init() {
	rootCmd.AddCommand(prettyFormatCmd)
	rootCmd.AddCommand(prettyPrintCmd)

	// Add flags for pretty-format (same as format but color is always enabled)
	prettyFormatCmd.Flags().StringVar(&lang, "lang", "sql", "SQL dialect (sql, postgresql, mysql, pl/sql, db2, n1ql)")
	prettyFormatCmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	prettyFormatCmd.Flags().BoolVarP(&write, "write", "w", false, "Write result to file instead of stdout")
	prettyFormatCmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert keywords to uppercase")
	prettyFormatCmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")

	// Add flags for pretty-print (same as pretty-format but no write option)
	prettyPrintCmd.Flags().StringVar(&lang, "lang", "sql", "SQL dialect (sql, postgresql, mysql, pl/sql, db2, n1ql)")
	prettyPrintCmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	prettyPrintCmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert keywords to uppercase")
	prettyPrintCmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
}

func runPrettyFormat(cmd *cobra.Command, args []string) error {
	config := buildConfig()

	// If no args or args is "-", read from stdin
	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		return prettyFormatStdin(config)
	}

	// Process files
	for _, filename := range args {
		if err := prettyFormatFile(filename, config); err != nil {
			return fmt.Errorf("failed to pretty format %s: %w", filename, err)
		}
	}

	return nil
}

func runPrettyPrint(cmd *cobra.Command, args []string) error {
	config := buildConfig()

	// If no args or args is "-", read from stdin
	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		return prettyPrintStdin(config)
	}

	// Process files
	for _, filename := range args {
		if err := prettyPrintFile(filename, config); err != nil {
			return fmt.Errorf("failed to pretty print %s: %w", filename, err)
		}
	}

	return nil
}

func prettyFormatStdin(config *sqlfmt.Config) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	formatted := sqlfmt.PrettyFormat(string(input), config)
	fmt.Print(formatted)
	return nil
}

func prettyFormatFile(filename string, config *sqlfmt.Config) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	formatted := sqlfmt.PrettyFormat(string(content), config)

	if write {
		// Write back to file
		if err := os.WriteFile(filename, []byte(formatted), 0o644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("Pretty formatted %s\n", filename)
	} else {
		// Output to stdout
		fmt.Print(formatted)
	}

	return nil
}

func prettyPrintStdin(config *sqlfmt.Config) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	sqlfmt.PrettyPrint(string(input), config)
	return nil
}

func prettyPrintFile(filename string, config *sqlfmt.Config) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	sqlfmt.PrettyPrint(string(content), config)
	return nil
}
