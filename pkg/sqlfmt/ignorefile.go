package sqlfmt

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// IgnoreFile represents a .sqlfmtignore file with patterns to exclude.
type IgnoreFile struct {
	patterns []string
}

// LoadIgnoreFile attempts to load .sqlfmtignore from the current directory and parent directories.
func LoadIgnoreFile() (*IgnoreFile, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Search for .sqlfmtignore in current directory and parents
	for {
		ignorePath := filepath.Join(dir, ".sqlfmtignore")
		if content, err := os.ReadFile(ignorePath); err == nil {
			var patterns []string
			scanner := bufio.NewScanner(strings.NewReader(string(content)))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				// Skip empty lines and comments
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				patterns = append(patterns, line)
			}
			return &IgnoreFile{patterns: patterns}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// No ignore file found
	return &IgnoreFile{}, nil
}

// ShouldIgnore checks if a file path should be ignored based on the patterns.
func (ig *IgnoreFile) ShouldIgnore(filePath string) bool {
	if len(ig.patterns) == 0 {
		return false
	}

	// Convert to relative path for pattern matching
	relPath, err := filepath.Rel(".", filePath)
	if err != nil {
		relPath = filePath
	}

	// Normalize path separators
	relPath = filepath.ToSlash(relPath)

	for _, pattern := range ig.patterns {
		if ig.matchPattern(relPath, pattern) {
			return true
		}
	}

	return false
}

// matchPattern checks if a path matches a glob pattern (simplified gitignore-style matching).
func (ig *IgnoreFile) matchPattern(path, pattern string) bool {
	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		pattern = strings.TrimSuffix(pattern, "/")
		if strings.HasPrefix(path, pattern+"/") || path == pattern {
			return true
		}
	}

	// Simple glob matching
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err == nil && matched {
		return true
	}

	// Check if pattern matches full path
	matched, err = filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}

	// Handle ** for recursive directory matching
	if strings.Contains(pattern, "**") {
		return ig.matchGlobstar(path, pattern)
	}

	return false
}

// matchGlobstar handles ** patterns for recursive directory matching.
func (ig *IgnoreFile) matchGlobstar(path, pattern string) bool {
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return false
	}

	prefix := parts[0]
	suffix := parts[1]

	// Remove trailing slash from prefix if present
	prefix = strings.TrimSuffix(prefix, "/")
	// Remove leading slash from suffix if present
	suffix = strings.TrimPrefix(suffix, "/")

	// If prefix is not empty, path must start with prefix/
	if prefix != "" && !strings.HasPrefix(path, prefix+"/") {
		return false
	}

	// If suffix is empty, just check prefix match
	if suffix == "" {
		return prefix == "" || strings.HasPrefix(path, prefix+"/")
	}

	// Calculate remaining path after prefix
	var remaining string
	if prefix != "" {
		remaining = strings.TrimPrefix(path, prefix+"/")
	} else {
		remaining = path
	}

	// Check suffix match
	if strings.Contains(suffix, "/") {
		// Complex pattern like **/subdir/*.sql
		matched, err := filepath.Match(suffix, remaining)
		return err == nil && matched
	} else {
		// Simple pattern like **/*.sql
		matched, err := filepath.Match(suffix, filepath.Base(remaining))
		return err == nil && matched
	}
}
