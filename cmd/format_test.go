package cmd

import (
	"bytes"
	"os"
	"path/filepath"
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
		{
			name:  "align select columns",
			input: "SELECT id, name, email FROM users",
			args:  []string{"--align-column-names", "-"},
			expected: `SELECT
  id   , name, email
FROM
  users`,
		},
		{
			name:  "align update assignments",
			input: "UPDATE users SET name = 'John', email = 'john@example.com' WHERE id = 1",
			args:  []string{"--align-assignments", "-"},
			expected: `UPDATE
  users
SET
  name = 'John'        , email = 'john@example.com'
WHERE
  id = 1`,
		},
		{
			name:  "align insert values",
			input: "INSERT INTO users (id, name) VALUES (1, 'John'), (2, 'Jane')",
			args:  []string{"--align-values", "-"},
			expected: `INSERT INTO
  users (id, name)
VALUES
  (1, 'John'),
  (2, 'Jane')`,
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
			alignColumnNames = false
			alignAssignments = false
			alignValues = false

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
			cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
			cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
			cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

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
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

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

func TestFormatCommandWithConfigFile(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "sqlfmt_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create config file in the temp directory
	tmpConfig := filepath.Join(tmpDir, ".sqlfmt.yaml")
	configContent := `language: sql
indent: "  "
align_column_names: true
align_assignments: true
align_values: true
`
	err = os.WriteFile(tmpConfig, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Create SQL file in the same temp directory
	tmpSQL := filepath.Join(tmpDir, "test.sql")
	testSQL := "SELECT id, name FROM users; UPDATE users SET name = 'John' WHERE id = 1; " +
		"INSERT INTO users (id, name) VALUES (1, 'John');"
	err = os.WriteFile(tmpSQL, []byte(testSQL), 0o644)
	require.NoError(t, err)

	// Change to the temp directory so config file is found
	oldWd, _ := os.Getwd()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()

	// Reset global flags
	lang = testSQLDialect
	indent = "  "
	write = false
	color = false
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	cmd.SetArgs([]string{tmpSQL})
	err = cmd.Execute()
	require.NoError(t, err)

	// Restore stdout and capture output
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	// Expected output with all alignments enabled
	expected := `SELECT
  id  , name
FROM
  users;

UPDATE
  users
SET
  name = 'John'
WHERE
  id = 1;

INSERT INTO
  users (id, name)
VALUES
  (1, 'John');`

	assert.Equal(t, expected, output)
}

func TestFormatCommandDialects(t *testing.T) {
	tests := []struct {
		name     string
		dialect  string
		input    string
		expected string
	}{
		{
			name:    "mysql dialect",
			dialect: "mysql",
			input:   "SELECT * FROM users",
			expected: `SELECT
  *
FROM
  users`,
		},
		{
			name:    "sqlite dialect",
			dialect: "sqlite",
			input:   "SELECT * FROM users",
			expected: `SELECT
  *
FROM
  users`,
		},
		{
			name:    "db2 dialect",
			dialect: "db2",
			input:   "SELECT * FROM users",
			expected: `SELECT
  *
FROM
  users`,
		},
		{
			name:    "n1ql dialect",
			dialect: "n1ql",
			input:   "SELECT * FROM users",
			expected: `SELECT
  *
FROM
  users`,
		},
		{
			name:    "unknown dialect defaults to standard sql",
			dialect: "unknown",
			input:   "SELECT * FROM users",
			expected: `SELECT
  *
FROM
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			lang = tt.dialect
			indent = "  "
			write = false
			color = false
			uppercase = false
			linesBetween = 2
			alignColumnNames = false
			alignAssignments = false
			alignValues = false

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
			cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
			cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
			cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

			// Mark the lang flag as changed to trigger applyLanguageFlag
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
			output := strings.TrimSpace(buf.String())

			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tt.expected), output)
		})
	}
}

func TestFormatCommandKeywordCase(t *testing.T) {
	tests := []struct {
		name        string
		keywordCase string
		input       string
		expected    string
	}{
		{
			name:        "keyword-case preserve",
			keywordCase: "preserve",
			input:       "select * from users",
			expected: `select
  *
from
  users`,
		},
		{
			name:        "keyword-case uppercase",
			keywordCase: "uppercase",
			input:       "select * from users",
			expected: `SELECT
  *
FROM
  users`,
		},
		{
			name:        "keyword-case lowercase",
			keywordCase: "lowercase",
			input:       "SELECT * FROM users",
			expected: `select
  *
from
  users`,
		},
		{
			name:        "keyword-case dialect",
			keywordCase: "dialect",
			input:       "select * from users",
			expected: `SELECT
  *
FROM
  users`,
		},
		{
			name:        "keyword-case unknown defaults to preserve",
			keywordCase: "unknown",
			input:       "select * from users",
			expected: `select
  *
from
  users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			lang = testSQLDialect
			indent = "  "
			write = false
			color = false
			uppercase = false
			keywordCase = tt.keywordCase
			linesBetween = 2
			alignColumnNames = false
			alignAssignments = false
			alignValues = false

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
			cmd.Flags().StringVar(&keywordCase, "keyword-case", "preserve", "Keyword casing")
			cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
			cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
			cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
			cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

			// Mark the keyword-case flag as changed to trigger applyKeywordCaseFlag
			_ = cmd.Flags().Set("keyword-case", tt.keywordCase)

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
			output := strings.TrimSpace(buf.String())

			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tt.expected), output)
		})
	}
}

