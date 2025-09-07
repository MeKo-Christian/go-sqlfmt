package sqlfmt

import (
	"testing"
)

// Basic SQLite tests.
func TestSQLite_BasicDialect(t *testing.T) {
	// Test that SQLite language constant works
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Simple SELECT - the core formatting should work
	query := "SELECT id, name FROM users;"
	result := Format(query, cfg)
	expected := "SELECT\n  id,\n  name\nFROM\n  users;"

	if result != expected {
		t.Errorf("SQLite basic formatting failed.\nGot:\n%q\nWant:\n%q", result, expected)
	}
}

func TestSQLite_TokenizerConfig(t *testing.T) {
	// Test that tokenizer config is created
	cfg := NewSQLiteTokenizerConfig()
	if cfg == nil {
		t.Fatal("NewSQLiteTokenizerConfig() returned nil")
	}

	// Verify SQLite-specific features are configured

	// Test comment types - SQLite only supports --, not #
	foundDoubleHyphen := false
	foundHash := false
	for _, comment := range cfg.LineCommentTypes {
		if comment == "--" {
			foundDoubleHyphen = true
		}
		if comment == "#" {
			foundHash = true
		}
	}

	if !foundDoubleHyphen {
		t.Error("Expected SQLite to support -- line comments")
	}
	if foundHash {
		t.Error("SQLite should not support # line comments (MySQL/Standard SQL feature)")
	}

	// Test that indexed placeholders include ?
	foundQuestion := false
	for _, placeholder := range cfg.IndexedPlaceholderTypes {
		if placeholder == "?" {
			foundQuestion = true
			break
		}
	}
	if !foundQuestion {
		t.Error("Expected SQLite to support ? indexed placeholders")
	}

	// Test that named placeholders include :, @, $
	expectedNamed := []string{":", "@", "$"}
	for _, expected := range expectedNamed {
		found := false
		for _, placeholder := range cfg.NamedPlaceholderTypes {
			if placeholder == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected SQLite to support %s named placeholders", expected)
		}
	}

	// Test string types include SQLite-specific formats
	foundBlob := false
	for _, stringType := range cfg.StringTypes {
		if stringType == "X''" {
			foundBlob = true
			break
		}
	}
	if !foundBlob {
		t.Error("Expected SQLite to support X'' blob literals")
	}
}

func TestSQLite_FormatterCreation(t *testing.T) {
	// Test that we can create a SQLite formatter via the factory
	cfg := &Config{Language: SQLite, Indent: "  "}
	formatter := NewSQLiteFormatter(cfg)

	if formatter == nil {
		t.Fatal("NewSQLiteFormatter() returned nil")
	}

	// Test that it can format something basic
	result := formatter.Format("SELECT 1;")
	if result == "" {
		t.Error("SQLite formatter returned empty string")
	}
	if result == "SELECT 1;" {
		t.Error("SQLite formatter returned input unchanged - no formatting applied")
	}
}

func TestSQLite_LanguageSelection(t *testing.T) {
	// Test that the dialect factory selects SQLite correctly
	cfg := &Config{Language: SQLite, Indent: "  "}

	// This should use the SQLite formatter internally
	result1 := Format("SELECT id FROM users;", cfg)

	// This should use the standard SQL formatter
	standardCfg := &Config{Language: StandardSQL, Indent: "  "}
	result2 := Format("SELECT id FROM users;", standardCfg)

	// Both should format (not return unchanged), but using SQLite vs Standard config
	if result1 == "SELECT id FROM users;" {
		t.Error("SQLite formatter didn't format the query")
	}
	if result2 == "SELECT id FROM users;" {
		t.Error("Standard formatter didn't format the query")
	}

	// For this simple query, they might produce the same output, which is OK
	// The key is that both formatters are working
}

// Test a few SQLite-specific features.
func TestSQLite_BasicFeatures(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test SQLite comment style (-- only, not #)
	query := "SELECT id -- user id\nFROM users;"
	result := Format(query, cfg)

	if !containsString(result, "-- user id") {
		t.Error("SQLite should preserve -- comments")
	}

	// Test basic identifier quoting (double quotes should work)
	query2 := `SELECT "user_name" FROM users;`
	result2 := Format(query2, cfg)

	if !containsString(result2, `"user_name"`) {
		t.Error("SQLite should preserve double-quoted identifiers")
	}
}

// Helper function for tests.
func containsString(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
