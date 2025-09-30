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

## Phase 1: Code Quality & Immediate Fixes

### 1.1 Format & Lint Issues ⚠️ URGENT

**Status**: Blocking CI/CD

- [x] Run `just fmt` to fix formatting violations
- [x] Run `just lint` to verify no remaining issues
- [x] Commit formatting fixes

### 1.2 SQLite Phase 2 Tokenizer Verification

**Status**: Implementation exists but checklist shows incomplete

- [ ] Verify identifier quoting works correctly:
  - [ ] Test double-quotes: `"table_name"`
  - [ ] Test backticks: `` `column_name` ``
  - [ ] Test brackets: `[identifier name]`
  - [ ] Test mixed usage in same query
- [ ] Verify blob literal handling:
  - [ ] Test uppercase: `X'DEADBEEF'`
  - [ ] Test lowercase: `x'cafebabe'`
  - [ ] Test in various contexts (SELECT, INSERT, WHERE)
- [ ] Verify comment handling:
  - [ ] Confirm `--` comments work
  - [ ] Confirm `/* */` comments work
  - [ ] Confirm `#` is NOT treated as comment (unlike MySQL)
- [ ] Update `PLAN-SQLITE.md` Phase 2 checkboxes to [x]

---

## Phase 2: PostgreSQL Known Issues

### 2.1 UPSERT Formatting Problem 🚧 BLOCKED

**Issue**: `INSERT ... ON CONFLICT ... DO UPDATE/DO NOTHING` doesn't format correctly

**Current Behavior**:

```sql
INSERT INTO users (id, name) VALUES (1, 'John')
ON CONFLICT (id)
DO
  UPDATE SET name = 'Jane';
```

**Expected Behavior**:

```sql
INSERT INTO users (id, name) VALUES (1, 'John')
ON CONFLICT (id) DO UPDATE SET name = 'Jane';
```

**Root Cause**:

- Tokenizer doesn't recognize compound keywords like "DO UPDATE" as single units
- Context detection between procedural `DO` blocks and UPSERT `DO` clauses fails
- Requires deeper architectural changes to tokenizer or formatter core

**Options**:

- [ ] **Option A**: Modify tokenizer core to support compound keyword lookahead
- [ ] **Option B**: Implement two-pass formatting with context awareness
- [ ] **Option C**: Add special handling in core formatting logic (not just token override)
- [ ] **Option D**: Accept current behavior and document as known limitation

**Decision**: Mark as known limitation for now; defer to future major refactor

### 2.2 Function Indentation Test Mismatches

**Issue**: Some PostgreSQL function definition tests fail due to indentation expectations

**Tasks**:

- [ ] Review failing test cases in `postgresql_formatter_test.go`
- [ ] Determine if issue is:
  - [ ] Test expectations need updating to match actual (correct) behavior
  - [ ] Code needs fixing to match test expectations
- [ ] Update tests or code accordingly
- [ ] Document decision in commit message

---

## Phase 3: Configuration File System Documentation

### 3.1 Document New Config File Feature

**Status**: Code implemented (`configfile.go`, `example.sqlfmt.yaml`) but not documented

- [ ] Update `docs/configuration.md`:
  - [ ] Add "Configuration Files" section
  - [ ] Document supported file names: `.sqlfmtrc`, `.sqlfmt.yaml`, `.sqlfmt.yml`, `sqlfmt.yaml`, `sqlfmt.yml`
  - [ ] Document search order:
    1. Current directory (and parents up to git root)
    2. User home directory
  - [ ] Document configuration precedence: CLI flags > config file > defaults
  - [ ] Add YAML configuration examples
  - [ ] Document all config options: `language`, `indent`, `keyword_case`, `lines_between_queries`
- [ ] Update `docs/cli-usage.md`:
  - [ ] Add section on config file usage
  - [ ] Show example of project-specific config
  - [ ] Show example of user-wide config
- [ ] Update `README.md`:
  - [ ] Add config file mention in "Quick Start"
  - [ ] Link to configuration documentation
- [ ] Consider adding `sqlfmt init` command to generate example config

### 3.2 Config File Testing

- [ ] Add tests for config file loading in `configfile_test.go`:
  - [ ] Test loading from current directory
  - [ ] Test loading from parent directories
  - [ ] Test loading from home directory
  - [ ] Test search order precedence
  - [ ] Test YAML parsing for all options
  - [ ] Test error handling for invalid YAML
  - [ ] Test error handling for unknown options

---

