# PostgreSQL Support Implementation Plan

Based on colleague's gap analysis, this plan adds comprehensive PostgreSQL support to go-sqlfmt in small, manageable phases.

## Phase 1: Foundation - Add PostgreSQL Language Constant

- [x] Add `PostgreSQL` constant to `Language` type in `config.go`
- [x] Update `getFormatter()` in `format.go` to handle PostgreSQL case
- [x] Create empty `postgresql_formatter.go` file with basic struct
  - [x] Create `PostgreSQLFormatter` struct
  - [x] Add `NewPostgreSQLFormatter()` constructor
  - [x] Add `Format()` method stub
- [x] Create empty `postgresql_formatter_test.go` file

## Phase 2: Basic PostgreSQL Tokenizer Config

- [x] Create `NewPostgreSQLTokenizerConfig()` function in `postgresql_formatter.go`
  - [x] Copy Standard SQL config as baseline
  - [x] Add PostgreSQL-specific line comment types (`--`)
  - [x] Add basic PostgreSQL reserved words (minimal set)
- [x] Test basic PostgreSQL formatter initialization
- [x] Test simple query formatting works

## Phase 3: Dollar-Quoted String Support

- [ ] Add dollar-quote pattern support to tokenizer
  - [ ] Add `$$` pattern to `createStringPattern()` in `tokenizer.go`
  - [ ] Create helper function `scanDollarQuotedString()` in tokenizer
  - [ ] Update `getStringToken()` to handle dollar-quoted strings
- [ ] Add comprehensive dollar-quote tests
  - [ ] Test basic dollar-quoted strings (`$$...$$`)
  - [ ] Test tagged dollar-quoted strings (`$tag$...$tag$`)
  - [ ] Test nested quotes inside dollar-quotes
  - [ ] Test multi-line dollar-quoted strings

## Phase 4: PostgreSQL Numbered Placeholders ($1, $2, ...)

- [ ] Update placeholder handling for PostgreSQL
  - [ ] Add `$` to `IndexedPlaceholderTypes` in PostgreSQL config
  - [ ] Update placeholder regex to recognize `$1`, `$2`, etc.
  - [ ] Modify `params.go` to handle 1-based indexing for `$n` placeholders
- [ ] Add numbered placeholder tests
  - [ ] Test numbered placeholder formatting
  - [ ] Test parameter substitution with `$n`
  - [ ] Test mixed numbered and named placeholders
  - [ ] Test edge cases (high numbers, $0)

## Phase 5: Type Cast Operator (::)

- [ ] Implement cast operator support
  - [ ] Recognize `::` as a special operator in tokenizer
  - [ ] Add `::` to operator regex pattern
  - [ ] Create special formatting rule for `::` (no spaces)
  - [ ] Update `formatWithSpaces()` to handle `::` specially
- [ ] Add type cast tests
  - [ ] Test basic type casts (`expr::type`)
  - [ ] Test complex type casts with functions
  - [ ] Test array type casts
  - [ ] Test chained operations with casts

## Phase 6: JSON/JSONB Operators

- [ ] Add JSON path operators
  - [ ] Add JSON operators to operator regex: `->`, `->>`
  - [ ] Add JSONB operators: `#>`, `#>>`
  - [ ] Ensure proper spacing around JSON operators
- [ ] Add containment and existence operators
  - [ ] Add containment operators: `@>`, `<@`
  - [ ] Add existence operators: `?`, `?|`, `?&`
  - [ ] Handle operator precedence correctly
- [ ] Add comprehensive JSON tests
  - [ ] Test JSON path extraction
  - [ ] Test JSONB containment queries
  - [ ] Test existence operators
  - [ ] Test complex nested JSON operations

## Phase 7: Pattern Matching Operators

- [ ] Add case-insensitive pattern matching
  - [ ] Add `ILIKE` to reserved words
  - [ ] Add `SIMILAR TO` to reserved words (multi-word)
  - [ ] Test pattern matching formatting
- [ ] Add regex operators
  - [ ] Add regex operators: `~`, `!~`, `~*`, `!~*`
  - [ ] Update operator regex to handle these multi-char operators
  - [ ] Test regex operator formatting
- [ ] Add pattern matching tests
  - [ ] Test ILIKE queries
  - [ ] Test SIMILAR TO queries
  - [ ] Test regex matching queries
  - [ ] Test negated pattern operators

## Phase 8: Core PostgreSQL Keywords - Part 1 (CTEs and RETURNING)

- [ ] Add Common Table Expression support
  - [ ] Add top-level words: `WITH`, `WITH RECURSIVE`
  - [ ] Ensure proper indentation for CTEs
  - [ ] Handle multiple CTEs correctly
- [ ] Add RETURNING clause support
  - [ ] Add `RETURNING` to top-level words
  - [ ] Handle RETURNING with INSERT/UPDATE/DELETE
- [ ] Add UPSERT support
  - [ ] Add `ON CONFLICT` to reserved words
  - [ ] Add `DO UPDATE`, `DO NOTHING` keywords
  - [ ] Handle conflict resolution formatting
