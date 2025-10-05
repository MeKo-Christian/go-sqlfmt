# go-sqlfmt Implementation Plan

This consolidated plan covers all remaining work for MySQL, PostgreSQL, and SQLite dialects, plus strategic enhancements for production readiness.

## Overview

**Remaining Focus Areas:**

1. ✅ ~~Fix failing PostgreSQL formatter test~~ - COMPLETED
2. ✅ ~~Fix failing MySQL formatter tests (11 tests)~~ - COMPLETED
3. Integration & performance tests
4. Documentation improvements

---

## Phase 1: Test Coverage Improvement

### 1.1 Integration Tests

#### Cross-Dialect Tests
- [ ] Test dialect-specific formatting differences
- [ ] Test migration between dialects
- [ ] Test edge cases across all dialects

#### Golden File Tests
- [ ] Expand golden files:
  - [ ] Add more complex PostgreSQL queries
  - [ ] Add more MySQL edge cases
  - [ ] Add SQLite-specific syntax
  - [ ] Add cross-dialect examples
- [ ] Automate golden file generation:
  - [ ] Script to generate test files
  - [ ] Validation of golden file quality

#### Scenario Tests
- [ ] Ensure all scenario files are tested
- [ ] Add more real-world scenarios:
  - [ ] ORM-generated queries
  - [ ] Migration scripts
  - [ ] Backup/restore scripts
  - [ ] Complex analytics queries

### 1.2 Performance & Stress Tests

#### Memory Tests (Already exist, run with `-short=false`)
- ✅ Memory leak detection
- ✅ Allocation profiling
- ✅ Concurrent formatting

#### Benchmark Tests
- [ ] Add comprehensive benchmarks:
  - [ ] Small queries (< 100 chars)
  - [ ] Medium queries (100-1000 chars)
  - [ ] Large queries (1000-10000 chars)
  - [ ] Very large queries (10000+ chars)
- [ ] Benchmark by dialect
- [ ] Benchmark configuration options impact

#### Fuzz Tests
- [ ] Add fuzzing for robustness:
  - [ ] Random SQL generation
  - [ ] Random character sequences
  - [ ] Invalid input handling
  - [ ] Unicode edge cases

### 1.3 Test Infrastructure Improvements

#### Test Helpers
- [ ] Create reusable test utilities:
  - [ ] Query builders for test cases
  - [ ] Assertion helpers for formatting
  - [ ] Config generators
  - [ ] Mock file system for config tests

#### CI/CD Integration
- [ ] Add coverage reporting to CI:
  - [ ] Generate coverage reports
  - [ ] Track coverage trends
  - [ ] Fail on coverage regression
  - [ ] Badge for README

#### Test Documentation
- [ ] Document testing approach:
  - [ ] How to run specific tests
  - [ ] How to add new tests
  - [ ] Test organization philosophy
  - [ ] Coverage goals per package

---

## Phase 2: CLI Enhancements

### 2.1 Watch Mode

- [ ] Implement file watcher using `fsnotify`
- [ ] Add `sqlfmt watch [path]` command
- [ ] Options:
  - [ ] `--recursive` - watch subdirectories
  - [ ] `--pattern` - file pattern to watch
  - [ ] `--debounce` - delay before formatting
- [ ] Add tests for watch functionality
- [ ] Document watch mode usage

### 2.2 Directory & Git Integration

- [ ] Add `sqlfmt format --recursive` for directory trees
- [ ] Add `sqlfmt format --git-diff` to format only changed files
- [ ] Add `sqlfmt format --git-staged` to format staged files
- [ ] Create installable git pre-commit hook:
  - [ ] Script in `hooks/pre-commit`
  - [ ] Installation command: `sqlfmt install-hook`
  - [ ] Document hook installation

---

## Phase 3: Library API Improvements

### 3.1 Streaming API

