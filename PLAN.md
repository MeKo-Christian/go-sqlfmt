# go-sqlfmt Implementation Plan

This consolidated plan covers all remaining work for MySQL, PostgreSQL, and SQLite dialects, plus strategic enhancements for production readiness.

## Phase 1: Advanced SQL Features

### 1.1 Enhanced Stored Procedure Support

- [ ] PostgreSQL PL/pgSQL improvements:
  - [ ] Better block indentation (BEGIN/END, IF/ELSE, LOOP)
  - [ ] Exception handling formatting
  - [ ] Variable declaration formatting
- [ ] MySQL stored procedure improvements:
  - [ ] Better DELIMITER handling
  - [ ] Control flow structure formatting
  - [ ] Cursor handling
- [ ] SQLite trigger improvements:
  - [ ] Multi-statement trigger bodies
  - [ ] BEFORE/AFTER/INSTEAD OF formatting

### 1.2 Complete DDL Support

- [ ] Full CREATE TABLE support:
  - [ ] All constraint types
  - [ ] Table inheritance (PostgreSQL)
  - [ ] Partitioning clauses
  - [ ] Storage parameters
- [ ] Index creation options:
  - [ ] Partial indexes
  - [ ] Expression indexes
  - [ ] Index storage parameters
- [ ] ALTER TABLE statements:
  - [ ] All modification types
  - [ ] Multi-column changes
  - [ ] Constraint modifications

## Phase 2: Compact IF/THEN Formatting

**Goal**: Format `IF condition THEN` on a single line instead of:
```sql
IF
  condition THEN
```

### Completed Work
- [x] Tokenizer: Sort close-paren patterns by length (longer matches first)
- [x] Tokenizer: `END IF`, `END WHILE`, `END LOOP`, `END REPEAT` now recognized as single tokens
- [x] Result: `END IF;` and `END LOOP label;` format on single lines

### Remaining Work

**Problem**: The formatter treats IF/THEN/ELSE identically to CASE/WHEN/ELSE. Adding special handling for IF breaks CASE statement formatting.

**Solution**: Track context (IF block vs CASE expression) to apply different formatting rules.

---

#### 2.1 Add Block Context Stack (Infrastructure - No Behavior Change) ✅

**Goal**: Add tracking infrastructure without changing any output.

**Files**: `pkg/sqlfmt/core/formatter.go`, `pkg/sqlfmt/core/formatter_blockstack_test.go`

- [x] Add `blockStack []string` field to `formatter` struct
- [x] Add helper methods:
  - `pushBlock(blockType string)` - push to stack
  - `popBlock()` - pop from stack
  - `currentBlock() string` - peek top (empty string if empty)
  - `isInBlock(blockType string) bool` - check if anywhere in stack
- [x] Call `pushBlock("CASE")` in `formatOpeningParentheses()` when token is CASE
- [x] Call `pushBlock("IF")` in `formatOpeningParentheses()` when token is IF
- [x] Call `pushBlock("BEGIN")` in `formatOpeningParentheses()` when token is BEGIN
- [x] Call `popBlock()` in `formatClosingParentheses()` for END, END IF, END LOOP, etc.
- [x] Comprehensive tests for helper methods
- [x] Tests for MySQL, PostgreSQL, and Standard SQL block tracking
- [x] Tests verify no output changes

**Acceptance**: ✅ All existing tests pass unchanged. New tests verify stack behavior.

---

### Architectural Analysis

**Why simple patches fail:**

The formatter was designed for flat SQL queries where `;` separates independent queries. Stored procedures introduce nested blocks where `;` terminates statements but doesn't separate queries. Key issues:

1. **Semicolon semantics**: `formatQuerySeparator()` calls `ResetIndentation()` unconditionally
2. **Indentation origin**: The `Indentation` struct tracks indent *levels* but not *sources* (top-level from CREATE PROCEDURE vs block-level from BEGIN)
3. **addNewline() coupling**: Always adds current indentation, causing double-indent when called by both semicolon and subsequent keyword
4. **ELSE ambiguity**: Same token used in IF blocks (procedural) and CASE expressions (declarative) with different formatting needs

