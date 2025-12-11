package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt"
	"github.com/spf13/cobra"
)

const (
	defaultSQLDialect = "sql"
)

var (
	lang                  string
	indent                string
	write                 bool
	color                 bool
	uppercase             bool
	keywordCase           string
	linesBetween          int
	autoDetect            bool
	alignColumnNames      bool
	alignAssignments      bool
	alignValues           bool
	maxLineLength         int
	preserveCommentIndent bool
	commentMinSpacing     int
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

	formatCmd.Flags().StringVar(&lang, "lang", defaultSQLDialect,
		"SQL dialect (sql, postgresql, mysql, pl/sql, db2, n1ql, sqlite)")
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
	formatCmd.Flags().BoolVar(&autoDetect, "auto-detect", false,
		"Automatically detect SQL dialect from file extension and content")
	formatCmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names vertically")
	formatCmd.Flags().BoolVar(&alignAssignments, "align-assignments", false,
		"Align UPDATE assignment operators vertically")
	formatCmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT VALUES vertically")
	formatCmd.Flags().IntVar(&maxLineLength, "max-line-length", 0, "Maximum line length (0 = unlimited)")
	formatCmd.Flags().BoolVar(&preserveCommentIndent, "preserve-comment-indent", false,
		"Preserve relative indentation of comments")
	formatCmd.Flags().IntVar(&commentMinSpacing, "comment-min-spacing", 1, "Minimum spaces before inline comments")
}

func runFormat(cmd *cobra.Command, args []string) error {
	config := buildConfig(cmd)

	// Load ignore file if available
	ignoreFile, err := sqlfmt.LoadIgnoreFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load ignore file: %v\n", err)
	}

	// If no args or args is "-", read from stdin
	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		return formatStdin(config)
	}

	// Process files, filtering out ignored ones
	for _, filename := range args {
		if ignoreFile.ShouldIgnore(filename) {
			continue
		}
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
	applyCommandLineFlags(cmd, config)

	if color {
		config.WithColorConfig(sqlfmt.NewDefaultColorConfig())
	}

	return config
}

func applyCommandLineFlags(cmd *cobra.Command, config *sqlfmt.Config) {
	// Handle auto-detection first (only if explicitly requested)
	if cmd.Flags().Changed("auto-detect") && autoDetect {
		// Auto-detection will be handled per-file in formatFile
		// For now, don't set language from --lang flag if auto-detect is used
		return
	}

	// Set language (only if explicitly provided via flag)
	if cmd.Flags().Changed("lang") {
		applyLanguageFlag(config)
	}

	// Set indentation (only if explicitly provided)
	if cmd.Flags().Changed("indent") {
		config.WithIndent(indent)
	}

	// Handle keyword casing
	applyKeywordCaseFlags(cmd, config)

	// Set lines between queries (only if explicitly provided)
	if cmd.Flags().Changed("lines-between") {
		config.WithLinesBetweenQueries(linesBetween)
	}

	// Set alignment options (only if explicitly provided)
	if cmd.Flags().Changed("align-column-names") {
		config.WithAlignColumnNames(alignColumnNames)
	}
	if cmd.Flags().Changed("align-assignments") {
		config.WithAlignAssignments(alignAssignments)
	}
	if cmd.Flags().Changed("align-values") {
		config.WithAlignValues(alignValues)
	}

	// Set max line length (only if explicitly provided)
	if cmd.Flags().Changed("max-line-length") {
		config.WithMaxLineLength(maxLineLength)
	}

	// Set comment options (only if explicitly provided)
	if cmd.Flags().Changed("preserve-comment-indent") {
		config.WithPreserveCommentIndent(preserveCommentIndent)
	}
	if cmd.Flags().Changed("comment-min-spacing") {
		config.WithCommentMinSpacing(commentMinSpacing)
	}
}

func applyLanguageFlag(config *sqlfmt.Config) {
	switch strings.ToLower(lang) {
	case defaultSQLDialect, "standard":
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

func applyKeywordCaseFlags(cmd *cobra.Command, config *sqlfmt.Config) {
	// --uppercase flag takes precedence for backward compatibility
	if cmd.Flags().Changed("uppercase") && uppercase {
		config.WithKeywordCase(sqlfmt.KeywordCaseUppercase)
	} else if cmd.Flags().Changed("keyword-case") {
		applyKeywordCaseFlag(config)
	}
}

func applyKeywordCaseFlag(config *sqlfmt.Config) {
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

func formatStdin(baseConfig *sqlfmt.Config) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	config := baseConfig

	// For stdin, auto-detection only uses content (no file path available)
	if autoDetect {
		detectedLang, detected := sqlfmt.DetectDialect("", string(input))
		if detected {
			// Create a new config with detected language
			config = &sqlfmt.Config{
				Language:            detectedLang,
				Indent:              baseConfig.Indent,
				KeywordCase:         baseConfig.KeywordCase,
				LinesBetweenQueries: baseConfig.LinesBetweenQueries,
				Params:              baseConfig.Params,
				ColorConfig:         baseConfig.ColorConfig,
				TokenizerConfig:     baseConfig.TokenizerConfig,
			}
		}
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

func formatFile(filename string, baseConfig *sqlfmt.Config) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)

	// Skip empty files - they're valid and don't need formatting
	if strings.TrimSpace(contentStr) == "" {
		if write {
			fmt.Printf("Skipped %s (empty file)\n", filename)
		}
		return nil
	}

	// Start with base config
	config := baseConfig

	// Load per-directory config file for this specific file
	if dirConfig, err := sqlfmt.LoadConfigFileForPath(filename); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config file for %s: %v\n", filename, err)
	} else {
		// Apply directory-specific config (this will override global config)
		if err := dirConfig.ApplyToConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to apply config file for %s: %v\n", filename, err)
		}
	}

	// Check for inline dialect hints (these override config file settings)
	if hintedLang, found := sqlfmt.ParseInlineDialectHint(contentStr); found {
		config = &sqlfmt.Config{
			Language:            hintedLang,
			Indent:              config.Indent,
			KeywordCase:         config.KeywordCase,
			LinesBetweenQueries: config.LinesBetweenQueries,
			Params:              config.Params,
			ColorConfig:         config.ColorConfig,
			TokenizerConfig:     config.TokenizerConfig,
		}
	}

	// Handle auto-detection if enabled (this overrides everything else)
	if autoDetect {
		detectedLang, detected := sqlfmt.DetectDialect(filename, contentStr)
		if detected {
			// Create a new config with detected language, preserving other settings
			config = &sqlfmt.Config{
				Language:            detectedLang,
				Indent:              config.Indent,
				KeywordCase:         config.KeywordCase,
				LinesBetweenQueries: config.LinesBetweenQueries,
				Params:              config.Params,
				ColorConfig:         config.ColorConfig,
				TokenizerConfig:     config.TokenizerConfig,
			}
		}
	}

	var formatted string
	if color {
		formatted = sqlfmt.PrettyFormat(contentStr, config)
	} else {
		formatted = sqlfmt.Format(contentStr, config)
	}

	if write {
		// Write back to file
		if err := os.WriteFile(filename, []byte(formatted), 0o644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("Formatted %s", filename)
		if autoDetect && config.Language != baseConfig.Language {
			fmt.Printf(" (detected as %s)", config.Language)
		}
		fmt.Println()
	} else {
		// Output to stdout
		fmt.Print(formatted)
	}

	return nil
}
