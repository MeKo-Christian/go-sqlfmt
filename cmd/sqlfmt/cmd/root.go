package cmd

import (
	"github.com/spf13/cobra"
)

const version = "v1.61.0"

var rootCmd = &cobra.Command{
	Use:   "sqlfmt",
	Short: "A SQL formatter for multiple dialects",
	Long: `sqlfmt is a library and CLI tool for formatting SQL queries with support for 
multiple SQL dialects including Standard SQL, PostgreSQL, N1QL, DB2, and PL/SQL.

It provides both programmatic access as a Go library and command-line formatting
capabilities with customizable indentation, colors, and dialect-specific formatting.`,
	Version: version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.SetVersionTemplate("sqlfmt version " + version + "\n")
}