func TestFormatCommandWriteFlag(t *testing.T) {
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
	write = true // Enable write flag
	color = false
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

	// Create format command
	cmd := &cobra.Command{
		Use:  "format [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runFormat,
	}
	cmd.Flags().StringVar(&lang, "lang", testSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&write, "write", true, "Write result to file")
	cmd.Flags().BoolVar(&color, "color", false, "Enable colors")
	cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
	cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

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

	// Verify the file was written with formatted content
	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	expected := `SELECT
  *
FROM
  users
WHERE
  name = 'john'`

	assert.Equal(t, expected, strings.TrimSpace(string(content)))
	assert.Contains(t, output, "Formatted "+tmpFile.Name())
}

func TestFormatCommandColorFlag(t *testing.T) {
	// Reset global flags
	lang = testSQLDialect
	indent = "  "
	write = false
	color = true // Enable color flag
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

	cmd := &cobra.Command{
		Use:  "format [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runFormat,
	}
	cmd.Flags().StringVar(&lang, "lang", testSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&write, "write", false, "Write result to file")
	cmd.Flags().BoolVar(&color, "color", true, "Enable colors")
	cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
	cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

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

func TestFormatCommandMultipleFiles(t *testing.T) {
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
	lang = testSQLDialect
	indent = "  "
	write = false
	color = false
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

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

	// Output should contain formatted content from both files
	assert.Contains(t, output, "users")
	assert.Contains(t, output, "orders")
}

func TestFormatCommandErrorNonExistentFile(t *testing.T) {
	// Reset global flags
	lang = testSQLDialect
	indent = "  "
	write = false
	color = false
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

	// Run the command with a non-existent file
	cmd.SetArgs([]string{"/nonexistent/file.sql"})
	err := cmd.Execute()

	// Should return an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to format")
}

func TestFormatCommandOtherFlags(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func(*cobra.Command)
		input    string
	}{
		{
			name: "max-line-length flag",
			setupCmd: func(cmd *cobra.Command) {
				maxLineLength = 50
				_ = cmd.Flags().Set("max-line-length", "50")
			},
			input: "SELECT id, name, email, created_at FROM users",
		},
		{
			name: "preserve-comment-indent flag",
			setupCmd: func(cmd *cobra.Command) {
				preserveCommentIndent = true
				_ = cmd.Flags().Set("preserve-comment-indent", "true")
			},
			input: "SELECT * FROM users -- comment",
		},
		{
			name: "comment-min-spacing flag",
			setupCmd: func(cmd *cobra.Command) {
				commentMinSpacing = 4
				_ = cmd.Flags().Set("comment-min-spacing", "4")
			},
			input: "SELECT * FROM users -- comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			lang = testSQLDialect
			indent = "  "
			write = false
			color = false
			uppercase = false
			linesBetween = 2
			alignColumnNames = false
			alignAssignments = false
			alignValues = false
			maxLineLength = 0
			preserveCommentIndent = false
			commentMinSpacing = 1

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
			cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
			cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
			cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")
			cmd.Flags().IntVar(&maxLineLength, "max-line-length", 0, "Maximum line length")
			cmd.Flags().BoolVar(&preserveCommentIndent, "preserve-comment-indent", false, "Preserve comment indent")
			cmd.Flags().IntVar(&commentMinSpacing, "comment-min-spacing", 1, "Comment min spacing")

			// Apply the test-specific setup
			tt.setupCmd(cmd)

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

			require.NoError(t, err)
			// Just verify it doesn't error - the actual formatting behavior
			// is tested in the library tests
		})
	}
}