- [ ] Add CTE and RETURNING tests
  - [ ] Test simple WITH queries
  - [ ] Test recursive CTE queries
  - [ ] Test INSERT...RETURNING
  - [ ] Test UPSERT queries with ON CONFLICT

## Phase 9: Core PostgreSQL Keywords - Part 2 (Window Functions)

- [ ] Add window function keywords
  - [ ] Add window function keywords: `WINDOW`, `OVER`, `PARTITION BY`
  - [ ] Add `FILTER (WHERE ...)` support
  - [ ] Handle window frame specifications: `RANGE`, `ROWS`
- [ ] Add lateral join support
  - [ ] Add `LATERAL` join support
  - [ ] Ensure proper JOIN formatting with LATERAL
- [ ] Add ordering enhancements
  - [ ] Add ordering modifiers: `NULLS FIRST`, `NULLS LAST`
  - [ ] Handle these as single units
- [ ] Add window function tests
  - [ ] Test basic window functions
  - [ ] Test aggregate functions with FILTER
  - [ ] Test LATERAL join queries
  - [ ] Test complex window specifications

## Phase 10: Array and Range Support

- [ ] Handle array subscript operations
  - [ ] Handle array subscripts: `[1]`, `[1:2]`
  - [ ] Ensure brackets don't break formatting
  - [ ] Handle multi-dimensional arrays
- [ ] Add concatenation operators
  - [ ] Add `||` operator for array/string concatenation
  - [ ] Handle operator precedence with ||
- [ ] Add array and range tests
  - [ ] Test array subscript operations
  - [ ] Test array concatenation
  - [ ] Test range operations
  - [ ] Test complex array expressions

## Phase 11: DO Blocks and PL/pgSQL

- [ ] Handle procedural blocks
  - [ ] Recognize `DO` as a keyword
  - [ ] Treat dollar-quoted body as opaque string (no formatting)
  - [ ] Handle `LANGUAGE` keyword properly
- [ ] Add function definition support
  - [ ] Handle `CREATE FUNCTION` statements
  - [ ] Handle `RETURNS` keyword
  - [ ] Handle function modifiers: `IMMUTABLE`, `STABLE`, `VOLATILE`
- [ ] Add procedural tests
  - [ ] Test DO blocks
  - [ ] Test function definitions
  - [ ] Test complex PL/pgSQL blocks
  - [ ] Test function calls with parameters

## Phase 12: DDL and Index Keywords

- [ ] Add concurrent operation support
  - [ ] Add `CONCURRENTLY` keyword
  - [ ] Handle concurrent index operations
- [ ] Add index method support
  - [ ] Add `USING` for index methods (GIN, GIST, BTREE, etc.)
  - [ ] Add `INCLUDE` for covering indexes
  - [ ] Handle index options properly
- [ ] Add DDL tests
  - [ ] Test CREATE INDEX statements
  - [ ] Test concurrent operations
  - [ ] Test ALTER TABLE statements
  - [ ] Test covering indexes

## Phase 13: Comprehensive Integration Testing

- [ ] Create golden test file with complex PostgreSQL queries
  - [ ] Test multi-CTE queries with RECURSIVE
  - [ ] Test UPSERT queries (INSERT...ON CONFLICT)
  - [ ] Test complex JOIN queries with LATERAL
  - [ ] Test aggregate functions with FILTER clauses
- [ ] Test real-world scenarios
  - [ ] Test migration scripts
  - [ ] Test stored procedure definitions
  - [ ] Test complex analytical queries
  - [ ] Test mixed DDL and DML operations
- [ ] Performance and edge case testing
  - [ ] Test with very large queries
  - [ ] Test deeply nested structures
  - [ ] Test malformed input handling
  - [ ] Test memory usage with complex queries

## Phase 14: Documentation and Examples

- [ ] Update project documentation
  - [ ] Update README.md with PostgreSQL usage examples
  - [ ] Add PostgreSQL to supported dialects list
  - [ ] Create comprehensive example code snippets
  - [ ] Update CLAUDE.md with PostgreSQL testing commands
- [ ] Add inline documentation
  - [ ] Document PostgreSQL-specific functions
  - [ ] Add code comments for complex logic
  - [ ] Document configuration options
  - [ ] Add usage examples in godoc

## Phase 15: Final Polish and Edge Cases

- [ ] Handle identifier edge cases
  - [ ] Handle quoted identifiers properly
  - [ ] Test case sensitivity rules
  - [ ] Test Unicode identifiers
- [ ] Final validation
  - [ ] Test with real-world PostgreSQL queries
  - [ ] Performance testing with large queries
  - [ ] Benchmark against existing formatters
  - [ ] Run full test suite and ensure no regressions
- [ ] Code cleanup
  - [ ] Refactor any duplicate code
  - [ ] Optimize performance bottlenecks
  - [ ] Add missing error handling
  - [ ] Final code review and cleanup

## Success Criteria

- [ ] All PostgreSQL-specific syntax is properly formatted
- [ ] No regressions in existing SQL dialect support
- [ ] Comprehensive test coverage (>90%)
- [ ] Performance comparable to existing formatters
- [ ] Documentation complete and accurate