- [ ] Add streaming format functions:
  ```go
  func FormatReader(r io.Reader, w io.Writer, cfg *Config) error
  func FormatFile(inputPath, outputPath string, cfg *Config) error
  ```
- [ ] Optimize for large file handling
- [ ] Add streaming tests
- [ ] Document streaming API

### 3.2 Parse Tree Access

- [ ] Expose parsed SQL structure:
  ```go
  type ParsedQuery struct {
      Tokens []Token
      Structure QueryStructure
  }
  func Parse(query string, cfg *Config) (*ParsedQuery, error)
  ```
- [ ] Add query analysis functions:
  - [ ] Detect query type (SELECT, INSERT, UPDATE, etc.)
  - [ ] Extract table names
  - [ ] Extract column references
- [ ] Document parse tree structure

---

## Phase 4: Advanced SQL Features

### 4.1 Enhanced Stored Procedure Support

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

### 4.2 Complete DDL Support

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

### 4.3 Extended Comment Support

- [ ] Structured comments for documentation:
  ```sql
  /**
   * @description: User authentication query
   * @param: $1 - username (text)
   * @param: $2 - password_hash (text)
   * @returns: user_id, role
   */
  SELECT id, role FROM users WHERE username = $1 AND password = $2;
  ```
- [ ] Parse and preserve structured comment format
- [ ] Add documentation generation from comments

---

## Phase 5: Documentation & Polish

### 5.1 Comprehensive Documentation

- [ ] Create dialect comparison guide:
  - [ ] Feature matrix across all dialects
  - [ ] Syntax differences
  - [ ] Migration guides between dialects
- [ ] Add troubleshooting guide:
  - [ ] Common formatting issues
  - [ ] Performance tips
  - [ ] Known limitations

### 5.2 Editor Integrations

- [ ] VSCode extension:
  - [ ] Format on save
  - [ ] Format selection
  - [ ] Configuration UI
  - [ ] Syntax highlighting integration
- [ ] Vim plugin:
  - [ ] Format command
  - [ ] Auto-format on write
  - [ ] Configuration variables
- [ ] Emacs mode:
  - [ ] sqlfmt-mode package
  - [ ] Interactive formatting
  - [ ] Configuration options

---

## Priority Matrix

### 🔴 High Priority (Next 1-2 Weeks)

1. **Test Coverage Improvement** (Phase 1) - Critical for production readiness
   - Fix failing PostgreSQL formatter test
   - Fix failing CLI test (if still exists)

### 🟡 Medium Priority (Next 2-4 Weeks)

1. **Watch mode** (Phase 2.1) - Developer workflow improvement
2. **Directory & Git integration** (Phase 2.2) - CI/CD integration

### 🟢 Low Priority (Future 1-3 Months)

1. **Streaming API** (Phase 3.1) - Large file handling
2. **Enhanced DDL support** (Phase 4.2) - Edge cases
3. **Editor integrations** (Phase 5.2) - Ecosystem expansion

---

## Success Metrics

### Code Quality

- [x] All tests passing (100%) - PostgreSQL ✅, MySQL ✅ - ALL TESTS PASSING!

### Documentation

- [ ] All features documented
- [ ] README examples work
- [ ] Config file guide complete
- [ ] Dialect guides complete

### Performance

- [ ] Format 1000-line file in < 100ms
- [ ] No memory leaks
- [ ] Concurrent formatting supported

### User Experience

- [ ] CLI intuitive and fast
- [ ] Config files easy to use
- [ ] Error messages helpful
- [ ] Works across all platforms

---

**Total Estimated Time**: 8-12 weeks for complete implementation

**Immediate Next Steps** (This Week):
1. ✅ ~~Fix failing PostgreSQL formatter test~~ - COMPLETED
2. ✅ ~~Fix MySQL formatter tests (11 failures)~~ - COMPLETED

**All Tests Passing!** 🎉

Next priorities:
- Integration & performance tests
- Documentation improvements
- Watch mode and CLI enhancements
