package sqlfmt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLoadConfigFromCurrentDirectory tests loading config from the current directory
func TestLoadConfigFromCurrentDirectory(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		content      string
		wantLanguage Language
		wantIndent   string
	}{
		{
			name:     "loads .sqlfmtrc",
			filename: ".sqlfmtrc",
			content: `language: postgresql
indent: "    "`,
			wantLanguage: PostgreSQL,
			wantIndent:   "    ",
		},
		{
			name:     "loads .sqlfmt.yaml",
			filename: ".sqlfmt.yaml",
			content: `language: mysql
indent: "\t"`,
			wantLanguage: MySQL,
			wantIndent:   "\t",
		},
		{
			name:     "loads .sqlfmt.yml",
			filename: ".sqlfmt.yml",
			content: `language: sqlite
indent: "  "`,
			wantLanguage: SQLite,
			wantIndent:   "  ",
		},
		{
			name:     "loads sqlfmt.yaml",
			filename: "sqlfmt.yaml",
			content: `language: standard
indent: "   "`,
			wantLanguage: StandardSQL,
			wantIndent:   "   ",
		},
		{
			name:     "loads sqlfmt.yml",
			filename: "sqlfmt.yml",
			content: `language: db2
indent: "\t\t"`,
			wantLanguage: DB2,
			wantIndent:   "\t\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and change to it
			tmpDir := t.TempDir()
			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(origDir)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			// Write config file
			err = os.WriteFile(tt.filename, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Load config
			configFile, err := LoadConfigFile()
			require.NoError(t, err)
			require.NotNil(t, configFile)

			// Apply to config
			config := NewDefaultConfig()
			err = configFile.ApplyToConfig(config)
			require.NoError(t, err)

			// Verify
			require.Equal(t, tt.wantLanguage, config.Language)
			require.Equal(t, tt.wantIndent, config.Indent)
		})
	}
}

// TestLoadConfigFromParentDirectories tests config loading from parent directories
func TestLoadConfigFromParentDirectories(t *testing.T) {
	// Create directory structure:
	// tmpDir/
	//   .sqlfmtrc (root config)
	//   subdir1/
	//     subdir2/
	//       (test runs here)

	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	// Create root config
	rootConfig := `language: postgresql
indent: "    "`
	err = os.WriteFile(filepath.Join(tmpDir, ".sqlfmtrc"), []byte(rootConfig), 0644)
	require.NoError(t, err)

	// Create subdirectories
	subDir1 := filepath.Join(tmpDir, "subdir1")
	subDir2 := filepath.Join(subDir1, "subdir2")
	err = os.MkdirAll(subDir2, 0755)
	require.NoError(t, err)

	// Change to deepest directory
	err = os.Chdir(subDir2)
	require.NoError(t, err)

	// Load config - should find parent config
	configFile, err := LoadConfigFile()
	require.NoError(t, err)
	require.NotNil(t, configFile)

	config := NewDefaultConfig()
	err = configFile.ApplyToConfig(config)
	require.NoError(t, err)

	require.Equal(t, PostgreSQL, config.Language)
	require.Equal(t, "    ", config.Indent)
}

// TestLoadConfigFromHomeDirectory tests loading config from home directory
func TestLoadConfigFromHomeDirectory(t *testing.T) {
	// Create temp directory to use as fake home
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpHome)

	// Create another temp directory for working directory (without config)
	tmpWorkDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	err = os.Chdir(tmpWorkDir)
	require.NoError(t, err)

	// Write config to home directory
	homeConfig := `language: mysql
keyword_case: uppercase`
	err = os.WriteFile(filepath.Join(tmpHome, ".sqlfmtrc"), []byte(homeConfig), 0644)
	require.NoError(t, err)

	// Load config - should find home config
	configFile, err := LoadConfigFile()
	require.NoError(t, err)
	require.NotNil(t, configFile)

	config := NewDefaultConfig()
	err = configFile.ApplyToConfig(config)
	require.NoError(t, err)

	require.Equal(t, MySQL, config.Language)
	require.Equal(t, KeywordCaseUppercase, config.KeywordCase)
}