**Sustainable solution:** Refactor the indentation system to understand procedural blocks as a distinct concept.

---

#### 2.2 Refactor Indentation to Track Sources ✅

**Goal**: Make the indentation system aware of WHERE each indent came from.

**Files**: `pkg/sqlfmt/utils/indentation.go`

- [x] Add `indentSource` field to track origin: `"top-level"`, `"block"`, `"procedural-block"`
- [x] Create `IndentEntry` struct: `{type: indentType, source: string, keyword: string}`
- [x] Replace `indentTypes []indentType` with `indentStack []IndentEntry`
- [x] Add methods:
  - `IncreaseProcedural(keyword string)` - for BEGIN, IF, LOOP, etc.
  - `DecreaseProcedural()` - only removes procedural indents
  - `GetProceduralDepth() int` - count of procedural blocks
  - `ResetToProceduralBase()` - keep procedural indents, reset top-level
- [x] Preserve backward compatibility: existing methods work unchanged

**Tests**:

- [x] Unit tests for new IndentEntry tracking
- [x] Verify existing tests pass (no behavior change yet)

**Acceptance**: ✅ All existing tests pass. New methods available for use.

---

#### 2.3 Add Procedural Block Tracking to Formatter ✅

**Goal**: Formatter tracks when it enters/exits procedural blocks (BEGIN/END).

**Files**: `pkg/sqlfmt/core/formatter.go`, `pkg/sqlfmt/core/formatter_procedural_test.go`

- [x] Add `proceduralDepth int` field to formatter struct
- [x] Increment in `formatOpeningParentheses()` when token is BEGIN
- [x] Decrement in `formatClosingParentheses()` when token is END
- [x] Add `isInProceduralBlock() bool` helper method
- [x] Use `IncreaseProcedural()` instead of `IncreaseBlockLevel()` for BEGIN
- [x] Document known issue in tests: END keywords currently at column 0 (will be fixed in 2.5)

**Tests**:
- Test procedural depth tracking across various queries
- Verify existing behavior unchanged
- Added TODO comments noting END indentation issue to be fixed in task 2.5

**Known Issue**: END keywords are currently formatted at column 0 instead of aligning with their opening BEGIN. This is documented in test comments and will be fixed in task 2.5.

**Acceptance**: ✅ Procedural depth accurately tracked. All tests pass.

---

#### 2.4 Differentiate Statement Terminators from Query Separators

**Goal**: Semicolons inside procedural blocks behave as statement terminators, not query separators.

**Files**: `pkg/sqlfmt/core/formatter.go`

- [x] Modify `formatQuerySeparator()`:
  ```go
  if f.isInProceduralBlock() {
      // Statement terminator: keep procedural indentation
      f.indentation.ResetToProceduralBase()
      trimSpacesEnd(query)
      query.WriteString(tok.Value)
      query.WriteString("\n")
  } else {
      // Query separator: full reset + blank lines
      f.indentation.ResetIndentation()
      trimSpacesEnd(query)
      query.WriteString(tok.Value)
      query.WriteString(strings.Repeat("\n", f.cfg.LinesBetweenQueries))
  }
  ```

**Tests**:
- `BEGIN DECLARE x; SELECT 1; END;` - statements maintain procedural indent
- `SELECT 1; SELECT 2;` - queries separated with blank lines (unchanged)
- Nested: `BEGIN BEGIN SELECT 1; END; END;` - proper indent levels

**Acceptance**: Semicolons respect procedural context. Existing query tests unchanged.

---

#### 2.5 Fix END Keyword Indentation ✅

**Goal**: END aligns with its opening keyword (BEGIN), not affected by top-level indents.

