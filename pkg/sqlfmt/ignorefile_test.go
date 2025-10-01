package sqlfmt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIgnoreFile_LoadIgnoreFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "sqlfmt-ignore-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Test with no ignore file
	ignoreFile, err := LoadIgnoreFile()
	require.NoError(t, err)
	require.Empty(t, ignoreFile.patterns)

	// Create .sqlfmtignore file
	ignoreContent := `# This is a comment
*.tmp
test/
dir/**/*.sql
`
	err = os.WriteFile(".sqlfmtignore", []byte(ignoreContent), 0o644)
	require.NoError(t, err)

	// Load ignore file
	ignoreFile, err = LoadIgnoreFile()
	require.NoError(t, err)
	require.Len(t, ignoreFile.patterns, 3)
	require.Contains(t, ignoreFile.patterns, "*.tmp")
	require.Contains(t, ignoreFile.patterns, "test/")
	require.Contains(t, ignoreFile.patterns, "dir/**/*.sql")
}

func TestIgnoreFile_ShouldIgnore(t *testing.T) {
	ignoreFile := &IgnoreFile{
		patterns: []string{"*.tmp", "test/", "dir/**/*.sql"},
	}

	tests := []struct {
		filePath string
		expected bool
	}{
		{"file.sql", false},
		{"file.tmp", true},
		{"test/file.sql", true},
		{"test/subdir/file.sql", true},
		{"dir/file.sql", true},
		{"dir/subdir/file.sql", true},
		{"other/file.sql", false},
		{"dir", false}, // directory itself should not be ignored unless pattern specifies it
	}

	for _, test := range tests {
		t.Run(test.filePath, func(t *testing.T) {
			result := ignoreFile.ShouldIgnore(test.filePath)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestIgnoreFile_matchPattern(t *testing.T) {
	ignoreFile := &IgnoreFile{}

	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		{"file.sql", "*.sql", true},
		{"file.txt", "*.sql", false},
		{"test/file.sql", "test/", true},
		{"test", "test/", true},
		{"other/file.sql", "test/", false},
		{"dir/sub/file.sql", "dir/**/*.sql", true},
		{"dir/file.sql", "dir/**/*.sql", true},
		{"other/file.sql", "dir/**/*.sql", false},
	}

	for _, test := range tests {
		t.Run(test.path+"_"+test.pattern, func(t *testing.T) {
			result := ignoreFile.matchPattern(test.path, test.pattern)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestIgnoreFile_LoadIgnoreFile_InParentDir(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "sqlfmt-ignore-parent-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create parent directory with ignore file
	parentDir := filepath.Join(tempDir, "parent")
	err = os.MkdirAll(parentDir, 0o755)
	require.NoError(t, err)

	// Create child directory
	childDir := filepath.Join(parentDir, "child")
	err = os.MkdirAll(childDir, 0o755)
	require.NoError(t, err)

	// Create .sqlfmtignore in parent directory
	ignoreContent := "*.tmp\n"
	err = os.WriteFile(filepath.Join(parentDir, ".sqlfmtignore"), []byte(ignoreContent), 0o644)
	require.NoError(t, err)

	// Change to child directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(childDir)
	require.NoError(t, err)

	// Load ignore file from child directory (should find parent .sqlfmtignore)
	ignoreFile, err := LoadIgnoreFile()
	require.NoError(t, err)
	require.Len(t, ignoreFile.patterns, 1)
	require.Equal(t, "*.tmp", ignoreFile.patterns[0])

	// Test that it correctly ignores files
	require.True(t, ignoreFile.ShouldIgnore("file.tmp"))
	require.False(t, ignoreFile.ShouldIgnore("file.sql"))
}