// TestConfigSearchOrderPrecedence tests that current dir takes precedence over home
func TestConfigSearchOrderPrecedence(t *testing.T) {
	// Create temp home with config
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpHome)

	homeConfig := `language: mysql`
	err := os.WriteFile(filepath.Join(tmpHome, ".sqlfmtrc"), []byte(homeConfig), 0644)
	require.NoError(t, err)

	// Create temp work dir with different config
	tmpWorkDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	err = os.Chdir(tmpWorkDir)
	require.NoError(t, err)

	workConfig := `language: postgresql`
	err = os.WriteFile(".sqlfmtrc", []byte(workConfig), 0644)
	require.NoError(t, err)

	// Load config - should prefer current directory
	configFile, err := LoadConfigFile()
	require.NoError(t, err)
	require.NotNil(t, configFile)

	config := NewDefaultConfig()
	err = configFile.ApplyToConfig(config)
	require.NoError(t, err)

	// Should use current directory config (postgresql), not home config (mysql)
	require.Equal(t, PostgreSQL, config.Language)
}

// TestParseAllConfigOptions tests parsing all configuration options
func TestParseAllConfigOptions(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	configContent := `language: postgresql
indent: "\t"
keyword_case: uppercase
lines_between_queries: 3`

	err = os.WriteFile(".sqlfmtrc", []byte(configContent), 0644)
	require.NoError(t, err)

	configFile, err := LoadConfigFile()
	require.NoError(t, err)

	config := NewDefaultConfig()
	err = configFile.ApplyToConfig(config)
	require.NoError(t, err)

	require.Equal(t, PostgreSQL, config.Language)
	require.Equal(t, "\t", config.Indent)
	require.Equal(t, KeywordCaseUppercase, config.KeywordCase)
	require.Equal(t, 3, config.LinesBetweenQueries)
}

// TestParseLanguageVariants tests all language name variants
func TestParseLanguageVariants(t *testing.T) {
	tests := []struct {
		name         string
		yamlLanguage string
		wantLanguage Language
	}{
		{"standard sql", "sql", StandardSQL},
		{"standard variant", "standard", StandardSQL},
		{"postgresql", "postgresql", PostgreSQL},
		{"postgres", "postgres", PostgreSQL},
		{"mysql", "mysql", MySQL},
		{"mariadb", "mariadb", MySQL},
		{"plsql", "pl/sql", PLSQL},
		{"plsql variant 1", "plsql", PLSQL},
		{"plsql variant 2", "oracle", PLSQL},
		{"db2", "db2", DB2},
		{"n1ql", "n1ql", N1QL},
		{"sqlite", "sqlite", SQLite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(origDir)
			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			configContent := "language: " + tt.yamlLanguage
			err = os.WriteFile(".sqlfmtrc", []byte(configContent), 0644)
			require.NoError(t, err)

			configFile, err := LoadConfigFile()
			require.NoError(t, err)

			config := NewDefaultConfig()
			err = configFile.ApplyToConfig(config)
			require.NoError(t, err)

			require.Equal(t, tt.wantLanguage, config.Language)
		})
	}
}

// TestParseKeywordCaseVariants tests all keyword_case variants
func TestParseKeywordCaseVariants(t *testing.T) {
	tests := []struct {
		name            string
		yamlKeywordCase string
		wantKeywordCase KeywordCase
	}{
		{"preserve", "preserve", KeywordCasePreserve},
		{"uppercase", "uppercase", KeywordCaseUppercase},
		{"lowercase", "lowercase", KeywordCaseLowercase},
		{"dialect", "dialect", KeywordCaseDialect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(origDir)
			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			configContent := "keyword_case: " + tt.yamlKeywordCase
			err = os.WriteFile(".sqlfmtrc", []byte(configContent), 0644)
			require.NoError(t, err)

			configFile, err := LoadConfigFile()
			require.NoError(t, err)

			config := NewDefaultConfig()
			err = configFile.ApplyToConfig(config)
			require.NoError(t, err)

			require.Equal(t, tt.wantKeywordCase, config.KeywordCase)
		})
	}
}

