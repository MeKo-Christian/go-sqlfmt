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
- [-] Add UPSERT support (**BLOCKED** - see Known Issues section)
  - [x] Add `ON CONFLICT` to reserved words
  - [x] Add `DO UPDATE`, `DO NOTHING` keywords
  - [ ] Handle conflict resolution formatting (tokenizer limitations)
- [x] Add CTE and RETURNING tests
  - [x] Test simple WITH queries
  - [x] Test recursive CTE queries
  - [x] Test INSERT...RETURNING
  - [ ] Test UPSERT queries with ON CONFLICT (failing due to formatting issues)

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

- [x] Handle procedural blocks
  - [x] Recognize `DO` as a keyword (implemented but conflicts with UPSERT)
  - [x] Treat dollar-quoted body as opaque string (no formatting)
  - [x] Handle `LANGUAGE` keyword properly
- [-] Add function definition support (**PARTIAL** - see Known Issues section)
  - [x] Handle `CREATE FUNCTION` statements
  - [x] Handle `RETURNS` keyword
  - [x] Handle function modifiers: `IMMUTABLE`, `STABLE`, `VOLATILE`
  - [ ] Function indentation formatting issues (some tests fail)
- [-] Add procedural tests (**PARTIAL**)
  - [x] Test DO blocks (basic tests pass)
  - [ ] Test function definitions (some indentation failures)
  - [x] Test complex PL/pgSQL blocks
  - [x] Test function calls with parameters

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

## Known Issues and Implementation Challenges

### UPSERT Formatting Problem (Phase 8)

**Issue**: UPSERT statements with `ON CONFLICT` clauses are not formatting correctly.

**Expected Behavior**:

```sql
INSERT INTO users (id, name, email)
VALUES (1, 'John', 'john@example.com') ON CONFLICT (id) DO NOTHING;
```

**Current Behavior**:

```sql
INSERT INTO users (id, name, email)
VALUES
  (1, 'John', 'john@example.com') ON CONFLICT (id)
DO
  NOTHING;
```

**Root Cause Analysis**:

1. **Conflicting Requirements**:
   - Procedural `DO` blocks (`DO $$ ... $$`) need to be top-level (new line)
   - UPSERT `DO` clauses (`DO UPDATE`, `DO NOTHING`) need to stay inline
2. **Tokenizer Limitation**:
   - Compound reserved words like "DO UPDATE" and "DO NOTHING" are not being tokenized as single units
   - The tokenizer sees "DO" as a separate token, making context detection difficult
3. **Context Detection Challenges**:
   - By the time the formatter processes "DO", there can be intervening tokens (identifiers, parentheses) between "ON CONFLICT" and "DO"
   - The `previousReservedWord` tracking doesn't reliably identify UPSERT context
   - Simple heuristics fail because of the complex token stream

**Attempted Solutions**:

1. ✗ Removed "DO" from top-level words + tokenizer override based on `previousReservedWord`
2. ✗ Context detection using previous reserved word patterns
3. ✗ Heuristic-based approach checking for procedural vs UPSERT indicators
4. ✗ Assumption that compound keywords would be tokenized as single units

**Technical Obstacles**:

- The tokenizer architecture doesn't support lookahead to determine if "DO" is part of "DO UPDATE"
- Token override functions only have access to the current token and previous reserved word
- Compound reserved word tokenization is not working as expected for multi-word UPSERT constructs
- PostgreSQL UPSERT syntax is more complex than initially anticipated

**Impact**:

- UPSERT tests (`TestPostgreSQLFormatter_UPSERT`) currently fail
- Procedural DO blocks work correctly
- This affects Phase 8 completion status

**Potential Future Solutions**:

1. Modify the tokenizer core to properly handle compound keywords like "DO UPDATE" as single tokens
2. Implement a two-pass formatter that can look ahead to determine context
3. Add special handling in the core formatting logic (not just token override)
4. Consider alternative test expectations that align with current formatting capabilities

**Status**: ❌ **Unresolved** - Requires deeper architectural changes to the tokenizer or formatting core.

### Function Indentation Problem (Phase 11)

**Issue**: Some PostgreSQL function definitions are not formatting with the expected indentation.

**Expected Behavior**:

```sql
CREATE FUNCTION
      get_secure_data(user_id INTEGER) RETURNS TABLE(id INTEGER, name TEXT) AS $$
      SELECT id, name FROM users WHERE id = user_id AND active = true;
      $$ LANGUAGE SQL STABLE SECURITY DEFINER;
```

**Current Behavior**:

```sql
CREATE FUNCTION
  get_secure_data(user_id INTEGER) RETURNS TABLE(id INTEGER, name TEXT) AS $$
			SELECT id, name FROM users WHERE id = user_id AND active = true;
			$$ LANGUAGE SQL STABLE SECURITY DEFINER;
```

**Root Cause**: Test expectations may not match the current indentation algorithm used by the formatter.

**Impact**:

- Some function tests in `TestPostgreSQLFormatter_Functions` fail
- Basic function formatting works correctly
- Only affects complex multi-line function definitions

**Status**: ❌ **Test expectation mismatch** - May require test updates rather than code changes.

---

## Success Criteria

- [ ] All PostgreSQL-specific syntax is properly formatted
- [ ] No regressions in existing SQL dialect support
- [ ] Comprehensive test coverage (>90%)
- [ ] Performance comparable to existing formatters
- [ ] Documentation complete and accurate

**Current Status**: Most PostgreSQL features implemented successfully. UPSERT formatting remains unresolved due to tokenizer architecture limitations.