## Phase 4: Testing Infrastructure Enhancements

### 4.1 Real-World Scenario Tests

**Status**: PostgreSQL Phase 13 partially complete

- [ ] Create `testdata/scenarios/` directory structure:
  - [ ] `migrations/` - Database migration scripts
  - [ ] `procedures/` - Stored procedures and functions
  - [ ] `analytics/` - Complex analytical queries
  - [ ] `mixed/` - Combined DDL and DML operations
- [ ] Add real-world test cases:
  - [ ] Migration scripts with multi-step changes
  - [ ] Production-style stored procedures with error handling
  - [ ] Complex window function queries
  - [ ] Large UNION/EXCEPT/INTERSECT queries
  - [ ] Deeply nested CTEs (3+ levels)
  - [ ] Mixed DDL/DML transactions
- [ ] Create `scenario_test.go` to test these cases

### 4.2 Performance & Stress Testing

- [ ] Add benchmark tests in `format_benchmark_test.go`:
  - [ ] Small queries (< 100 chars)
  - [ ] Medium queries (100-1000 chars)
  - [ ] Large queries (1000-10000 chars)
  - [ ] Very large queries (> 10000 chars)
  - [ ] Deeply nested queries (10+ subqueries)
- [ ] Add memory usage tests:
  - [ ] Profile memory allocation patterns
  - [ ] Test for memory leaks with repeated formatting
  - [ ] Test concurrent formatting operations
- [ ] Add fuzz testing (expand existing `sqlite_fuzz_test.go`):
  - [ ] Add MySQL fuzz tests
  - [ ] Add PostgreSQL fuzz tests
  - [ ] Test malformed SQL handling
  - [ ] Test edge cases and corner cases

### 4.3 Dialect-Specific Golden File Expansion

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

## Phase 5: Cross-Dialect Features

### 5.1 Dialect Auto-Detection

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

### 5.2 Multi-Dialect Support

**Goal**: Handle projects with multiple SQL dialects

- [ ] Add `.sqlfmtignore` file support (like `.gitignore`)
- [ ] Add per-directory config override support
- [ ] Add inline dialect hints: `-- sqlfmt: dialect=postgresql`
- [ ] Document multi-dialect project setup

---

## Phase 6: Advanced Formatting Options

### 6.1 Alignment Options

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

### 6.2 Line Length Limits

- [ ] Add `MaxLineLength int` configuration option
- [ ] Implement line breaking logic:
  - [ ] Break long SELECT column lists
  - [ ] Break long WHERE conditions
  - [ ] Break long function calls
  - [ ] Smart break at appropriate points (commas, operators)
- [ ] Add tests for line length enforcement
- [ ] Document behavior and limitations

### 6.3 Comment Handling Improvements

- [ ] Improve inline comment positioning:
  - [ ] Keep inline comments on same line when possible
  - [ ] Move to next line if line length exceeded
  - [ ] Preserve relative position in code
- [ ] Add comment formatting options:
  - [ ] `PreserveCommentIndent bool`
  - [ ] `CommentMinSpacing int` - spaces before inline comments
- [ ] Test comment handling edge cases

---

## Phase 7: CLI Enhancements

### 7.1 Watch Mode

- [ ] Implement file watcher using `fsnotify`
- [ ] Add `sqlfmt watch [path]` command
- [ ] Options:
  - [ ] `--recursive` - watch subdirectories
  - [ ] `--pattern` - file pattern to watch
  - [ ] `--debounce` - delay before formatting
- [ ] Add tests for watch functionality
- [ ] Document watch mode usage

### 7.2 Directory & Git Integration

- [ ] Add `sqlfmt format --recursive` for directory trees
- [ ] Add `sqlfmt format --git-diff` to format only changed files
- [ ] Add `sqlfmt format --git-staged` to format staged files
- [ ] Create installable git pre-commit hook:
  - [ ] Script in `hooks/pre-commit`
  - [ ] Installation command: `sqlfmt install-hook`
  - [ ] Document hook installation

### 7.3 Validation Improvements

- [ ] Enhance `sqlfmt validate` command:
  - [ ] Exit code: 0 (all valid), 1 (some invalid)
  - [ ] JSON output mode: `--output=json`
  - [ ] Summary report: files checked, files valid, files invalid
- [ ] Add `sqlfmt check` alias for validate
- [ ] Add `--diff` flag to show what would change

---

## Phase 8: Library API Improvements

### 8.1 Streaming API

