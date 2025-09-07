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

- [x] Add dollar-quote pattern support to tokenizer
  - [x] Add `$$` pattern to `createStringPattern()` in `tokenizer.go`
  - [x] Create helper function `scanDollarQuotedString()` in tokenizer
  - [x] Update `getStringToken()` to handle dollar-quoted strings
- [x] Add comprehensive dollar-quote tests
  - [x] Test basic dollar-quoted strings (`$$...$$`)
  - [x] Test tagged dollar-quoted strings (`$tag$...$tag$`)
  - [x] Test nested quotes inside dollar-quotes
  - [x] Test multi-line dollar-quoted strings

## Phase 4: PostgreSQL Numbered Placeholders ($1, $2, ...)

- [x] Update placeholder handling for PostgreSQL
  - [x] Add `$` to `IndexedPlaceholderTypes` in PostgreSQL config
  - [x] Update placeholder regex to recognize `$1`, `$2`, etc.
  - [x] Modify `params.go` to handle 1-based indexing for `$n` placeholders
- [x] Add numbered placeholder tests
  - [x] Test numbered placeholder formatting
  - [x] Test parameter substitution with `$n`
  - [x] Test mixed numbered and named placeholders
  - [x] Test edge cases (high numbers, $0)

## Phase 5: Type Cast Operator (::)

- [x] Implement cast operator support
  - [x] Recognize `::` as a special operator in tokenizer
  - [x] Add `::` to operator regex pattern
  - [x] Create special formatting rule for `::` (no spaces)
  - [x] Update `formatWithSpaces()` to handle `::` specially
- [x] Add type cast tests
  - [x] Test basic type casts (`expr::type`)
  - [x] Test complex type casts with functions
  - [x] Test array type casts
  - [x] Test chained operations with casts

## Phase 6: JSON/JSONB Operators

- [x] Add JSON path operators
  - [x] Add JSON operators to operator regex: `->`, `->>`
  - [x] Add JSONB operators: `#>`, `#>>`
  - [x] Ensure proper spacing around JSON operators
- [x] Add containment and existence operators
  - [x] Add containment operators: `@>`, `<@`
  - [x] Add existence operators: `?`, `?|`, `?&`
  - [x] Handle operator precedence correctly
- [x] Add comprehensive JSON tests
  - [x] Test JSON path extraction
  - [x] Test JSONB containment queries
  - [x] Test existence operators
  - [x] Test complex nested JSON operations

## Phase 7: Pattern Matching Operators

- [x] Add case-insensitive pattern matching
  - [x] Add `ILIKE` to reserved words
  - [x] Add `SIMILAR TO` to reserved words (multi-word)
  - [x] Test pattern matching formatting
- [x] Add regex operators
  - [x] Add regex operators: `~`, `!~`, `~*`, `!~*`
  - [x] Update operator regex to handle these multi-char operators
  - [x] Test regex operator formatting
- [x] Add pattern matching tests
  - [x] Test ILIKE queries
  - [x] Test SIMILAR TO queries
  - [x] Test regex matching queries
  - [x] Test negated pattern operators

## Phase 8: Core PostgreSQL Keywords - Part 1 (CTEs and RETURNING)

- [x] Add Common Table Expression support
  - [x] Add top-level words: `WITH`, `WITH RECURSIVE`
  - [x] Ensure proper indentation for CTEs
  - [x] Handle multiple CTEs correctly
- [x] Add RETURNING clause support
  - [x] Add `RETURNING` to top-level words
  - [x] Handle RETURNING with INSERT/UPDATE/DELETE
- [x] Add UPSERT support
  - [x] Add `ON CONFLICT` to reserved words
  - [x] Add `DO UPDATE`, `DO NOTHING` keywords
  - [x] Handle conflict resolution formatting
- [x] Add CTE and RETURNING tests
  - [x] Test simple WITH queries
  - [x] Test recursive CTE queries
  - [x] Test INSERT...RETURNING
  - [x] Test UPSERT queries with ON CONFLICT

## Phase 9: Core PostgreSQL Keywords - Part 2 (Window Functions)

- [x] Add window function keywords
  - [x] Add window function keywords: `WINDOW`, `OVER`, `PARTITION BY`
  - [x] Add `FILTER (WHERE ...)` support
  - [x] Handle window frame specifications: `RANGE`, `ROWS`
- [x] Add lateral join support
  - [x] Add `LATERAL` join support
  - [x] Ensure proper JOIN formatting with LATERAL
- [x] Add ordering enhancements
  - [x] Add ordering modifiers: `NULLS FIRST`, `NULLS LAST`
  - [x] Handle these as single units
- [x] Add window function tests
  - [x] Test basic window functions
  - [x] Test aggregate functions with FILTER
  - [x] Test LATERAL join queries
  - [x] Test complex window specifications

## Phase 10: Array and Range Support

- [x] Handle array subscript operations
  - [x] Handle array subscripts: `[1]`, `[1:2]`
  - [x] Ensure brackets don't break formatting
  - [x] Handle multi-dimensional arrays
- [x] Add concatenation operators
  - [x] Add `||` operator for array/string concatenation
  - [x] Handle operator precedence with ||
- [x] Add array and range tests
  - [x] Test array subscript operations
  - [x] Test array concatenation
  - [x] Test range operations
  - [x] Test complex array expressions

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

- [x] **Set up snapshot testing infrastructure**
  - [x] Add go-snaps dependency for Jest-like snapshot testing
  - [x] Create comprehensive snapshot test suite for all SQL dialects
  - [x] Add Justfile commands: `just test-snapshots`, `just update-snapshots`
  - [x] Configure TestMain for obsolete snapshot detection
  - [x] Test coverage for Standard SQL, PostgreSQL, N1QL, DB2, and PL/SQL
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
