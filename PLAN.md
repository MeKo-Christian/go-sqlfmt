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

#### 2.2 Preserve Indentation Inside BEGIN/END Blocks

**Goal**: Semicolons inside stored procedures shouldn't reset indentation.

**Files**: `pkg/sqlfmt/core/formatter.go`, `pkg/sqlfmt/utils/indentation.go`

- [ ] Add `IsInsideBlock() bool` method to `Indentation` struct
- [ ] Modify `formatQuerySeparator()`:
  - If `isInBlock("BEGIN")`: don't call `ResetIndentation()`, just add newline
  - Else: keep current behavior (reset + blank lines between queries)

**Test**: Write test with `BEGIN DECLARE x; SELECT 1; END;` - verify DECLARE and SELECT maintain indentation.

**Acceptance**: Stored procedure tests pass with proper indentation. All other tests unchanged.

---

#### 2.3 Compact IF Formatting (IF condition THEN on one line)

**Goal**: `IF x > 0 THEN` on single line instead of `IF\n  x > 0 THEN`.

**Files**: `pkg/sqlfmt/core/formatter.go`

- [ ] Write failing test: `IF x > 0 THEN y; END IF;` → expects `IF x > 0 THEN` on one line
- [ ] Modify `formatOpeningParentheses()`:
  - If token is IF: write space instead of newline after IF
- [ ] Verify test passes

**Acceptance**: New IF test passes. CASE tests still pass (CASE doesn't use OpenParen).

---

#### 2.4 Context-Aware ELSE/ELSEIF Handling

**Goal**: ELSE in IF blocks should align with IF. ELSE in CASE should stay indented.

**Files**: `pkg/sqlfmt/core/formatter.go`

- [ ] Write failing test for IF/ELSE alignment
- [ ] Write test verifying CASE/ELSE still works
- [ ] Modify `formatNewlineReservedWord()` for ELSE/ELSEIF:
  - If `currentBlock() == "IF"`: decrease block level, format, increase block level
  - Else: keep current behavior (just newline + format)
- [ ] Verify both tests pass

**Acceptance**: IF/ELSE aligns correctly. CASE/ELSE unchanged.

---

#### 2.5 END IF Indentation Fix

**Goal**: `END IF` should align with the opening `IF`.

**Files**: `pkg/sqlfmt/core/formatter.go`

- [ ] Write failing test: nested IF inside LOOP should have END IF at correct indent
- [ ] Modify `formatClosingParentheses()`:
  - Before decreasing block level, check if this closes an IF block
  - Ensure indentation matches the IF level

**Acceptance**: Nested control structures indent correctly.

---

### Implementation Order

1. **2.1** - Add infrastructure (safe, no behavior change)
2. **2.2** - Fix semicolon handling (fixes stored procedures)
3. **2.3** - Compact IF (isolated change, doesn't affect CASE)
4. **2.4** - ELSE handling (uses block context to differentiate IF vs CASE)
5. **2.5** - END IF alignment (polish)

Each step is independently committable with passing tests.
