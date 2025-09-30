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
	keywordCase  string
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

	formatCmd.Flags().StringVar(&lang, "lang", "sql", "SQL dialect (sql, postgresql, mysql, pl/sql, db2, n1ql, sqlite)")
	formatCmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	formatCmd.Flags().BoolVarP(&write, "write", "w", false, "Write result to file instead of stdout")
	formatCmd.Flags().BoolVar(&color, "color", false, "Enable ANSI color formatting")
	formatCmd.Flags().BoolVar(
		&uppercase,
		"uppercase",
		false,
		"Deprecated: convert keywords to uppercase",
	)
	formatCmd.Flags().StringVar(
		&keywordCase,
		"keyword-case",
		"preserve",
		"Keyword casing options",
	)
	formatCmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
}

func runFormat(cmd *cobra.Command, args []string) error {
	config := buildConfig(cmd)

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

func buildConfig(cmd *cobra.Command) *sqlfmt.Config {
	config := sqlfmt.NewDefaultConfig()

	// Load config file if available
	if configFile, err := sqlfmt.LoadConfigFile(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config file: %v\n", err)
	} else {
		if err := configFile.ApplyToConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to apply config file: %v\n", err)
		}
	}

	// Command-line flags override config file settings
	// We need to detect if flags were explicitly set vs using defaults

	// Set language (only if explicitly provided via flag)
	if cmd.Flags().Changed("lang") {
		switch strings.ToLower(lang) {
		case "sql", "standard":
			config.WithLang(sqlfmt.StandardSQL)
		case "postgresql", "postgres":
			config.WithLang(sqlfmt.PostgreSQL)
		case "mysql", "mariadb":
			config.WithLang(sqlfmt.MySQL)
		case "pl/sql", "plsql", "oracle":
			config.WithLang(sqlfmt.PLSQL)
		case "db2":
			config.WithLang(sqlfmt.DB2)
		case "n1ql":
			config.WithLang(sqlfmt.N1QL)
		case "sqlite":
			config.WithLang(sqlfmt.SQLite)
		default:
			fmt.Fprintf(os.Stderr, "Warning: unknown language %s, using standard SQL\n", lang)
			config.WithLang(sqlfmt.StandardSQL)
		}
	}

	// Set indentation (only if explicitly provided)
	if cmd.Flags().Changed("indent") {
		config.WithIndent(indent)
	}

	// Handle keyword casing - --uppercase flag takes precedence for backward compatibility
	if cmd.Flags().Changed("uppercase") && uppercase {
		config.WithKeywordCase(sqlfmt.KeywordCaseUppercase)
	} else if cmd.Flags().Changed("keyword-case") {
		// Convert string to KeywordCase type
		switch strings.ToLower(keywordCase) {
		case "preserve":
			config.WithKeywordCase(sqlfmt.KeywordCasePreserve)
		case "uppercase":
			config.WithKeywordCase(sqlfmt.KeywordCaseUppercase)
		case "lowercase":
			config.WithKeywordCase(sqlfmt.KeywordCaseLowercase)
		case "dialect":
			config.WithKeywordCase(sqlfmt.KeywordCaseDialect)
		default:
			fmt.Fprintf(os.Stderr, "Warning: unknown keyword-case %s, using preserve\n", keywordCase)
			config.WithKeywordCase(sqlfmt.KeywordCasePreserve)
		}
	}

	// Set lines between queries (only if explicitly provided)
	if cmd.Flags().Changed("lines-between") {
		config.WithLinesBetweenQueries(linesBetween)
	}

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