// TestInvalidYAMLHandling tests error handling for invalid YAML
func TestInvalidYAMLHandling(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Write invalid YAML
	invalidYAML := `language: postgresql
indent: [this is not valid
keyword_case: uppercase`
	err = os.WriteFile(".sqlfmtrc", []byte(invalidYAML), 0644)
	require.NoError(t, err)

	// Should return error
	_, err = LoadConfigFile()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse config file")
}

// TestUnknownLanguageHandling tests error handling for unknown language
func TestUnknownLanguageHandling(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	configContent := `language: unknown_language`
	err = os.WriteFile(".sqlfmtrc", []byte(configContent), 0644)
	require.NoError(t, err)

	configFile, err := LoadConfigFile()
	require.NoError(t, err)

	config := NewDefaultConfig()
	err = configFile.ApplyToConfig(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown language")
}

// TestUnknownKeywordCaseHandling tests error handling for unknown keyword_case
func TestUnknownKeywordCaseHandling(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	configContent := `keyword_case: unknown_case`
	err = os.WriteFile(".sqlfmtrc", []byte(configContent), 0644)
	require.NoError(t, err)

	configFile, err := LoadConfigFile()
	require.NoError(t, err)

	config := NewDefaultConfig()
	err = configFile.ApplyToConfig(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown keyword_case")
}

// TestNoConfigFileFound tests behavior when no config file exists
func TestNoConfigFileFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Don't create any config file
	configFile, err := LoadConfigFile()
	require.NoError(t, err)
	require.NotNil(t, configFile)

	// Should have empty/default values
	require.Equal(t, "", configFile.Language)
	require.Equal(t, "", configFile.Indent)
	require.Equal(t, "", configFile.KeywordCase)
	require.Equal(t, 0, configFile.LinesBetweenQueries)
}

// TestEmptyConfigFile tests behavior with empty config file
func TestEmptyConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Write empty config file
	err = os.WriteFile(".sqlfmtrc", []byte(""), 0644)
	require.NoError(t, err)

	configFile, err := LoadConfigFile()
	require.NoError(t, err)
	require.NotNil(t, configFile)

	// Apply to config - should keep defaults
	config := NewDefaultConfig()
	origLanguage := config.Language
	origIndent := config.Indent

	err = configFile.ApplyToConfig(config)
	require.NoError(t, err)

	// Should maintain defaults
	require.Equal(t, origLanguage, config.Language)
	require.Equal(t, origIndent, config.Indent)
}

// TestPartialConfigFile tests config file with only some options set
func TestPartialConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Only set language, leave others as default
	configContent := `language: postgresql`
	err = os.WriteFile(".sqlfmtrc", []byte(configContent), 0644)
	require.NoError(t, err)

	configFile, err := LoadConfigFile()
	require.NoError(t, err)

	config := NewDefaultConfig()
	origIndent := config.Indent
	origLinesBetween := config.LinesBetweenQueries

	err = configFile.ApplyToConfig(config)
	require.NoError(t, err)

	// Language should change, others stay default
	require.Equal(t, PostgreSQL, config.Language)
	require.Equal(t, origIndent, config.Indent)
	require.Equal(t, origLinesBetween, config.LinesBetweenQueries)
}

