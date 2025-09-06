package cmd

import (
	"fmt"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt"
	"github.com/spf13/cobra"
)

var dialectsCmd = &cobra.Command{
	Use:     "dialects",
	Aliases: []string{"list-dialects", "langs"},
	Short:   "List all supported SQL dialects",
	Long: `List all SQL dialects supported by sqlfmt.

Each dialect has specific formatting rules and keyword recognition
tailored for that particular SQL variant.`,
	Run: runDialects,
}

func init() {
	rootCmd.AddCommand(dialectsCmd)
}

func runDialects(cmd *cobra.Command, args []string) {
	fmt.Println("Supported SQL dialects:")
	fmt.Println()

	dialects := []struct {
		lang        sqlfmt.Language
		name        string
		description string
		aliases     []string
	}{
		{
			lang:        sqlfmt.StandardSQL,
			name:        "sql",
			description: "Standard SQL (ANSI SQL)",
			aliases:     []string{"standard"},
		},
		{
			lang:        sqlfmt.PostgreSQL,
			name:        "postgresql",
			description: "PostgreSQL with dialect-specific features",
			aliases:     []string{"postgres"},
		},
		{
			lang:        sqlfmt.PLSQL,
			name:        "pl/sql",
			description: "Oracle PL/SQL with procedural extensions",
			aliases:     []string{"plsql", "oracle"},
		},
		{
			lang:        sqlfmt.DB2,
			name:        "db2",
			description: "IBM DB2 SQL dialect",
			aliases:     []string{},
		},
		{
			lang:        sqlfmt.N1QL,
			name:        "n1ql",
			description: "Couchbase N1QL (SQL for JSON)",
			aliases:     []string{},
		},
	}

	for _, dialect := range dialects {
		fmt.Printf("  %s\n", dialect.name)
		fmt.Printf("    %s\n", dialect.description)
		if len(dialect.aliases) > 0 {
			fmt.Printf("    Aliases: %v\n", dialect.aliases)
		}
		fmt.Println()
	}

	fmt.Println("Usage:")
	fmt.Println("  sqlfmt format --lang=postgresql file.sql")
	fmt.Println("  sqlfmt format --lang=pl/sql file.sql")
}
