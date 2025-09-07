package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt"
	"github.com/spf13/cobra"
)

var (
	lang         string
	indent       string
	write        bool
	color        bool
	uppercase    bool
	linesBetween int
)

var formatCmd = &cobra.Command{
	Use:   "format [files...]",
	Short: "Format SQL files or stdin",
	Long: `Format SQL files or standard input using the specified SQL dialect.

Examples:
  sqlfmt format file.sql                    # Format file to stdout
  sqlfmt format --write file.sql           # Format file in place
  cat file.sql | sqlfmt format -            # Format stdin
  sqlfmt format --lang=postgresql file.sql # Format with PostgreSQL dialect
  sqlfmt format --color file.sql           # Format with ANSI colors`,
	Args: cobra.ArbitraryArgs,
	RunE: runFormat,
}

func init() {
	rootCmd.AddCommand(formatCmd)

	formatCmd.Flags().StringVar(&lang, "lang", "sql", "SQL dialect (sql, postgresql, pl/sql, db2, n1ql)")
	formatCmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	formatCmd.Flags().BoolVarP(&write, "write", "w", false, "Write result to file instead of stdout")
	formatCmd.Flags().BoolVar(&color, "color", false, "Enable ANSI color formatting")
	formatCmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert keywords to uppercase")
	formatCmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
}

func runFormat(cmd *cobra.Command, args []string) error {
	config := buildConfig()

	// If no args or args is "-", read from stdin
	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		return formatStdin(config)
	}

	// Process files
	for _, filename := range args {
		if err := formatFile(filename, config); err != nil {
			return fmt.Errorf("failed to format %s: %w", filename, err)
		}
	}

	return nil
}

func buildConfig() *sqlfmt.Config {
	config := sqlfmt.NewDefaultConfig()

	// Set language
	switch strings.ToLower(lang) {
	case "sql", "standard":
		config.WithLang(sqlfmt.StandardSQL)
	case "postgresql", "postgres":
		config.WithLang(sqlfmt.PostgreSQL)
	case "pl/sql", "plsql", "oracle":
		config.WithLang(sqlfmt.PLSQL)
	case "db2":
		config.WithLang(sqlfmt.DB2)
	case "n1ql":
		config.WithLang(sqlfmt.N1QL)
	default:
		fmt.Fprintf(os.Stderr, "Warning: unknown language %s, using standard SQL\n", lang)
		config.WithLang(sqlfmt.StandardSQL)
	}

	config.WithIndent(indent)
	if uppercase {
		config.WithUppercase()
	}
	config.WithLinesBetweenQueries(linesBetween)

	if color {
		config.WithColorConfig(sqlfmt.NewDefaultColorConfig())
	}

	return config
}

func formatStdin(config *sqlfmt.Config) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	var formatted string
	if color {
		formatted = sqlfmt.PrettyFormat(string(input), config)
	} else {
		formatted = sqlfmt.Format(string(input), config)
	}

	fmt.Print(formatted)
	return nil
}

func formatFile(filename string, config *sqlfmt.Config) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var formatted string
	if color {
		formatted = sqlfmt.PrettyFormat(string(content), config)
	} else {
		formatted = sqlfmt.Format(string(content), config)
	}

	if write {
		// Write back to file
		if err := os.WriteFile(filename, []byte(formatted), 0o644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("Formatted %s\n", filename)
	} else {
		// Output to stdout
		fmt.Print(formatted)
	}

	return nil
}
