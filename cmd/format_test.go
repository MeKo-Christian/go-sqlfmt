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

const testSQLDialect = "sql"

func TestFormatCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		args     []string
		expected string
		wantErr  bool
	}{
		{
			name:  "basic format",
			input: "SELECT * FROM users WHERE id = 1",
			args:  []string{"-"},
			expected: `SELECT
  *
FROM
  users
WHERE
  id = 1`,
		},
		{
			name:  "postgresql dialect",
			input: "SELECT 'test'::text FROM users",
			args:  []string{"--lang=postgresql", "-"},
			expected: `SELECT
  'test'::text
FROM
  users`,
		},
		{
			name:  "custom indentation",
			input: "SELECT * FROM users",
			args:  []string{"--indent=    ", "-"},
			expected: `SELECT
    *
FROM
    users`,
		},
		{
			name:  "uppercase keywords",
			input: "select * from users",
			args:  []string{"--uppercase", "-"},
			expected: `SELECT
  *
FROM
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags before each test
			lang = testSQLDialect
			indent = "  "
			write = false
			color = false
			uppercase = false
			linesBetween = 2

			// Create a new command for each test to ensure isolation
			cmd := &cobra.Command{
				Use:  "format [files...]",
				Args: cobra.ArbitraryArgs,
				RunE: runFormat,
			}
			cmd.Flags().StringVar(&lang, "lang", testSQLDialect, "SQL dialect")
			cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
			cmd.Flags().BoolVar(&write, "write", false, "Write result to file")
			cmd.Flags().BoolVar(&color, "color", false, "Enable colors")
			cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
			cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Capture stdin
			oldStdin := os.Stdin
			stdinReader, stdinWriter, _ := os.Pipe()
			os.Stdin = stdinReader

			// Write input to stdin
			go func() {
				defer func() { _ = stdinWriter.Close() }()
				_, _ = stdinWriter.WriteString(tt.input)
			}()

			// Run the command
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Restore stdout and capture output
			_ = w.Close()
			os.Stdout = oldStdout
			os.Stdin = oldStdin

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, strings.TrimSpace(tt.expected), output)
			}
		})
	}
}

func TestFormatFile(t *testing.T) {
	// Create a temporary SQL file
	tmpFile, err := os.CreateTemp("", "test*.sql")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	testSQL := "SELECT * FROM users WHERE name = 'john'"
	_, err = tmpFile.WriteString(testSQL)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Reset global flags
	lang = testSQLDialect
	indent = "  "
	write = false
	color = false
	uppercase = false
	linesBetween = 2

	// Create format command
	cmd := &cobra.Command{
		Use:  "format [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runFormat,
	}
	cmd.Flags().StringVar(&lang, "lang", testSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&write, "write", false, "Write result to file")
	cmd.Flags().BoolVar(&color, "color", false, "Enable colors")
	cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
	cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	cmd.SetArgs([]string{tmpFile.Name()})
	err = cmd.Execute()
	require.NoError(t, err)

	// Restore stdout and capture output
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	expected := `SELECT
  *
FROM
  users
WHERE
  name = 'john'`

	assert.Equal(t, expected, output)
}
