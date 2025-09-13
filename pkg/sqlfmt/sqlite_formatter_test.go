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

// Phase 5: Core Clauses Tests.
func TestSQLite_Phase5_LimitVariations(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test standard LIMIT OFFSET syntax
	query1 := "SELECT * FROM users LIMIT 10 OFFSET 5;"
	result1 := Format(query1, cfg)
	if !containsString(result1, "LIMIT") || !containsString(result1, "10 OFFSET 5") {
		t.Error("SQLite should format standard LIMIT OFFSET syntax")
	}

	// Test MySQL-style LIMIT with comma syntax (SQLite also supports this)
	query2 := "SELECT * FROM users LIMIT 5, 10;"
	result2 := Format(query2, cfg)
	// The formatter should preserve the comma syntax as valid SQLite
	if !containsString(result2, "LIMIT") || !containsString(result2, "5, 10") {
		t.Error("SQLite should preserve LIMIT with comma syntax")
	}

	// Test LIMIT without OFFSET
	query3 := "SELECT * FROM users LIMIT 20;"
	result3 := Format(query3, cfg)
	if !containsString(result3, "LIMIT") || !containsString(result3, "20") {
		t.Error("SQLite should format LIMIT without OFFSET")
	}

	// Test LIMIT with placeholder
	query4 := "SELECT * FROM users LIMIT ? OFFSET ?;"
	result4 := Format(query4, cfg)
	if !containsString(result4, "LIMIT") || !containsString(result4, "? OFFSET ?") {
		t.Error("SQLite should format LIMIT with placeholders")
	}

	// Test LIMIT in complex query
	query5 := `SELECT name, COUNT(*) as cnt 
		FROM users 
		WHERE active = 1 
		GROUP BY name 
		ORDER BY cnt DESC 
		LIMIT 10;`
	result5 := Format(query5, cfg)
	if !containsString(result5, "LIMIT") || !containsString(result5, "10") {
		t.Error("SQLite should format LIMIT in complex queries")
	}
}

func TestSQLite_Phase5_UpsertFeatures(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test INSERT OR REPLACE
	query1 := "INSERT OR REPLACE INTO users (id, name, email) VALUES (1, 'John', 'john@example.com');"
	result1 := Format(query1, cfg)
	if !containsString(result1, "INSERT OR REPLACE") {
		t.Error("SQLite should format INSERT OR REPLACE correctly")
	}

	// Test INSERT OR IGNORE
	query2 := "INSERT OR IGNORE INTO users (id, name, email) VALUES (2, 'Jane', 'jane@example.com');"
	result2 := Format(query2, cfg)
	if !containsString(result2, "INSERT OR IGNORE") {
		t.Error("SQLite should format INSERT OR IGNORE correctly")
	}

	// Test INSERT OR ABORT
	query3 := "INSERT OR ABORT INTO users (id, name) VALUES (3, 'Bob');"
	result3 := Format(query3, cfg)
	if !containsString(result3, "INSERT OR ABORT") {
		t.Error("SQLite should format INSERT OR ABORT correctly")
	}

	// Test INSERT OR FAIL
	query4 := "INSERT OR FAIL INTO users (id, name) VALUES (4, 'Alice');"
	result4 := Format(query4, cfg)
	if !containsString(result4, "INSERT OR FAIL") {
		t.Error("SQLite should format INSERT OR FAIL correctly")
	}

	// Test INSERT OR ROLLBACK
	query5 := "INSERT OR ROLLBACK INTO users (id, name) VALUES (5, 'Charlie');"
	result5 := Format(query5, cfg)
	if !containsString(result5, "INSERT OR ROLLBACK") {
		t.Error("SQLite should format INSERT OR ROLLBACK correctly")
	}

	// Test ON CONFLICT DO NOTHING
	query6 := "INSERT INTO users (id, name, email) VALUES (1, 'John', 'john@example.com') ON CONFLICT (id) DO NOTHING;"
	result6 := Format(query6, cfg)
	if !containsString(result6, "ON CONFLICT") || !containsString(result6, "DO NOTHING") {
		t.Error("SQLite should format ON CONFLICT DO NOTHING correctly")
	}

	// Test ON CONFLICT DO UPDATE
	query7 := `INSERT INTO users (id, name, email) VALUES (1, 'John', 'john@example.com') 
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email;`
	result7 := Format(query7, cfg)
	if !containsString(result7, "ON CONFLICT") || !containsString(result7, "DO UPDATE") {
		t.Error("SQLite should format ON CONFLICT DO UPDATE correctly")
	}

	// Test ON CONFLICT with multiple columns
	query8 := `INSERT INTO user_scores (user_id, game_id, score) VALUES (1, 100, 50) 
		ON CONFLICT (user_id, game_id) DO UPDATE SET score = MAX(score, EXCLUDED.score);`
	result8 := Format(query8, cfg)
	if !containsString(result8, "ON CONFLICT") || !containsString(result8, "(user_id, game_id)") {
		t.Error("SQLite should format ON CONFLICT with multiple columns")
	}

	// Test ON CONFLICT with WHERE clause
	query9 := `INSERT INTO products (sku, name, price) VALUES ('ABC123', 'Product A', 29.99)
		ON CONFLICT (sku) DO UPDATE SET price = EXCLUDED.price WHERE EXCLUDED.price < price;`
	result9 := Format(query9, cfg)
	if !containsString(result9, "ON CONFLICT") || !containsString(result9, "(sku)") || !containsString(result9, "DO UPDATE") {
		t.Error("SQLite should format ON CONFLICT with WHERE clause")
	}
}

