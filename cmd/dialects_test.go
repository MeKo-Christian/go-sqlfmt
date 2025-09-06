package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDialectsCommand(t *testing.T) {
	// Create dialects command
	cmd := &cobra.Command{
		Use: "dialects",
		Run: runDialects,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	// Restore stdout and capture output
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Check that all expected dialects are listed
	expectedDialects := []string{
		"sql",
		"Standard SQL (ANSI SQL)",
		"postgresql",
		"PostgreSQL with dialect-specific features",
		"pl/sql",
		"Oracle PL/SQL with procedural extensions",
		"db2",
		"IBM DB2 SQL dialect",
		"n1ql",
		"Couchbase N1QL (SQL for JSON)",
	}

	for _, expected := range expectedDialects {
		assert.Contains(t, output, expected, "Expected dialect information not found: %s", expected)
	}

	// Check for aliases
	assert.Contains(t, output, "Aliases: [standard]")
	assert.Contains(t, output, "Aliases: [postgres]")
	assert.Contains(t, output, "Aliases: [plsql oracle]")

	// Check for usage examples
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "sqlfmt format --lang=postgresql file.sql")
	assert.Contains(t, output, "sqlfmt format --lang=pl/sql file.sql")

	// Ensure the output is properly formatted
	lines := strings.Split(output, "\n")
	assert.Greater(t, len(lines), 10, "Output should have multiple lines")
}
