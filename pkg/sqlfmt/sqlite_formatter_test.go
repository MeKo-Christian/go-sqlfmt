package sqlfmt

import (
	"testing"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/utils"
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

	// Break into smaller test functions to reduce complexity
	testSQLiteCommentTypes(t, cfg)
	testSQLitePlaceholderTypes(t, cfg)
	testSQLiteStringTypes(t, cfg)
}

func testSQLiteCommentTypes(t *testing.T, cfg *TokenizerConfig) {
	t.Helper()
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
}

func testSQLitePlaceholderTypes(t *testing.T, cfg *TokenizerConfig) {
	t.Helper()
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
}

func testSQLiteStringTypes(t *testing.T, cfg *TokenizerConfig) {
	t.Helper()
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

// Phase 2 comprehensive SQLite tests.
func TestSQLite_Phase2_Comments(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test -- line comments
	query1 := "SELECT id -- this is a comment\nFROM users;"
	result1 := Format(query1, cfg)
	if !containsString(result1, "-- this is a comment") {
		t.Error("SQLite should preserve -- line comments")
	}

	// Test /* */ block comments
	query2 := "SELECT /* block comment */ id FROM users;"
	result2 := Format(query2, cfg)
	if !containsString(result2, "/* block comment */") {
		t.Error("SQLite should preserve /* */ block comments")
	}

	// Test that # comments are NOT treated as comments (they should be tokens)
	query3 := "SELECT # as symbol FROM users;"
	result3 := Format(query3, cfg)
	// The # should be treated as a regular token, not filtered out as a comment
	if !containsString(result3, "#") {
		t.Error("SQLite should treat # as a regular token, not a comment")
	}
}

func TestSQLite_Phase2_IdentifierQuoting(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test double-quoted identifiers (standard SQL)
	query1 := `SELECT "user_name", "user id" FROM "my table";`
	result1 := Format(query1, cfg)
	if !containsString(result1, `"user_name"`) ||
		!containsString(result1, `"user id"`) ||
		!containsString(result1, `"my table"`) {
		t.Error("SQLite should preserve double-quoted identifiers")
	}

	// Test backtick identifiers (MySQL compatibility)
	query2 := "SELECT `column_name`, `spaced column` FROM `table_name`;"
	result2 := Format(query2, cfg)
	if !containsString(result2, "`column_name`") ||
		!containsString(result2, "`spaced column`") ||
		!containsString(result2, "`table_name`") {
		t.Error("SQLite should preserve backtick identifiers")
	}

	// Test bracket identifiers (SQL Server compatibility)
	query3 := "SELECT [column_name], [spaced column] FROM [table name];"
	result3 := Format(query3, cfg)
	if !containsString(result3, "[column_name]") ||
		!containsString(result3, "[spaced column]") ||
		!containsString(result3, "[table name]") {
		t.Error("SQLite should preserve bracket identifiers")
	}
}

func TestSQLite_Phase2_StringsAndBlobs(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test single-quoted strings
	query1 := "SELECT 'hello world', 'with ''escaped'' quotes' FROM users;"
	result1 := Format(query1, cfg)
	if !containsString(result1, "'hello world'") ||
		!containsString(result1, "'with ''escaped'' quotes'") {
		t.Error("SQLite should preserve single-quoted strings")
	}

	// Test blob literals with uppercase X
	query2 := "SELECT X'DEADBEEF', X'48656C6C6F' as hex_data;"
	result2 := Format(query2, cfg)
	if !containsString(result2, "X'DEADBEEF'") ||
		!containsString(result2, "X'48656C6C6F'") {
		t.Error("SQLite should preserve uppercase X'' blob literals")
	}

	// Test blob literals with lowercase x
	query3 := "SELECT x'deadbeef', x'48656c6c6f' as hex_data;"
	result3 := Format(query3, cfg)
	if !containsString(result3, "x'deadbeef'") ||
		!containsString(result3, "x'48656c6c6f'") {
		t.Error("SQLite should preserve lowercase x'' blob literals")
	}
}

func TestSQLite_Phase2_Numbers(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test various number formats
	query := "SELECT 123, 123.456, -789, -12.34, 0.001 FROM numbers;"
	result := Format(query, cfg)

	numbers := []string{"123", "123.456", "-789", "-12.34", "0.001"}
	for _, num := range numbers {
		if !containsString(result, num) {
			t.Errorf("SQLite should preserve number: %s", num)
		}
	}
}

func TestSQLite_Phase2_IntegratedExample(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test a comprehensive example with all Phase 2 features
	query := `-- SQLite test query
SELECT
    "user_id",           -- double quotes
    ` + "`user_name`," + `          -- backticks
    [full name],         -- brackets with space
    'text data',         -- string
    X'DEADBEEF',         -- blob literal
    123.45               -- number
FROM /* table comment */ "users"
WHERE ` + "`status`" + ` = 'active'
  AND [user id] > 0;`

	result := Format(query, cfg)

	// Check that all identifier styles are preserved
	expectedElements := []string{
		"-- SQLite test query",
		`"user_id"`,
		"`user_name`",
		"[full name]",
		"'text data'",
		"X'DEADBEEF'",
		"123.45",
		"/* table comment */",
		`"users"`,
		"`status`",
		"'active'",
		"[user id]",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("SQLite integrated example should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

// Phase 3: Comprehensive Placeholder Tests.
func TestSQLite_Phase3_AllPlaceholderForms(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test basic ? placeholder
	query1 := "SELECT * FROM users WHERE id = ?;"
	result1 := Format(query1, cfg)
	if !containsString(result1, "= ?") {
		t.Error("SQLite should preserve ? placeholder")
	}

	// Test numbered ?NNN placeholders
	query2 := "SELECT * FROM users WHERE id = ?1 AND name = ?2 AND age > ?10;"
	result2 := Format(query2, cfg)
	if !containsString(result2, "= ?1") || !containsString(result2, "= ?2") || !containsString(result2, "> ?10") {
		t.Error("SQLite should preserve ?NNN numbered placeholders")
	}

	// Test named :name placeholders
	query3 := "SELECT * FROM users WHERE id = :user_id AND name = :user_name;"
	result3 := Format(query3, cfg)
	if !containsString(result3, "= :user_id") || !containsString(result3, "= :user_name") {
		t.Error("SQLite should preserve :name placeholders")
	}

	// Test named @name placeholders
	query4 := "SELECT * FROM users WHERE id = @user_id AND name = @user_name;"
	result4 := Format(query4, cfg)
	if !containsString(result4, "= @user_id") || !containsString(result4, "= @user_name") {
		t.Error("SQLite should preserve @name placeholders")
	}

	// Test named $name placeholders
	query5 := "SELECT * FROM users WHERE id = $user_id AND name = $user_name;"
	result5 := Format(query5, cfg)
	if !containsString(result5, "= $user_id") || !containsString(result5, "= $user_name") {
		t.Error("SQLite should preserve $name placeholders")
	}
}

func TestSQLite_Phase3_MixedPlaceholders(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test mixing different placeholder styles in one query
	query := `SELECT * FROM users
		WHERE id = ?1
		AND name = :user_name
		AND email = @email_addr
		AND status = $status
		AND created > ?;`

	result := Format(query, cfg)

	// Check that all placeholder types are preserved
	expectedPlaceholders := []string{"?1", ":user_name", "@email_addr", "$status", "?"}
	for _, placeholder := range expectedPlaceholders {
		if !containsString(result, placeholder) {
			t.Errorf("Mixed placeholders test should contain: %s\nFull result:\n%s", placeholder, result)
		}
	}
}

func TestSQLite_Phase3_PlaceholderIsolation(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test that placeholders inside strings are NOT treated as placeholders
	query1 := `SELECT 'This is a ?1 string with :name and @email and $var' FROM users WHERE id = ?1;`
	result1 := Format(query1, cfg)

	// The string should be preserved as-is, but the real placeholder should work
	if !containsString(result1, "'This is a ?1 string with :name and @email and $var'") {
		t.Error("SQLite should preserve placeholders inside strings as literal text")
	}
	if !containsString(result1, "id = ?1") {
		t.Error("SQLite should process real placeholders outside strings")
	}

	// Test that placeholders inside comments are NOT treated as placeholders
	query2 := `SELECT * FROM users -- This comment has ?1 and :name and @email
		WHERE id = ?1 /* Another comment with $var */ AND name = :real_name;`
	result2 := Format(query2, cfg)

	// Comments with placeholders should be preserved, real placeholders should work
	if !containsString(result2, "-- This comment has ?1 and :name and @email") {
		t.Error("SQLite should preserve placeholders inside comments as literal text")
	}
	if !containsString(result2, "/* Another comment with $var */") {
		t.Error("SQLite should preserve placeholders inside block comments")
	}
	if !containsString(result2, "id = ?1") || !containsString(result2, "= :real_name") {
		t.Error("SQLite should process real placeholders outside comments")
	}
}

func TestSQLite_Phase3_EdgeCases(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test ?0 (technically valid in SQLite, though unusual)
	query1 := "SELECT * FROM users WHERE id = ?0;"
	result1 := Format(query1, cfg)
	if !containsString(result1, "= ?0") {
		t.Error("SQLite should handle ?0 placeholder")
	}

	// Test high numbers
	query2 := "SELECT * FROM users WHERE id = ?999 AND name = ?1000;"
	result2 := Format(query2, cfg)
	if !containsString(result2, "= ?999") || !containsString(result2, "= ?1000") {
		t.Error("SQLite should handle high-numbered placeholders")
	}

	// Test named placeholders with underscores, numbers, dots
	query3 := "SELECT * FROM users WHERE id = :user_id_123 AND data = @data.field AND config = $config_v2;"
	result3 := Format(query3, cfg)
	if !containsString(result3, ":user_id_123") ||
		!containsString(result3, "@data.field") ||
		!containsString(result3, "$config_v2") {
		t.Error("SQLite should handle named placeholders with special characters")
	}

	// Test edge case: ensure ? near JSON operators doesn't interfere
	query4 := "SELECT data->'key' FROM users WHERE id = ? AND json_data ?| array['key1', 'key2'];"
	result4 := Format(query4, cfg)
	if !containsString(result4, "id = ?") {
		t.Error("SQLite should preserve ? placeholder even near JSON operators")
	}
}

func TestSQLite_Phase3_ParameterSubstitution(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test parameter substitution with basic placeholders
	query := "SELECT * FROM users WHERE id = ? AND name = :name AND email = @email AND status = $status;"

	// This test mainly ensures the formatting doesn't break parameter placeholders
	// The actual parameter substitution logic would be tested separately
	result := Format(query, cfg)

	// Verify all placeholder forms are preserved for later substitution
	if !containsString(result, "= ?") {
		t.Error("Parameter substitution test: ? placeholder should be preserved")
	}
	if !containsString(result, "= :name") {
		t.Error("Parameter substitution test: :name placeholder should be preserved")
	}
	if !containsString(result, "= @email") {
		t.Error("Parameter substitution test: @email placeholder should be preserved")
	}
	if !containsString(result, "= $status") {
		t.Error("Parameter substitution test: $status placeholder should be preserved")
	}
}

func TestSQLite_Phase3_ComplexPlaceholderScenarios(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test complex query with all features from previous phases + placeholders
	query := `-- SQLite comprehensive test with placeholders
SELECT
    "user_id",           -- double-quoted identifier
    ` + "`user_name`," + `          -- backtick identifier
    [full name],         -- bracket identifier
    'Hello ' || 'World', -- string concatenation
    X'DEADBEEF'          -- blob literal
FROM /* users table */ "users"
WHERE ` + "`user_id`" + ` = ?1          -- numbered placeholder
  AND [full name] = :full_name   -- named placeholder with colon
  AND email = @email_addr        -- named placeholder with @
  AND status = $user_status      -- named placeholder with $
  AND created > ?                -- basic ? placeholder
  AND data->'key' IS NOT NULL    -- JSON operator (should not conflict)
  AND 'not a :placeholder' = 'literal'; -- placeholder in string (ignored)`

	result := Format(query, cfg)

	// Test that all elements are preserved
	expectedElements := []string{
		"-- SQLite comprehensive test with placeholders",
		`"user_id"`,
		"`user_name`",
		"[full name]",
		"'Hello '",
		"||",
		"'World'",
		"X'DEADBEEF'",
		"/* users table */",
		`"users"`,
		"= ?1",
		"= :full_name",
		"= @email_addr",
		"= $user_status",
		"> ?",
		"data -> 'key'",
		"IS NOT NULL",
		"'not a :placeholder'",
		"'literal'",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Complex placeholder scenario should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

// Test SQLite 1-based indexing vs other dialects' 0-based indexing.
func TestSQLite_Phase3_IndexingBehavior(t *testing.T) {
	// Test that SQLite uses 1-based indexing for numbered placeholders
	sqliteParams := &utils.ParamsConfig{
		ListParams:        []string{"first", "second", "third"},
		UseSQLiteIndexing: true,
	}
	sqliteParamsObj := utils.NewParams(sqliteParams)

	// SQLite: ?1 should return "first" (ListParams[0])
	result1 := sqliteParamsObj.Get("1", "default")
	if result1 != "first" {
		t.Errorf("SQLite ?1 should map to ListParams[0] ('first'), got: %s", result1)
	}

	// SQLite: ?2 should return "second" (ListParams[1])
	result2 := sqliteParamsObj.Get("2", "default")
	if result2 != "second" {
		t.Errorf("SQLite ?2 should map to ListParams[1] ('second'), got: %s", result2)
	}

	// Test that standard SQL uses 0-based indexing
	standardParams := &utils.ParamsConfig{
		ListParams:        []string{"first", "second", "third"},
		UseSQLiteIndexing: false, // Default behavior
	}
	standardParamsObj := utils.NewParams(standardParams)

	// Standard SQL: ?0 should return "first" (ListParams[0])
	result0 := standardParamsObj.Get("0", "default")
	if result0 != "first" {
		t.Errorf("Standard SQL ?0 should map to ListParams[0] ('first'), got: %s", result0)
	}

	// Standard SQL: ?1 should return "second" (ListParams[1])
	result1_std := standardParamsObj.Get("1", "default")
	if result1_std != "second" {
		t.Errorf("Standard SQL ?1 should map to ListParams[1] ('second'), got: %s", result1_std)
	}

	// Edge case: SQLite ?0 should return default (out of range for 1-based)
	sqliteResult0 := sqliteParamsObj.Get("0", "default")
	if sqliteResult0 != "default" {
		t.Errorf("SQLite ?0 should return default value (out of range), got: %s", sqliteResult0)
	}
}

// Phase 4: Operators & Specials Tests.
func TestSQLite_Phase4_ConcatenationOperator(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test basic string concatenation
	query1 := "SELECT 'Hello' || ' ' || 'World' FROM users;"
	result1 := Format(query1, cfg)
	if !containsString(result1, "'Hello' || ' ' || 'World'") {
		t.Error("SQLite should format concatenation operator with proper spacing")
	}

	// Test concatenation with identifiers
	query2 := "SELECT first_name || ' ' || last_name as full_name FROM users;"
	result2 := Format(query2, cfg)
	if !containsString(result2, "first_name || ' ' || last_name") {
		t.Error("SQLite should format concatenation with identifiers")
	}

	// Test concatenation in WHERE clause
	query3 := "SELECT * FROM users WHERE first_name || last_name = 'JohnDoe';"
	result3 := Format(query3, cfg)
	if !containsString(result3, "first_name || last_name = 'JohnDoe'") {
		t.Error("SQLite should format concatenation in WHERE clauses")
	}
}

func TestSQLite_Phase4_JSONOperators(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test -> operator for JSON path
	query1 := "SELECT data->'key', info->'nested'->'field' FROM users;"
	result1 := Format(query1, cfg)
	if !containsString(result1, "data -> 'key'") || !containsString(result1, "info -> 'nested' -> 'field'") {
		t.Error("SQLite should format -> JSON operator with proper spacing")
	}

	// Test ->> operator for JSON text extraction
	query2 := "SELECT data->>'text_field', info->>'name' FROM users;"
	result2 := Format(query2, cfg)
	if !containsString(result2, "data ->> 'text_field'") || !containsString(result2, "info ->> 'name'") {
		t.Error("SQLite should format ->> JSON operator with proper spacing")
	}

	// Test mixed JSON operators
	query3 := "SELECT data->'user'->>'name', info->'settings'->'theme' FROM users WHERE id = ?;"
	result3 := Format(query3, cfg)
	if !containsString(result3, "data -> 'user' ->> 'name'") ||
		!containsString(result3, "info -> 'settings' -> 'theme'") {
		t.Error("SQLite should format mixed JSON operators correctly")
	}

	// Test JSON operators with placeholders (should not interfere)
	query4 := "SELECT data->'key' FROM users WHERE id = ? AND data ?| array['key1'];"
	result4 := Format(query4, cfg)
	if !containsString(result4, "data -> 'key'") || !containsString(result4, "id = ?") {
		t.Error("SQLite JSON operators should not interfere with placeholders")
	}
}

func TestSQLite_Phase4_NullHandling(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test IS NULL
	query1 := "SELECT * FROM users WHERE name IS NULL;"
	result1 := Format(query1, cfg)
	if !containsString(result1, "name IS NULL") {
		t.Error("SQLite should format IS NULL correctly")
	}

	// Test IS NOT NULL
	query2 := "SELECT * FROM users WHERE email IS NOT NULL;"
	result2 := Format(query2, cfg)
	if !containsString(result2, "email IS NOT NULL") {
		t.Error("SQLite should format IS NOT NULL correctly")
	}

	// Test IS DISTINCT FROM (SQLite 3.39+)
	query3 := "SELECT * FROM users WHERE status IS DISTINCT FROM 'active';"
	result3 := Format(query3, cfg)
	if !containsString(result3, "status IS DISTINCT FROM 'active'") {
		t.Error("SQLite should format IS DISTINCT FROM as single phrase")
	}

	// Test IS NOT DISTINCT FROM (SQLite 3.39+)
	query4 := "SELECT * FROM users WHERE status IS NOT DISTINCT FROM 'active';"
	result4 := Format(query4, cfg)
	if !containsString(result4, "status IS NOT DISTINCT FROM 'active'") {
		t.Error("SQLite should format IS NOT DISTINCT FROM as single phrase")
	}

	// Test complex NULL expressions
	query5 := "SELECT * FROM users WHERE name IS NULL OR email IS NOT DISTINCT FROM previous_email;"
	result5 := Format(query5, cfg)
	if !containsString(result5, "name IS NULL") ||
		!containsString(result5, "email IS NOT DISTINCT FROM previous_email") {
		t.Error("SQLite should handle complex NULL expressions correctly")
	}
}

func TestSQLite_Phase4_PatternMatching(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test LIKE pattern matching
	query1 := "SELECT * FROM users WHERE name LIKE 'John%';"
	result1 := Format(query1, cfg)
	if !containsString(result1, "name LIKE 'John%'") {
		t.Error("SQLite should format LIKE pattern matching correctly")
	}

	// Test NOT LIKE
	query2 := "SELECT * FROM users WHERE name NOT LIKE '%admin%';"
	result2 := Format(query2, cfg)
	if !containsString(result2, "name NOT LIKE '%admin%'") {
		t.Error("SQLite should format NOT LIKE correctly")
	}

	// Test GLOB pattern matching (SQLite-specific)
	query3 := "SELECT * FROM users WHERE name GLOB 'John*';"
	result3 := Format(query3, cfg)
	if !containsString(result3, "name GLOB 'John*'") {
		t.Error("SQLite should format GLOB pattern matching correctly")
	}

	// Test REGEXP (should be treated as identifier/function, not special operator)
	query4 := "SELECT * FROM users WHERE name REGEXP '^John.*';"
	result4 := Format(query4, cfg)
	if !containsString(result4, "name REGEXP '^John.*'") {
		t.Error("SQLite should treat REGEXP as regular identifier/function")
	}

	// Test case-insensitive pattern matching
	query5 := "SELECT * FROM users WHERE UPPER(name) LIKE UPPER('john%') ESCAPE '\\';"
	result5 := Format(query5, cfg)
	if !containsString(result5, "UPPER(name) LIKE UPPER('john%') ESCAPE '\\'") {
		t.Error("SQLite should handle complex LIKE expressions with ESCAPE")
	}
}

func TestSQLite_Phase4_IntegratedExample(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test comprehensive example with all Phase 4 features
	query := `-- Phase 4: Operators and Specials test
SELECT 
    first_name || ' ' || last_name as full_name,   -- concatenation
    data->'profile'->>'name' as profile_name,      -- JSON operators  
    CASE 
        WHEN email IS NULL THEN 'No email'
        WHEN status IS DISTINCT FROM 'active' THEN 'Inactive'
        ELSE 'Active'
    END as user_status
FROM users 
WHERE name LIKE 'John%'                            -- pattern matching
  AND data->'settings'->>'theme' IS NOT NULL      -- JSON + NULL handling
  AND created_at IS NOT DISTINCT FROM updated_at  -- NULL-safe comparison  
  AND name GLOB 'J*'                               -- SQLite GLOB
  AND notes REGEXP 'important|urgent'             -- REGEXP as function
ORDER BY first_name || last_name;` // concatenation in ORDER BY

	result := Format(query, cfg)

	// Test that all features are preserved and formatted correctly
	expectedElements := []string{
		"-- Phase 4: Operators and Specials test",
		"first_name || ' ' || last_name",
		"data -> 'profile' ->> 'name'",
		"email IS NULL",
		"status IS DISTINCT FROM 'active'",
		"name LIKE 'John%'",
		"data -> 'settings' ->> 'theme' IS NOT NULL",
		"created_at IS NOT DISTINCT FROM updated_at",
		"name GLOB 'J*'",
		"notes REGEXP 'important|urgent'",
		"first_name || last_name",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Phase 4 integrated example should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

func TestSQLite_Phase4_EdgeCases(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test operators in strings (should be ignored)
	query1 := `SELECT 'String with || and -> operators' FROM users WHERE name = 'test';`
	result1 := Format(query1, cfg)
	if !containsString(result1, "'String with || and -> operators'") {
		t.Error("SQLite should preserve operators inside strings as literal text")
	}

	// Test operators in comments (should be ignored)
	query2 := `SELECT name /* comment with || and -> */ FROM users; -- another || comment`
	result2 := Format(query2, cfg)
	if !containsString(result2, "/* comment with || and -> */") ||
		!containsString(result2, "-- another || comment") {
		t.Error("SQLite should preserve operators inside comments as literal text")
	}

	// Test operator precedence and grouping
	query3 := "SELECT (first_name || ' ') || (last_name || suffix) FROM users;"
	result3 := Format(query3, cfg)
	if !containsString(result3, "(first_name || ' ') || (last_name || suffix)") {
		t.Error("SQLite should preserve operator grouping with parentheses")
	}

	// Test JSON operators with complex expressions
	query4 := "SELECT data->(CASE WHEN type = 'user' THEN 'profile' ELSE 'settings' END)->>'name' FROM users;"
	result4 := Format(query4, cfg)
	// This complex case should be formatted reasonably - exact formatting may vary
	if !containsString(result4, "data ->") || !containsString(result4, "->> 'name'") {
		t.Error("SQLite should handle JSON operators with complex expressions")
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
