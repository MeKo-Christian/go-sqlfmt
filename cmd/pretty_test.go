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

const prettyTestSQLDialect = "sql"

func TestPrettyFormatStdin(t *testing.T) {
	// Reset global flags
	lang = prettyTestSQLDialect
	indent = "  "
	write = false
	uppercase = false
	linesBetween = 2

	cmd := &cobra.Command{
		Use:  "pretty-format [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runPrettyFormat,
	}
	cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&write, "write", false, "Write result to file")
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

	testInput := "SELECT * FROM users"

	// Write input to stdin
	go func() {
		defer func() { _ = stdinWriter.Close() }()
		_, _ = stdinWriter.WriteString(testInput)
	}()

	// Run the command
	cmd.SetArgs([]string{"-"})
	err := cmd.Execute()

	// Restore stdout and capture output
	_ = w.Close()
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	// Output should contain ANSI color codes
	assert.Contains(t, output, "\x1b[")
}

func TestPrettyFormatFile(t *testing.T) {
	// Create a temporary SQL file
	tmpFile, err := os.CreateTemp("", "test*.sql")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	testSQL := "SELECT * FROM users WHERE name = 'john'"
	_, err = tmpFile.WriteString(testSQL)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Reset global flags
	lang = prettyTestSQLDialect
	indent = "  "
	write = false
	uppercase = false
	linesBetween = 2

	cmd := &cobra.Command{
		Use:  "pretty-format [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runPrettyFormat,
	}
	cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&write, "write", false, "Write result to file")
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
	output := buf.String()

	// Output should contain ANSI color codes
	assert.Contains(t, output, "\x1b[")
}

func TestPrettyFormatFileWithWrite(t *testing.T) {
	// Create a temporary SQL file
	tmpFile, err := os.CreateTemp("", "test*.sql")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	testSQL := "SELECT * FROM users WHERE name = 'john'"
	_, err = tmpFile.WriteString(testSQL)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Reset global flags
	lang = prettyTestSQLDialect
	indent = "  "
	write = true // Enable write flag
	uppercase = false
	linesBetween = 2

	cmd := &cobra.Command{
		Use:  "pretty-format [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runPrettyFormat,
	}
	cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&write, "write", true, "Write result to file")
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

	// Verify the file was written
	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	// Content should have color codes
	assert.Contains(t, string(content), "\x1b[")
	assert.Contains(t, output, "Pretty formatted "+tmpFile.Name())
}

func TestPrettyFormatFileError(t *testing.T) {
	// Reset global flags
	lang = prettyTestSQLDialect
	indent = "  "
	write = false
	uppercase = false
	linesBetween = 2

	cmd := &cobra.Command{
		Use:  "pretty-format [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runPrettyFormat,
	}
	cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&write, "write", false, "Write result to file")
	cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
	cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")

	// Run the command with a non-existent file
	cmd.SetArgs([]string{"/nonexistent/file.sql"})
	err := cmd.Execute()

	// Should return an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pretty format")
}

func TestPrettyPrintStdin(t *testing.T) {
	// Reset global flags
	lang = prettyTestSQLDialect
	indent = "  "
	uppercase = false
	linesBetween = 2

	cmd := &cobra.Command{
		Use:  "pretty-print [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runPrettyPrint,
	}
	cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
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

	testInput := "SELECT * FROM users"

	// Write input to stdin
	go func() {
		defer func() { _ = stdinWriter.Close() }()
		_, _ = stdinWriter.WriteString(testInput)
	}()

	// Run the command
	cmd.SetArgs([]string{"-"})
	err := cmd.Execute()

	// Restore stdout and capture output
	_ = w.Close()
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	// Output should contain ANSI color codes
	assert.Contains(t, output, "\x1b[")
}

func TestPrettyPrintFile(t *testing.T) {
	// Create a temporary SQL file
	tmpFile, err := os.CreateTemp("", "test*.sql")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	testSQL := "SELECT * FROM users WHERE name = 'john'"
	_, err = tmpFile.WriteString(testSQL)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Reset global flags
	lang = prettyTestSQLDialect
	indent = "  "
	uppercase = false
	linesBetween = 2

	cmd := &cobra.Command{
		Use:  "pretty-print [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runPrettyPrint,
	}
	cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
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
	output := buf.String()

	// Output should contain ANSI color codes
	assert.Contains(t, output, "\x1b[")
}

func TestPrettyPrintFileError(t *testing.T) {
	// Reset global flags
	lang = prettyTestSQLDialect
	indent = "  "
	uppercase = false
	linesBetween = 2

	cmd := &cobra.Command{
		Use:  "pretty-print [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runPrettyPrint,
	}
	cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
	cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")

	// Run the command with a non-existent file
	cmd.SetArgs([]string{"/nonexistent/file.sql"})
	err := cmd.Execute()

	// Should return an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pretty print")
}

