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

// ApplyToConfig applies the configuration file settings to a Config struct.
func (cf *ConfigFile) ApplyToConfig(config *Config) error {
	if cf.Language != "" {
		langStr := strings.ToLower(cf.Language)
		switch langStr {
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
			return fmt.Errorf("unknown language in config: %s", cf.Language)
		}
	}

	if cf.Indent != "" {
		config.Indent = cf.Indent
	}

	if cf.KeywordCase != "" {
		kcStr := strings.ToLower(cf.KeywordCase)
		switch kcStr {
		case "preserve":
			config.KeywordCase = KeywordCasePreserve
		case "uppercase":
			config.KeywordCase = KeywordCaseUppercase
		case "lowercase":
			config.KeywordCase = KeywordCaseLowercase
		case "dialect":
			config.KeywordCase = KeywordCaseDialect
		default:
			return fmt.Errorf("unknown keyword_case in config: %s", cf.KeywordCase)
		}
	}

	if cf.LinesBetweenQueries > 0 {
		config.LinesBetweenQueries = cf.LinesBetweenQueries
	}

	return nil
}
