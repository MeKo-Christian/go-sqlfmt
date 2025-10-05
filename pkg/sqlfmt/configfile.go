package sqlfmt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigFile represents the structure of a sqlfmt configuration file.
type ConfigFile struct {
	Language            string `yaml:"language,omitempty"`
	Indent              string `yaml:"indent,omitempty"`
	KeywordCase         string `yaml:"keyword_case,omitempty"`
	LinesBetweenQueries int    `yaml:"lines_between_queries,omitempty"`
	AlignColumnNames    *bool  `yaml:"align_column_names,omitempty"`
	AlignAssignments    *bool  `yaml:"align_assignments,omitempty"`
	AlignValues         *bool  `yaml:"align_values,omitempty"`
	MaxLineLength       *int   `yaml:"max_line_length,omitempty"`
}

// LoadConfigFile attempts to load configuration from various locations.
func LoadConfigFile() (*ConfigFile, error) {
	searchPaths := getConfigSearchPaths()

	for _, path := range searchPaths {
		if content, err := os.ReadFile(path); err == nil {
			var config ConfigFile
			if err := yaml.Unmarshal(content, &config); err != nil {
				return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
			}
			return &config, nil
		}
	}

	// No config file found, return empty config
	return &ConfigFile{}, nil
}

// getConfigSearchPaths returns the list of paths to search for config files.
func getConfigSearchPaths() []string {
	var paths []string

	// 1. Current directory and parent directories (up to git root)
	dir, err := os.Getwd()
	if err == nil {
		paths = append(paths, findConfigInParentDirs(dir)...)
	}

	// 2. User home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			filepath.Join(homeDir, ".sqlfmtrc"),
			filepath.Join(homeDir, ".sqlfmt.yaml"),
			filepath.Join(homeDir, ".sqlfmt.yml"),
		)
	}

	return paths
}

// findConfigInParentDirs searches for config files in current dir and parents up to git root.
func findConfigInParentDirs(startDir string) []string {
	var paths []string
	dir := startDir

	for {
		// Add potential config files in this directory
		for _, filename := range []string{".sqlfmtrc", ".sqlfmt.yaml", ".sqlfmt.yml", "sqlfmt.yaml", "sqlfmt.yml"} {
			paths = append(paths, filepath.Join(dir, filename))
		}

		parent := filepath.Dir(dir)

		// Stop if we've reached the root or found a git directory
		if parent == dir || isGitRoot(dir) {
			break
		}

		dir = parent
	}

	return paths
}

// isGitRoot checks if the directory contains a .git directory.
func isGitRoot(dir string) bool {
	gitPath := filepath.Join(dir, ".git")
	if info, err := os.Stat(gitPath); err == nil {
		return info.IsDir()
	}
	return false
}

// LoadConfigFileForPath attempts to load configuration from various locations relative to a file path.
func LoadConfigFileForPath(filePath string) (*ConfigFile, error) {
	searchPaths := getConfigSearchPathsForPath(filePath)

	for _, path := range searchPaths {
		if content, err := os.ReadFile(path); err == nil {
			var config ConfigFile
			if err := yaml.Unmarshal(content, &config); err != nil {
				return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
			}
			return &config, nil
		}
	}

	// No config file found, return empty config
	return &ConfigFile{}, nil
}

// getConfigSearchPathsForPath returns the list of paths to search for config files relative to a file path.
func getConfigSearchPathsForPath(filePath string) []string {
	var paths []string

	// Get the directory of the file
	fileDir := filepath.Dir(filePath)

	// Search from file directory up to git root
	dir := fileDir
	for {
		// Add potential config files in this directory
		for _, filename := range []string{".sqlfmtrc", ".sqlfmt.yaml", ".sqlfmt.yml", "sqlfmt.yaml", "sqlfmt.yml"} {
			paths = append(paths, filepath.Join(dir, filename))
		}

		parent := filepath.Dir(dir)

		// Stop if we've reached the root or found a git directory
		if parent == dir || isGitRoot(dir) {
			break
		}

		dir = parent
	}

	// Also search user home directory (global configs)
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			filepath.Join(homeDir, ".sqlfmtrc"),
			filepath.Join(homeDir, ".sqlfmt.yaml"),
			filepath.Join(homeDir, ".sqlfmt.yml"),
		)
	}

	return paths
}