- [ ] Add streaming format functions:
  ```go
  func FormatReader(r io.Reader, w io.Writer, cfg *Config) error
  func FormatFile(inputPath, outputPath string, cfg *Config) error
  ```
- [ ] Optimize for large file handling
- [ ] Add streaming tests
- [ ] Document streaming API

### 8.2 Parse Tree Access

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

### 8.3 Custom Formatter Plugin System

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

## Phase 9: Advanced SQL Features

### 9.1 Enhanced Stored Procedure Support

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

### 9.2 Complete DDL Support

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

### 9.3 Extended Comment Support

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

## Phase 10: Documentation & Polish

### 10.1 Comprehensive Documentation

- [ ] Create dialect comparison guide:
  - [ ] Feature matrix across all dialects
  - [ ] Syntax differences
  - [ ] Migration guides between dialects
- [ ] Add troubleshooting guide:
  - [ ] Common formatting issues
  - [ ] Performance tips
  - [ ] Known limitations
- [ ] Create video tutorials:
  - [ ] Getting started
  - [ ] CLI usage
  - [ ] Library integration
  - [ ] Advanced features

### 10.2 Editor Integrations

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

### 10.3 Examples & Demos

- [ ] Create example projects:
  - [ ] PostgreSQL application
  - [ ] MySQL application
  - [ ] SQLite application
  - [ ] Multi-dialect project
- [ ] Add interactive online demo
- [ ] Create GIF demos for README

---

## Priority Matrix

### 🔴 High Priority (Immediate - Next 1-2 Weeks)

1. **Fix formatting issues** (Phase 1.1) - Blocks CI/CD
2. **SQLite Phase 2 verification** (Phase 1.2) - Complete existing work
3. **Config file documentation** (Phase 3.1) - Feature is implemented but undocumented
4. **PostgreSQL function tests** (Phase 2.2) - Fix test failures

### 🟡 Medium Priority (Next 2-4 Weeks)

1. **Dialect auto-detection** (Phase 5.1) - High user value
2. **Watch mode** (Phase 7.1) - Developer workflow improvement
3. **Real-world scenario tests** (Phase 4.1) - Quality assurance
4. **Performance benchmarks** (Phase 4.2) - Ensure scalability
5. **Directory & Git integration** (Phase 7.2) - CI/CD integration

### 🟢 Low Priority (Future 1-3 Months)

1. **Advanced alignment options** (Phase 6.1) - Nice-to-have
2. **Streaming API** (Phase 8.1) - Large file handling
3. **Line length limits** (Phase 6.2) - Style preference
4. **Enhanced DDL support** (Phase 9.2) - Edge cases
5. **Editor integrations** (Phase 10.2) - Ecosystem expansion

### ⚪ On Hold / Future

1. **PostgreSQL UPSERT formatting** (Phase 2.1) - Requires tokenizer redesign
2. **Plugin system** (Phase 8.3) - Complex feature, low demand
3. **Documentation generation** (Phase 9.3) - Advanced use case
4. **Video tutorials** (Phase 10.1) - Resource intensive

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

## Known Limitations

### PostgreSQL UPSERT Formatting 🚧

**Issue**: `ON CONFLICT ... DO UPDATE/DO NOTHING` adds unwanted line breaks

**Workaround**: Manually adjust UPSERT formatting after running formatter

**Long-term Solution**: Requires tokenizer architecture redesign to handle compound keywords

**Status**: Documented limitation, deferred to future major version

---

## Timeline Estimate

- **Phase 1 (Code Quality)**: 1-2 days
- **Phase 2 (PostgreSQL Issues)**: 2-3 days
- **Phase 3 (Config Documentation)**: 2-3 days
- **Phase 4 (Testing)**: 1-2 weeks
- **Phase 5 (Cross-Dialect)**: 1 week
- **Phase 6 (Advanced Formatting)**: 2 weeks
- **Phase 7 (CLI Enhancements)**: 2 weeks
- **Phase 8 (Library API)**: 1-2 weeks
- **Phase 9 (Advanced SQL)**: 2-3 weeks
- **Phase 10 (Documentation & Polish)**: 1-2 weeks

**Total Estimated Time**: 10-14 weeks for complete implementation

---

## Next Steps

1. Run `just fmt` to fix immediate formatting issues
2. Verify SQLite Phase 2 implementation
3. Document configuration file system
4. Fix PostgreSQL function indentation tests
5. Begin real-world testing scenarios

**Created**: 2025-09-30
**Status**: Active Development
**Last Updated**: 2025-09-30
