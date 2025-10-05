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

## Phase 1: CLI Enhancements

### 1.1 Watch Mode

- [ ] Implement file watcher using `fsnotify`
- [ ] Add `sqlfmt watch [path]` command
- [ ] Options:
  - [ ] `--recursive` - watch subdirectories
  - [ ] `--pattern` - file pattern to watch
  - [ ] `--debounce` - delay before formatting
- [ ] Add tests for watch functionality
- [ ] Document watch mode usage

### 1.2 Directory & Git Integration

- [ ] Add `sqlfmt format --recursive` for directory trees
- [ ] Add `sqlfmt format --git-diff` to format only changed files
- [ ] Add `sqlfmt format --git-staged` to format staged files
- [ ] Create installable git pre-commit hook:
  - [ ] Script in `hooks/pre-commit`
  - [ ] Installation command: `sqlfmt install-hook`
  - [ ] Document hook installation

---

## Phase 2: Library API Improvements

### 2.1 Streaming API

- [ ] Add streaming format functions:
  ```go
  func FormatReader(r io.Reader, w io.Writer, cfg *Config) error
  func FormatFile(inputPath, outputPath string, cfg *Config) error
  ```
- [ ] Optimize for large file handling
- [ ] Add streaming tests
- [ ] Document streaming API

### 2.2 Parse Tree Access

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

## Phase 3: Advanced SQL Features

### 3.1 Enhanced Stored Procedure Support

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

### 3.2 Complete DDL Support

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

### 3.3 Extended Comment Support

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

## Phase 4: Documentation & Polish

### 4.1 Comprehensive Documentation

- [ ] Create dialect comparison guide:
  - [ ] Feature matrix across all dialects
  - [ ] Syntax differences
  - [ ] Migration guides between dialects
- [ ] Add troubleshooting guide:
  - [ ] Common formatting issues
  - [ ] Performance tips
  - [ ] Known limitations

### 4.2 Editor Integrations

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

### 🟡 Medium Priority (Next 2-4 Weeks)

1. **Watch mode** (Phase 1.1) - Developer workflow improvement
2. **Directory & Git integration** (Phase 1.2) - CI/CD integration

### 🟢 Low Priority (Future 1-3 Months)

1. **Streaming API** (Phase 2.1) - Large file handling
2. **Enhanced DDL support** (Phase 3.2) - Edge cases
3. **Editor integrations** (Phase 4.2) - Ecosystem expansion

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
