// Package dialects provides PostgreSQL-specific SQL formatting functionality.
// This package implements comprehensive PostgreSQL support including dollar-quoted strings,
// type casting operators, JSON/JSONB operations, CTEs, window functions, and PL/pgSQL constructs.
package dialects

import (
	"strings"

	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/core"
	"github.com/MeKo-Christian/go-sqlfmt/pkg/sqlfmt/types"
)

var (
	// postgreSQLReservedWords extends standard SQL with PostgreSQL-specific keywords.
	// Includes pattern matching (ILIKE, SIMILAR TO), JSON operators, window functions,
	// procedural constructs, and DDL enhancements.
	postgreSQLReservedWords = append(standardSQLReservedWords, []string{
		"ILIKE", "SIMILAR TO", "ON CONFLICT", "DO UPDATE", "DO NOTHING",
		// Window function keywords
		"WINDOW", "OVER", "PARTITION BY", "FILTER", "RANGE", "ROWS", "GROUPS",
		"UNBOUNDED", "PRECEDING", "FOLLOWING", "CURRENT ROW",
		"EXCLUDE", "TIES", "NO OTHERS",
		// NULLS ordering
		"NULLS FIRST", "NULLS LAST",
		// LATERAL joins
		"LATERAL",
		// Array functions
		"ARRAY", "UNNEST",
		// Procedural keywords
		"LANGUAGE", "RETURNS", "AS", "DECLARE", "BEGIN",
		// Function modifiers
		"IMMUTABLE", "STABLE", "VOLATILE", "STRICT", "CALLED ON NULL INPUT",
		"SECURITY DEFINER", "SECURITY INVOKER", "LEAKPROOF", "NOT LEAKPROOF",
		// Return types
		"SETOF", "TABLE", "TRIGGER", "VOID",
		// Function options
		"COST", "ROWS", "SUPPORT", "PARALLEL SAFE", "PARALLEL UNSAFE", "PARALLEL RESTRICTED",
		// DDL and Index keywords
		"CONCURRENTLY", "IF NOT EXISTS", "IF EXISTS",
		// Index methods
		"BTREE", "HASH", "GIN", "GIST", "SPGIST", "BRIN",
		// Index options
		"INCLUDE", "TABLESPACE", "WITH", "FILLFACTOR", "FASTUPDATE",
		// Additional DDL keywords
		"REINDEX", "CLUSTER", "VACUUM", "ANALYZE",
	}...)

	// postgreSQLReservedTopLevelWords adds PostgreSQL-specific top-level keywords that start new statement sections.
	// These keywords cause new lines and reset indentation level.
	// Note: "DO" is included as top-level but is context-aware via tokenOverride to handle UPSERT correctly.
	postgreSQLReservedTopLevelWords = append(standardSQLReservedTopLevelWords, []string{
		"WITH", "WITH RECURSIVE", "RETURNING", "WINDOW",
		// Procedural blocks and functions
		"DO", "CREATE FUNCTION", "CREATE OR REPLACE FUNCTION",
		// DDL operations
		"CREATE INDEX", "CREATE UNIQUE INDEX", "DROP INDEX", "REINDEX",
	}...)

	// postgreSQLReservedTopLevelWordsNoIndent inherits from standard SQL - no PostgreSQL-specific additions needed.
	postgreSQLReservedTopLevelWordsNoIndent = standardSQLReservedTopLevelWordsNoIndent

	// postgreSQLReservedNewlineWords adds LATERAL join support and UPSERT keywords that trigger new lines.
	postgreSQLReservedNewlineWords = append(standardSQLReservedNewlineWords, []string{
		"LATERAL JOIN", "LEFT LATERAL JOIN", "RIGHT LATERAL JOIN", "CROSS JOIN LATERAL",
		// UPSERT clause keywords
		"ON CONFLICT",
	}...)
)

// PostgreSQLFormatter implements PostgreSQL-specific SQL formatting.
// It supports all PostgreSQL language features including dollar-quoted strings,
// type casting, JSON operations, CTEs, window functions, and PL/pgSQL constructs.
type PostgreSQLFormatter struct {
	cfg *Config
}

// NewPostgreSQLFormatter creates a new PostgreSQL formatter with dialect-specific configuration.
// The formatter automatically configures support for:
//   - Dollar-quoted strings ($$...$$, $tag$...$tag$)
//   - PostgreSQL operators (::, ->, ->>, #>, #>>, @>, <@, ?, ?|, ?&, ~, !~, ~*, !~*)
//   - Numbered placeholders ($1, $2, $3...)
//   - PostgreSQL-specific keywords and formatting rules
//
// Example usage:
//
//	cfg := &Config{Indent: "  ", Language: core.PostgreSQL}
//	formatter := NewPostgreSQLFormatter(cfg)
//	result := formatter.Format("SELECT data->>'name' FROM users WHERE id = $1")
func NewPostgreSQLFormatter(cfg *Config) *PostgreSQLFormatter {
	cfg.TokenizerConfig = NewPostgreSQLTokenizerConfig()
	return &PostgreSQLFormatter{cfg: cfg}
}

