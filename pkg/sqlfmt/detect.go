package sqlfmt

import (
	"path/filepath"
	"regexp"
	"strings"
)

// DetectDialect attempts to automatically detect the SQL dialect from a file path and/or content.
// It first checks file extensions, then falls back to content-based heuristics.
// Returns the detected Language and a boolean indicating if detection was successful.
func DetectDialect(filePath string, content string) (Language, bool) {
	// First, try file extension detection
	if lang, ok := detectFromFileExtension(filePath); ok {
		return lang, true
	}

	// Fall back to content-based detection
	if lang, ok := detectFromContent(content); ok {
		return lang, true
	}

	// Default to standard SQL if no detection succeeds
	return StandardSQL, false
}

// detectFromFileExtension detects dialect based on file extension.
func detectFromFileExtension(filePath string) (Language, bool) {
	base := strings.ToLower(filepath.Base(filePath))

	// Check for compound extensions first (more specific)
	if strings.HasSuffix(base, ".mysql.sql") || strings.HasSuffix(base, ".my.sql") {
		return MySQL, true
	}
	if strings.HasSuffix(base, ".psql.sql") || strings.HasSuffix(base, ".pgsql.sql") {
		return PostgreSQL, true
	}
	if strings.HasSuffix(base, ".sqlite.sql") || strings.HasSuffix(base, ".db.sql") {
		return SQLite, true
	}
	if strings.HasSuffix(base, ".plsql.sql") || strings.HasSuffix(base, ".ora.sql") {
		return PLSQL, true
	}

	// Check for simple extensions
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".psql", ".pgsql":
		return PostgreSQL, true
	case ".mysql":
		return MySQL, true
	case ".sqlite":
		return SQLite, true
	case ".plsql":
		return PLSQL, true
	}

	// Check for embedded extensions in filename
	if strings.Contains(base, ".mysql.") || strings.HasSuffix(base, ".mysql") {
		return MySQL, true
	}
	if strings.Contains(base, ".psql.") || strings.HasSuffix(base, ".psql") {
		return PostgreSQL, true
	}
	if strings.Contains(base, ".sqlite.") || strings.HasSuffix(base, ".sqlite") {
		return SQLite, true
	}
	if strings.Contains(base, ".plsql.") || strings.HasSuffix(base, ".plsql") ||
		strings.Contains(base, ".ora.") || strings.HasSuffix(base, ".ora") {
		return PLSQL, true
	}

	return StandardSQL, false
}

// detectFromContent detects dialect based on SQL content heuristics.
func detectFromContent(content string) (Language, bool) {
	content = strings.ToLower(content)

	// PostgreSQL indicators (highest priority - most specific)
	if hasPostgreSQLIndicators(content) {
		return PostgreSQL, true
	}

	// PL/SQL indicators (check before SQLite since some patterns overlap)
	if hasPLSQLIndicators(content) {
		return PLSQL, true
	}

	// MySQL indicators
	if hasMySQLIndicators(content) {
		return MySQL, true
	}

	// SQLite indicators (lowest priority)
	if hasSQLiteIndicators(content) {
		return SQLite, true
	}

	return StandardSQL, false
}

// hasPostgreSQLIndicators checks for PostgreSQL-specific syntax patterns.
func hasPostgreSQLIndicators(content string) bool {
	// PostgreSQL-specific operators and syntax
	postgresPatterns := []string{
		`::[a-zA-Z_][a-zA-Z0-9_]*`, // Type casting with ::
		`\$\$`,                     // Dollar quoting
		`\$[0-9]+`,                 // Positional parameters ($1, $2, etc.)
		`\breturning\b`,            // RETURNING clause
		`->>`,                      // JSON operators
		`#>`,                       // JSON path operators
		`@>`,                       // JSON containment
		`<@`,                       // JSON containment (reverse)
		`\bjsonb?\b`,               // JSON/JSONB types
		`\bserial\b`,               // PostgreSQL serial types
		`\bbigserial\b`,            // PostgreSQL bigserial
		`\bregclass\b`,             // PostgreSQL reg* types
		`\bregtype\b`,
		`\bregproc\b`,
		`\bregnamespace\b`,
		`\btsvector\b`, // Full-text search types
		`\btsquery\b`,
		`\bint4range\b`, // Range types
		`\bint8range\b`,
		`\bnumrange\b`,
		`\btsrange\b`,
		`\btstzrange\b`,
		`\bdaterange\b`,
		`\bgenerate_series\b`, // PostgreSQL functions
		`\bunnest\b`,
		`\barray_agg\b`,
		`\bstring_agg\b`,
		`\blateral\b`, // LATERAL joins
	}

	return matchesAnyPattern(content, postgresPatterns)
}