func TestFormatCommandColorWithFile(t *testing.T) {
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
	color = true // Enable color flag
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

	// Create format command
	cmd := &cobra.Command{
		Use:  "format [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runFormat,
	}
	cmd.Flags().StringVar(&lang, "lang", testSQLDialect, "SQL dialect")
	cmd.Flags().StringVar(&indent, "indent", "  ", "Indentation string")
	cmd.Flags().BoolVar(&write, "write", false, "Write result to file")
	cmd.Flags().BoolVar(&color, "color", true, "Enable colors")
	cmd.Flags().BoolVar(&uppercase, "uppercase", false, "Convert to uppercase")
	cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

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

func TestFormatCommandStandardDialectAlias(t *testing.T) {
	// Reset global flags
	lang = "standard"
	indent = "  "
	write = false
	color = false
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

	// Mark the lang flag as changed to trigger applyLanguageFlag
	_ = cmd.Flags().Set("lang", "standard")

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
	output := strings.TrimSpace(buf.String())

	require.NoError(t, err)
	expected := `SELECT
  *
FROM
  users`
	assert.Equal(t, expected, output)
}

func TestFormatCommandMariaDBAlias(t *testing.T) {
	// Reset global flags
	lang = "mariadb"
	indent = "  "
	write = false
	color = false
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

	// Mark the lang flag as changed to trigger applyLanguageFlag
	_ = cmd.Flags().Set("lang", "mariadb")

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
	output := strings.TrimSpace(buf.String())

	require.NoError(t, err)
	expected := `SELECT
  *
FROM
  users`
	assert.Equal(t, expected, output)
}

func TestFormatCommandLowercaseFlag(t *testing.T) {
	// Reset global flags
	lang = testSQLDialect
	indent = "  "
	write = false
	color = false
	uppercase = false
	keywordCase = "lowercase"
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().StringVar(&keywordCase, "keyword-case", "preserve", "Keyword casing")
	cmd.Flags().IntVar(&linesBetween, "lines-between", 2, "Lines between queries")
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

	// Mark flags as changed
	_ = cmd.Flags().Set("keyword-case", "lowercase")

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
	output := strings.TrimSpace(buf.String())

	require.NoError(t, err)
	expected := `select
  *
from
  users`
	assert.Equal(t, expected, output)
}

func TestFormatCommandOracleAlias(t *testing.T) {
	// Reset global flags
	lang = "oracle"
	indent = "  "
	write = false
	color = false
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

	// Mark the lang flag as changed to trigger applyLanguageFlag
	_ = cmd.Flags().Set("lang", "oracle")

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
	output := strings.TrimSpace(buf.String())

	require.NoError(t, err)
	expected := `SELECT
  *
FROM
  users`
	assert.Equal(t, expected, output)
}

func TestFormatCommandPostgresAlias(t *testing.T) {
	// Reset global flags
	lang = "postgres"
	indent = "  "
	write = false
	color = false
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

	// Mark the lang flag as changed to trigger applyLanguageFlag
	_ = cmd.Flags().Set("lang", "postgres")

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
	output := strings.TrimSpace(buf.String())

	require.NoError(t, err)
	expected := `SELECT
  *
FROM
  users`
	assert.Equal(t, expected, output)
}

func TestFormatCommandPLSQLAlias(t *testing.T) {
	// Reset global flags
	lang = "plsql"
	indent = "  "
	write = false
	color = false
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

	// Mark the lang flag as changed to trigger applyLanguageFlag
	_ = cmd.Flags().Set("lang", "plsql")

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
	output := strings.TrimSpace(buf.String())

	require.NoError(t, err)
	expected := `SELECT
  *
FROM
  users`
	assert.Equal(t, expected, output)
}

func TestFormatCommandPLSQLWithSlash(t *testing.T) {
	// Reset global flags
	lang = "pl/sql"
	indent = "  "
	write = false
	color = false
	uppercase = false
	linesBetween = 2
	alignColumnNames = false
	alignAssignments = false
	alignValues = false

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
	cmd.Flags().BoolVar(&alignColumnNames, "align-column-names", false, "Align SELECT column names")
	cmd.Flags().BoolVar(&alignAssignments, "align-assignments", false, "Align UPDATE assignments")
	cmd.Flags().BoolVar(&alignValues, "align-values", false, "Align INSERT values")

	// Mark the lang flag as changed to trigger applyLanguageFlag
	_ = cmd.Flags().Set("lang", "pl/sql")

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
	output := strings.TrimSpace(buf.String())

	require.NoError(t, err)
	expected := `SELECT
  *
FROM
  users`
	assert.Equal(t, expected, output)
}