// NewPostgreSQLTokenizerConfig creates a tokenizer configuration for PostgreSQL dialect.
// Configures support for all PostgreSQL-specific syntax elements:
//   - Dollar-quoted strings: $$ and $tag$ varieties
//   - Numbered placeholders: $1, $2, $3... (1-based indexing)
//   - Named placeholders: @param, :param
//   - PostgreSQL line comments: --
//   - Standard string types with PostgreSQL extensions
//
// The tokenizer handles PostgreSQL operators and keywords through the reserved word lists
// and provides proper recognition of PostgreSQL-specific constructs like ILIKE, SIMILAR TO,
// JSON operators, window functions, and procedural language elements.
func NewPostgreSQLTokenizerConfig() *TokenizerConfig {
	return &TokenizerConfig{
		ReservedWords:                 postgreSQLReservedWords,
		ReservedTopLevelWords:         postgreSQLReservedTopLevelWords,
		ReservedNewlineWords:          postgreSQLReservedNewlineWords,
		ReservedTopLevelWordsNoIndent: postgreSQLReservedTopLevelWordsNoIndent,
		StringTypes:                   []string{`""`, "N''", "''", "``", "$$"},
		OpenParens:                    []string{"(", "CASE"},
		CloseParens:                   []string{")", "END"},
		IndexedPlaceholderTypes:       []string{"$"},
		NamedPlaceholderTypes:         []string{"@", ":"},
		LineCommentTypes:              []string{"--"},
	}
}

// Format formats a PostgreSQL query string according to PostgreSQL formatting conventions.
// Handles all PostgreSQL-specific syntax including type casts, JSON operations, and procedural constructs.
func (psf *PostgreSQLFormatter) Format(query string) string {
	return core.FormatQuery(
		psf.cfg,
		psf.tokenOverride,
		query,
	)
}

// tokenOverride handles PostgreSQL-specific token formatting overrides.
// Implements context-aware formatting for:
//
// 1. Type cast operator (::) - formatted without spaces following PostgreSQL conventions
// 2. DO keyword - context-aware handling for UPSERT vs standalone DO blocks
//
// Examples:
//
//	Type casting:
//	  'text'::varchar  -> 'text'::varchar  (no spaces around ::)
//	  value::numeric   -> value::numeric   (no spaces around ::)
//
//	DO keyword context awareness:
//	  ON CONFLICT (id) DO UPDATE SET ...  -> DO stays inline (UPSERT context)
//	  DO $$ BEGIN ... END $$;             -> DO creates newline (standalone block)
//
//nolint:cyclop // Context-aware logic requires checking multiple conditions
func (psf *PostgreSQLFormatter) tokenOverride(tok types.Token, previousReservedWord types.Token) types.Token {
	// Handle type cast operator :: - format without spaces (PostgreSQL convention)
	if tok.Type == types.TokenTypeOperator && tok.Value == "::" {
		return types.Token{
			Type:  types.TokenTypeSpecialOperator,
			Value: tok.Value,
			Key:   tok.Key,
		}
	}

	// Handle DO keyword context-awareness for UPSERT vs standalone blocks
	// In UPSERT context (after ON CONFLICT), DO should not create a top-level break
	// In standalone context, DO should create a top-level break for PL/pgSQL blocks
	if tok.Type == types.TokenTypeReservedTopLevel && tok.Value == "DO" {
		// Check if we're in UPSERT context by looking at previous reserved word
		// Previous word could be "ON CONFLICT" directly, or "WHERE" (in ON CONFLICT ... WHERE ... DO pattern)
		if !previousReservedWord.Empty() {
			prevVal := strings.ToUpper(previousReservedWord.Value)
			if strings.Contains(prevVal, "CONFLICT") || prevVal == "WHERE" {
				// In UPSERT context: downgrade to regular reserved
				// This prevents unwanted newline and keeps UPSERT clause together
				return types.Token{
					Type:  types.TokenTypeReserved,
					Value: tok.Value,
					Key:   tok.Key,
				}
			}
		}
		// Otherwise keep as top-level for standalone DO blocks
	}

	// Handle UPDATE keyword context-awareness for UPSERT
	// After "DO" in UPSERT context, UPDATE should not create top-level break
	if tok.Type == types.TokenTypeReservedTopLevel && tok.Value == "UPDATE" {
		// Check if previous reserved word is "DO" or starts with "DO " (for "DO NOTHING")
		prevVal := strings.ToUpper(previousReservedWord.Value)
		if prevVal == "DO" || strings.HasPrefix(prevVal, "DO ") {
			// In UPSERT context after DO: downgrade UPDATE to regular reserved
			return types.Token{
				Type:  types.TokenTypeReserved,
				Value: tok.Value,
				Key:   tok.Key,
			}
		}
	}

	return tok
}