func TestPrettyFormatMultipleFiles(t *testing.T) {
	// Create two temporary SQL files
	tmpFile1, err := os.CreateTemp("", "test1*.sql")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile1.Name()) }()

	tmpFile2, err := os.CreateTemp("", "test2*.sql")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile2.Name()) }()

	testSQL1 := "SELECT * FROM users"
	testSQL2 := "SELECT * FROM orders"

	_, err = tmpFile1.WriteString(testSQL1)
	require.NoError(t, err)
	_ = tmpFile1.Close()

	_, err = tmpFile2.WriteString(testSQL2)
	require.NoError(t, err)
	_ = tmpFile2.Close()

	// Reset global flags
	lang = prettyTestSQLDialect
	indent = "  "
	write = false
	uppercase = false
	linesBetween = 2

	cmd := &cobra.Command{
		Use:  "pretty-format [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runPrettyFormat,
	}
	cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&write, "write", false, "Write result to file")
	cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
	cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command with both files
	cmd.SetArgs([]string{tmpFile1.Name(), tmpFile2.Name()})
	err = cmd.Execute()
	require.NoError(t, err)

	// Restore stdout and capture output
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Output should contain ANSI color codes and content from both files
	assert.Contains(t, output, "\x1b[")
	assert.Contains(t, output, "users")
	assert.Contains(t, output, "orders")
}

func TestPrettyPrintMultipleFiles(t *testing.T) {
	// Create two temporary SQL files
	tmpFile1, err := os.CreateTemp("", "test1*.sql")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile1.Name()) }()

	tmpFile2, err := os.CreateTemp("", "test2*.sql")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile2.Name()) }()

	testSQL1 := "SELECT * FROM users"
	testSQL2 := "SELECT * FROM orders"

	_, err = tmpFile1.WriteString(testSQL1)
	require.NoError(t, err)
	_ = tmpFile1.Close()

	_, err = tmpFile2.WriteString(testSQL2)
	require.NoError(t, err)
	_ = tmpFile2.Close()

	// Reset global flags
	lang = prettyTestSQLDialect
	indent = "  "
	uppercase = false
	linesBetween = 2

	cmd := &cobra.Command{
		Use:  "pretty-print [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runPrettyPrint,
	}
	cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
	cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command with both files
	cmd.SetArgs([]string{tmpFile1.Name(), tmpFile2.Name()})
	err = cmd.Execute()
	require.NoError(t, err)

	// Restore stdout and capture output
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Output should contain ANSI color codes and content from both files
	assert.Contains(t, output, "\x1b[")
	assert.Contains(t, output, "users")
	assert.Contains(t, output, "orders")
}

func TestPrettyFormatWithDialects(t *testing.T) {
	tests := []struct {
		name     string
		dialect  string
		input    string
	}{
		{
			name:    "postgresql dialect",
			dialect: "postgresql",
			input:   "SELECT * FROM users",
		},
		{
			name:    "mysql dialect",
			dialect: "mysql",
			input:   "SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			lang = tt.dialect
			indent = "  "
			write = false
			uppercase = false
			linesBetween = 2

			cmd := &cobra.Command{
				Use:  "pretty-format [files...]",
				Args: cobra.ArbitraryArgs,
				RunE: runPrettyFormat,
			}
			cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
			cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
			cmd.Flags().BoolVar(&write, "write", false, "Write result to file")
			cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
			cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")

			// Mark the lang flag as changed
			_ = cmd.Flags().Set("lang", tt.dialect)

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
			cmd.SetArgs([]string{"-"})
			err := cmd.Execute()

			// Restore stdout and capture output
			_ = w.Close()
			os.Stdout = oldStdout
			os.Stdin = oldStdin

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			require.NoError(t, err)
			// Output should contain ANSI color codes
			assert.Contains(t, output, "\x1b[")
		})
	}
}

func TestPrettyPrintWithDialects(t *testing.T) {
	tests := []struct {
		name     string
		dialect  string
		input    string
	}{
		{
			name:    "postgresql dialect",
			dialect: "postgresql",
			input:   "SELECT * FROM users",
		},
		{
			name:    "mysql dialect",
			dialect: "mysql",
			input:   "SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			lang = tt.dialect
			indent = "  "
			uppercase = false
			linesBetween = 2

			cmd := &cobra.Command{
				Use:  "pretty-print [files...]",
				Args: cobra.ArbitraryArgs,
				RunE: runPrettyPrint,
			}
			cmd.Flags().StringVar(&lang, "lang", prettyTestSQLDialect, "SQL dialect")
			cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
			cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
			cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")

			// Mark the lang flag as changed
			_ = cmd.Flags().Set("lang", tt.dialect)

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
			cmd.SetArgs([]string{"-"})
			err := cmd.Execute()

			// Restore stdout and capture output
			_ = w.Close()
			os.Stdout = oldStdout
			os.Stdin = oldStdin

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			require.NoError(t, err)
			// Output should contain ANSI color codes
			assert.Contains(t, output, "\x1b[")
		})
	}
}