// TestGitRootStopsSearch tests that config search stops at git root
func TestGitRootStopsSearch(t *testing.T) {
	// Create directory structure:
	// tmpDir/
	//   .sqlfmtrc (should NOT be found)
	//   subdir/
	//     .git/ (git root)
	//     subdir2/
	//       (test runs here)

	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	// Create config above git root
	rootConfig := `language: mysql`
	err = os.WriteFile(filepath.Join(tmpDir, ".sqlfmtrc"), []byte(rootConfig), 0644)
	require.NoError(t, err)

	// Create git root
	gitRootDir := filepath.Join(tmpDir, "subdir")
	gitDir := filepath.Join(gitRootDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create test directory below git root
	testDir := filepath.Join(gitRootDir, "subdir2")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	// Change to test directory
	err = os.Chdir(testDir)
	require.NoError(t, err)

	// Load config - should NOT find config above git root
	configFile, err := LoadConfigFile()
	require.NoError(t, err)
	require.NotNil(t, configFile)

	// Should be empty since config is above git root
	require.Equal(t, "", configFile.Language)
}

// TestConfigWithGitRoot tests that config at git root is found
func TestConfigWithGitRoot(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create config at git root
	rootConfig := `language: postgresql`
	err = os.WriteFile(filepath.Join(tmpDir, ".sqlfmtrc"), []byte(rootConfig), 0644)
	require.NoError(t, err)

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Change to subdirectory
	err = os.Chdir(subDir)
	require.NoError(t, err)

	// Load config - should find config at git root
	configFile, err := LoadConfigFile()
	require.NoError(t, err)
	require.NotNil(t, configFile)

	config := NewDefaultConfig()
	err = configFile.ApplyToConfig(config)
	require.NoError(t, err)

	require.Equal(t, PostgreSQL, config.Language)
}

// TestMultipleConfigFilesPrecedence tests precedence when multiple config files exist
func TestMultipleConfigFilesPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create multiple config files - .sqlfmtrc should be checked first
	err = os.WriteFile(".sqlfmtrc", []byte("language: postgresql"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(".sqlfmt.yaml", []byte("language: mysql"), 0644)
	require.NoError(t, err)
	err = os.WriteFile("sqlfmt.yml", []byte("language: sqlite"), 0644)
	require.NoError(t, err)

	// Should load .sqlfmtrc first
	configFile, err := LoadConfigFile()
	require.NoError(t, err)

	config := NewDefaultConfig()
	err = configFile.ApplyToConfig(config)
	require.NoError(t, err)

	require.Equal(t, PostgreSQL, config.Language)
}

// TestCaseInsensitiveLanguageParsing tests that language parsing is case-insensitive
func TestCaseInsensitiveLanguageParsing(t *testing.T) {
	tests := []struct {
		name         string
		yamlLanguage string
		wantLanguage Language
	}{
		{"uppercase", "POSTGRESQL", PostgreSQL},
		{"mixed case", "PostgreSQL", PostgreSQL},
		{"mixed case mysql", "MySQL", MySQL},
		{"uppercase standard", "SQL", StandardSQL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(origDir)
			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			configContent := "language: " + tt.yamlLanguage
			err = os.WriteFile(".sqlfmtrc", []byte(configContent), 0644)
			require.NoError(t, err)

			configFile, err := LoadConfigFile()
			require.NoError(t, err)

			config := NewDefaultConfig()
			err = configFile.ApplyToConfig(config)
			require.NoError(t, err)

			require.Equal(t, tt.wantLanguage, config.Language)
		})
	}
}

// TestCaseInsensitiveKeywordCaseParsing tests that keyword_case parsing is case-insensitive
func TestCaseInsensitiveKeywordCaseParsing(t *testing.T) {
	tests := []struct {
		name            string
		yamlKeywordCase string
		wantKeywordCase KeywordCase
	}{
		{"uppercase preserve", "PRESERVE", KeywordCasePreserve},
		{"mixed case uppercase", "UpperCase", KeywordCaseUppercase},
		{"mixed case lowercase", "LowerCase", KeywordCaseLowercase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(origDir)
			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			configContent := "keyword_case: " + tt.yamlKeywordCase
			err = os.WriteFile(".sqlfmtrc", []byte(configContent), 0644)
			require.NoError(t, err)

			configFile, err := LoadConfigFile()
			require.NoError(t, err)

			config := NewDefaultConfig()
			err = configFile.ApplyToConfig(config)
			require.NoError(t, err)

			require.Equal(t, tt.wantKeywordCase, config.KeywordCase)
		})
	}
}