**Files**: `pkg/sqlfmt/core/formatter.go`, `pkg/sqlfmt/core/formatter_end_indentation_test.go`

- [x] Modify `formatClosingParentheses()` for END:
  - Call `ResetToProceduralBase()` before formatting END to clear top-level indents
  - Call `DecreaseProcedural()` for procedural block closers (BEGIN, IF, etc.)
  - Keep normal `DecreaseBlockLevel()` for CASE END
- [x] Handle END IF, END LOOP, END WHILE, END REPEAT similarly
- [x] Created comprehensive test suite for END keyword indentation
- [x] Fixed existing tests that were affected by changed END indentation

**Tests**:
- `CREATE PROCEDURE foo() BEGIN SELECT 1; END;` - END at column 0 ✅
- Nested IF inside BEGIN - END IF at procedural base ✅
- Multiple nested blocks ✅
- CASE END maintains normal indentation ✅

**Acceptance**: ✅ All END keywords align correctly. Procedural block ENDs reset to procedural base, clearing accumulated top-level indents from statements inside the block.

---

#### 2.6 Compact IF Formatting

**Goal**: `IF condition THEN` on single line.

**Files**: `pkg/sqlfmt/core/formatter.go`

- [ ] Modify `formatOpeningParentheses()` for IF:
  - Use `IncreaseProcedural("IF")`
  - Write space instead of newline after IF
- [ ] Condition and THEN stay on same line as IF

**Tests**:
- `IF x > 0 THEN SELECT 1; END IF;` → `IF x > 0 THEN` on one line
- Nested IF statements
- IF with complex conditions

**Acceptance**: IF formatting is compact. CASE formatting unchanged.

---

#### 2.7 Context-Aware ELSE/ELSEIF Handling

**Goal**: ELSE in IF blocks aligns with IF. ELSE in CASE stays indented.

**Files**: `pkg/sqlfmt/core/formatter.go`

- [ ] Modify `formatNewlineReservedWord()` for ELSE/ELSEIF:
  ```go
  if f.currentBlock() == "IF" {
      f.indentation.DecreaseProcedural()  // Close IF block
      f.addNewline(query)
      f.formatKeyword(tok, query)
      f.indentation.IncreaseProcedural("ELSE")  // Open ELSE block
  } else {
      // CASE ELSE: default behavior
      f.addNewline(query)
      f.formatKeyword(tok, query)
  }
  ```

**Tests**:
- IF/ELSE alignment test
- IF/ELSEIF/ELSE alignment test
- CASE/WHEN/ELSE test (must remain unchanged)
- Nested IF inside CASE

**Acceptance**: IF and CASE ELSE format correctly and independently.

---

#### 2.8 Comprehensive Integration Tests

**Goal**: Verify all stored procedure patterns work correctly across dialects.

**Files**: `pkg/sqlfmt/mysql_formatter_ddl_test.go`, `pkg/sqlfmt/postgresql_formatter_test.go`

- [ ] MySQL stored procedures with all control structures
- [ ] PostgreSQL PL/pgSQL functions
- [ ] Nested control structures (IF inside LOOP inside BEGIN)
- [ ] Multiple procedures in one file
- [ ] Edge cases: empty blocks, single-statement blocks

**Acceptance**: All dialect-specific stored procedure tests pass.

---

### Implementation Order

```
2.2 Refactor Indentation (foundation)
 ↓
2.3 Procedural Block Tracking (uses 2.2)
 ↓
2.4 Statement vs Query Separator (uses 2.3)
 ↓
2.5 END Indentation Fix (uses 2.4)
 ↓
2.6 Compact IF (uses 2.5)
 ↓
2.7 ELSE Handling (uses 2.6)
 ↓
2.8 Integration Tests (validates all)
```

**Each step is independently committable with passing tests.**

**Estimated complexity**: Medium-high. The indentation refactor (2.2) is the foundation - get it right and subsequent steps become straightforward.