// hasMySQLIndicators checks for MySQL-specific syntax patterns.
func hasMySQLIndicators(content string) bool {
	// MySQL-specific syntax
	mysqlPatterns := []string{
		"`[^`]+`",                     // Backtick-quoted identifiers
		`\bon duplicate key update\b`, // ON DUPLICATE KEY UPDATE
		`\binsert ignore\b`,           // INSERT IGNORE
		`\breplace into\b`,            // REPLACE INTO
		`\blower\([^)]+\)`,            // LOWER() function (MySQL is case-insensitive)
		`\bupper\([^)]+\)`,            // UPPER() function
		`\bgroup_concat\b`,            // GROUP_CONCAT function
		`\bfound_rows\b`,              // FOUND_ROWS()
		`\brow_count\b`,               // ROW_COUNT()
		`\blast_insert_id\b`,          // LAST_INSERT_ID()
		`\bauto_increment\b`,          // AUTO_INCREMENT
		`\bengine\s*=\s*[a-zA-Z_]+`,   // ENGINE=InnoDB
		`\bcharset\s*=\s*[a-zA-Z_]+`,  // CHARSET=utf8
		`\bcollate\s*=\s*[a-zA-Z_]+`,  // COLLATE=utf8_general_ci
		`\bstraight_join\b`,           // STRAIGHT_JOIN
		`\bforce index\b`,             // FORCE INDEX
		`\buse index\b`,               // USE INDEX
		`\bignore index\b`,            // IGNORE INDEX
		`\block in share mode\b`,      // LOCK IN SHARE MODE
		`\bfor update\b`,              // FOR UPDATE (with MySQL-specific context)
		`\bjson_extract\b`,            // MySQL JSON functions
		`\bjson_unquote\b`,
		`\bjson_type\b`,
	}

	return matchesAnyPattern(content, mysqlPatterns)
}

// hasSQLiteIndicators checks for SQLite-specific syntax patterns.
func hasSQLiteIndicators(content string) bool {
	// SQLite-specific syntax
	sqlitePatterns := []string{
		`\bpragma\s+[a-zA-Z_][a-zA-Z0-9_]*\b`, // PRAGMA with specific SQLite pragmas
		`\bwithout rowid\b`,                   // WITHOUT ROWID tables
		`\bautoincrement\b`,                   // AUTOINCREMENT (SQLite spelling)
		`\battach\b.*\bdatabase\b`,            // ATTACH DATABASE (more specific)
		`\bdetach\b.*\bdatabase\b`,            // DETACH DATABASE
		`\breindex\b`,                         // REINDEX
		`\bvacuum\b`,                          // VACUUM
		`\banalyze\b`,                         // ANALYZE
		`\bexplain query plan\b`,              // EXPLAIN QUERY PLAN
		`\browid\b`,                           // ROWID
		`\b_oid\b`,                            // _ROWID_
		`\blic\b`,                             // SQLite license pragma
		`\bforeign_keys\b`,                    // Foreign key pragmas
		`\bjournal_mode\b`,
		`\bsynchronous\b`,
		`\bcache_size\b`,
		`\btemp_store\b`,
		`\btable_info\b`, // SQLite-specific pragma functions
		`\bindex_info\b`,
		`\bindex_list\b`,
		`\bdatabase_list\b`,
		`\bforeign_key_list\b`,
		`\bcollation_list\b`,
		`\bfunction_list\b`,
		`\bmodule_list\b`,
		`\bpragma_list\b`,
		`\bstatistics\b`,
		`\bcompile_options\b`,
	}

	return matchesAnyPattern(content, sqlitePatterns)
}

// hasPLSQLIndicators checks for PL/SQL-specific syntax patterns.
func hasPLSQLIndicators(content string) bool {
	// PL/SQL-specific syntax - made more specific to avoid SQLite conflicts
	plsqlPatterns := []string{
		`\bbegin\b[^;]*\bexception\b`, // BEGIN with EXCEPTION block
		`\bbegin\b.*?\bend\b\s*;`,     // BEGIN ... END; with semicolon (more flexible)
		`\bexception\b.*\bwhen\b`,     // EXCEPTION WHEN blocks
		`\bcreate\b.*\bprocedure\b`,   // CREATE PROCEDURE
		`\bcreate\b.*\bfunction\b`,    // CREATE FUNCTION
		`\bcreate\b.*\bpackage\b`,     // CREATE PACKAGE
		`\bcreate\b.*\btrigger\b`,     // CREATE TRIGGER
		`\bpackage\b.*\bbody\b`,       // PACKAGE BODY
		`\braise\b`,                   // RAISE statements
		`\bexecute immediate\b`,       // Dynamic SQL
		`\bopen\b.*\bfor\b`,           // Cursor FOR loops
		`\bfetch\b.*\binto\b`,         // FETCH INTO
		`\bclose\b`,                   // CLOSE cursor
		`\bref cursor\b`,              // REF CURSOR
		`\bsys_refcursor\b`,           // SYS_REFCURSOR
		`\bbulk collect\b`,            // BULK COLLECT
		`\bforall\b`,                  // FORALL loops
		`\btype\b.*\bis\b`,            // TYPE definitions
		`\brecord\b`,                  // RECORD types
		`\bvarray\b`,                  // VARRAY types
		`\bnested table\b`,            // NESTED TABLE types
		`\bindex by\b`,                // INDEX BY tables
		`\bpls_integer\b`,             // PLS_INTEGER type
		`\bbinary_integer\b`,          // BINARY_INTEGER type
		`\bnatural\b`,                 // NATURAL types
		`\bpositive\b`,                // POSITIVE types
		`\bsimple_integer\b`,          // SIMPLE_INTEGER type
		`\bboolean\b`,                 // BOOLEAN type (though not SQL standard)
		`\btrue\b`,                    // TRUE literal
		`\bfalse\b`,                   // FALSE literal
	}

	return matchesAnyPattern(content, plsqlPatterns)
}

// matchesAnyPattern checks if the content matches any of the given regex patterns.
func matchesAnyPattern(content string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}
	return false
}