func TestSQLite_Phase5_WithoutRowid(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test WITHOUT ROWID table creation
	query1 := `CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE
	) WITHOUT ROWID;`
	result1 := Format(query1, cfg)
	if !containsString(result1, "WITHOUT ROWID") {
		t.Error("SQLite should format WITHOUT ROWID table option")
	}

	// Test WITHOUT ROWID with STRICT
	query2 := `CREATE TABLE products (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		price REAL
	) STRICT, WITHOUT ROWID;`
	result2 := Format(query2, cfg)
	if !containsString(result2, "WITHOUT ROWID") {
		t.Error("SQLite should format WITHOUT ROWID with STRICT")
	}

	// Test regular table (should not have WITHOUT ROWID)
	query3 := `CREATE TABLE logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	result3 := Format(query3, cfg)
	if containsString(result3, "WITHOUT ROWID") {
		t.Error("Regular table should not have WITHOUT ROWID automatically added")
	}
}

func TestSQLite_Phase5_IntegratedExample(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test comprehensive Phase 5 example with all features
	query := `-- Phase 5: Core Clauses comprehensive test

-- Create table without rowid
CREATE TABLE user_preferences (
    user_id INTEGER,
    setting_key TEXT,
    setting_value TEXT,
    PRIMARY KEY (user_id, setting_key)
) WITHOUT ROWID;

-- Use INSERT OR REPLACE for upsert-like behavior
INSERT OR REPLACE INTO user_preferences (user_id, setting_key, setting_value)
VALUES (1, 'theme', 'dark'),
       (1, 'notifications', 'enabled'),
       (2, 'theme', 'light');

-- Use modern UPSERT syntax with ON CONFLICT
INSERT INTO user_preferences (user_id, setting_key, setting_value)
VALUES (1, 'language', 'en')
ON CONFLICT (user_id, setting_key) 
DO UPDATE SET setting_value = EXCLUDED.setting_value,
              updated_at = CURRENT_TIMESTAMP;

-- Use both LIMIT styles in different queries  
SELECT * FROM user_preferences WHERE user_id = 1 LIMIT 5 OFFSET 0;
SELECT * FROM user_preferences WHERE setting_key = 'theme' LIMIT 0, 10;

-- Complex query with multiple Phase 5 features
INSERT INTO audit_log (user_id, action, details) 
SELECT user_id, 'preference_change' as action, 
       'Updated ' || setting_key || ' to ' || setting_value as details
FROM user_preferences 
WHERE user_id = ? 
ORDER BY setting_key 
LIMIT 100
ON CONFLICT (user_id, action, created_at) DO NOTHING;`

	result := Format(query, cfg)

	// Test that all Phase 5 features are preserved and formatted
	expectedElements := []string{
		"-- Phase 5: Core Clauses comprehensive test",
		"WITHOUT ROWID",
		"INSERT OR REPLACE",
		"ON CONFLICT",
		"(user_id, setting_key)",
		"DO UPDATE",
		"EXCLUDED.setting_value",
		"5 OFFSET 0",
		"0, 10",
		"100",
		"DO NOTHING",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Phase 5 integrated example should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

func TestSQLite_Phase5_EdgeCases(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test nested UPSERT with complex expressions
	query1 := `INSERT INTO stats (user_id, metric, value) 
		VALUES (?, 'login_count', 1) 
		ON CONFLICT (user_id, metric) 
		DO UPDATE SET value = value + EXCLUDED.value, updated_at = datetime('now');`
	result1 := Format(query1, cfg)
	if !containsString(result1, "ON CONFLICT") || !containsString(result1, "(user_id, metric)") || !containsString(result1, "DO UPDATE") {
		t.Error("SQLite should handle complex UPSERT expressions")
	}

	// Test LIMIT with expressions
	query2 := "SELECT * FROM users LIMIT (SELECT COUNT(*) FROM users) / 2 OFFSET 10;"
	result2 := Format(query2, cfg)
	if !containsString(result2, "LIMIT") || !containsString(result2, "OFFSET 10") {
		t.Error("SQLite should handle LIMIT with complex expressions")
	}

	// Test INSERT OR with subquery
	query3 := `INSERT OR IGNORE INTO user_backup 
		SELECT * FROM users WHERE created_at > datetime('now', '-1 day');`
	result3 := Format(query3, cfg)
	if !containsString(result3, "INSERT OR IGNORE") {
		t.Error("SQLite should handle INSERT OR with subqueries")
	}

	// Test multiple ON CONFLICT clauses (shouldn't exist but should handle gracefully)
	query4 := `INSERT INTO complex_table (a, b, c) VALUES (1, 2, 3) 
		ON CONFLICT (a) DO UPDATE SET b = EXCLUDED.b 
		WHERE EXCLUDED.c > c;`
	result4 := Format(query4, cfg)
	if !containsString(result4, "ON CONFLICT") || !containsString(result4, "(a)") {
		t.Error("SQLite should handle ON CONFLICT with WHERE conditions")
	}

	// Test UPSERT with placeholders and complex WHERE
	query5 := `INSERT INTO user_settings (user_id, key, value) 
		VALUES (:user_id, :key, :value)
		ON CONFLICT (user_id, key) DO UPDATE SET 
		value = CASE 
			WHEN :force_update = 1 THEN EXCLUDED.value 
			WHEN json_valid(EXCLUDED.value) THEN EXCLUDED.value 
			ELSE value 
		END;`
	result5 := Format(query5, cfg)
	if !containsString(result5, "ON CONFLICT") || !containsString(result5, "(user_id, key)") || !containsString(result5, ":force_update") {
		t.Error("SQLite should handle complex UPSERT with placeholders and CASE expressions")
	}
}

// Phase 6: CTEs & Window Functions Tests.
func TestSQLite_Phase6_BasicCTE(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test simple WITH clause
	query1 := `WITH active_users AS (
		SELECT user_id, name FROM users WHERE active = 1
	) SELECT * FROM active_users;`
	result1 := Format(query1, cfg)
	if !containsString(result1, "WITH") || !containsString(result1, "active_users AS") {
		t.Error("SQLite should format basic CTE with WITH clause")
	}

	// Test multiple CTEs
	query2 := `WITH 
		active_users AS (SELECT user_id, name FROM users WHERE active = 1),
		orders_summary AS (SELECT user_id, COUNT(*) as order_count FROM orders GROUP BY user_id)
	SELECT au.name, os.order_count FROM active_users au JOIN orders_summary os ON au.user_id = os.user_id;`
	result2 := Format(query2, cfg)
	if !containsString(result2, "WITH") || !containsString(result2, "active_users AS") || !containsString(result2, "orders_summary AS") {
		t.Error("SQLite should format multiple CTEs correctly")
	}
}

func TestSQLite_Phase6_RecursiveCTE(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test WITH RECURSIVE for hierarchical queries
	query := `WITH RECURSIVE employee_hierarchy(id, name, manager_id, level) AS (
		-- Anchor: top-level employees (no manager)
		SELECT id, name, manager_id, 0 as level
		FROM employees 
		WHERE manager_id IS NULL
		
		UNION ALL
		
		-- Recursive: employees with managers
		SELECT e.id, e.name, e.manager_id, eh.level + 1
		FROM employees e
		JOIN employee_hierarchy eh ON e.manager_id = eh.id
	)
	SELECT id, name, level FROM employee_hierarchy ORDER BY level, name;`

	result := Format(query, cfg)

	// Check that WITH RECURSIVE is handled correctly
	if !containsString(result, "WITH") || !containsString(result, "RECURSIVE") {
		t.Error("SQLite should handle WITH RECURSIVE keywords")
	}
	if !containsString(result, "employee_hierarchy") {
		t.Error("SQLite should preserve CTE name in recursive query")
	}
	if !containsString(result, "UNION ALL") {
		t.Error("SQLite should handle UNION ALL in recursive CTE")
	}
}

func TestSQLite_Phase6_CTEWithPlaceholders(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test CTE with all SQLite placeholder types
	query := `WITH filtered_users AS (
		SELECT user_id, name, email 
		FROM users 
		WHERE active = ?1 
		  AND created_at > :min_date
		  AND department = @dept
		  AND status = $status
	)
	SELECT * FROM filtered_users WHERE email LIKE ?;`

	result := Format(query, cfg)

	// Check that placeholders are preserved in CTE context
	if !containsString(result, "= ?1") || !containsString(result, "> :min_date") ||
		!containsString(result, "= @dept") || !containsString(result, "= $status") || !containsString(result, "LIKE ?") {
		t.Error("SQLite should preserve all placeholder types in CTE queries")
	}
}

func TestSQLite_Phase6_CTEComplexExample(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test complex CTE with multiple features from previous phases
	query := `-- Complex CTE example with SQLite features
	WITH RECURSIVE 
		category_tree AS (
			-- Base case: root categories
			SELECT 
				id, 
				name, 
				parent_id, 
				0 as depth,
				CAST(id AS TEXT) as path
			FROM categories 
			WHERE parent_id IS NULL
			
			UNION ALL
			
			-- Recursive case: child categories  
			SELECT 
				c.id,
				c.name,
				c.parent_id,
				ct.depth + 1,
				ct.path || '/' || CAST(c.id AS TEXT) as path
			FROM categories c
			JOIN category_tree ct ON c.parent_id = ct.id
			WHERE ct.depth < :max_depth  -- Named placeholder
		),
		product_stats AS (
			SELECT 
				category_id,
				COUNT(*) as product_count,
				AVG(price) as avg_price,
				data->'metadata'->>'supplier' as supplier_name  -- JSON operators
			FROM products 
			WHERE active = 1 
			  AND price > ?1  -- Numbered placeholder
			GROUP BY category_id, data->'metadata'->>'supplier'
		)
	SELECT 
		ct.name || ' (' || ct.depth || ')' as category_display,  -- Concatenation
		ps.product_count,
		ps.avg_price,
		ps.supplier_name
	FROM category_tree ct
	LEFT JOIN product_stats ps ON ct.id = ps.category_id
	WHERE ct.depth <= @max_display_depth  -- @ placeholder
	ORDER BY ct.path;`

	result := Format(query, cfg)

	// Verify all SQLite features work together with CTEs
	expectedElements := []string{
		"-- Complex CTE example with SQLite features",
		"WITH",
		"RECURSIVE",
		"category_tree AS",
		"product_stats AS",
		"parent_id IS NULL",
		"UNION ALL",
		"ct.depth + 1",
		"|| '/' ||",                          // Concatenation
		"< :max_depth",                       // Named placeholder
		"data -> 'metadata' ->> 'supplier'",  // JSON operators
		"> ?1",                               // Numbered placeholder
		"@max_display_depth",                 // @ placeholder (no space)
		"ct.name || ' (' || ct.depth || ')'", // Complex concatenation
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Complex CTE example should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

func TestSQLite_Phase6_BasicWindowFunctions(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test basic window functions
	query1 := `SELECT 
		employee_id,
		salary,
		ROW_NUMBER() OVER (ORDER BY salary DESC) as rank,
		RANK() OVER (PARTITION BY department ORDER BY salary DESC) as dept_rank
	FROM employees;`
	result1 := Format(query1, cfg)

	if !containsString(result1, "ROW_NUMBER() OVER") || !containsString(result1, "RANK() OVER") {
		t.Error("SQLite should format basic window functions")
	}
	if !containsString(result1, "PARTITION BY department") {
		t.Error("SQLite should format PARTITION BY clause")
	}

	// Test more window functions
	query2 := `SELECT 
		id,
		value,
		LAG(value, 1) OVER (ORDER BY date) as prev_value,
		LEAD(value, 2, 0) OVER (ORDER BY date) as next_value,
		DENSE_RANK() OVER (PARTITION BY category ORDER BY value DESC) as dense_rank
	FROM measurements;`
	result2 := Format(query2, cfg)

	if !containsString(result2, "LAG(value, 1) OVER") || !containsString(result2, "LEAD(value, 2, 0) OVER") {
		t.Error("SQLite should format LAG/LEAD window functions with parameters")
	}
	if !containsString(result2, "DENSE_RANK() OVER") {
		t.Error("SQLite should format DENSE_RANK window function")
	}
}

func TestSQLite_Phase6_WindowFrames(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test window frames with ROWS
	query1 := `SELECT 
		id,
		value,
		SUM(value) OVER (
			ORDER BY date 
			ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
		) as running_total
	FROM sales;`
	result1 := Format(query1, cfg)

	if !containsString(result1, "ROWS BETWEEN") || !containsString(result1, "UNBOUNDED PRECEDING") || !containsString(result1, "CURRENT ROW") {
		t.Error("SQLite should format ROWS frame specification")
	}

	// Test window frames with RANGE
	query2 := `SELECT 
		date,
		amount,
		AVG(amount) OVER (
			PARTITION BY category
			ORDER BY date
			RANGE BETWEEN INTERVAL '7 days' PRECEDING AND CURRENT ROW
		) as weekly_avg
	FROM transactions;`
	result2 := Format(query2, cfg)

	if !containsString(result2, "RANGE BETWEEN") && !containsString(result2, "PRECEDING AND CURRENT ROW") {
		t.Error("SQLite should format RANGE frame specification")
	}

	// Test GROUPS frame (SQLite 3.28+)
	query3 := `SELECT 
		category,
		value,
		COUNT(*) OVER (
			PARTITION BY category
			ORDER BY value
			GROUPS BETWEEN 1 PRECEDING AND 1 FOLLOWING  
		) as group_count
	FROM data;`
	result3 := Format(query3, cfg)

	if !containsString(result3, "GROUPS BETWEEN") && !containsString(result3, "1 PRECEDING AND 1 FOLLOWING") {
		t.Error("SQLite should format GROUPS frame specification")
	}
}

func TestSQLite_Phase6_WindowFunctionsWithPlaceholders(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test window functions with placeholders
	query := `SELECT 
		user_id,
		score,
		ROW_NUMBER() OVER (
			PARTITION BY category 
			ORDER BY score DESC
		) as rank
	FROM user_scores 
	WHERE category = ?1
	  AND score > :min_score
	  AND created_at > @date_filter
	  AND status = $status
	ORDER BY rank
	LIMIT ?;`

	result := Format(query, cfg)

	// Check window functions work with placeholders
	if !containsString(result, "ROW_NUMBER() OVER") {
		t.Error("SQLite window functions should work with placeholders in WHERE clause")
	}
	if !containsString(result, "= ?1") || !containsString(result, "> :min_score") ||
		!containsString(result, "> @date_filter") || !containsString(result, "= $status") || !containsString(result, "LIMIT") {
		t.Error("SQLite should preserve placeholders when using window functions")
	}
}

func TestSQLite_Phase6_CTEWithWindowFunctions(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test CTE combined with window functions
	query := `WITH monthly_sales AS (
		SELECT 
			DATE(order_date, 'start of month') as month,
			SUM(amount) as total_sales
		FROM orders
		WHERE order_date >= ?1
		GROUP BY DATE(order_date, 'start of month')
	)
	SELECT 
		month,
		total_sales,
		LAG(total_sales, 1) OVER (ORDER BY month) as prev_month_sales,
		total_sales - LAG(total_sales, 1) OVER (ORDER BY month) as sales_change,
		RANK() OVER (ORDER BY total_sales DESC) as sales_rank
	FROM monthly_sales
	ORDER BY month;`

	result := Format(query, cfg)

	// Verify CTE and window functions work together
	expectedElements := []string{
		"monthly_sales AS",
		"DATE(order_date, 'start of month')",
		"LAG(total_sales, 1) OVER",
		"ORDER BY",
		"month",
		"RANK() OVER",
		"total_sales DESC",
		">= ?1",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("CTE with window functions should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

func TestSQLite_Phase6_ComplexWindowExample(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test comprehensive example with CTEs, window functions, and all SQLite features
	query := `-- Phase 6: CTEs and Window Functions comprehensive test
	WITH RECURSIVE 
		sales_hierarchy AS (
			-- Manager hierarchy for sales teams
			SELECT 
				emp_id,
				name,
				manager_id,
				0 as level,
				CAST(emp_id AS TEXT) as path
			FROM employees 
			WHERE department = 'Sales' AND manager_id IS NULL
			
			UNION ALL
			
			SELECT 
				e.emp_id,
				e.name,
				e.manager_id, 
				sh.level + 1,
				sh.path || '/' || CAST(e.emp_id AS TEXT)
			FROM employees e
			JOIN sales_hierarchy sh ON e.manager_id = sh.emp_id
			WHERE e.department = 'Sales' AND sh.level < :max_levels
		),
		performance_data AS (
			SELECT 
				emp_id,
				quarter,
				sales_amount,
				data->'metrics'->>'bonus_eligible' as bonus_eligible,  -- JSON
				'Q' || quarter || '_' || emp_id as period_key            -- Concatenation
			FROM quarterly_sales 
			WHERE year = ?1 
			  AND sales_amount > @min_sales
		)
	SELECT 
		sh.name,
		sh.level,
		pd.quarter,
		pd.sales_amount,
		-- Window functions with various frames
		ROW_NUMBER() OVER (
			PARTITION BY sh.level, pd.quarter 
			ORDER BY pd.sales_amount DESC
		) as quarterly_rank,
		RANK() OVER (
			PARTITION BY sh.level
			ORDER BY pd.sales_amount DESC
		) as annual_rank,
		LAG(pd.sales_amount, 1) OVER (
			PARTITION BY sh.emp_id 
			ORDER BY pd.quarter
		) as prev_quarter_sales,
		SUM(pd.sales_amount) OVER (
			PARTITION BY sh.emp_id
			ORDER BY pd.quarter
			ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW  
		) as cumulative_sales,
		AVG(pd.sales_amount) OVER (
			PARTITION BY sh.level
			ORDER BY pd.quarter
			RANGE BETWEEN 1 PRECEDING AND 1 FOLLOWING
		) as level_avg_sales,
		-- Complex expression with concatenation and JSON
		CASE 
			WHEN pd.bonus_eligible = 'true' 
			THEN 'Eligible: ' || sh.name || ' (Level ' || sh.level || ')'
			ELSE 'Not eligible'
		END as bonus_status
	FROM sales_hierarchy sh
	JOIN performance_data pd ON sh.emp_id = pd.emp_id
	WHERE sh.level <= $max_display_level  -- $ placeholder
	  AND pd.sales_amount IS NOT NULL
	ORDER BY sh.level, pd.quarter, pd.sales_amount DESC;`

	result := Format(query, cfg)

	// Test comprehensive integration
	expectedElements := []string{
		"-- Phase 6: CTEs and Window Functions comprehensive test",
		"WITH",
		"RECURSIVE",
		"sales_hierarchy AS",
		"performance_data AS",
		"UNION ALL",
		"< :max_levels",                          // Named placeholder
		"data -> 'metrics' ->> 'bonus_eligible'", // JSON operators
		"'Q' || quarter || '_' || emp_id",        // Concatenation
		"= ?1",                                   // Numbered placeholder
		"> @min_sales",                           // @ placeholder
		"ROW_NUMBER() OVER",                      // Window function
		"PARTITION BY",                           // Partition clause
		"sh.level",                               // Partition columns
		"pd.quarter",                             // Partition columns
		"LAG(pd.sales_amount, 1) OVER",           // LAG function
		"ROWS BETWEEN",                           // Frame specification
		"UNBOUNDED PRECEDING",                    // Frame bound
		"RANGE BETWEEN",                          // Range frame
		"1 PRECEDING",                            // Range bound
		"'Eligible: ' || sh.name",                // Complex concatenation
		"<= $max_display_level",                  // $ placeholder
		"IS NOT NULL",                            // NULL handling
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Phase 6 comprehensive example should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

// Phase 7: DDL Essentials Tests.
func TestSQLite_Phase7_CreateTableWithGeneratedColumns(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test basic generated VIRTUAL column
	query1 := `CREATE TABLE products (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		price DECIMAL(10, 2),
		slug GENERATED ALWAYS AS (LOWER(REPLACE(name, ' ', '-'))) VIRTUAL
	);`
	result1 := Format(query1, cfg)
	if !containsString(result1, "CREATE TABLE") || !containsString(result1, "GENERATED ALWAYS AS") || !containsString(result1, "VIRTUAL") {
		t.Error("SQLite should format CREATE TABLE with VIRTUAL generated column")
	}

	// Test generated STORED column
	query2 := `CREATE TABLE inventory (
		product_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		reorder_level INTEGER DEFAULT 10,
		needs_reorder GENERATED ALWAYS AS (quantity <= reorder_level) STORED
	);`
	result2 := Format(query2, cfg)
	if !containsString(result2, "GENERATED ALWAYS AS") || !containsString(result2, "STORED") {
		t.Error("SQLite should format CREATE TABLE with STORED generated column")
	}

	// Test complex generated column expression
	query3 := `CREATE TABLE orders (
		quantity INTEGER,
		unit_price DECIMAL(10, 2),
		discount_percent DECIMAL(5, 2) DEFAULT 0.0,
		total_amount GENERATED ALWAYS AS (quantity * unit_price * (1 - discount_percent / 100.0)) STORED
	);`
	result3 := Format(query3, cfg)
	if !containsString(result3, "GENERATED ALWAYS AS") || !containsString(result3, "quantity * unit_price") {
		t.Error("SQLite should format complex generated column expressions")
	}
}

func TestSQLite_Phase7_CreateTableStrict(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test STRICT mode table
	query1 := `CREATE TABLE users STRICT (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE,
		age INTEGER CHECK(age >= 0)
	);`
	result1 := Format(query1, cfg)
	if !containsString(result1, "CREATE TABLE") || !containsString(result1, "users STRICT") {
		t.Error("SQLite should format CREATE TABLE with STRICT mode")
	}

	// Test STRICT with generated columns
	query2 := `CREATE TABLE products STRICT (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		price DECIMAL(10, 2) NOT NULL,
		slug GENERATED ALWAYS AS (LOWER(name)) VIRTUAL,
		is_expensive GENERATED ALWAYS AS (price > 100.0) STORED
	);`
	result2 := Format(query2, cfg)
	if !containsString(result2, "products STRICT") || !containsString(result2, "GENERATED ALWAYS AS") {
		t.Error("SQLite should format STRICT table with generated columns")
	}

	// Test STRICT with WITHOUT ROWID
	query3 := `CREATE TABLE sessions STRICT (
		session_id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at INTEGER NOT NULL,
		is_expired GENERATED ALWAYS AS (expires_at < unixepoch()) VIRTUAL
	) WITHOUT ROWID;`
	result3 := Format(query3, cfg)
	if !containsString(result3, "sessions STRICT") || !containsString(result3, "WITHOUT ROWID") {
		t.Error("SQLite should format STRICT table with WITHOUT ROWID")
	}
}

func TestSQLite_Phase7_CreateIndexWithIfNotExists(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test CREATE INDEX IF NOT EXISTS
	query1 := "CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);"
	result1 := Format(query1, cfg)
	if !containsString(result1, "CREATE INDEX") || !containsString(result1, "IF NOT EXISTS") {
		t.Error("SQLite should format CREATE INDEX IF NOT EXISTS")
	}

	// Test CREATE UNIQUE INDEX IF NOT EXISTS
	query2 := "CREATE UNIQUE INDEX IF NOT EXISTS idx_products_sku ON products(sku);"
	result2 := Format(query2, cfg)
	if !containsString(result2, "CREATE UNIQUE INDEX") || !containsString(result2, "IF NOT EXISTS") {
		t.Error("SQLite should format CREATE UNIQUE INDEX IF NOT EXISTS")
	}

	// Test partial index with WHERE clause
	query3 := "CREATE INDEX IF NOT EXISTS idx_active_users ON users(created_at) WHERE active = 1;"
	result3 := Format(query3, cfg)
	if !containsString(result3, "CREATE INDEX") || !containsString(result3, "IF NOT EXISTS") || !containsString(result3, "active = 1") {
		t.Error("SQLite should format partial index with WHERE clause")
	}

	// Test multi-column index
	query4 := "CREATE INDEX IF NOT EXISTS idx_orders_customer_date ON orders(customer_id, order_date DESC);"
	result4 := Format(query4, cfg)
	if !containsString(result4, "CREATE INDEX") || !containsString(result4, "(customer_id, order_date DESC)") {
		t.Error("SQLite should format multi-column index")
	}
}

func TestSQLite_Phase7_PragmaStatements(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test basic PRAGMA statements
	query1 := "PRAGMA foreign_keys = ON;"
	result1 := Format(query1, cfg)
	if !containsString(result1, "PRAGMA") || !containsString(result1, "foreign_keys = ON") {
		t.Error("SQLite should format basic PRAGMA statements")
	}

	// Test PRAGMA with string values
	query2 := "PRAGMA journal_mode = WAL;"
	result2 := Format(query2, cfg)
	if !containsString(result2, "PRAGMA") || !containsString(result2, "journal_mode = WAL") {
		t.Error("SQLite should format PRAGMA with string values")
	}

	// Test PRAGMA with function call
	query3 := "PRAGMA table_info(users);"
	result3 := Format(query3, cfg)
	if !containsString(result3, "PRAGMA") || !containsString(result3, "table_info(users)") {
		t.Error("SQLite should format PRAGMA with function call")
	}

	// Test PRAGMA with quoted table name
	query4 := "PRAGMA index_list('my_table');"
	result4 := Format(query4, cfg)
	if !containsString(result4, "PRAGMA") || !containsString(result4, "index_list('my_table')") {
		t.Error("SQLite should format PRAGMA with quoted parameters")
	}

	// Test multiple PRAGMA statements
	queries := []string{
		"PRAGMA cache_size = -64000;",
		"PRAGMA temp_store = MEMORY;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA mmap_size = 268435456;",
	}

	for _, query := range queries {
		result := Format(query, cfg)
		if !containsString(result, "PRAGMA") {
			t.Errorf("SQLite should format PRAGMA statement: %s", query)
		}
	}
}

func TestSQLite_Phase7_DDLIntegratedExample(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test comprehensive DDL example combining all Phase 7 features
	query := `-- Phase 7: DDL Essentials comprehensive test
	
-- PRAGMA setup
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

-- CREATE TABLE with generated columns and STRICT mode
CREATE TABLE user_profiles STRICT (
    user_id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    
    -- Generated columns
    full_name GENERATED ALWAYS AS (first_name || ' ' || last_name) VIRTUAL,
    email_domain GENERATED ALWAYS AS (substr(email, instr(email, '@') + 1)) STORED,
    is_recent_user GENERATED ALWAYS AS (
        julianday('now') - julianday(created_at) < 30
    ) VIRTUAL
);

-- CREATE TABLE WITHOUT ROWID with generated columns
CREATE TABLE user_sessions STRICT (
    session_token TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    last_activity INTEGER DEFAULT (unixepoch()),
    
    -- Generated columns for session management
    is_expired GENERATED ALWAYS AS (expires_at < unixepoch()) VIRTUAL,
    session_duration GENERATED ALWAYS AS (expires_at - created_at) STORED,
    is_active GENERATED ALWAYS AS (
        last_activity > unixepoch() - 3600 AND NOT is_expired
    ) VIRTUAL,
    
    -- Foreign key
    FOREIGN KEY (user_id) REFERENCES user_profiles(user_id)
) WITHOUT ROWID;

-- CREATE INDEX statements with IF NOT EXISTS
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_email 
ON user_profiles(email);

CREATE INDEX IF NOT EXISTS idx_profiles_domain 
ON user_profiles(email_domain);

CREATE INDEX IF NOT EXISTS idx_profiles_recent 
ON user_profiles(is_recent_user) 
WHERE is_recent_user = 1;

CREATE INDEX IF NOT EXISTS idx_sessions_user_active 
ON user_sessions(user_id, last_activity) 
WHERE is_active = 1;

-- More PRAGMA statements
PRAGMA table_info(user_profiles);
PRAGMA index_list('user_sessions');`

	result := Format(query, cfg)

	// Test that all Phase 7 features are preserved and formatted correctly
	expectedElements := []string{
		"-- Phase 7: DDL Essentials comprehensive test",
		"foreign_keys = ON",
		"journal_mode = WAL",
		"CREATE TABLE",
		"user_profiles STRICT",
		"GENERATED ALWAYS AS",
		"(first_name || ' ' || last_name)",
		"VIRTUAL",
		"STORED",
		"julianday('now') - julianday(created_at) < 30",
		"user_sessions STRICT",
		"WITHOUT ROWID",
		"is_expired GENERATED ALWAYS AS",
		"expires_at < unixepoch()",
		"CREATE UNIQUE INDEX",
		"IF NOT EXISTS",
		"idx_profiles_email",
		"CREATE INDEX",
		"idx_sessions_user_active",
		"is_active = 1",
		"table_info(user_profiles)",
		"index_list('user_sessions')",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Phase 7 integrated example should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

func TestSQLite_Phase7_DDLWithPreviousPhaseFeatures(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test DDL combined with features from previous phases (placeholders, JSON, etc.)
	query := `-- DDL with all SQLite features combined

-- PRAGMA with placeholder-like syntax (should be preserved as-is)
PRAGMA user_version = 1;
PRAGMA application_id = 12345;

-- CREATE TABLE with JSON column and generated JSON accessors
CREATE TABLE user_data STRICT (
    id INTEGER PRIMARY KEY,
    profile JSON NOT NULL DEFAULT '{}',
    settings JSON NOT NULL DEFAULT '{}',
    created_at TEXT DEFAULT (datetime('now')),
    
    -- Generated columns using JSON operators and concatenation
    display_name GENERATED ALWAYS AS (
        COALESCE(
            json_extract(profile, '$.display_name'),
            json_extract(profile, '$.first_name') || ' ' || json_extract(profile, '$.last_name'),
            'Anonymous User'
        )
    ) VIRTUAL,
    
    theme GENERATED ALWAYS AS (
        COALESCE(json_extract(settings, '$.theme'), 'default')
    ) STORED,
    
    is_premium GENERATED ALWAYS AS (
        json_extract(profile, '$.subscription.type') = 'premium'
    ) VIRTUAL,
    
    -- Complex generated column with CASE and JSON
    user_tier GENERATED ALWAYS AS (
        CASE 
            WHEN json_extract(profile, '$.subscription.type') = 'premium' THEN 'PREMIUM'
            WHEN json_extract(profile, '$.verified') = 1 THEN 'VERIFIED' 
            ELSE 'BASIC'
        END
    ) STORED
);

-- INSERT with UPSERT and placeholders (testing interaction with DDL)  
INSERT INTO user_data (id, profile, settings) 
VALUES (?1, :profile_json, @settings_json)
ON CONFLICT (id) DO UPDATE SET 
    profile = json_patch(profile, EXCLUDED.profile),
    settings = EXCLUDED.settings;

-- CREATE INDEX on generated columns with JSON expressions
CREATE INDEX IF NOT EXISTS idx_user_theme 
ON user_data(theme);

CREATE INDEX IF NOT EXISTS idx_user_tier_premium 
ON user_data(user_tier, is_premium) 
WHERE user_tier IN ('PREMIUM', 'VERIFIED');

-- Final PRAGMA to check the schema
PRAGMA table_xinfo(user_data);`

	result := Format(query, cfg)

	// Verify integration of all features
	expectedElements := []string{
		"-- DDL with all SQLite features combined",
		"user_version = 1",
		"application_id = 12345",
		"CREATE TABLE",
		"user_data STRICT",
		"profile JSON NOT NULL DEFAULT '{}'",
		"GENERATED ALWAYS AS",
		"json_extract(profile, '$.display_name')",
		"|| ' ' ||", // Concatenation
		"COALESCE(", // SQL function
		"VIRTUAL",
		"STORED",
		"json_extract(profile, '$.subscription.type')", // JSON operators
		"= 'premium'",
		"CASE",
		"(?1, :profile_json, @settings_json)", // All placeholder types
		"ON CONFLICT",                         // UPSERT
		"DO UPDATE",
		"json_patch(profile, EXCLUDED.profile)", // JSON function
		"CREATE INDEX",
		"IF NOT EXISTS",
		"idx_user_theme",
		"user_tier IN ('PREMIUM', 'VERIFIED')", // Complex WHERE
		"table_xinfo(user_data)",               // Schema inspection
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Phase 7 with previous features should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

func TestSQLite_Phase7_EdgeCases(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test edge cases and complex scenarios

	// Generated column with complex expression and parentheses
	query1 := `CREATE TABLE calculations (
		a REAL,
		b REAL,
		c REAL,
		result GENERATED ALWAYS AS (
			CASE 
				WHEN b * b - 4 * a * c >= 0 
				THEN (-b + sqrt(b * b - 4 * a * c)) / (2 * a)
				ELSE NULL 
			END
		) STORED
	);`
	result1 := Format(query1, cfg)
	if !containsString(result1, "GENERATED ALWAYS AS") || !containsString(result1, "sqrt(b * b - 4 * a * c)") {
		t.Error("SQLite should handle complex mathematical expressions in generated columns")
	}

	// Generated column referencing other generated columns (should be formatted but may have runtime limitations)
	query2 := `CREATE TABLE derived (
		base_value INTEGER,
		doubled GENERATED ALWAYS AS (base_value * 2) STORED,
		quadrupled GENERATED ALWAYS AS (doubled * 2) VIRTUAL
	);`
	result2 := Format(query2, cfg)
	if !containsString(result2, "doubled GENERATED ALWAYS AS") || !containsString(result2, "quadrupled GENERATED ALWAYS AS") {
		t.Error("SQLite should format generated columns that reference other columns")
	}

	// Multiple CREATE INDEX statements with various options
	query3 := `CREATE INDEX IF NOT EXISTS idx_complex 
		ON my_table(col1 COLLATE NOCASE, col2 DESC, col3 ASC) 
		WHERE col1 IS NOT NULL AND col2 > 0;`
	result3 := Format(query3, cfg)
	if !containsString(result3, "COLLATE NOCASE") || !containsString(result3, "col2 DESC") || !containsString(result3, "col1 IS NOT NULL") {
		t.Error("SQLite should handle complex CREATE INDEX with collation, ordering, and WHERE clause")
	}

	// PRAGMA statements with complex values
	query4 := `PRAGMA secure_delete = FAST;
		PRAGMA journal_size_limit = 67108864;
		PRAGMA compile_options;`
	result4 := Format(query4, cfg)
	if !containsString(result4, "secure_delete = FAST") || !containsString(result4, "journal_size_limit = 67108864") || !containsString(result4, "compile_options") {
		t.Error("SQLite should handle various PRAGMA statement formats")
	}
}

// Phase 8: Triggers & Views Tests.
func TestSQLite_Phase8_CreateTriggerBasic(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test basic BEFORE INSERT trigger
	query1 := `CREATE TRIGGER before_user_insert 
		BEFORE INSERT ON users 
		BEGIN 
			UPDATE stats SET insert_count = insert_count + 1;
		END;`
	result1 := Format(query1, cfg)
	if !containsString(result1, "CREATE TRIGGER") {
		t.Error("SQLite should format CREATE TRIGGER statement")
	}
	if !containsString(result1, "BEFORE") || !containsString(result1, "INSERT") {
		t.Error("SQLite should format BEFORE INSERT trigger type")
	}
	if !containsString(result1, "BEGIN") || !containsString(result1, "END") {
		t.Error("SQLite should format BEGIN/END trigger body blocks")
	}

	// Test AFTER UPDATE trigger
	query2 := `CREATE TRIGGER after_user_update 
		AFTER UPDATE ON users FOR EACH ROW
		BEGIN
			INSERT INTO audit_log (table_name, operation, old_id, new_id) 
			VALUES ('users', 'UPDATE', OLD.id, NEW.id);
		END;`
	result2 := Format(query2, cfg)
	if !containsString(result2, "AFTER") || !containsString(result2, "UPDATE") {
		t.Error("SQLite should format AFTER UPDATE trigger type")
	}
	if !containsString(result2, "FOR EACH ROW") {
		t.Error("SQLite should format FOR EACH ROW trigger scope")
	}
	if !containsString(result2, "OLD.id") || !containsString(result2, "NEW.id") {
		t.Error("SQLite should preserve OLD/NEW references in trigger body")
	}

	// Test BEFORE DELETE trigger
	query3 := `CREATE TRIGGER before_user_delete 
		BEFORE DELETE ON users FOR EACH ROW
		BEGIN
			INSERT INTO deleted_users SELECT * FROM users WHERE id = OLD.id;
		END;`
	result3 := Format(query3, cfg)
	if !containsString(result3, "BEFORE") || !containsString(result3, "DELETE") {
		t.Error("SQLite should format BEFORE DELETE trigger type")
	}
}

func TestSQLite_Phase8_CreateTriggerWithWhen(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test trigger with WHEN condition
	query1 := `CREATE TRIGGER update_modified_time 
		BEFORE UPDATE ON products FOR EACH ROW
		WHEN NEW.name != OLD.name OR NEW.price != OLD.price
		BEGIN
			UPDATE products SET modified_at = datetime('now') WHERE id = NEW.id;
		END;`
	result1 := Format(query1, cfg)
	if !containsString(result1, "WHEN") {
		t.Error("SQLite should format WHEN condition in triggers")
	}
	if !containsString(result1, "NEW.name != OLD.name") {
		t.Error("SQLite should format NEW/OLD comparisons in WHEN clause")
	}

	// Test trigger with complex WHEN condition
	query2 := `CREATE TRIGGER validate_user_email 
		BEFORE INSERT ON users FOR EACH ROW
		WHEN NEW.email IS NOT NULL AND NEW.email NOT LIKE '%@%'
		BEGIN
			SELECT RAISE(ABORT, 'Invalid email format');
		END;`
	result2 := Format(query2, cfg)
	if !containsString(result2, "NEW.email IS NOT NULL") {
		t.Error("SQLite should format NULL checks in WHEN clause")
	}
	if !containsString(result2, "RAISE(ABORT") {
		t.Error("SQLite should format RAISE function in trigger body")
	}
}

func TestSQLite_Phase8_CreateTriggerWithPlaceholders(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test trigger body with placeholders (though unusual, should be preserved)
	query := `CREATE TRIGGER log_changes 
		AFTER UPDATE ON sensitive_data FOR EACH ROW
		WHEN NEW.status = ?1
		BEGIN
			INSERT INTO change_log (table_name, record_id, changed_by, timestamp)
			VALUES ('sensitive_data', NEW.id, :user_id, @timestamp);
		END;`
	result := Format(query, cfg)

	// Placeholders should be preserved
	if !containsString(result, "= ?1") {
		t.Error("SQLite should preserve numbered placeholders in trigger WHEN clause")
	}
	if !containsString(result, ":user_id") {
		t.Error("SQLite should preserve :name placeholders in trigger body")
	}
	if !containsString(result, "@timestamp") {
		t.Error("SQLite should preserve @name placeholders in trigger body")
	}
}

func TestSQLite_Phase8_CreateTriggerComplexBody(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test trigger with complex body including multiple statements
	query := `CREATE TRIGGER complex_user_trigger 
		AFTER INSERT ON users FOR EACH ROW
		BEGIN
			-- Update statistics
			UPDATE user_stats SET total_users = total_users + 1;
			
			-- Create default user preferences
			INSERT INTO user_preferences (user_id, theme, notifications)
			VALUES (NEW.id, 'default', 1);
			
			-- Log the creation with JSON data
			INSERT INTO activity_log (
				user_id, 
				action, 
				details,
				metadata
			) VALUES (
				NEW.id,
				'user_created',
				'New user: ' || NEW.name || ' <' || NEW.email || '>',
				json_object('user_id', NEW.id, 'email_domain', substr(NEW.email, instr(NEW.email, '@') + 1))
			);
			
			-- Conditional logic
			UPDATE users SET welcome_sent = 1 
			WHERE id = NEW.id AND NEW.email IS NOT NULL;
		END;`
	result := Format(query, cfg)

	// Check that complex formatting is preserved
	expectedElements := []string{
		"CREATE TRIGGER",
		"complex_user_trigger",
		"AFTER",
		"INSERT",
		"FOR EACH ROW",
		"BEGIN",
		"-- Update statistics",
		"user_stats",
		"total_users = total_users + 1",
		"user_preferences",
		"NEW.id",
		"activity_log",
		"'New user: '", // Concatenation components
		"NEW.name",
		"NEW.email",
		"json_object", // JSON function
		"'user_id'",
		"substr(NEW.email", // String functions
		"instr(NEW.email, '@')",
		"NEW.email IS NOT NULL", // NULL check
		"END",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Complex trigger should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

func TestSQLite_Phase8_CreateTriggerIfNotExists(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test CREATE TRIGGER IF NOT EXISTS
	query := `CREATE TRIGGER IF NOT EXISTS audit_trigger
		AFTER UPDATE ON important_table FOR EACH ROW
		BEGIN
			INSERT INTO audit_log (table_name, record_id, change_time)
			VALUES ('important_table', NEW.id, datetime('now'));
		END;`
	result := Format(query, cfg)

	if !containsString(result, "CREATE TRIGGER") {
		t.Error("SQLite should format CREATE TRIGGER IF NOT EXISTS")
	}
	if !containsString(result, "IF NOT EXISTS") {
		t.Error("SQLite should format IF NOT EXISTS clause")
	}
}

func TestSQLite_Phase8_CreateViewBasic(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test basic CREATE VIEW
	query1 := "CREATE VIEW active_users AS SELECT id, name, email FROM users WHERE active = 1;"
	result1 := Format(query1, cfg)
	if !containsString(result1, "CREATE VIEW") {
		t.Error("SQLite should format CREATE VIEW statement")
	}
	if !containsString(result1, "active_users AS") {
		t.Error("SQLite should format view name with AS clause")
	}

	// Test CREATE VIEW with complex SELECT
	query2 := `CREATE VIEW user_summary AS
		SELECT 
			u.id,
			u.name,
			u.email,
			COUNT(o.id) as order_count,
			SUM(o.total) as total_spent
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.active = 1
		GROUP BY u.id, u.name, u.email
		ORDER BY total_spent DESC;`
	result2 := Format(query2, cfg)
	if !containsString(result2, "CREATE VIEW") || !containsString(result2, "user_summary AS") {
		t.Error("SQLite should format CREATE VIEW with complex SELECT")
	}
	if !containsString(result2, "LEFT JOIN") || !containsString(result2, "GROUP BY") {
		t.Error("SQLite should format JOIN and GROUP BY in view definition")
	}
}

func TestSQLite_Phase8_CreateViewIfNotExists(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test CREATE VIEW IF NOT EXISTS
	query := `CREATE VIEW IF NOT EXISTS recent_orders AS
		SELECT 
			o.id,
			o.user_id,
			u.name as user_name,
			o.total,
			o.created_at
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE o.created_at > datetime('now', '-30 days');`
	result := Format(query, cfg)

	if !containsString(result, "CREATE VIEW") {
		t.Error("SQLite should format CREATE VIEW IF NOT EXISTS")
	}
	if !containsString(result, "IF NOT EXISTS") {
		t.Error("SQLite should format IF NOT EXISTS clause for views")
	}
	if !containsString(result, "recent_orders AS") {
		t.Error("SQLite should format view name correctly")
	}
}

func TestSQLite_Phase8_CreateViewWithCTE(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test CREATE VIEW containing CTE
	query := `CREATE VIEW department_stats AS
		WITH employee_counts AS (
			SELECT 
				department_id,
				COUNT(*) as employee_count,
				AVG(salary) as avg_salary
			FROM employees
			WHERE active = 1
			GROUP BY department_id
		),
		department_budgets AS (
			SELECT 
				id as department_id,
				name,
				budget
			FROM departments
			WHERE budget > 0
		)
		SELECT 
			d.name as department_name,
			ec.employee_count,
			ec.avg_salary,
			d.budget,
			d.budget / ec.employee_count as budget_per_employee
		FROM department_budgets d
		JOIN employee_counts ec ON d.department_id = ec.department_id;`
	result := Format(query, cfg)

	// Check that CTE and view work together
	expectedElements := []string{
		"CREATE VIEW",
		"department_stats AS",
		"WITH",
		"employee_counts AS",
		"department_budgets AS",
		"AVG(salary)",
		"GROUP BY",
		"department_id",
		"d.budget / ec.employee_count",
		"JOIN employee_counts ec",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("CREATE VIEW with CTE should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

func TestSQLite_Phase8_CreateViewWithWindowFunctions(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test CREATE VIEW with window functions
	query := `CREATE VIEW user_rankings AS
		SELECT 
			id,
			name,
			score,
			ROW_NUMBER() OVER (ORDER BY score DESC) as rank,
			RANK() OVER (PARTITION BY category ORDER BY score DESC) as category_rank,
			LAG(score, 1) OVER (ORDER BY score DESC) as prev_score
		FROM user_scores
		WHERE active = 1;`
	result := Format(query, cfg)

	if !containsString(result, "CREATE VIEW") || !containsString(result, "user_rankings AS") {
		t.Error("SQLite should format CREATE VIEW with window functions")
	}
	if !containsString(result, "ROW_NUMBER() OVER") || !containsString(result, "PARTITION BY category") {
		t.Error("SQLite should format window functions in view definition")
	}
	if !containsString(result, "LAG(score, 1) OVER") {
		t.Error("SQLite should format LAG window function in view")
	}
}

func TestSQLite_Phase8_IntegratedExample(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test comprehensive example combining triggers and views with all SQLite features
	query := `-- Phase 8: Triggers & Views comprehensive test

-- Create a view with CTE, JSON operators, and window functions
CREATE VIEW IF NOT EXISTS user_activity_summary AS
	WITH recent_activities AS (
		SELECT 
			user_id,
			action_type,
			json_extract(metadata, '$.source') as source,
			json_extract(metadata, '$.ip_address') as ip_address,
			created_at
		FROM activity_log
		WHERE created_at > datetime('now', '-7 days')
	),
	activity_stats AS (
		SELECT 
			user_id,
			COUNT(*) as total_actions,
			COUNT(DISTINCT source) as unique_sources,
			MAX(created_at) as last_activity,
			ROW_NUMBER() OVER (ORDER BY COUNT(*) DESC) as activity_rank
		FROM recent_activities
		GROUP BY user_id
	)
	SELECT 
		u.id,
		u.name,
		u.email,
		as_stats.total_actions,
		as_stats.unique_sources,
		as_stats.last_activity,
		as_stats.activity_rank,
		CASE 
			WHEN as_stats.total_actions > 100 THEN 'Very Active'
			WHEN as_stats.total_actions > 50 THEN 'Active'
			ELSE 'Low Activity'
		END as activity_level
	FROM users u
	LEFT JOIN activity_stats as_stats ON u.id = as_stats.user_id;

-- Create a trigger that logs changes with JSON metadata and uses placeholders  
CREATE TRIGGER IF NOT EXISTS user_change_logger
	AFTER UPDATE ON users FOR EACH ROW
	WHEN NEW.email != OLD.email OR NEW.name != OLD.name
	BEGIN
		-- Log the change with detailed JSON metadata
		INSERT INTO change_audit (
			table_name,
			record_id,
			change_type,
			old_values,
			new_values,
			metadata,
			changed_at
		) VALUES (
			'users',
			NEW.id,
			'UPDATE',
			json_object(
				'email', OLD.email,
				'name', OLD.name,
				'updated_at', OLD.updated_at
			),
			json_object(
				'email', NEW.email,
				'name', NEW.name,  
				'updated_at', NEW.updated_at
			),
			json_object(
				'email_changed', NEW.email != OLD.email,
				'name_changed', NEW.name != OLD.name,
				'change_summary', 
				CASE 
					WHEN NEW.email != OLD.email AND NEW.name != OLD.name 
					THEN 'Both email and name changed'
					WHEN NEW.email != OLD.email 
					THEN 'Email changed from ' || OLD.email || ' to ' || NEW.email
					ELSE 'Name changed from ' || OLD.name || ' to ' || NEW.name
				END
			),
			datetime('now')
		);
		
		-- Update activity summary
		INSERT OR REPLACE INTO user_activity_summary_cache (
			user_id, 
			last_change_type, 
			change_count
		) VALUES (
			NEW.id,
			'profile_update',
			COALESCE((SELECT change_count FROM user_activity_summary_cache WHERE user_id = NEW.id), 0) + 1
		);
	END;`

	result := Format(query, cfg)

	// Test comprehensive integration of all features
	expectedElements := []string{
		"-- Phase 8: Triggers & Views comprehensive test",
		"CREATE VIEW",
		"IF NOT EXISTS",
		"user_activity_summary",
		"WITH",
		"recent_activities",
		"activity_stats",
		"json_extract", // JSON operators
		"'$.source'",
		"datetime", // Date functions
		"'-7 days'",
		"ROW_NUMBER", // Window functions
		"OVER",
		"COUNT(*)", // Window ordering
		"DESC",
		"GROUP BY", // Aggregation
		"user_id",
		"LEFT JOIN", // JOINs
		"CREATE TRIGGER",
		"user_change_logger",
		"AFTER",
		"UPDATE",
		"FOR EACH ROW",
		"WHEN",
		"NEW.email", // NEW/OLD references
		"OLD.email",
		"BEGIN",
		"change_audit",
		"json_object", // JSON functions
		"'email'",     // JSON construction
		"OLD.email",
		"NEW.email",                     // Boolean expressions
		"'Both email and name changed'", // String literals
		"OLD.email",                     // Concatenation components
		"NEW.email",
		"datetime('now')",   // Date functions
		"INSERT OR REPLACE", // UPSERT
		"COALESCE",          // NULL handling
		"END",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Phase 8 integrated example should contain: %s\nFull result:\n%s", element, result)
		}
	}
}

// Phase 11: Final Polish & Edge Cases Tests.
func TestSQLite_Phase11_PragmaValuePreservation(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test PRAGMA values are preserved as-is, not uppercased
	queries := []struct {
		input    string
		expected string
	}{
		{
			"PRAGMA journal_mode = WAL;",
			"WAL", // Should preserve case, not become "wal"
		},
		{
			"PRAGMA encoding = 'UTF-8';",
			"'UTF-8'", // String values should be preserved
		},
		{
			"PRAGMA table_info(\"users\");",
			"table_info(\"users\")", // Function calls preserved
		},
		{
			"PRAGMA cache_size = -64000;",
			"-64000", // Numeric values preserved
		},
		{
			"PRAGMA temp_store = MEMORY;",
			"MEMORY", // Identifier values preserved
		},
	}

	for _, test := range queries {
		result := Format(test.input, cfg)
		if !containsString(result, test.expected) {
			t.Errorf("PRAGMA value preservation failed.\nInput: %s\nExpected to contain: %s\nGot: %s",
				test.input, test.expected, result)
		}
	}
}

func TestSQLite_Phase11_PragmaValuePreservationWithUppercase(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  ", Uppercase: true}

	// Even with uppercase flag, PRAGMA values should NOT be uppercased
	query := "PRAGMA journal_mode = WAL; PRAGMA encoding = 'UTF-8';"
	result := Format(query, cfg)

	// Keywords should be uppercase, but values should be preserved
	if !containsString(result, "PRAGMA") {
		t.Error("PRAGMA keyword should be uppercase")
	}
	if !containsString(result, "WAL") || containsString(result, "wal") {
		t.Error("PRAGMA value 'WAL' should preserve case even with uppercase flag")
	}
	if !containsString(result, "'UTF-8'") {
		t.Error("PRAGMA string value should preserve case")
	}
}

func TestSQLite_Phase11_SemicolonsInStringsAndComments(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test that semicolons inside strings and comments don't split statements
	query := `-- Comment with ; semicolon
	SELECT 'test; string' AS col1, /* inline; comment */ 'another;value' FROM users;
	SELECT * FROM posts;`

	result := Format(query, cfg)

	// Should result in exactly 2 SELECT statements, not more due to embedded semicolons
	selectCount := 0
	for i := range len(result) {
		if i+6 < len(result) && result[i:i+6] == "SELECT" {
			selectCount++
		}
	}

	if selectCount != 2 {
		t.Errorf("Expected exactly 2 SELECT statements, found %d. Result:\n%s", selectCount, result)
	}

	// Verify semicolons inside strings are preserved
	if !containsString(result, "'test; string'") {
		t.Error("Semicolon inside string should be preserved")
	}
	if !containsString(result, "'another;value'") {
		t.Error("Semicolon inside string should be preserved")
	}
}

func TestSQLite_Phase11_UnicodeIdentifierPreservation(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test Unicode identifiers are preserved, not altered by case conversion
	query := `SELECT "ÄÖÜ" AS "αβγ", ` + "`カタカナ`" + `, [中文字段], "Ñoël" FROM [unicode_table];`
	result := Format(query, cfg)

	expectedUnicodeElements := []string{
		"\"ÄÖÜ\"",         // German umlauts
		"\"αβγ\"",         // Greek letters
		"`カタカナ`",          // Japanese katakana
		"[中文字段]",          // Chinese characters
		"\"Ñoël\"",        // Accented characters
		"[unicode_table]", // Bracketed table name
	}

	for _, element := range expectedUnicodeElements {
		if !containsString(result, element) {
			t.Errorf("Unicode identifier should be preserved: %s\nFull result:\n%s", element, result)
		}
	}
}

func TestSQLite_Phase11_UnicodeWithUppercaseFlag(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  ", Uppercase: true}

	// Unicode identifiers should NOT be affected by uppercase flag
	query := `SELECT "ÄÖÜ" AS "αβγ", ` + "`カタカナ`" + ` FROM [unicode_table];`
	result := Format(query, cfg)

	// Keywords should be uppercase
	if !containsString(result, "SELECT") || !containsString(result, "FROM") {
		t.Error("Keywords should be uppercase")
	}

	// Unicode identifiers should preserve their exact case
	if !containsString(result, "\"ÄÖÜ\"") || !containsString(result, "\"αβγ\"") {
		t.Error("Unicode identifiers should preserve case even with uppercase flag")
	}
}

func TestSQLite_Phase11_LargeComplexQueries(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test very large, deeply nested query
	query := `WITH RECURSIVE 
		fibonacci(n, a, b) AS (
			SELECT 1, 0, 1
			UNION ALL
			SELECT n+1, b, a+b FROM fibonacci WHERE n < 100
		),
		stats AS (
			SELECT 
				COUNT(*) as total,
				MAX(b) as max_fib,
				MIN(b) as min_fib,
				AVG(CAST(b AS REAL)) as avg_fib
			FROM fibonacci
		),
		categories AS (
			SELECT 
				n,
				CASE
					WHEN b % 15 = 0 THEN 'FizzBuzz'
					WHEN b % 3 = 0 THEN 'Fizz'
					WHEN b % 5 = 0 THEN 'Buzz'
					WHEN b % 2 = 0 THEN 'Even'
					ELSE 'Odd'
				END as category,
				json_object(
					'number', n,
					'fib_value', b,
					'is_prime', (
						SELECT COUNT(*) = 0 
						FROM (
							SELECT 1 
							FROM generate_series(2, CAST(sqrt(b) AS INTEGER)) AS divisor
							WHERE b % divisor = 0
						)
					)
				) as metadata
			FROM fibonacci
		)
		SELECT 
			c.n,
			c.category,
			c.metadata ->> 'is_prime' as is_prime,
			s.total,
			ROW_NUMBER() OVER (
				PARTITION BY c.category 
				ORDER BY c.n DESC
			) as category_rank,
			LAG(c.n, 1) OVER (ORDER BY c.n) as prev_n
		FROM categories c
		CROSS JOIN stats s
		WHERE c.n <= 50
		ORDER BY c.n;`

	result := Format(query, cfg)

	// Should handle the complexity without errors and preserve key elements
	expectedElements := []string{
		"WITH",
		"RECURSIVE",
		"fibonacci",
		"UNION ALL",
		"json_object",
		"ROW_NUMBER() OVER",
		"PARTITION BY",
		"CROSS JOIN",
		"ORDER BY",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Large query should contain: %s", element)
		}
	}

	// Verify proper indentation is maintained
	if !containsString(result, "    SELECT") { // Should have nested indentation
		t.Error("Large query should maintain proper nested indentation")
	}
}

func TestSQLite_Phase11_MalformedInputHandling(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test that formatter doesn't panic on malformed input
	malformedQueries := []string{
		"SELECT * FROM (SELECT 1",         // Missing closing paren
		"SELECT 'unterminated string",     // Unclosed string
		"CREATE TABLE users (id INTEGER,", // Incomplete DDL
		"SELECT * FROM",                   // Incomplete FROM clause
		"PRAGMA",                          // Incomplete PRAGMA
		"INSERT INTO users VALUES (",      // Incomplete VALUES
		"SELECT /* unclosed comment",      // Unclosed comment
	}

	for _, query := range malformedQueries {
		// Should not panic - capture any panics
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Formatter panicked on malformed input: %s\nPanic: %v", query, r)
				}
			}()
			result := Format(query, cfg)
			// Result should contain at least some of the input, even if malformed
			if len(result) == 0 {
				t.Errorf("Formatter returned empty result for: %s", query)
			}
		}()
	}
}

func TestSQLite_Phase11_EdgeCaseParameters(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Test edge cases with parameters and special characters
	query := `SELECT 
		?1 as param1,
		:named_param as param2,
		@at_param as param3,
		$dollar_param as param4,
		'?' as not_param,
		"?" as not_param2,
		-- Comment with ? and :name and @param
		/* Block comment with $var */ 
		json_extract(data, '$.path?key') as json_path
	FROM test_table
	WHERE column = ?2 AND other = :test;`

	result := Format(query, cfg)

	// Parameters should be preserved
	parameterElements := []string{
		"?1", "?2",
		":named_param", ":test",
		"@at_param",
		"$dollar_param",
	}

	for _, param := range parameterElements {
		if !containsString(result, param) {
			t.Errorf("Parameter should be preserved: %s", param)
		}
	}

	// Non-parameters should be preserved as strings
	if !containsString(result, "'?'") || !containsString(result, "\"?\"") {
		t.Error("Question marks in strings should not be treated as parameters")
	}
}

func TestSQLite_Phase11_IntegratedScenario(t *testing.T) {
	cfg := &Config{Language: SQLite, Indent: "  "}

	// Comprehensive test combining all Phase 11 requirements
	query := `-- Phase 11 integration test with Unicode: Тест системы
	PRAGMA foreign_keys = ON;
	PRAGMA encoding = 'UTF-8'; 
	
	CREATE TABLE "测试表" (
		"用户ID" INTEGER PRIMARY KEY,
		"姓名" TEXT,
		"邮箱" TEXT UNIQUE,
		metadata JSON DEFAULT '{}',
		created_at DATETIME DEFAULT (datetime('now'))
	) STRICT;
	
	-- Test with semicolons in strings and Unicode
	INSERT INTO "测试表" ("姓名", "邮箱", metadata) VALUES 
		('张三; 测试用户', 'zhang@test; fake.com', json_object('notes', 'Test; with semicolons')),
		('李四', 'li@example.org', json_object('αβγ', 'Greek letters', 'カタカナ', 'Japanese')),
		('Müller', 'mueller@täst.de', json_object('Ñoël', 'Accented names'));
	
	WITH user_stats AS (
		SELECT 
			COUNT(*) as total_users,
			COUNT(*) FILTER (WHERE "邮箱" LIKE '%@%.com') as com_users,
			json_group_array(json_object(
				'name', "姓名",
				'email', "邮箱", 
				'meta', json_extract(metadata, '$')
			)) as all_users
		FROM "测试表"
		WHERE "姓名" IS NOT NULL
	)
	SELECT 
		us.total_users,
		us.com_users,
		json_extract(us.all_users, '$[0].name') as first_user,
		?1 as param_test,
		:user_filter as named_param,
		@limit as at_param
	FROM user_stats us
	WHERE us.total_users > 0
	ORDER BY us.total_users DESC;
	
	PRAGMA table_xinfo("测试表");`

	result := Format(query, cfg)

	// Verify all key Phase 11 elements are handled correctly
	phase11Elements := []string{
		// PRAGMA values preserved
		"PRAGMA", "foreign_keys = ON", "encoding = 'UTF-8'", "table_xinfo",
		// Unicode identifiers preserved
		"\"测试表\"", "\"用户ID\"", "\"姓名\"", "\"邮箱\"",
		// Semicolons in strings preserved (not splitting statements)
		"'zhang@test; fake.com'", "'Test; with semicolons'",
		// Unicode in strings preserved
		"'张三; 测试用户'", "'αβγ'", "'カタカナ'",
		// Parameters preserved
		"?1", ":user_filter", "@limit",
		// Complex JSON and CTEs formatted
		"json_object", "json_extract", "WITH", "user_stats",
		// Comments with Unicode preserved
		"-- Phase 11 integration test with Unicode: Тест системы",
	}

	for _, element := range phase11Elements {
		if !containsString(result, element) {
			t.Errorf("Phase 11 integrated test should contain: %s\nFull result:\n%s", element, result)
		}
	}

	// Verify structure is maintained
	if !containsString(result, "CREATE TABLE") || !containsString(result, "INSERT INTO") {
		t.Error("DDL structure should be preserved")
	}
	if !containsString(result, "WITH") || !containsString(result, "SELECT") {
		t.Error("CTE and SELECT structure should be preserved")
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
