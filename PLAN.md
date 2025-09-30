# go-sqlfmt Implementation Plan

This consolidated plan covers all remaining work for MySQL, PostgreSQL, and SQLite dialects, plus strategic enhancements for production readiness.

## Overview

Most dialect-specific features have been successfully implemented. This plan focuses on:

1. Completing remaining checklist items
2. Fixing known issues and bugs
3. Enhancing testing coverage
4. Improving documentation
5. Adding strategic new features

---

## Phase 1: Testing Infrastructure Enhancements

### 1.1 Performance & Stress Testing

- [x] Add benchmark tests in `format_benchmark_test.go`:
  - [x] Small queries (< 100 chars)
  - [x] Medium queries (100-1000 chars)
  - [x] Large queries (1000-10000 chars)
  - [x] Very large queries (> 10000 chars)
  - [x] Deeply nested queries (10+ subqueries)
- [x] Add memory usage tests:
  - [x] Profile memory allocation patterns
  - [x] Test for memory leaks with repeated formatting
  - [x] Test concurrent formatting operations
- [x] Add fuzz testing (expand existing `sqlite_fuzz_test.go`):
  - [x] Add MySQL fuzz tests
  - [x] Add PostgreSQL fuzz tests
  - [x] Test malformed SQL handling
  - [x] Test edge cases and corner cases

### 1.2 Dialect-Specific Golden File Expansion

- [ ] Add more PostgreSQL golden files:
  - [ ] Complex window functions with frames
  - [ ] Recursive CTEs with multiple branches
  - [ ] Advanced JSON/JSONB operations
  - [ ] Full text search queries
- [ ] Add more MySQL golden files:
  - [ ] Complex stored procedures with control flow
  - [ ] Window functions (MySQL 8.0+)
  - [ ] JSON table functions
  - [ ] Full text search with MATCH AGAINST
- [ ] Add more SQLite golden files:
  - [ ] Trigger definitions with complex logic
  - [ ] View definitions with CTEs
  - [ ] PRAGMA statements
  - [ ] FTS5 full-text search queries

---

## Phase 2: Cross-Dialect Features

### 2.1 Dialect Auto-Detection

**Goal**: Automatically detect SQL dialect from file or content

- [ ] Add dialect detection logic in `pkg/sqlfmt/detect.go`:
  - [ ] File extension detection:
    - `.psql`, `.pgsql` → PostgreSQL
    - `.mysql`, `.my.sql` → MySQL
    - `.sqlite`, `.db.sql` → SQLite
    - `.plsql`, `.ora.sql` → PL/SQL
  - [ ] Content-based detection (heuristics):
    - PostgreSQL: `::`, `$$`, `$1` placeholders, `RETURNING`
    - MySQL: backticks, `?` placeholders, `ON DUPLICATE KEY UPDATE`
    - SQLite: `?`, `:name` placeholders, `WITHOUT ROWID`, `PRAGMA`
    - PL/SQL: `BEGIN ... END;`, `EXCEPTION`
- [ ] Add `--auto-detect` CLI flag
- [ ] Add tests for detection accuracy
- [ ] Document detection logic and limitations

### 2.2 Multi-Dialect Support

**Goal**: Handle projects with multiple SQL dialects

- [ ] Add `.sqlfmtignore` file support (like `.gitignore`)
- [ ] Add per-directory config override support
- [ ] Add inline dialect hints: `-- sqlfmt: dialect=postgresql`
- [ ] Document multi-dialect project setup

---

## Phase 3: Advanced Formatting Options

### 3.1 Alignment Options

- [ ] Add vertical alignment configuration:
  - [ ] Column name alignment in SELECT
  - [ ] Assignment operator alignment in UPDATE
  - [ ] Values alignment in INSERT
- [ ] Add configuration options in `Config`:
  - [ ] `AlignColumnNames bool`
  - [ ] `AlignAssignments bool`
  - [ ] `AlignValues bool`
- [ ] Implement alignment logic in formatter
- [ ] Add tests for alignment options
- [ ] Document in configuration guide

### 3.2 Line Length Limits

- [ ] Add `MaxLineLength int` configuration option
- [ ] Implement line breaking logic:
  - [ ] Break long SELECT column lists
  - [ ] Break long WHERE conditions
  - [ ] Break long function calls
  - [ ] Smart break at appropriate points (commas, operators)
- [ ] Add tests for line length enforcement
- [ ] Document behavior and limitations

### 3.3 Comment Handling Improvements

- [ ] Improve inline comment positioning:
  - [ ] Keep inline comments on same line when possible
  - [ ] Move to next line if line length exceeded
  - [ ] Preserve relative position in code
- [ ] Add comment formatting options:
  - [ ] `PreserveCommentIndent bool`
  - [ ] `CommentMinSpacing int` - spaces before inline comments
