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

const validationSQLDialect = "sql"

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		args        []string
		expectValid bool
		wantErr     bool
	}{
		{
			name:        "already formatted",
			input:       "SELECT\n  *\nFROM\n  users",
			args:        []string{"-"},
			expectValid: true,
		},
		{
			name:        "needs formatting",
			input:       "SELECT * FROM users",
			args:        []string{"-"},
			expectValid: false,
		},
		{
			name:        "postgresql already formatted",
			input:       "SELECT\n  'test' :: text\nFROM\n  users",
			args:        []string{"--lang=postgresql", "-"},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags before each test
			lang = validationSQLDialect
			indent = "  "
			uppercase = false
			linesBetween = 2

			// Create a new command for each test to ensure isolation
			cmd := &cobra.Command{
				Use:  "validate [files...]",
				Args: cobra.ArbitraryArgs,
				RunE: runValidateTest,
			}
			cmd.Flags().StringVar(&lang, "lang", validationSQLDialect, "SQL dialect")
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
				assert.NoError(t, err)
				if tt.expectValid {
					assert.Contains(t, output, "properly formatted")
				} else {
					assert.Contains(t, output, "needs formatting")
				}
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	// Create a temporary SQL file with unformatted SQL
	tmpFile, err := os.CreateTemp("", "test*.sql")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	testSQL := "SELECT * FROM users WHERE name = 'john'"
	_, err = tmpFile.WriteString(testSQL)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Reset global flags
	lang = validationSQLDialect
	indent = "  "
	uppercase = false
	linesBetween = 2

	// Create validate command
	cmd := &cobra.Command{
		Use:  "validate [files...]",
		Args: cobra.ArbitraryArgs,
		RunE: runValidateTest,
	}
	cmd.Flags().StringVar(&lang, "lang", validationSQLDialect, "SQL dialect")
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
	output := strings.TrimSpace(buf.String())

	assert.Contains(t, output, "needs formatting")
}

// runValidateTest is like runValidate but doesn't call os.Exit.
func runValidateTest(_ *cobra.Command, args []string) error {
	config := buildConfig()

	// If no args or args is "-", validate stdin
	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		_, err := validateStdin(config)
		return err
	}

	// Process files
	for _, filename := range args {
		_, err := validateFile(filename, config)
		if err != nil {
			return err
		}
	}

	return nil
}
