package sqlfmt

import (
	"testing"
)

func TestDetectDialect(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		content      string
		expected     Language
		shouldDetect bool
	}{
		// File extension detection
		{
			name:         "PostgreSQL file extension .psql",
			filePath:     "query.psql",
			content:      "SELECT * FROM users",
			expected:     PostgreSQL,
			shouldDetect: true,
		},
		{
			name:         "PostgreSQL file extension .pgsql",
			filePath:     "query.pgsql",
			content:      "SELECT * FROM users",
			expected:     PostgreSQL,
			shouldDetect: true,
		},
		{
			name:         "MySQL file extension .mysql",
			filePath:     "query.mysql",
			content:      "SELECT * FROM users",
			expected:     MySQL,
			shouldDetect: true,
		},
		{
			name:         "MySQL compound extension",
			filePath:     "query.mysql.sql",
			content:      "SELECT * FROM users",
			expected:     MySQL,
			shouldDetect: true,
		},
		{
			name:         "SQLite file extension .sqlite",
			filePath:     "query.sqlite",
			content:      "SELECT * FROM users",
			expected:     SQLite,
			shouldDetect: true,
		},
		{
			name:         "PL/SQL file extension .plsql",
			filePath:     "query.plsql",
			content:      "SELECT * FROM users",
			expected:     PLSQL,
			shouldDetect: true,
		},
		{
			name:         "PL/SQL file extension .ora.sql",
			filePath:     "query.ora.sql",
			content:      "SELECT * FROM users",
			expected:     PLSQL,
			shouldDetect: true,
		},

		// Content-based detection - PostgreSQL
		{
			name:         "PostgreSQL type casting",
			filePath:     "query.sql",
			content:      "SELECT id::integer FROM users",
			expected:     PostgreSQL,
			shouldDetect: true,
		},
		{
			name:         "PostgreSQL dollar quoting",
			filePath:     "query.sql",
			content:      "SELECT $$hello world$$",
			expected:     PostgreSQL,
			shouldDetect: true,
		},
		{
			name:         "PostgreSQL positional parameters",
			filePath:     "query.sql",
			content:      "SELECT * FROM users WHERE id = $1",
			expected:     PostgreSQL,
			shouldDetect: true,
		},
		{
			name:         "PostgreSQL RETURNING clause",
			filePath:     "query.sql",
			content:      "INSERT INTO users (name) VALUES ('John') RETURNING id",
			expected:     PostgreSQL,
			shouldDetect: true,
		},
		{
			name:         "PostgreSQL JSON operators",
			filePath:     "query.sql",
			content:      "SELECT data->>'name' FROM users",
			expected:     PostgreSQL,
			shouldDetect: true,
		},

		// Content-based detection - MySQL
		{
			name:         "MySQL backticks",
			filePath:     "query.sql",
			content:      "SELECT `id`, `name` FROM `users`",
			expected:     MySQL,
			shouldDetect: true,
		},
		{
			name:         "MySQL ON DUPLICATE KEY UPDATE",
			filePath:     "query.sql",
			content:      "INSERT INTO users (id, name) VALUES (1, 'John') ON DUPLICATE KEY UPDATE name = 'John'",
			expected:     MySQL,
			shouldDetect: true,
		},
		{
			name:         "MySQL INSERT IGNORE",
			filePath:     "query.sql",
			content:      "INSERT IGNORE INTO users (id, name) VALUES (1, 'John')",
			expected:     MySQL,
			shouldDetect: true,
		},
		{
			name:         "MySQL ENGINE clause",
			filePath:     "query.sql",
			content:      "CREATE TABLE users (id INT) ENGINE=InnoDB",
			expected:     MySQL,
			shouldDetect: true,
		},

		// Content-based detection - SQLite
		{
			name:         "SQLite PRAGMA",
			filePath:     "query.sql",
			content:      "PRAGMA foreign_keys = ON",
			expected:     SQLite,
			shouldDetect: true,
		},
		{
			name:         "SQLite WITHOUT ROWID",
			filePath:     "query.sql",
			content:      "CREATE TABLE users (id INTEGER PRIMARY KEY) WITHOUT ROWID",
			expected:     SQLite,
			shouldDetect: true,
		},
		{
			name:         "SQLite AUTOINCREMENT",
			filePath:     "query.sql",
			content:      "CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT)",
			expected:     SQLite,
			shouldDetect: true,
		},
		{
			name:         "SQLite ATTACH",
			filePath:     "query.sql",
			content:      "ATTACH DATABASE 'other.db' AS other",
			expected:     SQLite,
			shouldDetect: true,
		},

		// Content-based detection - PL/SQL
		{
			name:         "PL/SQL BEGIN END block",
			filePath:     "query.sql",
			content:      "BEGIN SELECT * FROM users; END;",
			expected:     PLSQL,
			shouldDetect: true,
		},
		{
			name:         "PL/SQL EXCEPTION block",
			filePath:     "query.sql",
			content:      "BEGIN SELECT * FROM users; EXCEPTION WHEN OTHERS THEN NULL; END;",
			expected:     PLSQL,
			shouldDetect: true,
		},
		{
			name:         "PL/SQL PROCEDURE",
			filePath:     "query.sql",
			content:      "CREATE PROCEDURE test_proc AS BEGIN NULL; END;",
			expected:     PLSQL,
			shouldDetect: true,
		},
		{
			name:         "PL/SQL FUNCTION",
			filePath:     "query.sql",
			content:      "CREATE FUNCTION test_func RETURN NUMBER AS BEGIN RETURN 1; END;",
			expected:     PLSQL,
			shouldDetect: true,
		},

		// Fallback cases
		{
			name:         "Unknown extension falls back to content",
			filePath:     "query.unknown",
			content:      "SELECT id::integer FROM users",
			expected:     PostgreSQL,
			shouldDetect: true,
		},
		{
			name:         "No detection possible",
			filePath:     "query.sql",
			content:      "SELECT * FROM users WHERE id = 1",
			expected:     StandardSQL,
			shouldDetect: false,
		},
		{
			name:         "Empty content",
			filePath:     "query.sql",
			content:      "",
			expected:     StandardSQL,
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detected, ok := DetectDialect(tt.filePath, tt.content)

			if ok != tt.shouldDetect {
				t.Errorf("DetectDialect() detection success = %v, want %v", ok, tt.shouldDetect)
			}

			if detected != tt.expected {
				t.Errorf("DetectDialect() = %v, want %v", detected, tt.expected)
			}
		})
	}
}

func TestDetectFromFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected Language
		found    bool
	}{
		{"PostgreSQL .psql", "file.psql", PostgreSQL, true},
		{"PostgreSQL .pgsql", "file.pgsql", PostgreSQL, true},
		{"MySQL .mysql", "file.mysql", MySQL, true},
		{"MySQL .my.sql", "file.my.sql", MySQL, true},
		{"SQLite .sqlite", "file.sqlite", SQLite, true},
		{"SQLite .db.sql", "file.db.sql", SQLite, true},
		{"PL/SQL .plsql", "file.plsql", PLSQL, true},
		{"PL/SQL .ora.sql", "file.ora.sql", PLSQL, true},
		{"Unknown extension", "file.sql", StandardSQL, false},
		{"No extension", "file", StandardSQL, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang, found := detectFromFileExtension(tt.filePath)
			if found != tt.found || lang != tt.expected {
				t.Errorf("detectFromFileExtension(%s) = (%v, %v), want (%v, %v)",
					tt.filePath, lang, found, tt.expected, tt.found)
			}
		})
	}
}

func TestDetectFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected Language
		found    bool
	}{
		// PostgreSQL patterns
		{"PostgreSQL :: casting", "SELECT id::integer", PostgreSQL, true},
		{"PostgreSQL $$ quoting", "SELECT $$test$$", PostgreSQL, true},
		{"PostgreSQL $ parameters", "SELECT $1", PostgreSQL, true},
		{"PostgreSQL RETURNING", "INSERT ... RETURNING id", PostgreSQL, true},
		{"PostgreSQL JSON ->>", "SELECT data->>'key'", PostgreSQL, true},

		// MySQL patterns
		{"MySQL backticks", "SELECT `column`", MySQL, true},
		{"MySQL ON DUPLICATE", "INSERT ... ON DUPLICATE KEY UPDATE", MySQL, true},
		{"MySQL ENGINE", "CREATE TABLE ... ENGINE=InnoDB", MySQL, true},

		// SQLite patterns
		{"SQLite PRAGMA", "PRAGMA foreign_keys", SQLite, true},
		{"SQLite WITHOUT ROWID", "CREATE TABLE ... WITHOUT ROWID", SQLite, true},
		{"SQLite ATTACH", "ATTACH DATABASE", SQLite, true},

		// PL/SQL patterns
		{"PL/SQL BEGIN END", "BEGIN SELECT 1; END;", PLSQL, true},
		{"PL/SQL EXCEPTION", "EXCEPTION WHEN OTHERS", PLSQL, true},
		{"PL/SQL PROCEDURE", "CREATE PROCEDURE", PLSQL, true},

		// No detection
		{"Standard SQL", "SELECT * FROM users", StandardSQL, false},
		{"Empty", "", StandardSQL, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang, found := detectFromContent(tt.content)
			if found != tt.found || (found && lang != tt.expected) {
				t.Errorf("detectFromContent(%s) = (%v, %v), want (%v, %v)",
					tt.content, lang, found, tt.expected, tt.found)
			}
		})
	}
}