- [ ] Test comment handling edge cases

---

## Phase 4: CLI Enhancements

### 4.1 Watch Mode

- [ ] Implement file watcher using `fsnotify`
- [ ] Add `sqlfmt watch [path]` command
- [ ] Options:
  - [ ] `--recursive` - watch subdirectories
  - [ ] `--pattern` - file pattern to watch
  - [ ] `--debounce` - delay before formatting
- [ ] Add tests for watch functionality
- [ ] Document watch mode usage

### 4.2 Directory & Git Integration

- [ ] Add `sqlfmt format --recursive` for directory trees
- [ ] Add `sqlfmt format --git-diff` to format only changed files
- [ ] Add `sqlfmt format --git-staged` to format staged files
- [ ] Create installable git pre-commit hook:
  - [ ] Script in `hooks/pre-commit`
  - [ ] Installation command: `sqlfmt install-hook`
  - [ ] Document hook installation

### 4.3 Validation Improvements

- [ ] Enhance `sqlfmt validate` command:
  - [ ] Exit code: 0 (all valid), 1 (some invalid)
  - [ ] JSON output mode: `--output=json`
  - [ ] Summary report: files checked, files valid, files invalid
- [ ] Add `sqlfmt check` alias for validate
- [ ] Add `--diff` flag to show what would change

---

## Phase 5: Library API Improvements

### 5.1 Streaming API

- [ ] Add streaming format functions:
  ```go
  func FormatReader(r io.Reader, w io.Writer, cfg *Config) error
  func FormatFile(inputPath, outputPath string, cfg *Config) error
  ```
- [ ] Optimize for large file handling
- [ ] Add streaming tests
- [ ] Document streaming API

### 5.2 Parse Tree Access

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

### 5.3 Custom Formatter Plugin System

- [ ] Define plugin interface:
  ```go
  type FormatterPlugin interface {
      Name() string
      Format(query string, cfg *Config) (string, error)
      SupportsDialect(lang Language) bool
  }
  ```
- [ ] Add plugin registration system
- [ ] Create example plugins
- [ ] Document plugin development

---

## Phase 6: Advanced SQL Features

### 6.1 Enhanced Stored Procedure Support

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

### 6.2 Complete DDL Support

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

### 6.3 Extended Comment Support

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

## Phase 7: Documentation & Polish

### 7.1 Comprehensive Documentation

- [ ] Create dialect comparison guide:
  - [ ] Feature matrix across all dialects
  - [ ] Syntax differences
  - [ ] Migration guides between dialects
- [ ] Add troubleshooting guide:
  - [ ] Common formatting issues
  - [ ] Performance tips
  - [ ] Known limitations

### 7.2 Editor Integrations

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

### 7.3 Examples & Demos

- [ ] Create example projects:
  - [ ] PostgreSQL application
  - [ ] MySQL application
  - [ ] SQLite application
  - [ ] Multi-dialect project
- [ ] Add interactive online demo
- [ ] Create GIF demos for README

---

## Priority Matrix

### 🟡 Medium Priority (Next 2-4 Weeks)

1. **Dialect auto-detection** (Phase 2.1) - High user value
2. **Watch mode** (Phase 4.1) - Developer workflow improvement
3. ~~**Performance benchmarks** (Phase 1.1) - Ensure scalability~~ ✅ **COMPLETED**
4. **Directory & Git integration** (Phase 4.2) - CI/CD integration

### 🟢 Low Priority (Future 1-3 Months)

1. **Advanced alignment options** (Phase 3.1) - Nice-to-have
2. **Streaming API** (Phase 5.1) - Large file handling
3. **Line length limits** (Phase 3.2) - Style preference
4. **Enhanced DDL support** (Phase 6.2) - Edge cases
5. **Editor integrations** (Phase 7.2) - Ecosystem expansion

### ⚪ On Hold / Future

1. **Plugin system** (Phase 5.3) - Complex feature, low demand
2. **Documentation generation** (Phase 6.3) - Advanced use case

---

## Success Metrics

### Code Quality

- [ ] All tests passing (100%)
- [ ] Test coverage > 90%
- [ ] No linting errors
- [ ] All code formatted consistently

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

## Timeline Estimate

- **Phase 1 (Testing)**: 1-2 weeks
- **Phase 2 (Cross-Dialect)**: 1 week
- **Phase 3 (Advanced Formatting)**: 2 weeks
- **Phase 4 (CLI Enhancements)**: 2 weeks
- **Phase 5 (Library API)**: 1-2 weeks
- **Phase 6 (Advanced SQL)**: 2-3 weeks
- **Phase 7 (Documentation & Polish)**: 1-2 weeks

**Total Estimated Time**: 10-14 weeks for complete implementation