// ParseInlineDialectHint parses SQL comments for dialect hints like "-- sqlfmt: dialect=mysql".
func ParseInlineDialectHint(content string) (Language, bool) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "--") {
			comment := strings.TrimSpace(strings.TrimPrefix(line, "--"))
			if strings.HasPrefix(comment, "sqlfmt:") {
				directive := strings.TrimSpace(strings.TrimPrefix(comment, "sqlfmt:"))
				if strings.HasPrefix(directive, "dialect=") {
					dialectStr := strings.TrimSpace(strings.TrimPrefix(directive, "dialect="))
					switch strings.ToLower(dialectStr) {
					case string(StandardSQL), "standard":
						return StandardSQL, true
					case string(PostgreSQL), "postgres":
						return PostgreSQL, true
					case string(MySQL), "mariadb":
						return MySQL, true
					case string(PLSQL), "plsql", "oracle":
						return PLSQL, true
					case string(DB2):
						return DB2, true
					case string(N1QL):
						return N1QL, true
					case string(SQLite):
						return SQLite, true
					}
				}
			}
		}
		// Stop at first non-comment line
		if line != "" && !strings.HasPrefix(line, "--") {
			break
		}
	}
	return StandardSQL, false
}

// ApplyToConfig applies the configuration file settings to a Config struct.
func (cf *ConfigFile) ApplyToConfig(config *Config) error {
	if cf.Language != "" {
		if err := applyConfigLanguage(cf.Language, config); err != nil {
			return err
		}
	}

	if cf.Indent != "" {
		config.Indent = cf.Indent
	}

	if cf.KeywordCase != "" {
		if err := applyConfigKeywordCase(cf.KeywordCase, config); err != nil {
			return err
		}
	}

	if cf.LinesBetweenQueries > 0 {
		config.LinesBetweenQueries = cf.LinesBetweenQueries
	}

	// Apply alignment options (only if explicitly set in config)
	if cf.AlignColumnNames != nil {
		config.AlignColumnNames = *cf.AlignColumnNames
	}
	if cf.AlignAssignments != nil {
		config.AlignAssignments = *cf.AlignAssignments
	}
	if cf.AlignValues != nil {
		config.AlignValues = *cf.AlignValues
	}

	// Apply max line length (only if explicitly set in config)
	if cf.MaxLineLength != nil {
		config.MaxLineLength = *cf.MaxLineLength
	}

	return nil
}

func applyConfigLanguage(langStr string, config *Config) error {
	switch strings.ToLower(langStr) {
	case "sql", "standard":
		config.Language = StandardSQL
	case "postgresql", "postgres":
		config.Language = PostgreSQL
	case "mysql", "mariadb":
		config.Language = MySQL
	case "pl/sql", "plsql", "oracle":
		config.Language = PLSQL
	case "db2":
		config.Language = DB2
	case "n1ql":
		config.Language = N1QL
	case "sqlite":
		config.Language = SQLite
	default:
		return fmt.Errorf("unknown language in config: %s", langStr)
	}
	return nil
}

func applyConfigKeywordCase(kcStr string, config *Config) error {
	switch strings.ToLower(kcStr) {
	case "preserve":
		config.KeywordCase = KeywordCasePreserve
	case "uppercase":
		config.KeywordCase = KeywordCaseUppercase
	case "lowercase":
		config.KeywordCase = KeywordCaseLowercase
	case "dialect":
		config.KeywordCase = KeywordCaseDialect
	default:
		return fmt.Errorf("unknown keyword_case in config: %s", kcStr)
	}
	return nil
}
